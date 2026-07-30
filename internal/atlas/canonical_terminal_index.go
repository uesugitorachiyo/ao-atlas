package atlas

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	canonicalTerminalInputContract    = "ao.canonical-terminal-index-input.v1"
	canonicalTerminalIndexContract    = "ao.canonical-terminal-index.v1"
	canonicalTerminalArtifactMaxBytes = 1 << 20
	canonicalTerminalTotalMaxBytes    = 16 << 20
	canonicalTerminalMaxArtifacts     = 128
)

type CanonicalTerminalArtifact struct {
	Role     string `json:"role"`
	Sequence int    `json:"sequence"`
	Path     string `json:"path"`
	Schema   string `json:"schema"`
	SHA256   string `json:"sha256"`
	State    string `json:"state"`
}

type CanonicalTerminalLineage struct {
	FromSequence int    `json:"from_sequence"`
	ToSequence   int    `json:"to_sequence"`
	Relation     string `json:"relation"`
}

type CanonicalTerminalCounts struct {
	Total     int `json:"total"`
	Minimum   int `json:"minimum"`
	Completed int `json:"completed"`
	Ready     int `json:"ready"`
	Blocked   int `json:"blocked"`
	Failed    int `json:"failed"`
}

type CanonicalTerminalLease struct {
	MinimumMinutes int    `json:"minimum_minutes"`
	TargetMinutes  int    `json:"target_minutes"`
	MaximumMinutes int    `json:"maximum_minutes"`
	ElapsedMinutes int    `json:"elapsed_minutes"`
	Status         string `json:"status"`
}

type CanonicalTerminalSafety struct {
	ExecutesWork        bool `json:"executes_work"`
	ApprovesWork        bool `json:"approves_work"`
	MutatesRepositories bool `json:"mutates_repositories"`
	CallsProviders      bool `json:"calls_providers"`
	Publishes           bool `json:"publishes"`
	Releases            bool `json:"releases"`
	Deploys             bool `json:"deploys"`
	AdvancesAuthority   bool `json:"advances_authority"`
}

type CanonicalTerminalIndex struct {
	ContractVersion      string                      `json:"contract_version"`
	SchemaDigest         string                      `json:"schema_digest"`
	MissionID            string                      `json:"mission_id"`
	EvidenceRoot         string                      `json:"evidence_root"`
	GeneratedAtUTC       string                      `json:"generated_at_utc"`
	Artifacts            []CanonicalTerminalArtifact `json:"artifacts"`
	Lineage              []CanonicalTerminalLineage  `json:"lineage"`
	TerminalReference    string                      `json:"terminal_reference"`
	Counts               CanonicalTerminalCounts     `json:"counts"`
	Lease                CanonicalTerminalLease      `json:"lease"`
	CompletionObserved   bool                        `json:"completion_observed"`
	CanonicalAgreement   bool                        `json:"canonical_evidence_agreement"`
	ReadinessPassed      bool                        `json:"readiness_passed"`
	ReturnGateStatus     string                      `json:"return_gate_status"`
	FinalResponseAllowed bool                        `json:"final_response_allowed"`
	Conflicts            []string                    `json:"conflict_codes"`
	ConflictSummaries    []string                    `json:"conflict_summaries"`
	ExactNextAction      string                      `json:"exact_next_action"`
	Safety               CanonicalTerminalSafety     `json:"safety_boundaries"`
	Digest               string                      `json:"digest"`
}

type canonicalTerminalManifest struct {
	ContractVersion string `json:"contract_version"`
	MissionID       string `json:"mission_id"`
	EvidenceRoot    string `json:"evidence_root"`
	GeneratedAtUTC  string `json:"generated_at_utc"`
	Artifacts       []struct {
		Role     string `json:"role"`
		Sequence int    `json:"sequence"`
		Path     string `json:"path"`
		State    string `json:"state"`
		SHA256   string `json:"sha256"`
	} `json:"artifacts"`
}

type terminalObservation struct {
	missionID      string
	completed      int
	ready          int
	blocked        int
	failed         int
	elapsed        int
	minimumNodes   int
	minimumMinutes int
	targetMinutes  int
	maximumMinutes int
	leaseStatus    string
	final          bool
	nextAction     string
	safety         CanonicalTerminalSafety
	schema         string
}

func BuildCanonicalTerminalIndex(root, manifestPath string) (CanonicalTerminalIndex, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return CanonicalTerminalIndex{}, err
	}
	manifestData, err := readBoundedRegularFile(manifestPath, canonicalTerminalArtifactMaxBytes)
	if err != nil {
		return CanonicalTerminalIndex{}, fmt.Errorf("manifest: %w", err)
	}
	var manifest canonicalTerminalManifest
	if err := decodeStrictJSON(manifestData, &manifest); err != nil {
		return CanonicalTerminalIndex{}, fmt.Errorf("manifest: %w", err)
	}
	if manifest.ContractVersion != canonicalTerminalInputContract {
		return CanonicalTerminalIndex{}, fmt.Errorf("manifest contract_version must be %s", canonicalTerminalInputContract)
	}
	if strings.TrimSpace(manifest.MissionID) == "" {
		return CanonicalTerminalIndex{}, errors.New("manifest mission_id must not be empty")
	}
	if manifest.EvidenceRoot != "." {
		return CanonicalTerminalIndex{}, errors.New("manifest evidence_root must be .")
	}
	if _, err := time.Parse(time.RFC3339, manifest.GeneratedAtUTC); err != nil {
		return CanonicalTerminalIndex{}, errors.New("manifest generated_at_utc must be RFC3339")
	}
	if len(manifest.Artifacts) == 0 || len(manifest.Artifacts) > canonicalTerminalMaxArtifacts {
		return CanonicalTerminalIndex{}, errors.New("manifest artifact count is outside limits")
	}

	index := CanonicalTerminalIndex{
		ContractVersion: canonicalTerminalIndexContract,
		SchemaDigest:    DigestBytes([]byte(canonicalTerminalIndexContract)),
		MissionID:       manifest.MissionID,
		EvidenceRoot:    ".",
		GeneratedAtUTC:  manifest.GeneratedAtUTC,
		Conflicts:       []string{},
		Lineage:         []CanonicalTerminalLineage{},
		Safety:          CanonicalTerminalSafety{},
	}
	observations := map[string]terminalObservation{}
	seenRoles := map[string]bool{}
	total := 0
	previousSequence := -1
	for _, item := range manifest.Artifacts {
		if item.Sequence <= previousSequence {
			return CanonicalTerminalIndex{}, errors.New("artifact sequence is non-monotonic")
		}
		previousSequence = item.Sequence
		if seenRoles[item.Role] {
			return CanonicalTerminalIndex{}, fmt.Errorf("duplicate artifact role %q", item.Role)
		}
		seenRoles[item.Role] = true
		if err := validateArtifactState(item.Role, item.State); err != nil {
			return CanonicalTerminalIndex{}, err
		}
		path, err := containedEvidencePath(rootAbs, item.Path)
		if err != nil {
			return CanonicalTerminalIndex{}, err
		}
		data, err := readBoundedRegularFile(path, canonicalTerminalArtifactMaxBytes)
		if err != nil {
			return CanonicalTerminalIndex{}, fmt.Errorf("%s artifact: %w", item.Role, err)
		}
		total += len(data)
		if total > canonicalTerminalTotalMaxBytes {
			return CanonicalTerminalIndex{}, errors.New("evidence total exceeds size limit")
		}
		if DigestBytes(data) != item.SHA256 {
			return CanonicalTerminalIndex{}, fmt.Errorf("%s artifact digest mismatch", item.Role)
		}
		var raw map[string]any
		if err := decodeStrictJSON(data, &raw); err != nil {
			return CanonicalTerminalIndex{}, fmt.Errorf("%s artifact: %w", item.Role, err)
		}
		observation := observeTerminalArtifact(raw)
		if observation.missionID != "" && observation.missionID != manifest.MissionID {
			return CanonicalTerminalIndex{}, fmt.Errorf("%s artifact mission identity mismatch", item.Role)
		}
		observations[item.Role] = observation
		index.Artifacts = append(index.Artifacts, CanonicalTerminalArtifact{
			Role: item.Role, Sequence: item.Sequence, Path: item.Path, Schema: observation.schema,
			SHA256: item.SHA256, State: item.State,
		})
		if len(index.Artifacts) > 1 {
			prior := index.Artifacts[len(index.Artifacts)-2]
			index.Lineage = append(index.Lineage, CanonicalTerminalLineage{
				FromSequence: prior.Sequence, ToSequence: item.Sequence, Relation: "precedes",
			})
		}
	}
	if !seenRoles["lease"] || !seenRoles["root"] {
		return CanonicalTerminalIndex{}, errors.New("lease and root artifacts are required")
	}
	rootState := observations["root"]
	leaseState := observations["lease"]
	index.Counts.Minimum = leaseState.minimumNodes
	index.Lease.MinimumMinutes = leaseState.minimumMinutes
	index.Lease.TargetMinutes = leaseState.targetMinutes
	index.Lease.MaximumMinutes = leaseState.maximumMinutes

	terminal, hasTerminal := observations["terminal"]
	if !hasTerminal {
		index.Counts.Total = rootState.completed + rootState.ready + rootState.blocked + rootState.failed
		index.Counts.Completed = rootState.completed
		index.Counts.Ready = rootState.ready
		index.Counts.Blocked = rootState.blocked
		index.Counts.Failed = rootState.failed
		index.Conflicts = append(index.Conflicts, "canonical_terminal_missing")
		index.ExactNextAction = "Produce a governed terminal observation and rebuild the canonical index."
		index.Lease.Status = "terminal_missing"
		index.ReturnGateStatus = "final_response_denied"
		return finalizeCanonicalTerminalIndex(index), nil
	}
	index.TerminalReference = artifactPathForRole(index.Artifacts, "terminal")
	if index.Counts.Minimum == 0 {
		index.Counts.Minimum = terminal.minimumNodes
	}
	if index.Counts.Minimum == 0 {
		index.Counts.Minimum = rootState.minimumNodes
	}
	if terminal.completed < rootState.completed {
		return CanonicalTerminalIndex{}, errors.New("terminal completed count is non-monotonic")
	}
	if rootState.ready > 0 && terminal.completed == rootState.completed && terminal.ready < rootState.ready {
		return CanonicalTerminalIndex{}, errors.New("root and terminal disagree without advancing lineage")
	}
	index.Counts.Completed = terminal.completed
	index.Counts.Ready = terminal.ready
	index.Counts.Blocked = terminal.blocked
	index.Counts.Failed = terminal.failed
	index.Counts.Total = terminal.completed + terminal.ready + terminal.blocked + terminal.failed
	rootTotal := rootState.completed + rootState.ready + rootState.blocked + rootState.failed
	if rootTotal > 0 && index.Counts.Total != rootTotal {
		return CanonicalTerminalIndex{}, errors.New("root and terminal node counts are non-monotonic")
	}
	index.Lease.ElapsedMinutes = terminal.elapsed
	index.CompletionObserved = terminal.completed >= index.Counts.Minimum && index.Counts.Minimum > 0
	index.CanonicalAgreement = true
	index.Safety = terminal.safety
	if index.Lease.MinimumMinutes <= 0 || index.Lease.MaximumMinutes <= 0 ||
		index.Lease.MaximumMinutes < index.Lease.MinimumMinutes {
		index.Conflicts = append(index.Conflicts, "lease_window_invalid")
	}
	if index.Lease.TargetMinutes <= 0 {
		index.Conflicts = append(index.Conflicts, "lease_target_missing")
	}

	if duration, ok := observations["duration"]; ok {
		if duration.completed != 0 && duration.completed != terminal.completed {
			index.Conflicts = append(index.Conflicts, "duration_state_stale")
		}
		if duration.elapsed != 0 && duration.elapsed != terminal.elapsed {
			index.Conflicts = append(index.Conflicts, "duration_state_stale")
		}
	}
	switch {
	case terminal.elapsed < index.Lease.MinimumMinutes:
		index.Lease.Status = "minimum_not_met"
		index.Conflicts = append(index.Conflicts, "lease_minimum_not_met")
	case index.Lease.MaximumMinutes > 0 && terminal.elapsed > index.Lease.MaximumMinutes:
		index.Lease.Status = "maximum_exceeded"
		index.Conflicts = append(index.Conflicts, "lease_maximum_exceeded")
	default:
		index.Lease.Status = "within_window"
	}
	if terminal.leaseStatus != "" && terminal.leaseStatus != index.Lease.Status {
		index.Conflicts = append(index.Conflicts, "terminal_lease_status_mismatch")
	}
	if terminal.ready > 0 || terminal.blocked > 0 {
		index.Conflicts = append(index.Conflicts, "unfinished_work_final_response")
	}
	if terminal.failed > 0 {
		index.Conflicts = append(index.Conflicts, "failed_work_final_response")
	}
	if terminal.final && (terminal.ready > 0 || terminal.blocked > 0 || terminal.failed > 0 || index.Lease.Status != "within_window") {
		index.Conflicts = append(index.Conflicts, "terminal_final_response_allowed_despite_violation")
	}
	if terminal.nextAction != "" && !isNoAction(terminal.nextAction) && terminal.final {
		index.Conflicts = append(index.Conflicts, "terminal_exact_next_action_requires_execution")
	}
	if terminal.safety.ExecutesWork || terminal.safety.ApprovesWork || terminal.safety.MutatesRepositories ||
		terminal.safety.CallsProviders || terminal.safety.Publishes || terminal.safety.Releases ||
		terminal.safety.Deploys || terminal.safety.AdvancesAuthority {
		index.Conflicts = append(index.Conflicts, "unsafe_boundary_flag")
	}
	index.ReadinessPassed = index.CompletionObserved && len(index.Conflicts) == 0
	index.FinalResponseAllowed = index.ReadinessPassed && terminal.final
	if index.ReadinessPassed {
		index.ExactNextAction = "none"
	} else {
		index.ExactNextAction = "Review the canonical conflict codes and produce a fresh governed terminal observation."
	}
	if index.FinalResponseAllowed {
		index.ReturnGateStatus = "final_response_allowed"
	} else {
		index.ReturnGateStatus = "final_response_denied"
	}
	return finalizeCanonicalTerminalIndex(index), nil
}

func VerifyCanonicalTerminalIndex(root string, index CanonicalTerminalIndex) error {
	if index.ContractVersion != canonicalTerminalIndexContract {
		return fmt.Errorf("index contract_version must be %s", canonicalTerminalIndexContract)
	}
	if index.SchemaDigest != DigestBytes([]byte(canonicalTerminalIndexContract)) {
		return errors.New("index schema digest mismatch")
	}
	if index.MissionID == "" || index.EvidenceRoot != "." {
		return errors.New("index identity is invalid")
	}
	if _, err := time.Parse(time.RFC3339, index.GeneratedAtUTC); err != nil {
		return errors.New("index generation timestamp is invalid")
	}
	expected := index
	expected.Digest = ""
	data, err := json.Marshal(expected)
	if err != nil {
		return err
	}
	if index.Digest != DigestBytes(data) {
		return errors.New("index digest mismatch")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	seenRoles := map[string]bool{}
	previousSequence := -1
	for position, artifact := range index.Artifacts {
		if artifact.Sequence <= previousSequence || seenRoles[artifact.Role] {
			return errors.New("index artifact lineage is invalid")
		}
		previousSequence = artifact.Sequence
		seenRoles[artifact.Role] = true
		if artifact.Schema == "" {
			return fmt.Errorf("%s artifact schema is missing", artifact.Role)
		}
		if err := validateArtifactState(artifact.Role, artifact.State); err != nil {
			return err
		}
		if position > 0 {
			edge := index.Lineage[position-1]
			prior := index.Artifacts[position-1]
			if edge.FromSequence != prior.Sequence || edge.ToSequence != artifact.Sequence || edge.Relation != "precedes" {
				return errors.New("index explicit lineage is invalid")
			}
		}
		path, err := containedEvidencePath(rootAbs, artifact.Path)
		if err != nil {
			return err
		}
		data, err := readBoundedRegularFile(path, canonicalTerminalArtifactMaxBytes)
		if err != nil {
			return err
		}
		if DigestBytes(data) != artifact.SHA256 {
			return fmt.Errorf("%s artifact digest mismatch", artifact.Role)
		}
	}
	if len(index.Lineage) != maxInt(0, len(index.Artifacts)-1) {
		return errors.New("index lineage length is invalid")
	}
	if index.TerminalReference != artifactPathForRole(index.Artifacts, "terminal") {
		return errors.New("index terminal reference is invalid")
	}
	if index.Counts.Total != index.Counts.Completed+index.Counts.Ready+index.Counts.Blocked+index.Counts.Failed {
		return errors.New("index count total is contradictory")
	}
	expectedLeaseStatus := "within_window"
	if index.TerminalReference == "" {
		expectedLeaseStatus = "terminal_missing"
	} else if index.Lease.ElapsedMinutes < index.Lease.MinimumMinutes {
		expectedLeaseStatus = "minimum_not_met"
	} else if index.Lease.MaximumMinutes > 0 && index.Lease.ElapsedMinutes > index.Lease.MaximumMinutes {
		expectedLeaseStatus = "maximum_exceeded"
	}
	if index.Lease.Status != expectedLeaseStatus {
		return errors.New("index lease status is contradictory")
	}
	unsafe := index.Safety.ExecutesWork || index.Safety.ApprovesWork || index.Safety.MutatesRepositories ||
		index.Safety.CallsProviders || index.Safety.Publishes || index.Safety.Releases ||
		index.Safety.Deploys || index.Safety.AdvancesAuthority
	semanticReadiness := index.CompletionObserved && index.Lease.Status == "within_window" &&
		index.Counts.Ready == 0 && index.Counts.Blocked == 0 && index.Counts.Failed == 0 &&
		len(index.Conflicts) == 0 && !unsafe
	if index.ReadinessPassed != semanticReadiness {
		return errors.New("index readiness conclusion is contradictory")
	}
	expectedReturnGate := "final_response_denied"
	if index.FinalResponseAllowed {
		expectedReturnGate = "final_response_allowed"
	}
	if index.ReturnGateStatus != expectedReturnGate {
		return errors.New("index return gate is contradictory")
	}
	if index.FinalResponseAllowed && (!index.ReadinessPassed || index.Counts.Ready != 0 || index.Counts.Blocked != 0 || index.Counts.Failed != 0 || len(index.Conflicts) != 0) {
		return errors.New("index permits an unsafe final response")
	}
	return nil
}

func finalizeCanonicalTerminalIndex(index CanonicalTerminalIndex) CanonicalTerminalIndex {
	sort.Strings(index.Conflicts)
	index.Conflicts = compactStrings(index.Conflicts)
	index.ConflictSummaries = index.ConflictSummaries[:0]
	for _, conflict := range index.Conflicts {
		index.ConflictSummaries = append(index.ConflictSummaries, strings.ReplaceAll(conflict, "_", " "))
	}
	if len(index.ConflictSummaries) == 0 {
		index.ConflictSummaries = []string{"none"}
	}
	unsigned := index
	unsigned.Digest = ""
	data, _ := json.Marshal(unsigned)
	index.Digest = DigestBytes(data)
	return index
}

func observeTerminalArtifact(raw map[string]any) terminalObservation {
	counts := mapValue(raw, "counts")
	supervisor := mapValue(raw, "supervisor")
	safety := mapValue(raw, "safety_boundaries")
	return terminalObservation{
		schema:         firstString(raw, "schema", "contract_version"),
		missionID:      stringValue(raw, "mission_id"),
		completed:      firstInt(raw, counts, "completed", "completed_nodes"),
		ready:          firstInt(raw, counts, "ready", "ready_nodes"),
		blocked:        firstInt(raw, counts, "blocked", "blocked_nodes"),
		failed:         firstInt(raw, counts, "failed", "failed_nodes"),
		elapsed:        firstInt(raw, nil, "elapsed_minutes"),
		minimumNodes:   firstInt(raw, supervisor, "minimum_nodes", "min_nodes"),
		minimumMinutes: firstInt(raw, supervisor, "minimum_minutes", "min_minutes"),
		targetMinutes:  firstInt(raw, supervisor, "target_minutes"),
		maximumMinutes: firstInt(raw, supervisor, "maximum_minutes", "max_minutes"),
		leaseStatus:    stringValue(raw, "lease_time_status"),
		final:          boolValue(raw, "final_response_allowed"),
		nextAction:     stringValue(raw, "exact_next_action"),
		safety: CanonicalTerminalSafety{
			ExecutesWork:        boolValue(raw, "executes_work") || boolValue(safety, "executes_work"),
			ApprovesWork:        boolValue(raw, "approves_work") || boolValue(safety, "approves_work"),
			MutatesRepositories: boolValue(raw, "mutates_repositories") || boolValue(safety, "mutates_repositories"),
			CallsProviders:      boolValue(raw, "calls_providers") || boolValue(safety, "calls_providers"),
			Publishes:           boolValue(raw, "publishes") || boolValue(safety, "publishes"),
			Releases:            boolValue(raw, "releases") || boolValue(safety, "releases"),
			Deploys:             boolValue(raw, "deploys") || boolValue(safety, "deploys"),
			AdvancesAuthority:   boolValue(raw, "claims_authority_advance") || boolValue(safety, "advances_authority"),
		},
	}
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeJSONValue(decoder)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid JSON: trailing token %v", token)
		}
		return fmt.Errorf("invalid JSON: %w", err)
	}
	normalized, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(normalized, target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func decodeJSONValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch delimiter := token.(type) {
	case json.Delim:
		switch delimiter {
		case '{':
			object := map[string]any{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, errors.New("object key is not a string")
				}
				if _, exists := object[key]; exists {
					return nil, fmt.Errorf("duplicate JSON key %q", key)
				}
				value, err := decodeJSONValue(decoder)
				if err != nil {
					return nil, err
				}
				object[key] = value
			}
			_, err := decoder.Token()
			return object, err
		case '[':
			array := []any{}
			for decoder.More() {
				value, err := decodeJSONValue(decoder)
				if err != nil {
					return nil, err
				}
				array = append(array, value)
			}
			_, err := decoder.Token()
			return array, err
		}
	}
	return token, nil
}

func containedEvidencePath(root, relative string) (string, error) {
	if relative == "" || filepath.IsAbs(relative) || filepath.Clean(relative) != relative || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe evidence path %q", relative)
	}
	path := filepath.Join(root, relative)
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe evidence path %q", relative)
	}
	return path, nil
}

func readBoundedRegularFile(path string, limit int) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("evidence must be a regular file")
	}
	if info.Size() > int64(limit) {
		return nil, errors.New("evidence exceeds size limit")
	}
	return os.ReadFile(path)
}

func mapValue(raw map[string]any, key string) map[string]any {
	value, _ := raw[key].(map[string]any)
	return value
}

func stringValue(raw map[string]any, key string) string {
	value, _ := raw[key].(string)
	return value
}

func firstString(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringValue(raw, key); value != "" {
			return value
		}
	}
	return ""
}

func boolValue(raw map[string]any, key string) bool {
	value, _ := raw[key].(bool)
	return value
}

func firstInt(primary, secondary map[string]any, keys ...string) int {
	for _, source := range []map[string]any{primary, secondary} {
		for _, key := range keys {
			switch value := source[key].(type) {
			case json.Number:
				integer, _ := strconv.Atoi(value.String())
				return integer
			case float64:
				return int(value)
			}
		}
	}
	return 0
}

func artifactPathForRole(artifacts []CanonicalTerminalArtifact, role string) string {
	for _, artifact := range artifacts {
		if artifact.Role == role {
			return artifact.Path
		}
	}
	return ""
}

func isNoAction(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "" || value == "none" || value == "no further action" ||
		value == "fresh 60-node mission-to-atlas soak complete; no further execution is authorized."
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func validateArtifactState(role, state string) error {
	expected := map[string]string{
		"lease":      "lease_authority",
		"root":       "initial_snapshot",
		"checkpoint": "checkpoint",
		"duration":   "duration_snapshot",
		"terminal":   "terminal_candidate",
		"closure":    "closure_support",
	}
	want, ok := expected[role]
	if !ok {
		return fmt.Errorf("unsupported artifact role %q", role)
	}
	if state != want {
		return fmt.Errorf("%s artifact state must be %q", role, want)
	}
	return nil
}
