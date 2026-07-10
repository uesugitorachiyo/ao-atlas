package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BFoundryTicketSchemaCompatibilityLedgerRecordsCompatibleEntries(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-14")
	inputPath := filepath.Join(nodeDir, "foundry-ticket-schema-compatibility-ledger-input.json")
	recordedPath := filepath.Join(nodeDir, "foundry-ticket-schema-compatibility-ledger.json")
	outPath := filepath.Join(t.TempDir(), "foundry-ticket-schema-compatibility-ledger.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "foundry-ticket-schema-compatibility-ledger",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry-ticket-schema-compatibility-ledger command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=foundry_ticket_schema_compatible") ||
		!strings.Contains(out.String(), "all_entries_compatible=true") ||
		!strings.Contains(out.String(), "entry_count=2") {
		t.Fatalf("foundry ticket schema compatibility output missing summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("foundry ticket schema compatibility ledger changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["all_entries_compatible"] != true ||
		generated["entry_count"] != float64(2) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("foundry ticket schema compatibility ledger lost safety state: %#v", generated)
	}
}

func TestP0BFoundryTicketSchemaCompatibilityLedgerUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-14", "foundry-ticket-schema-compatibility-ledger.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.foundry-ticket-schema-compatibility-ledger.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:foundry-ticket-schema-compatibility-ledger" {
		t.Fatalf("expected typed foundry ticket schema compatibility ledger validator, got %s", validator)
	}
}
