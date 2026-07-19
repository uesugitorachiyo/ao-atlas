package atlas

import (
	"bytes"
	"crypto/sha256"
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
	evidenceCount := 0
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
		if entry.Schema == "ao.foundry.compact-evidence-manifest.v0.1" {
			compactEntries, err := verifiedCompactFoundryEvidenceEntries(key, resolvedPath, raw)
			if err != nil {
				return AtlasRunLinkDigestCheck{}, err
			}
			evidenceCount += len(compactEntries)
			schemaBoundCount += len(compactEntries)
			entries = append(entries, compactEntries...)
			continue
		}
		evidenceCount++
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
		EvidenceCount:            evidenceCount,
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

type compactFoundryManifest struct {
	Schema        string                `json:"schema"`
	FormatVersion string                `json:"format_version"`
	SourceRun     string                `json:"source_run"`
	ChunkOrder    []string              `json:"chunk_order"`
	TotalRecords  int                   `json:"total_record_count"`
	Chunks        []compactFoundryChunk `json:"chunks"`
	Lookup        struct {
		Strategy string                     `json:"strategy"`
		Ranges   []compactFoundryChunkRange `json:"ranges"`
	} `json:"lookup"`
	ManifestDigest string `json:"manifest_digest"`
}

type compactFoundryChunk struct {
	Path          string `json:"path"`
	RecordCount   int    `json:"record_count"`
	FirstRecordID string `json:"first_record_id"`
	LastRecordID  string `json:"last_record_id"`
	SHA256        string `json:"sha256"`
}

type compactFoundryChunkRange struct {
	Chunk         string `json:"chunk"`
	FirstRecordID string `json:"first_record_id"`
	LastRecordID  string `json:"last_record_id"`
}

type compactFoundryRecord struct {
	RecordID string          `json:"record_id"`
	Kind     string          `json:"kind"`
	Payload  json.RawMessage `json:"payload"`
}

func verifiedCompactFoundryEvidenceEntries(key, manifestPath string, raw map[string]any) ([]AtlasRunLinkDigestCheckEntry, error) {
	body, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var manifest compactFoundryManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, err
	}
	if manifest.Schema != "ao.foundry.compact-evidence-manifest.v0.1" {
		return nil, fmt.Errorf("compact manifest schema must be ao.foundry.compact-evidence-manifest.v0.1")
	}
	if manifest.FormatVersion != "v0.1" {
		return nil, fmt.Errorf("compact manifest format_version must be v0.1")
	}
	declaredOrder := make([]string, 0, len(manifest.Chunks))
	for i, chunk := range manifest.Chunks {
		if !safeCompactFoundryChunkPath(chunk.Path) {
			return nil, fmt.Errorf("chunks[%d].path must be a safe relative path", i)
		}
		declaredOrder = append(declaredOrder, chunk.Path)
	}
	if strings.Join(manifest.ChunkOrder, "\x00") != strings.Join(declaredOrder, "\x00") {
		return nil, fmt.Errorf("chunk_order must match manifest chunk order")
	}
	if manifest.Lookup.Strategy != "ordered_chunk_ranges" {
		return nil, fmt.Errorf("lookup.strategy must be ordered_chunk_ranges")
	}
	if len(manifest.Lookup.Ranges) != len(manifest.Chunks) {
		return nil, fmt.Errorf("lookup.ranges must match chunk count")
	}
	if manifest.ManifestDigest == "" {
		return nil, fmt.Errorf("manifest_digest is required")
	}
	if !compactFoundryManifestDigestMatches(manifest, raw) {
		return nil, fmt.Errorf("manifest_digest mismatch")
	}

	manifestDir := filepath.Dir(manifestPath)
	entries := []AtlasRunLinkDigestCheckEntry{}
	seen := map[string]bool{}
	previousRecordID := ""
	totalRecords := 0
	for i, chunk := range manifest.Chunks {
		body, err := os.ReadFile(filepath.Join(manifestDir, chunk.Path))
		if err != nil {
			return nil, fmt.Errorf("%s is missing", chunk.Path)
		}
		chunkSum := sha256.Sum256(body)
		if fmt.Sprintf("%x", chunkSum[:]) != chunk.SHA256 {
			return nil, fmt.Errorf("%s sha256 mismatch", chunk.Path)
		}
		records, err := parseCompactFoundryChunkRecords(chunk.Path, body)
		if err != nil {
			return nil, err
		}
		totalRecords += len(records)
		if chunk.RecordCount != len(records) {
			return nil, fmt.Errorf("%s record_count mismatch", chunk.Path)
		}
		firstRecordID, lastRecordID := "", ""
		if len(records) > 0 {
			firstRecordID = records[0].RecordID
			lastRecordID = records[len(records)-1].RecordID
		}
		if chunk.FirstRecordID != firstRecordID {
			return nil, fmt.Errorf("%s first_record_id mismatch", chunk.Path)
		}
		if chunk.LastRecordID != lastRecordID {
			return nil, fmt.Errorf("%s last_record_id mismatch", chunk.Path)
		}
		if manifest.Lookup.Ranges[i].Chunk != chunk.Path {
			return nil, fmt.Errorf("lookup.ranges[%d].chunk mismatch", i)
		}
		if manifest.Lookup.Ranges[i].FirstRecordID != chunk.FirstRecordID {
			return nil, fmt.Errorf("lookup.ranges[%d].first_record_id mismatch", i)
		}
		if manifest.Lookup.Ranges[i].LastRecordID != chunk.LastRecordID {
			return nil, fmt.Errorf("lookup.ranges[%d].last_record_id mismatch", i)
		}
		for _, record := range records {
			if seen[record.RecordID] {
				return nil, fmt.Errorf("duplicate record_id %s", record.RecordID)
			}
			if previousRecordID != "" && record.RecordID <= previousRecordID {
				return nil, fmt.Errorf("record_id %s is out of order", record.RecordID)
			}
			seen[record.RecordID] = true
			previousRecordID = record.RecordID
			entries = append(entries, AtlasRunLinkDigestCheckEntry{
				Key:    key + ":" + record.RecordID,
				Path:   publicArtifactRef(manifestPath) + "#" + record.RecordID,
				Schema: "ao.foundry.compact-evidence-record.v0.1",
				Status: "schema_bound",
			})
		}
	}
	if manifest.TotalRecords != totalRecords {
		return nil, fmt.Errorf("total_record_count mismatch")
	}
	return entries, nil
}

func compactFoundryManifestDigestMatches(manifest compactFoundryManifest, raw map[string]any) bool {
	manifestCopy := manifest
	manifestCopy.ManifestDigest = ""
	if body, err := json.Marshal(manifestCopy); err == nil {
		sum := sha256.Sum256(body)
		if fmt.Sprintf("%x", sum[:]) == manifest.ManifestDigest {
			return true
		}
	}
	rawCopy := map[string]any{}
	for key, value := range raw {
		rawCopy[key] = value
	}
	rawCopy["manifest_digest"] = ""
	if body, err := json.Marshal(rawCopy); err == nil {
		sum := sha256.Sum256(body)
		if fmt.Sprintf("%x", sum[:]) == manifest.ManifestDigest {
			return true
		}
	}
	return false
}

func parseCompactFoundryChunkRecords(path string, body []byte) ([]compactFoundryRecord, error) {
	records := []compactFoundryRecord{}
	for i, line := range bytes.Split(body, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var record compactFoundryRecord
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, fmt.Errorf("%s line %d malformed JSON", path, i+1)
		}
		if strings.TrimSpace(record.RecordID) == "" {
			return nil, fmt.Errorf("%s line %d record_id is required", path, i+1)
		}
		records = append(records, record)
	}
	return records, nil
}

func safeCompactFoundryChunkPath(path string) bool {
	if path == "" || filepath.IsAbs(path) || driveAbsPattern.MatchString(path) || strings.Contains(path, "\\") {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return false
		}
	}
	return true
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
