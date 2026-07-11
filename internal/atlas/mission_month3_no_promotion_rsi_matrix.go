package atlas

import (
	"fmt"
	"path/filepath"
	"strings"
)

func BuildAtlasMonth3NoPromotionRSIMatrix(nodeID, sourceReadbackPath, sourceWorkgraphPath, evidenceRoot, expectedNextNode string, readback AtlasRecommendationReadback, workgraph Workgraph) (AtlasMonth3NoPromotionRSIMatrix, error) {
	completed := completedRecommendationNodeIDs(workgraph)
	entries := make([]AtlasMonth3NoPromotionRSIMatrixEntry, 0, len(completed))
	for _, completedNode := range completed {
		nodeRoot := filepath.Join(evidenceRoot, "nodes", completedNode)
		promoterPath := filepath.Join(nodeRoot, "promoter_no_promotion.json")
		commandPath := filepath.Join(nodeRoot, "command_readback.json")
		sentinelPath := filepath.Join(nodeRoot, "sentinel_public_safety.json")
		promoter, err := LoadJSON[AtlasNodePromoterNoPromotionEvidence](promoterPath)
		if err != nil {
			return AtlasMonth3NoPromotionRSIMatrix{}, err
		}
		command, err := LoadJSON[AtlasNodeCommandReadbackEvidence](commandPath)
		if err != nil {
			return AtlasMonth3NoPromotionRSIMatrix{}, err
		}
		sentinel, err := LoadJSON[AtlasNodeSentinelPublicSafetyEvidence](sentinelPath)
		if err != nil {
			return AtlasMonth3NoPromotionRSIMatrix{}, err
		}
		noPromotion := promoterNoPromotionStatusAllowed(promoter.Status) &&
			!promoter.PromotionClaimed &&
			!promoter.ClaimsAuthorityAdvance &&
			strings.HasPrefix(command.Status, "readback_agrees_no_promotion") &&
			!command.FinalResponseAllowed
		rsiDenied := sentinel.RSIRemainsDenied && promoter.RSIRemainsDenied && command.RSIRemainsDenied
		entries = append(entries, AtlasMonth3NoPromotionRSIMatrixEntry{
			NodeID:                     completedNode,
			PromoterPath:               publicArtifactRef(promoterPath),
			PromoterStatus:             promoter.Status,
			PromotionClaimed:           promoter.PromotionClaimed,
			CommandReadbackPath:        publicArtifactRef(commandPath),
			CommandReadbackStatus:      command.Status,
			SentinelPublicSafetyPath:   publicArtifactRef(sentinelPath),
			SentinelPublicSafetyStatus: sentinel.Status,
			SentinelRSIRemainsDenied:   sentinel.RSIRemainsDenied,
			ClaimsAuthorityAdvance:     promoter.ClaimsAuthorityAdvance,
			NoPromotionInvariantHolds:  noPromotion,
			RSIDenialInvariantHolds:    rsiDenied,
		})
	}
	matrix := AtlasMonth3NoPromotionRSIMatrix{
		Schema:                          AtlasMonth3NoPromotionRSIMatrixContract,
		NodeID:                          strings.TrimSpace(nodeID),
		Status:                          "asserted",
		SourceReadbackPath:              publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:            digestValue(readback),
		SourceWorkgraphPath:             publicArtifactRef(sourceWorkgraphPath),
		SourceWorkgraphDigest:           digestValue(workgraph),
		EvidenceRoot:                    publicArtifactRef(evidenceRoot),
		CompletedNodes:                  len(entries),
		PromoterNoPromotionFiles:        len(entries),
		CommandReadbackFiles:            len(entries),
		SentinelPublicSafetyFiles:       len(entries),
		AllowedPromoterStatuses:         []string{"no_promotion_requested", "no_promotion", "recorded"},
		AllowedCommandStatusPrefixes:    []string{"readback_agrees_no_promotion", "recorded"},
		Entries:                         entries,
		ExpectedNextNodeAfterCompletion: strings.TrimSpace(expectedNextNode),
		PromotionRequested:              false,
		PromotionGranted:                false,
		ClaimsAuthorityAdvance:          false,
		RSIRemainsDenied:                true,
	}
	for _, entry := range entries {
		if !entry.PromotionClaimed {
			matrix.PromotionRequestedFalseCount++
			matrix.PromotionGrantedFalseCount++
		}
		if entry.SentinelRSIRemainsDenied {
			matrix.SentinelRSIDeniedCount++
		}
	}
	matrix.NoPromotionInvariantHolds = matrix.CompletedNodes > 0 &&
		matrix.PromotionRequestedFalseCount == matrix.CompletedNodes &&
		matrix.PromotionGrantedFalseCount == matrix.CompletedNodes &&
		allMonth3NoPromotionEntriesHold(entries)
	matrix.RSIDenialInvariantHolds = matrix.CompletedNodes > 0 &&
		matrix.SentinelRSIDeniedCount == matrix.CompletedNodes &&
		allMonth3RSIEntriesHold(entries)
	if err := ValidateAtlasMonth3NoPromotionRSIMatrix(matrix); err != nil {
		return AtlasMonth3NoPromotionRSIMatrix{}, err
	}
	return matrix, nil
}

func completedRecommendationNodeIDs(workgraph Workgraph) []string {
	completed := []string{}
	for _, node := range workgraph.Nodes {
		if node.Status == "completed" {
			completed = append(completed, node.ID)
		}
	}
	return completed
}

func promoterNoPromotionStatusAllowed(status string) bool {
	return status == "no_promotion_requested" || status == "no_promotion" || status == "recorded"
}

func allMonth3NoPromotionEntriesHold(entries []AtlasMonth3NoPromotionRSIMatrixEntry) bool {
	for _, entry := range entries {
		if !entry.NoPromotionInvariantHolds {
			return false
		}
	}
	return true
}

func allMonth3RSIEntriesHold(entries []AtlasMonth3NoPromotionRSIMatrixEntry) bool {
	for _, entry := range entries {
		if !entry.RSIDenialInvariantHolds {
			return false
		}
	}
	return true
}

func ValidateAtlasMonth3NoPromotionRSIMatrix(matrix AtlasMonth3NoPromotionRSIMatrix) error {
	var errs []string
	requireContract(&errs, "month3_no_promotion_rsi_matrix", matrix.Schema, AtlasMonth3NoPromotionRSIMatrixContract)
	for field, value := range map[string]string{
		"node_id":                             matrix.NodeID,
		"status":                              matrix.Status,
		"source_readback_path":                matrix.SourceReadbackPath,
		"source_readback_digest":              matrix.SourceReadbackDigest,
		"source_workgraph_path":               matrix.SourceWorkgraphPath,
		"source_workgraph_digest":             matrix.SourceWorkgraphDigest,
		"evidence_root":                       matrix.EvidenceRoot,
		"expected_next_node_after_completion": matrix.ExpectedNextNodeAfterCompletion,
	} {
		requireField(&errs, field, value)
	}
	checkPublicPath(&errs, "node_id", matrix.NodeID, true)
	checkPublicPath(&errs, "source_readback_path", matrix.SourceReadbackPath, true)
	checkPublicPath(&errs, "source_workgraph_path", matrix.SourceWorkgraphPath, true)
	checkPublicPath(&errs, "evidence_root", matrix.EvidenceRoot, true)
	checkPublicPath(&errs, "expected_next_node_after_completion", matrix.ExpectedNextNodeAfterCompletion, true)
	checkOptionalDigest(&errs, "source_readback_digest", matrix.SourceReadbackDigest)
	checkOptionalDigest(&errs, "source_workgraph_digest", matrix.SourceWorkgraphDigest)
	if matrix.Status != "asserted" {
		errs = append(errs, "status must be asserted")
	}
	if matrix.CompletedNodes <= 0 || matrix.CompletedNodes != len(matrix.Entries) {
		errs = append(errs, "completed_nodes must match entries")
	}
	if matrix.PromoterNoPromotionFiles != matrix.CompletedNodes ||
		matrix.CommandReadbackFiles != matrix.CompletedNodes ||
		matrix.SentinelPublicSafetyFiles != matrix.CompletedNodes {
		errs = append(errs, "evidence file counts must match completed_nodes")
	}
	if matrix.PromotionRequestedFalseCount != matrix.CompletedNodes ||
		matrix.PromotionGrantedFalseCount != matrix.CompletedNodes ||
		matrix.SentinelRSIDeniedCount != matrix.CompletedNodes {
		errs = append(errs, "safety counts must match completed_nodes")
	}
	if !containsStringValue(matrix.AllowedPromoterStatuses, "no_promotion_requested") {
		errs = append(errs, "allowed_promoter_statuses must include no_promotion_requested")
	}
	if !containsStringValue(matrix.AllowedCommandStatusPrefixes, "readback_agrees_no_promotion") {
		errs = append(errs, "allowed_command_status_prefixes must include readback_agrees_no_promotion")
	}
	for i, entry := range matrix.Entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		for field, value := range map[string]string{
			prefix + ".node_id":                       entry.NodeID,
			prefix + ".promoter_path":                 entry.PromoterPath,
			prefix + ".promoter_status":               entry.PromoterStatus,
			prefix + ".command_readback_path":         entry.CommandReadbackPath,
			prefix + ".command_readback_status":       entry.CommandReadbackStatus,
			prefix + ".sentinel_public_safety_path":   entry.SentinelPublicSafetyPath,
			prefix + ".sentinel_public_safety_status": entry.SentinelPublicSafetyStatus,
		} {
			requireField(&errs, field, value)
		}
		checkPublicPath(&errs, prefix+".node_id", entry.NodeID, true)
		checkPublicPath(&errs, prefix+".promoter_path", entry.PromoterPath, true)
		checkPublicPath(&errs, prefix+".command_readback_path", entry.CommandReadbackPath, true)
		checkPublicPath(&errs, prefix+".sentinel_public_safety_path", entry.SentinelPublicSafetyPath, true)
		if !promoterNoPromotionStatusAllowed(entry.PromoterStatus) {
			errs = append(errs, prefix+".promoter_status must be an allowed no-promotion status")
		}
		if !strings.HasPrefix(entry.CommandReadbackStatus, "readback_agrees_no_promotion") && entry.CommandReadbackStatus != "recorded" {
			errs = append(errs, prefix+".command_readback_status must be an allowed no-promotion readback")
		}
		if entry.PromotionClaimed {
			errs = append(errs, prefix+".promotion_claimed must be false")
		}
		if entry.ClaimsAuthorityAdvance {
			errs = append(errs, prefix+".claims_authority_advance must be false")
		}
		if !entry.SentinelRSIRemainsDenied || !entry.NoPromotionInvariantHolds || !entry.RSIDenialInvariantHolds {
			errs = append(errs, prefix+" must preserve no-promotion and RSI denial invariants")
		}
	}
	if !matrix.NoPromotionInvariantHolds {
		errs = append(errs, "no_promotion_invariant_holds must be true")
	}
	if !matrix.RSIDenialInvariantHolds {
		errs = append(errs, "rsi_denial_invariant_holds must be true")
	}
	if matrix.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if matrix.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, false, false, false, matrix.ClaimsAuthorityAdvance, matrix.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3NoPromotionRSIMatrix(path string, matrix AtlasMonth3NoPromotionRSIMatrix) error {
	return WriteJSON(path, matrix)
}
