package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AtlasRecommendationWaveOptions struct {
	RecommendationsPath  string
	TargetInstance       string
	MinTasks             int
	NodeBudget           int
	EstimatedMinutes     int
	MinMinutes           int
	MinMinutesSet        bool
	MaxMinutes           int
	ContinueIfFastTarget int
	ReturnOnlyWhen       string
	CheckpointPolicy     string
	EvidencePolicy       string
	FinalReportContract  string
}

type AtlasNextWaveFeatureDepthExportOptions struct {
	MissionID           string
	SourceEvidenceRoot  string
	SourceReadbackPath  string
	SourceAssertionPath string
	MinTasks            int
}

type AtlasNextWaveRefactoringExportOptions struct {
	MissionID                  string
	SourceEvidenceRoot         string
	SourceReadbackPath         string
	SourceAssertionPath        string
	NextTrackDecisionPath      string
	ConsumedRecommendationPath string
	MinTasks                   int
}

type AtlasRecommendationWaveResult struct {
	Wave      AtlasRecommendationWave
	Workgraph Workgraph
	Prompt    string
}

type AtlasRecommendationReadbackOptions struct {
	WavePath               string
	WorkgraphPath          string
	EvidenceRoot           string
	StartedAt              string
	CompletedAt            string
	ElapsedMinutes         int
	LeaseTimingMode        string
	PublicSafetyScanStatus string
	SchemaHealthStatus     string
}

type AtlasRecommendationLeaseStartOptions struct {
	WavePath      string
	WorkgraphPath string
	EvidenceRoot  string
	StartedAt     string
}

type AtlasRecommendationWorkgraphReadinessPacketOptions struct {
	WavePath      string
	WorkgraphPath string
	ReadbackPath  string
}

type AtlasRecommendationCompleteNodeOptions struct {
	ExpectedNodeID string
	EvidenceRoot   string
}

func BuildAtlasRecommendationWave(options AtlasRecommendationWaveOptions) (AtlasRecommendationWaveResult, error) {
	minTasks := options.MinTasks
	if minTasks <= 0 {
		minTasks = 30
	}
	nodeBudget := options.NodeBudget
	if nodeBudget <= 0 {
		nodeBudget = 40
	}
	continueIfFastTarget := options.ContinueIfFastTarget
	if continueIfFastTarget <= 0 {
		continueIfFastTarget = nodeBudget
	}
	minMinutes := options.MinMinutes
	if minMinutes < 0 {
		return AtlasRecommendationWaveResult{}, fmt.Errorf("min_minutes must be zero or greater")
	}
	if !options.MinMinutesSet && minMinutes <= 0 {
		if minTasks >= 30 || nodeBudget >= 40 || continueIfFastTarget >= 40 {
			minMinutes = 120
		} else {
			minMinutes = 90
		}
	}
	maxMinutes := options.MaxMinutes
	if maxMinutes <= 0 {
		if minMinutes >= 120 || minTasks >= 30 || nodeBudget >= 40 || continueIfFastTarget >= 40 {
			maxMinutes = 180
		} else if minMinutes > 0 {
			maxMinutes = minMinutes
		} else {
			maxMinutes = 90
		}
	}
	estimatedMinutes := options.EstimatedMinutes
	if estimatedMinutes <= 0 {
		if minTasks >= 30 || nodeBudget >= 40 || continueIfFastTarget >= 40 {
			estimatedMinutes = 120
		} else if minMinutes > 0 {
			estimatedMinutes = minMinutes
		} else {
			estimatedMinutes = 90
		}
	}
	returnOnlyWhen := strings.TrimSpace(options.ReturnOnlyWhen)
	if returnOnlyWhen == "" {
		returnOnlyWhen = fmt.Sprintf("all_generated_nodes_done_or_%d_nodes_done_or_true_hard_blocker", minTasks)
	}
	checkpointPolicy := strings.TrimSpace(options.CheckpointPolicy)
	if checkpointPolicy == "" {
		checkpointPolicy = "after_each_node_or_timed_interval"
	}
	evidencePolicy := strings.TrimSpace(options.EvidencePolicy)
	if evidencePolicy == "" {
		evidencePolicy = "node_gate_candidate_rollback_tests_verification_public_safety_promoter_command"
	}
	finalReportContract := strings.TrimSpace(options.FinalReportContract)
	if finalReportContract == "" {
		finalReportContract = "ao.atlas.long-run-final-report.v0.2"
	}
	if strings.TrimSpace(options.TargetInstance) == "" {
		return AtlasRecommendationWaveResult{}, fmt.Errorf("target_instance is required")
	}
	var bundle AOMissionFeatureDepthRecommendations
	if err := readJSONIfPossible(options.RecommendationsPath, &bundle); err != nil {
		return AtlasRecommendationWaveResult{}, err
	}
	if err := ValidateAOMissionFeatureDepthRecommendations(bundle, minTasks); err != nil {
		return AtlasRecommendationWaveResult{}, err
	}
	if err := rejectSaturatedFeatureDepthContinuation(bundle.SourceEvidenceRoot, bundle.SourceReadbackPath); err != nil {
		return AtlasRecommendationWaveResult{}, err
	}
	sourceDigest, err := digestFile(options.RecommendationsPath)
	if err != nil {
		return AtlasRecommendationWaveResult{}, err
	}
	selected := atlasOwnedRecommendationTasks(bundle.Tasks, nodeBudget)
	if len(selected) < minTasks || len(selected) < continueIfFastTarget {
		if continueIfFastTarget > minTasks {
			return AtlasRecommendationWaveResult{}, fmt.Errorf("AO Atlas recommendation wave requires at least %d AO Atlas-owned tasks and %d tasks for continue-if-fast target, got %d", minTasks, continueIfFastTarget, len(selected))
		}
		return AtlasRecommendationWaveResult{}, fmt.Errorf("AO Atlas recommendation wave requires at least %d tasks, got %d", minTasks, len(selected))
	}
	tasks := make([]AtlasRecommendationTask, 0, len(selected))
	for _, item := range selected {
		nodeID := "mission-recommendation-" + sanitizeMissionProvenanceNodeName(item.ID)
		if nodeID == "mission-recommendation-" {
			nodeID = "mission-recommendation-" + sanitizeMissionProvenanceNodeName(item.Task)
		}
		tasks = append(tasks, AtlasRecommendationTask{
			ID:                item.ID,
			Owner:             item.Owner,
			Task:              item.Task,
			NodeID:            nodeID,
			TaskID:            nodeID + "-task",
			MutationClass:     "low_risk_code",
			SourceTaskDigest:  digestValue(item),
			TargetFactoryRepo: "ao-atlas",
			FactoryFolder:     "factory/ao-atlas-recommendations/" + strings.TrimPrefix(nodeID, "mission-recommendation-"),
			RequiredGates: []string{
				"node_gate",
				"candidate_record",
				"rollback_record",
				"tests",
				"verification",
				"sentinel_public_safety",
				"promoter_no_promotion",
				"command_readback",
			},
			Verification: []string{
				"go test ./... -count=1",
				"go vet ./...",
				"go build ./cmd/atlas",
				"scripts/production-readiness.sh",
				"scripts/atlas-foundry-roundtrip-smoke.sh",
			},
			SafetyLimits: []string{
				"no provider calls",
				"no credential inspection",
				"no direct main mutation",
				"no release deploy publish upload tag",
				"no dependency updates without separate authorization",
				"no auth policy config widening",
				"no broad RSI claim",
			},
		})
	}
	wave := AtlasRecommendationWave{
		ContractVersion:  AtlasRecommendationWaveContract,
		MissionID:        bundle.MissionID,
		TargetInstance:   options.TargetInstance,
		Status:           "ready",
		SourceDigest:     sourceDigest,
		MinimumTasks:     minTasks,
		TotalTasks:       len(tasks),
		NodeBudget:       nodeBudget,
		EstimatedMinutes: estimatedMinutes,
		Supervisor: &AtlasLongRunSupervisor{
			ContractVersion:      "ao.atlas.long-run-supervisor.v0.2",
			MinNodes:             minTasks,
			MinMinutes:           minMinutes,
			MaxMinutes:           maxMinutes,
			ContinueIfFastTarget: continueIfFastTarget,
			ReturnOnlyWhen:       returnOnlyWhen,
			CheckpointPolicy:     checkpointPolicy,
			EvidencePolicy:       evidencePolicy,
			FinalReportContract:  finalReportContract,
		},
		Tasks:                  tasks,
		FinalResponseAllowed:   false,
		FinalResponseReason:    "ready nodes or exact next actions remain",
		PromoterReadbackStatus: "required_not_bound",
		CommandReadbackStatus:  "required_not_bound",
		PublicSafetyScanStatus: "required_pending_verification",
		SafeToExecute:          false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
	}
	prompt := buildAtlasRecommendationPrompt(wave)
	wave.NextRecommendedPrompt = prompt
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		return AtlasRecommendationWaveResult{}, err
	}
	workgraph, err := BuildAtlasRecommendationWorkgraph(wave)
	if err != nil {
		return AtlasRecommendationWaveResult{}, err
	}
	return AtlasRecommendationWaveResult{Wave: wave, Workgraph: workgraph, Prompt: prompt}, nil
}

func ValidateAOMissionFeatureDepthRecommendations(bundle AOMissionFeatureDepthRecommendations, minTasks int) error {
	var errs []string
	if bundle.Schema != "ao.mission.feature-depth-recommendations.v0.3" {
		errs = append(errs, "schema must be ao.mission.feature-depth-recommendations.v0.3")
	}
	requireField(&errs, "mission_id", bundle.MissionID)
	if bundle.Status != "ready" {
		errs = append(errs, "status must be ready")
	}
	if bundle.MinimumTasks < minTasks {
		errs = append(errs, fmt.Sprintf("minimum_tasks must be at least %d", minTasks))
	}
	if len(bundle.Tasks) < minTasks {
		errs = append(errs, fmt.Sprintf("tasks must include at least %d tasks", minTasks))
	}
	if bundle.RecommendationCount != 0 && bundle.RecommendationCount != len(bundle.Tasks) {
		errs = append(errs, "recommendation_count must match tasks length")
	}
	for i, task := range bundle.Tasks {
		prefix := fmt.Sprintf("tasks[%d]", i)
		requireField(&errs, prefix+".id", task.ID)
		requireField(&errs, prefix+".owner", task.Owner)
		requireField(&errs, prefix+".task", task.Task)
		if len(strings.Fields(task.Task)) < 6 {
			errs = append(errs, prefix+".task must be a concrete actionable task")
		}
		checkPublicPath(&errs, prefix+".id", task.ID, true)
		checkPublicPath(&errs, prefix+".owner", task.Owner, true)
		checkPublicPath(&errs, prefix+".task", task.Task, true)
	}
	if bundle.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if bundle.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if bundle.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if bundle.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if bundle.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}

func BuildAtlasNextWaveFeatureDepthRecommendations(options AtlasNextWaveFeatureDepthExportOptions) (AOMissionFeatureDepthRecommendations, error) {
	minTasks := options.MinTasks
	if minTasks <= 0 {
		minTasks = 40
	}
	if minTasks > 40 {
		return AOMissionFeatureDepthRecommendations{}, fmt.Errorf("next-wave exporter currently supports at most 40 ranked tasks, requested %d", minTasks)
	}
	missionID := strings.TrimSpace(options.MissionID)
	if missionID == "" {
		missionID = "ao-atlas-next-feature-depth-wave-v01"
	}
	sourceEvidenceRoot := filepath.ToSlash(strings.TrimSpace(options.SourceEvidenceRoot))
	sourceReadbackPath := filepath.ToSlash(strings.TrimSpace(options.SourceReadbackPath))
	sourceAssertionPath := filepath.ToSlash(strings.TrimSpace(options.SourceAssertionPath))
	for label, path := range map[string]string{
		"source_evidence_root":  sourceEvidenceRoot,
		"source_readback_path":  sourceReadbackPath,
		"source_assertion_path": sourceAssertionPath,
	} {
		if err := validateNextWaveSourcePath(label, path); err != nil {
			return AOMissionFeatureDepthRecommendations{}, err
		}
	}
	if err := rejectSaturatedFeatureDepthContinuation(sourceEvidenceRoot, sourceReadbackPath); err != nil {
		return AOMissionFeatureDepthRecommendations{}, err
	}

	tasks := make([]AOMissionFeatureDepthTask, 0, 40)
	for i, seed := range nextWaveFeatureDepthSeeds() {
		rank := i + 1
		tasks = append(tasks, AOMissionFeatureDepthTask{
			Rank:         rank,
			ID:           fmt.Sprintf("feature-depth-next-wave-%02d", rank),
			Owner:        "ao-atlas",
			Theme:        seed.theme,
			Task:         seed.task,
			EvidenceRefs: []string{sourceReadbackPath, sourceAssertionPath},
		})
	}
	bundle := AOMissionFeatureDepthRecommendations{
		Schema:              "ao.mission.feature-depth-recommendations.v0.3",
		MissionID:           missionID,
		Status:              "ready",
		MinimumTasks:        minTasks,
		RecommendationCount: len(tasks),
		SourceEvidenceRoot:  sourceEvidenceRoot,
		SourceReadbackPath:  sourceReadbackPath,
		SourceAssertionPath: sourceAssertionPath,
		Tasks:               tasks,
		SafeToExecute:       false,
		SchedulesWork:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
	}
	if err := ValidateAtlasNextWaveFeatureDepthRecommendations(bundle, minTasks); err != nil {
		return AOMissionFeatureDepthRecommendations{}, err
	}
	return bundle, nil
}

func BuildAtlasNextWaveRefactoringRecommendations(options AtlasNextWaveRefactoringExportOptions) (AOMissionRefactoringRecommendations, error) {
	minTasks := options.MinTasks
	if minTasks <= 0 {
		minTasks = 40
	}
	if minTasks > 40 {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("refactoring exporter currently supports at most 40 ranked tasks, requested %d", minTasks)
	}
	missionID := strings.TrimSpace(options.MissionID)
	if missionID == "" {
		missionID = "ao-atlas-refactoring-wave-v01"
	}
	sourceEvidenceRoot := filepath.ToSlash(strings.TrimSpace(options.SourceEvidenceRoot))
	sourceReadbackPath := filepath.ToSlash(strings.TrimSpace(options.SourceReadbackPath))
	sourceAssertionPath := filepath.ToSlash(strings.TrimSpace(options.SourceAssertionPath))
	nextTrackDecisionPath := filepath.ToSlash(strings.TrimSpace(options.NextTrackDecisionPath))
	consumedRecommendationPath := filepath.ToSlash(strings.TrimSpace(options.ConsumedRecommendationPath))
	for label, path := range map[string]string{
		"source_evidence_root":  sourceEvidenceRoot,
		"source_readback_path":  sourceReadbackPath,
		"source_assertion_path": sourceAssertionPath,
	} {
		if err := validateNextWaveSourcePath(label, path); err != nil {
			return AOMissionRefactoringRecommendations{}, err
		}
	}
	if nextTrackDecisionPath == "" {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("next_track_decision_path is required")
	}
	if consumedRecommendationPath == "" {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("consumed_recommendation_ledger_path is required")
	}
	decision, err := LoadJSON[AtlasRecommendationNextTrackDecision](nextTrackDecisionPath)
	if err != nil {
		return AOMissionRefactoringRecommendations{}, err
	}
	if err := ValidateAtlasRecommendationNextTrackDecision(decision); err != nil {
		return AOMissionRefactoringRecommendations{}, err
	}
	if decision.SourceEvidenceRoot != sourceEvidenceRoot || filepath.ToSlash(decision.SourceReadbackPath) != sourceReadbackPath {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("next-track decision must match refactoring export source")
	}
	sourceReadbackDigest, err := digestFile(sourceReadbackPath)
	if err != nil {
		return AOMissionRefactoringRecommendations{}, err
	}
	if decision.SourceReadbackDigest != sourceReadbackDigest {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("next-track decision source_readback_digest %s does not match current source readback digest %s", decision.SourceReadbackDigest, sourceReadbackDigest)
	}
	if decision.RecommendedTrack != "refactoring" || decision.RefactoringStatus != "recommended_next" || decision.RSITrackStatus != "boundary_hardening_only_denied" {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("next-track decision must recommend refactoring with RSI boundary denied")
	}
	decisionDigest, err := digestFile(nextTrackDecisionPath)
	if err != nil {
		return AOMissionRefactoringRecommendations{}, err
	}
	ledger, err := LoadJSON[AtlasConsumedRecommendationLedger](consumedRecommendationPath)
	if err != nil {
		return AOMissionRefactoringRecommendations{}, err
	}
	if err := ValidateAtlasConsumedRecommendationLedger(ledger); err != nil {
		return AOMissionRefactoringRecommendations{}, err
	}
	if ledger.SourceEvidenceRoot != sourceEvidenceRoot || filepath.ToSlash(ledger.SourceReadbackPath) != sourceReadbackPath {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("consumed recommendation ledger must match refactoring export source")
	}
	if filepath.ToSlash(ledger.NextTrackDecisionPath) != nextTrackDecisionPath {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("consumed recommendation ledger next_track_decision_path must match refactoring export decision")
	}
	if ledger.SourceReadbackDigest != sourceReadbackDigest {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("consumed recommendation ledger source_readback_digest %s does not match current source readback digest %s", ledger.SourceReadbackDigest, sourceReadbackDigest)
	}
	if ledger.NextTrackDecisionDigest != decisionDigest {
		return AOMissionRefactoringRecommendations{}, fmt.Errorf("consumed recommendation ledger next_track_decision_digest %s does not match current next-track decision digest %s", ledger.NextTrackDecisionDigest, decisionDigest)
	}
	ledgerDigest, err := digestFile(consumedRecommendationPath)
	if err != nil {
		return AOMissionRefactoringRecommendations{}, err
	}
	nextTrackDecisionRef := publicArtifactRef(nextTrackDecisionPath)
	consumedRecommendationRef := publicArtifactRef(consumedRecommendationPath)

	tasks := make([]AOMissionRefactoringTask, 0, 40)
	for i, seed := range nextWaveRefactoringSeeds() {
		rank := i + 1
		tasks = append(tasks, AOMissionRefactoringTask{
			Rank:         rank,
			ID:           fmt.Sprintf("refactoring-next-wave-%02d", rank),
			Owner:        "ao-atlas",
			Theme:        seed.theme,
			Task:         seed.task,
			EvidenceRefs: []string{sourceReadbackPath, sourceAssertionPath, nextTrackDecisionRef},
		})
	}
	bundle := AOMissionRefactoringRecommendations{
		Schema:                  AOMissionRefactoringRecommendationsContract,
		MissionID:               missionID,
		Status:                  "ready",
		Track:                   "refactoring",
		MinimumTasks:            minTasks,
		RecommendationCount:     len(tasks),
		SourceEvidenceRoot:      sourceEvidenceRoot,
		SourceReadbackPath:      sourceReadbackPath,
		SourceReadbackDigest:    sourceReadbackDigest,
		SourceAssertionPath:     sourceAssertionPath,
		NextTrackDecisionPath:   nextTrackDecisionRef,
		NextTrackDecisionDigest: decisionDigest,
		ConsumedLedgerPath:      consumedRecommendationRef,
		ConsumedLedgerDigest:    ledgerDigest,
		Tasks:                   tasks,
		NoPromotionRequested:    true,
		PromotionGranted:        false,
		ClaimsAuthorityAdvance:  false,
		RSIRemainsDenied:        true,
		SafeToExecute:           false,
		SchedulesWork:           false,
		ExecutesWork:            false,
		ApprovesWork:            false,
		MutatesRepositories:     false,
	}
	if err := ValidateAtlasNextWaveRefactoringRecommendations(bundle, minTasks); err != nil {
		return AOMissionRefactoringRecommendations{}, err
	}
	return bundle, nil
}

func BuildAtlasNextWaveRecommendationExport(bundle AOMissionFeatureDepthRecommendations, sourceReadback AtlasRecommendationReadback, nodeID, expectedNextNode string) (AtlasNextWaveRecommendationExport, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasNextWaveRecommendationExport{}, fmt.Errorf("node_id is required")
	}
	expectedNextNode = strings.TrimSpace(expectedNextNode)
	if expectedNextNode == "" {
		return AtlasNextWaveRecommendationExport{}, fmt.Errorf("expected_next_node is required")
	}
	if err := ValidateAtlasNextWaveFeatureDepthRecommendations(bundle, 40); err != nil {
		return AtlasNextWaveRecommendationExport{}, err
	}
	fixture := AtlasNextWaveRecommendationExport{
		Schema:                 "ao.atlas.next-wave-recommendation-export.v0.1",
		NodeID:                 nodeID,
		Status:                 "exported",
		SourceEvidenceRoot:     bundle.SourceEvidenceRoot,
		SourceReadbackPath:     bundle.SourceReadbackPath,
		SourceAssertionPath:    bundle.SourceAssertionPath,
		CompletedNodesBefore:   sourceReadback.CompletedNodes,
		ReadyNodesBefore:       sourceReadback.ReadyNodes,
		ExpectedNextNode:       expectedNextNode,
		MinimumRankedTasks:     40,
		RecommendationCount:    bundle.RecommendationCount,
		RankedTaskFloorMet:     len(bundle.Tasks) >= 40,
		NoPromotionRequested:   true,
		PromotionGranted:       false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
		FeatureDepthExport:     bundle,
	}
	if !fixture.RankedTaskFloorMet {
		return AtlasNextWaveRecommendationExport{}, fmt.Errorf("ranked task floor not met")
	}
	return fixture, nil
}

func ValidateAtlasNextWaveFeatureDepthRecommendations(bundle AOMissionFeatureDepthRecommendations, minTasks int) error {
	if minTasks <= 0 {
		minTasks = 40
	}
	var errs []string
	if err := ValidateAOMissionFeatureDepthRecommendations(bundle, minTasks); err != nil {
		errs = append(errs, err.Error())
	}
	requireField(&errs, "source_evidence_root", bundle.SourceEvidenceRoot)
	requireField(&errs, "source_readback_path", bundle.SourceReadbackPath)
	requireField(&errs, "source_assertion_path", bundle.SourceAssertionPath)
	checkPublicPath(&errs, "source_evidence_root", bundle.SourceEvidenceRoot, true)
	checkPublicPath(&errs, "source_readback_path", bundle.SourceReadbackPath, true)
	checkPublicPath(&errs, "source_assertion_path", bundle.SourceAssertionPath, true)
	if len(bundle.Tasks) < minTasks {
		errs = append(errs, fmt.Sprintf("ranked tasks must include at least %d tasks", minTasks))
	}
	themes := map[string]bool{}
	for i, task := range bundle.Tasks {
		prefix := fmt.Sprintf("tasks[%d]", i)
		wantRank := i + 1
		if task.Rank != wantRank {
			errs = append(errs, fmt.Sprintf("%s.rank must be %d", prefix, wantRank))
		}
		if task.ID != fmt.Sprintf("feature-depth-next-wave-%02d", wantRank) {
			errs = append(errs, fmt.Sprintf("%s.id must match ranked next-wave id", prefix))
		}
		requireField(&errs, prefix+".theme", task.Theme)
		if len(task.EvidenceRefs) == 0 {
			errs = append(errs, prefix+".evidence_refs must not be empty")
		}
		if !containsValue(task.EvidenceRefs, bundle.SourceReadbackPath) {
			errs = append(errs, prefix+".evidence_refs must include source_readback_path")
		}
		checkPublicStrings(&errs, prefix+".evidence_refs", task.EvidenceRefs, true)
		checkPublicPath(&errs, prefix+".theme", task.Theme, true)
		themes[task.Theme] = true
	}
	if len(themes) < 10 {
		errs = append(errs, "tasks must span at least 10 Feature Depth themes")
	}
	return joinErrors(errs)
}

func ValidateAtlasNextWaveRefactoringRecommendations(bundle AOMissionRefactoringRecommendations, minTasks int) error {
	var errs []string
	if minTasks <= 0 {
		minTasks = 40
	}
	if bundle.Schema != AOMissionRefactoringRecommendationsContract {
		errs = append(errs, "schema must be "+AOMissionRefactoringRecommendationsContract)
	}
	requireField(&errs, "mission_id", bundle.MissionID)
	if bundle.Status != "ready" {
		errs = append(errs, "status must be ready")
	}
	if bundle.Track != "refactoring" {
		errs = append(errs, "track must be refactoring")
	}
	if bundle.MinimumTasks < minTasks {
		errs = append(errs, fmt.Sprintf("minimum_tasks must be at least %d", minTasks))
	}
	if len(bundle.Tasks) < minTasks {
		errs = append(errs, fmt.Sprintf("tasks must include at least %d tasks", minTasks))
	}
	if bundle.RecommendationCount != len(bundle.Tasks) {
		errs = append(errs, "recommendation_count must match tasks length")
	}
	for field, path := range map[string]string{
		"source_evidence_root":     bundle.SourceEvidenceRoot,
		"source_readback_path":     bundle.SourceReadbackPath,
		"source_assertion_path":    bundle.SourceAssertionPath,
		"next_track_decision_path": bundle.NextTrackDecisionPath,
	} {
		requireField(&errs, field, path)
		checkPublicPath(&errs, field, path, true)
	}
	if !digestPattern.MatchString(bundle.NextTrackDecisionDigest) {
		errs = append(errs, "next_track_decision_digest must be sha256 digest")
	}
	if !digestPattern.MatchString(bundle.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if bundle.ConsumedLedgerPath != "" || bundle.ConsumedLedgerDigest != "" {
		requireField(&errs, "consumed_recommendation_ledger_path", bundle.ConsumedLedgerPath)
		checkPublicPath(&errs, "consumed_recommendation_ledger_path", bundle.ConsumedLedgerPath, true)
		if !digestPattern.MatchString(bundle.ConsumedLedgerDigest) {
			errs = append(errs, "consumed_recommendation_ledger_digest must be sha256 digest")
		}
	}
	seen := map[string]bool{}
	themes := map[string]bool{}
	for i, task := range bundle.Tasks {
		prefix := fmt.Sprintf("tasks[%d]", i)
		wantRank := i + 1
		if task.Rank != wantRank {
			errs = append(errs, fmt.Sprintf("%s.rank must be %d", prefix, wantRank))
		}
		if task.ID != fmt.Sprintf("refactoring-next-wave-%02d", wantRank) {
			errs = append(errs, prefix+".id must match ranked refactoring id")
		}
		if seen[task.ID] {
			errs = append(errs, prefix+".id must be unique")
		}
		seen[task.ID] = true
		requireField(&errs, prefix+".owner", task.Owner)
		if task.Owner != "ao-atlas" {
			errs = append(errs, prefix+".owner must be ao-atlas")
		}
		requireField(&errs, prefix+".theme", task.Theme)
		requireField(&errs, prefix+".task", task.Task)
		if len(strings.Fields(task.Task)) < 6 {
			errs = append(errs, prefix+".task must be a concrete actionable task")
		}
		if len(task.EvidenceRefs) < 3 {
			errs = append(errs, prefix+".evidence_refs must include readback, assertion, and next-track decision")
		}
		if !containsValue(task.EvidenceRefs, bundle.SourceReadbackPath) {
			errs = append(errs, prefix+".evidence_refs must include source_readback_path")
		}
		if !containsValue(task.EvidenceRefs, bundle.SourceAssertionPath) {
			errs = append(errs, prefix+".evidence_refs must include source_assertion_path")
		}
		if !containsValue(task.EvidenceRefs, bundle.NextTrackDecisionPath) {
			errs = append(errs, prefix+".evidence_refs must include next_track_decision_path")
		}
		checkPublicPath(&errs, prefix+".id", task.ID, true)
		checkPublicPath(&errs, prefix+".owner", task.Owner, true)
		checkPublicPath(&errs, prefix+".theme", task.Theme, true)
		checkPublicPath(&errs, prefix+".task", task.Task, true)
		checkPublicStrings(&errs, prefix+".evidence_refs", task.EvidenceRefs, true)
		themes[task.Theme] = true
	}
	if len(themes) < 10 {
		errs = append(errs, "tasks must span at least 10 refactoring themes")
	}
	if !bundle.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if bundle.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if bundle.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !bundle.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if bundle.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if bundle.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if bundle.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if bundle.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if bundle.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}

func validateNextWaveSourcePath(label, path string) error {
	var errs []string
	requireField(&errs, label, path)
	checkPublicPath(&errs, label, path, true)
	return joinErrors(errs)
}

func rejectSaturatedFeatureDepthContinuation(sourceEvidenceRoot, sourceReadbackPath string) error {
	sourceEvidenceRoot = filepath.ToSlash(strings.TrimSpace(sourceEvidenceRoot))
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	if sourceEvidenceRoot == "" || sourceReadbackPath == "" || !isAtlasFeatureDepthEvidenceRoot(sourceEvidenceRoot) {
		return nil
	}

	var readback AtlasRecommendationReadback
	if err := readJSONIfPossible(sourceReadbackPath, &readback); err != nil {
		return fmt.Errorf("read source_readback_path for Feature Depth saturation check: %w", err)
	}
	if !isSaturatedFeatureDepthReadback(readback) {
		return nil
	}
	return fmt.Errorf(
		"feature depth recommendations saturated: source readback %s completed %d/%d Feature Depth nodes with final_response_allowed=true; route to AO Atlas refactoring/strategy review instead of exporting another Feature Depth wave",
		filepath.ToSlash(sourceReadbackPath),
		readback.CompletedNodes,
		readback.TotalNodes,
	)
}

func isAtlasFeatureDepthEvidenceRoot(sourceEvidenceRoot string) bool {
	base := filepath.Base(filepath.ToSlash(strings.TrimSpace(sourceEvidenceRoot)))
	return strings.HasPrefix(base, "ao-atlas-feature-depth-wave-") ||
		strings.HasPrefix(base, "ao-atlas-feature-depth-followup-")
}

func isSaturatedFeatureDepthReadback(readback AtlasRecommendationReadback) bool {
	if !strings.Contains(readback.MissionID, "feature-depth") &&
		!strings.Contains(readback.TargetInstance, "feature-depth") {
		return false
	}
	return readback.Status == "completed" &&
		readback.TotalNodes > 0 &&
		readback.CompletedNodes == readback.TotalNodes &&
		readback.ReadyNodes == 0 &&
		readback.BlockedNodes == 0 &&
		readback.FailedNodes == 0 &&
		readback.FinalResponseAllowed &&
		readback.ReturnGateStatus == "final_response_allowed"
}

func nextWaveFeatureDepthSeeds() []struct {
	theme string
	task  string
} {
	return []struct {
		theme string
		task  string
	}{
		{"mission-readback-durability", "Bind AO Mission readback deltas to deterministic checkpoint comparison evidence."},
		{"mission-readback-durability", "Add resumable readback diff fixtures for completed and ready node transitions."},
		{"mission-readback-durability", "Create stale checkpoint rejection evidence for outdated Mission continuation prompts."},
		{"mission-readback-durability", "Add operator summary checks that preserve exact next action wording."},
		{"evidence-schema-coverage", "Extend evidence schema validation to typed node closure rollup artifacts."},
		{"evidence-schema-coverage", "Add schema coverage summaries for every generated run link artifact."},
		{"evidence-schema-coverage", "Validate required node evidence fields before recommendation readback advances."},
		{"evidence-schema-coverage", "Record schema validator drift evidence for regenerated fixture directories."},
		{"pr-ci-telemetry", "Build aggregate PR and CI timing summaries across consolidation wave nodes."},
		{"pr-ci-telemetry", "Add long running Windows check threshold evidence to PR ledger rows."},
		{"pr-ci-telemetry", "Create failed check replay fixtures for retry and no merge decisions."},
		{"pr-ci-telemetry", "Bind merge commit readbacks to passed required check conclusions."},
		{"branch-lifecycle-recovery", "Add post merge branch deletion readback to every node closure packet."},
		{"branch-lifecycle-recovery", "Create stale remote branch repair evidence for interrupted cleanup handoffs."},
		{"branch-lifecycle-recovery", "Validate local main synchronization before selecting the next executable node."},
		{"branch-lifecycle-recovery", "Record branch cleanup ledger summaries in final operator handoff evidence."},
		{"compaction-resume-fidelity", "Generate compaction resume prompts that preserve lease timing and active node state."},
		{"compaction-resume-fidelity", "Add resume prompt regression fixtures for ready nodes and exact next actions."},
		{"compaction-resume-fidelity", "Bind checkpoint digests into resume prompts for interruption recovery audits."},
		{"compaction-resume-fidelity", "Create resume denial evidence when final response remains blocked by ready work."},
		{"public-safety-scanning", "Bind Sentinel wording scan results into final closure readback status fields."},
		{"public-safety-scanning", "Add scoped public safety scans for changed evidence and prompt artifacts."},
		{"public-safety-scanning", "Create negative wording fixtures for unsafe authority promotion statements."},
		{"public-safety-scanning", "Summarize public safety scan coverage in machine readable closure rollups."},
		{"promoter-command-rollups", "Aggregate Promoter no promotion verdicts across completed hardening and closure nodes."},
		{"promoter-command-rollups", "Bind Command compact readback agreement to Promoter no promotion summaries."},
		{"promoter-command-rollups", "Add regression evidence for no promotion rollup count mismatches."},
		{"promoter-command-rollups", "Create final response denial evidence when Command and Promoter disagree."},
		{"foundry-handoff-binding", "Bind Foundry import readiness records to exactly one active mutation node."},
		{"foundry-handoff-binding", "Add Foundry run link digest checks for completed node evidence packets."},
		{"foundry-handoff-binding", "Create Foundry handoff replay fixtures for resumed bounded implementation nodes."},
		{"foundry-handoff-binding", "Validate Foundry terminal status examples against recommendation readback enums."},
		{"dashboard-provenance", "Bind Atlas final closure evidence into multi repo Mission dashboard rows."},
		{"dashboard-provenance", "Add dashboard provenance links for Foundry Promoter Command and Sentinel evidence."},
		{"dashboard-provenance", "Create dashboard freshness checks for merged PR and synced main state."},
		{"dashboard-provenance", "Summarize blocked versus ready Mission nodes in compact dashboard filters."},
		{"next-wave-generation", "Export ranked Feature Depth tasks from final closure readback evidence."},
		{"next-wave-generation", "Add next wave prompt generation with minimum two hour work budget language."},
		{"next-wave-generation", "Validate next wave recommendations remain planning only until imported."},
		{"next-wave-generation", "Generate final Feature Depth recommendations for operator handoff review."},
	}
}

func nextWaveRefactoringSeeds() []struct {
	theme string
	task  string
} {
	return []struct {
		theme string
		task  string
	}{
		{"recommendation-routing", "Extract recommendation routing command dispatch into a typed registry with deterministic help output."},
		{"recommendation-routing", "Replace duplicated recommendation command lists with one shared registry-backed command catalog."},
		{"recommendation-routing", "Bind completed Feature Depth routing decisions to refactoring wave generation without stale exports."},
		{"recommendation-routing", "Add consumed recommendation ledger checks before any next wave exporter can run."},
		{"recommendation-routing", "Separate planning-only recommendation export commands from mutation-capable node execution commands."},
		{"evidence-schema-registry", "Refactor recommendation evidence schema registry entries into typed constructors with drift tests."},
		{"evidence-schema-registry", "Move schema contract constants into grouped recommendation evidence namespaces with coverage tests."},
		{"evidence-schema-registry", "Add registry-backed typed validator lookup for recommendation control-plane artifacts."},
		{"evidence-schema-registry", "Collapse duplicated schema coverage failure wording into reusable validation helpers."},
		{"evidence-schema-registry", "Introduce schema registry golden fixtures for command output and typed validator bindings."},
		{"readback-gates", "Centralize final response gate evaluation for readback, execution readback, and closure readbacks."},
		{"readback-gates", "Refactor exact next action preservation into a shared readback transition helper."},
		{"readback-gates", "Add compact readback status normalization for ready, completed, blocked, and failed node states."},
		{"readback-gates", "Bind return gate denial reasons to structured fields instead of repeated text fragments."},
		{"readback-gates", "Create regression fixtures for stale readback rejection across all recommendation tracks."},
		{"run-ledger", "Unify command run-ledger, rollup, and coverage-check builders behind a common artifact summary type."},
		{"run-ledger", "Refactor run-ledger output status classification into reusable pass, ready, failed, and blocked categories."},
		{"run-ledger", "Add ledger coverage checks for refactoring exporters and track routing artifacts."},
		{"run-ledger", "Bind run-ledger rollups to final operator summaries without self-referential ledger requirements."},
		{"run-ledger", "Create long-run ledger fixture packs for repeated command retries and resumed sessions."},
		{"pr-ci-lifecycle", "Normalize PR and CI ledger rows across feature depth, closure, and refactoring waves."},
		{"pr-ci-lifecycle", "Extract Windows long-running check telemetry into shared threshold and wait-state helpers."},
		{"pr-ci-lifecycle", "Add merge readiness guard helpers that require passed checks before branch cleanup evidence."},
		{"pr-ci-lifecycle", "Refactor post-merge branch deletion readbacks into reusable local and remote cleanup records."},
		{"pr-ci-lifecycle", "Create PR lifecycle replay fixtures for interrupted merge, sync, and cleanup handoffs."},
		{"prompt-generation", "Refactor continuation prompt generation to consume structured wave budgets and stop conditions."},
		{"prompt-generation", "Add prompt compaction resume fixtures that preserve next node, exact action, and final gate denial."},
		{"prompt-generation", "Move prompt safety boundary rendering into one audited template helper."},
		{"prompt-generation", "Bind generated prompts to source readback digests and consumed recommendation ledgers."},
		{"prompt-generation", "Add long-run prompt regression fixtures for two to three hour refactoring waves."},
		{"dashboard-provenance", "Refactor mission dashboard rows to share provenance digest and freshness evaluation helpers."},
		{"dashboard-provenance", "Add dashboard filters for recommendation track, schema health, CI state, and cleanup state."},
		{"dashboard-provenance", "Bind dashboard closure rows to Promoter, Command, Sentinel, and Foundry rollup evidence."},
		{"dashboard-provenance", "Create stale dashboard evidence detection for superseded wave readbacks and old exports."},
		{"dashboard-provenance", "Add dashboard compact rendering tests for completed waves with no ready nodes."},
		{"test-structure", "Split oversized recommendation tests into focused files by routing, evidence, readback, and lifecycle domain."},
		{"test-structure", "Extract shared recommendation test fixture builders for waves, nodes, and readbacks."},
		{"test-structure", "Add table-driven validator tests for no-promotion and RSI-denied boundary fields."},
		{"architecture-boundaries", "Create targeted regression suites that avoid rerunning unrelated long-wave fixture assertions."},
		{"developer-experience", "Document the refactoring wave handoff with ranked tasks and verification gates."},
	}
}

func ValidateAtlasRecommendationWave(wave AtlasRecommendationWave) error {
	var errs []string
	requireContract(&errs, "atlas_recommendation_wave", wave.ContractVersion, AtlasRecommendationWaveContract)
	requireField(&errs, "mission_id", wave.MissionID)
	requireField(&errs, "target_instance", wave.TargetInstance)
	if wave.Status != "ready" {
		errs = append(errs, "status must be ready")
	}
	if !digestPattern.MatchString(wave.SourceDigest) {
		errs = append(errs, "source_digest must be sha256 digest")
	}
	if wave.MinimumTasks < 1 {
		errs = append(errs, "minimum_tasks must be positive")
	}
	if wave.TotalTasks != len(wave.Tasks) {
		errs = append(errs, "total_tasks must match tasks length")
	}
	if wave.TotalTasks < wave.MinimumTasks {
		errs = append(errs, "total_tasks must meet minimum_tasks")
	}
	if wave.NodeBudget < wave.MinimumTasks || wave.NodeBudget > wave.TotalTasks {
		errs = append(errs, "node_budget must be between minimum_tasks and total_tasks")
	}
	if wave.MinimumTasks >= 20 && wave.EstimatedMinutes < 90 {
		errs = append(errs, "estimated_minutes must be at least 90 for a 20-task wave")
	}
	if wave.Supervisor != nil {
		if wave.Supervisor.ContractVersion != "ao.atlas.long-run-supervisor.v0.2" {
			errs = append(errs, "supervisor.contract_version must be ao.atlas.long-run-supervisor.v0.2")
		}
		if wave.Supervisor.MinNodes != wave.MinimumTasks {
			errs = append(errs, "supervisor.min_nodes must match minimum_tasks")
		}
		if wave.Supervisor.MinMinutes < 0 {
			errs = append(errs, "supervisor.min_minutes must be zero or greater")
		}
		if wave.Supervisor.MaxMinutes < wave.Supervisor.MinMinutes {
			errs = append(errs, "supervisor.max_minutes must be greater than or equal to min_minutes")
		}
		if wave.Supervisor.MinNodes >= 30 && wave.Supervisor.MinMinutes > 0 && wave.Supervisor.MinMinutes < 120 {
			errs = append(errs, "supervisor.min_minutes must be at least 120 for a 30-node wave")
		}
		if wave.Supervisor.MinNodes >= 30 && wave.Supervisor.MaxMinutes < 180 {
			errs = append(errs, "supervisor.max_minutes must support a 2-3 hour wave")
		}
		if wave.Supervisor.ContinueIfFastTarget < wave.Supervisor.MinNodes || wave.Supervisor.ContinueIfFastTarget > wave.TotalTasks {
			errs = append(errs, "supervisor.continue_if_fast_target must be between min_nodes and total_tasks")
		}
		requireField(&errs, "supervisor.return_only_when", wave.Supervisor.ReturnOnlyWhen)
		requireField(&errs, "supervisor.checkpoint_policy", wave.Supervisor.CheckpointPolicy)
		requireField(&errs, "supervisor.evidence_policy", wave.Supervisor.EvidencePolicy)
		requireField(&errs, "supervisor.final_report_contract", wave.Supervisor.FinalReportContract)
	}
	if wave.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while recommendation nodes are ready")
	}
	requireField(&errs, "final_response_reason", wave.FinalResponseReason)
	requireField(&errs, "promoter_readback_status", wave.PromoterReadbackStatus)
	requireField(&errs, "command_readback_status", wave.CommandReadbackStatus)
	requireField(&errs, "public_safety_scan_status", wave.PublicSafetyScanStatus)
	requireField(&errs, "next_recommended_prompt", wave.NextRecommendedPrompt)
	for i, task := range wave.Tasks {
		prefix := fmt.Sprintf("tasks[%d]", i)
		requireField(&errs, prefix+".id", task.ID)
		if task.Owner != "ao-atlas" {
			errs = append(errs, prefix+".owner must be ao-atlas")
		}
		requireField(&errs, prefix+".task", task.Task)
		requireField(&errs, prefix+".node_id", task.NodeID)
		requireField(&errs, prefix+".task_id", task.TaskID)
		if task.MutationClass != "low_risk_code" {
			errs = append(errs, prefix+".mutation_class must be low_risk_code")
		}
		if !digestPattern.MatchString(task.SourceTaskDigest) {
			errs = append(errs, prefix+".source_task_digest must be sha256 digest")
		}
		if task.TargetFactoryRepo != "ao-atlas" {
			errs = append(errs, prefix+".target_factory_repo must be ao-atlas")
		}
		requireField(&errs, prefix+".factory_folder", task.FactoryFolder)
		requireList(&errs, prefix+".required_gates", task.RequiredGates)
		requireList(&errs, prefix+".verification_commands", task.Verification)
		requireList(&errs, prefix+".safety_limits", task.SafetyLimits)
		checkPublicStrings(&errs, prefix+".required_gates", task.RequiredGates, true)
		checkPublicStrings(&errs, prefix+".verification_commands", task.Verification, true)
		checkPublicStrings(&errs, prefix+".safety_limits", task.SafetyLimits, true)
	}
	if wave.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if wave.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if wave.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if wave.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationWorkgraph(wave AtlasRecommendationWave) (Workgraph, error) {
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		return Workgraph{}, err
	}
	nodes := make([]WorkgraphNode, 0, len(wave.Tasks))
	for i, item := range wave.Tasks {
		deps := []string{}
		if i > 0 {
			deps = append(deps, wave.Tasks[i-1].NodeID)
		}
		nodes = append(nodes, WorkgraphNode{
			ID:           item.NodeID,
			Status:       "ready",
			Dependencies: deps,
			Blockers:     []string{},
			StitchTask:   i%5 == 0,
			FactoryTask: FactoryTask{
				ContractVersion:   FactoryTaskContract,
				ID:                item.TaskID,
				Objective:         item.Task,
				TargetFactoryRepo: item.TargetFactoryRepo,
				FactoryFolder:     item.FactoryFolder,
				MutationClass:     item.MutationClass,
				Acceptance: []string{
					"node gate, candidate record, rollback record, tests, verification, and readback evidence are recorded",
					"Atlas final response remains denied while ready work or exact next actions remain",
				},
				NonGoals: []string{
					"do not execute Foundry work from Atlas",
					"do not widen AO authority or claim broad RSI",
				},
				WriteScope: []string{
					"internal/atlas",
					"schemas",
					"examples",
					"docs/evidence",
				},
				RequiredGates:     append([]string(nil), item.RequiredGates...),
				RollbackScope:     []string{"revert node-specific Atlas changes and generated evidence"},
				Verification:      append([]string(nil), item.Verification...),
				RequiredEvidence:  []string{"source_digest:" + wave.SourceDigest, "source_recommendation:" + item.ID, "source_task_digest:" + item.SourceTaskDigest},
				SafetyLimits:      append([]string(nil), item.SafetyLimits...),
				AuthorityBoundary: "atlas_recommendation_planning_only",
				DependencyRefs:    append([]string(nil), deps...),
				ContextPackRefs:   []string{},
			},
		})
	}
	workgraph := Workgraph{
		ContractVersion: WorkgraphContract,
		ID:              "ao-atlas-recommendation-wave-" + sanitizeMissionProvenanceNodeName(wave.MissionID),
		TargetInstance:  wave.TargetInstance,
		Nodes:           nodes,
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return Workgraph{}, err
	}
	return workgraph, nil
}

func WriteAtlasRecommendationWaveArtifacts(outDir string, result AtlasRecommendationWaveResult) error {
	if strings.TrimSpace(outDir) == "" {
		return fmt.Errorf("out directory is required")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "recommendation-wave.json"), result.Wave); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "recommendation-workgraph.json"), result.Workgraph); err != nil {
		return err
	}
	leaseStart, err := BuildAtlasRecommendationLeaseStart(result.Wave, result.Workgraph, AtlasRecommendationLeaseStartOptions{
		WavePath:      filepath.Join(outDir, "recommendation-wave.json"),
		WorkgraphPath: filepath.Join(outDir, "recommendation-workgraph.json"),
		EvidenceRoot:  filepath.ToSlash(outDir),
	})
	if err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "lease-start.json"), leaseStart); err != nil {
		return err
	}
	readback, err := BuildAtlasRecommendationReadback(result.Wave, result.Workgraph, AtlasRecommendationReadbackOptions{
		WavePath:      filepath.Join(outDir, "recommendation-wave.json"),
		WorkgraphPath: filepath.Join(outDir, "recommendation-workgraph.json"),
		EvidenceRoot:  filepath.ToSlash(outDir),
	})
	if err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "recommendation-readback.json"), readback); err != nil {
		return err
	}
	readinessPacket, err := BuildAtlasRecommendationWorkgraphReadinessPacket(readback, AtlasRecommendationWorkgraphReadinessPacketOptions{
		WavePath:      filepath.Join(outDir, "recommendation-wave.json"),
		WorkgraphPath: filepath.Join(outDir, "recommendation-workgraph.json"),
		ReadbackPath:  filepath.Join(outDir, "recommendation-readback.json"),
	})
	if err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "workgraph-readiness-packet.json"), readinessPacket); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "next-recommended-prompt.md"), []byte(result.Prompt), 0o644)
}

func BuildAtlasRecommendationLeaseStart(wave AtlasRecommendationWave, workgraph Workgraph, options AtlasRecommendationLeaseStartOptions) (AtlasRecommendationLeaseStart, error) {
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		return AtlasRecommendationLeaseStart{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return AtlasRecommendationLeaseStart{}, err
	}
	if wave.TargetInstance != workgraph.TargetInstance {
		return AtlasRecommendationLeaseStart{}, fmt.Errorf("target_instance mismatch between recommendation wave and workgraph")
	}
	startedAt := strings.TrimSpace(options.StartedAt)
	if startedAt == "" {
		startedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if _, err := time.Parse(time.RFC3339, startedAt); err != nil {
		return AtlasRecommendationLeaseStart{}, fmt.Errorf("started_at must be RFC3339: %w", err)
	}
	minMinutes := wave.EstimatedMinutes
	maxMinutes := wave.EstimatedMinutes
	continueIfFastTarget := wave.NodeBudget
	checkpointPolicy := "after_each_node_or_timed_interval"
	if wave.Supervisor != nil {
		minMinutes = wave.Supervisor.MinMinutes
		maxMinutes = wave.Supervisor.MaxMinutes
		continueIfFastTarget = wave.Supervisor.ContinueIfFastTarget
		checkpointPolicy = wave.Supervisor.CheckpointPolicy
	}
	waveDigest := digestValue(wave)
	if strings.TrimSpace(options.WavePath) != "" {
		digest, err := digestFile(options.WavePath)
		if err != nil {
			return AtlasRecommendationLeaseStart{}, err
		}
		waveDigest = digest
	}
	workgraphDigest := digestValue(workgraph)
	if strings.TrimSpace(options.WorkgraphPath) != "" {
		digest, err := digestFile(options.WorkgraphPath)
		if err != nil {
			return AtlasRecommendationLeaseStart{}, err
		}
		workgraphDigest = digest
	}
	leaseStart := AtlasRecommendationLeaseStart{
		Schema:                 "ao.atlas.recommendation-lease-start.v0.1",
		Status:                 "started",
		MissionID:              wave.MissionID,
		TargetInstance:         wave.TargetInstance,
		EvidenceRoot:           filepath.ToSlash(strings.TrimSpace(options.EvidenceRoot)),
		StartedAt:              startedAt,
		MinMinutes:             minMinutes,
		MaxMinutes:             maxMinutes,
		ContinueIfFastTarget:   continueIfFastTarget,
		CheckpointPolicy:       checkpointPolicy,
		WaveDigest:             waveDigest,
		WorkgraphDigest:        workgraphDigest,
		FinalResponseAllowed:   false,
		FinalResponseReason:    "lease start marker does not allow final response",
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		MutatesRepositories:    false,
		CallsProviders:         false,
		ClaimsAuthorityAdvance: false,
	}
	if err := ValidateAtlasRecommendationLeaseStart(leaseStart); err != nil {
		return AtlasRecommendationLeaseStart{}, err
	}
	return leaseStart, nil
}
