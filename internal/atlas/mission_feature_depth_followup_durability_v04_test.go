package atlas

import (
	"path/filepath"
	"testing"
)

func TestFeatureDepthFollowupDurabilityV04CompletesFortyNodeLongRunWave(t *testing.T) {
	root := filepath.Join(repoRoot(t), "docs", "evidence", "ao-atlas-feature-depth-followup-durability-v04")

	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(root, "recommendation-wave.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if wave.MissionID != "ao-atlas-next-feature-depth-followup-durability-v04" ||
		wave.TargetInstance != "ao-atlas-feature-depth-followup-durability-v04" ||
		wave.TotalTasks != 40 ||
		wave.MinimumTasks != 40 ||
		wave.NodeBudget != 40 ||
		wave.FinalResponseAllowed {
		t.Fatalf("v04 wave must preserve the 40-node Feature Depth contract: %#v", wave)
	}
	if wave.Supervisor == nil ||
		wave.Supervisor.MinNodes != 40 ||
		wave.Supervisor.MinMinutes != 120 ||
		wave.Supervisor.MaxMinutes != 180 ||
		wave.Supervisor.ContinueIfFastTarget != 40 {
		t.Fatalf("v04 wave must preserve the long-run supervisor lease: %#v", wave.Supervisor)
	}

	readback := mustLoadJSON[AtlasRecommendationReadback](t, filepath.Join(root, "nodes", "mission-recommendation-feature-depth-next-wave-40", "recommendation-readback-after.json"))
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		t.Fatal(err)
	}
	if readback.CompletedNodes != 40 ||
		readback.ReadyNodes != 0 ||
		readback.BlockedNodes != 0 ||
		readback.FailedNodes != 0 ||
		!readback.FinalResponseAllowed ||
		readback.ReturnGateStatus != "final_response_allowed" ||
		readback.LeaseTimeStatus != "minimum_minutes_met" ||
		!readback.MinMinutesMet ||
		readback.PublicSafetyScanStatus != "passed" {
		t.Fatalf("v04 final readback must close the long-run wave cleanly: %#v", readback)
	}

	report := mustLoadJSON[AtlasRecommendationEvidenceValidationReport](t, filepath.Join(root, "nodes", "mission-recommendation-feature-depth-next-wave-40", "feature-depth-evidence-validation-report.json"))
	if report.Schema != AtlasRecommendationEvidenceValidationReportContract ||
		report.Status != "passed" ||
		report.NodeCount != 40 ||
		report.JSONFileCount != 821 ||
		report.SchemaBoundFiles != report.JSONFileCount ||
		report.TypedValidatorFiles != 443 ||
		!report.RequiredFilenamesCovered ||
		len(report.MissingRequiredFiles) != 0 ||
		len(report.MissingSchemaFiles) != 0 ||
		len(report.FailedFiles) != 0 {
		t.Fatalf("v04 validation report must cover all node evidence without failures: %#v", report)
	}

	handoff := mustLoadJSON[finalFeatureDepthRecommendationHandoff](t, filepath.Join(root, "nodes", "mission-recommendation-feature-depth-next-wave-40", "final-feature-depth-recommendations.json"))
	if handoff.Schema != "ao.atlas.final-feature-depth-recommendations.v0.1" ||
		handoff.NodeID != "mission-recommendation-feature-depth-next-wave-40" ||
		handoff.Status != "ready_for_operator_handoff" ||
		handoff.CompletedNodesBefore != 39 ||
		handoff.ReadyNodesBefore != 1 ||
		handoff.ExpectedCompletedNodesAfter != 40 ||
		handoff.ExpectedReadyNodesAfter != 0 ||
		!handoff.ExpectedFinalResponseAllowedAfter ||
		handoff.TotalRecommendationCount != 40 ||
		handoff.OperatorReviewRecommendationCount != 40 ||
		len(handoff.Recommendations) != 40 ||
		!handoff.NoPromotionRequested ||
		handoff.PromotionGranted ||
		handoff.ClaimsAuthorityAdvance ||
		!handoff.RSIRemainsDenied ||
		handoff.SafeToExecute ||
		handoff.SchedulesWork ||
		handoff.ExecutesWork ||
		handoff.ApprovesWork ||
		handoff.MutatesRepositories {
		t.Fatalf("v04 final handoff must preserve 40 no-promotion Feature Depth recommendations: %#v", handoff)
	}
	for i, item := range handoff.Recommendations {
		if item.Rank != i+1 || item.ID == "" || item.Theme == "" || item.Task == "" {
			t.Fatalf("recommendation %d is not ranked and review-ready: %#v", i, item)
		}
	}
}
