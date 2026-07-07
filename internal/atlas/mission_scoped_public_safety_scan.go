package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type scopedPublicSafetyUnsafePatternSpec struct {
	ID       string
	Category string
	Pattern  *regexp.Regexp
}

var scopedPublicSafetyUnsafePatternSpecs = []scopedPublicSafetyUnsafePatternSpec{
	{ID: "promotion_granted_true", Category: "json_authority_boolean", Pattern: regexp.MustCompile(`"promotion_granted"\s*:\s*true`)},
	{ID: "promotion_claimed_true", Category: "json_promotion_boolean", Pattern: regexp.MustCompile(`"promotion_claimed"\s*:\s*true`)},
	{ID: "claims_authority_advance_true", Category: "json_authority_boolean", Pattern: regexp.MustCompile(`"claims_authority_advance"\s*:\s*true`)},
	{ID: "fully_unsupervised_complex_mutation_live_proven_true", Category: "json_promotion_boolean", Pattern: regexp.MustCompile(`"fully_unsupervised_complex_mutation_live_proven"\s*:\s*true`)},
	{ID: "rsi_is_proven_phrase", Category: "text_rsi_claim", Pattern: regexp.MustCompile(`(?i)\brsi\s+is\s+(proven|live|promoted)\b`)},
	{ID: "rsi_proof_granted_phrase", Category: "text_rsi_claim", Pattern: regexp.MustCompile(`(?i)\brsi proof granted\b`)},
	{ID: "fully_unsupervised_complex_mutation_is_live_proven_phrase", Category: "text_promotion_claim", Pattern: regexp.MustCompile(`(?i)\bfully_unsupervised_complex_mutation\s+is\s+(proven|live-proven|live)\b`)},
}

func BuildAtlasScopedPublicSafetyScan(nodeID string, scopes []string) (AtlasScopedPublicSafetyScan, error) {
	if len(scopes) == 0 {
		return AtlasScopedPublicSafetyScan{}, fmt.Errorf("at least one scope is required")
	}
	scannedScopes := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		scannedScopes = append(scannedScopes, publicArtifactRef(scope))
	}
	if len(scannedScopes) == 0 {
		return AtlasScopedPublicSafetyScan{}, fmt.Errorf("at least one non-empty scope is required")
	}
	files, err := scopedPublicSafetyScanFiles(scopes)
	if err != nil {
		return AtlasScopedPublicSafetyScan{}, err
	}
	unsafeMatches := 0
	evidenceFiles := 0
	promptArtifacts := 0
	scannedFiles := make([]string, 0, len(files))
	for _, path := range files {
		ref := publicArtifactRef(path)
		scannedFiles = append(scannedFiles, ref)
		switch {
		case strings.HasSuffix(ref, ".json"):
			evidenceFiles++
		case strings.HasSuffix(ref, ".md"):
			promptArtifacts++
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return AtlasScopedPublicSafetyScan{}, err
		}
		text := string(data)
		for _, spec := range scopedPublicSafetyUnsafePatternSpecs {
			unsafeMatches += len(spec.Pattern.FindAllStringIndex(text, -1))
		}
	}
	sort.Strings(scannedScopes)
	sort.Strings(scannedFiles)
	status := "passed"
	if unsafeMatches > 0 {
		status = "failed"
	}
	scan := AtlasScopedPublicSafetyScan{
		Schema:                    AtlasScopedPublicSafetyScanContract,
		NodeID:                    strings.TrimSpace(nodeID),
		Status:                    status,
		ScannedScopes:             scannedScopes,
		ScannedFiles:              scannedFiles,
		ScannedFileCount:          len(scannedFiles),
		ChangedEvidenceFiles:      evidenceFiles,
		ChangedPromptArtifacts:    promptArtifacts,
		ForbiddenPatternsRedacted: true,
		UnsafeMatchCount:          unsafeMatches,
		PublicSafetyScanPassed:    unsafeMatches == 0,
		SchedulesWork:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		ClaimsAuthorityAdvance:    false,
		RSIRemainsDenied:          true,
	}
	if err := ValidateAtlasScopedPublicSafetyScan(scan); err != nil {
		return AtlasScopedPublicSafetyScan{}, err
	}
	if unsafeMatches > 0 {
		return scan, fmt.Errorf("scoped public-safety scan found %d unsafe matches", unsafeMatches)
	}
	return scan, nil
}

func ValidateAtlasScopedPublicSafetyScan(scan AtlasScopedPublicSafetyScan) error {
	var errs []string
	requireContract(&errs, "scoped_public_safety_scan", scan.Schema, AtlasScopedPublicSafetyScanContract)
	requireField(&errs, "node_id", scan.NodeID)
	checkPublicPath(&errs, "node_id", scan.NodeID, true)
	if !oneOf(scan.Status, "passed", "failed") {
		errs = append(errs, "status must be passed or failed")
	}
	if len(scan.ScannedScopes) == 0 {
		errs = append(errs, "scanned_scopes must not be empty")
	}
	checkPublicStrings(&errs, "scanned_scopes", scan.ScannedScopes, true)
	if len(scan.ScannedFiles) == 0 {
		errs = append(errs, "scanned_files must not be empty")
	}
	checkPublicStrings(&errs, "scanned_files", scan.ScannedFiles, true)
	if scan.ScannedFileCount != len(scan.ScannedFiles) {
		errs = append(errs, "scanned_file_count must match scanned_files length")
	}
	if scan.ChangedEvidenceFiles <= 0 {
		errs = append(errs, "changed_evidence_files must be positive")
	}
	if scan.ChangedPromptArtifacts <= 0 {
		errs = append(errs, "changed_prompt_artifacts must be positive")
	}
	if !scan.ForbiddenPatternsRedacted {
		errs = append(errs, "forbidden_patterns_redacted must be true")
	}
	if scan.UnsafeMatchCount < 0 {
		errs = append(errs, "unsafe_match_count must be non-negative")
	}
	if scan.Status == "passed" && scan.UnsafeMatchCount != 0 {
		errs = append(errs, "passed scan must have zero unsafe matches")
	}
	if scan.Status == "failed" && scan.UnsafeMatchCount == 0 {
		errs = append(errs, "failed scan must have unsafe matches")
	}
	if scan.PublicSafetyScanPassed != (scan.Status == "passed" && scan.UnsafeMatchCount == 0) {
		errs = append(errs, "public_safety_scan_passed must match status and unsafe_match_count")
	}
	validateNoAuthorityEffects(&errs, scan.SchedulesWork, scan.ExecutesWork, scan.ApprovesWork, scan.ClaimsAuthorityAdvance, scan.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasScopedPublicSafetyScan(path string, scan AtlasScopedPublicSafetyScan) error {
	return WriteJSON(path, scan)
}

func scopedPublicSafetyScanFiles(scopes []string) ([]string, error) {
	seen := map[string]bool{}
	files := []string{}
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		info, err := os.Stat(scope)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if scopedPublicSafetyFile(scope) {
				clean := filepath.Clean(scope)
				seen[clean] = true
				files = append(files, clean)
			}
			continue
		}
		if err := filepath.WalkDir(scope, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !scopedPublicSafetyFile(path) {
				return nil
			}
			clean := filepath.Clean(path)
			if seen[clean] {
				return nil
			}
			seen[clean] = true
			files = append(files, clean)
			return nil
		}); err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("scoped public-safety scan found no json or markdown artifacts")
	}
	return files, nil
}

func scopedPublicSafetyFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".md":
		return true
	default:
		return false
	}
}
