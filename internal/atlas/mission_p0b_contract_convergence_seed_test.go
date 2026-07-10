package atlas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BContractConvergenceWaveSeedsThirtyMissionBoundNodes(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01")

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

type p0BContractConvergenceSeedSummary struct {
	Schema                 string `json:"schema"`
	Status                 string `json:"status"`
	MissionID              string `json:"mission_id"`
	TargetInstance         string `json:"target_instance"`
	EventIndexDigest        string `json:"event_index_digest"`
	FinalRollupDigest       string `json:"final_rollup_digest"`
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
