package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3ArchitectureSourceTruthChecklistBindsCurrentAuthorityCorrections(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-15-architecture-source-truth")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-14-compaction-resume-generator", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "architecture-source-truth-checklist.json")
	outPath := filepath.Join(t.TempDir(), "architecture-source-truth-checklist.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-architecture-source-truth",
		"--node-id", "mission-recommendation-month3-final-closure-15-architecture-source-truth",
		"--source-readback", sourceReadback,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-architecture-source-truth command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=checklist_ready") ||
		!strings.Contains(out.String(), "completed_nodes=14") ||
		!strings.Contains(out.String(), "corrections=4") {
		t.Fatalf("architecture source-truth output missing checklist state: %s", out.String())
	}

	recorded := mustLoadJSON[AtlasMonth3ArchitectureSourceTruthChecklist](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3ArchitectureSourceTruthChecklist](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("architecture source-truth checklist drifted\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3ArchitectureSourceTruthChecklist(recorded); err != nil {
		t.Fatalf("recorded architecture source-truth checklist invalid: %v", err)
	}
	if recorded.CurrentAuthorityStatement != "highest_proven_live_class=complex_repo_mutation; fully_unsupervised_complex_mutation=denied; RSI=denied" ||
		recorded.CompletedNodes != 14 ||
		recorded.ReadyNodes != 16 ||
		!recorded.CorrectionsRequired ||
		len(recorded.Checklist) != 4 ||
		recorded.PromotionRequested ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("architecture source-truth checklist lost authority correction boundaries: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, AtlasMonth3ArchitectureSourceTruthChecklistContract)
	if err != nil {
		t.Fatalf("typed validator rejected architecture source-truth checklist: %v", err)
	}
	if validator != "typed:month3-architecture-source-truth" {
		t.Fatalf("expected architecture source-truth typed validator, got %s", validator)
	}
}
