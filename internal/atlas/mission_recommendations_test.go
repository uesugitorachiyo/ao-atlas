package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsImportBuildsDoubleSizeWaveAndWorkgraph(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	outDir := filepath.Join(dir, "recommendations-out")
	writeFeatureDepthBundle(t, recommendationsPath, 20, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--min-tasks", "20",
		"--node-budget", "20",
		"--estimated-minutes", "90",
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "recommendation_tasks=20") ||
		!strings.Contains(out.String(), "estimated_minutes=90") {
		t.Fatalf("import output missing long-run counts: %s", out.String())
	}

	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(outDir, "recommendation-wave.json"))
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		t.Fatal(err)
	}
	if wave.MissionID != "mission-long-wave" || wave.TotalTasks != 20 || wave.EstimatedMinutes != 90 || wave.NodeBudget != 20 {
		t.Fatalf("bad recommendation wave summary: %#v", wave)
	}
	if wave.SafeToExecute || wave.SchedulesWork || wave.ExecutesWork || wave.ApprovesWork {
		t.Fatalf("recommendation wave widened authority: %#v", wave)
	}
	if !strings.Contains(wave.NextRecommendedPrompt, "at least 20 bounded Atlas nodes") ||
		!strings.Contains(wave.NextRecommendedPrompt, "Return only after") {
		t.Fatalf("wave missing continuation prompt: %q", wave.NextRecommendedPrompt)
	}

	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(outDir, "recommendation-workgraph.json"))
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) != 20 {
		t.Fatalf("expected 20 recommendation nodes, got %d", len(workgraph.Nodes))
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.ExecutableReadyNodeIDs) != 1 || state.ExecutableReadyNodeIDs[0] != "mission-recommendation-next-01" {
		t.Fatalf("expected exactly one executable-ready node, got %#v", state.ExecutableReadyNodeIDs)
	}
	for _, node := range workgraph.Nodes {
		if node.FactoryTask.TargetFactoryRepo != "ao-atlas" {
			t.Fatalf("recommendation task should be Atlas-owned: %+v", node.FactoryTask)
		}
		for _, want := range []string{"node_gate", "candidate_record", "rollback_record", "tests", "verification"} {
			if !containsString(node.FactoryTask.RequiredGates, want) {
				t.Fatalf("task %s missing required gate %q: %#v", node.FactoryTask.ID, want, node.FactoryTask.RequiredGates)
			}
		}
		if !containsString(node.FactoryTask.SafetyLimits, "no provider calls") ||
			!containsString(node.FactoryTask.SafetyLimits, "no credential inspection") ||
			!containsString(node.FactoryTask.SafetyLimits, "no direct main mutation") {
			t.Fatalf("task %s missing safety limits: %#v", node.FactoryTask.ID, node.FactoryTask.SafetyLimits)
		}
	}
	prompt, err := os.ReadFile(filepath.Join(outDir, "next-recommended-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(prompt), "You are AO Atlas") ||
		!strings.Contains(string(prompt), "Double the previous short batch") {
		t.Fatalf("next prompt missing operator-ready continuation text:\n%s", string(prompt))
	}
}

func TestMissionRecommendationsRejectShallowAndUnsafeBundles(t *testing.T) {
	dir := t.TempDir()
	shallowPath := filepath.Join(dir, "shallow.json")
	unsafePath := filepath.Join(dir, "unsafe.json")
	writeFeatureDepthBundle(t, shallowPath, 3, false)
	writeFeatureDepthBundle(t, unsafePath, 20, true)

	for _, tc := range []struct {
		name string
		path string
		want string
	}{
		{name: "shallow", path: shallowPath, want: "at least 20 tasks"},
		{name: "unsafe", path: unsafePath, want: "safe_to_execute must be false"},
	} {
		var out bytes.Buffer
		code := Run([]string{
			"mission", "recommendations", "import",
			"--recommendations", tc.path,
			"--target-instance", "demo-stack",
			"--min-tasks", "20",
			"--node-budget", "20",
			"--estimated-minutes", "90",
			"--out", filepath.Join(dir, tc.name+"-out"),
		}, &out, &out)
		if code == 0 {
			t.Fatalf("%s bundle was accepted", tc.name)
		}
		if !strings.Contains(out.String(), tc.want) {
			t.Fatalf("%s error missing %q: %s", tc.name, tc.want, out.String())
		}
	}
}

func TestMissionRecommendationsDefaultToTwoToThreeHourSupervisorWave(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	outDir := filepath.Join(dir, "recommendations-out")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("recommendation import failed: %s", out.String())
	}
	wave := mustLoadJSON[AtlasRecommendationWave](t, filepath.Join(outDir, "recommendation-wave.json"))
	if wave.TotalTasks != 40 || wave.NodeBudget != 40 {
		t.Fatalf("default wave should generate 40 nodes for continue-if-fast policy: %#v", wave)
	}
	if wave.MinimumTasks != 30 || wave.EstimatedMinutes != 120 {
		t.Fatalf("default wave should require 30 nodes and 120 minute floor: %#v", wave)
	}
	if wave.Supervisor == nil {
		t.Fatalf("default wave missing long-run supervisor: %#v", wave)
	}
	if wave.Supervisor.MinNodes != 30 ||
		wave.Supervisor.MinMinutes != 120 ||
		wave.Supervisor.MaxMinutes != 180 ||
		wave.Supervisor.ContinueIfFastTarget != 40 ||
		wave.Supervisor.ReturnOnlyWhen != "all_generated_nodes_done_or_30_nodes_done_or_true_hard_blocker" ||
		wave.Supervisor.CheckpointPolicy != "after_each_node_or_timed_interval" ||
		wave.Supervisor.EvidencePolicy != "node_gate_candidate_rollback_tests_verification_public_safety_promoter_command" ||
		wave.Supervisor.FinalReportContract != "ao.atlas.long-run-final-report.v0.2" {
		t.Fatalf("bad long-run supervisor: %#v", wave.Supervisor)
	}
	if wave.FinalResponseAllowed || wave.FinalResponseReason != "ready nodes or exact next actions remain" {
		t.Fatalf("default wave must deny final response while ready nodes remain: %#v", wave)
	}
	if wave.PromoterReadbackStatus != "required_not_bound" || wave.CommandReadbackStatus != "required_not_bound" || wave.PublicSafetyScanStatus != "required_pending_verification" {
		t.Fatalf("wave should require promoter, command, and public-safety readbacks: %#v", wave)
	}
	workgraph := mustLoadJSON[Workgraph](t, filepath.Join(outDir, "recommendation-workgraph.json"))
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) != 40 || len(state.ExecutableReadyNodeIDs) != 1 {
		t.Fatalf("expected 40 dependency-chained nodes with one executable-ready node, nodes=%d ready=%#v", len(workgraph.Nodes), state.ExecutableReadyNodeIDs)
	}
	prompt := wave.NextRecommendedPrompt
	for _, want := range []string{
		"Current state:",
		"Problem:",
		"Goal:",
		"Minimum work budget:",
		"Safety boundaries:",
		"Required work:",
		"Per-node requirements:",
		"Regression tests:",
		"Verification:",
		"Final response only after completion or true hard blocker:",
		"Target 2-3 hours",
		"Complete at least 30 bounded implementation/evidence nodes",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("v0.2 prompt missing section %q:\n%s", want, prompt)
		}
	}
}

func TestMissionRecommendationsRejectMixedOwnerDefaultWaveWithExactReadback(t *testing.T) {
	dir := t.TempDir()
	recommendationsPath := filepath.Join(dir, "feature-depth-recommendations.json")
	writeFeatureDepthBundle(t, recommendationsPath, 40, false)
	var bundle AOMissionFeatureDepthRecommendations
	if err := readJSONIfPossible(recommendationsPath, &bundle); err != nil {
		t.Fatal(err)
	}
	bundle.Tasks[39].Owner = "ao-foundry"
	if err := WriteJSON(recommendationsPath, bundle); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "import",
		"--recommendations", recommendationsPath,
		"--target-instance", "demo-stack",
		"--out", filepath.Join(dir, "out"),
	}, &out, &out)
	if code == 0 {
		t.Fatal("mixed-owner default wave was accepted")
	}
	if !strings.Contains(out.String(), "requires at least 30 AO Atlas-owned tasks and 40 tasks for continue-if-fast target") {
		t.Fatalf("mixed-owner error did not report exact readback: %s", out.String())
	}
}

func TestProductionReadinessExercisesMissionRecommendationsImport(t *testing.T) {
	root := repoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "production-readiness.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read production readiness script: %v", err)
	}
	script := string(content)
	for _, want := range []string{
		"mission recommendations import",
		"--recommendations examples/valid/ao-mission/feature-depth-recommendations.json",
		"--min-tasks 30",
		"--min-minutes 120",
		"--max-minutes 180",
		"--continue-if-fast-target 40",
		"recommendation-workgraph.json",
		"next-recommended-prompt.md",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("production readiness script missing recommendation coverage %q", want)
		}
	}
}

func writeFeatureDepthBundle(t *testing.T, path string, taskCount int, unsafe bool) {
	t.Helper()
	tasks := make([]map[string]string, 0, taskCount)
	for i := 1; i <= taskCount; i++ {
		tasks = append(tasks, map[string]string{
			"id":    "next-" + twoDigit(i),
			"owner": "ao-atlas",
			"task":  "Implement Atlas long-run recommendation node " + twoDigit(i) + " with tests, readback evidence, and continuation prompt coverage.",
		})
	}
	if err := WriteJSON(path, map[string]any{
		"schema":               "ao.mission.feature-depth-recommendations.v0.3",
		"mission_id":           "mission-long-wave",
		"status":               "ready",
		"minimum_tasks":        taskCount,
		"recommendation_count": taskCount,
		"tasks":                tasks,
		"safe_to_execute":      unsafe,
		"executes_work":        false,
		"approves_work":        false,
	}); err != nil {
		t.Fatal(err)
	}
}

func twoDigit(value int) string {
	if value < 10 {
		return "0" + string(rune('0'+value))
	}
	return "10"[:0] + string(rune('0'+value/10)) + string(rune('0'+value%10))
}
