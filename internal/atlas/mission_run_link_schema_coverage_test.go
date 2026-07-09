package atlas

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveRunLinkSchemaCoverageSummarizesEveryGeneratedRunLink(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-06")
	fixturePath := filepath.Join(nodeDir, "run-link-schema-coverage.json")
	recorded := mustLoadJSON[AtlasRunLinkSchemaCoverage](t, fixturePath)
	checkpointRoot := runLinkSchemaCoverageCheckpointRoot(t, waveRoot, recorded)

	coverage, err := BuildAtlasRunLinkSchemaCoverage(checkpointRoot)
	if err != nil {
		t.Fatal(err)
	}
	coverage.EvidenceRoot = recorded.EvidenceRoot
	if err := ValidateAtlasRunLinkSchemaCoverage(coverage); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasRunLinkSchemaCoverage(recorded); err != nil {
		t.Fatal(err)
	}
	if digestValue(coverage) != digestValue(recorded) {
		t.Fatalf("run-link schema coverage fixture drifted\nwant %s\ngot  %s", digestValue(coverage), digestValue(recorded))
	}
	if coverage.RunLinkCount != 6 ||
		coverage.CompletedRunLinks != 6 ||
		coverage.SchemaCounts[RunLinkContract] != 6 ||
		coverage.ValidatorCounts["typed:run-link"] != 6 ||
		coverage.SchedulesWork ||
		coverage.ExecutesWork ||
		coverage.ApprovesWork ||
		coverage.ClaimsAuthorityAdvance ||
		!coverage.RSIRemainsDenied {
		t.Fatalf("run-link coverage must summarize six completed run links without authority effects: %#v", coverage)
	}
	for _, entry := range coverage.Entries {
		if entry.Schema != RunLinkContract ||
			entry.Validator != "typed:run-link" ||
			entry.Status != "completed" ||
			entry.EvidenceKeyCount == 0 {
			t.Fatalf("run-link coverage entry must bind typed run-link evidence: %#v", entry)
		}
	}
}

func TestFeatureDepthWaveV02RunLinkSchemaCoverageSummarizesEveryGeneratedRunLink(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-06")
	fixturePath := filepath.Join(nodeDir, "run-link-schema-coverage.json")
	recorded := mustLoadJSON[AtlasRunLinkSchemaCoverage](t, fixturePath)
	checkpointRoot := runLinkSchemaCoverageCheckpointRoot(t, waveRoot, recorded)

	coverage, err := BuildAtlasRunLinkSchemaCoverage(checkpointRoot)
	if err != nil {
		t.Fatal(err)
	}
	coverage.EvidenceRoot = recorded.EvidenceRoot
	if err := ValidateAtlasRunLinkSchemaCoverage(coverage); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAtlasRunLinkSchemaCoverage(recorded); err != nil {
		t.Fatal(err)
	}
	if digestValue(coverage) != digestValue(recorded) {
		t.Fatalf("v02 run-link schema coverage fixture drifted\nwant %s\ngot  %s", digestValue(coverage), digestValue(recorded))
	}
	if coverage.RunLinkCount != 6 ||
		coverage.CompletedRunLinks != 6 ||
		coverage.SchemaCounts[RunLinkContract] != 6 ||
		coverage.ValidatorCounts["typed:run-link"] != 6 ||
		coverage.SchedulesWork ||
		coverage.ExecutesWork ||
		coverage.ApprovesWork ||
		coverage.ClaimsAuthorityAdvance ||
		!coverage.RSIRemainsDenied {
		t.Fatalf("v02 run-link coverage must summarize six completed run links without authority effects: %#v", coverage)
	}
	for _, entry := range coverage.Entries {
		if entry.Schema != RunLinkContract ||
			entry.Validator != "typed:run-link" ||
			entry.Status != "completed" ||
			entry.EvidenceKeyCount == 0 {
			t.Fatalf("v02 run-link coverage entry must bind typed run-link evidence: %#v", entry)
		}
	}
}

func TestMissionRecommendationsRunLinkSchemaCoverageCLIWritesDeterministicArtifact(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-06")
	recorded := mustLoadJSON[AtlasRunLinkSchemaCoverage](t, filepath.Join(nodeDir, "run-link-schema-coverage.json"))
	checkpointRoot := runLinkSchemaCoverageCheckpointRoot(t, waveRoot, recorded)
	outPath := filepath.Join(t.TempDir(), "run-link-schema-coverage.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "run-link-schema-coverage",
		"--evidence-root", checkpointRoot,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-link-schema-coverage command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=complete") ||
		!strings.Contains(out.String(), "run_link_count=6") ||
		!strings.Contains(out.String(), "typed_run_link_validators=6") {
		t.Fatalf("run-link-schema-coverage output missing coverage summary: %s", out.String())
	}
	generated := mustLoadJSON[AtlasRunLinkSchemaCoverage](t, outPath)
	if err := ValidateAtlasRunLinkSchemaCoverage(generated); err != nil {
		t.Fatal(err)
	}
	generated.EvidenceRoot = recorded.EvidenceRoot
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("CLI run-link coverage output drifted\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
}

func runLinkSchemaCoverageCheckpointRoot(t *testing.T, waveRoot string, recorded AtlasRunLinkSchemaCoverage) string {
	t.Helper()
	checkpointRoot := filepath.Join(t.TempDir(), "evidence")
	for _, entry := range recorded.Entries {
		src := filepath.Join(waveRoot, filepath.FromSlash(entry.Path))
		dst := filepath.Join(checkpointRoot, filepath.FromSlash(entry.Path))
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return checkpointRoot
}
