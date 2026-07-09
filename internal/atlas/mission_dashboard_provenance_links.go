package atlas

import (
	"fmt"
	"strings"
)

var missionDashboardRequiredProvenanceRepos = []string{"ao-command", "ao-foundry", "ao-promoter", "ao-sentinel"}

func BuildAtlasMissionDashboardProvenanceLinks(nodeID, dashboardBindingPath string) (AtlasMissionDashboardProvenanceLinks, error) {
	nodeID = strings.TrimSpace(nodeID)
	dashboardBindingPath = strings.TrimSpace(dashboardBindingPath)
	for name, value := range map[string]string{
		"node id":                nodeID,
		"dashboard binding path": dashboardBindingPath,
	} {
		if value == "" {
			return AtlasMissionDashboardProvenanceLinks{}, fmt.Errorf("%s is required", name)
		}
	}
	binding, err := LoadJSON[AtlasMissionDashboardClosureBinding](dashboardBindingPath)
	if err != nil {
		return AtlasMissionDashboardProvenanceLinks{}, err
	}
	if err := ValidateAtlasMissionDashboardClosureBinding(binding); err != nil {
		return AtlasMissionDashboardProvenanceLinks{}, err
	}
	sourceDigest, err := digestTextFileWithNormalizedLineEndings(dashboardBindingPath)
	if err != nil {
		return AtlasMissionDashboardProvenanceLinks{}, err
	}

	rowsByRepo := map[string]AtlasMissionDashboardClosureBindingRow{}
	for _, row := range binding.Rows {
		rowsByRepo[row.Repo] = row
	}
	links := make([]AtlasMissionDashboardProvenanceLink, 0, len(missionDashboardRequiredProvenanceRepos))
	for _, repo := range missionDashboardRequiredProvenanceRepos {
		row := rowsByRepo[repo]
		artifact, err := buildMissionDashboardEvidenceArtifact(row.ClosureEvidencePath)
		if err != nil {
			return AtlasMissionDashboardProvenanceLinks{}, err
		}
		link := buildMissionDashboardProvenanceLinkFromArtifact(repo, row.Role, artifact, row.ClosureEvidenceDigest, row.FinalResponseAllowed, row.RSIRemainsDenied)
		link.DashboardRowMatched = strings.TrimSpace(row.Repo) == repo
		link.AuthorityAdvanceClaimed = row.AuthorityAdvanceClaimed
		links = append(links, link)
	}

	linksFixture := AtlasMissionDashboardProvenanceLinks{
		Schema:                            AtlasMissionDashboardProvenanceLinksContract,
		NodeID:                            nodeID,
		Status:                            "dashboard_provenance_links_bound",
		SourceDashboardBindingPath:        publicArtifactRef(dashboardBindingPath),
		SourceDashboardBindingDigest:      sourceDigest,
		SourceRowCount:                    binding.RowCount,
		RequiredRepos:                     append([]string(nil), missionDashboardRequiredProvenanceRepos...),
		ProvenanceLinkCount:               len(links),
		ProvenanceLinks:                   links,
		AllRequiredProvenanceLinked:       missionDashboardAllRequiredProvenanceLinked(links),
		EveryLinkMatchesDashboard:         missionDashboardEveryLinkMatchesDashboard(links),
		EveryLinkedArtifactDigestVerified: missionDashboardEveryLinkedArtifactDigestVerified(links),
		FinalResponseAllowed:              binding.FinalResponseAllowed,
		SchedulesWork:                     false,
		ExecutesWork:                      false,
		ApprovesWork:                      false,
		ClaimsAuthorityAdvance:            false,
		RSIRemainsDenied:                  binding.RSIRemainsDenied,
	}
	if err := ValidateAtlasMissionDashboardProvenanceLinks(linksFixture); err != nil {
		return AtlasMissionDashboardProvenanceLinks{}, err
	}
	return linksFixture, nil
}

func ValidateAtlasMissionDashboardProvenanceLinks(links AtlasMissionDashboardProvenanceLinks) error {
	var errs []string
	requireContract(&errs, "mission_dashboard_provenance_links", links.Schema, AtlasMissionDashboardProvenanceLinksContract)
	requireField(&errs, "node_id", links.NodeID)
	checkPublicPath(&errs, "node_id", links.NodeID, true)
	if links.Status != "dashboard_provenance_links_bound" {
		errs = append(errs, "status must be dashboard_provenance_links_bound")
	}
	requireField(&errs, "source_dashboard_binding_path", links.SourceDashboardBindingPath)
	checkPublicPath(&errs, "source_dashboard_binding_path", links.SourceDashboardBindingPath, true)
	if !digestPattern.MatchString(links.SourceDashboardBindingDigest) {
		errs = append(errs, "source_dashboard_binding_digest must be sha256 digest")
	}
	if links.SourceRowCount != 6 {
		errs = append(errs, "source_row_count must be 6")
	}
	if !equalStringSlices(links.RequiredRepos, missionDashboardRequiredProvenanceRepos) {
		errs = append(errs, "required_repos must list command, foundry, promoter, and sentinel in order")
	}
	if links.ProvenanceLinkCount != len(links.ProvenanceLinks) || links.ProvenanceLinkCount != len(missionDashboardRequiredProvenanceRepos) {
		errs = append(errs, "provenance_link_count must match four required subsystem links")
	}
	if !links.AllRequiredProvenanceLinked {
		errs = append(errs, "all_required_provenance_linked must be true")
	}
	if !links.EveryLinkMatchesDashboard {
		errs = append(errs, "every_link_matches_dashboard must be true")
	}
	if !links.EveryLinkedArtifactDigestVerified {
		errs = append(errs, "every_linked_artifact_digest_verified must be true")
	}
	if links.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateMissionDashboardProvenanceLinkRows(&errs, links.ProvenanceLinks)
	validateNoAuthorityEffects(&errs, links.SchedulesWork, links.ExecutesWork, links.ApprovesWork, links.ClaimsAuthorityAdvance, links.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateMissionDashboardProvenanceLinkRows(errs *[]string, links []AtlasMissionDashboardProvenanceLink) {
	expectedRoles := map[string]string{
		"ao-command":  "compact_readback_owner",
		"ao-foundry":  "bounded_implementation_handoff",
		"ao-promoter": "no_promotion_rollup_owner",
		"ao-sentinel": "public_safety_scan_owner",
	}
	seen := map[string]bool{}
	previousRepo := ""
	for i, link := range links {
		prefix := fmt.Sprintf("provenance_links[%d]", i)
		requireField(errs, prefix+".repo", link.Repo)
		requireField(errs, prefix+".role", link.Role)
		requireField(errs, prefix+".evidence_path", link.EvidencePath)
		requireField(errs, prefix+".provenance_link_status", link.ProvenanceLinkStatus)
		checkPublicPath(errs, prefix+".repo", link.Repo, true)
		checkPublicPath(errs, prefix+".role", link.Role, true)
		checkPublicPath(errs, prefix+".evidence_path", link.EvidencePath, true)
		if !digestPattern.MatchString(link.EvidenceDigest) {
			*errs = append(*errs, prefix+".evidence_digest must be sha256 digest")
		}
		if seen[link.Repo] {
			*errs = append(*errs, "provenance_links repos must be unique")
		}
		seen[link.Repo] = true
		if expectedRoles[link.Repo] != link.Role {
			*errs = append(*errs, prefix+".role must match required repo role")
		}
		if previousRepo != "" && link.Repo < previousRepo {
			*errs = append(*errs, "provenance_links must be sorted by repo")
		}
		previousRepo = link.Repo
		if link.ProvenanceLinkStatus != "linked" {
			*errs = append(*errs, prefix+".provenance_link_status must be linked")
		}
		if !link.DashboardRowMatched {
			*errs = append(*errs, prefix+".dashboard_row_matched must be true")
		}
		if !link.ClosureEvidenceDigestMatches {
			*errs = append(*errs, prefix+".closure_evidence_digest_matches must be true")
		}
		if !link.ArtifactDigestVerified {
			*errs = append(*errs, prefix+".artifact_digest_verified must be true")
		}
		if link.FinalResponseAllowed {
			*errs = append(*errs, prefix+".final_response_allowed must be false")
		}
		if !link.RSIRemainsDenied {
			*errs = append(*errs, prefix+".rsi_remains_denied must be true")
		}
		if link.AuthorityAdvanceClaimed {
			*errs = append(*errs, prefix+".authority_advance_claimed must be false")
		}
	}
	for _, repo := range missionDashboardRequiredProvenanceRepos {
		if !seen[repo] {
			*errs = append(*errs, "provenance_links missing "+repo)
		}
	}
}

func missionDashboardAllRequiredProvenanceLinked(links []AtlasMissionDashboardProvenanceLink) bool {
	if len(links) != len(missionDashboardRequiredProvenanceRepos) {
		return false
	}
	for i, repo := range missionDashboardRequiredProvenanceRepos {
		if links[i].Repo != repo || links[i].ProvenanceLinkStatus != "linked" {
			return false
		}
	}
	return true
}

func missionDashboardEveryLinkMatchesDashboard(links []AtlasMissionDashboardProvenanceLink) bool {
	for _, link := range links {
		if !link.DashboardRowMatched || !link.ClosureEvidenceDigestMatches {
			return false
		}
	}
	return len(links) > 0
}

func missionDashboardEveryLinkedArtifactDigestVerified(links []AtlasMissionDashboardProvenanceLink) bool {
	for _, link := range links {
		if !link.ArtifactDigestVerified || !digestPattern.MatchString(link.EvidenceDigest) {
			return false
		}
	}
	return len(links) > 0
}

func WriteAtlasMissionDashboardProvenanceLinks(path string, links AtlasMissionDashboardProvenanceLinks) error {
	return WriteJSON(path, links)
}
