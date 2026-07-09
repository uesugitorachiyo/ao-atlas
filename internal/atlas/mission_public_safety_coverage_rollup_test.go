package atlas

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestFeatureDepthWavePublicSafetyCoverageRollupSummarizesCompletedCoverage(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-24")
	sourceReadbackPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-23", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "public-safety-coverage-rollup.json")
	outPath := filepath.Join(t.TempDir(), "public-safety-coverage-rollup.json")
	sourceReadback := mustLoadJSON[AtlasRecommendationReadback](t, sourceReadbackPath)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "public-safety-coverage-rollup",
		"--node-id", "mission-recommendation-feature-depth-next-wave-24",
		"--source-readback", sourceReadbackPath,
		"--evidence-root", waveRoot,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("public-safety-coverage-rollup command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasPublicSafetyCoverageRollup](t, recordedPath)
	generated := mustLoadJSON[AtlasPublicSafetyCoverageRollup](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("public safety coverage rollup fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasPublicSafetyCoverageRollup(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "passed" ||
		recorded.CompletedNodesBefore != sourceReadback.CompletedNodes ||
		recorded.ReadyNodesBefore != sourceReadback.ReadyNodes ||
		recorded.FirstExecutableNodeBefore != sourceReadback.FirstExecutableNode ||
		recorded.FinalResponseAllowedBefore ||
		recorded.PublicSafetyScanStatus != "passed" ||
		recorded.SentinelEvidenceCount != sourceReadback.CompletedNodes ||
		recorded.CompletedNodesWithSentinel != sourceReadback.CompletedNodes ||
		!recorded.AllCompletedNodesCovered ||
		!recorded.AllSentinelStatusesPassed ||
		!recorded.AllScopedScansPassed ||
		!recorded.PublicSafetyScanPassed ||
		!recorded.MachineReadableClosureRollup ||
		recorded.UnsafeMatchCountTotal != 0 ||
		recorded.ScopedScanCount < 2 ||
		recorded.ChangedPromptArtifactsTotal < 2 ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("public-safety coverage rollup lost coverage or safety state: %#v", recorded)
	}
	if len(recorded.MissingSentinelNodes) != 0 {
		t.Fatalf("all completed nodes should have Sentinel evidence: %#v", recorded.MissingSentinelNodes)
	}
	if len(recorded.SentinelEvidenceFiles) != recorded.SentinelEvidenceCount ||
		len(recorded.ScopedScanFiles) != recorded.ScopedScanCount {
		t.Fatalf("rollup path counts must match recorded lists: %#v", recorded)
	}
	for _, path := range append(recorded.SentinelEvidenceFiles, recorded.ScopedScanFiles...) {
		if path == "" || filepath.IsAbs(path) {
			t.Fatalf("rollup paths must be portable relative paths: %q", path)
		}
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.public-safety-coverage-rollup.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:public-safety-coverage-rollup" {
		t.Fatalf("expected typed public safety coverage rollup validator, got %s", validator)
	}
}

func TestFeatureDepthWaveV02PublicSafetyCoverageRollupSummarizesCompletedCoverage(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-24")
	sourceReadbackPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-23", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "public-safety-coverage-rollup.json")
	outPath := filepath.Join(t.TempDir(), "public-safety-coverage-rollup.json")
	sourceReadback := mustLoadJSON[AtlasRecommendationReadback](t, sourceReadbackPath)

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "public-safety-coverage-rollup",
		"--node-id", "mission-recommendation-feature-depth-next-wave-24",
		"--source-readback", sourceReadbackPath,
		"--evidence-root", waveRoot,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("public-safety-coverage-rollup command failed: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasPublicSafetyCoverageRollup](t, recordedPath)
	generated := mustLoadJSON[AtlasPublicSafetyCoverageRollup](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 public safety coverage rollup fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasPublicSafetyCoverageRollup(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "passed" ||
		recorded.CompletedNodesBefore != sourceReadback.CompletedNodes ||
		recorded.ReadyNodesBefore != sourceReadback.ReadyNodes ||
		recorded.FirstExecutableNodeBefore != sourceReadback.FirstExecutableNode ||
		recorded.FinalResponseAllowedBefore ||
		recorded.PublicSafetyScanStatus != "passed" ||
		recorded.SentinelEvidenceCount != sourceReadback.CompletedNodes ||
		recorded.CompletedNodesWithSentinel != sourceReadback.CompletedNodes ||
		!recorded.AllCompletedNodesCovered ||
		!recorded.AllSentinelStatusesPassed ||
		!recorded.AllScopedScansPassed ||
		!recorded.PublicSafetyScanPassed ||
		!recorded.MachineReadableClosureRollup ||
		recorded.UnsafeMatchCountTotal != 0 ||
		recorded.ScopedScanCount < 1 ||
		recorded.ChangedPromptArtifactsTotal < 1 ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("v02 public-safety coverage rollup lost coverage or safety state: %#v", recorded)
	}
}
