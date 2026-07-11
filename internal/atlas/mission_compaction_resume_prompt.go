package atlas

import (
	"encoding/json"
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
	CheckpointReadbackPath          string
	EvidenceRoot                    string
	ExpectedNextNodeAfterCompletion string
}

func BuildAtlasCompactionResumePromptFixture(readback AtlasRecommendationReadback, options AtlasCompactionResumePromptOptions) (AtlasCompactionResumePrompt, string, error) {
	checkpointDigest, err := digestJSONArtifact(options.CheckpointReadbackPath)
	if err != nil {
		return AtlasCompactionResumePrompt{}, "", err
	}
	fixture := AtlasCompactionResumePrompt{
		Schema:                          AtlasCompactionResumePromptContract,
		NodeID:                          strings.TrimSpace(options.NodeID),
		Status:                          "generated",
		SourceReadbackPath:              publicArtifactRef(options.SourceReadbackPath),
		SourceReadbackDigest:            digestValue(readback),
		PromptPath:                      publicArtifactRef(options.PromptPath),
		LeaseStartPath:                  publicArtifactRef(options.LeaseStartPath),
		WorkgraphPath:                   publicArtifactRef(options.WorkgraphPath),
		CheckpointReadbackPath:          publicArtifactRef(options.CheckpointReadbackPath),
		CheckpointReadbackDigest:        checkpointDigest,
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
		SchemaHealthStatus:              readback.SchemaHealthStatus,
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
	b.WriteString(fmt.Sprintf("You are AO Atlas, resuming the AO Atlas %s after context compaction.\n\n", atlasCompactionResumeWaveLabel(readback.TargetInstance+" "+readback.MissionID+" "+readback.EvidenceRoot)))
	b.WriteString("Load and preserve this state exactly:\n")
	b.WriteString(fmt.Sprintf("- Evidence root: `%s`\n", filepath.ToSlash(evidenceRoot)))
	b.WriteString(fmt.Sprintf("- Lease start: `%s`\n", fixture.LeaseStartPath))
	b.WriteString(fmt.Sprintf("- Current workgraph: `%s`\n", fixture.WorkgraphPath))
	b.WriteString(fmt.Sprintf("- Current readback: `%s`\n", fixture.SourceReadbackPath))
	if fixture.CheckpointReadbackPath != "" {
		b.WriteString(fmt.Sprintf("- Checkpoint readback: `%s`\n", fixture.CheckpointReadbackPath))
		b.WriteString(fmt.Sprintf("- Checkpoint readback digest: `%s`\n", fixture.CheckpointReadbackDigest))
	}
	b.WriteString("\n")
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
	if fixture.SchemaHealthStatus != "" {
		b.WriteString(fmt.Sprintf("- Schema health status: `%s`\n", fixture.SchemaHealthStatus))
	}
	b.WriteString(fmt.Sprintf("- Final response allowed: `%t`\n\n", fixture.FinalResponseAllowed))
	b.WriteString("Execution rules:\n")
	b.WriteString("- Emit Foundry import for exactly one active node at a time.\n")
	b.WriteString("- Do not restart completed nodes.\n")
	b.WriteString("- Do not produce a final response while ready nodes or exact next action remain.\n")
	b.WriteString("- Continue from the next executable node named above unless a true hard blocker remains.\n\n")
	writeAtlasPromptSafetyBoundaries(&b, AtlasPromptSafetyBoundaryOptions{})
	return b.String()
}

func atlasCompactionResumeWaveLabel(targetInstance string) string {
	target := strings.ToLower(strings.TrimSpace(targetInstance))
	switch {
	case strings.Contains(target, "refactoring"):
		return "refactoring wave"
	case strings.Contains(target, "feature-depth"):
		return "feature-depth wave"
	case strings.Contains(target, "final-closure"):
		return "final-closure consolidation wave"
	default:
		return "recommendation wave"
	}
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
	checkOptionalPublicPath(&errs, "checkpoint_readback_path", fixture.CheckpointReadbackPath)
	checkOptionalDigest(&errs, "checkpoint_readback_digest", fixture.CheckpointReadbackDigest)
	if fixture.CheckpointReadbackPath != "" && fixture.CheckpointReadbackDigest == "" {
		errs = append(errs, "checkpoint_readback_digest is required when checkpoint_readback_path is set")
	}
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
	checkPublicPath(&errs, "schema_health_status", fixture.SchemaHealthStatus, true)
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

func digestJSONArtifact(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return "", err
	}
	return digestValue(value), nil
}

func checkOptionalPublicPath(errs *[]string, field, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	checkPublicPath(errs, field, value, true)
}
