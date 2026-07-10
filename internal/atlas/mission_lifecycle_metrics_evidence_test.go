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
