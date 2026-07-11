package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveCompactionResumePromptPreservesLeaseAndActiveNode(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-17")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-16", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-16", "workgraph-after.json")
	recordedPath := filepath.Join(nodeDir, "compaction-resume-prompt.json")
	recordedPromptPath := filepath.Join(nodeDir, "compaction-resume-prompt.md")
	outDir := t.TempDir()
	outFixture := filepath.Join(outDir, "compaction-resume-prompt.json")
	outPrompt := filepath.Join(outDir, "compaction-resume-prompt.md")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-prompt",
		"--source-readback", sourceReadback,
		"--workgraph", sourceWorkgraph,
		"--lease-start", filepath.Join(waveRoot, "lease-start.json"),
		"--evidence-root", filepath.ToSlash(filepath.Join("docs", "evidence", "ao-atlas-feature-depth-wave-v01")),
		"--node-id", "mission-recommendation-feature-depth-next-wave-17",
		"--expected-next-node-after-completion", "mission-recommendation-feature-depth-next-wave-18",
		"--prompt-out", outPrompt,
		"--fixture-out", outFixture,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-prompt command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=generated") ||
		!strings.Contains(out.String(), "first_executable_node=mission-recommendation-feature-depth-next-wave-17") ||
		!strings.Contains(out.String(), "elapsed_minutes=375") {
		t.Fatalf("compaction-resume-prompt output missing resume state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outFixture)
	generated["prompt_path"] = recorded["prompt_path"]
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("compaction resume prompt fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	promptBytes, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatalf("read recorded prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"You are AO Atlas, resuming the AO Atlas feature-depth wave after context compaction.",
		"Current readback: `docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-16/recommendation-readback-after.json`",
		"Completed nodes: 16 / 40",
		"Ready nodes: 24",
		"Next executable node: `mission-recommendation-feature-depth-next-wave-17`",
		"Elapsed minutes: `375`",
		"Final response allowed: `false`",
		"Emit Foundry import for mission-recommendation-feature-depth-next-wave-17 and execute exactly one active node.",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("compaction resume prompt missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "mission-recommendation-feature-depth-next-wave-16 and execute") {
		t.Fatalf("compaction resume prompt must not restart completed node 16:\n%s", prompt)
	}
}

func TestFeatureDepthWaveV02CompactionResumePromptPreservesLeaseAndActiveNode(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-17")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-16", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-16", "workgraph-after.json")
	recordedPath := filepath.Join(nodeDir, "compaction-resume-prompt.json")
	recordedPromptPath := filepath.Join(nodeDir, "compaction-resume-prompt.md")
	outDir := t.TempDir()
	outFixture := filepath.Join(outDir, "compaction-resume-prompt.json")
	outPrompt := filepath.Join(outDir, "compaction-resume-prompt.md")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-prompt",
		"--source-readback", sourceReadback,
		"--workgraph", sourceWorkgraph,
		"--lease-start", filepath.Join(waveRoot, "lease-start.json"),
		"--evidence-root", filepath.ToSlash(filepath.Join("docs", "evidence", "ao-atlas-feature-depth-wave-v02")),
		"--node-id", "mission-recommendation-feature-depth-next-wave-17",
		"--expected-next-node-after-completion", "mission-recommendation-feature-depth-next-wave-18",
		"--prompt-out", outPrompt,
		"--fixture-out", outFixture,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-prompt command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=generated") ||
		!strings.Contains(out.String(), "first_executable_node=mission-recommendation-feature-depth-next-wave-17") ||
		!strings.Contains(out.String(), "final_response_allowed=false") {
		t.Fatalf("compaction-resume-prompt output missing v02 resume state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outFixture)
	generated["prompt_path"] = recorded["prompt_path"]
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 compaction resume prompt fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	promptBytes, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatalf("read recorded prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"You are AO Atlas, resuming the AO Atlas feature-depth wave after context compaction.",
		"Current readback: `docs/evidence/ao-atlas-feature-depth-wave-v02/nodes/mission-recommendation-feature-depth-next-wave-16/recommendation-readback-after.json`",
		"Completed nodes: 16 / 40",
		"Ready nodes: 24",
		"Next executable node: `mission-recommendation-feature-depth-next-wave-17`",
		"Final response allowed: `false`",
		"Emit Foundry import for mission-recommendation-feature-depth-next-wave-17 and execute exactly one active node.",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("v02 compaction resume prompt missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "mission-recommendation-feature-depth-next-wave-16 and execute") {
		t.Fatalf("v02 compaction resume prompt must not restart completed node 16:\n%s", prompt)
	}
}

func TestMonth3FinalClosureCompactionResumePromptPreservesNode14Continuation(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-14-compaction-resume-generator")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-13-rollback-replay-negative", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-13-rollback-replay-negative", "workgraph-after.json")
	checkpointReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-13-rollback-replay-negative", "checkpoint-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "compaction-resume-prompt.json")
	recordedPromptPath := filepath.Join(nodeDir, "compaction-resume-prompt.md")
	outDir := t.TempDir()
	outFixture := filepath.Join(outDir, "compaction-resume-prompt.json")
	outPrompt := filepath.Join(outDir, "compaction-resume-prompt.md")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-prompt",
		"--source-readback", sourceReadback,
		"--workgraph", sourceWorkgraph,
		"--lease-start", filepath.Join(waveRoot, "lease-start.json"),
		"--checkpoint-readback", checkpointReadback,
		"--evidence-root", filepath.ToSlash(filepath.Join("docs", "evidence", "ao-m3-final-closure-v01")),
		"--node-id", "mission-recommendation-month3-final-closure-14-compaction-resume-generator",
		"--expected-next-node-after-completion", "mission-recommendation-month3-final-closure-15-architecture-source-truth",
		"--prompt-out", outPrompt,
		"--fixture-out", outFixture,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-prompt command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=generated") ||
		!strings.Contains(out.String(), "first_executable_node=mission-recommendation-month3-final-closure-14-compaction-resume-generator") ||
		!strings.Contains(out.String(), "final_response_allowed=false") {
		t.Fatalf("compaction-resume-prompt output missing Month 3 resume state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outFixture)
	generated["prompt_path"] = recorded["prompt_path"]
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 compaction resume prompt fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	promptBytes, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatalf("read recorded prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"You are AO Atlas, resuming the AO Atlas final-closure consolidation wave after context compaction.",
		"Current readback: `docs/evidence/ao-m3-final-closure-v01/nodes/mission-recommendation-month3-final-closure-13-rollback-replay-negative/recommendation-readback-after.json`",
		"Checkpoint readback: `docs/evidence/ao-m3-final-closure-v01/nodes/mission-recommendation-month3-final-closure-13-rollback-replay-negative/checkpoint-readback-after.json`",
		"Completed nodes: 13 / 30",
		"Ready nodes: 17",
		"Next executable node: `mission-recommendation-month3-final-closure-14-compaction-resume-generator`",
		"Final response allowed: `false`",
		"Emit Foundry import for mission-recommendation-month3-final-closure-14-compaction-resume-generator and execute exactly one active node.",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("Month 3 compaction resume prompt missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "mission-recommendation-month3-final-closure-13-rollback-replay-negative and execute") {
		t.Fatalf("compaction resume prompt must not restart completed node 13:\n%s", prompt)
	}
}

func TestFeatureDepthWaveCompactionResumePromptUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-17", "compaction-resume-prompt.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.compaction-resume-prompt.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:compaction-resume-prompt" {
		t.Fatalf("expected typed compaction resume prompt validator, got %s", validator)
	}
}

func TestRefactoringWaveCompactionResumePromptPreservesNextActionAndFinalGateDenial(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-refactoring-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "refactoring-next-wave-27")
	sourceReadback := filepath.Join(waveRoot, "nodes", "refactoring-next-wave-26", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "refactoring-next-wave-26", "workgraph-after.json")
	recordedPath := filepath.Join(nodeDir, "compaction-resume-prompt.json")
	recordedPromptPath := filepath.Join(nodeDir, "compaction-resume-prompt.md")
	outDir := t.TempDir()
	outFixture := filepath.Join(outDir, "compaction-resume-prompt.json")
	outPrompt := filepath.Join(outDir, "compaction-resume-prompt.md")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-prompt",
		"--source-readback", sourceReadback,
		"--workgraph", sourceWorkgraph,
		"--lease-start", filepath.Join(waveRoot, "lease-start.json"),
		"--evidence-root", filepath.ToSlash(filepath.Join("docs", "evidence", "ao-atlas-refactoring-wave-v01")),
		"--node-id", "refactoring-next-wave-27",
		"--expected-next-node-after-completion", "refactoring-next-wave-28",
		"--prompt-out", outPrompt,
		"--fixture-out", outFixture,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-prompt command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=generated") ||
		!strings.Contains(out.String(), "first_executable_node=refactoring-next-wave-27") {
		t.Fatalf("compaction-resume-prompt output missing refactoring resume state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outFixture)
	generated["prompt_path"] = recorded["prompt_path"]
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("refactoring compaction resume prompt fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	promptBytes, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatalf("read recorded prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"You are AO Atlas, resuming the AO Atlas refactoring wave after context compaction.",
		"Completed nodes: 26 / 40",
		"Ready nodes: 14",
		"Next executable node: `refactoring-next-wave-27`",
		"Exact next action: Add prompt compaction resume fixtures that preserve next node, exact action, and final gate denial.",
		"Return gate: `final_response_denied_ready_work_remains`",
		"Final response allowed: `false`",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("refactoring compaction resume prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestP0BContractConvergenceCompactionResumePromptPreservesNode25Continuation(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-p0b-contract-convergence-25")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-p0b-contract-convergence-24", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "mission-recommendation-p0b-contract-convergence-24", "workgraph-after-complete.json")
	checkpointReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-p0b-contract-convergence-24", "checkpoint-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "compaction-resume-prompt.json")
	recordedPromptPath := filepath.Join(nodeDir, "compaction-resume-prompt.md")
	outDir := t.TempDir()
	outFixture := filepath.Join(outDir, "compaction-resume-prompt.json")
	outPrompt := filepath.Join(outDir, "compaction-resume-prompt.md")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-prompt",
		"--source-readback", sourceReadback,
		"--workgraph", sourceWorkgraph,
		"--lease-start", filepath.Join(waveRoot, "lease-start.json"),
		"--checkpoint-readback", checkpointReadback,
		"--evidence-root", filepath.ToSlash(filepath.Join("docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01")),
		"--node-id", "mission-recommendation-p0b-contract-convergence-25",
		"--expected-next-node-after-completion", "mission-recommendation-p0b-contract-convergence-26",
		"--prompt-out", outPrompt,
		"--fixture-out", outFixture,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-prompt command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=generated") ||
		!strings.Contains(out.String(), "first_executable_node=mission-recommendation-p0b-contract-convergence-25") ||
		!strings.Contains(out.String(), "final_response_allowed=false") {
		t.Fatalf("compaction-resume-prompt output missing P0-B resume state: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outFixture)
	generated["prompt_path"] = recorded["prompt_path"]
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("P0-B compaction resume prompt fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	promptBytes, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatalf("read recorded prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"You are AO Atlas, resuming the AO Atlas recommendation wave after context compaction.",
		"Current readback: `docs/evidence/ao-stack-p0b-contract-convergence-wave-v01/nodes/mission-recommendation-p0b-contract-convergence-24/recommendation-readback-after.json`",
		"Completed nodes: 24 / 30",
		"Ready nodes: 6",
		"Next executable node: `mission-recommendation-p0b-contract-convergence-25`",
		"Checkpoint count: `24`",
		"Final response allowed: `false`",
		"Emit Foundry import for mission-recommendation-p0b-contract-convergence-25 and execute exactly one active node.",
		"Do not produce a final response while ready nodes or exact next action remain.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("P0-B compaction resume prompt missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "mission-recommendation-p0b-contract-convergence-24 and execute") {
		t.Fatalf("P0-B compaction resume prompt must not restart completed node 24:\n%s", prompt)
	}
}

func TestFeatureDepthWaveCompactionResumePromptBindsCheckpointDigest(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-19")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-18", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-18", "workgraph-after.json")
	sourceCheckpoint := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-18", "checkpoint-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "checkpoint-digest-resume-prompt.json")
	recordedPromptPath := filepath.Join(nodeDir, "checkpoint-digest-resume-prompt.md")
	outDir := t.TempDir()
	outFixture := filepath.Join(outDir, "checkpoint-digest-resume-prompt.json")
	outPrompt := filepath.Join(outDir, "checkpoint-digest-resume-prompt.md")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-prompt",
		"--source-readback", sourceReadback,
		"--workgraph", sourceWorkgraph,
		"--lease-start", filepath.Join(waveRoot, "lease-start.json"),
		"--checkpoint-readback", sourceCheckpoint,
		"--evidence-root", filepath.ToSlash(filepath.Join("docs", "evidence", "ao-atlas-feature-depth-wave-v01")),
		"--node-id", "mission-recommendation-feature-depth-next-wave-19",
		"--expected-next-node-after-completion", "mission-recommendation-feature-depth-next-wave-20",
		"--prompt-out", outPrompt,
		"--fixture-out", outFixture,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-prompt command failed: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outFixture)
	generated["prompt_path"] = recorded["prompt_path"]
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("checkpoint digest resume prompt fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded["checkpoint_readback_path"] != "docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-18/checkpoint-readback-after.json" {
		t.Fatalf("checkpoint readback path not bound: %#v", recorded["checkpoint_readback_path"])
	}
	checkpointDigest, ok := recorded["checkpoint_readback_digest"].(string)
	if !ok || !digestPattern.MatchString(checkpointDigest) {
		t.Fatalf("checkpoint digest missing or invalid: %#v", recorded["checkpoint_readback_digest"])
	}
	promptBytes, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatalf("read recorded prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"Checkpoint readback: `docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-18/checkpoint-readback-after.json`",
		"Checkpoint readback digest: `" + checkpointDigest + "`",
		"Next executable node: `mission-recommendation-feature-depth-next-wave-19`",
		"Do not produce a final response while ready nodes or exact next action remain.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("checkpoint digest resume prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestFeatureDepthWaveV02CompactionResumePromptBindsCheckpointDigest(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-19")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-18", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-18", "workgraph-after.json")
	sourceCheckpoint := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-18", "checkpoint-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "checkpoint-digest-resume-prompt.json")
	recordedPromptPath := filepath.Join(nodeDir, "checkpoint-digest-resume-prompt.md")
	outDir := t.TempDir()
	outFixture := filepath.Join(outDir, "checkpoint-digest-resume-prompt.json")
	outPrompt := filepath.Join(outDir, "checkpoint-digest-resume-prompt.md")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-prompt",
		"--source-readback", sourceReadback,
		"--workgraph", sourceWorkgraph,
		"--lease-start", filepath.Join(waveRoot, "lease-start.json"),
		"--checkpoint-readback", sourceCheckpoint,
		"--evidence-root", filepath.ToSlash(filepath.Join("docs", "evidence", "ao-atlas-feature-depth-wave-v02")),
		"--node-id", "mission-recommendation-feature-depth-next-wave-19",
		"--expected-next-node-after-completion", "mission-recommendation-feature-depth-next-wave-20",
		"--prompt-out", outPrompt,
		"--fixture-out", outFixture,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-prompt command failed: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outFixture)
	generated["prompt_path"] = recorded["prompt_path"]
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 checkpoint digest resume prompt fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if recorded["checkpoint_readback_path"] != "docs/evidence/ao-atlas-feature-depth-wave-v02/nodes/mission-recommendation-feature-depth-next-wave-18/checkpoint-readback-after.json" {
		t.Fatalf("v02 checkpoint readback path not bound: %#v", recorded["checkpoint_readback_path"])
	}
	checkpointDigest, ok := recorded["checkpoint_readback_digest"].(string)
	if !ok || !digestPattern.MatchString(checkpointDigest) {
		t.Fatalf("v02 checkpoint digest missing or invalid: %#v", recorded["checkpoint_readback_digest"])
	}
	promptBytes, err := os.ReadFile(recordedPromptPath)
	if err != nil {
		t.Fatalf("read recorded prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"Checkpoint readback: `docs/evidence/ao-atlas-feature-depth-wave-v02/nodes/mission-recommendation-feature-depth-next-wave-18/checkpoint-readback-after.json`",
		"Checkpoint readback digest: `" + checkpointDigest + "`",
		"Next executable node: `mission-recommendation-feature-depth-next-wave-19`",
		"Do not produce a final response while ready nodes or exact next action remain.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("v02 checkpoint digest resume prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestCompactionResumePromptCarriesSchemaHealthStatusWhenReadbackHasIt(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-16", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-16", "workgraph-after.json")
	readback := mustLoadJSON[AtlasRecommendationReadback](t, sourceReadback)
	readback.SchemaHealthStatus = "failed_missing_registry_artifacts"

	outDir := t.TempDir()
	syntheticReadback := filepath.Join(outDir, "recommendation-readback-after.json")
	outFixture := filepath.Join(outDir, "compaction-resume-prompt.json")
	outPrompt := filepath.Join(outDir, "compaction-resume-prompt.md")
	if err := WriteJSON(syntheticReadback, readback); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "compaction-resume-prompt",
		"--source-readback", syntheticReadback,
		"--workgraph", sourceWorkgraph,
		"--lease-start", filepath.Join(waveRoot, "lease-start.json"),
		"--evidence-root", filepath.ToSlash(filepath.Join("docs", "evidence", "ao-atlas-feature-depth-wave-v01")),
		"--node-id", "mission-recommendation-feature-depth-next-wave-17",
		"--expected-next-node-after-completion", "mission-recommendation-feature-depth-next-wave-18",
		"--prompt-out", outPrompt,
		"--fixture-out", outFixture,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("compaction-resume-prompt command failed: %s", out.String())
	}
	fixture := mustLoadJSON[AtlasCompactionResumePrompt](t, outFixture)
	if fixture.SchemaHealthStatus != "failed_missing_registry_artifacts" {
		t.Fatalf("compaction resume prompt lost schema health status: %#v", fixture.SchemaHealthStatus)
	}
	promptBytes, err := os.ReadFile(outPrompt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(promptBytes), "Schema health status: `failed_missing_registry_artifacts`") {
		t.Fatalf("compaction resume prompt markdown missing schema health status:\n%s", string(promptBytes))
	}
	assertSchemaHasProperty(t, filepath.Join(root, "schemas", "compaction-resume-prompt.schema.json"), "schema_health_status")
}
