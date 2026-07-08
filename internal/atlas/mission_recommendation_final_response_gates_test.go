package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionRecommendationsFinalResponseGatesPublishReturnCriteria(t *testing.T) {
	root := repoRoot(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	}()

	gatesPath := filepath.Join(t.TempDir(), "recommendation-final-response-gates.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "final-response-gates",
		"--out", gatesPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("final-response-gates failed: %s", out.String())
	}
	for _, want := range []string{
		"status=ready",
		"gate_count=10",
		"final_response_allowed_requires_all_gates=true",
		"rsi_remains_denied=true",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("final-response-gates output missing %q: %s", want, out.String())
		}
	}

	gates := mustLoadJSON[map[string]any](t, gatesPath)
	if gates["schema"] != "ao.atlas.recommendation-final-response-gates.v0.1" ||
		gates["status"] != "ready" ||
		gates["final_response_allowed_requires_all_gates"] != true ||
		gates["no_promotion_requested"] != true ||
		gates["promotion_granted"] != false ||
		gates["claims_authority_advance"] != false ||
		gates["rsi_remains_denied"] != true ||
		gates["safe_to_execute"] != false ||
		gates["schedules_work"] != false ||
		gates["executes_work"] != false ||
		gates["approves_work"] != false ||
		gates["mutates_repositories"] != false {
		t.Fatalf("final response gate registry did not publish safe criteria: %#v", gates)
	}
	entries, ok := gates["gates"].([]any)
	if !ok || len(entries) != 10 {
		t.Fatalf("expected 10 final-response gates, got %#v", gates["gates"])
	}
	wantGates := []string{
		"completed_nodes_equal_total",
		"ready_nodes_zero",
		"blocked_nodes_zero",
		"failed_nodes_zero",
		"return_gate_final_response_allowed",
		"local_verification_passed",
		"public_safety_scan_passed",
		"promoter_no_promotion",
		"command_readback_agrees",
		"rsi_denied",
	}
	for i, want := range wantGates {
		entry, ok := entries[i].(map[string]any)
		if !ok || entry["gate"] != want || entry["required"] != true {
			t.Fatalf("gate %d is wrong: %#v", i, entries[i])
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(gatesPath, "ao.atlas.recommendation-final-response-gates.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:recommendation-final-response-gates" {
		t.Fatalf("expected typed recommendation final-response gates validator, got %s", validator)
	}
}
