package atlas

import (
	"path/filepath"
	"testing"
)

func TestMonth6GoldenPathDryRunRehearsalsFixtureBindsThreeExternalRepos(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month6-beta-launch-v01", "nodes", "month6-recommendation-03-golden-path-dry-run-rehearsals", "golden-path-dry-run-rehearsals.json")

	fixture := mustLoadJSON[AtlasMonth6GoldenPathDryRunRehearsals](t, fixturePath)
	if err := ValidateAtlasMonth6GoldenPathDryRunRehearsals(fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Status != "dry_run_rehearsals_bound" ||
		fixture.SourceRecommendationRank != 3 ||
		fixture.RehearsalCount != 3 ||
		len(fixture.Rehearsals) != 3 ||
		!fixture.FixtureOnly ||
		!fixture.DryRunOnly ||
		fixture.ProviderCallsAllowed ||
		fixture.CredentialUseAllowed ||
		fixture.LiveMutationAllowed ||
		fixture.ReleaseOrPublishAllowed ||
		fixture.ApprovalGranted ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.MutatesRepositories {
		t.Fatalf("Month 6 golden-path rehearsal fixture lost safety state: %#v", fixture)
	}
	seen := map[string]bool{}
	for _, rehearsal := range fixture.Rehearsals {
		if seen[rehearsal.ID] {
			t.Fatalf("duplicate rehearsal id %q", rehearsal.ID)
		}
		seen[rehearsal.ID] = true
		if rehearsal.RepoClass != "external_non_ao" ||
			!rehearsal.FixtureOnly ||
			!rehearsal.DryRunOnly ||
			rehearsal.ProviderCallsAllowed ||
			rehearsal.CredentialUseAllowed ||
			rehearsal.LiveMutationAllowed ||
			rehearsal.ReleaseOrPublishAllowed ||
			rehearsal.ApprovalGranted ||
			rehearsal.SafeToExecute ||
			rehearsal.ExecutesWork ||
			rehearsal.MutatesRepository {
			t.Fatalf("unsafe external non-AO rehearsal: %#v", rehearsal)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(fixturePath, AtlasMonth6GoldenPathDryRunRehearsalsContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month6-golden-path-dry-run-rehearsals" {
		t.Fatalf("expected typed Month 6 golden path dry-run rehearsal validator, got %s", validator)
	}
}

func TestMonth6GoldenPathDryRunRehearsalsRejectsProviderExecution(t *testing.T) {
	fixture := validMonth6GoldenPathDryRunRehearsalsFixtureForTest()
	fixture.Rehearsals[0].ProviderCallsAllowed = true
	fixture.ProviderCallsAllowed = true
	if err := ValidateAtlasMonth6GoldenPathDryRunRehearsals(fixture); err == nil {
		t.Fatal("expected provider execution to be rejected")
	}
}

func validMonth6GoldenPathDryRunRehearsalsFixtureForTest() AtlasMonth6GoldenPathDryRunRehearsals {
	return AtlasMonth6GoldenPathDryRunRehearsals{
		Schema:                   AtlasMonth6GoldenPathDryRunRehearsalsContract,
		NodeID:                   "month6-recommendation-03-golden-path-dry-run-rehearsals",
		Status:                   "dry_run_rehearsals_bound",
		SourceRecommendationRank: 3,
		SourceRecommendationTask: "Run three dry-run golden path rehearsals on non-AO sample repos",
		SafetyGate:               "planning_only_no_provider_no_release",
		RehearsalCount:           3,
		Rehearsals: []AtlasMonth6GoldenPathDryRunRehearsal{
			validMonth6GoldenPathDryRunRehearsalForTest("external-non-ao-cli-sample"),
			validMonth6GoldenPathDryRunRehearsalForTest("external-non-ao-library-sample"),
			validMonth6GoldenPathDryRunRehearsalForTest("external-non-ao-docs-sample"),
		},
		FixtureOnly:             true,
		DryRunOnly:              true,
		NoPromotionRequested:    true,
		ClaimsAuthorityAdvance:  false,
		RSIRemainsDenied:        true,
		SafeToExecute:           false,
		ExecutesWork:            false,
		ApprovesWork:            false,
		MutatesRepositories:     false,
		ProviderCallsAllowed:    false,
		CredentialUseAllowed:    false,
		LiveMutationAllowed:     false,
		ReleaseOrPublishAllowed: false,
		ApprovalGranted:         false,
	}
}

func validMonth6GoldenPathDryRunRehearsalForTest(id string) AtlasMonth6GoldenPathDryRunRehearsal {
	return AtlasMonth6GoldenPathDryRunRehearsal{
		ID:                      id,
		Repository:              id,
		RepoClass:               "external_non_ao",
		ObjectiveDigest:         "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		BaseCommit:              "0000000000000000000000000000000000000000",
		DiffDigestPlaceholder:   "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		RollbackReceiptRef:      "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-03-golden-path-dry-run-rehearsals/rollback-receipts/" + id + ".json",
		CommandReadbackRef:      "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-03-golden-path-dry-run-rehearsals/command-readbacks/" + id + ".json",
		SentinelPublicSafetyRef: "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-03-golden-path-dry-run-rehearsals/sentinel-public-safety/" + id + ".json",
		PromoterNoPromotionRef:  "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-03-golden-path-dry-run-rehearsals/promoter-no-promotion/" + id + ".json",
		FixtureOnly:             true,
		DryRunOnly:              true,
		ProviderCallsAllowed:    false,
		CredentialUseAllowed:    false,
		LiveMutationAllowed:     false,
		ReleaseOrPublishAllowed: false,
		ApprovalGranted:         false,
		SafeToExecute:           false,
		ExecutesWork:            false,
		MutatesRepository:       false,
	}
}
