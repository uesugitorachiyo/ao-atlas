package atlas

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

func BuildAtlasRunLinkSchemaCoverage(evidenceRoot string) (AtlasRunLinkSchemaCoverage, error) {
	root := filepath.Clean(evidenceRoot)
	entries := []AtlasRunLinkSchemaCoverageItem{}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Base(path) != "run-link.json" {
			return nil
		}
		link, err := LoadJSON[RunLink](path)
		if err != nil {
			return err
		}
		if err := ValidateRunLink(link); err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		entries = append(entries, AtlasRunLinkSchemaCoverageItem{
			Path:             filepath.ToSlash(rel),
			TaskID:           link.TaskID,
			Status:           link.Status,
			Schema:           link.ContractVersion,
			Validator:        "typed:run-link",
			EvidenceKeyCount: len(link.Evidence),
			Digest:           digestValue(link),
		})
		return nil
	}); err != nil {
		return AtlasRunLinkSchemaCoverage{}, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	schemaCounts := map[string]int{}
	validatorCounts := map[string]int{}
	completed := 0
	for _, entry := range entries {
		schemaCounts[entry.Schema]++
		validatorCounts[entry.Validator]++
		if entry.Status == "completed" {
			completed++
		}
	}
	coverage := AtlasRunLinkSchemaCoverage{
		Schema:                 AtlasRunLinkSchemaCoverageContract,
		Status:                 "complete",
		EvidenceRoot:           publicArtifactRef(evidenceRoot),
		RunLinkCount:           len(entries),
		CompletedRunLinks:      completed,
		SchemaCounts:           schemaCounts,
		ValidatorCounts:        validatorCounts,
		Entries:                entries,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasRunLinkSchemaCoverage(coverage); err != nil {
		return AtlasRunLinkSchemaCoverage{}, err
	}
	return coverage, nil
}

func ValidateAtlasRunLinkSchemaCoverage(coverage AtlasRunLinkSchemaCoverage) error {
	var errs []string
	requireContract(&errs, "run_link_schema_coverage", coverage.Schema, AtlasRunLinkSchemaCoverageContract)
	if coverage.Status != "complete" {
		errs = append(errs, "status must be complete")
	}
	requireField(&errs, "evidence_root", coverage.EvidenceRoot)
	checkPublicPath(&errs, "evidence_root", coverage.EvidenceRoot, true)
	if coverage.RunLinkCount != len(coverage.Entries) {
		errs = append(errs, "run_link_count must match entries length")
	}
	if coverage.CompletedRunLinks != coverage.RunLinkCount {
		errs = append(errs, "completed_run_links must match run_link_count")
	}
	if coverage.SchemaCounts[RunLinkContract] != coverage.RunLinkCount {
		errs = append(errs, "schema_counts run-link count must match run_link_count")
	}
	if coverage.ValidatorCounts["typed:run-link"] != coverage.RunLinkCount {
		errs = append(errs, "validator_counts typed:run-link count must match run_link_count")
	}
	previousPath := ""
	seenPaths := map[string]bool{}
	for i, entry := range coverage.Entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		requireField(&errs, prefix+".path", entry.Path)
		checkPublicPath(&errs, prefix+".path", entry.Path, true)
		if seenPaths[entry.Path] {
			errs = append(errs, "entries paths must be unique")
		}
		seenPaths[entry.Path] = true
		if previousPath != "" && entry.Path < previousPath {
			errs = append(errs, "entries must be sorted by path")
		}
		previousPath = entry.Path
		requireField(&errs, prefix+".task_id", entry.TaskID)
		checkPublicPath(&errs, prefix+".task_id", entry.TaskID, true)
		if entry.Status != "completed" {
			errs = append(errs, prefix+".status must be completed")
		}
		if entry.Schema != RunLinkContract {
			errs = append(errs, prefix+".schema must be "+RunLinkContract)
		}
		if entry.Validator != "typed:run-link" {
			errs = append(errs, prefix+".validator must be typed:run-link")
		}
		if entry.EvidenceKeyCount <= 0 {
			errs = append(errs, prefix+".evidence_key_count must be greater than zero")
		}
		if !digestPattern.MatchString(entry.Digest) {
			errs = append(errs, prefix+".digest must be sha256 digest")
		}
	}
	if coverage.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if coverage.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if coverage.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if coverage.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !coverage.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}
