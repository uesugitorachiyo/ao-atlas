package atlas

import (
	"fmt"
	"strings"
)

type atlasDashboardFreshnessPostMergeLifecycle struct {
	Schema                       string `json:"schema"`
	NodeID                       string `json:"node_id"`
	Status                       string `json:"status"`
	PRNumber                     int    `json:"pr_number"`
	MergeCommit                  string `json:"merge_commit"`
	CIStatus                     string `json:"ci_status"`
	LocalMainSynced              bool   `json:"local_main_synced"`
	OriginMainSynced             bool   `json:"origin_main_synced"`
	LocalBranchDeleted           bool   `json:"local_branch_deleted"`
	RemoteBranchDeleted          bool   `json:"remote_branch_deleted"`
	LocalCodexBranchesRemaining  int    `json:"local_codex_branches_remaining"`
	RemoteCodexBranchesRemaining int    `json:"remote_codex_branches_remaining"`
	FinalHead                    string `json:"final_head"`
}

func BuildAtlasMissionDashboardFreshnessChecks(nodeID, provenanceLinksPath, sourceReadbackPath, postMergeLifecyclePath string) (AtlasMissionDashboardFreshnessChecks, error) {
	nodeID = strings.TrimSpace(nodeID)
	provenanceLinksPath = strings.TrimSpace(provenanceLinksPath)
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	postMergeLifecyclePath = strings.TrimSpace(postMergeLifecyclePath)
	for name, value := range map[string]string{
		"node id":                   nodeID,
		"provenance links path":     provenanceLinksPath,
		"source readback path":      sourceReadbackPath,
		"post merge lifecycle path": postMergeLifecyclePath,
	} {
		if value == "" {
			return AtlasMissionDashboardFreshnessChecks{}, fmt.Errorf("%s is required", name)
		}
	}
	provenance, err := LoadJSON[AtlasMissionDashboardProvenanceLinks](provenanceLinksPath)
	if err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}
	if err := ValidateAtlasMissionDashboardProvenanceLinks(provenance); err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}
	lifecycle, err := LoadJSON[atlasDashboardFreshnessPostMergeLifecycle](postMergeLifecyclePath)
	if err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}
	provenanceArtifact, err := buildMissionDashboardEvidenceArtifact(provenanceLinksPath)
	if err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}
	readbackArtifact, err := buildMissionDashboardEvidenceArtifact(sourceReadbackPath)
	if err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}
	lifecycleArtifact, err := buildMissionDashboardEvidenceArtifact(postMergeLifecyclePath)
	if err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}

	prMergedAndCleaned := lifecycle.Status == "merged_and_cleaned" && lifecycle.PRNumber > 0 && lifecycle.CIStatus == "passed"
	mainSynced := lifecycle.LocalMainSynced && lifecycle.OriginMainSynced && lifecycle.MergeCommit != "" && lifecycle.FinalHead == lifecycle.MergeCommit
	branchesClean := lifecycle.LocalBranchDeleted && lifecycle.RemoteBranchDeleted && lifecycle.LocalCodexBranchesRemaining == 0 && lifecycle.RemoteCodexBranchesRemaining == 0
	dashboardFresh := provenance.NodeID == lifecycle.NodeID && readback.CompletedNodes == 34 && readback.FirstExecutableNode == "mission-recommendation-feature-depth-next-wave-35" && !readback.FinalResponseAllowed

	checks := []AtlasMissionDashboardFreshnessCheck{
		buildMissionDashboardFreshnessCheckFromArtifact("pr_merged", prMergedAndCleaned, lifecycleArtifact),
		buildMissionDashboardFreshnessCheckFromArtifact("ci_passed", lifecycle.CIStatus == "passed", lifecycleArtifact),
		buildMissionDashboardFreshnessCheckFromArtifact("main_synced_to_merge_commit", mainSynced, lifecycleArtifact),
		buildMissionDashboardFreshnessCheckFromArtifact("branches_deleted", lifecycle.LocalBranchDeleted && lifecycle.RemoteBranchDeleted, lifecycleArtifact),
		buildMissionDashboardFreshnessCheckFromArtifact("codex_branches_clean", branchesClean, lifecycleArtifact),
		buildMissionDashboardFreshnessCheckFromArtifact("dashboard_sources_match_completed_readback", dashboardFresh, provenanceArtifact),
	}

	fixture := AtlasMissionDashboardFreshnessChecks{
		Schema:                       AtlasMissionDashboardFreshnessChecksContract,
		NodeID:                       nodeID,
		Status:                       "dashboard_freshness_verified",
		SourceProvenanceLinksPath:    provenanceArtifact.PublicPath,
		SourceProvenanceLinksDigest:  provenanceArtifact.Digest,
		SourceReadbackPath:           readbackArtifact.PublicPath,
		SourceReadbackDigest:         readbackArtifact.Digest,
		PostMergeLifecyclePath:       lifecycleArtifact.PublicPath,
		PostMergeLifecycleDigest:     lifecycleArtifact.Digest,
		SourceCompletedNodes:         readback.CompletedNodes,
		SourceReadyNodes:             readback.ReadyNodes,
		SourceFirstExecutableNode:    readback.FirstExecutableNode,
		PRNumber:                     lifecycle.PRNumber,
		MergeCommit:                  lifecycle.MergeCommit,
		FinalHead:                    lifecycle.FinalHead,
		LocalMainSynced:              lifecycle.LocalMainSynced,
		OriginMainSynced:             lifecycle.OriginMainSynced,
		LocalBranchDeleted:           lifecycle.LocalBranchDeleted,
		RemoteBranchDeleted:          lifecycle.RemoteBranchDeleted,
		LocalCodexBranchesRemaining:  lifecycle.LocalCodexBranchesRemaining,
		RemoteCodexBranchesRemaining: lifecycle.RemoteCodexBranchesRemaining,
		FreshnessCheckCount:          len(checks),
		FreshnessChecks:              checks,
		PRMergedAndCleaned:           prMergedAndCleaned,
		MainSyncedToMergeCommit:      mainSynced,
		DashboardSourceStillFresh:    dashboardFresh,
		AllFreshnessChecksPassed:     missionDashboardFreshnessChecksPassed(checks),
		FinalResponseAllowed:         readback.FinalResponseAllowed,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		ClaimsAuthorityAdvance:       false,
		RSIRemainsDenied:             readback.SafetyBoundaries["rsi_remains_denied"] && provenance.RSIRemainsDenied,
	}
	if err := ValidateAtlasMissionDashboardFreshnessChecks(fixture); err != nil {
		return AtlasMissionDashboardFreshnessChecks{}, err
	}
	return fixture, nil
}

func ValidateAtlasMissionDashboardFreshnessChecks(fixture AtlasMissionDashboardFreshnessChecks) error {
	var errs []string
	requireContract(&errs, "mission_dashboard_freshness_checks", fixture.Schema, AtlasMissionDashboardFreshnessChecksContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if fixture.Status != "dashboard_freshness_verified" {
		errs = append(errs, "status must be dashboard_freshness_verified")
	}
	for field, value := range map[string]string{
		"source_provenance_links_path": fixture.SourceProvenanceLinksPath,
		"source_readback_path":         fixture.SourceReadbackPath,
		"post_merge_lifecycle_path":    fixture.PostMergeLifecyclePath,
		"source_first_executable_node": fixture.SourceFirstExecutableNode,
		"merge_commit":                 fixture.MergeCommit,
		"final_head":                   fixture.FinalHead,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"source_provenance_links_digest": fixture.SourceProvenanceLinksDigest,
		"source_readback_digest":         fixture.SourceReadbackDigest,
		"post_merge_lifecycle_digest":    fixture.PostMergeLifecycleDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if fixture.SourceCompletedNodes != 34 || fixture.SourceReadyNodes != 6 {
		errs = append(errs, "source node counts must be completed_nodes=34 and ready_nodes=6")
	}
	if fixture.SourceFirstExecutableNode != "mission-recommendation-feature-depth-next-wave-35" {
		errs = append(errs, "source_first_executable_node must be mission-recommendation-feature-depth-next-wave-35")
	}
	if fixture.PRNumber <= 0 {
		errs = append(errs, "pr_number must be positive")
	}
	if fixture.MergeCommit != fixture.FinalHead {
		errs = append(errs, "merge_commit must match final_head")
	}
	if !fixture.LocalMainSynced || !fixture.OriginMainSynced {
		errs = append(errs, "local and origin main must be synced")
	}
	if !fixture.LocalBranchDeleted || !fixture.RemoteBranchDeleted {
		errs = append(errs, "local and remote node branches must be deleted")
	}
	if fixture.LocalCodexBranchesRemaining != 0 || fixture.RemoteCodexBranchesRemaining != 0 {
		errs = append(errs, "codex branch counts must be zero")
	}
	if fixture.FreshnessCheckCount != len(fixture.FreshnessChecks) || fixture.FreshnessCheckCount != 6 {
		errs = append(errs, "freshness_check_count must be 6")
	}
	validateMissionDashboardFreshnessCheckRows(&errs, fixture.FreshnessChecks)
	if !fixture.PRMergedAndCleaned {
		errs = append(errs, "pr_merged_and_cleaned must be true")
	}
	if !fixture.MainSyncedToMergeCommit {
		errs = append(errs, "main_synced_to_merge_commit must be true")
	}
	if !fixture.DashboardSourceStillFresh {
		errs = append(errs, "dashboard_source_still_fresh must be true")
	}
	if !fixture.AllFreshnessChecksPassed {
		errs = append(errs, "all_freshness_checks_passed must be true")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateMissionDashboardFreshnessCheckRows(errs *[]string, checks []AtlasMissionDashboardFreshnessCheck) {
	expected := []string{
		"pr_merged",
		"ci_passed",
		"main_synced_to_merge_commit",
		"branches_deleted",
		"codex_branches_clean",
		"dashboard_sources_match_completed_readback",
	}
	if len(checks) != len(expected) {
		*errs = append(*errs, "freshness_checks must contain six checks")
		return
	}
	for i, check := range checks {
		prefix := fmt.Sprintf("freshness_checks[%d]", i)
		if check.Name != expected[i] {
			*errs = append(*errs, prefix+".name must be "+expected[i])
		}
		if check.Status != "passed" {
			*errs = append(*errs, prefix+".status must be passed")
		}
		requireField(errs, prefix+".evidence_path", check.EvidencePath)
		checkPublicPath(errs, prefix+".evidence_path", check.EvidencePath, true)
		if !digestPattern.MatchString(check.EvidenceDigest) {
			*errs = append(*errs, prefix+".evidence_digest must be sha256 digest")
		}
	}
}

func missionDashboardFreshnessChecksPassed(checks []AtlasMissionDashboardFreshnessCheck) bool {
	for _, check := range checks {
		if check.Status != "passed" {
			return false
		}
	}
	return len(checks) == 6
}

func WriteAtlasMissionDashboardFreshnessChecks(path string, fixture AtlasMissionDashboardFreshnessChecks) error {
	return WriteJSON(path, fixture)
}
