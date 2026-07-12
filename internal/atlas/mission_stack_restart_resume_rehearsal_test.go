package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth6StackRestartResumeRehearsal(t *testing.T) {
	root := repoRoot(t)
	recordedPath := filepath.Join(root, "examples", "valid", "stack-restart-resume-rehearsal.json")
	outPath := filepath.Join(t.TempDir(), "stack-restart-resume-rehearsal.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "stack-restart-resume-rehearsal",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("stack-restart-resume-rehearsal command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=restart_resume_rehearsal_ready",
		"component_count=3",
		"mission_checkpoint_bound=true",
		"atlas_workgraph_bound=true",
		"foundry_safe_next_work_bound=true",
		"final_response_allowed=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("stack restart resume output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("stack restart resume rehearsal changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["no_lost_evidence"] != true ||
		generated["single_active_node_preserved"] != true ||
		generated["final_response_allowed"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("stack restart resume rehearsal lost safety state: %#v", generated)
	}
}

func TestMonth6StackRestartResumeRehearsalUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "valid", "stack-restart-resume-rehearsal.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.stack-restart-resume-rehearsal.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:stack-restart-resume-rehearsal" {
		t.Fatalf("expected typed stack restart resume rehearsal validator, got %s", validator)
	}
}

func TestMonth6StackRestartResumeRehearsalRejectsFinalResponse(t *testing.T) {
	fixture, err := BuildAtlasStackRestartResumeRehearsal()
	if err != nil {
		t.Fatal(err)
	}
	fixture.FinalResponseAllowed = true
	if err := ValidateAtlasStackRestartResumeRehearsal(fixture); err == nil || !strings.Contains(err.Error(), "final_response_allowed must be false") {
		t.Fatalf("expected final response rejection, got %v", err)
	}
}
