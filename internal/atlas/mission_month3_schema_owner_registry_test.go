package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3SchemaOwnerRegistryProposalBindsCovenantOwnerAndConsumers(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-07-schema-owner-registry")
	recordedPath := filepath.Join(nodeDir, "month3-schema-owner-registry-proposal.json")
	outPath := filepath.Join(t.TempDir(), "month3-schema-owner-registry-proposal.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-schema-owner-registry",
		"--node-id", "mission-recommendation-month3-final-closure-07-schema-owner-registry",
		"--registry-manifest", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-8", "canonical-contract-registry-manifest.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-schema-owner-registry command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3SchemaOwnerRegistryProposal](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3SchemaOwnerRegistryProposal](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 schema owner registry proposal changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3SchemaOwnerRegistryProposal(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "schema_owner_registry_proposal_ready" ||
		recorded.RegistryAuthorityOwner != "ao-covenant" ||
		recorded.ContractCount != 4 ||
		recorded.ConsumerCompatibilityCheckCount != 9 ||
		!recorded.CovenantOwnsRegistry ||
		!recorded.ProducersRetainContractImplementation ||
		!recorded.ConsumerCompatibilityRequired ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("schema owner registry proposal lost safety state: %#v", recorded)
	}
	for _, contract := range recorded.Contracts {
		if contract.RegistryOwner != "ao-covenant" ||
			contract.ProducerOwner == "" ||
			len(contract.ConsumerCompatibilityChecks) == 0 {
			t.Fatalf("contract owner mapping incomplete: %#v", contract)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-schema-owner-registry-proposal.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-schema-owner-registry-proposal" {
		t.Fatalf("expected typed Month 3 schema owner registry proposal validator, got %s", validator)
	}
}
