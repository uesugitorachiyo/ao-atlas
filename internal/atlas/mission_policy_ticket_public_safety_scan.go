package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasPolicyTicketPublicSafetyScan(inputPath string) (AtlasPolicyTicketPublicSafetyScan, error) {
	input, err := LoadJSON[AtlasPolicyTicketPublicSafetyScanInput](inputPath)
	if err != nil {
		return AtlasPolicyTicketPublicSafetyScan{}, err
	}
	if err := ValidateAtlasPolicyTicketPublicSafetyScanInput(input); err != nil {
		return AtlasPolicyTicketPublicSafetyScan{}, err
	}
	scan := AtlasPolicyTicketPublicSafetyScan{
		Schema:                 AtlasPolicyTicketPublicSafetyScanContract,
		Status:                 "passed_policy_ticket_public_safety_scan",
		SourceInputPath:        publicArtifactRef(inputPath),
		SourceInputDigest:      digestValue(input),
		ClaimCount:             len(input.Claims),
		BlockedPhraseCount:     len(input.BlockedPhrases),
		Claims:                 append([]string(nil), input.Claims...),
		BlockedPhrases:         append([]string(nil), input.BlockedPhrases...),
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	scan.UnsafeMatches = findPolicyTicketUnsafeMatches(input.Claims, input.BlockedPhrases)
	scan.UnsafeClaimsFound = len(scan.UnsafeMatches)
	if scan.UnsafeClaimsFound != 0 {
		scan.Status = "failed_policy_ticket_public_safety_scan"
	}
	if err := ValidateAtlasPolicyTicketPublicSafetyScan(scan); err != nil {
		return AtlasPolicyTicketPublicSafetyScan{}, err
	}
	return scan, nil
}

func ValidateAtlasPolicyTicketPublicSafetyScanInput(input AtlasPolicyTicketPublicSafetyScanInput) error {
	var errs []string
	requireContract(&errs, "policy_ticket_public_safety_scan_input", input.Schema, AtlasPolicyTicketPublicSafetyScanInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validatePolicyTicketPublicSafetyText(&errs, "claims", input.Claims)
	validatePolicyTicketPublicSafetyText(&errs, "blocked_phrases", input.BlockedPhrases)
	matches := findPolicyTicketUnsafeMatches(input.Claims, input.BlockedPhrases)
	if len(matches) != 0 {
		errs = append(errs, fmt.Sprintf("claims contain blocked phrases: %s", strings.Join(matches, ", ")))
	}
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasPolicyTicketPublicSafetyScan(scan AtlasPolicyTicketPublicSafetyScan) error {
	var errs []string
	requireContract(&errs, "policy_ticket_public_safety_scan", scan.Schema, AtlasPolicyTicketPublicSafetyScanContract)
	if scan.Status != "passed_policy_ticket_public_safety_scan" {
		errs = append(errs, "status must be passed_policy_ticket_public_safety_scan")
	}
	requireField(&errs, "source_input_path", scan.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", scan.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", scan.SourceInputDigest)
	validatePolicyTicketPublicSafetyText(&errs, "claims", scan.Claims)
	validatePolicyTicketPublicSafetyText(&errs, "blocked_phrases", scan.BlockedPhrases)
	if scan.ClaimCount != len(scan.Claims) {
		errs = append(errs, "claim_count must match claims")
	}
	if scan.BlockedPhraseCount != len(scan.BlockedPhrases) {
		errs = append(errs, "blocked_phrase_count must match blocked_phrases")
	}
	matches := findPolicyTicketUnsafeMatches(scan.Claims, scan.BlockedPhrases)
	if len(matches) != scan.UnsafeClaimsFound {
		errs = append(errs, "unsafe_claims_found must match unsafe_matches")
	}
	if len(matches) != len(scan.UnsafeMatches) {
		errs = append(errs, "unsafe_matches must match blocked phrase scan")
	}
	if scan.UnsafeClaimsFound != 0 {
		errs = append(errs, "unsafe_claims_found must be zero")
	}
	validateNoAuthorityEffects(&errs, scan.SchedulesWork, scan.ExecutesWork, scan.ApprovesWork, scan.ClaimsAuthorityAdvance, scan.RSIRemainsDenied)
	return joinErrors(errs)
}

func validatePolicyTicketPublicSafetyText(errs *[]string, field string, values []string) {
	if len(values) == 0 {
		*errs = append(*errs, field+" must not be empty")
	}
	for i, value := range values {
		requireField(errs, fmt.Sprintf("%s[%d]", field, i), value)
	}
}

func findPolicyTicketUnsafeMatches(claims []string, blockedPhrases []string) []string {
	var matches []string
	for _, claim := range claims {
		normalizedClaim := strings.ToLower(claim)
		for _, phrase := range blockedPhrases {
			normalizedPhrase := strings.ToLower(strings.TrimSpace(phrase))
			if normalizedPhrase != "" && strings.Contains(normalizedClaim, normalizedPhrase) {
				matches = append(matches, phrase)
			}
		}
	}
	return matches
}
