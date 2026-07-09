package atlas

import (
	"fmt"
	"strings"
)

type AtlasMissionDashboardStaleEvidenceDetection struct {
	Schema                      string                                   `json:"schema"`
	NodeID                      string                                   `json:"node_id"`
	Status                      string                                   `json:"status"`
	LatestReadbackPath          string                                   `json:"latest_readback_path"`
	LatestReadbackDigest        string                                   `json:"latest_readback_digest"`
	LatestCompletedNodes        int                                      `json:"latest_completed_nodes"`
	DashboardEvidenceCount      int                                      `json:"dashboard_evidence_count"`
	StaleDashboardEvidenceCount int                                      `json:"stale_dashboard_evidence_count"`
	FreshDashboardEvidenceCount int                                      `json:"fresh_dashboard_evidence_count"`
	StaleDashboardEvidence      []AtlasMissionDashboardStaleEvidenceItem `json:"stale_dashboard_evidence"`
	OldExportRootCount          int                                      `json:"old_export_root_count"`
	OldExportRoots              []AtlasMissionDashboardStaleExportRoot   `json:"old_export_roots"`
	RecommendedAction           string                                   `json:"recommended_action"`
	PromotionRequested          bool                                     `json:"promotion_requested"`
	PromotionGranted            bool                                     `json:"promotion_granted"`
	ClaimsAuthorityAdvance      bool                                     `json:"claims_authority_advance"`
	SchedulesWork               bool                                     `json:"schedules_work"`
	ExecutesWork                bool                                     `json:"executes_work"`
	ApprovesWork                bool                                     `json:"approves_work"`
	RSIRemainsDenied            bool                                     `json:"rsi_remains_denied"`
}

type AtlasMissionDashboardStaleEvidenceItem struct {
	Path                 string `json:"path"`
	Schema               string `json:"schema"`
	SourceReadbackPath   string `json:"source_readback_path"`
	SourceReadbackDigest string `json:"source_readback_digest"`
	LatestReadbackDigest string `json:"latest_readback_digest"`
	Stale                bool   `json:"stale"`
	StaleReason          string `json:"stale_reason"`
}

type AtlasMissionDashboardStaleExportRoot struct {
	Path        string `json:"path"`
	Stale       bool   `json:"stale"`
	StaleReason string `json:"stale_reason"`
}

func BuildAtlasMissionDashboardStaleEvidenceDetection(nodeID, latestReadbackPath string, dashboardPaths, oldExportRoots []string) (AtlasMissionDashboardStaleEvidenceDetection, error) {
	nodeID = strings.TrimSpace(nodeID)
	latestReadbackPath = strings.TrimSpace(latestReadbackPath)
	if nodeID == "" {
		return AtlasMissionDashboardStaleEvidenceDetection{}, fmt.Errorf("node id is required")
	}
	if latestReadbackPath == "" {
		return AtlasMissionDashboardStaleEvidenceDetection{}, fmt.Errorf("latest readback path is required")
	}
	latest, err := LoadJSON[AtlasRecommendationReadback](latestReadbackPath)
	if err != nil {
		return AtlasMissionDashboardStaleEvidenceDetection{}, err
	}
	if err := ValidateAtlasRecommendationReadback(latest); err != nil {
		return AtlasMissionDashboardStaleEvidenceDetection{}, err
	}
	latestDigest, err := digestTextFileWithNormalizedLineEndings(latestReadbackPath)
	if err != nil {
		return AtlasMissionDashboardStaleEvidenceDetection{}, err
	}

	staleItems := []AtlasMissionDashboardStaleEvidenceItem{}
	freshCount := 0
	for _, path := range dashboardPaths {
		item, err := buildMissionDashboardStaleEvidenceItem(path, latestDigest)
		if err != nil {
			return AtlasMissionDashboardStaleEvidenceDetection{}, err
		}
		if item.Stale {
			staleItems = append(staleItems, item)
		} else {
			freshCount++
		}
	}

	oldRoots := make([]AtlasMissionDashboardStaleExportRoot, 0, len(oldExportRoots))
	for _, root := range oldExportRoots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		oldRoots = append(oldRoots, AtlasMissionDashboardStaleExportRoot{
			Path:        publicArtifactRef(root),
			Stale:       true,
			StaleReason: "old_export_root_superseded_by_latest_readback",
		})
	}

	status := "dashboard_evidence_fresh"
	if len(staleItems) > 0 || len(oldRoots) > 0 {
		status = "stale_dashboard_evidence_detected"
	}
	return AtlasMissionDashboardStaleEvidenceDetection{
		Schema:                      "ao.atlas.mission-dashboard-stale-evidence-detection.v0.1",
		NodeID:                      nodeID,
		Status:                      status,
		LatestReadbackPath:          publicArtifactRef(latestReadbackPath),
		LatestReadbackDigest:        latestDigest,
		LatestCompletedNodes:        latest.CompletedNodes,
		DashboardEvidenceCount:      len(dashboardPaths),
		StaleDashboardEvidenceCount: len(staleItems),
		FreshDashboardEvidenceCount: freshCount,
		StaleDashboardEvidence:      staleItems,
		OldExportRootCount:          len(oldRoots),
		OldExportRoots:              oldRoots,
		RecommendedAction:           "regenerate_dashboard_evidence_from_latest_readback_before_operator_handoff",
		PromotionRequested:          false,
		PromotionGranted:            false,
		ClaimsAuthorityAdvance:      false,
		SchedulesWork:               false,
		ExecutesWork:                false,
		ApprovesWork:                false,
		RSIRemainsDenied:            true,
	}, nil
}

func buildMissionDashboardStaleEvidenceItem(path, latestDigest string) (AtlasMissionDashboardStaleEvidenceItem, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return AtlasMissionDashboardStaleEvidenceItem{}, fmt.Errorf("dashboard evidence path is required")
	}
	raw, err := LoadJSON[map[string]any](path)
	if err != nil {
		return AtlasMissionDashboardStaleEvidenceItem{}, err
	}
	schema, _ := raw["schema"].(string)
	sourceReadbackPath, _ := raw["source_readback_path"].(string)
	sourceReadbackDigest, _ := raw["source_readback_digest"].(string)
	stale := sourceReadbackDigest != "" && sourceReadbackDigest != latestDigest
	reason := ""
	if stale {
		reason = "source_readback_digest_superseded_by_latest_readback"
	}
	return AtlasMissionDashboardStaleEvidenceItem{
		Path:                 publicArtifactRef(path),
		Schema:               schema,
		SourceReadbackPath:   sourceReadbackPath,
		SourceReadbackDigest: sourceReadbackDigest,
		LatestReadbackDigest: latestDigest,
		Stale:                stale,
		StaleReason:          reason,
	}, nil
}
