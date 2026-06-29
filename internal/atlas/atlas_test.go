package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstanceInitValidateAndRegistry(t *testing.T) {
	dir := t.TempDir()
	instancePath := filepath.Join(dir, "instance.json")
	var out bytes.Buffer
	code := Run([]string{"instance", "init", "--id", "demo", "--state-root", ".atlas-local/state", "--toolchain-root", "../toolchain", "--out", instancePath}, &out, &out)
	if code != 0 {
		t.Fatalf("init failed: %s", out.String())
	}
	code = Run([]string{"instance", "validate", "--instance", instancePath}, &out, &out)
	if code != 0 {
		t.Fatalf("validate failed: %s", out.String())
	}
	registryPath := filepath.Join(dir, "foundry-registry.json")
	code = Run([]string{"instance", "registry", "emit", "--instance", instancePath, "--out", registryPath}, &out, &out)
	if code != 0 {
		t.Fatalf("registry failed: %s", out.String())
	}
	if _, err := os.Stat(registryPath); err != nil {
		t.Fatal(err)
	}
}

func TestProductionReadinessOpsWorkflowRunsAtlasReadinessGate(t *testing.T) {
	root := repoRoot(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "production-readiness-ops.yml")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read production-readiness-ops workflow: %v", err)
	}
	workflow := string(content)
	for _, want := range []string{
		"name: Production Readiness Ops",
		"workflow_dispatch:",
		"contents: read",
		"actions/checkout@v7",
		"actions/setup-go@v6",
		"scripts/production-readiness.sh",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("production-readiness-ops workflow missing %q", want)
		}
	}
	for _, forbidden := range []string{"gh release", "git push", "upload-artifact", "OPENAI_" + "API_KEY", "ANTHROPIC_" + "API_KEY"} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("production-readiness-ops workflow contains forbidden capability %q", forbidden)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestInstanceDoctorValidatesRootsAndRegistryParity(t *testing.T) {
	dir := t.TempDir()
	instancePath := filepath.Join(dir, "instance.json")
	registryPath := filepath.Join(dir, "registry.json")
	doctorPath := filepath.Join(dir, "doctor.json")
	var out bytes.Buffer
	if code := Run([]string{"instance", "init", "--id", "demo", "--state-root", ".atlas-local/state", "--toolchain-root", "../toolchain", "--out", instancePath}, &out, &out); code != 0 {
		t.Fatalf("init failed: %s", out.String())
	}
	if code := Run([]string{"instance", "registry", "emit", "--instance", instancePath, "--out", registryPath}, &out, &out); code != 0 {
		t.Fatalf("registry failed: %s", out.String())
	}
	if code := Run([]string{"instance", "doctor", "--instance", instancePath, "--registry", registryPath, "--out", doctorPath}, &out, &out); code != 0 {
		t.Fatalf("doctor failed: %s", out.String())
	}
	report, err := LoadJSON[InstanceDoctorReport](doctorPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateInstanceDoctorReport(report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ready" || report.InstanceID != "demo" {
		t.Fatalf("unexpected doctor report: %#v", report)
	}
}

func TestInstanceDoctorRejectsRegistryParityMismatch(t *testing.T) {
	dir := t.TempDir()
	instancePath := filepath.Join(dir, "instance.json")
	registryPath := filepath.Join(dir, "registry.json")
	var out bytes.Buffer
	if code := Run([]string{"instance", "init", "--id", "demo", "--state-root", ".atlas-local/state", "--toolchain-root", "../toolchain", "--out", instancePath}, &out, &out); code != 0 {
		t.Fatalf("init failed: %s", out.String())
	}
	if err := WriteJSON(registryPath, AtlasRegistry{
		ContractVersion: "ao.atlas.foundry-registry.v0.1",
		InstanceID:      "other-demo",
		ToolchainRoot:   "../toolchain",
		Roots:           map[string]string{"mission": ".atlas-local/state/demo/mission"},
		SchedulesWork:   false,
		ExecutesWork:    false,
	}); err != nil {
		t.Fatal(err)
	}
	code := Run([]string{"instance", "doctor", "--instance", instancePath, "--registry", registryPath, "--out", filepath.Join(dir, "doctor.json")}, &out, &out)
	if code == 0 {
		t.Fatal("expected registry parity mismatch to fail")
	}
	if !strings.Contains(out.String(), "registry instance_id") {
		t.Fatalf("expected registry parity error, got %s", out.String())
	}
}

func TestIntakeUnderspecifiedEmitsBlueprintRequest(t *testing.T) {
	intake := Intake{ContractVersion: IntakeContract, ID: "short", BroadPrompt: "fix it"}
	request, err := ValidateIntake(intake)
	if err != nil {
		t.Fatal(err)
	}
	if request.Status != "blueprint_required" {
		t.Fatalf("expected blueprint request, got %#v", request)
	}
	if len(request.Missing) == 0 {
		t.Fatal("expected missing fields")
	}
	if err := ValidateBlueprintRequest(request); err != nil {
		t.Fatal(err)
	}
}

func TestMissionStatusSummarizesIntakeWorkgraphAndRunLinks(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "mission-status.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph-completed.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status failed: %s", out.String())
	}
	status, err := LoadJSON[MissionStatus](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateMissionStatus(status); err != nil {
		t.Fatal(err)
	}
	if status.IntakeID != "atlas-intake-demo" || status.WorkgraphID != "atlas-readiness-workgraph" {
		t.Fatalf("unexpected mission status: %#v", status)
	}
	if status.CompletionStatus != "completed" || status.NodeCounts["completed"] != 2 {
		t.Fatalf("expected completed mission status, got %#v", status)
	}
}

func TestMissionStatusReportsBlockedWhenRunLinkBlocked(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "mission-status.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"),
		"--run-link", filepath.Join("..", "..", "examples", "invalid", "run-link-blocked.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status failed: %s", out.String())
	}
	status, err := LoadJSON[MissionStatus](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if status.CompletionStatus != "blocked" {
		t.Fatalf("expected blocked mission status, got %#v", status)
	}
}

func TestBlueprintRequestFixtureIsValidAndPublicSafe(t *testing.T) {
	request, err := LoadJSON[BlueprintRequest](filepath.Join("..", "..", "examples", "valid", "blueprint-request.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateBlueprintRequest(request); err != nil {
		t.Fatal(err)
	}
	if request.Status != "blueprint_required" {
		t.Fatalf("expected blueprint_required, got %s", request.Status)
	}
}

func TestContextPackRejectsAbsoluteLocalPath(t *testing.T) {
	pack := ContextPack{
		ContractVersion: ContextPackContract,
		ID:              "bad-pack",
		TaskID:          "task",
		BudgetBytes:     4096,
		SourceRefs:      []SourceRef{{Ref: "/" + "absolute/context.md", Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},
		Summaries:       []string{"summary"},
		MissingProtocol: "ask Blueprint for missing context",
	}
	if err := ValidateContextPack(pack, 0); err == nil || !strings.Contains(err.Error(), "absolute local path") {
		t.Fatalf("expected absolute path rejection, got %v", err)
	}
}

func TestContextPackRepackWritesBoundedPackForNeedsContext(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "context-pack.json")
	var out bytes.Buffer
	code := Run([]string{
		"context-pack", "repack",
		"--task", filepath.Join("..", "..", "examples", "valid", "factory-task.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-needs-context.json"),
		"--source-ref", "docs/sdd/AO-ATLAS-CONTEXT-PACKS.md",
		"--source-digest", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"--budget", "4096",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("repack failed: %s", out.String())
	}
	pack, err := LoadJSON[ContextPack](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateContextPack(pack, 0); err != nil {
		t.Fatal(err)
	}
	if pack.TaskID != "atlas-readiness-task" || len(pack.SourceRefs) != 1 {
		t.Fatalf("unexpected context pack: %#v", pack)
	}
}

func TestContextPackRepackRejectsCompletedRunLink(t *testing.T) {
	assertContextPackRepackFails(t, filepath.Join("..", "..", "examples", "valid", "run-link.json"), "blocked or failed")
}

func TestContextPackRepackRejectsBlockedRunLinkWithoutNeedsContext(t *testing.T) {
	assertContextPackRepackFails(t, filepath.Join("..", "..", "examples", "invalid", "run-link-blocked.json"), "needs_context")
}

func assertContextPackRepackFails(t *testing.T, runLinkPath, want string) {
	t.Helper()
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{
		"context-pack", "repack",
		"--task", filepath.Join("..", "..", "examples", "valid", "factory-task.json"),
		"--run-link", runLinkPath,
		"--source-ref", "docs/sdd/AO-ATLAS-CONTEXT-PACKS.md",
		"--source-digest", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"--budget", "4096",
		"--out", filepath.Join(dir, "context-pack.json"),
	}, &out, &out)
	if code == 0 {
		t.Fatalf("expected context-pack repack to fail for %s", runLinkPath)
	}
	if !strings.Contains(out.String(), want) {
		t.Fatalf("expected error containing %q, got %s", want, out.String())
	}
}

func TestWorkgraphNextSkipsBlockedAndDependencyIncomplete(t *testing.T) {
	wg := fixtureWorkgraph()
	node, ok := NextReadyNode(wg)
	if !ok {
		t.Fatal("expected ready node")
	}
	if node.ID != "task-ready" {
		t.Fatalf("expected task-ready, got %s", node.ID)
	}
}

func TestFoundryHandoffUsesReadyTasksOnly(t *testing.T) {
	wg := fixtureWorkgraph()
	handoff := BuildFoundryHandoff(wg)
	if err := ValidateFoundryHandoff(handoff); err != nil {
		t.Fatal(err)
	}
	if len(handoff.Tasks) != 2 {
		t.Fatalf("expected two ready tasks, got %d", len(handoff.Tasks))
	}
}

func TestFoundryImportWritesTaskFixturesForReadyNodes(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "foundry-import")
	var out bytes.Buffer
	code := Run([]string{"foundry", "import", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--out", outDir}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry import failed: %s", out.String())
	}
	manifest, err := LoadJSON[FoundryImport](filepath.Join(outDir, "foundry-import.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryImport(manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Tasks) != 1 {
		t.Fatalf("expected one dependency-ready task fixture, got %#v", manifest.Tasks)
	}
	if manifest.Tasks[0].TaskID != "atlas-readiness-task" {
		t.Fatalf("unexpected task fixture: %#v", manifest.Tasks[0])
	}
	if _, err := os.Stat(filepath.Join(outDir, manifest.Tasks[0].Path)); err != nil {
		t.Fatal(err)
	}
	if manifest.SchedulesWork || manifest.ExecutesWork || manifest.ApprovesWork {
		t.Fatalf("foundry import must be fixture-only readback: %#v", manifest)
	}
}

func TestFoundryImportRejectsNoReadyNodes(t *testing.T) {
	wg := fixtureWorkgraph()
	for i := range wg.Nodes {
		if wg.Nodes[i].Status == "ready" {
			wg.Nodes[i].Status = "blocked"
			wg.Nodes[i].Blockers = []string{"not ready"}
		}
	}
	_, err := BuildFoundryImport(wg)
	if err == nil || !strings.Contains(err.Error(), "tasks must not be empty") {
		t.Fatalf("expected no ready tasks rejection, got %v", err)
	}
}

func TestFoundryImportRejectsSameInputAndOutputPath(t *testing.T) {
	var out bytes.Buffer
	path := filepath.Join("..", "..", "examples", "valid", "workgraph.json")
	code := Run([]string{"foundry", "import", "--workgraph", path, "--out", path}, &out, &out)
	if code == 0 {
		t.Fatal("expected same input/output path to fail")
	}
	if !strings.Contains(out.String(), "overwrite input") {
		t.Fatalf("expected overwrite error, got %s", out.String())
	}
}

func TestFoundryImportFixtureIsValid(t *testing.T) {
	foundryImport, err := LoadJSON[FoundryImport](filepath.Join("..", "..", "examples", "valid", "foundry-import.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryImport(foundryImport); err != nil {
		t.Fatal(err)
	}
}

func TestFoundryImportRejectsExecutionAuthority(t *testing.T) {
	foundryImport, err := LoadJSON[FoundryImport](filepath.Join("..", "..", "examples", "invalid", "foundry-import-executes-work.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryImport(foundryImport); err == nil || !strings.Contains(err.Error(), "executes_work") {
		t.Fatalf("expected executes_work rejection, got %v", err)
	}
}

func TestFoundryRoundtripSmokeValidatesFoundryImport(t *testing.T) {
	script, err := os.ReadFile(filepath.Join("..", "..", "scripts", "atlas-foundry-roundtrip-smoke.sh"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(script)
	for _, want := range []string{
		"foundry import",
		"--workgraph examples/valid/workgraph.json",
		"foundry atlas import validate",
		"foundry_import_validation",
		"foundry atlas readback",
		"FOUNDRY_READBACK",
		"foundry_readback",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("roundtrip smoke missing %q", want)
		}
	}
}

func TestFactoryMaterializeDryRunWritesBoundedSkeleton(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "factory-materialization")
	var out bytes.Buffer
	code := Run([]string{"factory", "materialize", "--task", filepath.Join("..", "..", "examples", "valid", "factory-task.json"), "--out", outDir, "--dry-run"}, &out, &out)
	if code != 0 {
		t.Fatalf("materialize failed: %s", out.String())
	}
	for _, rel := range []string{"README.md", "task.json", "verification.txt", "materialization.json", filepath.Join("evidence", "README.md"), filepath.Join("context", "README.md")} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	manifest, err := LoadJSON[FactoryMaterialization](filepath.Join(outDir, "materialization.json"))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Mode != "dry_run" || manifest.ExecutesWork {
		t.Fatalf("unexpected materialization manifest: %#v", manifest)
	}
	if strings.Contains(manifest.OutputRoot, string(os.PathSeparator)) {
		t.Fatalf("manifest must not record local output path: %#v", manifest)
	}
}

func TestFactoryMaterializeRequiresDryRun(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{"factory", "materialize", "--task", filepath.Join("..", "..", "examples", "valid", "factory-task.json"), "--out", filepath.Join(dir, "factory-materialization")}, &out, &out)
	if code == 0 {
		t.Fatal("expected materialize without --dry-run to fail")
	}
	if !strings.Contains(out.String(), "--dry-run") {
		t.Fatalf("expected dry-run error, got %s", out.String())
	}
}

func TestWorkgraphMaterializeNextDryRunUsesNextReadyTask(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "next-materialization")
	var out bytes.Buffer
	code := Run([]string{"workgraph", "materialize-next", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--out", outDir, "--dry-run"}, &out, &out)
	if code != 0 {
		t.Fatalf("materialize-next failed: %s", out.String())
	}
	manifest, err := LoadJSON[FactoryMaterialization](filepath.Join(outDir, "materialization.json"))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.TaskID != "atlas-readiness-task" {
		t.Fatalf("expected atlas-readiness-task, got %s", manifest.TaskID)
	}
	if !strings.Contains(out.String(), "node=readiness-ready") {
		t.Fatalf("expected node readback, got %s", out.String())
	}
}

func TestWorkgraphMaterializeNextRequiresDryRun(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{"workgraph", "materialize-next", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--out", filepath.Join(dir, "next-materialization")}, &out, &out)
	if code == 0 {
		t.Fatal("expected materialize-next without --dry-run to fail")
	}
	if !strings.Contains(out.String(), "--dry-run") {
		t.Fatalf("expected dry-run error, got %s", out.String())
	}
}

func TestWorkgraphCompleteWritesNewCompletedWorkgraph(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "workgraph-completed.json")
	var out bytes.Buffer
	inputPath := filepath.Join("..", "..", "examples", "valid", "workgraph.json")
	code := Run([]string{"workgraph", "complete", "--workgraph", inputPath, "--run-link", filepath.Join("..", "..", "examples", "valid", "run-link.json"), "--out", outPath}, &out, &out)
	if code != 0 {
		t.Fatalf("complete failed: %s", out.String())
	}
	completed, err := LoadJSON[Workgraph](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Nodes[1].Status != "completed" {
		t.Fatalf("expected completed node, got %#v", completed.Nodes[1])
	}
	original, err := LoadJSON[Workgraph](inputPath)
	if err != nil {
		t.Fatal(err)
	}
	if original.Nodes[1].Status != "ready" {
		t.Fatalf("input workgraph was modified: %#v", original.Nodes[1])
	}
	if !strings.Contains(out.String(), "node=readiness-ready") {
		t.Fatalf("expected node readback, got %s", out.String())
	}
}

func TestWorkgraphCompleteRejectsBlockedRunLink(t *testing.T) {
	assertWorkgraphCompleteFails(t, filepath.Join("..", "..", "examples", "valid", "workgraph.json"), filepath.Join("..", "..", "examples", "invalid", "run-link-blocked.json"), "completed")
}

func TestWorkgraphCompleteRejectsMissingNode(t *testing.T) {
	assertWorkgraphCompleteFails(t, filepath.Join("..", "..", "examples", "valid", "workgraph.json"), filepath.Join("..", "..", "examples", "invalid", "run-link-missing-node.json"), "matching")
}

func TestWorkgraphCompleteRejectsIncompleteDependency(t *testing.T) {
	assertWorkgraphCompleteFails(t, filepath.Join("..", "..", "examples", "invalid", "workgraph-complete-incomplete-dependency.json"), filepath.Join("..", "..", "examples", "valid", "run-link.json"), "dependencies")
}

func TestWorkgraphCompleteRejectsSameInputAndOutputPath(t *testing.T) {
	var out bytes.Buffer
	path := filepath.Join("..", "..", "examples", "valid", "workgraph.json")
	code := Run([]string{"workgraph", "complete", "--workgraph", path, "--run-link", filepath.Join("..", "..", "examples", "valid", "run-link.json"), "--out", path}, &out, &out)
	if code == 0 {
		t.Fatal("expected same input/output path to fail")
	}
	if !strings.Contains(out.String(), "overwrite input") {
		t.Fatalf("expected overwrite error, got %s", out.String())
	}
}

func TestWorkgraphRepairPlanEmitsRepairTaskForBlockedRunLink(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "repair-plan.json")
	var out bytes.Buffer
	code := Run([]string{"workgraph", "repair-plan", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--run-link", filepath.Join("..", "..", "examples", "invalid", "run-link-blocked.json"), "--out", outPath}, &out, &out)
	if code != 0 {
		t.Fatalf("repair-plan failed: %s", out.String())
	}
	plan, err := LoadJSON[WorkgraphRepairPlan](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkgraphRepairPlan(plan); err != nil {
		t.Fatal(err)
	}
	if plan.TaskID != "atlas-readiness-task" || len(plan.RepairTasks) != 1 {
		t.Fatalf("unexpected repair plan: %#v", plan)
	}
	if plan.RepairTasks[0].ID != "repair-atlas-readiness-task" {
		t.Fatalf("unexpected repair task: %#v", plan.RepairTasks[0])
	}
}

func TestWorkgraphRepairPlanRejectsCompletedRunLink(t *testing.T) {
	assertWorkgraphRepairPlanFails(t, filepath.Join("..", "..", "examples", "valid", "run-link.json"), "blocked or failed")
}

func TestWorkgraphRepairPlanRejectsMissingNode(t *testing.T) {
	assertWorkgraphRepairPlanFails(t, filepath.Join("..", "..", "examples", "invalid", "run-link-missing-node-blocked.json"), "matching")
}

func assertWorkgraphRepairPlanFails(t *testing.T, runLinkPath, want string) {
	t.Helper()
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{"workgraph", "repair-plan", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--run-link", runLinkPath, "--out", filepath.Join(dir, "repair-plan.json")}, &out, &out)
	if code == 0 {
		t.Fatalf("expected repair-plan to fail for %s", runLinkPath)
	}
	if !strings.Contains(out.String(), want) {
		t.Fatalf("expected error containing %q, got %s", want, out.String())
	}
}

func assertWorkgraphCompleteFails(t *testing.T, workgraphPath, runLinkPath, want string) {
	t.Helper()
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{"workgraph", "complete", "--workgraph", workgraphPath, "--run-link", runLinkPath, "--out", filepath.Join(dir, "completed.json")}, &out, &out)
	if code == 0 {
		t.Fatalf("expected complete to fail for %s", runLinkPath)
	}
	if !strings.Contains(out.String(), want) {
		t.Fatalf("expected error containing %q, got %s", want, out.String())
	}
}

func TestRunLinkAttachWritesDigestBoundPublicSafeLink(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "run-link.json")
	var out bytes.Buffer
	code := Run([]string{
		"run-link", "attach",
		"--task-id", "atlas-readiness-task",
		"--status", "completed",
		"--evidence", "foundry=evidence/foundry/atlas-readiness.json",
		"--evidence", "forge=evidence/forge/atlas-readiness.json",
		"--evidence", "ao2=evidence/ao2/atlas-readiness.json",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-link attach failed: %s", out.String())
	}
	link, err := LoadJSON[RunLink](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateRunLink(link); err != nil {
		t.Fatal(err)
	}
	if link.TaskID != "atlas-readiness-task" || link.Status != "completed" {
		t.Fatalf("unexpected run link: %#v", link)
	}
	if link.Evidence["ao2"] != "evidence/ao2/atlas-readiness.json" {
		t.Fatalf("expected ao2 evidence path, got %#v", link.Evidence)
	}
	if !strings.HasPrefix(link.Digest, "sha256:") || link.Digest == "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatalf("expected computed digest, got %s", link.Digest)
	}
}

func TestRunLinkAttachRejectsPrivateEvidencePath(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{
		"run-link", "attach",
		"--task-id", "atlas-readiness-task",
		"--status", "completed",
		"--evidence", "ao2=/" + "Users/example/private.json",
		"--out", filepath.Join(dir, "run-link.json"),
	}, &out, &out)
	if code == 0 {
		t.Fatal("expected private evidence path to fail")
	}
	if !strings.Contains(out.String(), "private or machine-local path") {
		t.Fatalf("expected public-safety error, got %s", out.String())
	}
}

func fixtureWorkgraph() Workgraph {
	baseTask := FactoryTask{
		ContractVersion:   FactoryTaskContract,
		ID:                "factory-task",
		Objective:         "Create bounded AO Atlas task material.",
		TargetFactoryRepo: "ao-foundry",
		FactoryFolder:     "factory/atlas-demo",
		Acceptance:        []string{"evidence exists"},
		NonGoals:          []string{"do not execute"},
		WriteScope:        []string{"factory/atlas-demo"},
		Verification:      []string{"go test ./..."},
		RequiredEvidence:  []string{"summary.json"},
		SafetyLimits:      []string{"no provider calls"},
	}
	return Workgraph{
		ContractVersion: WorkgraphContract,
		ID:              "wg",
		TargetInstance:  "demo",
		Nodes: []WorkgraphNode{
			{ID: "done", Status: "completed", FactoryTask: baseTask},
			{ID: "task-ready", Status: "ready", Dependencies: []string{"done"}, FactoryTask: baseTask},
			{ID: "task-blocked", Status: "blocked", Blockers: []string{"needs Blueprint"}, FactoryTask: baseTask},
			{ID: "task-ready-2", Status: "ready", FactoryTask: baseTask},
		},
	}
}
