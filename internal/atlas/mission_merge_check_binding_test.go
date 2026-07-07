package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveMergeCheckBindingFixtureBindsMergeCommitsToPassedChecks(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-12")
	inputPath := filepath.Join(nodeDir, "merge-check-binding-input.json")
	recorded := mustLoadJSON[map[string]any](t, filepath.Join(nodeDir, "merge-check-binding.json"))
	outPath := filepath.Join(t.TempDir(), "merge-check-binding.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "merge-check-binding",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("merge-check-binding command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=required_checks_bound") ||
		!strings.Contains(out.String(), "row_count=3") ||
		!strings.Contains(out.String(), "passed_required_check_rows=3") {
		t.Fatalf("merge-check-binding output missing binding summary: %s", out.String())
	}
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("merge check binding fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["status"] != "required_checks_bound" ||
		generated["row_count"] != float64(3) ||
		generated["passed_required_check_rows"] != float64(3) ||
		generated["unbound_merge_commits"] != float64(0) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("merge check binding must bind merge commits to passed checks without authority effects: %#v", generated)
	}
}

func TestFeatureDepthWaveMergeCheckBindingFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-12", "merge-check-binding.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.merge-check-binding.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:merge-check-binding" {
		t.Fatalf("expected typed merge check binding validator, got %s", validator)
	}
}
