package atlas

import (
	"fmt"
	"path/filepath"
	"strings"
)

type AtlasRecommendationResumePromptOptions struct {
	EvidenceRoot   string
	LeaseStartPath string
	WorkgraphPath  string
	ReadbackPath   string
}

func BuildAtlasRecommendationResumePrompt(readback AtlasRecommendationReadback, options AtlasRecommendationResumePromptOptions) string {
	evidenceRoot := strings.TrimSpace(options.EvidenceRoot)
	if evidenceRoot == "" {
		evidenceRoot = readback.EvidenceRoot
	}
	leaseStartPath := filepath.ToSlash(strings.TrimSpace(options.LeaseStartPath))
	workgraphPath := filepath.ToSlash(strings.TrimSpace(options.WorkgraphPath))
	readbackPath := filepath.ToSlash(strings.TrimSpace(options.ReadbackPath))
	minMinutes := readback.ElapsedMinutes
	if readback.Supervisor != nil {
		minMinutes = readback.Supervisor.MinMinutes
	}
	nextNode := readback.FirstExecutableNode
	if strings.TrimSpace(nextNode) == "" {
		nextNode = "none"
	}

	var b strings.Builder
	b.WriteString("You are AO Atlas, continuing the AO Atlas long-run recommendation wave.\n\n")
	b.WriteString("Do not ask the operator for permission. Do not reset the lease clock. Load and preserve:\n\n")
	if evidenceRoot != "" {
		b.WriteString(fmt.Sprintf("- Evidence root: `%s`\n", filepath.ToSlash(evidenceRoot)))
	}
	if leaseStartPath != "" {
		b.WriteString(fmt.Sprintf("- Lease start: `%s`\n", leaseStartPath))
	}
	if workgraphPath != "" {
		b.WriteString(fmt.Sprintf("- Current workgraph: `%s`\n", workgraphPath))
	}
	if readbackPath != "" {
		b.WriteString(fmt.Sprintf("- Current readback: `%s`\n", readbackPath))
	}
	b.WriteString("\nCurrent status:\n")
	b.WriteString(fmt.Sprintf("- Completed nodes: %d / %d\n", readback.CompletedNodes, readback.TotalNodes))
	b.WriteString(fmt.Sprintf("- Ready nodes: %d\n", readback.ReadyNodes))
	b.WriteString(fmt.Sprintf("- Elapsed minutes at latest checkpoint: %d\n", readback.ElapsedMinutes))
	b.WriteString(fmt.Sprintf("- Minimum minutes: %d\n", minMinutes))
	b.WriteString(fmt.Sprintf("- `min_minutes_met=%t`\n", readback.MinMinutesMet))
	b.WriteString(fmt.Sprintf("- `final_response_allowed=%t`\n", readback.FinalResponseAllowed))
	b.WriteString(fmt.Sprintf("- Return gate: `%s`\n", readback.ReturnGateStatus))
	b.WriteString(fmt.Sprintf("- Continuation contract reason: `%s`\n", readback.ContinuationContract.Reason))
	b.WriteString(fmt.Sprintf("- Early-return risk: `%s`\n", readback.EarlyReturnRiskStatus))
	b.WriteString(fmt.Sprintf("- Checkpoint count: %d\n", readback.CheckpointCount))
	b.WriteString(fmt.Sprintf("- Next executable node: `%s`\n\n", nextNode))
	b.WriteString("Goal:\n")
	b.WriteString("Continue the useful 2-3 hour Atlas-owned hardening wave. Execute exactly one bounded node at a time, preserving the original `started_at` from `lease-start.json`, until all ready work is handled or a true hard blocker remains after safe repair attempts.\n\n")
	b.WriteString("Exact next action:\n")
	b.WriteString(fmt.Sprintf("- %s\n\n", readback.ExactNextAction))
	b.WriteString("Blocked-node continuation:\n")
	b.WriteString("- If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair.\n\n")
	writeAtlasPromptSafetyBoundaries(&b, AtlasPromptSafetyBoundaryOptions{
		SuffixLines: []string{"Keep exactly one executable mutation node active at a time."},
	})
	b.WriteString("\n")
	b.WriteString("Verification:\n")
	b.WriteString("- `go test ./... -count=1`\n")
	b.WriteString("- `go vet ./...`\n")
	b.WriteString("- `go build ./cmd/atlas`\n")
	b.WriteString("- `scripts/production-readiness.sh`\n")
	b.WriteString("- `scripts/atlas-foundry-roundtrip-smoke.sh`\n")
	b.WriteString("- Public-safety wording scan over changed docs and evidence.\n\n")
	b.WriteString("Final response is allowed only when the authoritative recommendation readback has `final_response_allowed=true`, the execution readback agrees, Command and Foundry summaries agree, Promoter records no promotion, verification passes, the repo is clean and synced, and no ready nodes or exact next actions remain.\n")
	b.WriteString("If `ready_nodes > 0` or `exact_next_action` is non-empty, do not produce a final response.\n")
	return b.String()
}

type AtlasRecommendationPromptBudget struct {
	MinNodes         int
	MinMinutes       int
	MaxMinutes       int
	MaxIterations    int
	ReturnOnlyWhen   string
	CheckpointPolicy string
	StopConditions   []string
}

func BuildAtlasRecommendationPromptBudget(wave AtlasRecommendationWave) AtlasRecommendationPromptBudget {
	budget := AtlasRecommendationPromptBudget{
		MinNodes:         wave.MinimumTasks,
		MinMinutes:       wave.EstimatedMinutes,
		MaxMinutes:       wave.EstimatedMinutes,
		MaxIterations:    wave.NodeBudget,
		ReturnOnlyWhen:   fmt.Sprintf("all generated nodes complete, at least %d nodes complete, or a true hard blocker remains", wave.MinimumTasks),
		CheckpointPolicy: "after each node or timed interval",
	}
	if wave.Supervisor != nil {
		budget.MinNodes = wave.Supervisor.MinNodes
		if budget.MinNodes == 0 {
			budget.MinNodes = wave.MinimumTasks
		}
		budget.MinMinutes = wave.Supervisor.MinMinutes
		budget.MaxMinutes = wave.Supervisor.MaxMinutes
		budget.MaxIterations = wave.Supervisor.ContinueIfFastTarget
		budget.ReturnOnlyWhen = wave.Supervisor.ReturnOnlyWhen
		budget.CheckpointPolicy = wave.Supervisor.CheckpointPolicy
	}
	budget.StopConditions = []string{
		fmt.Sprintf("Target duration: %d to %d minutes.", budget.MinMinutes, budget.MaxMinutes),
		fmt.Sprintf("Node floor stop gate: complete at least %d nodes before final response unless a true hard blocker remains.", budget.MinNodes),
		fmt.Sprintf("Lease floor stop gate: do not return before min_minutes=%d unless a true hard blocker remains.", budget.MinMinutes),
		fmt.Sprintf("Continue-if-fast stop gate: if %d nodes finish quickly and no blocker remains, continue through %d nodes.", budget.MinNodes, budget.MaxIterations),
		"Ready-work stop gate: if ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.",
	}
	return budget
}

func buildAtlasRecommendationPrompt(wave AtlasRecommendationWave) string {
	var b strings.Builder
	budget := BuildAtlasRecommendationPromptBudget(wave)
	b.WriteString("You are AO Atlas, coordinating AO Foundry, AO Promoter, AO Command, and AO Architecture for a long-run Atlas-owned correction wave.\n\n")
	b.WriteString("Do not ask the operator for permission. Do not stop after one repo, one PR, one CI pass, one Foundry import, one route decision, one evidence artifact, or one short batch. Continue until the workgraph is completed with evidence, at least the lease minimum is met, or a true hard blocker remains after all safe repair, repack, and support work has been attempted.\n\n")
	b.WriteString("Current state:\n")
	b.WriteString(fmt.Sprintf("- Mission: %s.\n", wave.MissionID))
	b.WriteString(fmt.Sprintf("- Target instance: %s.\n", wave.TargetInstance))
	b.WriteString(fmt.Sprintf("- Generated Atlas-owned nodes: %d.\n", wave.TotalTasks))
	b.WriteString(fmt.Sprintf("- Lease minimum: %d nodes, %d to %d minutes.\n", budget.MinNodes, budget.MinMinutes, budget.MaxMinutes))
	b.WriteString(fmt.Sprintf("- Continue-if-fast target: %d nodes.\n", budget.MaxIterations))
	b.WriteString(fmt.Sprintf("- Final response allowed: %t, because %s.\n", wave.FinalResponseAllowed, wave.FinalResponseReason))
	b.WriteString(fmt.Sprintf("- Source digest: %s.\n\n", wave.SourceDigest))
	b.WriteString("Problem:\n")
	b.WriteString("- Recent AO Atlas/Mission recommendation prompts returned after short batches instead of sustaining 2-3 hour workgraphs.\n")
	b.WriteString("- Double the previous short batch when explicitly requested, and otherwise use the v0.2 2-3 hour supervisor default.\n")
	b.WriteString("- This continuation must behave like a long-run supervisor: Atlas owns sequencing, Foundry owns bounded implementation nodes, and Blueprint is used only for genuinely new requirements or authorization.\n\n")
	b.WriteString("Goal:\n")
	b.WriteString(fmt.Sprintf("- Target 2-3 hours and complete a durable AO Atlas long-run wave for %s.\n", wave.MissionID))
	b.WriteString(fmt.Sprintf("- Execute at least %d bounded Atlas nodes from the generated workgraph.\n", budget.MinNodes))
	b.WriteString(fmt.Sprintf("- Complete at least %d bounded implementation/evidence nodes before final response unless a true hard blocker remains.\n", budget.MinNodes))
	b.WriteString(fmt.Sprintf("- If the first %d nodes finish quickly and no blocker remains, continue through the %d-node continue-if-fast target.\n\n", budget.MinNodes, budget.MaxIterations))
	b.WriteString(fmt.Sprintf("Return only after all generated nodes complete, at least %d bounded Atlas nodes complete, or a true hard blocker remains after safe repair attempts.\n\n", budget.MinNodes))
	b.WriteString("Minimum work budget:\n")
	b.WriteString(fmt.Sprintf("- min_nodes: %d\n", budget.MinNodes))
	b.WriteString(fmt.Sprintf("- min_minutes: %d\n", budget.MinMinutes))
	b.WriteString(fmt.Sprintf("- max_minutes: %d\n", budget.MaxMinutes))
	b.WriteString(fmt.Sprintf("- max_iterations: %d\n", budget.MaxIterations))
	b.WriteString(fmt.Sprintf("- return_only_when: %s\n", budget.ReturnOnlyWhen))
	b.WriteString(fmt.Sprintf("- checkpoint_policy: %s\n\n", budget.CheckpointPolicy))
	b.WriteString("Stop gates:\n")
	for _, condition := range budget.StopConditions {
		b.WriteString(fmt.Sprintf("- %s\n", condition))
	}
	b.WriteString("- Checkpoint stop gate: record a checkpoint after each node or timed interval before evaluating final response.\n\n")
	writeAtlasPromptSafetyBoundaries(&b, AtlasPromptSafetyBoundaryOptions{
		PrefixLines: []string{"Keep exactly one executable mutation node active at a time."},
		SuffixLines: []string{"Use existing repo auth only for normal PR, CI, and merge if available without exposing credentials."},
	})
	b.WriteString("\n")
	b.WriteString("Required work:\n")
	for _, task := range wave.Tasks {
		b.WriteString(fmt.Sprintf("%s. %s\n", strings.TrimPrefix(task.ID, "next-"), task.Task))
	}
	b.WriteString("\nPer-node requirements:\n")
	b.WriteString("- Generate or validate node gate, candidate record, rollback record, implementation evidence, tests, verification command output, Sentinel/public-safety wording evidence where applicable, Promoter/no-promotion or promotion-readiness evidence where applicable, and Command/readback evidence where applicable.\n")
	b.WriteString("- Emit a Foundry import for exactly one active node at a time, execute the node, verify locally, record run-link evidence, complete the node in Atlas, evaluate the next stop gate, and continue.\n\n")
	b.WriteString("Regression tests:\n")
	b.WriteString("- Prove the recommendation wave defaults to at least 30 nodes and 120 minutes.\n")
	b.WriteString("- Prove the continue-if-fast target generates 40 bounded Atlas-owned tasks.\n")
	b.WriteString("- Prove mixed-owner default waves are rejected with exact readback.\n")
	b.WriteString("- Prove final response remains denied while ready nodes or exact next actions remain.\n\n")
	b.WriteString("Verification:\n")
	b.WriteString("- go test ./... -count=1\n")
	b.WriteString("- go vet ./...\n")
	b.WriteString("- go build ./cmd/atlas\n")
	b.WriteString("- scripts/production-readiness.sh\n")
	b.WriteString("- scripts/atlas-foundry-roundtrip-smoke.sh\n")
	b.WriteString("- Public-safety wording scan over changed docs and readbacks.\n\n")
	b.WriteString("Final response only after completion or true hard blocker:\n")
	b.WriteString("- Include `early_return_risk_status` in continuation prompts and treat any blocked risk status as final-response denial evidence.\n")
	b.WriteString("- If ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.\n")
	b.WriteString("- If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair.\n")
	b.WriteString("- completed nodes / total nodes\n")
	b.WriteString("- list of node statuses\n")
	b.WriteString("- merged PRs by repo or local commits if remote lifecycle is blocked\n")
	b.WriteString("- evidence roots\n")
	b.WriteString("- final AO Atlas long-run supervisor status\n")
	b.WriteString("- Foundry rollup\n")
	b.WriteString("- Command readback\n")
	b.WriteString(fmt.Sprintf("- Feature Depth Recommendations, at least %d tasks\n", wave.TotalTasks))
	b.WriteString("- verification results\n")
	b.WriteString("- clean/synced repo status\n")
	b.WriteString("- exact next action\n")
	return b.String()
}
