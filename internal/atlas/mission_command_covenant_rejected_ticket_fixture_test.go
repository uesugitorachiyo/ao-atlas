package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BCommandCovenantRejectedTicketFixturePreservesNativeReason(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-04")
	inputPath := filepath.Join(nodeDir, "command-covenant-rejected-ticket-input.json")
	recordedPath := filepath.Join(nodeDir, "command-covenant-rejected-ticket-fixture.json")
	outPath := filepath.Join(t.TempDir(), "command-covenant-rejected-ticket-fixture.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "command-covenant-rejected-ticket-fixture",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("command-covenant-rejected-ticket-fixture command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=rejected_ticket_reason_preserved") ||
		!strings.Contains(out.String(), "command_accepts_ticket=false") ||
		!strings.Contains(out.String(), "reason_preserved=true") {
		t.Fatalf("fixture command output missing rejected ticket summary: %s", out.String())
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("rejected ticket fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["command_accepts_ticket"] != false ||
		generated["reason_preserved"] != true ||
		generated["request_sha256"] != "sha256:45e0d47d6247758d8103700d1eb6ba54ac85e914915da78dd991e57ea142e4bc" ||
		generated["ticket_sha256"] != "sha256:a890471900d26321afb6df3fcb293cac702c229beab44a29870b1b17da0a4b3e" ||
		generated["covenant_native_reason"] != "policy_decision_rejected: requested resource is outside authorized write scope" ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("rejected ticket fixture lost digest, reason, or authority state: %#v", generated)
	}
}

func TestP0BCommandCovenantRejectedTicketFixtureUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-04", "command-covenant-rejected-ticket-fixture.json")

	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.command-covenant-rejected-ticket-fixture.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:command-covenant-rejected-ticket-fixture" {
		t.Fatalf("expected typed command/covenant rejected ticket fixture validator, got %s", validator)
	}
}
