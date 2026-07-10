package atlas

import (
	"path/filepath"
	"testing"
)

func TestP0CReadinessBridgeBindsP0BClosureCriteria(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-29", "p0c-readiness-criteria.json")
	criteria := mustLoadJSON[AtlasP0CReadinessCriteria](t, path)
	if err := ValidateAtlasP0CReadinessCriteria(criteria); err != nil {
		t.Fatal(err)
	}
	if criteria.CompletedNodesBefore != 28 ||
		criteria.ReadyNodesBefore != 2 ||
		criteria.RequiredCriterionCount != len(criteria.RequiredCriteria) ||
		criteria.RequiredCriterionCount < 10 ||
		criteria.NextExecutableNode != "mission-recommendation-p0b-contract-convergence-30" ||
		!criteria.RequiresP0BTerminalReadback ||
		!criteria.RequiresMissionToFoundryCompletePath ||
		criteria.FinalResponseAllowed ||
		criteria.PromotionRequested ||
		criteria.ClaimsAuthorityAdvance ||
		!criteria.RSIRemainsDenied {
		t.Fatalf("P0-C readiness criteria drifted from P0-B closure handoff boundary: %#v", criteria)
	}
}

func TestP0CReadinessBridgeUsesTypedEvidenceValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-29", "p0c-readiness-criteria.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, AtlasP0CReadinessCriteriaContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:p0c-readiness-criteria" {
		t.Fatalf("unexpected validator: %s", validator)
	}
}
