package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveMissionDashboardProvenanceLinksBindSubsystemEvidence(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-34")
	sourceBindingPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-33", "mission-dashboard-closure-binding.json")
	recordedPath := filepath.Join(nodeDir, "mission-dashboard-provenance-links.json")
	outPath := filepath.Join(t.TempDir(), "mission-dashboard-provenance-links.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "mission-dashboard-provenance-links",
		"--node-id", "mission-recommendation-feature-depth-next-wave-34",
		"--dashboard-binding", sourceBindingPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission-dashboard-provenance-links command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=dashboard_provenance_links_bound") ||
		!strings.Contains(out.String(), "provenance_link_count=4") ||
		!strings.Contains(out.String(), "all_required_provenance_linked=true") {
		t.Fatalf("dashboard provenance links output missing expected state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasMissionDashboardProvenanceLinks](t, recordedPath)
	generated := mustLoadJSON[AtlasMissionDashboardProvenanceLinks](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Mission dashboard provenance links fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasMissionDashboardProvenanceLinks(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "dashboard_provenance_links_bound" ||
		recorded.SourceRowCount != 6 ||
		recorded.ProvenanceLinkCount != 4 ||
		!recorded.AllRequiredProvenanceLinked ||
		!recorded.EveryLinkMatchesDashboard ||
		!recorded.EveryLinkedArtifactDigestVerified ||
		recorded.FinalResponseAllowed ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("dashboard provenance links lost closure guards: %#v", recorded)
	}
	links := map[string]AtlasMissionDashboardProvenanceLink{}
	for _, link := range recorded.ProvenanceLinks {
		links[link.Repo] = link
	}
	for _, repo := range []string{"ao-command", "ao-foundry", "ao-promoter", "ao-sentinel"} {
		link, ok := links[repo]
		if !ok {
			t.Fatalf("provenance link missing %s", repo)
		}
		if link.EvidencePath == "" || !link.DashboardRowMatched || !link.ClosureEvidenceDigestMatches || !link.ArtifactDigestVerified || link.AuthorityAdvanceClaimed || !link.RSIRemainsDenied {
			t.Fatalf("provenance link %s lost evidence binding or safety state: %#v", repo, link)
		}
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.mission-dashboard-provenance-links.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:mission-dashboard-provenance-links" {
		t.Fatalf("expected typed Mission dashboard provenance links validator, got %s", validator)
	}
}
