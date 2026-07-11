package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMonth3RealRunAcceptanceCriteriaBindThreeExternalRepos(t *testing.T) {
	root := repoRoot(t)
	sourceRoot := filepath.Join(root, "docs", "evidence", "ao-stack-golden-path-month3-wave-v01")
	closureRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-05-real-run-acceptance")
	recordedPath := filepath.Join(nodeDir, "month3-real-run-acceptance-criteria.json")
	outPath := filepath.Join(t.TempDir(), "month3-real-run-acceptance-criteria.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-real-run-acceptance",
		"--node-id", "mission-recommendation-month3-final-closure-05-real-run-acceptance",
		"--readiness-matrix", filepath.Join(sourceRoot, "nodes", "mission-recommendation-month3-golden-path-40", "golden-path-readiness-matrix.json"),
		"--non-ao-replay", filepath.Join(closureRoot, "nodes", "mission-recommendation-month3-final-closure-04-non-ao-dry-run-replay", "month3-non-ao-dry-run-replay.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-real-run-acceptance command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMonth3RealRunAcceptanceCriteria](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3RealRunAcceptanceCriteria](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 real-run acceptance criteria fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3RealRunAcceptanceCriteria(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "real_run_acceptance_ready" ||
		recorded.ExternalRepoCount != 3 ||
		recorded.CriteriaPerRepo != 7 ||
		!recorded.NonAOReplayBound ||
		!recorded.RequiresExplicitOperatorApproval ||
		!recorded.RequiresReviewedPR ||
		!recorded.RequiresRollbackReceipt ||
		!recorded.RequiresObserverReadback ||
		recorded.ExecutesWork ||
		recorded.SchedulesWork ||
		recorded.ApprovesWork ||
		recorded.PromotionRequested ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("real-run acceptance criteria lost safety state: %#v", recorded)
	}
	if len(recorded.ExternalRepos) != 3 {
		t.Fatalf("expected three external repo criteria, got %#v", recorded.ExternalRepos)
	}
	for _, repo := range recorded.ExternalRepos {
		if repo.RepoClass != "external_non_ao" ||
			repo.ProviderExecutionAllowed ||
			!repo.RequiresIsolatedWorktree ||
			!repo.RequiresExactDigestApproval ||
			!repo.RequiresRollbackReceipt ||
			!repo.RequiresObserverReadback ||
			!repo.RequiresNoPromotionVerdict {
			t.Fatalf("external repo acceptance criteria unsafe: %#v", repo)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.month3-real-run-acceptance-criteria.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month3-real-run-acceptance-criteria" {
		t.Fatalf("expected typed Month 3 real-run acceptance validator, got %s", validator)
	}
}
