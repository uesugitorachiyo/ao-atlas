package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3ArchitectureSourceTruthChecklist(nodeID, sourceReadbackPath string, readback AtlasRecommendationReadback) (AtlasMonth3ArchitectureSourceTruthChecklist, error) {
	checklist := AtlasMonth3ArchitectureSourceTruthChecklist{
		Schema:               AtlasMonth3ArchitectureSourceTruthChecklistContract,
		NodeID:               strings.TrimSpace(nodeID),
		Status:               "checklist_ready",
		SourceReadbackPath:   publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest: digestValue(readback),
		MissionID:            readback.MissionID,
		CompletedNodes:       readback.CompletedNodes,
		ReadyNodes:           readback.ReadyNodes,
		CurrentAuthorityStatement: "highest_proven_live_class=complex_repo_mutation; " +
			"fully_unsupervised_complex_mutation=denied; RSI=denied",
		TargetArchitectureFiles: []string{
			"ao-architecture/README.md",
			"ao-architecture/overview/README.md",
			"ao-architecture/overview/PRODUCTION-READINESS.md",
			"ao-architecture/docs/superpowers/specs/",
		},
		Checklist: []AtlasMonth3ArchitectureSourceTruthItem{
			{
				Area:           "authority_ladder",
				Status:         "correction_required",
				RequiredAction: "Align Architecture authority statements to keep complex_repo_mutation as highest proven live class.",
				SourceOfTruth:  "Month 3 terminal readback and Promoter no-promotion evidence",
			},
			{
				Area:           "fully_unsupervised_complex_mutation",
				Status:         "correction_required",
				RequiredAction: "State fully_unsupervised_complex_mutation remains denied until final promotion evidence exists.",
				SourceOfTruth:  "AO Mission and Atlas final-response denial gates",
			},
			{
				Area:           "readiness_inventory",
				Status:         "correction_required",
				RequiredAction: "Include Mission and Blueprint in readiness inventory rather than documenting only partial stack coverage.",
				SourceOfTruth:  "AO Architecture Month 1 roadmap disposition",
			},
			{
				Area:           "evidence_catalog",
				Status:         "correction_required",
				RequiredAction: "Move historical campaign evidence into generated catalog links and keep current behavior as the README source of truth.",
				SourceOfTruth:  "AO stack six-month consolidation roadmap",
			},
		},
		CorrectionsRequired:    true,
		PromotionRequested:     false,
		PromotionGranted:       false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasMonth3ArchitectureSourceTruthChecklist(checklist); err != nil {
		return AtlasMonth3ArchitectureSourceTruthChecklist{}, err
	}
	return checklist, nil
}

func ValidateAtlasMonth3ArchitectureSourceTruthChecklist(checklist AtlasMonth3ArchitectureSourceTruthChecklist) error {
	var errs []string
	requireContract(&errs, "month3_architecture_source_truth_checklist", checklist.Schema, AtlasMonth3ArchitectureSourceTruthChecklistContract)
	for field, value := range map[string]string{
		"node_id":                     checklist.NodeID,
		"status":                      checklist.Status,
		"source_readback_path":        checklist.SourceReadbackPath,
		"source_readback_digest":      checklist.SourceReadbackDigest,
		"mission_id":                  checklist.MissionID,
		"current_authority_statement": checklist.CurrentAuthorityStatement,
	} {
		requireField(&errs, field, value)
	}
	checkPublicPath(&errs, "node_id", checklist.NodeID, true)
	checkPublicPath(&errs, "source_readback_path", checklist.SourceReadbackPath, true)
	checkOptionalDigest(&errs, "source_readback_digest", checklist.SourceReadbackDigest)
	if checklist.Status != "checklist_ready" {
		errs = append(errs, "status must be checklist_ready")
	}
	if checklist.CompletedNodes <= 0 || checklist.ReadyNodes <= 0 {
		errs = append(errs, "completed_nodes and ready_nodes must be positive")
	}
	if !strings.Contains(checklist.CurrentAuthorityStatement, "complex_repo_mutation") ||
		!strings.Contains(checklist.CurrentAuthorityStatement, "fully_unsupervised_complex_mutation=denied") ||
		!strings.Contains(checklist.CurrentAuthorityStatement, "RSI=denied") {
		errs = append(errs, "current_authority_statement must preserve denied authority boundaries")
	}
	if len(checklist.TargetArchitectureFiles) == 0 {
		errs = append(errs, "target_architecture_files must not be empty")
	}
	checkPublicStrings(&errs, "target_architecture_files", checklist.TargetArchitectureFiles, true)
	if len(checklist.Checklist) != 4 {
		errs = append(errs, "checklist must contain exactly four correction items")
	}
	for i, item := range checklist.Checklist {
		prefix := fmt.Sprintf("checklist[%d]", i)
		for field, value := range map[string]string{
			prefix + ".area":            item.Area,
			prefix + ".status":          item.Status,
			prefix + ".required_action": item.RequiredAction,
			prefix + ".source_of_truth": item.SourceOfTruth,
		} {
			requireField(&errs, field, value)
		}
		if item.Status != "correction_required" {
			errs = append(errs, prefix+".status must be correction_required")
		}
	}
	if !checklist.CorrectionsRequired {
		errs = append(errs, "corrections_required must be true")
	}
	if checklist.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if checklist.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, false, false, false, checklist.ClaimsAuthorityAdvance, checklist.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3ArchitectureSourceTruthChecklist(path string, checklist AtlasMonth3ArchitectureSourceTruthChecklist) error {
	return WriteJSON(path, checklist)
}
