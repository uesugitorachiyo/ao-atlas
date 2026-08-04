package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const aoMissionArtifactManifestV02MaxBytes = 1 << 20

type aoMissionArtifactManifestV02 struct {
	Schema         string                    `json:"schema"`
	MissionID      string                    `json:"mission_id"`
	ArtifactRefs   []aoMissionArtifactRefV02 `json:"artifact_refs"`
	ManifestDigest string                    `json:"manifest_digest"`
	Signature      string                    `json:"signature"`
	SafeToExecute  bool                      `json:"safe_to_execute"`
	ExecutesWork   bool                      `json:"executes_work"`
	ApprovesWork   bool                      `json:"approves_work"`
	GeneratedAtUTC string                    `json:"generated_at_utc,omitempty"`
}

type aoMissionArtifactRefV02 struct {
	Schema     string `json:"schema"`
	Ref        string `json:"ref"`
	ContentRef string `json:"content_ref"`
	Digest     string `json:"digest"`
	Kind       string `json:"kind,omitempty"`
}

func BuildAOMissionImport(recordPath, commandStatusPath, artifactManifestPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithRouteHistory(recordPath, commandStatusPath, artifactManifestPath, "")
}

func BuildAOMissionImportWithRouteHistory(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithMissionReadbacks(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, "", "")
}

func BuildAOMissionImportWithMissionReadbacks(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithMissionArchive(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, "")
}

func BuildAOMissionImportWithMissionArchive(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, missionArchivePath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithGatewayReadiness(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, missionArchivePath, "")
}

func BuildAOMissionImportWithGatewayReadiness(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, missionArchivePath, gatewayReadinessRollupPath string) (AOMissionImport, error) {
	return BuildAOMissionImportWithTimelineCompaction(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, "", missionArchivePath, gatewayReadinessRollupPath)
}

func BuildAOMissionImportWithTimelineCompaction(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, timelineCompactionPath, missionArchivePath, gatewayReadinessRollupPath string) (AOMissionImport, error) {
	var record map[string]any
	if err := readJSONIfPossible(recordPath, &record); err != nil {
		return AOMissionImport{}, err
	}
	var commandStatus map[string]any
	if err := readJSONIfPossible(commandStatusPath, &commandStatus); err != nil {
		return AOMissionImport{}, err
	}
	manifest, err := readAOMissionArtifactManifest(artifactManifestPath)
	if err != nil {
		return AOMissionImport{}, err
	}
	missionID, _ := record["mission_id"].(string)
	if missionID == "" {
		return AOMissionImport{}, fmt.Errorf("mission record requires mission_id")
	}
	if commandMissionID, _ := commandStatus["mission_id"].(string); commandMissionID != missionID {
		return AOMissionImport{}, fmt.Errorf("command status mission_id mismatch")
	}
	if manifestMissionID, _ := manifest["mission_id"].(string); manifestMissionID != missionID {
		return AOMissionImport{}, fmt.Errorf("artifact manifest mission_id mismatch")
	}
	for _, field := range []string{"safe_to_execute", "executes_work", "approves_work", "mutates_repositories"} {
		if value, ok := commandStatus[field].(bool); ok && value {
			return AOMissionImport{}, fmt.Errorf("command status %s must be false", field)
		}
	}
	for _, field := range []string{"executes_work", "approves_work"} {
		if value, ok := manifest[field].(bool); ok && value {
			return AOMissionImport{}, fmt.Errorf("artifact manifest %s must be false", field)
		}
	}
	if strings.TrimSpace(routeHistoryPath) != "" {
		if err := validateAOMissionRouteHistory(routeHistoryPath, missionID); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(schedulerRecoveryPath) != "" {
		if err := validateAOMissionReadback(schedulerRecoveryPath, missionID, "ao.mission.scheduler-recovery-readback.v0.1", "scheduler recovery"); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(ledgerCompactionPath) != "" {
		if err := validateAOMissionReadback(ledgerCompactionPath, missionID, "ao.mission.ledger-compaction-readback.v0.1", "ledger compaction"); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(timelineCompactionPath) != "" {
		if err := validateAOMissionReadback(timelineCompactionPath, missionID, "ao.mission.timeline-compaction-readback.v0.1", "timeline compaction"); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(missionArchivePath) != "" {
		if err := validateAOMissionArchive(missionArchivePath, missionID); err != nil {
			return AOMissionImport{}, err
		}
	}
	if strings.TrimSpace(gatewayReadinessRollupPath) != "" {
		if err := validateAOMissionReadback(gatewayReadinessRollupPath, missionID, "ao.mission.gateway-readiness-rollup.v0.1", "gateway readiness rollup"); err != nil {
			return AOMissionImport{}, err
		}
	}
	sources, err := aoMissionSourceArtifacts(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, timelineCompactionPath, missionArchivePath, gatewayReadinessRollupPath)
	if err != nil {
		return AOMissionImport{}, err
	}
	status, _ := record["status"].(string)
	route, _ := record["current_route"].(string)
	return AOMissionImport{
		ContractVersion: AOMissionImportContract,
		MissionID:       missionID,
		Status:          status,
		CurrentRoute:    route,
		SourceArtifacts: sources,
		NextAction:      "compile AO Mission context into Atlas workgraph before Foundry import",
		SafeToExecute:   false,
		SchedulesWork:   false,
		ExecutesWork:    false,
		ApprovesWork:    false,
	}, nil
}

func readAOMissionArtifactManifest(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	switch envelope.Schema {
	case "ao.mission.artifact-manifest.v0.1":
		var manifest map[string]any
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, err
		}
		if err := validateAOMissionManifestRefs(manifest, path); err != nil {
			return nil, err
		}
		return manifest, nil
	case "ao.mission.artifact-manifest.v0.2":
		return readAOMissionArtifactManifestV02(data, path)
	default:
		return nil, fmt.Errorf("artifact manifest schema must be ao.mission.artifact-manifest.v0.1 or ao.mission.artifact-manifest.v0.2")
	}
}

func readAOMissionArtifactManifestV02(data []byte, manifestPath string) (map[string]any, error) {
	if len(data) > aoMissionArtifactManifestV02MaxBytes {
		return nil, fmt.Errorf("artifact manifest exceeds %d-byte size limit", aoMissionArtifactManifestV02MaxBytes)
	}
	var document map[string]any
	if err := decodeStrictJSON(data, &document); err != nil {
		return nil, err
	}
	if err := validateAOMissionArtifactManifestV02Structure(document); err != nil {
		return nil, err
	}
	normalized, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	var manifest aoMissionArtifactManifestV02
	if err := json.Unmarshal(normalized, &manifest); err != nil {
		return nil, err
	}
	if err := validateAOMissionArtifactManifestV02(manifest, manifestPath); err != nil {
		return nil, err
	}
	return document, nil
}

func validateAOMissionArtifactManifestV02Structure(document map[string]any) error {
	if document == nil {
		return errors.New("artifact manifest v0.2 must be a JSON object")
	}
	if err := requireAOMissionObjectFields(document, "artifact manifest v0.2", map[string]string{
		"schema": "string", "mission_id": "string", "artifact_refs": "array", "manifest_digest": "string",
		"signature": "string", "safe_to_execute": "boolean", "executes_work": "boolean",
		"approves_work": "boolean", "generated_at_utc": "string",
	}, []string{"schema", "mission_id", "artifact_refs", "manifest_digest", "signature", "safe_to_execute", "executes_work", "approves_work"}); err != nil {
		return err
	}
	refs := document["artifact_refs"].([]any)
	for i, raw := range refs {
		ref, ok := raw.(map[string]any)
		if !ok || ref == nil {
			return fmt.Errorf("artifact manifest v0.2 artifact_refs[%d] must be an object", i)
		}
		if err := requireAOMissionObjectFields(ref, fmt.Sprintf("artifact manifest v0.2 artifact_refs[%d]", i), map[string]string{
			"schema": "string", "ref": "string", "content_ref": "string", "digest": "string", "kind": "string",
		}, []string{"schema", "ref", "content_ref", "digest"}); err != nil {
			return err
		}
	}
	return nil
}

func requireAOMissionObjectFields(object map[string]any, label string, allowed map[string]string, required []string) error {
	for key, value := range object {
		expected, ok := allowed[key]
		if !ok {
			return fmt.Errorf("%s has unknown field %q", label, key)
		}
		if !aOMissionValueHasType(value, expected) {
			return fmt.Errorf("%s field %q must be %s", label, key, expected)
		}
	}
	for _, key := range required {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("%s requires field %q", label, key)
		}
	}
	return nil
}

func aOMissionValueHasType(value any, expected string) bool {
	switch expected {
	case "string":
		_, ok := value.(string)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	default:
		return false
	}
}

func validateAOMissionArtifactManifestV02(manifest aoMissionArtifactManifestV02, manifestPath string) error {
	if manifest.Schema != "ao.mission.artifact-manifest.v0.2" {
		return fmt.Errorf("artifact manifest schema must be ao.mission.artifact-manifest.v0.2")
	}
	if manifest.SafeToExecute || manifest.ExecutesWork || manifest.ApprovesWork {
		return errors.New("artifact manifest must not claim execution or approval authority")
	}
	if !isCanonicalAOMissionSHA256(manifest.ManifestDigest) {
		return errors.New("artifact manifest digest must be a canonical sha256 digest")
	}
	expectedDigest, err := aOMissionArtifactManifestV02Digest(manifest)
	if err != nil {
		return err
	}
	if manifest.ManifestDigest != expectedDigest {
		return errors.New("artifact manifest digest mismatch")
	}
	if manifest.Signature != "ao-mission-local-digest:"+manifest.ManifestDigest {
		return errors.New("artifact manifest signature does not bind manifest digest")
	}
	for _, ref := range manifest.ArtifactRefs {
		if strings.TrimSpace(ref.Ref) == "" {
			return errors.New("artifact manifest refs require ref")
		}
		if ref.Schema != "ao.mission.artifact-ref.v0.1" {
			return errors.New("artifact manifest artifact ref schema must be ao.mission.artifact-ref.v0.1")
		}
		if !isCanonicalAOMissionSHA256(ref.Digest) {
			return fmt.Errorf("artifact manifest ref %q digest must be a canonical sha256 digest", ref.Ref)
		}
		expectedRef := "artifacts/sha256/" + strings.TrimPrefix(ref.Digest, "sha256:")
		if ref.ContentRef != expectedRef {
			return fmt.Errorf("artifact manifest ref %q content_ref must be contained and digest-addressed", ref.Ref)
		}
		data, err := readAOMissionV02RetainedContent(manifestPath, ref.ContentRef)
		if err != nil {
			return fmt.Errorf("artifact manifest ref %q: %w", ref.Ref, err)
		}
		if DigestBytes(data) != ref.Digest {
			return fmt.Errorf("artifact manifest ref %q digest mismatch", ref.Ref)
		}
	}
	return nil
}

func aOMissionArtifactManifestV02Digest(manifest aoMissionArtifactManifestV02) (string, error) {
	body, err := json.Marshal(struct {
		Schema       string                    `json:"schema"`
		MissionID    string                    `json:"mission_id"`
		ArtifactRefs []aoMissionArtifactRefV02 `json:"artifact_refs"`
	}{manifest.Schema, manifest.MissionID, manifest.ArtifactRefs})
	if err != nil {
		return "", err
	}
	return DigestBytes(body), nil
}

func isCanonicalAOMissionSHA256(digest string) bool {
	encoded := strings.TrimPrefix(digest, "sha256:")
	if encoded == digest || len(encoded) != sha256.Size*2 || encoded != strings.ToLower(encoded) {
		return false
	}
	_, err := hex.DecodeString(encoded)
	return err == nil
}

func readAOMissionV02RetainedContent(manifestPath, contentRef string) ([]byte, error) {
	root, err := os.OpenRoot(filepath.Dir(manifestPath))
	if err != nil {
		return nil, fmt.Errorf("open manifest root: %w", err)
	}
	defer root.Close()
	if err := validateAOMissionRetainedDirectories(root); err != nil {
		return nil, err
	}
	before, err := root.Lstat(contentRef)
	if err != nil {
		return nil, fmt.Errorf("inspect retained artifact: %w", err)
	}
	if !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 || before.Size() > aoMissionArtifactManifestV02MaxBytes {
		return nil, errors.New("retained artifact must be a bounded regular non-symlink file")
	}
	file, err := root.Open(contentRef)
	if err != nil {
		return nil, fmt.Errorf("open retained artifact: %w", err)
	}
	opened, statErr := file.Stat()
	if statErr != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat retained artifact: %w", statErr)
	}
	if !opened.Mode().IsRegular() || !os.SameFile(before, opened) || opened.Size() > aoMissionArtifactManifestV02MaxBytes {
		_ = file.Close()
		return nil, errors.New("retained artifact changed while opening")
	}
	body, readErr := io.ReadAll(io.LimitReader(file, aoMissionArtifactManifestV02MaxBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read retained artifact: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close retained artifact: %w", closeErr)
	}
	if len(body) > aoMissionArtifactManifestV02MaxBytes {
		return nil, errors.New("retained artifact exceeds size limit")
	}
	after, err := root.Lstat(contentRef)
	if err != nil {
		return nil, fmt.Errorf("reinspect retained artifact: %w", err)
	}
	if !after.Mode().IsRegular() || after.Mode()&os.ModeSymlink != 0 || !os.SameFile(opened, after) || after.Size() != int64(len(body)) {
		return nil, errors.New("retained artifact changed while reading")
	}
	if err := validateAOMissionRetainedDirectories(root); err != nil {
		return nil, err
	}
	return body, nil
}

func validateAOMissionRetainedDirectories(root *os.Root) error {
	for _, path := range []string{"artifacts", "artifacts/sha256"} {
		info, err := root.Lstat(path)
		if err != nil {
			return fmt.Errorf("inspect retained artifact directory: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("retained artifact directory must be a non-symlink directory")
		}
	}
	return nil
}

func BuildAOMissionWorkgraphMetadata(importPath, workgraphPath string) (AOMissionWorkgraphMetadata, error) {
	importRecord, err := LoadJSON[AOMissionImport](importPath)
	if err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	if importRecord.ContractVersion != AOMissionImportContract {
		return AOMissionWorkgraphMetadata{}, fmt.Errorf("invalid AO Mission import contract_version")
	}
	if importRecord.SafeToExecute || importRecord.SchedulesWork || importRecord.ExecutesWork || importRecord.ApprovesWork {
		return AOMissionWorkgraphMetadata{}, fmt.Errorf("AO Mission import must not claim execution, scheduling, or approval authority")
	}
	workgraph, err := LoadWorkgraph(workgraphPath)
	if err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	if workgraphRequiresOperationalBinding(workgraph) && importRecord.MissionID != workgraph.MissionID {
		return AOMissionWorkgraphMetadata{}, fmt.Errorf("AO Mission import mission_id must match operational workgraph")
	}
	importDigest, err := digestFile(importPath)
	if err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	workgraphDigest, err := digestFile(workgraphPath)
	if err != nil {
		return AOMissionWorkgraphMetadata{}, err
	}
	provenance := aoMissionProvenanceCounts(importRecord)
	return AOMissionWorkgraphMetadata{
		ContractVersion:          AOMissionWorkgraphMetadataContract,
		MissionID:                importRecord.MissionID,
		WorkgraphID:              workgraph.ID,
		TargetInstance:           workgraph.TargetInstance,
		CurrentRoute:             importRecord.CurrentRoute,
		NodeCounts:               aoMissionWorkgraphNodeCounts(workgraph),
		MissionProvenance:        provenance,
		ProvenanceNodes:          sortedMissionProvenanceKeys(provenance),
		PrimaryMissionProvenance: primaryMissionProvenance(provenance),
		ProvenanceDiagnostics:    missionProvenanceDiagnostics(provenance),
		SourceArtifacts: map[string]string{
			"ao_mission_import": importDigest,
			"workgraph":         workgraphDigest,
		},
		NextAction:    "send the first safe Atlas workgraph node to AO Foundry import",
		SafeToExecute: false,
		SchedulesWork: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	}, nil
}

func BuildAOMissionProvenanceWorkgraph(importRecord AOMissionImport, workgraph Workgraph) (Workgraph, error) {
	if importRecord.ContractVersion != AOMissionImportContract {
		return Workgraph{}, fmt.Errorf("invalid AO Mission import contract_version")
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return Workgraph{}, err
	}
	if workgraphRequiresOperationalBinding(workgraph) {
		return Workgraph{}, fmt.Errorf("operational workgraph provenance augmentation is denied because it would expand the approved planner partitions")
	}
	augmented := workgraph
	augmented.Nodes = append([]WorkgraphNode(nil), workgraph.Nodes...)
	existing := map[string]bool{}
	for _, node := range augmented.Nodes {
		existing[node.ID] = true
	}
	for _, source := range importRecord.SourceArtifacts {
		name := sanitizeMissionProvenanceNodeName(source.Name)
		if name == "" {
			continue
		}
		nodeID := "ao-mission-provenance-" + name
		if existing[nodeID] {
			continue
		}
		augmented.Nodes = append(augmented.Nodes, WorkgraphNode{
			ID:           nodeID,
			Status:       "blocked",
			Dependencies: []string{},
			Blockers:     []string{"AO Mission provenance node is readback-only and requires explicit downstream evidence binding before execution"},
			StitchTask:   false,
			FactoryTask: FactoryTask{
				ContractVersion:   FactoryTaskContract,
				ID:                nodeID + "-task",
				Objective:         "Bind AO Mission provenance source " + name + " as a first-class Atlas workgraph readback node.",
				TargetFactoryRepo: "ao-foundry",
				FactoryFolder:     "factory/ao-mission-provenance/" + name,
				Acceptance:        []string{"provenance source is digest-bound without granting execution authority"},
				NonGoals:          []string{"do not execute mutation", "do not approve AO Mission authority"},
				WriteScope:        []string{"factory/ao-mission-provenance/" + name},
				Verification:      []string{"scripts/production-readiness.sh"},
				RequiredEvidence:  []string{source.Path + "@" + source.SHA256},
				SafetyLimits:      []string{"readback-only", "no provider calls", "no credentials", "no public claim broadening"},
				AuthorityBoundary: "atlas_provenance_readback_only",
			},
		})
		existing[nodeID] = true
	}
	if err := ValidateWorkgraph(augmented); err != nil {
		return Workgraph{}, err
	}
	return augmented, nil
}

func BuildAOMissionProvenanceRender(metadata AOMissionWorkgraphMetadata) (AOMissionProvenanceRender, error) {
	if metadata.ContractVersion != AOMissionWorkgraphMetadataContract {
		return AOMissionProvenanceRender{}, fmt.Errorf("invalid AO Mission workgraph metadata contract_version")
	}
	if metadata.SafeToExecute || metadata.SchedulesWork || metadata.ExecutesWork || metadata.ApprovesWork {
		return AOMissionProvenanceRender{}, fmt.Errorf("AO Mission workgraph metadata must not claim execution, scheduling, or approval authority")
	}
	return AOMissionProvenanceRender{
		ContractVersion:          "ao.atlas.ao-mission-provenance-render.v0.1",
		Status:                   "ready",
		MissionID:                metadata.MissionID,
		WorkgraphID:              metadata.WorkgraphID,
		PrimaryMissionProvenance: metadata.PrimaryMissionProvenance,
		TotalProvenanceSources:   countMissionProvenanceSources(metadata.MissionProvenance),
		ProvenanceSummary:        metadata.ProvenanceDiagnostics,
		ProvenanceNodes:          append([]string(nil), metadata.ProvenanceNodes...),
		MissionProvenance:        copyStringIntMap(metadata.MissionProvenance),
		NextAction:               "render AO Mission provenance for operator review, then send safe workgraph nodes through Foundry gates",
		SafeToExecute:            false,
		SchedulesWork:            false,
		ExecutesWork:             false,
		ApprovesWork:             false,
	}, nil
}

func ValidateAOMissionWorkgraphMetadata(metadata AOMissionWorkgraphMetadata, workgraph Workgraph) error {
	if metadata.ContractVersion != AOMissionWorkgraphMetadataContract {
		return fmt.Errorf("invalid AO Mission workgraph metadata contract_version")
	}
	if strings.TrimSpace(metadata.MissionID) == "" {
		return fmt.Errorf("AO Mission workgraph metadata requires mission_id")
	}
	if workgraphRequiresOperationalBinding(workgraph) && metadata.MissionID != workgraph.MissionID {
		return fmt.Errorf("AO Mission workgraph metadata mission_id must match operational workgraph")
	}
	if metadata.WorkgraphID != workgraph.ID {
		return fmt.Errorf("AO Mission workgraph metadata workgraph_id must match workgraph")
	}
	if metadata.TargetInstance != workgraph.TargetInstance {
		return fmt.Errorf("AO Mission workgraph metadata target_instance must match workgraph")
	}
	if metadata.SafeToExecute || metadata.SchedulesWork || metadata.ExecutesWork || metadata.ApprovesWork {
		return fmt.Errorf("AO Mission workgraph metadata must not claim execution, scheduling, or approval authority")
	}
	if metadata.NodeCounts["total"] != len(workgraph.Nodes) {
		return fmt.Errorf("AO Mission workgraph metadata node_counts.total must match workgraph")
	}
	for _, node := range workgraph.Nodes {
		if metadata.NodeCounts[node.Status] == 0 {
			return fmt.Errorf("AO Mission workgraph metadata missing node_counts for status %q", node.Status)
		}
	}
	if len(metadata.SourceArtifacts) == 0 {
		return fmt.Errorf("AO Mission workgraph metadata requires source_artifacts")
	}
	if len(metadata.MissionProvenance) == 0 {
		return fmt.Errorf("AO Mission workgraph metadata requires mission_provenance")
	}
	return nil
}

func countMissionProvenanceSources(values map[string]int) int {
	total := 0
	for _, count := range values {
		total += count
	}
	return total
}

func copyStringIntMap(values map[string]int) map[string]int {
	out := make(map[string]int, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func aoMissionWorkgraphNodeCounts(workgraph Workgraph) map[string]int {
	counts := map[string]int{"total": len(workgraph.Nodes)}
	for _, node := range workgraph.Nodes {
		counts[node.Status]++
	}
	return counts
}

func sanitizeMissionProvenanceNodeName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, ".", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "\\", "-")
	fields := strings.Fields(value)
	value = strings.Join(fields, "-")
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "-")
}

func aoMissionProvenanceCounts(importRecord AOMissionImport) map[string]int {
	counts := map[string]int{}
	for _, source := range importRecord.SourceArtifacts {
		counts[source.Name]++
	}
	return counts
}

func primaryMissionProvenance(counts map[string]int) string {
	keys := sortedMissionProvenanceKeys(counts)
	if len(keys) == 0 {
		return ""
	}
	best := keys[0]
	for _, key := range keys[1:] {
		if counts[key] > counts[best] {
			best = key
		}
	}
	return best
}

func missionProvenanceDiagnostics(counts map[string]int) string {
	keys := sortedMissionProvenanceKeys(counts)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
	}
	return strings.Join(parts, ",")
}

func sortedMissionProvenanceKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func aoMissionSourceArtifacts(recordPath, commandStatusPath, artifactManifestPath, routeHistoryPath, schedulerRecoveryPath, ledgerCompactionPath, timelineCompactionPath, missionArchivePath, gatewayReadinessRollupPath string) ([]AOMissionSourceArtifact, error) {
	inputs := []struct {
		name string
		path string
	}{
		{name: "mission_record", path: recordPath},
		{name: "command_status", path: commandStatusPath},
		{name: "artifact_manifest", path: artifactManifestPath},
	}
	if strings.TrimSpace(routeHistoryPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "route_history", path: routeHistoryPath})
	}
	if strings.TrimSpace(schedulerRecoveryPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "scheduler_recovery", path: schedulerRecoveryPath})
	}
	if strings.TrimSpace(ledgerCompactionPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "ledger_compaction", path: ledgerCompactionPath})
	}
	if strings.TrimSpace(timelineCompactionPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "timeline_compaction", path: timelineCompactionPath})
	}
	if strings.TrimSpace(missionArchivePath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "mission_archive", path: missionArchivePath})
	}
	if strings.TrimSpace(gatewayReadinessRollupPath) != "" {
		inputs = append(inputs, struct {
			name string
			path string
		}{name: "gateway_readiness_rollup", path: gatewayReadinessRollupPath})
	}
	sources := make([]AOMissionSourceArtifact, 0, len(inputs))
	for _, input := range inputs {
		digest, err := digestFile(input.path)
		if err != nil {
			return nil, err
		}
		sources = append(sources, AOMissionSourceArtifact{Name: input.name, Path: filepath.ToSlash(input.path), SHA256: digest})
	}
	return sources, nil
}

func validateAOMissionArchive(path, missionID string) error {
	var archive map[string]any
	if err := readJSONIfPossible(path, &archive); err != nil {
		return err
	}
	if schema, _ := archive["schema"].(string); schema != "ao.mission.archive.v0.1" {
		return fmt.Errorf("mission archive schema must be ao.mission.archive.v0.1")
	}
	if archiveMissionID, _ := archive["mission_id"].(string); archiveMissionID != missionID {
		return fmt.Errorf("mission archive mission_id mismatch")
	}
	if strings.TrimSpace(fmt.Sprint(archive["archive_digest"])) == "" {
		return fmt.Errorf("mission archive requires archive_digest")
	}
	for _, field := range []string{"safe_to_execute", "executes_work", "approves_work", "mutates_repositories"} {
		if value, ok := archive[field].(bool); ok && value {
			return fmt.Errorf("mission archive %s must be false", field)
		}
	}
	return nil
}

func validateAOMissionRouteHistory(path, missionID string) error {
	var history []map[string]any
	if err := readJSONIfPossible(path, &history); err != nil {
		return err
	}
	if len(history) == 0 {
		return fmt.Errorf("route history requires at least one item")
	}
	for i, item := range history {
		if schema, _ := item["schema"].(string); schema != "ao.mission.route-decision.v0.1" {
			return fmt.Errorf("route history item %d schema must be ao.mission.route-decision.v0.1", i)
		}
		if got, _ := item["mission_id"].(string); got != missionID {
			return fmt.Errorf("route history item %d mission_id mismatch", i)
		}
		for _, field := range []string{"safe_to_execute", "executes_work", "approves_work", "mutates_repositories"} {
			if value, ok := item[field].(bool); ok && value {
				return fmt.Errorf("route history must not claim execution, approval, or repository mutation authority")
			}
		}
	}
	return nil
}

func validateAOMissionReadback(path, missionID, expectedSchema, label string) error {
	var readback map[string]any
	if err := readJSONIfPossible(path, &readback); err != nil {
		return err
	}
	if schema, _ := readback["schema"].(string); schema != expectedSchema {
		return fmt.Errorf("%s schema must be %s", label, expectedSchema)
	}
	if got, _ := readback["mission_id"].(string); got != missionID {
		return fmt.Errorf("%s mission_id mismatch", label)
	}
	for _, field := range []string{"safe_to_execute", "schedules_work", "executes_work", "approves_work", "mutates_repositories", "provider_calls", "release_or_publish", "credential_use", "direct_main_mutation", "concurrent_mutation"} {
		if value, ok := readback[field].(bool); ok && value {
			return fmt.Errorf("%s %s must be false", label, field)
		}
	}
	return nil
}

func validateAOMissionManifestRefs(manifest map[string]any, manifestPath string) error {
	refs, ok := manifest["artifact_refs"].([]any)
	if !ok {
		return nil
	}
	for i, raw := range refs {
		ref, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("artifact manifest artifact_refs[%d] must be an object", i)
		}
		path, _ := ref["ref"].(string)
		if path == "" {
			path, _ = ref["path"].(string)
		}
		want, _ := ref["digest"].(string)
		if want == "" {
			want, _ = ref["sha256"].(string)
		}
		if strings.TrimSpace(path) == "" || strings.TrimSpace(want) == "" {
			return fmt.Errorf("artifact manifest artifact_refs[%d] requires ref/path and digest/sha256", i)
		}
		if !strings.HasPrefix(want, "sha256:") {
			return fmt.Errorf("artifact manifest artifact_refs[%d] digest must start with sha256:", i)
		}
		actualPath, err := resolveAOMissionManifestRef(manifestPath, path)
		if err != nil {
			return fmt.Errorf("artifact manifest ref %q: %w", path, err)
		}
		got, err := digestFile(actualPath)
		if err != nil {
			return fmt.Errorf("artifact manifest ref %q: %w", path, err)
		}
		if got != want {
			return fmt.Errorf("artifact manifest ref %q digest mismatch", path)
		}
	}
	return nil
}

func resolveAOMissionManifestRef(manifestPath, ref string) (string, error) {
	if filepath.IsAbs(ref) {
		if _, err := os.Stat(ref); err != nil {
			return "", err
		}
		return ref, nil
	}
	if _, err := os.Stat(ref); err == nil {
		return ref, nil
	}
	candidate := filepath.Join(filepath.Dir(manifestPath), ref)
	if _, err := os.Stat(candidate); err != nil {
		return "", err
	}
	return candidate, nil
}
