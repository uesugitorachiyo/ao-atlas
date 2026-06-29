package atlas

import (
	"bytes"
	"os"
	"path/filepath"
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
