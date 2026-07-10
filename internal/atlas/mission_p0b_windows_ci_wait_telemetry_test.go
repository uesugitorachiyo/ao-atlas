package atlas

import (
	"path/filepath"
	"testing"
)

func TestP0BWindowsCIWaitTelemetryCoversCovenantAndCommandShards(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-22", "windows-ci-wait-state-telemetry.json")
	telemetry := mustLoadJSON[AtlasP0BWindowsCIWaitTelemetry](t, path)
	if err := ValidateAtlasP0BWindowsCIWaitTelemetry(telemetry); err != nil {
		t.Fatal(err)
	}
	if telemetry.Status != "recorded" ||
		telemetry.WaitThresholdSeconds != 600 ||
		telemetry.CommandShardCount == 0 ||
		telemetry.CovenantShardCount == 0 ||
		!telemetry.PendingStateObserved ||
		!telemetry.CompletedPassStateObserved ||
		telemetry.FailedStateObserved ||
		telemetry.FinalResponseAllowed ||
		telemetry.PromotionRequested ||
		telemetry.PromotionGranted ||
		!telemetry.RSIRemainsDenied {
		t.Fatalf("P0-B Windows telemetry summary drifted from safe wait-state coverage: %#v", telemetry)
	}
	for _, sample := range telemetry.WindowsCheckDurationSamples {
		state, err := ClassifyAtlasWindowsCIWaitState(AtlasWindowsCIWaitStateInput{
			CheckName:        sample.CheckName,
			GitHubStatus:     sample.FinalStatus,
			GitHubConclusion: sample.FinalConclusion,
			DurationSeconds:  sample.DurationSeconds,
			ThresholdSeconds: telemetry.WaitThresholdSeconds,
		})
		if err != nil {
			t.Fatal(err)
		}
		if state.WaitState != sample.WaitState ||
			state.FinalResponseAllowed ||
			state.ClaimsAuthorityAdvance ||
			!state.RSIRemainsDenied {
			t.Fatalf("wait-state sample must be reproducible and authority neutral: sample=%#v state=%#v", sample, state)
		}
	}
}

func TestP0BWindowsCIWaitTelemetryUsesTypedEvidenceValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-22", "windows-ci-wait-state-telemetry.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, AtlasP0BWindowsCIWaitTelemetryContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:p0b-windows-ci-wait-telemetry" {
		t.Fatalf("unexpected validator: %s", validator)
	}
}
