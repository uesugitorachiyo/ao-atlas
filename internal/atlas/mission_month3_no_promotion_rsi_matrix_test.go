package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth3NoPromotionRSIMatrixCoversCompletedArtifacts(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-m3-final-closure-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-16-no-promotion-rsi-matrix")
	sourceReadback := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-15-architecture-source-truth", "recommendation-readback-after.json")
	sourceWorkgraph := filepath.Join(waveRoot, "nodes", "mission-recommendation-month3-final-closure-15-architecture-source-truth", "workgraph-after.json")
	recordedPath := filepath.Join(nodeDir, "no-promotion-rsi-matrix.json")
	outPath := filepath.Join(t.TempDir(), "no-promotion-rsi-matrix.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "month3-no-promotion-rsi-matrix",
		"--node-id", "mission-recommendation-month3-final-closure-16-no-promotion-rsi-matrix",
		"--source-readback", sourceReadback,
		"--source-workgraph", sourceWorkgraph,
		"--evidence-root", waveRoot,
		"--expected-next-node-after-completion", "mission-recommendation-month3-final-closure-17-workspace-root-preflight",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("month3-no-promotion-rsi-matrix command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=asserted") ||
		!strings.Contains(out.String(), "completed_nodes=15") ||
		!strings.Contains(out.String(), "no_promotion_invariant_holds=true") ||
		!strings.Contains(out.String(), "rsi_denial_invariant_holds=true") {
		t.Fatalf("no-promotion RSI matrix output missing assertion state: %s", out.String())
	}

	recorded := mustLoadJSON[AtlasMonth3NoPromotionRSIMatrix](t, recordedPath)
	generated := mustLoadJSON[AtlasMonth3NoPromotionRSIMatrix](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Month 3 no-promotion RSI matrix drifted\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMonth3NoPromotionRSIMatrix(recorded); err != nil {
		t.Fatalf("recorded no-promotion RSI matrix invalid: %v", err)
	}
	if recorded.CompletedNodes != 15 ||
		recorded.PromoterNoPromotionFiles != 15 ||
		recorded.CommandReadbackFiles != 15 ||
		recorded.SentinelPublicSafetyFiles != 15 ||
		recorded.PromotionRequestedFalseCount != 15 ||
		recorded.PromotionGrantedFalseCount != 15 ||
		recorded.SentinelRSIDeniedCount != 15 ||
		recorded.ExpectedNextNodeAfterCompletion != "mission-recommendation-month3-final-closure-17-workspace-root-preflight" ||
		!recorded.NoPromotionInvariantHolds ||
		!recorded.RSIDenialInvariantHolds ||
		recorded.PromotionRequested ||
		recorded.PromotionGranted ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("no-promotion RSI matrix lost completed artifact coverage or safety boundary: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, AtlasMonth3NoPromotionRSIMatrixContract)
	if err != nil {
		t.Fatalf("typed validator rejected no-promotion RSI matrix: %v", err)
	}
	if validator != "typed:month3-no-promotion-rsi-matrix" {
		t.Fatalf("expected no-promotion RSI matrix typed validator, got %s", validator)
	}
}
