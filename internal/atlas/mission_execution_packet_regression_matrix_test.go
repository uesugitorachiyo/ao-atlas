package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3ExecutionPacketRegressionMatrix(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-22")
	recordedPath := filepath.Join(nodeDir, "execution-packet-regression-matrix.json")
	outPath := filepath.Join(t.TempDir(), "execution-packet-regression-matrix.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "execution-packet-regression-matrix",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("execution-packet-regression-matrix command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=execution_packet_regression_matrix_ready",
		"cases=2",
		"provider_invocation_allowed=false",
		"silent_changed_result_allowed=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("execution packet regression output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("execution packet regression matrix changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["provider_invocation_allowed"] != false ||
		generated["silent_changed_result_allowed"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("execution packet regression matrix lost authority state: %#v", generated)
	}
}

func TestMonth3ExecutionPacketRegressionMatrixUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-22", "execution-packet-regression-matrix.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.execution-packet-regression-matrix.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:execution-packet-regression-matrix" {
		t.Fatalf("expected typed execution packet regression matrix validator, got %s", validator)
	}
}

func TestMonth3ExecutionPacketRegressionMatrixRejectsSilentChangedResult(t *testing.T) {
	matrix, err := BuildAtlasExecutionPacketRegressionMatrix()
	if err != nil {
		t.Fatal(err)
	}
	matrix.SilentChangedResultAllowed = true
	if err := ValidateAtlasExecutionPacketRegressionMatrix(matrix); err == nil || !strings.Contains(err.Error(), "silent_changed_result_allowed must be false") {
		t.Fatalf("expected silent changed result rejection, got %v", err)
	}
}
