package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3ContractCompatibilityInventoryRequiresOwnerAndConsumerTests(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-9")
	recordedPath := filepath.Join(nodeDir, "contract-compatibility-inventory.json")
	outPath := filepath.Join(t.TempDir(), "contract-compatibility-inventory.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "contract-compatibility-inventory",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("contract-compatibility-inventory command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=compatibility_inventory_ready",
		"contract_count=4",
		"gate_critical_count=4",
		"consumer_test_count=9",
		"missing_owner_count=0",
		"missing_consumer_test_count=0",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("compatibility inventory output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("contract compatibility inventory changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["missing_owner_count"] != float64(0) ||
		generated["missing_consumer_test_count"] != float64(0) ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("contract compatibility inventory lost safety state: %#v", generated)
	}
}

func TestMonth3ContractCompatibilityInventoryRejectsGateCriticalEntryWithoutOwnerOrConsumerTest(t *testing.T) {
	inventory := AtlasContractCompatibilityInventory{
		Schema:                 AtlasContractCompatibilityInventoryContract,
		Status:                 "compatibility_inventory_ready",
		Contracts:              []AtlasContractCompatibilityEntry{{ID: "missing", SchemaName: "ao.example.v0", GateCritical: true}},
		ContractCount:          1,
		GateCriticalCount:      1,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	err := ValidateAtlasContractCompatibilityInventory(inventory)
	if err == nil {
		t.Fatal("gate-critical entry without owner or consumer tests was accepted")
	}
	if !strings.Contains(err.Error(), "owner must not be empty") ||
		!strings.Contains(err.Error(), "consumer_tests must not be empty") {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestMonth3ContractCompatibilityInventoryUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01", "nodes", "mission-recommendation-month3-golden-path-9", "contract-compatibility-inventory.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.contract-compatibility-inventory.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:contract-compatibility-inventory" {
		t.Fatalf("expected typed contract compatibility inventory validator, got %s", validator)
	}
}
