package atlas

import (
	"os"
	"strings"
)

type AtlasCompactionResumeRegressionOptions struct {
	NodeID                          string
	SourcePromptFixturePath         string
	SourcePromptMarkdownPath        string
	SourceReadbackPath              string
	ExpectedNextNodeAfterCompletion string
}

func BuildAtlasCompactionResumeRegression(options AtlasCompactionResumeRegressionOptions) (AtlasCompactionResumeRegression, error) {
	promptFixture, err := LoadJSON[AtlasCompactionResumePrompt](options.SourcePromptFixturePath)
	if err != nil {
		return AtlasCompactionResumeRegression{}, err
	}
	if err := ValidateAtlasCompactionResumePrompt(promptFixture); err != nil {
		return AtlasCompactionResumeRegression{}, err
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](options.SourceReadbackPath)
	if err != nil {
		return AtlasCompactionResumeRegression{}, err
	}
	promptBytes, err := os.ReadFile(options.SourcePromptMarkdownPath)
	if err != nil {
		return AtlasCompactionResumeRegression{}, err
	}
	prompt := string(promptBytes)
	regression := AtlasCompactionResumeRegression{
		Schema:                           AtlasCompactionResumeRegressionContract,
		NodeID:                           strings.TrimSpace(options.NodeID),
		Status:                           "guarded",
		SourcePromptFixturePath:          publicArtifactRef(options.SourcePromptFixturePath),
		SourcePromptFixtureDigest:        digestValue(promptFixture),
		SourcePromptMarkdownPath:         publicArtifactRef(options.SourcePromptMarkdownPath),
		SourcePromptMarkdownDigest:       digestBytes(promptBytes),
		SourceReadbackPath:               publicArtifactRef(options.SourceReadbackPath),
		SourceReadbackDigest:             digestValue(readback),
		SourcePromptFirstExecutableNode:  promptFixture.FirstExecutableNode,
		SourcePromptExactNextAction:      promptFixture.ExactNextAction,
		SourcePromptActiveNodePreserved:  strings.Contains(prompt, "Next executable node: `"+promptFixture.FirstExecutableNode+"`"),
		SourcePromptExactActionPreserved: strings.Contains(prompt, promptFixture.ExactNextAction),
		CompletedNodesBefore:             readback.CompletedNodes,
		TotalNodes:                       readback.TotalNodes,
		ReadyNodesBefore:                 readback.ReadyNodes,
		BlockedNodesBefore:               readback.BlockedNodes,
		FailedNodesBefore:                readback.FailedNodes,
		FirstExecutableNodeBefore:        readback.FirstExecutableNode,
		ExactNextActionBefore:            readback.ExactNextAction,
		ReturnGateStatusBefore:           readback.ReturnGateStatus,
		ContinuationContractReasonBefore: readback.ContinuationContract.Reason,
		EarlyReturnRiskStatusBefore:      readback.EarlyReturnRiskStatus,
		FinalResponseAllowedBefore:       readback.FinalResponseAllowed,
		RefusesFinalResponseBefore:       readback.ContinuationContract.RefusesFinalResponse,
		RegressionAssertions: []string{
			"first_executable_node_preserved",
			"exact_next_action_preserved",
			"return_gate_status_preserved",
			"final_response_denied_until_ready_work_consumed",
			"rsi_denial_preserved",
		},
		ExpectedNextNodeAfterCompletion: strings.TrimSpace(options.ExpectedNextNodeAfterCompletion),
		PromotionRequested:              false,
		PromotionGranted:                false,
		ClaimsAuthorityAdvance:          false,
		RSIRemainsDenied:                true,
	}
	if err := ValidateAtlasCompactionResumeRegression(regression); err != nil {
		return AtlasCompactionResumeRegression{}, err
	}
	return regression, nil
}

func ValidateAtlasCompactionResumeRegression(regression AtlasCompactionResumeRegression) error {
	var errs []string
	requireContract(&errs, "compaction_resume_regression", regression.Schema, AtlasCompactionResumeRegressionContract)
	if regression.Status != "guarded" {
		errs = append(errs, "status must be guarded")
	}
	requireField(&errs, "node_id", regression.NodeID)
	checkPublicPath(&errs, "node_id", regression.NodeID, true)
	requireField(&errs, "source_prompt_fixture_path", regression.SourcePromptFixturePath)
	checkPublicPath(&errs, "source_prompt_fixture_path", regression.SourcePromptFixturePath, true)
	checkOptionalDigest(&errs, "source_prompt_fixture_digest", regression.SourcePromptFixtureDigest)
	requireField(&errs, "source_prompt_markdown_path", regression.SourcePromptMarkdownPath)
	checkPublicPath(&errs, "source_prompt_markdown_path", regression.SourcePromptMarkdownPath, true)
	checkOptionalDigest(&errs, "source_prompt_markdown_digest", regression.SourcePromptMarkdownDigest)
	requireField(&errs, "source_readback_path", regression.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", regression.SourceReadbackPath, true)
	checkOptionalDigest(&errs, "source_readback_digest", regression.SourceReadbackDigest)
	if regression.SourcePromptFirstExecutableNode != "" {
		checkPublicPath(&errs, "source_prompt_first_executable_node", regression.SourcePromptFirstExecutableNode, true)
	}
	if regression.SourcePromptExactNextAction != "" {
		checkPublicStrings(&errs, "source_prompt_exact_next_action", []string{regression.SourcePromptExactNextAction}, true)
	}
	if regression.CompletedNodesBefore <= 0 || regression.TotalNodes < regression.CompletedNodesBefore || regression.ReadyNodesBefore < 0 || regression.BlockedNodesBefore < 0 || regression.FailedNodesBefore < 0 {
		errs = append(errs, "node counts must be positive and internally consistent")
	}
	requireField(&errs, "first_executable_node_before", regression.FirstExecutableNodeBefore)
	checkPublicPath(&errs, "first_executable_node_before", regression.FirstExecutableNodeBefore, true)
	requireField(&errs, "exact_next_action_before", regression.ExactNextActionBefore)
	checkPublicStrings(&errs, "exact_next_action_before", []string{regression.ExactNextActionBefore}, true)
	requireField(&errs, "return_gate_status_before", regression.ReturnGateStatusBefore)
	requireField(&errs, "continuation_contract_reason_before", regression.ContinuationContractReasonBefore)
	requireField(&errs, "early_return_risk_status_before", regression.EarlyReturnRiskStatusBefore)
	if regression.ReadyNodesBefore > 0 && regression.FinalResponseAllowedBefore {
		errs = append(errs, "final_response_allowed_before must be false while ready nodes remain")
	}
	if regression.ReadyNodesBefore > 0 && !regression.RefusesFinalResponseBefore {
		errs = append(errs, "refuses_final_response_before must be true while ready nodes remain")
	}
	for _, assertion := range []string{
		"first_executable_node_preserved",
		"exact_next_action_preserved",
		"return_gate_status_preserved",
		"final_response_denied_until_ready_work_consumed",
		"rsi_denial_preserved",
	} {
		if !containsStringValue(regression.RegressionAssertions, assertion) {
			errs = append(errs, "regression_assertions missing "+assertion)
		}
	}
	requireField(&errs, "expected_next_node_after_completion", regression.ExpectedNextNodeAfterCompletion)
	checkPublicPath(&errs, "expected_next_node_after_completion", regression.ExpectedNextNodeAfterCompletion, true)
	if regression.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if regression.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, false, false, false, regression.ClaimsAuthorityAdvance, regression.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasCompactionResumeRegression(path string, regression AtlasCompactionResumeRegression) error {
	return WriteJSON(path, regression)
}

func checkOptionalDigest(errs *[]string, field, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if !digestPattern.MatchString(value) {
		*errs = append(*errs, field+" must be sha256 digest")
	}
}

func digestBytes(data []byte) string {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return DigestBytes([]byte(text))
}
