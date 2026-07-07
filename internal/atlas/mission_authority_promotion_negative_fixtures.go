package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasAuthorityPromotionNegativeFixtures(nodeID string) (AtlasAuthorityPromotionNegativeFixtures, error) {
	cases := make([]AtlasAuthorityPromotionNegativeFixtureCase, 0, len(scopedPublicSafetyUnsafePatternSpecs))
	for _, spec := range scopedPublicSafetyUnsafePatternSpecs {
		cases = append(cases, AtlasAuthorityPromotionNegativeFixtureCase{
			ID:                    spec.ID,
			ScannerPatternID:      spec.ID,
			Category:              spec.Category,
			StatementTokens:       authorityPromotionNegativeTokens(spec.ID),
			ExpectedScanStatus:    "failed",
			ExpectedUnsafeMatches: 1,
			RedactionNote:         "unsafe wording is represented as separated tokens; no raw unsafe literal is stored in this fixture",
		})
	}
	fixture := AtlasAuthorityPromotionNegativeFixtures{
		Schema:                         AtlasAuthorityPromotionNegativeFixturesContract,
		NodeID:                         strings.TrimSpace(nodeID),
		Status:                         "passed",
		FixtureEncoding:                "redacted_token_sequences",
		ScannerContract:                AtlasScopedPublicSafetyScanContract,
		CaseCount:                      len(cases),
		Cases:                          cases,
		ForbiddenPatternsRedacted:      true,
		UnsafeLiteralStored:            false,
		ExpectedScanStatus:             "failed",
		ExpectedPublicSafetyScanPassed: false,
		SchedulesWork:                  false,
		ExecutesWork:                   false,
		ApprovesWork:                   false,
		ClaimsAuthorityAdvance:         false,
		RSIRemainsDenied:               true,
	}
	if err := ValidateAtlasAuthorityPromotionNegativeFixtures(fixture); err != nil {
		return AtlasAuthorityPromotionNegativeFixtures{}, err
	}
	return fixture, nil
}

func ValidateAtlasAuthorityPromotionNegativeFixtures(fixture AtlasAuthorityPromotionNegativeFixtures) error {
	var errs []string
	requireContract(&errs, "authority_promotion_negative_fixtures", fixture.Schema, AtlasAuthorityPromotionNegativeFixturesContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "passed" {
		errs = append(errs, "status must be passed")
	}
	if fixture.FixtureEncoding != "redacted_token_sequences" {
		errs = append(errs, "fixture_encoding must be redacted_token_sequences")
	}
	if fixture.ScannerContract != AtlasScopedPublicSafetyScanContract {
		errs = append(errs, "scanner_contract must reference scoped public-safety scan contract")
	}
	if fixture.CaseCount != len(fixture.Cases) {
		errs = append(errs, "case_count must match cases length")
	}
	if fixture.CaseCount < len(scopedPublicSafetyUnsafePatternSpecs) {
		errs = append(errs, "case_count must cover every scoped public-safety unsafe pattern")
	}
	if !fixture.ForbiddenPatternsRedacted {
		errs = append(errs, "forbidden_patterns_redacted must be true")
	}
	if fixture.UnsafeLiteralStored {
		errs = append(errs, "unsafe_literal_stored must be false")
	}
	if fixture.ExpectedScanStatus != "failed" {
		errs = append(errs, "expected_scan_status must be failed")
	}
	if fixture.ExpectedPublicSafetyScanPassed {
		errs = append(errs, "expected_public_safety_scan_passed must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)

	seen := map[string]bool{}
	expectedIDs := map[string]bool{}
	for _, spec := range scopedPublicSafetyUnsafePatternSpecs {
		expectedIDs[spec.ID] = true
	}
	for i, testCase := range fixture.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".id", testCase.ID)
		requireField(&errs, prefix+".scanner_pattern_id", testCase.ScannerPatternID)
		requireField(&errs, prefix+".category", testCase.Category)
		if seen[testCase.ScannerPatternID] {
			errs = append(errs, prefix+".scanner_pattern_id must be unique")
		}
		seen[testCase.ScannerPatternID] = true
		if !expectedIDs[testCase.ScannerPatternID] {
			errs = append(errs, prefix+".scanner_pattern_id is not a scoped public-safety pattern")
		}
		if len(testCase.StatementTokens) < 2 {
			errs = append(errs, prefix+".statement_tokens must contain redacted token sequence")
		}
		checkPublicStrings(&errs, prefix+".statement_tokens", testCase.StatementTokens, true)
		if testCase.ExpectedScanStatus != "failed" {
			errs = append(errs, prefix+".expected_scan_status must be failed")
		}
		if testCase.ExpectedUnsafeMatches <= 0 {
			errs = append(errs, prefix+".expected_unsafe_matches must be positive")
		}
		requireField(&errs, prefix+".redaction_note", testCase.RedactionNote)
	}
	for expectedID := range expectedIDs {
		if !seen[expectedID] {
			errs = append(errs, "cases must include "+expectedID)
		}
	}
	return joinErrors(errs)
}

func WriteAtlasAuthorityPromotionNegativeFixtures(path string, fixture AtlasAuthorityPromotionNegativeFixtures) error {
	return WriteJSON(path, fixture)
}

func authorityPromotionNegativeTokens(patternID string) []string {
	switch patternID {
	case "promotion_granted_true":
		return []string{"json_field", "promotion_granted", "boolean_true"}
	case "promotion_claimed_true":
		return []string{"json_field", "promotion_claimed", "boolean_true"}
	case "claims_authority_advance_true":
		return []string{"json_field", "claims_authority_advance", "boolean_true"}
	case "fully_unsupervised_complex_mutation_live_proven_true":
		return []string{"json_field", "fully_unsupervised_complex_mutation_live_proven", "boolean_true"}
	case "rsi_is_proven_phrase":
		return []string{"RSI", "is", "proven"}
	case "rsi_proof_granted_phrase":
		return []string{"RSI", "proof", "granted"}
	case "fully_unsupervised_complex_mutation_is_live_proven_phrase":
		return []string{"fully_unsupervised_complex_mutation", "is", "live-proven"}
	default:
		return []string{"unknown", patternID}
	}
}
