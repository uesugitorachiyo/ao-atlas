package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecommendationWaveExplicitZeroMinimumMinutesIsPreserved(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "useful-work",
		MinMinutes:          0,
		MinMinutesSet:       true,
		MaxMinutes:          180,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Wave.Supervisor == nil || result.Wave.Supervisor.MinMinutes != 0 {
		t.Fatalf("explicit zero minimum was not preserved: %#v", result.Wave.Supervisor)
	}
	if result.Wave.EstimatedMinutes != 120 {
		t.Fatalf("target estimate must remain independent from elapsed minimum: %d", result.Wave.EstimatedMinutes)
	}
	if strings.Contains(result.Prompt, "do not return before min_minutes=0") {
		t.Fatalf("useful-work prompt must not require elapsed-time padding: %s", result.Prompt)
	}
}

func TestRecommendationWaveOmittedMinimumMinutesKeepsHistoricalDefault(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "historical-default",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Wave.Supervisor == nil || result.Wave.Supervisor.MinMinutes != 120 {
		t.Fatalf("omitted minimum must retain the historical default: %#v", result.Wave.Supervisor)
	}
}

func TestRecommendationWaveUsefulWorkSmallWaveGetsIndependentDefaults(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 20, false)

	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath:  recommendationsPath,
		TargetInstance:       "small-useful-work",
		MinTasks:             20,
		NodeBudget:           20,
		ContinueIfFastTarget: 20,
		MinMinutes:           0,
		MinMinutesSet:        true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Wave.EstimatedMinutes != 90 || result.Wave.Supervisor == nil || result.Wave.Supervisor.MinMinutes != 0 || result.Wave.Supervisor.MaxMinutes != 90 {
		t.Fatalf("small useful-work defaults must keep target/min/max independent: %#v", result.Wave)
	}
}

func TestMissionRecommendationsImportPersistsExplicitZeroMinimum(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	outDir := filepath.Join(dir, "out")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "useful-work-cli",
		"--min-minutes", "0",
		"--max-minutes", "180",
		"--started-at", "2026-08-03T01:00:00Z",
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("explicit zero import failed: %s", out.String())
	}
	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(outDir, "recommendation-wave.json"))
	lease := mustLoadJSON[AtlasRecommendationLeaseStart](t, filepath.Join(outDir, "lease-start.json"))
	if wave.Supervisor == nil || wave.Supervisor.MinMinutes != 0 || lease.MinMinutes != 0 {
		t.Fatalf("CLI did not preserve explicit zero: wave=%#v lease=%#v", wave.Supervisor, lease)
	}
	if err := ValidateAtlasRecommendationLeaseStart(lease); err != nil {
		t.Fatalf("zero-minimum lease start should validate: %v", err)
	}
}

func TestMissionRecommendationsImportRejectsNegativeMinimum(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "negative-minimum",
		"--min-minutes", "-1",
		"--json",
	}, &out, &out)
	if code == 0 || !strings.Contains(out.String(), "min_minutes must be zero or greater") {
		t.Fatalf("negative minimum did not fail closed: %s", out.String())
	}
}

func TestRecommendationReadbackUsefulWorkMayCompleteAtZeroElapsedMinutes(t *testing.T) {
	result := usefulWorkRecommendationWave(t)
	completed := completeRecommendationNodes(result.Workgraph, 40)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
		StartedAt:      "2026-08-03T01:00:00Z",
		CompletedAt:    "2026-08-03T01:00:00Z",
		ElapsedMinutes: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !readback.FinalResponseAllowed || !readback.MinMinutesMet || readback.LeaseTimeStatus != "minimum_minutes_not_required" {
		t.Fatalf("useful work should close without elapsed padding: %#v", readback)
	}
}

func TestRecommendationReadbackUsefulWorkStillRequiresTimingForMaximum(t *testing.T) {
	result := usefulWorkRecommendationWave(t)
	completed := completeRecommendationNodes(result.Workgraph, 40)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if readback.FinalResponseAllowed || readback.LeaseTimeStatus != "lease_timing_missing" || readback.ReturnGateStatus != "blocked_lease_timing_missing" {
		t.Fatalf("missing timing cannot prove the hard maximum: %#v", readback)
	}
}

func TestRecommendationReadbackFailsClosedAboveMaximumMinutes(t *testing.T) {
	result := usefulWorkRecommendationWave(t)
	completed := completeRecommendationNodes(result.Workgraph, 40)
	readback, err := BuildAtlasRecommendationReadback(result.Wave, completed, AtlasRecommendationReadbackOptions{
		StartedAt:      "2026-08-03T01:00:00Z",
		CompletedAt:    "2026-08-03T04:01:00Z",
		ElapsedMinutes: 181,
	})
	if err != nil {
		t.Fatal(err)
	}
	if readback.FinalResponseAllowed || readback.LeaseTimeStatus != "maximum_minutes_exceeded" || readback.ReturnGateStatus != "blocked_maximum_minutes_exceeded" {
		t.Fatalf("over-maximum work must fail closed: %#v", readback)
	}
	if readback.FinalResponseReason != "maximum lease minutes exceeded" || !strings.Contains(readback.ExactNextAction, "reconcile") {
		t.Fatalf("over-maximum readback needs an exact reconciliation action: %#v", readback)
	}
	checkpoint := BuildAtlasRecommendationCheckpointReadback(readback)
	if err := ValidateAtlasRecommendationCheckpointReadback(checkpoint); err != nil {
		t.Fatalf("checkpoint rejected maximum violation: %v", err)
	}
	command := BuildAtlasRecommendationCommandReadback(readback)
	execution := BuildAtlasRecommendationExecutionReadback(readback)
	if err := ValidateAtlasRecommendationExecutionReadback(execution, readback); err != nil {
		t.Fatalf("execution readback rejected maximum violation: %v", err)
	}
	if checkpoint.LeaseTimeStatus != readback.LeaseTimeStatus || command.LeaseTimeStatus != readback.LeaseTimeStatus || execution.ReturnGateStatus != readback.ReturnGateStatus {
		t.Fatalf("terminal surfaces disagree on maximum violation: checkpoint=%#v command=%#v execution=%#v", checkpoint, command, execution)
	}

	tampered := readback
	tampered.ElapsedMinutes = 180
	if err := ValidateAtlasRecommendationReadback(tampered); err == nil || !strings.Contains(err.Error(), "maximum_minutes_exceeded requires useful-work mode") {
		t.Fatalf("stale maximum violation status was accepted: %v", err)
	}
}

func TestRecommendationWaveAPIRejectsNegativeMinimum(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	_, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "negative-api",
		MinMinutes:          -1,
	})
	if err == nil || !strings.Contains(err.Error(), "min_minutes must be zero or greater") {
		t.Fatalf("negative API minimum was accepted: %v", err)
	}
}

func usefulWorkRecommendationWave(t *testing.T) AtlasRecommendationWaveResult {
	t.Helper()
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	result, err := BuildAtlasRecommendationWave(AtlasRecommendationWaveOptions{
		RecommendationsPath: recommendationsPath,
		TargetInstance:      "useful-work-readback",
		MinMinutes:          0,
		MinMinutesSet:       true,
		MaxMinutes:          180,
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
