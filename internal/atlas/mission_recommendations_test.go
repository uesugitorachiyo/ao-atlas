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
		"--min-tasks 20",
		"--estimated-minutes 90",
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
		"minimum_tasks":        20,
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
