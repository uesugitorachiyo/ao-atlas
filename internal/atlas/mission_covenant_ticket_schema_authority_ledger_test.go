package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BCovenantTicketSchemaAuthorityLedgerRecordsCompatibleEntries(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-16")
	inputPath := filepath.Join(nodeDir, "covenant-ticket-schema-authority-ledger-input.json")
	recordedPath := filepath.Join(nodeDir, "covenant-ticket-schema-authority-ledger.json")
	outPath := filepath.Join(t.TempDir(), "covenant-ticket-schema-authority-ledger.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "covenant-ticket-schema-authority-ledger",
		"--input", inputPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("covenant-ticket-schema-authority-ledger command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=covenant_ticket_schema_authority_compatible") ||
		!strings.Contains(out.String(), "all_entries_compatible=true") ||
		!strings.Contains(out.String(), "entry_count=3") {
		t.Fatalf("covenant ticket schema authority output missing summary: %s", out.String())
	}
	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("covenant ticket schema authority ledger changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["all_entries_compatible"] != true ||
		generated["entry_count"] != float64(3) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("covenant ticket schema authority ledger lost safety state: %#v", generated)
	}
}

func TestP0BCovenantTicketSchemaAuthorityLedgerUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-16", "covenant-ticket-schema-authority-ledger.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.covenant-ticket-schema-authority-ledger.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:covenant-ticket-schema-authority-ledger" {
		t.Fatalf("expected typed covenant ticket schema authority ledger validator, got %s", validator)
	}
}
