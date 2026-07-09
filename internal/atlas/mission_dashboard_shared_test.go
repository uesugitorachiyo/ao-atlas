package atlas

import (
	"path/filepath"
	"testing"
)

func TestMissionDashboardEvidenceArtifactFeedsProvenanceAndFreshnessRows(t *testing.T) {
	root := repoRoot(t)
	evidencePath := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01", "nodes", "mission-recommendation-feature-depth-next-wave-34", "mission-dashboard-provenance-links.json")

	artifact, err := buildMissionDashboardEvidenceArtifact(evidencePath)
	if err != nil {
		t.Fatal(err)
	}

	if artifact.PublicPath != "docs/evidence/ao-atlas-feature-depth-wave-v01/nodes/mission-recommendation-feature-depth-next-wave-34/mission-dashboard-provenance-links.json" {
		t.Fatalf("unexpected public path: %#v", artifact)
	}
	if !digestPattern.MatchString(artifact.Digest) {
		t.Fatalf("artifact digest must be normalized sha256: %#v", artifact)
	}

	check := buildMissionDashboardFreshnessCheckFromArtifact("dashboard_sources_match_completed_readback", true, artifact)
	if check.EvidencePath != artifact.PublicPath || check.EvidenceDigest != artifact.Digest || check.Status != "passed" {
		t.Fatalf("freshness check did not reuse artifact provenance: check=%#v artifact=%#v", check, artifact)
	}

	link := buildMissionDashboardProvenanceLinkFromArtifact("ao-sentinel", "public_safety_scan_owner", artifact, artifact.Digest, false, true)
	if link.EvidencePath != artifact.PublicPath || link.EvidenceDigest != artifact.Digest || !link.ArtifactDigestVerified || link.AuthorityAdvanceClaimed || !link.RSIRemainsDenied {
		t.Fatalf("provenance link did not reuse artifact provenance: link=%#v artifact=%#v", link, artifact)
	}
}
