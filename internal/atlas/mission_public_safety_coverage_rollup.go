package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildAtlasPublicSafetyCoverageRollup(nodeID, sourceReadbackPath, evidenceRoot string) (AtlasPublicSafetyCoverageRollup, error) {
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	evidenceRoot = strings.TrimSpace(evidenceRoot)
	if sourceReadbackPath == "" {
		return AtlasPublicSafetyCoverageRollup{}, fmt.Errorf("source readback path is required")
	}
	if evidenceRoot == "" {
		return AtlasPublicSafetyCoverageRollup{}, fmt.Errorf("evidence root is required")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasPublicSafetyCoverageRollup{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasPublicSafetyCoverageRollup{}, err
	}
	readbackDigest, err := digestFile(sourceReadbackPath)
	if err != nil {
		return AtlasPublicSafetyCoverageRollup{}, err
	}

	completedNodeIDs := []string{}
	for _, evidence := range readback.NodeEvidence {
		if evidence.Status == "completed" {
			completedNodeIDs = append(completedNodeIDs, evidence.NodeID)
		}
	}
	sentinelFiles := []string{}
	scopedScanFiles := []string{}
	missingSentinelNodes := []string{}
	completedNodesWithSentinel := 0
	allSentinelPassed := true
	allScopedPassed := true
	scannedFileCountTotal := 0
	changedEvidenceFilesTotal := 0
	changedPromptArtifactsTotal := 0
	unsafeMatchCountTotal := 0

	for _, completedNodeID := range completedNodeIDs {
		nodeDir := filepath.Join(evidenceRoot, "nodes", completedNodeID)
		sentinelPath := filepath.Join(nodeDir, "sentinel_public_safety.json")
		if _, err := os.Stat(sentinelPath); err != nil {
			missingSentinelNodes = append(missingSentinelNodes, completedNodeID)
			allSentinelPassed = false
		} else {
			sentinel, err := LoadJSON[AtlasNodeSentinelPublicSafetyEvidence](sentinelPath)
			if err != nil {
				return AtlasPublicSafetyCoverageRollup{}, err
			}
			if err := ValidateAtlasNodeSentinelPublicSafetyEvidence(sentinel); err != nil {
				return AtlasPublicSafetyCoverageRollup{}, err
			}
			if sentinel.Status != "passed" || sentinel.UnsafePublicClaimDetected || sentinel.PromotionClaimDetected || sentinel.RSIClaimDetected || !sentinel.RSIRemainsDenied {
				allSentinelPassed = false
			}
			completedNodesWithSentinel++
			sentinelFiles = append(sentinelFiles, publicArtifactRef(sentinelPath))
		}

		scopedScanPath := filepath.Join(nodeDir, "scoped-public-safety-scan.json")
		if _, err := os.Stat(scopedScanPath); err == nil {
			scan, err := LoadJSON[AtlasScopedPublicSafetyScan](scopedScanPath)
			if err != nil {
				return AtlasPublicSafetyCoverageRollup{}, err
			}
			if err := ValidateAtlasScopedPublicSafetyScan(scan); err != nil {
				return AtlasPublicSafetyCoverageRollup{}, err
			}
			if scan.Status != "passed" || !scan.PublicSafetyScanPassed || scan.UnsafeMatchCount != 0 || !scan.RSIRemainsDenied {
				allScopedPassed = false
			}
			scannedFileCountTotal += scan.ScannedFileCount
			changedEvidenceFilesTotal += scan.ChangedEvidenceFiles
			changedPromptArtifactsTotal += scan.ChangedPromptArtifacts
			unsafeMatchCountTotal += scan.UnsafeMatchCount
			scopedScanFiles = append(scopedScanFiles, publicArtifactRef(scopedScanPath))
		}
	}

	allCompletedCovered := len(missingSentinelNodes) == 0 && completedNodesWithSentinel == readback.CompletedNodes
	publicSafetyScanPassed := readback.PublicSafetyScanStatus == "passed" && allCompletedCovered && allSentinelPassed && allScopedPassed && unsafeMatchCountTotal == 0
	status := "passed"
	if !publicSafetyScanPassed {
		status = "failed"
	}
	rollup := AtlasPublicSafetyCoverageRollup{
		Schema:                       AtlasPublicSafetyCoverageRollupContract,
		NodeID:                       strings.TrimSpace(nodeID),
		Status:                       status,
		SourceReadbackPath:           publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:         readbackDigest,
		EvidenceRoot:                 publicArtifactRef(evidenceRoot),
		CompletedNodesBefore:         readback.CompletedNodes,
		ReadyNodesBefore:             readback.ReadyNodes,
		FirstExecutableNodeBefore:    readback.FirstExecutableNode,
		FinalResponseAllowedBefore:   readback.FinalResponseAllowed,
		ExactNextActionBefore:        readback.ExactNextAction,
		PublicSafetyScanStatus:       readback.PublicSafetyScanStatus,
		SentinelEvidenceCount:        len(sentinelFiles),
		CompletedNodesWithSentinel:   completedNodesWithSentinel,
		MissingSentinelNodes:         missingSentinelNodes,
		ScopedScanCount:              len(scopedScanFiles),
		SentinelEvidenceFiles:        sentinelFiles,
		ScopedScanFiles:              scopedScanFiles,
		ScannedFileCountTotal:        scannedFileCountTotal,
		ChangedEvidenceFilesTotal:    changedEvidenceFilesTotal,
		ChangedPromptArtifactsTotal:  changedPromptArtifactsTotal,
		UnsafeMatchCountTotal:        unsafeMatchCountTotal,
		AllCompletedNodesCovered:     allCompletedCovered,
		AllSentinelStatusesPassed:    allSentinelPassed,
		AllScopedScansPassed:         allScopedPassed,
		PublicSafetyScanPassed:       publicSafetyScanPassed,
		MachineReadableClosureRollup: true,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		ClaimsAuthorityAdvance:       false,
		RSIRemainsDenied:             true,
	}
	if err := ValidateAtlasPublicSafetyCoverageRollup(rollup); err != nil {
		return AtlasPublicSafetyCoverageRollup{}, err
	}
	return rollup, nil
}

func ValidateAtlasPublicSafetyCoverageRollup(rollup AtlasPublicSafetyCoverageRollup) error {
	var errs []string
	requireContract(&errs, "public_safety_coverage_rollup", rollup.Schema, AtlasPublicSafetyCoverageRollupContract)
	requireField(&errs, "node_id", rollup.NodeID)
	checkPublicPath(&errs, "node_id", rollup.NodeID, true)
	if !oneOf(rollup.Status, "passed", "failed") {
		errs = append(errs, "status must be passed or failed")
	}
	requireField(&errs, "source_readback_path", rollup.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", rollup.SourceReadbackPath, true)
	if !digestPattern.MatchString(rollup.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	requireField(&errs, "evidence_root", rollup.EvidenceRoot)
	checkPublicPath(&errs, "evidence_root", rollup.EvidenceRoot, true)
	if rollup.CompletedNodesBefore <= 0 {
		errs = append(errs, "completed_nodes_before must be positive")
	}
	if rollup.ReadyNodesBefore < 0 {
		errs = append(errs, "ready_nodes_before must not be negative")
	}
	requireField(&errs, "first_executable_node_before", rollup.FirstExecutableNodeBefore)
	if rollup.FinalResponseAllowedBefore {
		errs = append(errs, "final_response_allowed_before must be false")
	}
	requireField(&errs, "exact_next_action_before", rollup.ExactNextActionBefore)
	if !oneOf(rollup.PublicSafetyScanStatus, "passed", "required_pending_verification", "failed") {
		errs = append(errs, "public_safety_scan_status must be passed, failed, or required_pending_verification")
	}
	if rollup.SentinelEvidenceCount != len(rollup.SentinelEvidenceFiles) {
		errs = append(errs, "sentinel_evidence_count must match sentinel_evidence_files length")
	}
	if rollup.SentinelEvidenceCount <= 0 {
		errs = append(errs, "sentinel_evidence_count must be positive")
	}
	if rollup.CompletedNodesWithSentinel+len(rollup.MissingSentinelNodes) != rollup.CompletedNodesBefore {
		errs = append(errs, "completed sentinel coverage must account for every completed node")
	}
	if rollup.AllCompletedNodesCovered != (len(rollup.MissingSentinelNodes) == 0 && rollup.CompletedNodesWithSentinel == rollup.CompletedNodesBefore) {
		errs = append(errs, "all_completed_nodes_covered must match missing sentinel nodes")
	}
	if rollup.ScopedScanCount != len(rollup.ScopedScanFiles) {
		errs = append(errs, "scoped_scan_count must match scoped_scan_files length")
	}
	if rollup.ScopedScanCount <= 0 {
		errs = append(errs, "scoped_scan_count must be positive")
	}
	if rollup.ScannedFileCountTotal <= 0 {
		errs = append(errs, "scanned_file_count_total must be positive")
	}
	if rollup.ChangedEvidenceFilesTotal <= 0 {
		errs = append(errs, "changed_evidence_files_total must be positive")
	}
	if rollup.ChangedPromptArtifactsTotal <= 0 {
		errs = append(errs, "changed_prompt_artifacts_total must be positive")
	}
	if rollup.UnsafeMatchCountTotal < 0 {
		errs = append(errs, "unsafe_match_count_total must not be negative")
	}
	if rollup.PublicSafetyScanPassed != (rollup.PublicSafetyScanStatus == "passed" && rollup.AllCompletedNodesCovered && rollup.AllSentinelStatusesPassed && rollup.AllScopedScansPassed && rollup.UnsafeMatchCountTotal == 0) {
		errs = append(errs, "public_safety_scan_passed must match coverage and scan status")
	}
	if rollup.Status == "passed" && !rollup.PublicSafetyScanPassed {
		errs = append(errs, "passed rollup requires public_safety_scan_passed")
	}
	if !rollup.MachineReadableClosureRollup {
		errs = append(errs, "machine_readable_closure_rollup must be true")
	}
	checkPublicStrings(&errs, "missing_sentinel_nodes", rollup.MissingSentinelNodes, true)
	checkPublicStrings(&errs, "sentinel_evidence_files", rollup.SentinelEvidenceFiles, true)
	checkPublicStrings(&errs, "scoped_scan_files", rollup.ScopedScanFiles, true)
	validateNoAuthorityEffects(&errs, rollup.SchedulesWork, rollup.ExecutesWork, rollup.ApprovesWork, rollup.ClaimsAuthorityAdvance, rollup.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasPublicSafetyCoverageRollup(path string, rollup AtlasPublicSafetyCoverageRollup) error {
	return WriteJSON(path, rollup)
}
