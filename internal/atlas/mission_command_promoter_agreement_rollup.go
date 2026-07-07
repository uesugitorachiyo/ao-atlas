package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasCommandPromoterAgreementRollup(nodeID, promoterRollupPath, commandReadbackPath, sourceReadbackPath string) (AtlasCommandPromoterAgreementRollup, error) {
	nodeID = strings.TrimSpace(nodeID)
	promoterRollupPath = strings.TrimSpace(promoterRollupPath)
	commandReadbackPath = strings.TrimSpace(commandReadbackPath)
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	if nodeID == "" {
		return AtlasCommandPromoterAgreementRollup{}, fmt.Errorf("node id is required")
	}
	if promoterRollupPath == "" {
		return AtlasCommandPromoterAgreementRollup{}, fmt.Errorf("promoter rollup path is required")
	}
	if commandReadbackPath == "" {
		return AtlasCommandPromoterAgreementRollup{}, fmt.Errorf("command readback path is required")
	}
	if sourceReadbackPath == "" {
		return AtlasCommandPromoterAgreementRollup{}, fmt.Errorf("source readback path is required")
	}
	promoter, err := LoadJSON[AtlasPromoterNoPromotionRollup](promoterRollupPath)
	if err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	if err := ValidateAtlasPromoterNoPromotionRollup(promoter); err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	command, err := LoadJSON[AtlasNodeCommandReadbackEvidence](commandReadbackPath)
	if err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	if err := ValidateAtlasNodeCommandReadbackEvidence(command); err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	promoterDigest, err := digestTextFileWithNormalizedLineEndings(promoterRollupPath)
	if err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	commandDigest, err := digestTextFileWithNormalizedLineEndings(commandReadbackPath)
	if err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	readbackAgrees := command.ExpectedCompletedNodesAfter == readback.CompletedNodes &&
		command.ExpectedReadyNodesAfter == readback.ReadyNodes &&
		command.ExpectedNextExecutableNodeAfter == readback.FirstExecutableNode &&
		command.FinalResponseAllowed == readback.FinalResponseAllowed
	commandAgreesNoPromotion := command.Status == "readback_agrees_no_promotion" &&
		promoter.NoPromotionInvariantHolds &&
		!promoter.PromotionRequested &&
		!promoter.PromotionGranted &&
		!promoter.ClaimsAuthorityAdvance &&
		promoter.RSIRemainsDenied
	rollup := AtlasCommandPromoterAgreementRollup{
		Schema:                             AtlasCommandPromoterAgreementRollupContract,
		NodeID:                             nodeID,
		Status:                             "command_agrees_with_promoter_no_promotion",
		SourcePromoterRollupPath:           publicArtifactRef(promoterRollupPath),
		SourcePromoterRollupDigest:         promoterDigest,
		SourceCommandReadbackPath:          publicArtifactRef(commandReadbackPath),
		SourceCommandReadbackDigest:        commandDigest,
		SourceReadbackPath:                 publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:               readbackDigest,
		PromoterRollupStatus:               promoter.Status,
		PromoterNoPromotionFiles:           promoter.PromoterNoPromotionFiles,
		PromoterNoPromotionInvariantHolds:  promoter.NoPromotionInvariantHolds,
		CommandStatus:                      command.Status,
		CommandExpectedCompletedNodesAfter: command.ExpectedCompletedNodesAfter,
		CommandExpectedReadyNodesAfter:     command.ExpectedReadyNodesAfter,
		CommandExpectedNextExecutableNode:  command.ExpectedNextExecutableNodeAfter,
		ReadbackCompletedNodes:             readback.CompletedNodes,
		ReadbackReadyNodes:                 readback.ReadyNodes,
		ReadbackFirstExecutableNode:        readback.FirstExecutableNode,
		CommandAgreesNoPromotion:           commandAgreesNoPromotion,
		ReadbackAgreesWithCommand:          readbackAgrees,
		AggregatePromotionStatus:           promoter.AggregatePromotionStatus,
		PromotionRequested:                 promoter.PromotionRequested,
		PromotionGranted:                   promoter.PromotionGranted,
		FinalResponseAllowed:               readback.FinalResponseAllowed,
		SchedulesWork:                      false,
		ExecutesWork:                       false,
		ApprovesWork:                       false,
		ClaimsAuthorityAdvance:             promoter.ClaimsAuthorityAdvance,
		RSIRemainsDenied:                   promoter.RSIRemainsDenied && command.RSIRemainsDenied && readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if !rollup.CommandAgreesNoPromotion || !rollup.ReadbackAgreesWithCommand || !rollup.RSIRemainsDenied {
		rollup.Status = "command_promoter_agreement_failed"
	}
	if err := ValidateAtlasCommandPromoterAgreementRollup(rollup); err != nil {
		return AtlasCommandPromoterAgreementRollup{}, err
	}
	return rollup, nil
}

func ValidateAtlasCommandPromoterAgreementRollup(rollup AtlasCommandPromoterAgreementRollup) error {
	var errs []string
	requireContract(&errs, "command_promoter_agreement_rollup", rollup.Schema, AtlasCommandPromoterAgreementRollupContract)
	requireField(&errs, "node_id", rollup.NodeID)
	checkPublicPath(&errs, "node_id", rollup.NodeID, true)
	if !oneOf(rollup.Status, "command_agrees_with_promoter_no_promotion", "command_promoter_agreement_failed") {
		errs = append(errs, "status must be command_agrees_with_promoter_no_promotion or command_promoter_agreement_failed")
	}
	for field, value := range map[string]string{
		"source_promoter_rollup_path":    rollup.SourcePromoterRollupPath,
		"source_command_readback_path":   rollup.SourceCommandReadbackPath,
		"source_readback_path":           rollup.SourceReadbackPath,
		"command_expected_next_node":     rollup.CommandExpectedNextExecutableNode,
		"readback_first_executable_node": rollup.ReadbackFirstExecutableNode,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_promoter_rollup_digest":  rollup.SourcePromoterRollupDigest,
		"source_command_readback_digest": rollup.SourceCommandReadbackDigest,
		"source_readback_digest":         rollup.SourceReadbackDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if rollup.PromoterRollupStatus != "no_promotion_rollup_bound" {
		errs = append(errs, "promoter_rollup_status must be no_promotion_rollup_bound")
	}
	if rollup.PromoterNoPromotionFiles <= 0 {
		errs = append(errs, "promoter_no_promotion_files must be positive")
	}
	if !rollup.PromoterNoPromotionInvariantHolds {
		errs = append(errs, "promoter_no_promotion_invariant_holds must be true")
	}
	if rollup.CommandStatus != "readback_agrees_no_promotion" {
		errs = append(errs, "command_status must be readback_agrees_no_promotion")
	}
	if rollup.CommandExpectedCompletedNodesAfter != rollup.ReadbackCompletedNodes {
		errs = append(errs, "Command completed node expectation must match readback")
	}
	if rollup.CommandExpectedReadyNodesAfter != rollup.ReadbackReadyNodes {
		errs = append(errs, "Command ready node expectation must match readback")
	}
	if rollup.CommandExpectedNextExecutableNode != rollup.ReadbackFirstExecutableNode {
		errs = append(errs, "Command next executable expectation must match readback")
	}
	if !rollup.CommandAgreesNoPromotion {
		errs = append(errs, "command_agrees_no_promotion must be true")
	}
	if !rollup.ReadbackAgreesWithCommand {
		errs = append(errs, "readback_agrees_with_command must be true")
	}
	if rollup.AggregatePromotionStatus != "no_promotion_requested" {
		errs = append(errs, "aggregate_promotion_status must be no_promotion_requested")
	}
	if rollup.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if rollup.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if rollup.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false")
	}
	validateNoAuthorityEffects(&errs, rollup.SchedulesWork, rollup.ExecutesWork, rollup.ApprovesWork, rollup.ClaimsAuthorityAdvance, rollup.RSIRemainsDenied)
	if rollup.Status == "command_agrees_with_promoter_no_promotion" && (!rollup.CommandAgreesNoPromotion || !rollup.ReadbackAgreesWithCommand || !rollup.RSIRemainsDenied) {
		errs = append(errs, "agreement status requires Command, readback, and RSI agreement")
	}
	return joinErrors(errs)
}

func WriteAtlasCommandPromoterAgreementRollup(path string, rollup AtlasCommandPromoterAgreementRollup) error {
	return WriteJSON(path, rollup)
}
