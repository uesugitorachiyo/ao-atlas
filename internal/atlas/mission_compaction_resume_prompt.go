package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AtlasCompactionResumePromptOptions struct {
	NodeID                          string
	SourceReadbackPath              string
	PromptPath                      string
	LeaseStartPath                  string
	WorkgraphPath                   string
	EvidenceRoot                    string
	ExpectedNextNodeAfterCompletion string
}

func BuildAtlasCompactionResumePromptFixture(readback AtlasRecommendationReadback, options AtlasCompactionResumePromptOptions) (AtlasCompactionResumePrompt, string, error) {
	fixture := AtlasCompactionResumePrompt{
		Schema:                          AtlasCompactionResumePromptContract,
		NodeID:                          strings.TrimSpace(options.NodeID),
		Status:                          "generated",
		SourceReadbackPath:              publicArtifactRef(options.SourceReadbackPath),
		SourceReadbackDigest:            digestValue(readback),
		PromptPath:                      publicArtifactRef(options.PromptPath),
		LeaseStartPath:                  publicArtifactRef(options.LeaseStartPath),
		WorkgraphPath:                   publicArtifactRef(options.WorkgraphPath),
		StartedAt:                       readback.StartedAt,
		CompletedAt:                     readback.CompletedAt,
		ElapsedMinutes:                  readback.ElapsedMinutes,
		LeaseTimeStatus:                 readback.LeaseTimeStatus,
		CheckpointCount:                 readback.CheckpointCount,
		CompletedNodes:                  readback.CompletedNodes,
		TotalNodes:                      readback.TotalNodes,
		ReadyNodes:                      readback.ReadyNodes,
		BlockedNodes:                    readback.BlockedNodes,
		FailedNodes:                     readback.FailedNodes,
		FirstExecutableNode:             readback.FirstExecutableNode,
		ExactNextAction:                 readback.ExactNextAction,
		ReturnGateStatus:                readback.ReturnGateStatus,
		ContinuationContractReason:      readback.ContinuationContract.Reason,
		EarlyReturnRiskStatus:           readback.EarlyReturnRiskStatus,
		FinalResponseAllowed:            readback.FinalResponseAllowed,
		RefusesFinalResponse:            readback.ContinuationContract.RefusesFinalResponse,
		ExpectedNextNodeAfterCompletion: strings.TrimSpace(options.ExpectedNextNodeAfterCompletion),
		PromotionRequested:              false,
		PromotionGranted:                false,
		ClaimsAuthorityAdvance:          false,
		RSIRemainsDenied:                true,
	}
	if readback.Supervisor != nil {
		fixture.MinMinutes = readback.Supervisor.MinMinutes
		fixture.MaxMinutes = readback.Supervisor.MaxMinutes
	}
	prompt := BuildAtlasCompactionResumePrompt(readback, fixture, options)
	if err := ValidateAtlasCompactionResumePrompt(fixture); err != nil {
		return AtlasCompactionResumePrompt{}, "", err
	}
	return fixture, prompt, nil
}

func BuildAtlasCompactionResumePrompt(readback AtlasRecommendationReadback, fixture AtlasCompactionResumePrompt, options AtlasCompactionResumePromptOptions) string {
	evidenceRoot := strings.TrimSpace(options.EvidenceRoot)
	if evidenceRoot == "" {
		evidenceRoot = readback.EvidenceRoot
	}
	var b strings.Builder
	b.WriteString("You are AO Atlas, resuming the AO Atlas feature-depth wave after context compaction.\n\n")
	b.WriteString("Load and preserve this state exactly:\n")
	b.WriteString(fmt.Sprintf("- Evidence root: `%s`\n", filepath.ToSlash(evidenceRoot)))
	b.WriteString(fmt.Sprintf("- Lease start: `%s`\n", fixture.LeaseStartPath))
	b.WriteString(fmt.Sprintf("- Current workgraph: `%s`\n", fixture.WorkgraphPath))
	b.WriteString(fmt.Sprintf("- Current readback: `%s`\n\n", fixture.SourceReadbackPath))
	b.WriteString("Current status:\n")
	b.WriteString(fmt.Sprintf("- Completed nodes: %d / %d\n", fixture.CompletedNodes, fixture.TotalNodes))
	b.WriteString(fmt.Sprintf("- Ready nodes: %d\n", fixture.ReadyNodes))
	b.WriteString(fmt.Sprintf("- Blocked nodes: %d\n", fixture.BlockedNodes))
	b.WriteString(fmt.Sprintf("- Failed nodes: %d\n", fixture.FailedNodes))
	b.WriteString(fmt.Sprintf("- Next executable node: `%s`\n", fixture.FirstExecutableNode))
	b.WriteString(fmt.Sprintf("- Exact next action: %s\n", fixture.ExactNextAction))
	b.WriteString(fmt.Sprintf("- Elapsed minutes: `%d`\n", fixture.ElapsedMinutes))
	b.WriteString(fmt.Sprintf("- Lease time status: `%s`\n", fixture.LeaseTimeStatus))
	b.WriteString(fmt.Sprintf("- Checkpoint count: `%d`\n", fixture.CheckpointCount))
	b.WriteString(fmt.Sprintf("- Return gate: `%s`\n", fixture.ReturnGateStatus))
	b.WriteString(fmt.Sprintf("- Continuation contract reason: `%s`\n", fixture.ContinuationContractReason))
	b.WriteString(fmt.Sprintf("- Early-return risk: `%s`\n", fixture.EarlyReturnRiskStatus))
	b.WriteString(fmt.Sprintf("- Final response allowed: `%t`\n\n", fixture.FinalResponseAllowed))
	b.WriteString("Execution rules:\n")
	b.WriteString("- Emit Foundry import for exactly one active node at a time.\n")
	b.WriteString("- Do not restart completed nodes.\n")
	b.WriteString("- Do not produce a final response while ready nodes or exact next action remain.\n")
	b.WriteString("- Continue from the next executable node named above unless a true hard blocker remains.\n\n")
	b.WriteString("Safety boundaries:\n")
	b.WriteString("- No provider calls.\n")
	b.WriteString("- No credential or token inspection.\n")
	b.WriteString("- No direct main mutation.\n")
	b.WriteString("- No release, deploy, publish, upload, or tag.\n")
	b.WriteString("- No dependency updates unless separately authorized.\n")
	b.WriteString("- No auth, policy, or config widening.\n")
	b.WriteString("- No hidden instruction mutation.\n")
	b.WriteString("- No broad RSI claim.\n")
	b.WriteString("- RSI remains denied.\n")
	return b.String()
}

func ValidateAtlasCompactionResumePrompt(fixture AtlasCompactionResumePrompt) error {
	var errs []string
	requireContract(&errs, "compaction_resume_prompt", fixture.Schema, AtlasCompactionResumePromptContract)
	if fixture.Status != "generated" {
		errs = append(errs, "status must be generated")
	}
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	requireField(&errs, "source_readback_path", fixture.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", fixture.SourceReadbackPath, true)
	if fixture.SourceReadbackDigest != "" && !digestPattern.MatchString(fixture.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	requireField(&errs, "prompt_path", fixture.PromptPath)
	checkPublicPath(&errs, "prompt_path", fixture.PromptPath, true)
	checkOptionalPublicPath(&errs, "lease_start_path", fixture.LeaseStartPath)
	checkOptionalPublicPath(&errs, "workgraph_path", fixture.WorkgraphPath)
	if fixture.CompletedNodes <= 0 || fixture.TotalNodes < fixture.CompletedNodes || fixture.ReadyNodes < 0 || fixture.BlockedNodes < 0 || fixture.FailedNodes < 0 {
		errs = append(errs, "node counts must be positive and internally consistent")
	}
	requireField(&errs, "first_executable_node", fixture.FirstExecutableNode)
	checkPublicPath(&errs, "first_executable_node", fixture.FirstExecutableNode, true)
	requireField(&errs, "exact_next_action", fixture.ExactNextAction)
	checkPublicStrings(&errs, "exact_next_action", []string{fixture.ExactNextAction}, true)
	requireField(&errs, "return_gate_status", fixture.ReturnGateStatus)
	requireField(&errs, "continuation_contract_reason", fixture.ContinuationContractReason)
	requireField(&errs, "early_return_risk_status", fixture.EarlyReturnRiskStatus)
	if fixture.ReadyNodes > 0 && fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready nodes remain")
	}
	if fixture.ReadyNodes > 0 && !fixture.RefusesFinalResponse {
		errs = append(errs, "refuses_final_response must be true while ready nodes remain")
	}
	requireField(&errs, "expected_next_node_after_completion", fixture.ExpectedNextNodeAfterCompletion)
	checkPublicPath(&errs, "expected_next_node_after_completion", fixture.ExpectedNextNodeAfterCompletion, true)
	if fixture.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if fixture.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, false, false, false, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasCompactionResumePrompt(promptPath, fixturePath string, fixture AtlasCompactionResumePrompt, prompt string) error {
	if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return err
	}
	return WriteJSON(fixturePath, fixture)
}

func checkOptionalPublicPath(errs *[]string, field, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	checkPublicPath(errs, field, value, true)
}
