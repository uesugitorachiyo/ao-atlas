package atlas

import (
	"fmt"
	"sort"
	"strings"
)

func BuildAtlasFoundryTerminalStatusExamplesValidation(nodeID, sourceReadbackPath string) (AtlasFoundryTerminalStatusExamplesValidation, error) {
	nodeID = strings.TrimSpace(nodeID)
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	for name, value := range map[string]string{
		"node id":              nodeID,
		"source readback path": sourceReadbackPath,
	} {
		if value == "" {
			return AtlasFoundryTerminalStatusExamplesValidation{}, fmt.Errorf("%s is required", name)
		}
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasFoundryTerminalStatusExamplesValidation{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasFoundryTerminalStatusExamplesValidation{}, err
	}
	sourceDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasFoundryTerminalStatusExamplesValidation{}, err
	}

	keys := make([]string, 0, len(readback.FoundryTerminalStatusReadback))
	for key := range readback.FoundryTerminalStatusReadback {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	terminalByStatus := map[string]AtlasFoundryTerminalStatusExample{}
	for _, example := range readback.FoundryTerminalStatusExamples {
		terminalByStatus[example.SourceStatus] = example
	}
	deniedExamplesSafe := true
	for _, example := range readback.FoundryDeniedTerminalExamples {
		if !example.RSIRemainsDenied || example.AuthorityAdvanceClaimed || !example.RequiresExactMissingEvidence {
			deniedExamplesSafe = false
			break
		}
	}

	fixture := AtlasFoundryTerminalStatusExamplesValidation{
		Schema:                 AtlasFoundryTerminalStatusExamplesContract,
		NodeID:                 nodeID,
		Status:                 "terminal_status_examples_validated",
		SourceReadbackPath:     publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:   sourceDigest,
		TerminalStatusReadback: copyStringMap(readback.FoundryTerminalStatusReadback),
		TerminalStatusKeys:     keys,
		TerminalStatusKeyCount: len(keys),
		TerminalExamples:       append([]AtlasFoundryTerminalStatusExample(nil), readback.FoundryTerminalStatusExamples...),
		TerminalExampleCount:   len(readback.FoundryTerminalStatusExamples),
		DeniedExamples:         append([]AtlasFoundryDeniedTerminalExample(nil), readback.FoundryDeniedTerminalExamples...),
		DeniedExampleCount:     len(readback.FoundryDeniedTerminalExamples),
		ExamplesMatchReadbackEnums: terminalExamplesMatchReadbackEnums(
			readback.FoundryTerminalStatusReadback,
			terminalByStatus,
		),
		PromotedRequiresCommandPromoterAgreement: strings.Contains(readback.FoundryTerminalStatusReadback["promoted"], "promoter_and_command") &&
			strings.Contains(terminalByStatus["promoted"].RequiredReadback, "Promoter and Command"),
		DeniedRequiresExactEvidence: strings.Contains(readback.FoundryTerminalStatusReadback["denied"], "exact_missing_evidence") &&
			strings.Contains(terminalByStatus["denied"].RequiredReadback, "exact missing evidence"),
		BlockedRequiresRepairOrResume: strings.Contains(readback.FoundryTerminalStatusReadback["blocked"], "repair_or_checkpoint_resume") &&
			!terminalByStatus["blocked"].CanCloseMission &&
			strings.Contains(terminalByStatus["blocked"].RequiredReadback, "repair or resume"),
		DeniedExamplesSafe:     deniedExamplesSafe,
		ReadyNodes:             readback.ReadyNodes,
		FinalResponseAllowed:   readback.FinalResponseAllowed,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if err := ValidateAtlasFoundryTerminalStatusExamplesValidation(fixture); err != nil {
		return AtlasFoundryTerminalStatusExamplesValidation{}, err
	}
	return fixture, nil
}

func ValidateAtlasFoundryTerminalStatusExamplesValidation(fixture AtlasFoundryTerminalStatusExamplesValidation) error {
	var errs []string
	requireContract(&errs, "foundry_terminal_status_examples", fixture.Schema, AtlasFoundryTerminalStatusExamplesContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "terminal_status_examples_validated" {
		errs = append(errs, "status must be terminal_status_examples_validated")
	}
	requireField(&errs, "source_readback_path", fixture.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", fixture.SourceReadbackPath, true)
	if !digestPattern.MatchString(fixture.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	expectedReadback := map[string]string{
		"blocked":   "terminal_blocker_requires_repair_or_checkpoint_resume",
		"completed": "terminal_success_can_close_when_all_nodes_and_readbacks_are_complete",
		"denied":    "terminal_denial_requires_exact_missing_evidence_readback",
		"promoted":  "terminal_success_can_close_when_promoter_and_command_agree",
	}
	if fixture.TerminalStatusKeyCount != len(expectedReadback) || len(fixture.TerminalStatusKeys) != len(expectedReadback) {
		errs = append(errs, "terminal status key count must match expected statuses")
	}
	for _, key := range []string{"blocked", "completed", "denied", "promoted"} {
		if fixture.TerminalStatusReadback[key] != expectedReadback[key] {
			errs = append(errs, "terminal_status_readback."+key+" must match recommendation readback enum")
		}
		if !containsStringValue(fixture.TerminalStatusKeys, key) {
			errs = append(errs, "terminal_status_keys missing "+key)
		}
	}
	if fixture.TerminalExampleCount != len(fixture.TerminalExamples) || fixture.TerminalExampleCount != 4 {
		errs = append(errs, "terminal_example_count must be 4")
	}
	if fixture.DeniedExampleCount != len(fixture.DeniedExamples) || fixture.DeniedExampleCount != 3 {
		errs = append(errs, "denied_example_count must be 3")
	}
	if err := validateFoundryTerminalStatusExamples(fixture.TerminalExamples); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validateFoundryDeniedTerminalExamples(fixture.DeniedExamples); err != nil {
		errs = append(errs, err.Error())
	}
	if !fixture.ExamplesMatchReadbackEnums {
		errs = append(errs, "examples_match_readback_enums must be true")
	}
	if !fixture.PromotedRequiresCommandPromoterAgreement {
		errs = append(errs, "promoted_requires_command_promoter_agreement must be true")
	}
	if !fixture.DeniedRequiresExactEvidence {
		errs = append(errs, "denied_requires_exact_evidence must be true")
	}
	if !fixture.BlockedRequiresRepairOrResume {
		errs = append(errs, "blocked_requires_repair_or_resume must be true")
	}
	if !fixture.DeniedExamplesSafe {
		errs = append(errs, "denied_examples_safe must be true")
	}
	if fixture.ReadyNodes <= 0 {
		errs = append(errs, "ready_nodes must be positive for this replay source")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready nodes remain")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasFoundryTerminalStatusExamplesValidation(path string, fixture AtlasFoundryTerminalStatusExamplesValidation) error {
	return WriteJSON(path, fixture)
}

func terminalExamplesMatchReadbackEnums(readback map[string]string, terminalByStatus map[string]AtlasFoundryTerminalStatusExample) bool {
	expected := map[string]string{
		"blocked":   "blocked",
		"completed": "completed",
		"denied":    "denied",
		"promoted":  "completed",
	}
	for sourceStatus, normalized := range expected {
		if strings.TrimSpace(readback[sourceStatus]) == "" {
			return false
		}
		if terminalByStatus[sourceStatus].NormalizedStatus != normalized {
			return false
		}
	}
	return true
}
