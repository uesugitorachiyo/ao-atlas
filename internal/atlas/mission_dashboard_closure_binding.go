package atlas

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type missionDashboardClosureRowSpec struct {
	repo            string
	role            string
	path            string
	readinessStatus string
}

func BuildAtlasMissionDashboardClosureBinding(nodeID, sourceReadbackPath, sourceNodeDir string) (AtlasMissionDashboardClosureBinding, error) {
	nodeID = strings.TrimSpace(nodeID)
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	sourceNodeDir = strings.TrimSpace(sourceNodeDir)
	for name, value := range map[string]string{
		"node id":              nodeID,
		"source readback path": sourceReadbackPath,
		"source node dir":      sourceNodeDir,
	} {
		if value == "" {
			return AtlasMissionDashboardClosureBinding{}, fmt.Errorf("%s is required", name)
		}
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMissionDashboardClosureBinding{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMissionDashboardClosureBinding{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMissionDashboardClosureBinding{}, err
	}

	specs := []missionDashboardClosureRowSpec{
		{
			repo:            "ao-atlas",
			role:            "workgraph_state_owner",
			path:            sourceReadbackPath,
			readinessStatus: "ready_work_remains",
		},
		{
			repo:            "ao-foundry",
			role:            "bounded_implementation_handoff",
			path:            filepath.Join(sourceNodeDir, "foundry-import.json"),
			readinessStatus: "single_active_import_closed",
		},
		{
			repo:            "ao-command",
			role:            "compact_readback_owner",
			path:            filepath.Join(sourceNodeDir, "command_readback.json"),
			readinessStatus: "readback_agrees",
		},
		{
			repo:            "ao-promoter",
			role:            "no_promotion_rollup_owner",
			path:            filepath.Join(sourceNodeDir, "promoter_no_promotion.json"),
			readinessStatus: "no_promotion_requested",
		},
		{
			repo:            "ao-sentinel",
			role:            "public_safety_scan_owner",
			path:            filepath.Join(sourceNodeDir, "sentinel_public_safety.json"),
			readinessStatus: "public_safety_passed",
		},
		{
			repo:            "ao-mission",
			role:            "checkpoint_readback_owner",
			path:            filepath.Join(sourceNodeDir, "checkpoint-readback-after.json"),
			readinessStatus: "checkpoint_readback_bound",
		},
	}

	rows := make([]AtlasMissionDashboardClosureBindingRow, 0, len(specs))
	for _, spec := range specs {
		digest, err := digestTextFileWithNormalizedLineEndings(spec.path)
		if err != nil {
			return AtlasMissionDashboardClosureBinding{}, err
		}
		rows = append(rows, AtlasMissionDashboardClosureBindingRow{
			Repo:                    spec.repo,
			Role:                    spec.role,
			ClosureEvidencePath:     publicArtifactRef(spec.path),
			ClosureEvidenceDigest:   digest,
			ReadinessStatus:         spec.readinessStatus,
			EvidenceStatus:          "bound",
			ProvenanceRequired:      true,
			FinalResponseAllowed:    readback.FinalResponseAllowed,
			RSIRemainsDenied:        readback.SafetyBoundaries["rsi_remains_denied"],
			AuthorityAdvanceClaimed: false,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Repo < rows[j].Repo
	})

	binding := AtlasMissionDashboardClosureBinding{
		Schema:                     AtlasMissionDashboardClosureBindingContract,
		NodeID:                     nodeID,
		Status:                     "mission_dashboard_closure_bound",
		SourceReadbackPath:         publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:       readbackDigest,
		SourceNodeDir:              publicArtifactRef(sourceNodeDir),
		CompletedNodesBefore:       readback.CompletedNodes,
		ReadyNodesBefore:           readback.ReadyNodes,
		BlockedNodesBefore:         readback.BlockedNodes,
		FailedNodesBefore:          readback.FailedNodes,
		FirstExecutableNode:        readback.FirstExecutableNode,
		ExactNextAction:            readback.ExactNextAction,
		SchemaHealthStatus:         readback.SchemaHealthStatus,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		RowCount:                   len(rows),
		Rows:                       rows,
		AtlasClosureEvidenceBound:  missionDashboardAtlasRowBound(rows, publicArtifactRef(sourceReadbackPath), readbackDigest),
		EveryRowHasClosureEvidence: missionDashboardRowsHaveClosureEvidence(rows),
		EveryRowPreservesSafety:    missionDashboardRowsPreserveSafety(rows),
		DashboardBindingStatus:     "multi_repo_rows_bound_to_closure_evidence",
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
		RSIRemainsDenied:           readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if err := ValidateAtlasMissionDashboardClosureBinding(binding); err != nil {
		return AtlasMissionDashboardClosureBinding{}, err
	}
	return binding, nil
}

func ValidateAtlasMissionDashboardClosureBinding(binding AtlasMissionDashboardClosureBinding) error {
	var errs []string
	requireContract(&errs, "mission_dashboard_closure_binding", binding.Schema, AtlasMissionDashboardClosureBindingContract)
	requireField(&errs, "node_id", binding.NodeID)
	checkPublicPath(&errs, "node_id", binding.NodeID, true)
	if binding.Status != "mission_dashboard_closure_bound" {
		errs = append(errs, "status must be mission_dashboard_closure_bound")
	}
	for field, value := range map[string]string{
		"source_readback_path":     binding.SourceReadbackPath,
		"source_node_dir":          binding.SourceNodeDir,
		"first_executable_node":    binding.FirstExecutableNode,
		"exact_next_action":        binding.ExactNextAction,
		"dashboard_binding_status": binding.DashboardBindingStatus,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	checkPublicPath(&errs, "schema_health_status", binding.SchemaHealthStatus, true)
	if !digestPattern.MatchString(binding.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if binding.CompletedNodesBefore <= 0 || binding.ReadyNodesBefore <= 0 || binding.BlockedNodesBefore < 0 || binding.FailedNodesBefore < 0 {
		errs = append(errs, "node counts must prove completed and ready work with no negative counts")
	}
	if binding.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready dashboard work remains")
	}
	if binding.RowCount != len(binding.Rows) || binding.RowCount != 6 {
		errs = append(errs, "row_count must match the six required Mission dashboard rows")
	}
	if !binding.AtlasClosureEvidenceBound {
		errs = append(errs, "atlas_closure_evidence_bound must be true")
	}
	if !binding.EveryRowHasClosureEvidence {
		errs = append(errs, "every_row_has_closure_evidence must be true")
	}
	if !binding.EveryRowPreservesSafety {
		errs = append(errs, "every_row_preserves_safety must be true")
	}
	if binding.DashboardBindingStatus != "multi_repo_rows_bound_to_closure_evidence" {
		errs = append(errs, "dashboard_binding_status must be multi_repo_rows_bound_to_closure_evidence")
	}
	validateMissionDashboardClosureRows(&errs, binding.Rows)
	validateNoAuthorityEffects(&errs, binding.SchedulesWork, binding.ExecutesWork, binding.ApprovesWork, binding.ClaimsAuthorityAdvance, binding.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateMissionDashboardClosureRows(errs *[]string, rows []AtlasMissionDashboardClosureBindingRow) {
	expectedRoles := map[string]string{
		"ao-atlas":    "workgraph_state_owner",
		"ao-command":  "compact_readback_owner",
		"ao-foundry":  "bounded_implementation_handoff",
		"ao-mission":  "checkpoint_readback_owner",
		"ao-promoter": "no_promotion_rollup_owner",
		"ao-sentinel": "public_safety_scan_owner",
	}
	seen := map[string]bool{}
	previousRepo := ""
	for i, row := range rows {
		prefix := fmt.Sprintf("rows[%d]", i)
		requireField(errs, prefix+".repo", row.Repo)
		requireField(errs, prefix+".role", row.Role)
		requireField(errs, prefix+".closure_evidence_path", row.ClosureEvidencePath)
		requireField(errs, prefix+".readiness_status", row.ReadinessStatus)
		requireField(errs, prefix+".evidence_status", row.EvidenceStatus)
		checkPublicPath(errs, prefix+".repo", row.Repo, true)
		checkPublicPath(errs, prefix+".role", row.Role, true)
		checkPublicPath(errs, prefix+".closure_evidence_path", row.ClosureEvidencePath, true)
		checkPublicPath(errs, prefix+".readiness_status", row.ReadinessStatus, true)
		if !digestPattern.MatchString(row.ClosureEvidenceDigest) {
			*errs = append(*errs, prefix+".closure_evidence_digest must be sha256 digest")
		}
		if seen[row.Repo] {
			*errs = append(*errs, "rows repo values must be unique")
		}
		seen[row.Repo] = true
		if expectedRoles[row.Repo] != row.Role {
			*errs = append(*errs, prefix+".role must match required repo role")
		}
		if previousRepo != "" && row.Repo < previousRepo {
			*errs = append(*errs, "rows must be sorted by repo")
		}
		previousRepo = row.Repo
		if row.EvidenceStatus != "bound" {
			*errs = append(*errs, prefix+".evidence_status must be bound")
		}
		if !row.ProvenanceRequired {
			*errs = append(*errs, prefix+".provenance_required must be true")
		}
		if row.FinalResponseAllowed {
			*errs = append(*errs, prefix+".final_response_allowed must be false")
		}
		if !row.RSIRemainsDenied {
			*errs = append(*errs, prefix+".rsi_remains_denied must be true")
		}
		if row.AuthorityAdvanceClaimed {
			*errs = append(*errs, prefix+".authority_advance_claimed must be false")
		}
	}
	for repo := range expectedRoles {
		if !seen[repo] {
			*errs = append(*errs, "rows missing "+repo)
		}
	}
}

func missionDashboardAtlasRowBound(rows []AtlasMissionDashboardClosureBindingRow, sourceReadbackPath, sourceReadbackDigest string) bool {
	for _, row := range rows {
		if row.Repo == "ao-atlas" {
			return row.ClosureEvidencePath == sourceReadbackPath && row.ClosureEvidenceDigest == sourceReadbackDigest && row.EvidenceStatus == "bound"
		}
	}
	return false
}

func missionDashboardRowsHaveClosureEvidence(rows []AtlasMissionDashboardClosureBindingRow) bool {
	for _, row := range rows {
		if strings.TrimSpace(row.ClosureEvidencePath) == "" || !digestPattern.MatchString(row.ClosureEvidenceDigest) || row.EvidenceStatus != "bound" {
			return false
		}
	}
	return len(rows) > 0
}

func missionDashboardRowsPreserveSafety(rows []AtlasMissionDashboardClosureBindingRow) bool {
	for _, row := range rows {
		if row.FinalResponseAllowed || !row.RSIRemainsDenied || row.AuthorityAdvanceClaimed || !row.ProvenanceRequired {
			return false
		}
	}
	return len(rows) > 0
}

func WriteAtlasMissionDashboardClosureBinding(path string, binding AtlasMissionDashboardClosureBinding) error {
	return WriteJSON(path, binding)
}
