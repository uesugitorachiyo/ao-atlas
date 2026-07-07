package atlas

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func BuildAtlasRunLinkDigestCheck(nodeID, runLinkPath, evidenceRoot string) (AtlasRunLinkDigestCheck, error) {
	nodeID = strings.TrimSpace(nodeID)
	runLinkPath = strings.TrimSpace(runLinkPath)
	evidenceRoot = strings.TrimSpace(evidenceRoot)
	for name, value := range map[string]string{
		"node id":       nodeID,
		"run-link path": runLinkPath,
		"evidence root": evidenceRoot,
	} {
		if value == "" {
			return AtlasRunLinkDigestCheck{}, fmt.Errorf("%s is required", name)
		}
	}

	link, err := LoadJSON[RunLink](runLinkPath)
	if err != nil {
		return AtlasRunLinkDigestCheck{}, err
	}
	if err := ValidateRunLink(link); err != nil {
		return AtlasRunLinkDigestCheck{}, err
	}
	sourceDigest, err := digestTextFileWithNormalizedLineEndings(runLinkPath)
	if err != nil {
		return AtlasRunLinkDigestCheck{}, err
	}
	recomputedDigest := digestRunLink(link)

	keys := make([]string, 0, len(link.Evidence))
	for key := range link.Evidence {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]AtlasRunLinkDigestCheckEntry, 0, len(keys))
	missingEvidence := []string{}
	schemaBoundCount := 0
	for _, key := range keys {
		evidencePath := strings.TrimSpace(link.Evidence[key])
		resolvedPath := resolveRunLinkEvidencePath(evidenceRoot, evidencePath)
		entry := AtlasRunLinkDigestCheckEntry{
			Key:    key,
			Path:   publicArtifactRef(resolvedPath),
			Status: "missing",
		}
		data, err := os.ReadFile(resolvedPath)
		if err != nil {
			missingEvidence = append(missingEvidence, evidencePath)
			entries = append(entries, entry)
			continue
		}
		raw := map[string]any{}
		if err := json.Unmarshal(data, &raw); err != nil {
			entry.Status = "invalid_json"
			missingEvidence = append(missingEvidence, evidencePath)
			entries = append(entries, entry)
			continue
		}
		entry.Schema = evidenceSchemaMarker(raw)
		if entry.Schema == "" {
			entry.Status = "missing_schema"
			missingEvidence = append(missingEvidence, evidencePath)
			entries = append(entries, entry)
			continue
		}
		entry.Status = "schema_bound"
		schemaBoundCount++
		entries = append(entries, entry)
	}

	check := AtlasRunLinkDigestCheck{
		Schema:                   AtlasRunLinkDigestCheckContract,
		NodeID:                   nodeID,
		Status:                   "run_link_digest_verified",
		SourceRunLinkPath:        publicArtifactRef(runLinkPath),
		SourceRunLinkFileDigest:  sourceDigest,
		TaskID:                   link.TaskID,
		RunLinkStatus:            link.Status,
		RecordedDigest:           link.Digest,
		RecomputedDigest:         recomputedDigest,
		DigestMatches:            link.Digest == recomputedDigest,
		EvidenceRoot:             publicArtifactRef(evidenceRoot),
		EvidenceCount:            len(link.Evidence),
		SchemaBoundEvidenceCount: schemaBoundCount,
		MissingEvidence:          missingEvidence,
		EvidenceEntries:          entries,
		SchedulesWork:            false,
		ExecutesWork:             false,
		ApprovesWork:             false,
		ClaimsAuthorityAdvance:   false,
		RSIRemainsDenied:         true,
	}
	if check.RunLinkStatus != "completed" ||
		!check.DigestMatches ||
		len(check.MissingEvidence) != 0 ||
		check.SchemaBoundEvidenceCount != check.EvidenceCount {
		check.Status = "run_link_digest_failed"
	}
	if err := ValidateAtlasRunLinkDigestCheck(check); err != nil {
		return AtlasRunLinkDigestCheck{}, err
	}
	return check, nil
}

func resolveRunLinkEvidencePath(evidenceRoot, evidencePath string) string {
	if filepath.IsAbs(evidencePath) || driveAbsPattern.MatchString(evidencePath) {
		return filepath.Clean(evidencePath)
	}
	return filepath.Clean(filepath.Join(evidenceRoot, evidencePath))
}

func ValidateAtlasRunLinkDigestCheck(check AtlasRunLinkDigestCheck) error {
	var errs []string
	requireContract(&errs, "run_link_digest_check", check.Schema, AtlasRunLinkDigestCheckContract)
	requireField(&errs, "node_id", check.NodeID)
	checkPublicPath(&errs, "node_id", check.NodeID, true)
	if check.Status != "run_link_digest_verified" {
		errs = append(errs, "status must be run_link_digest_verified")
	}
	for field, value := range map[string]string{
		"source_run_link_path":        check.SourceRunLinkPath,
		"task_id":                     check.TaskID,
		"run_link_status":             check.RunLinkStatus,
		"recorded_digest":             check.RecordedDigest,
		"recomputed_digest":           check.RecomputedDigest,
		"source_run_link_file_digest": check.SourceRunLinkFileDigest,
		"evidence_root":               check.EvidenceRoot,
	} {
		requireField(&errs, field, value)
		if strings.Contains(field, "path") || field == "evidence_root" {
			checkPublicPath(&errs, field, value, true)
		}
	}
	for field, value := range map[string]string{
		"source_run_link_file_digest": check.SourceRunLinkFileDigest,
		"recorded_digest":             check.RecordedDigest,
		"recomputed_digest":           check.RecomputedDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if check.RunLinkStatus != "completed" {
		errs = append(errs, "run_link_status must be completed")
	}
	if !strings.HasSuffix(check.TaskID, "-task") {
		errs = append(errs, "task_id must be a task id")
	}
	if !check.DigestMatches {
		errs = append(errs, "digest_matches must be true")
	}
	if check.RecordedDigest != check.RecomputedDigest {
		errs = append(errs, "recorded_digest must match recomputed_digest")
	}
	if check.EvidenceCount <= 0 {
		errs = append(errs, "evidence_count must be greater than zero")
	}
	if check.SchemaBoundEvidenceCount != check.EvidenceCount {
		errs = append(errs, "schema_bound_evidence_count must match evidence_count")
	}
	if len(check.MissingEvidence) != 0 {
		errs = append(errs, "missing_evidence must be empty")
	}
	if len(check.EvidenceEntries) != check.EvidenceCount {
		errs = append(errs, "evidence_entries length must match evidence_count")
	}
	previousKey := ""
	seenKeys := map[string]bool{}
	for i, entry := range check.EvidenceEntries {
		prefix := fmt.Sprintf("evidence_entries[%d]", i)
		requireField(&errs, prefix+".key", entry.Key)
		checkPublicPath(&errs, prefix+".key", entry.Key, true)
		requireField(&errs, prefix+".path", entry.Path)
		checkPublicPath(&errs, prefix+".path", entry.Path, true)
		requireField(&errs, prefix+".schema", entry.Schema)
		if entry.Status != "schema_bound" {
			errs = append(errs, prefix+".status must be schema_bound")
		}
		if seenKeys[entry.Key] {
			errs = append(errs, "evidence_entries keys must be unique")
		}
		seenKeys[entry.Key] = true
		if previousKey != "" && entry.Key < previousKey {
			errs = append(errs, "evidence_entries must be sorted by key")
		}
		previousKey = entry.Key
	}
	validateNoAuthorityEffects(&errs, check.SchedulesWork, check.ExecutesWork, check.ApprovesWork, check.ClaimsAuthorityAdvance, check.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasRunLinkDigestCheck(path string, check AtlasRunLinkDigestCheck) error {
	return WriteJSON(path, check)
}
