package atlas

import (
	"path/filepath"
	"testing"
)

func TestP0BCommandPromoterAgreementBindsCompletedNodes(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-24", "command-promoter-agreement.json")
	agreement := mustLoadJSON[AtlasP0BCommandPromoterAgreement](t, path)
	if err := ValidateAtlasP0BCommandPromoterAgreement(agreement); err != nil {
		t.Fatal(err)
	}
	if agreement.CoveredNodeStart != 1 ||
		agreement.CoveredNodeEnd != 23 ||
		agreement.EntryCount != 23 ||
		!agreement.AllCommandReadbacksAgree ||
		!agreement.AllPromoterReadbacksNoPromotion ||
		agreement.PromotionRequestedCount != 0 ||
		agreement.PromotionGrantedCount != 0 ||
		agreement.AuthorityAdvanceClaimCount != 0 ||
		agreement.FinalResponseAllowed {
		t.Fatalf("P0-B Command/Promoter agreement drifted from no-promotion coverage: %#v", agreement)
	}
}

func TestP0BCommandPromoterAgreementUsesTypedEvidenceValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-24", "command-promoter-agreement.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, AtlasP0BCommandPromoterAgreementContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:p0b-command-promoter-agreement" {
		t.Fatalf("unexpected validator: %s", validator)
	}
}
