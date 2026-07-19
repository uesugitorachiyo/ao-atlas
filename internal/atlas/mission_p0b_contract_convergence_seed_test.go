package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BContractConvergenceWaveSeedsThirtyMissionBoundNodes(t *testing.T) {
	root := repoRoot(t)
	waveRootRel := filepath.Join("docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01")
	waveRoot := filepath.Join(root, waveRootRel)

	source := mustLoadJSON[AOMissionFeatureDepthRecommendations](t, filepath.Join(waveRoot, "source-feature-depth-recommendations.json"))
	if err := ValidateAtlasNextWaveFeatureDepthRecommendations(source, 30); err != nil {
		t.Fatal(err)
	}
	if source.MissionID != "mission-710327df54728420" ||
		source.RecommendationCount != 30 ||
		source.MinimumTasks != 30 ||
		source.Status != "ready" ||
		source.SafeToExecute ||
		source.SchedulesWork ||
		source.ExecutesWork ||
		source.ApprovesWork ||
		source.MutatesRepositories {
		t.Fatalf("P0-B source recommendations must be a 30-node planning-only bundle: %#v", source)
	}

	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(waveRoot, "recommendation-wave.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if wave.MissionID != "mission-710327df54728420" ||
		wave.TargetInstance != "ao-stack-p0b-contract-convergence-wave-v01" ||
		wave.TotalTasks != 30 ||
		wave.MinimumTasks != 30 ||
		wave.NodeBudget != 30 ||
		wave.EstimatedMinutes != 150 ||
		wave.FinalResponseAllowed ||
		wave.SafeToExecute ||
		wave.SchedulesWork ||
		wave.ExecutesWork ||
		wave.ApprovesWork {
		t.Fatalf("P0-B wave must preserve long-run planning-only import state: %#v", wave)
	}
	if wave.Supervisor == nil ||
		wave.Supervisor.MinNodes != 30 ||
		wave.Supervisor.MinMinutes != 120 ||
		wave.Supervisor.MaxMinutes != 180 ||
		wave.Supervisor.ContinueIfFastTarget != 30 ||
		wave.Supervisor.ReturnOnlyWhen != "all_30_p0b_contract_convergence_nodes_complete_or_true_hard_blocker" {
		t.Fatalf("P0-B wave must preserve 2-3 hour supervisor contract: %#v", wave.Supervisor)
	}

	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(waveRoot, "recommendation-readback.json"))
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	if readback.CompletedNodes != 0 ||
		readback.ReadyNodes != 30 ||
		readback.BlockedNodes != 0 ||
		readback.FailedNodes != 0 ||
		readback.FinalResponseAllowed ||
		readback.FirstExecutableNode != "mission-recommendation-p0b-contract-convergence-01" ||
		readback.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		!readback.ContinuationContract.RefusesFinalResponse ||
		readback.ContinuationContract.Reason != "ready_nodes_or_exact_next_action_remain" {
		t.Fatalf("P0-B readback must require continuation across all ready nodes: %#v", readback)
	}

	summary := mustLoadJSON[p0BContractConvergenceSeedSummary](t, filepath.Join(waveRoot, "p0b-contract-convergence-seed-summary.json"))
	if summary.Schema != "ao.atlas.p0b-contract-convergence-seed-summary.v0.1" ||
		summary.Status != "ready_for_long_run_execution" ||
		summary.MissionID != "mission-710327df54728420" ||
		summary.EventIndexDigest != "sha256:fffb805b58f76f46980f6b1cb6351b31e89cc8e0e0b00d5456ae509f7a0a4423" ||
		summary.FinalRollupDigest != "sha256:7d81cd3bb89aa9236ee58e0978e1a6cdcfc617e232d415c1c25dfd883bd51f38" ||
		summary.RecommendationCount != 30 ||
		summary.ReadyNodes != 30 ||
		summary.CompletedNodes != 0 ||
		summary.FinalResponseAllowed ||
		!summary.NoPromotionRequested ||
		summary.PromotionGranted ||
		summary.ClaimsAuthorityAdvance ||
		!summary.RSIRemainsDenied ||
		summary.SafeToExecute ||
		summary.SchedulesWork ||
		summary.ExecutesWork ||
		summary.ApprovesWork ||
		summary.MutatesRepositories {
		t.Fatalf("P0-B seed summary lost Mission digest or safety contract: %#v", summary)
	}

	promptBytes, err := os.ReadFile(filepath.Join(waveRoot, "next-recommended-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"Mission: mission-710327df54728420",
		"Target instance: ao-stack-p0b-contract-convergence-wave-v01",
		"Lease minimum: 30 nodes, 120 to 180 minutes.",
		"event_index_digest=sha256:fffb805b58f76f46980f6b1cb6351b31e89cc8e0e0b00d5456ae509f7a0a4423",
		"final_rollup_digest=sha256:7d81cd3bb89aa9236ee58e0978e1a6cdcfc617e232d415c1c25dfd883bd51f38",
		"Do not ask the operator for permission.",
		"p0b-contract-convergence-30. Generate P0-C Mission to Foundry real complete-path handoff.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("P0-B prompt missing %q", want)
		}
	}
}

func TestP0BContractConvergenceWorkgraphReadinessMaterializesFirstExecutableNode(t *testing.T) {
	root := repoRoot(t)
	waveRootRel := filepath.Join("docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01")
	waveRoot := filepath.Join(root, waveRootRel)

	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(waveRoot, "recommendation-wave.json"))
	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(waveRoot, "recommendation-workgraph.json"))
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	if workgraph.ID != "ao-atlas-recommendation-wave-mission-710327df54728420" ||
		workgraph.TargetInstance != "ao-stack-p0b-contract-convergence-wave-v01" ||
		len(workgraph.Nodes) != 30 {
		t.Fatalf("P0-B workgraph lost mission identity or node budget: %#v", workgraph)
	}
	first := workgraph.Nodes[0]
	if first.ID != "mission-recommendation-p0b-contract-convergence-01" ||
		first.Status != "ready" ||
		len(first.Dependencies) != 0 ||
		!first.StitchTask ||
		first.FactoryTask.ID != "mission-recommendation-p0b-contract-convergence-01-task" ||
		first.FactoryTask.Objective != wave.Tasks[0].Task ||
		first.FactoryTask.AuthorityBoundary != "atlas_recommendation_planning_only" {
		t.Fatalf("first P0-B workgraph node is not the Mission-recorded executable node: %#v", first)
	}
	second := workgraph.Nodes[1]
	if len(second.Dependencies) != 1 || second.Dependencies[0] != first.ID {
		t.Fatalf("second P0-B node must remain serialized behind first node: %#v", second)
	}
	for _, want := range []string{
		"source_digest:" + wave.SourceDigest,
		"source_recommendation:p0b-contract-convergence-01",
		"source_task_digest:" + wave.Tasks[0].SourceTaskDigest,
	} {
		if !containsString(first.FactoryTask.RequiredEvidence, want) {
			t.Fatalf("first P0-B node lost source evidence binding %q: %#v", want, first.FactoryTask.RequiredEvidence)
		}
	}
	for _, want := range []string{
		"no provider calls",
		"no credential inspection",
		"no direct main mutation",
		"no broad RSI claim",
	} {
		if !containsString(first.FactoryTask.SafetyLimits, want) {
			t.Fatalf("first P0-B node lost safety limit %q: %#v", want, first.FactoryTask.SafetyLimits)
		}
	}

	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(waveRoot, "recommendation-readback.json"))
	packet := mustLoadJSON[AtlasRecommendationWorkgraphReadinessPacket](t, filepath.Join(waveRoot, "workgraph-readiness-packet.json"))
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(packet, readback); err != nil {
		t.Fatal(err)
	}
	if packet.TotalNodes != 30 ||
		packet.MinimumNodes != 30 ||
		packet.NodeBudget != 30 ||
		packet.ContinueIfFastTarget != 30 ||
		packet.ExecutableReadyNodes != 1 ||
		packet.FirstExecutableNode != first.ID ||
		packet.FinalResponseAllowed ||
		!packet.RefusesFinalResponse ||
		packet.ReturnGateStatus != "blocked_ready_nodes_remain" ||
		!packet.RSIRemainsDenied {
		t.Fatalf("P0-B readiness packet must deny final response and expose exactly one executable node: %#v", packet)
	}

	instance := mustLoadJSON[Instance](t, filepath.Join(waveRoot, "stack-instance.json"))
	if err := ValidateInstance(instance); err != nil {
		t.Fatal(err)
	}
	if instance.ID != workgraph.TargetInstance {
		t.Fatalf("stack instance id must match P0-B workgraph target: instance=%s workgraph=%s", instance.ID, workgraph.TargetInstance)
	}

	foundryOut := filepath.Join("..", "..", "target", "p0b-contract-convergence-first-node-foundry-import")
	foundryOutAbs := filepath.Join(root, "target", "p0b-contract-convergence-first-node-foundry-import")
	if err := os.RemoveAll(foundryOutAbs); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(foundryOutAbs)
	})
	var out bytes.Buffer
	code := Run([]string{
		"foundry", "import",
		"--workgraph", filepath.Join("..", "..", waveRootRel, "recommendation-workgraph.json"),
		"--instance", filepath.Join("..", "..", waveRootRel, "stack-instance.json"),
		"--node", first.ID,
		"--out", foundryOut,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("P0-B foundry import failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "tasks=1") ||
		!strings.Contains(out.String(), "next_recommended_action=Move to ../ao-foundry") {
		t.Fatalf("P0-B foundry import output lost single-node continuation readback: %s", out.String())
	}
	manifest := mustLoadJSON[FoundryImport](t, filepath.Join(foundryOutAbs, "foundry-import.json"))
	if err := ValidateFoundryImport(manifest); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryImportMatchesWorkgraph(workgraph, manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Tasks) != 1 ||
		manifest.Tasks[0].NodeID != first.ID ||
		manifest.Tasks[0].TaskID != first.FactoryTask.ID ||
		manifest.Tasks[0].AuthorityBoundary != "atlas_recommendation_planning_only" ||
		manifest.SchedulesWork ||
		manifest.ExecutesWork ||
		manifest.ApprovesWork {
		t.Fatalf("P0-B foundry import must stay fixture-only for exactly the first node: %#v", manifest)
	}
}

type p0BContractConvergenceSeedSummary struct {
	Schema                 string `json:"schema"`
	Status                 string `json:"status"`
	MissionID              string `json:"mission_id"`
	TargetInstance         string `json:"target_instance"`
	EventIndexDigest       string `json:"event_index_digest"`
	FinalRollupDigest      string `json:"final_rollup_digest"`
	RecommendationCount    int    `json:"recommendation_count"`
	ReadyNodes             int    `json:"ready_nodes"`
	CompletedNodes         int    `json:"completed_nodes"`
	FinalResponseAllowed   bool   `json:"final_response_allowed"`
	NoPromotionRequested   bool   `json:"no_promotion_requested"`
	PromotionGranted       bool   `json:"promotion_granted"`
	ClaimsAuthorityAdvance bool   `json:"claims_authority_advance"`
	RSIRemainsDenied       bool   `json:"rsi_remains_denied"`
	SafeToExecute          bool   `json:"safe_to_execute"`
	SchedulesWork          bool   `json:"schedules_work"`
	ExecutesWork           bool   `json:"executes_work"`
	ApprovesWork           bool   `json:"approves_work"`
	MutatesRepositories    bool   `json:"mutates_repositories"`
}
