package atlas

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionLifecycleMetricsEvidenceRejectsHandoffCompletionClaim(t *testing.T) {
	path := filepath.Join(repoRoot(t), "examples", "invalid", "mission-lifecycle-metrics-handoff-count.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, MissionLifecycleMetricsEvidenceContract)
	if err == nil {
		t.Fatal("lifecycle metrics evidence accepted handoff steps as completed nodes")
	}
	if validator != "typed:mission-lifecycle-metrics" {
		t.Fatalf("unexpected lifecycle metrics validator: %q", validator)
	}
	if !strings.Contains(err.Error(), "handoff_steps_count_as_completed_nodes must be false") {
		t.Fatalf("unexpected lifecycle metrics rejection: %v", err)
	}
}

func TestMissionLifecycleMetricsEvidenceAcceptsEvidenceBasedCompletion(t *testing.T) {
	path := filepath.Join(repoRoot(t), "examples", "valid", "mission-lifecycle-metrics-readback.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, MissionLifecycleMetricsEvidenceContract)
	if err != nil {
		t.Fatalf("evidence-based lifecycle metrics rejected: %v", err)
	}
	if validator != "typed:mission-lifecycle-metrics" {
		t.Fatalf("unexpected lifecycle metrics validator: %q", validator)
	}
}

func TestMonth3FinalClosureMissionRecoveryInvariantFixture(t *testing.T) {
	path := filepath.Join(repoRoot(t), "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-20-mission-recovery-invariant", "mission-recovery-invariant.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, MissionLifecycleMetricsEvidenceContract)
	if err != nil {
		t.Fatalf("final-closure Mission recovery invariant rejected: %v", err)
	}
	if validator != "typed:mission-lifecycle-metrics" {
		t.Fatalf("unexpected lifecycle metrics validator: %q", validator)
	}
	fixture := mustLoadJSON[MissionLifecycleMetricsEvidence](t, path)
	if fixture.CompletedNodes != 19 ||
		fixture.EvidenceCompletedNodes != 19 ||
		fixture.HandoffStepsCountAsCompletedNodes ||
		fixture.CompletionBasis != "downstream_evidence_not_handoff_steps" ||
		fixture.SafeToExecute ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.MutatesRepositories ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("final-closure Mission recovery invariant lost handoff boundary: %#v", fixture)
	}
}
