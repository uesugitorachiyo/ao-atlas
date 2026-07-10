package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0BContractConvergenceEvidenceValidationCommandCoversWholeWave(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-p0b-contract-convergence-26")
	recordedPath := filepath.Join(nodeDir, "evidence-validation-report.json")
	outPath := filepath.Join(t.TempDir(), "evidence-validation-report.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "validate-evidence",
		"--evidence-root", waveRoot,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("validate-evidence failed: %s", out.String())
	}
	for _, want := range []string{
		"status=passed",
		"node_count=26",
		"failed_files=0",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("validate-evidence output missing %q:\n%s", want, out.String())
		}
	}

	recorded := mustLoadJSON[AtlasRecommendationEvidenceValidationReport](t, recordedPath)
	if err := ValidateAtlasRecommendationEvidenceValidationReport(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "passed" ||
		recorded.NodeCount != 26 ||
		recorded.JSONFileCount == 0 ||
		recorded.ValidatedJSONFiles != recorded.JSONFileCount ||
		len(recorded.MissingSchemaFiles) != 0 ||
		len(recorded.FailedFiles) != 0 ||
		!recorded.RequiredFilenamesCovered {
		t.Fatalf("P0-B evidence validation report did not cover the whole wave cleanly: %#v", recorded)
	}
	if recorded.SchemaCounts[AtlasCompactionResumePromptContract] == 0 ||
		recorded.SchemaCounts[AtlasP0BCommandPromoterAgreementContract] == 0 ||
		recorded.Validators["typed:compaction-resume-prompt"] == 0 ||
		recorded.Validators["typed:p0b-command-promoter-agreement"] == 0 {
		t.Fatalf("P0-B evidence validation report missing recent typed evidence coverage: %#v", recorded.SchemaCounts)
	}
}
