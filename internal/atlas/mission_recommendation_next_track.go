package atlas

import (
	"fmt"
	"path/filepath"
	"strings"
)

func BuildAtlasRecommendationNextTrackDecision(sourceEvidenceRoot, sourceReadbackPath string) (AtlasRecommendationNextTrackDecision, error) {
	sourceEvidenceRoot = filepath.ToSlash(strings.TrimSpace(sourceEvidenceRoot))
	sourceReadbackPath = filepath.ToSlash(strings.TrimSpace(sourceReadbackPath))
	if sourceEvidenceRoot == "" {
		return AtlasRecommendationNextTrackDecision{}, fmt.Errorf("source evidence root is required")
	}
	if sourceReadbackPath == "" {
		return AtlasRecommendationNextTrackDecision{}, fmt.Errorf("source readback path is required")
	}
	if err := validateNextWaveSourcePath("source_evidence_root", sourceEvidenceRoot); err != nil {
		return AtlasRecommendationNextTrackDecision{}, err
	}
	if err := validateNextWaveSourcePath("source_readback_path", sourceReadbackPath); err != nil {
		return AtlasRecommendationNextTrackDecision{}, err
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasRecommendationNextTrackDecision{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasRecommendationNextTrackDecision{}, err
	}
	readbackDigest, err := digestFile(sourceReadbackPath)
	if err != nil {
		return AtlasRecommendationNextTrackDecision{}, err
	}

	currentTrack := recommendationCurrentTrack(sourceEvidenceRoot, readback)
	currentTrackStatus := "in_progress"
	recommendedTrack := "feature_depth"
	featureDepthStatus := "available"
	refactoringStatus := "pending_after_feature_depth"
	exactNextAction := readback.ExactNextAction
	if exactNextAction == "" {
		exactNextAction = "Continue the current AO Atlas recommendation track from the latest readback."
	}
	if currentTrack == "feature_depth" && isSaturatedFeatureDepthReadback(readback) {
		currentTrackStatus = "completed_saturated"
		recommendedTrack = "refactoring"
		featureDepthStatus = "saturated_completed"
		refactoringStatus = "recommended_next"
		exactNextAction = "Start AO Atlas refactoring wave for recommendation routing, consumed-task ledger, final-response gates, and non-self-referential handoffs."
	}

	decision := AtlasRecommendationNextTrackDecision{
		Schema:                       AtlasRecommendationNextTrackDecisionContract,
		Status:                       "routed",
		SourceEvidenceRoot:           sourceEvidenceRoot,
		SourceReadbackPath:           sourceReadbackPath,
		SourceReadbackDigest:         readbackDigest,
		MissionID:                    readback.MissionID,
		TargetInstance:               readback.TargetInstance,
		CompletedNodes:               readback.CompletedNodes,
		TotalNodes:                   readback.TotalNodes,
		ReadyNodes:                   readback.ReadyNodes,
		BlockedNodes:                 readback.BlockedNodes,
		FailedNodes:                  readback.FailedNodes,
		FinalResponseAllowedObserved: readback.FinalResponseAllowed,
		ReturnGateStatus:             readback.ReturnGateStatus,
		CurrentTrack:                 currentTrack,
		CurrentTrackStatus:           currentTrackStatus,
		RecommendedTrack:             recommendedTrack,
		PriorityOrder:                []string{recommendedTrack, "feature_depth", "rsi_boundary_hardening"},
		FeatureDepthStatus:           featureDepthStatus,
		RefactoringStatus:            refactoringStatus,
		RSITrackStatus:               "boundary_hardening_only_denied",
		ExactNextAction:              exactNextAction,
		NoPromotionRequested:         true,
		PromotionGranted:             false,
		ClaimsAuthorityAdvance:       false,
		RSIRemainsDenied:             true,
		SafeToExecute:                false,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		MutatesRepositories:          false,
	}
	if decision.PriorityOrder[1] == decision.PriorityOrder[0] {
		decision.PriorityOrder[1] = "refactoring"
	}
	if err := ValidateAtlasRecommendationNextTrackDecision(decision); err != nil {
		return AtlasRecommendationNextTrackDecision{}, err
	}
	return decision, nil
}

func BuildAtlasConsumedRecommendationLedger(sourceEvidenceRoot, sourceReadbackPath, nextTrackDecisionPath string) (AtlasConsumedRecommendationLedger, error) {
	sourceEvidenceRoot = filepath.ToSlash(strings.TrimSpace(sourceEvidenceRoot))
	sourceReadbackPath = filepath.ToSlash(strings.TrimSpace(sourceReadbackPath))
	nextTrackDecisionPath = filepath.ToSlash(strings.TrimSpace(nextTrackDecisionPath))
	if sourceEvidenceRoot == "" {
		return AtlasConsumedRecommendationLedger{}, fmt.Errorf("source evidence root is required")
	}
	if sourceReadbackPath == "" {
		return AtlasConsumedRecommendationLedger{}, fmt.Errorf("source readback path is required")
	}
	if nextTrackDecisionPath == "" {
		return AtlasConsumedRecommendationLedger{}, fmt.Errorf("next-track decision path is required")
	}
	if err := validateNextWaveSourcePath("source_evidence_root", sourceEvidenceRoot); err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}
	if err := validateNextWaveSourcePath("source_readback_path", sourceReadbackPath); err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}
	decision, err := LoadJSON[AtlasRecommendationNextTrackDecision](nextTrackDecisionPath)
	if err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}
	if err := ValidateAtlasRecommendationNextTrackDecision(decision); err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}
	if decision.SourceEvidenceRoot != sourceEvidenceRoot {
		return AtlasConsumedRecommendationLedger{}, fmt.Errorf("next-track decision source_evidence_root %s does not match %s", decision.SourceEvidenceRoot, sourceEvidenceRoot)
	}
	if filepath.ToSlash(decision.SourceReadbackPath) != sourceReadbackPath {
		return AtlasConsumedRecommendationLedger{}, fmt.Errorf("next-track decision source_readback_path %s does not match %s", decision.SourceReadbackPath, sourceReadbackPath)
	}
	if decision.CurrentTrack != "feature_depth" ||
		decision.CurrentTrackStatus != "completed_saturated" ||
		decision.RecommendedTrack != "refactoring" {
		return AtlasConsumedRecommendationLedger{}, fmt.Errorf("next-track decision does not record completed Feature Depth routing to refactoring")
	}
	if !isSaturatedFeatureDepthReadback(readback) {
		return AtlasConsumedRecommendationLedger{}, fmt.Errorf("source readback is not a completed saturated Feature Depth readback")
	}
	readbackDigest, err := digestFile(sourceReadbackPath)
	if err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}
	decisionDigest, err := digestFile(nextTrackDecisionPath)
	if err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}

	ledger := AtlasConsumedRecommendationLedger{
		Schema:                       AtlasConsumedRecommendationLedgerContract,
		Status:                       "consumed_recorded",
		SourceEvidenceRoot:           sourceEvidenceRoot,
		SourceReadbackPath:           sourceReadbackPath,
		SourceReadbackDigest:         readbackDigest,
		NextTrackDecisionPath:        nextTrackDecisionPath,
		NextTrackDecisionDigest:      decisionDigest,
		MissionID:                    readback.MissionID,
		TargetInstance:               readback.TargetInstance,
		CompletedNodes:               readback.CompletedNodes,
		TotalNodes:                   readback.TotalNodes,
		FinalResponseAllowedObserved: readback.FinalResponseAllowed,
		CurrentTrack:                 decision.CurrentTrack,
		CurrentTrackStatus:           decision.CurrentTrackStatus,
		ConsumedTrack:                decision.CurrentTrack,
		ConsumedReason:               "completed_saturated_feature_depth_routed_to_refactoring",
		RecommendedTrack:             decision.RecommendedTrack,
		ExactNextAction:              decision.ExactNextAction,
		DuplicateExportBlocked:       true,
		ImportBypassBlocked:          true,
		NoPromotionRequested:         true,
		PromotionGranted:             false,
		ClaimsAuthorityAdvance:       false,
		RSIRemainsDenied:             true,
		SafeToExecute:                false,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		MutatesRepositories:          false,
	}
	if err := ValidateAtlasConsumedRecommendationLedger(ledger); err != nil {
		return AtlasConsumedRecommendationLedger{}, err
	}
	return ledger, nil
}

func recommendationCurrentTrack(sourceEvidenceRoot string, readback AtlasRecommendationReadback) string {
	combined := strings.ToLower(sourceEvidenceRoot + " " + readback.MissionID + " " + readback.TargetInstance)
	switch {
	case strings.Contains(combined, "feature-depth"):
		return "feature_depth"
	case strings.Contains(combined, "refactor"):
		return "refactoring"
	case strings.Contains(combined, "rsi"):
		return "rsi_boundary_hardening"
	default:
		return "unknown"
	}
}

func ValidateAtlasConsumedRecommendationLedger(ledger AtlasConsumedRecommendationLedger) error {
	var errs []string
	requireContract(&errs, "consumed_recommendation_ledger", ledger.Schema, AtlasConsumedRecommendationLedgerContract)
	if ledger.Status != "consumed_recorded" {
		errs = append(errs, "status must be consumed_recorded")
	}
	requireField(&errs, "source_evidence_root", ledger.SourceEvidenceRoot)
	requireField(&errs, "source_readback_path", ledger.SourceReadbackPath)
	requireField(&errs, "next_track_decision_path", ledger.NextTrackDecisionPath)
	checkPublicPath(&errs, "source_evidence_root", ledger.SourceEvidenceRoot, true)
	checkPublicPath(&errs, "source_readback_path", ledger.SourceReadbackPath, true)
	if !digestPattern.MatchString(ledger.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if !digestPattern.MatchString(ledger.NextTrackDecisionDigest) {
		errs = append(errs, "next_track_decision_digest must be sha256 digest")
	}
	requireField(&errs, "mission_id", ledger.MissionID)
	requireField(&errs, "target_instance", ledger.TargetInstance)
	if ledger.TotalNodes <= 0 {
		errs = append(errs, "total_nodes must be positive")
	}
	if ledger.CompletedNodes < 0 || ledger.CompletedNodes > ledger.TotalNodes {
		errs = append(errs, "completed_nodes must be between 0 and total_nodes")
	}
	if !ledger.FinalResponseAllowedObserved {
		errs = append(errs, "final_response_allowed_observed must be true")
	}
	if ledger.CurrentTrack != "feature_depth" {
		errs = append(errs, "current_track must be feature_depth")
	}
	if ledger.CurrentTrackStatus != "completed_saturated" {
		errs = append(errs, "current_track_status must be completed_saturated")
	}
	if ledger.ConsumedTrack != ledger.CurrentTrack {
		errs = append(errs, "consumed_track must match current_track")
	}
	if ledger.ConsumedReason != "completed_saturated_feature_depth_routed_to_refactoring" {
		errs = append(errs, "consumed_reason is invalid")
	}
	if ledger.RecommendedTrack != "refactoring" {
		errs = append(errs, "recommended_track must be refactoring")
	}
	requireField(&errs, "exact_next_action", ledger.ExactNextAction)
	checkPublicStrings(&errs, "exact_next_action", []string{ledger.ExactNextAction}, true)
	if !ledger.DuplicateExportBlocked {
		errs = append(errs, "duplicate_export_blocked must be true")
	}
	if !ledger.ImportBypassBlocked {
		errs = append(errs, "import_bypass_blocked must be true")
	}
	if !ledger.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if ledger.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if ledger.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !ledger.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if ledger.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if ledger.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if ledger.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if ledger.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if ledger.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}

func ValidateAtlasRecommendationNextTrackDecision(decision AtlasRecommendationNextTrackDecision) error {
	var errs []string
	requireContract(&errs, "recommendation_next_track_decision", decision.Schema, AtlasRecommendationNextTrackDecisionContract)
	if !oneOf(decision.Status, "routed", "blocked") {
		errs = append(errs, "status must be routed or blocked")
	}
	requireField(&errs, "source_evidence_root", decision.SourceEvidenceRoot)
	requireField(&errs, "source_readback_path", decision.SourceReadbackPath)
	checkPublicPath(&errs, "source_evidence_root", decision.SourceEvidenceRoot, true)
	checkPublicPath(&errs, "source_readback_path", decision.SourceReadbackPath, true)
	if !digestPattern.MatchString(decision.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	requireField(&errs, "mission_id", decision.MissionID)
	requireField(&errs, "target_instance", decision.TargetInstance)
	if decision.TotalNodes <= 0 {
		errs = append(errs, "total_nodes must be positive")
	}
	if decision.CompletedNodes < 0 || decision.CompletedNodes > decision.TotalNodes {
		errs = append(errs, "completed_nodes must be between 0 and total_nodes")
	}
	if decision.ReadyNodes < 0 || decision.BlockedNodes < 0 || decision.FailedNodes < 0 {
		errs = append(errs, "node counters must not be negative")
	}
	requireField(&errs, "return_gate_status", decision.ReturnGateStatus)
	if !oneOf(decision.CurrentTrack, "feature_depth", "refactoring", "rsi_boundary_hardening", "unknown") {
		errs = append(errs, "current_track is invalid")
	}
	if !oneOf(decision.CurrentTrackStatus, "in_progress", "completed_saturated") {
		errs = append(errs, "current_track_status is invalid")
	}
	if !oneOf(decision.RecommendedTrack, "feature_depth", "refactoring", "rsi_boundary_hardening") {
		errs = append(errs, "recommended_track is invalid")
	}
	requireList(&errs, "priority_order", decision.PriorityOrder)
	if len(decision.PriorityOrder) < 3 {
		errs = append(errs, "priority_order must include at least 3 tracks")
	}
	checkPublicStrings(&errs, "priority_order", decision.PriorityOrder, true)
	requireField(&errs, "feature_depth_status", decision.FeatureDepthStatus)
	requireField(&errs, "refactoring_status", decision.RefactoringStatus)
	if decision.RSITrackStatus != "boundary_hardening_only_denied" {
		errs = append(errs, "rsi_track_status must be boundary_hardening_only_denied")
	}
	requireField(&errs, "exact_next_action", decision.ExactNextAction)
	checkPublicStrings(&errs, "exact_next_action", []string{decision.ExactNextAction}, true)
	if !decision.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if decision.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if decision.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !decision.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if decision.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if decision.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if decision.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if decision.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if decision.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}
