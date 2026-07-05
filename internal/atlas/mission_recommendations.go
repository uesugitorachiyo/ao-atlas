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
	MaxMinutes           int
	ContinueIfFastTarget int
	ReturnOnlyWhen       string
	CheckpointPolicy     string
	EvidencePolicy       string
	FinalReportContract  string
}

type AtlasRecommendationWaveResult struct {
	Wave      AtlasRecommendationWave
	Workgraph Workgraph
	Prompt    string
}

type AtlasRecommendationReadbackOptions struct {
	WavePath        string
	WorkgraphPath   string
	EvidenceRoot    string
	StartedAt       string
	CompletedAt     string
	ElapsedMinutes  int
	LeaseTimingMode string
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
	if minMinutes <= 0 {
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
		} else {
			maxMinutes = minMinutes
		}
	}
	estimatedMinutes := options.EstimatedMinutes
	if estimatedMinutes <= 0 {
		estimatedMinutes = minMinutes
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
		if wave.Supervisor.MinMinutes < 1 {
			errs = append(errs, "supervisor.min_minutes must be positive")
		}
		if wave.Supervisor.MaxMinutes < wave.Supervisor.MinMinutes {
			errs = append(errs, "supervisor.max_minutes must be greater than or equal to min_minutes")
		}
		if wave.Supervisor.MinNodes >= 30 && wave.Supervisor.MinMinutes < 120 {
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

func BuildAtlasRecommendationReadback(wave AtlasRecommendationWave, workgraph Workgraph, options AtlasRecommendationReadbackOptions) (AtlasRecommendationReadback, error) {
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		return AtlasRecommendationReadback{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return AtlasRecommendationReadback{}, err
	}
	if wave.TargetInstance != workgraph.TargetInstance {
		return AtlasRecommendationReadback{}, fmt.Errorf("target_instance mismatch between recommendation wave and workgraph")
	}
	if len(workgraph.Nodes) != len(wave.Tasks) {
		return AtlasRecommendationReadback{}, fmt.Errorf("workgraph node count must match recommendation tasks")
	}
	taskByNode := map[string]AtlasRecommendationTask{}
	for _, task := range wave.Tasks {
		taskByNode[task.NodeID] = task
	}
	for _, node := range workgraph.Nodes {
		task, ok := taskByNode[node.ID]
		if !ok {
			return AtlasRecommendationReadback{}, fmt.Errorf("workgraph node %s is not present in recommendation wave", node.ID)
		}
		if task.TaskID != node.FactoryTask.ID {
			return AtlasRecommendationReadback{}, fmt.Errorf("workgraph node %s task_id mismatch", node.ID)
		}
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return AtlasRecommendationReadback{}, err
	}
	completed := state.NodeCounts["completed"]
	ready := state.NodeCounts["ready"]
	blocked := state.NodeCounts["blocked"]
	failed := state.NodeCounts["failed"]
	executableReady := len(state.ExecutableReadyNodeIDs)
	firstExecutable := ""
	if executableReady > 0 {
		firstExecutable = state.ExecutableReadyNodeIDs[0]
	}
	nodesComplete := completed == wave.TotalTasks && ready == 0 && blocked == 0 && failed == 0
	leaseTiming, err := buildRecommendationLeaseTiming(wave, options, nodesComplete)
	if err != nil {
		return AtlasRecommendationReadback{}, err
	}
	finalAllowed := nodesComplete && leaseTiming.MinMinutesMet
	finalReason := "ready nodes or exact next actions remain"
	exactNextAction := "Complete dependency chain so the next Atlas recommendation node becomes executable-ready."
	leaseHealth := "minimum_unmet"
	earlyReturnRisk := "blocked_final_response_minimum_unmet"
	if finalAllowed {
		finalReason = "all generated nodes complete and no ready nodes remain"
		exactNextAction = "Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks."
		leaseHealth = "all_generated_nodes_complete"
		earlyReturnRisk = "cleared_no_ready_nodes_remain"
	} else if nodesComplete && leaseTiming.LeaseTimeStatus == "lease_timing_missing" {
		finalReason = "minimum lease timing evidence missing"
		exactNextAction = "Record started_at, completed_at, and elapsed_minutes before evaluating final response for the long-run lease."
		leaseHealth = "minimum_minutes_timing_missing"
		earlyReturnRisk = "blocked_final_response_minimum_timing_missing"
	} else if nodesComplete && !leaseTiming.MinMinutesMet {
		finalReason = "minimum lease minutes unmet"
		exactNextAction = "Generate and execute the next useful Atlas recommendation wave until elapsed_minutes meets supervisor.min_minutes, or record a true hard blocker."
		leaseHealth = "minimum_minutes_unmet_continue_next_wave"
		earlyReturnRisk = "blocked_final_response_minimum_minutes_unmet"
	} else if blocked > 0 || failed > 0 {
		finalReason = "true hard blocker requires exact repair evidence"
		exactNextAction = "Resolve blocked or failed recommendation node with exact repair evidence."
		leaseHealth = "hard_blocker_requires_repair"
		earlyReturnRisk = "hard_blocker_requires_exact_missing_evidence"
	} else {
		if completed >= wave.MinimumTasks {
			leaseHealth = "minimum_met_continue_if_fast"
		}
		if ready > 0 {
			earlyReturnRisk = "blocked_final_response_ready_nodes_remain"
		}
		if firstExecutable != "" {
			exactNextAction = fmt.Sprintf("Emit Foundry import for %s and execute exactly one active node.", firstExecutable)
		}
	}
	returnGateStatus := recommendationReturnGateStatus(finalAllowed, nodesComplete, leaseTiming, ready, blocked, failed)
	status := "ready"
	if finalAllowed {
		status = "completed"
	} else if blocked > 0 || failed > 0 {
		status = "blocked"
	} else if completed > 0 {
		status = "in_progress"
	}
	foundryRollupStatus := "required_pending_first_node_import"
	promoterReadbackStatus := wave.PromoterReadbackStatus
	promoterNoPromotionStatus := "required_not_bound_until_promotion_evidence_exists"
	commandReadbackStatus := wave.CommandReadbackStatus
	commandTimelineStatus := "compact_timeline_required_before_closure"
	if completed > 0 {
		foundryRollupStatus = "in_progress_node_run_links_recorded"
		promoterNoPromotionStatus = "in_progress_no_promotion_recorded"
		commandTimelineStatus = "in_progress_compact_timeline_recorded"
	}
	if finalAllowed {
		foundryRollupStatus = "completed_all_node_run_links_recorded"
		promoterReadbackStatus = "no_promotion_recorded"
		promoterNoPromotionStatus = "recorded_no_promotion_for_recommendation_wave"
		commandReadbackStatus = "compact_timeline_recorded"
		commandTimelineStatus = "recorded_compact_timeline_for_completed_wave"
	}
	waveDigest := digestValue(wave)
	if strings.TrimSpace(options.WavePath) != "" {
		if digest, err := digestFile(options.WavePath); err == nil {
			waveDigest = digest
		} else {
			return AtlasRecommendationReadback{}, err
		}
	}
	workgraphDigest := digestValue(workgraph)
	if strings.TrimSpace(options.WorkgraphPath) != "" {
		if digest, err := digestFile(options.WorkgraphPath); err == nil {
			workgraphDigest = digest
		} else {
			return AtlasRecommendationReadback{}, err
		}
	}
	readback := AtlasRecommendationReadback{
		ContractVersion:           AtlasRecommendationReadbackContract,
		MissionID:                 wave.MissionID,
		TargetInstance:            wave.TargetInstance,
		Status:                    status,
		SourceDigest:              wave.SourceDigest,
		WaveDigest:                waveDigest,
		WorkgraphDigest:           workgraphDigest,
		EvidenceRoot:              filepath.ToSlash(strings.TrimSpace(options.EvidenceRoot)),
		Supervisor:                wave.Supervisor,
		StartedAt:                 leaseTiming.StartedAt,
		CompletedAt:               leaseTiming.CompletedAt,
		ElapsedMinutes:            leaseTiming.ElapsedMinutes,
		MinMinutesMet:             leaseTiming.MinMinutesMet,
		LeaseTimeStatus:           leaseTiming.LeaseTimeStatus,
		TotalNodes:                wave.TotalTasks,
		MinimumNodes:              wave.MinimumTasks,
		CompletedNodes:            completed,
		ReadyNodes:                ready,
		BlockedNodes:              blocked,
		FailedNodes:               failed,
		ExecutableReadyNodes:      executableReady,
		FirstExecutableNode:       firstExecutable,
		LeaseHealthStatus:         leaseHealth,
		CheckpointFreshnessStatus: "fresh_checkpoint_required_after_each_node_or_timed_interval",
		StaleRouteDecisionStatus:  "fresh_atlas_supervises_foundry_owns_one_active_node",
		EarlyReturnRiskStatus:     earlyReturnRisk,
		FoundryRollupStatus:       foundryRollupStatus,
		FoundryTerminalStatusReadback: map[string]string{
			"completed": "terminal_success_can_close_when_all_nodes_and_readbacks_are_complete",
			"promoted":  "terminal_success_can_close_when_promoter_and_command_agree",
			"denied":    "terminal_denial_requires_exact_missing_evidence_readback",
			"blocked":   "terminal_blocker_requires_repair_or_checkpoint_resume",
		},
		FoundryTerminalStatusExamples:   foundryTerminalStatusExamples(),
		FoundryDeniedTerminalExamples:   foundryDeniedTerminalExamples(),
		PromoterReadbackStatus:          promoterReadbackStatus,
		PromoterNoPromotionStatus:       promoterNoPromotionStatus,
		PromoterNoPromotionPlaceholders: promoterNoPromotionPlaceholders(),
		CommandReadbackStatus:           commandReadbackStatus,
		CommandTimelineStatus:           commandTimelineStatus,
		CommandTimelinePlaceholders:     commandTimelinePlaceholders(),
		PublicSafetyScanStatus:          wave.PublicSafetyScanStatus,
		ReturnGateStatus:                returnGateStatus,
		CheckpointCount:                 completed,
		FinalResponseAllowed:            finalAllowed,
		FinalResponseDenialGate:         recommendationFinalResponseDenialGate(finalAllowed, returnGateStatus),
		FinalResponseReason:             finalReason,
		ExactNextAction:                 exactNextAction,
		ContinuationContract:            buildAtlasContinuationContract(ready, exactNextAction, returnGateStatus, finalAllowed),
		ExactNextActionReadback:         buildExactNextActionReadback(exactNextAction, firstExecutable, returnGateStatus, finalAllowed),
		NodeEvidence:                    recommendationNodeEvidence(workgraph),
		FeatureDepthRecommendations:     featureDepthRecommendationReadback(wave.Tasks, 10),
		SafetyBoundaries: map[string]bool{
			"provider_calls":                    false,
			"credential_inspection":             false,
			"direct_main_mutation":              false,
			"release_deploy_publish_upload_tag": false,
			"dependency_updates":                false,
			"auth_policy_config_widening":       false,
			"hidden_instruction_mutation":       false,
			"broad_rsi_claim":                   false,
			"rsi_remains_denied":                true,
		},
		SchedulesWork: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasRecommendationReadback{}, err
	}
	return readback, nil
}

func buildExactNextActionReadback(action, nextExecutableNode, returnGateStatus string, finalResponseAllowed bool) AtlasExactNextActionReadback {
	status := "continuation_required"
	if finalResponseAllowed {
		status = "ready_for_final_response"
	}
	return AtlasExactNextActionReadback{
		Status:               status,
		Action:               action,
		NextExecutableNode:   nextExecutableNode,
		ReturnGateStatus:     returnGateStatus,
		FinalResponseAllowed: finalResponseAllowed,
		Source:               "recommendation_readback",
	}
}

func commandTimelinePlaceholders() []AtlasCommandTimelinePlaceholder {
	return []AtlasCommandTimelinePlaceholder{
		{
			Slot:                        "checkpoint",
			Source:                      "recommendation_readback",
			Status:                      "pending_command_timeline",
			Summary:                     "Bind completed_nodes, ready_nodes, checkpoint_count, and elapsed_minutes into the compact Command timeline.",
			RequiredBeforeFinalResponse: true,
		},
		{
			Slot:                        "exact_next_action",
			Source:                      "recommendation_readback",
			Status:                      "pending_command_timeline",
			Summary:                     "Bind exact_next_action and first_executable_node into the compact Command timeline.",
			RequiredBeforeFinalResponse: true,
		},
		{
			Slot:                        "return_gate",
			Source:                      "recommendation_readback",
			Status:                      "pending_command_timeline",
			Summary:                     "Bind return_gate_status and final_response_allowed into the compact Command timeline.",
			RequiredBeforeFinalResponse: true,
		},
	}
}

func promoterNoPromotionPlaceholders() []AtlasPromoterNoPromotionPlaceholder {
	return []AtlasPromoterNoPromotionPlaceholder{
		{
			Slot:                        "promotion_claim",
			Source:                      "recommendation_readback",
			Status:                      "pending_promoter_no_promotion",
			Summary:                     "Bind promotion_claimed=false and the no-promotion summary before closure.",
			RequiredBeforeFinalResponse: true,
		},
		{
			Slot:                        "rsi_boundary",
			Source:                      "recommendation_readback",
			Status:                      "pending_promoter_no_promotion",
			Summary:                     "Bind rsi_remains_denied=true and next_denied_class=RSI before closure.",
			RequiredBeforeFinalResponse: true,
		},
		{
			Slot:                        "authority_advance",
			Source:                      "recommendation_readback",
			Status:                      "pending_promoter_no_promotion",
			Summary:                     "Bind claims_authority_advance=false plus no scheduling, execution, or approval authority.",
			RequiredBeforeFinalResponse: true,
		},
	}
}

func foundryTerminalStatusExamples() []AtlasFoundryTerminalStatusExample {
	return []AtlasFoundryTerminalStatusExample{
		{
			SourceStatus:     "completed",
			NormalizedStatus: "completed",
			Terminal:         true,
			CanCloseMission:  true,
			RequiredReadback: "Foundry rollup reports completed, all node evidence exists, and no ready nodes remain.",
		},
		{
			SourceStatus:     "promoted",
			NormalizedStatus: "completed",
			Terminal:         true,
			CanCloseMission:  true,
			RequiredReadback: "Promoter and Command agree promotion is terminal, RSI remains denied, and no ready nodes remain.",
		},
		{
			SourceStatus:     "denied",
			NormalizedStatus: "denied",
			Terminal:         true,
			CanCloseMission:  true,
			RequiredReadback: "Denial readback includes exact missing evidence, no ready repair node remains, and no authority advance is claimed.",
		},
		{
			SourceStatus:     "blocked",
			NormalizedStatus: "blocked",
			Terminal:         true,
			CanCloseMission:  false,
			RequiredReadback: "Blocker readback names the exact repair or resume action before final response can close.",
		},
	}
}

func foundryDeniedTerminalExamples() []AtlasFoundryDeniedTerminalExample {
	return []AtlasFoundryDeniedTerminalExample{
		{
			DenialReason:                 "missing_node_evidence",
			NormalizedStatus:             "denied",
			Terminal:                     true,
			CanCloseMission:              true,
			RequiresExactMissingEvidence: true,
			RequiredReadback:             "Denied rollup names the missing node id, missing evidence key, and expected evidence path.",
			RSIRemainsDenied:             true,
			AuthorityAdvanceClaimed:      false,
		},
		{
			DenialReason:                 "missing_stop_gate_evidence",
			NormalizedStatus:             "denied",
			Terminal:                     true,
			CanCloseMission:              true,
			RequiresExactMissingEvidence: true,
			RequiredReadback:             "Denied rollup names the uncleared stop gate and the exact artifact needed before promotion can be reconsidered.",
			RSIRemainsDenied:             true,
			AuthorityAdvanceClaimed:      false,
		},
		{
			DenialReason:                 "forbidden_surface_or_rsi_claim",
			NormalizedStatus:             "denied",
			Terminal:                     true,
			CanCloseMission:              true,
			RequiresExactMissingEvidence: true,
			RequiredReadback:             "Denied rollup records the forbidden surface or RSI claim, refuses promotion, and keeps RSI denied.",
			RSIRemainsDenied:             true,
			AuthorityAdvanceClaimed:      false,
		},
	}
}

func recommendationReturnGateStatus(finalAllowed bool, nodesComplete bool, leaseTiming atlasRecommendationLeaseTiming, ready, blocked, failed int) string {
	if finalAllowed {
		return "final_response_allowed"
	}
	if blocked > 0 || failed > 0 {
		return "blocked_hard_blocker"
	}
	if nodesComplete && leaseTiming.LeaseTimeStatus == "lease_timing_missing" {
		return "blocked_lease_timing_missing"
	}
	if nodesComplete && !leaseTiming.MinMinutesMet {
		return "blocked_minimum_minutes_unmet"
	}
	if ready > 0 {
		return "blocked_ready_nodes_remain"
	}
	return "blocked_no_executable_ready_node"
}

func recommendationFinalResponseDenialGate(finalAllowed bool, returnGateStatus string) string {
	if finalAllowed {
		return "allow_final_response"
	}
	if returnGateStatus == "blocked_hard_blocker" {
		return "blocked_hard_blocker"
	}
	return "deny_ready_nodes_or_exact_next_action_remain"
}

func buildAtlasContinuationContract(readyNodes int, exactNextAction, returnGateStatus string, finalResponseAllowed bool) AtlasContinuationContract {
	status := "ready_for_final_response"
	refusesFinalResponse := false
	reason := "final response allowed by recommendation readback"
	if !finalResponseAllowed {
		status = "continuation_required"
		refusesFinalResponse = true
		reason = atlasContinuationContractReason(readyNodes, exactNextAction, returnGateStatus)
		if readyNodes == 0 && strings.TrimSpace(exactNextAction) == "" {
			status = "blocked"
		}
	}
	return AtlasContinuationContract{
		ContractVersion:      "ao.atlas.continuation-contract.v0.1",
		Status:               status,
		ReadyNodes:           readyNodes,
		ExactNextAction:      exactNextAction,
		ReturnGateStatus:     returnGateStatus,
		FinalResponseAllowed: finalResponseAllowed,
		RefusesFinalResponse: refusesFinalResponse,
		Reason:               reason,
		Source:               "recommendation_readback",
	}
}

func atlasContinuationContractReason(readyNodes int, exactNextAction, returnGateStatus string) string {
	hasExactNextAction := strings.TrimSpace(exactNextAction) != ""
	switch {
	case readyNodes > 0 && hasExactNextAction:
		return "ready_nodes_or_exact_next_action_remain"
	case readyNodes > 0:
		return "ready_nodes_remain"
	case hasExactNextAction:
		return "exact_next_action_remains"
	default:
		return returnGateStatus
	}
}

type atlasRecommendationLeaseTiming struct {
	StartedAt       string
	CompletedAt     string
	ElapsedMinutes  int
	MinMinutesMet   bool
	LeaseTimeStatus string
}

func buildRecommendationLeaseTiming(wave AtlasRecommendationWave, options AtlasRecommendationReadbackOptions, nodesComplete bool) (atlasRecommendationLeaseTiming, error) {
	minMinutes := wave.EstimatedMinutes
	if wave.Supervisor != nil {
		minMinutes = wave.Supervisor.MinMinutes
	}
	startedAt := strings.TrimSpace(options.StartedAt)
	completedAt := strings.TrimSpace(options.CompletedAt)
	elapsedMinutes := options.ElapsedMinutes
	if elapsedMinutes < 0 {
		return atlasRecommendationLeaseTiming{}, fmt.Errorf("elapsed_minutes must be non-negative")
	}
	var started time.Time
	var completed time.Time
	var hasStarted bool
	var hasCompleted bool
	if startedAt != "" {
		parsed, err := time.Parse(time.RFC3339, startedAt)
		if err != nil {
			return atlasRecommendationLeaseTiming{}, fmt.Errorf("started_at must be RFC3339: %w", err)
		}
		started = parsed
		hasStarted = true
	}
	if completedAt != "" {
		parsed, err := time.Parse(time.RFC3339, completedAt)
		if err != nil {
			return atlasRecommendationLeaseTiming{}, fmt.Errorf("completed_at must be RFC3339: %w", err)
		}
		completed = parsed
		hasCompleted = true
	}
	if hasStarted && hasCompleted && completed.Before(started) {
		return atlasRecommendationLeaseTiming{}, fmt.Errorf("completed_at must be greater than or equal to started_at")
	}
	hasTimingEvidence := elapsedMinutes > 0 ||
		startedAt != "" ||
		completedAt != "" ||
		strings.TrimSpace(options.LeaseTimingMode) != ""
	if elapsedMinutes == 0 && hasStarted && hasCompleted {
		elapsedMinutes = ceilDurationMinutes(completed.Sub(started))
	}
	status := "in_progress_timing_pending"
	minMinutesMet := false
	if minMinutes <= 0 {
		status = "minimum_minutes_not_required"
		minMinutesMet = true
	} else if hasTimingEvidence {
		if elapsedMinutes >= minMinutes {
			status = "minimum_minutes_met"
			minMinutesMet = true
		} else {
			status = "minimum_minutes_unmet"
		}
	} else if nodesComplete {
		status = "lease_timing_missing"
	}
	return atlasRecommendationLeaseTiming{
		StartedAt:       startedAt,
		CompletedAt:     completedAt,
		ElapsedMinutes:  elapsedMinutes,
		MinMinutesMet:   minMinutesMet,
		LeaseTimeStatus: status,
	}, nil
}

func ceilDurationMinutes(duration time.Duration) int {
	if duration <= 0 {
		return 0
	}
	minutes := int(duration / time.Minute)
	if duration%time.Minute != 0 {
		minutes++
	}
	return minutes
}

func ValidateAtlasRecommendationReadback(readback AtlasRecommendationReadback) error {
	var errs []string
	requireContract(&errs, "atlas_recommendation_readback", readback.ContractVersion, AtlasRecommendationReadbackContract)
	requireField(&errs, "mission_id", readback.MissionID)
	requireField(&errs, "target_instance", readback.TargetInstance)
	if !oneOf(readback.Status, "ready", "in_progress", "blocked", "completed") {
		errs = append(errs, "status must be ready, in_progress, blocked, or completed")
	}
	if !digestPattern.MatchString(readback.SourceDigest) {
		errs = append(errs, "source_digest must be sha256 digest")
	}
	if strings.TrimSpace(readback.WaveDigest) != "" && !digestPattern.MatchString(readback.WaveDigest) {
		errs = append(errs, "wave_digest must be sha256 digest")
	}
	if strings.TrimSpace(readback.WorkgraphDigest) != "" && !digestPattern.MatchString(readback.WorkgraphDigest) {
		errs = append(errs, "workgraph_digest must be sha256 digest")
	}
	if readback.TotalNodes < 1 {
		errs = append(errs, "total_nodes must be positive")
	}
	if readback.MinimumNodes < 1 || readback.MinimumNodes > readback.TotalNodes {
		errs = append(errs, "minimum_nodes must be between 1 and total_nodes")
	}
	if readback.CompletedNodes+readback.ReadyNodes+readback.BlockedNodes+readback.FailedNodes != readback.TotalNodes {
		errs = append(errs, "node counts must sum to total_nodes")
	}
	if readback.ElapsedMinutes < 0 {
		errs = append(errs, "elapsed_minutes must be non-negative")
	}
	requireField(&errs, "lease_time_status", readback.LeaseTimeStatus)
	if readback.Supervisor != nil && readback.Supervisor.MinMinutes > 0 {
		if readback.MinMinutesMet && readback.ElapsedMinutes < readback.Supervisor.MinMinutes {
			errs = append(errs, "min_minutes_met requires elapsed_minutes to meet supervisor.min_minutes")
		}
		if readback.FinalResponseAllowed && !readback.MinMinutesMet {
			errs = append(errs, "final_response_allowed requires min_minutes_met")
		}
	}
	if readback.FinalResponseAllowed {
		requireField(&errs, "started_at", readback.StartedAt)
		requireField(&errs, "completed_at", readback.CompletedAt)
		if readback.ElapsedMinutes == 0 {
			errs = append(errs, "final_response_allowed requires elapsed_minutes")
		}
	}
	if readback.ExecutableReadyNodes > readback.ReadyNodes {
		errs = append(errs, "executable_ready_nodes cannot exceed ready_nodes")
	}
	requireField(&errs, "lease_health_status", readback.LeaseHealthStatus)
	requireField(&errs, "checkpoint_freshness_status", readback.CheckpointFreshnessStatus)
	requireField(&errs, "stale_route_decision_status", readback.StaleRouteDecisionStatus)
	requireField(&errs, "early_return_risk_status", readback.EarlyReturnRiskStatus)
	requireField(&errs, "foundry_rollup_status", readback.FoundryRollupStatus)
	for _, key := range []string{"completed", "promoted", "denied", "blocked"} {
		requireField(&errs, "foundry_terminal_status_readback."+key, readback.FoundryTerminalStatusReadback[key])
	}
	if err := validateFoundryTerminalStatusExamples(readback.FoundryTerminalStatusExamples); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validateFoundryDeniedTerminalExamples(readback.FoundryDeniedTerminalExamples); err != nil {
		errs = append(errs, err.Error())
	}
	requireField(&errs, "promoter_readback_status", readback.PromoterReadbackStatus)
	requireField(&errs, "promoter_no_promotion_status", readback.PromoterNoPromotionStatus)
	if err := validatePromoterNoPromotionPlaceholders(readback.PromoterNoPromotionPlaceholders); err != nil {
		errs = append(errs, err.Error())
	}
	requireField(&errs, "command_readback_status", readback.CommandReadbackStatus)
	requireField(&errs, "command_timeline_status", readback.CommandTimelineStatus)
	if err := validateCommandTimelinePlaceholders(readback.CommandTimelinePlaceholders); err != nil {
		errs = append(errs, err.Error())
	}
	requireField(&errs, "public_safety_scan_status", readback.PublicSafetyScanStatus)
	if strings.TrimSpace(readback.ReturnGateStatus) != "" &&
		!oneOf(readback.ReturnGateStatus, "final_response_allowed", "blocked_hard_blocker", "blocked_lease_timing_missing", "blocked_minimum_minutes_unmet", "blocked_ready_nodes_remain", "blocked_no_executable_ready_node") {
		errs = append(errs, "return_gate_status has unsupported value")
	}
	if readback.CheckpointCount < 0 {
		errs = append(errs, "checkpoint_count must be non-negative")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && readback.CheckpointCount != readback.CompletedNodes {
		errs = append(errs, "checkpoint_count must match completed_nodes when return_gate_status is recorded")
	}
	requireField(&errs, "final_response_denial_gate", readback.FinalResponseDenialGate)
	requireField(&errs, "final_response_reason", readback.FinalResponseReason)
	requireField(&errs, "exact_next_action", readback.ExactNextAction)
	if err := validateAtlasContinuationContract(readback); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validateExactNextActionReadback(readback); err != nil {
		errs = append(errs, err.Error())
	}
	if readback.ReadyNodes > 0 && readback.BlockedNodes == 0 && readback.FailedNodes == 0 {
		if readback.FinalResponseAllowed {
			errs = append(errs, "ready nodes require final_response_allowed=false")
		}
		if readback.ReturnGateStatus != "blocked_ready_nodes_remain" {
			errs = append(errs, "ready nodes require return_gate_status=blocked_ready_nodes_remain")
		}
		if readback.FinalResponseReason != "ready nodes or exact next actions remain" {
			errs = append(errs, "ready nodes require final_response_reason=ready nodes or exact next actions remain")
		}
		if readback.ExecutableReadyNodes > 0 && !strings.Contains(readback.ExactNextAction, readback.FirstExecutableNode) {
			errs = append(errs, "ready nodes require exact_next_action to name first_executable_node")
		}
	}
	if readback.FinalResponseAllowed {
		if readback.Status != "completed" {
			errs = append(errs, "final_response_allowed requires status=completed")
		}
		if readback.ReturnGateStatus != "final_response_allowed" {
			errs = append(errs, "final_response_allowed requires return_gate_status=final_response_allowed")
		}
		if readback.FinalResponseReason != "all generated nodes complete and no ready nodes remain" {
			errs = append(errs, "final_response_allowed requires final_response_reason=all generated nodes complete and no ready nodes remain")
		}
		if readback.ExactNextAction != "Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks." {
			errs = append(errs, "final_response_allowed requires final exact_next_action")
		}
		if readback.FinalResponseDenialGate != "allow_final_response" {
			errs = append(errs, "final_response_allowed requires final_response_denial_gate=allow_final_response")
		}
	} else if readback.ReturnGateStatus == "blocked_hard_blocker" {
		if readback.FinalResponseDenialGate != "blocked_hard_blocker" {
			errs = append(errs, "hard blocker requires final_response_denial_gate=blocked_hard_blocker")
		}
	} else if readback.ReadyNodes > 0 || strings.TrimSpace(readback.ExactNextAction) != "" {
		if readback.FinalResponseDenialGate != "deny_ready_nodes_or_exact_next_action_remain" {
			errs = append(errs, "ready nodes or exact next action require final_response_denial_gate=deny_ready_nodes_or_exact_next_action_remain")
		}
	}
	if readback.FinalResponseAllowed && (readback.ReadyNodes > 0 || readback.BlockedNodes > 0 || readback.FailedNodes > 0) {
		errs = append(errs, "final_response_allowed requires no ready, blocked, or failed nodes")
	}
	if len(readback.NodeEvidence) != readback.TotalNodes {
		errs = append(errs, "node_evidence length must match total_nodes")
	}
	if len(readback.FeatureDepthRecommendations) < 10 && readback.TotalNodes >= 10 {
		errs = append(errs, "feature_depth_recommendations must include at least 10 tasks")
	}
	if readback.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if readback.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if readback.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	for i, evidence := range readback.NodeEvidence {
		prefix := fmt.Sprintf("node_evidence[%d]", i)
		requireField(&errs, prefix+".node_id", evidence.NodeID)
		requireField(&errs, prefix+".task_id", evidence.TaskID)
		requireField(&errs, prefix+".status", evidence.Status)
		requireField(&errs, prefix+".node_gate", evidence.NodeGate)
		requireField(&errs, prefix+".candidate_record", evidence.CandidateRecord)
		requireField(&errs, prefix+".rollback_record", evidence.RollbackRecord)
		requireField(&errs, prefix+".implementation_evidence", evidence.ImplementationEvidence)
		requireField(&errs, prefix+".tests", evidence.Tests)
		requireField(&errs, prefix+".verification", evidence.Verification)
		requireField(&errs, prefix+".public_safety_wording", evidence.PublicSafetyWording)
		requireField(&errs, prefix+".promoter_readback", evidence.PromoterReadback)
		requireField(&errs, prefix+".command_readback", evidence.CommandReadback)
		requireList(&errs, prefix+".required_gates", evidence.RequiredGates)
		requireList(&errs, prefix+".verification_commands", evidence.VerificationCommands)
	}
	return joinErrors(errs)
}

func validateAtlasContinuationContract(readback AtlasRecommendationReadback) error {
	contract := readback.ContinuationContract
	if contract.ContractVersion != "ao.atlas.continuation-contract.v0.1" {
		return fmt.Errorf("continuation_contract.contract_version must be ao.atlas.continuation-contract.v0.1")
	}
	if contract.Source != "recommendation_readback" {
		return fmt.Errorf("continuation_contract.source must be recommendation_readback")
	}
	if contract.ReadyNodes != readback.ReadyNodes {
		return fmt.Errorf("continuation_contract.ready_nodes must match ready_nodes")
	}
	if contract.ExactNextAction != readback.ExactNextAction {
		return fmt.Errorf("continuation_contract.exact_next_action must match exact_next_action")
	}
	if contract.ReturnGateStatus != readback.ReturnGateStatus {
		return fmt.Errorf("continuation_contract.return_gate_status must match return_gate_status")
	}
	if contract.FinalResponseAllowed != readback.FinalResponseAllowed {
		return fmt.Errorf("continuation_contract.final_response_allowed must match final_response_allowed")
	}
	if strings.TrimSpace(contract.Reason) == "" {
		return fmt.Errorf("continuation_contract.reason is required")
	}
	if readback.FinalResponseAllowed {
		if contract.Status != "ready_for_final_response" {
			return fmt.Errorf("continuation_contract.status must be ready_for_final_response when final response is allowed")
		}
		if contract.RefusesFinalResponse {
			return fmt.Errorf("continuation_contract.refuses_final_response must be false when final response is allowed")
		}
		return nil
	}
	if readback.ReadyNodes > 0 || strings.TrimSpace(readback.ExactNextAction) != "" {
		if contract.Status != "continuation_required" {
			return fmt.Errorf("continuation_contract.status must be continuation_required while ready nodes or exact next action remain")
		}
		if !contract.RefusesFinalResponse {
			return fmt.Errorf("continuation_contract.refuses_final_response must be true while ready nodes or exact next action remain")
		}
		expectedReason := atlasContinuationContractReason(readback.ReadyNodes, readback.ExactNextAction, readback.ReturnGateStatus)
		if contract.Reason != expectedReason {
			return fmt.Errorf("continuation_contract.reason must be %s while ready nodes or exact next action remain", expectedReason)
		}
	}
	return nil
}

func validateExactNextActionReadback(readback AtlasRecommendationReadback) error {
	action := readback.ExactNextActionReadback
	if strings.TrimSpace(action.Status) == "" {
		return fmt.Errorf("exact_next_action_readback.status is required")
	}
	if action.Action != readback.ExactNextAction {
		return fmt.Errorf("exact_next_action_readback.action must match exact_next_action")
	}
	if action.NextExecutableNode != readback.FirstExecutableNode {
		return fmt.Errorf("exact_next_action_readback.next_executable_node must match first_executable_node")
	}
	if action.ReturnGateStatus != readback.ReturnGateStatus {
		return fmt.Errorf("exact_next_action_readback.return_gate_status must match return_gate_status")
	}
	if action.FinalResponseAllowed != readback.FinalResponseAllowed {
		return fmt.Errorf("exact_next_action_readback.final_response_allowed must match final_response_allowed")
	}
	if action.Source != "recommendation_readback" {
		return fmt.Errorf("exact_next_action_readback.source must be recommendation_readback")
	}
	if readback.FinalResponseAllowed {
		if action.Status != "ready_for_final_response" {
			return fmt.Errorf("exact_next_action_readback.status must be ready_for_final_response")
		}
	} else if action.Status != "continuation_required" {
		return fmt.Errorf("exact_next_action_readback.status must be continuation_required")
	}
	return nil
}

func validatePromoterNoPromotionPlaceholders(placeholders []AtlasPromoterNoPromotionPlaceholder) error {
	required := map[string]bool{
		"promotion_claim":   false,
		"rsi_boundary":      false,
		"authority_advance": false,
	}
	if len(placeholders) < len(required) {
		return fmt.Errorf("promoter_no_promotion_placeholders must include promotion_claim, rsi_boundary, and authority_advance")
	}
	for _, placeholder := range placeholders {
		slot := strings.TrimSpace(placeholder.Slot)
		if _, ok := required[slot]; !ok {
			return fmt.Errorf("promoter_no_promotion_placeholders has unsupported slot %q", placeholder.Slot)
		}
		if required[slot] {
			return fmt.Errorf("promoter_no_promotion_placeholders duplicate slot %q", slot)
		}
		required[slot] = true
		if placeholder.Source != "recommendation_readback" {
			return fmt.Errorf("promoter_no_promotion_placeholders.%s source must be recommendation_readback", slot)
		}
		if placeholder.Status != "pending_promoter_no_promotion" {
			return fmt.Errorf("promoter_no_promotion_placeholders.%s status must be pending_promoter_no_promotion", slot)
		}
		if strings.TrimSpace(placeholder.Summary) == "" {
			return fmt.Errorf("promoter_no_promotion_placeholders.%s summary is required", slot)
		}
		if !placeholder.RequiredBeforeFinalResponse {
			return fmt.Errorf("promoter_no_promotion_placeholders.%s must be required before final response", slot)
		}
	}
	for slot, seen := range required {
		if !seen {
			return fmt.Errorf("promoter_no_promotion_placeholders missing %s", slot)
		}
	}
	return nil
}

func validateCommandTimelinePlaceholders(placeholders []AtlasCommandTimelinePlaceholder) error {
	required := map[string]bool{
		"checkpoint":        false,
		"exact_next_action": false,
		"return_gate":       false,
	}
	if len(placeholders) < len(required) {
		return fmt.Errorf("command_timeline_placeholders must include checkpoint, exact_next_action, and return_gate")
	}
	for _, placeholder := range placeholders {
		slot := strings.TrimSpace(placeholder.Slot)
		if _, ok := required[slot]; !ok {
			return fmt.Errorf("command_timeline_placeholders has unsupported slot %q", placeholder.Slot)
		}
		if required[slot] {
			return fmt.Errorf("command_timeline_placeholders duplicate slot %q", slot)
		}
		required[slot] = true
		if placeholder.Source != "recommendation_readback" {
			return fmt.Errorf("command_timeline_placeholders.%s source must be recommendation_readback", slot)
		}
		if placeholder.Status != "pending_command_timeline" {
			return fmt.Errorf("command_timeline_placeholders.%s status must be pending_command_timeline", slot)
		}
		if strings.TrimSpace(placeholder.Summary) == "" {
			return fmt.Errorf("command_timeline_placeholders.%s summary is required", slot)
		}
		if !placeholder.RequiredBeforeFinalResponse {
			return fmt.Errorf("command_timeline_placeholders.%s must be required before final response", slot)
		}
	}
	for slot, seen := range required {
		if !seen {
			return fmt.Errorf("command_timeline_placeholders missing %s", slot)
		}
	}
	return nil
}

func validateFoundryTerminalStatusExamples(examples []AtlasFoundryTerminalStatusExample) error {
	required := map[string]bool{
		"completed": false,
		"promoted":  false,
		"denied":    false,
		"blocked":   false,
	}
	if len(examples) != len(required) {
		return fmt.Errorf("foundry_terminal_status_examples must include completed, promoted, denied, and blocked examples")
	}
	for _, example := range examples {
		source := strings.TrimSpace(example.SourceStatus)
		if _, ok := required[source]; !ok {
			return fmt.Errorf("foundry_terminal_status_examples has unsupported source_status %q", example.SourceStatus)
		}
		if required[source] {
			return fmt.Errorf("foundry_terminal_status_examples duplicate source_status %q", source)
		}
		required[source] = true
		if strings.TrimSpace(example.NormalizedStatus) == "" {
			return fmt.Errorf("foundry_terminal_status_examples.%s normalized_status is required", source)
		}
		if strings.TrimSpace(example.RequiredReadback) == "" {
			return fmt.Errorf("foundry_terminal_status_examples.%s required_readback is required", source)
		}
		if !example.Terminal {
			return fmt.Errorf("foundry_terminal_status_examples.%s must be terminal", source)
		}
		switch source {
		case "completed":
			if example.NormalizedStatus != "completed" || !example.CanCloseMission {
				return fmt.Errorf("foundry_terminal_status_examples.completed must close as completed")
			}
		case "promoted":
			if example.NormalizedStatus != "completed" || !example.CanCloseMission {
				return fmt.Errorf("foundry_terminal_status_examples.promoted must close as completed")
			}
		case "denied":
			if example.NormalizedStatus != "denied" || !example.CanCloseMission {
				return fmt.Errorf("foundry_terminal_status_examples.denied must close with exact denial evidence")
			}
		case "blocked":
			if example.NormalizedStatus != "blocked" || example.CanCloseMission {
				return fmt.Errorf("foundry_terminal_status_examples.blocked must remain open for repair or resume")
			}
		}
	}
	for source, seen := range required {
		if !seen {
			return fmt.Errorf("foundry_terminal_status_examples missing %s", source)
		}
	}
	return nil
}

func validateFoundryDeniedTerminalExamples(examples []AtlasFoundryDeniedTerminalExample) error {
	required := map[string]bool{
		"missing_node_evidence":          false,
		"missing_stop_gate_evidence":     false,
		"forbidden_surface_or_rsi_claim": false,
	}
	if len(examples) < len(required) {
		return fmt.Errorf("foundry_denied_terminal_examples must include missing node, stop gate, and forbidden surface examples")
	}
	for _, example := range examples {
		reason := strings.TrimSpace(example.DenialReason)
		if _, ok := required[reason]; !ok {
			return fmt.Errorf("foundry_denied_terminal_examples has unsupported denial_reason %q", example.DenialReason)
		}
		if required[reason] {
			return fmt.Errorf("foundry_denied_terminal_examples duplicate denial_reason %q", reason)
		}
		required[reason] = true
		if example.NormalizedStatus != "denied" {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must normalize to denied", reason)
		}
		if !example.Terminal {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must be terminal", reason)
		}
		if !example.CanCloseMission {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must be closable as final denial", reason)
		}
		if !example.RequiresExactMissingEvidence {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must require exact missing evidence", reason)
		}
		if strings.TrimSpace(example.RequiredReadback) == "" {
			return fmt.Errorf("foundry_denied_terminal_examples.%s required_readback is required", reason)
		}
		if !example.RSIRemainsDenied {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must keep RSI denied", reason)
		}
		if example.AuthorityAdvanceClaimed {
			return fmt.Errorf("foundry_denied_terminal_examples.%s must not claim authority advance", reason)
		}
	}
	for reason, seen := range required {
		if !seen {
			return fmt.Errorf("foundry_denied_terminal_examples missing %s", reason)
		}
	}
	return nil
}

func ValidateAtlasRecommendationLeaseStart(leaseStart AtlasRecommendationLeaseStart) error {
	var errs []string
	if leaseStart.Schema != "ao.atlas.recommendation-lease-start.v0.1" {
		errs = append(errs, "schema must be ao.atlas.recommendation-lease-start.v0.1")
	}
	if leaseStart.Status != "started" {
		errs = append(errs, "status must be started")
	}
	requireField(&errs, "mission_id", leaseStart.MissionID)
	requireField(&errs, "target_instance", leaseStart.TargetInstance)
	requireField(&errs, "started_at", leaseStart.StartedAt)
	if strings.TrimSpace(leaseStart.StartedAt) != "" {
		if _, err := time.Parse(time.RFC3339, leaseStart.StartedAt); err != nil {
			errs = append(errs, "started_at must be RFC3339")
		}
	}
	if leaseStart.MinMinutes < 1 {
		errs = append(errs, "min_minutes must be positive")
	}
	if leaseStart.MaxMinutes < leaseStart.MinMinutes {
		errs = append(errs, "max_minutes must be greater than or equal to min_minutes")
	}
	if leaseStart.ContinueIfFastTarget < 1 {
		errs = append(errs, "continue_if_fast_target must be positive")
	}
	if !digestPattern.MatchString(leaseStart.WaveDigest) {
		errs = append(errs, "wave_digest must be sha256 digest")
	}
	if !digestPattern.MatchString(leaseStart.WorkgraphDigest) {
		errs = append(errs, "workgraph_digest must be sha256 digest")
	}
	if leaseStart.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false for lease start marker")
	}
	requireField(&errs, "final_response_reason", leaseStart.FinalResponseReason)
	if leaseStart.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if leaseStart.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if leaseStart.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if leaseStart.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	if leaseStart.CallsProviders {
		errs = append(errs, "calls_providers must be false")
	}
	if leaseStart.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationCheckpointReadback(readback AtlasRecommendationReadback) AtlasRecommendationCheckpointReadback {
	minMinutes := readback.ElapsedMinutes
	maxMinutes := readback.ElapsedMinutes
	if readback.Supervisor != nil {
		minMinutes = readback.Supervisor.MinMinutes
		maxMinutes = readback.Supervisor.MaxMinutes
	}
	status := "fresh"
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		status = "blocked"
	}
	return AtlasRecommendationCheckpointReadback{
		Schema:                    "ao.atlas.recommendation-checkpoint-readback.v0.1",
		Status:                    status,
		MissionID:                 readback.MissionID,
		EvidenceRoot:              readback.EvidenceRoot,
		StartedAt:                 readback.StartedAt,
		CompletedAt:               readback.CompletedAt,
		ElapsedMinutes:            readback.ElapsedMinutes,
		MinMinutes:                minMinutes,
		MaxMinutes:                maxMinutes,
		MinMinutesMet:             readback.MinMinutesMet,
		LeaseTimeStatus:           readback.LeaseTimeStatus,
		LeaseHealthStatus:         readback.LeaseHealthStatus,
		CheckpointFreshnessStatus: "elapsed_minutes_recorded_after_node_checkpoint",
		CompletedNodes:            readback.CompletedNodes,
		ReadyNodes:                readback.ReadyNodes,
		BlockedNodes:              readback.BlockedNodes,
		FailedNodes:               readback.FailedNodes,
		TotalNodes:                readback.TotalNodes,
		FirstExecutableNode:       readback.FirstExecutableNode,
		FinalResponseAllowed:      readback.FinalResponseAllowed,
		FinalResponseReason:       readback.FinalResponseReason,
		ExactNextAction:           readback.ExactNextAction,
		SchedulesWork:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		ClaimsAuthorityAdvance:    false,
	}
}

func ValidateAtlasRecommendationCheckpointReadback(checkpoint AtlasRecommendationCheckpointReadback) error {
	var errs []string
	if checkpoint.Schema != "ao.atlas.recommendation-checkpoint-readback.v0.1" {
		errs = append(errs, "schema must be ao.atlas.recommendation-checkpoint-readback.v0.1")
	}
	if !oneOf(checkpoint.Status, "fresh", "blocked") {
		errs = append(errs, "status must be fresh or blocked")
	}
	requireField(&errs, "mission_id", checkpoint.MissionID)
	if checkpoint.CompletedNodes+checkpoint.ReadyNodes+checkpoint.BlockedNodes+checkpoint.FailedNodes != checkpoint.TotalNodes {
		errs = append(errs, "node counts must sum to total_nodes")
	}
	if checkpoint.ElapsedMinutes < 0 {
		errs = append(errs, "elapsed_minutes must be non-negative")
	}
	requireField(&errs, "lease_time_status", checkpoint.LeaseTimeStatus)
	requireField(&errs, "lease_health_status", checkpoint.LeaseHealthStatus)
	requireField(&errs, "checkpoint_freshness_status", checkpoint.CheckpointFreshnessStatus)
	requireField(&errs, "final_response_reason", checkpoint.FinalResponseReason)
	requireField(&errs, "exact_next_action", checkpoint.ExactNextAction)
	if checkpoint.FinalResponseAllowed && !checkpoint.MinMinutesMet {
		errs = append(errs, "final_response_allowed requires min_minutes_met")
	}
	if checkpoint.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if checkpoint.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if checkpoint.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if checkpoint.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationCommandReadback(readback AtlasRecommendationReadback) AtlasRecommendationCommandReadback {
	minMinutes := readback.ElapsedMinutes
	if readback.Supervisor != nil {
		minMinutes = readback.Supervisor.MinMinutes
	}
	nodeStatus := "nodes_in_progress"
	if readback.CompletedNodes == readback.TotalNodes && readback.ReadyNodes == 0 && readback.BlockedNodes == 0 && readback.FailedNodes == 0 {
		nodeStatus = "all_nodes_complete"
	}
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		nodeStatus = "blocked_or_failed_nodes_present"
	}
	compactTimeline := fmt.Sprintf("%d/%d recommendation nodes complete; elapsed_minutes=%d; lease_time_status=%s; final_response_allowed=%t", readback.CompletedNodes, readback.TotalNodes, readback.ElapsedMinutes, readback.LeaseTimeStatus, readback.FinalResponseAllowed)
	return AtlasRecommendationCommandReadback{
		Schema:                     "ao.atlas.recommendation-command-readback.v0.1",
		Status:                     readback.Status,
		MissionID:                  readback.MissionID,
		EvidenceRoot:               readback.EvidenceRoot,
		CompletedNodes:             readback.CompletedNodes,
		ReadyNodes:                 readback.ReadyNodes,
		BlockedNodes:               readback.BlockedNodes,
		FailedNodes:                readback.FailedNodes,
		TotalNodes:                 readback.TotalNodes,
		StartedAt:                  readback.StartedAt,
		CompletedAt:                readback.CompletedAt,
		ElapsedMinutes:             readback.ElapsedMinutes,
		MinMinutes:                 minMinutes,
		MinMinutesMet:              readback.MinMinutesMet,
		LeaseTimeStatus:            readback.LeaseTimeStatus,
		LeaseHealthStatus:          readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
		NodeCompletionStatus:       nodeStatus,
		ReturnGateStatus:           readback.ReturnGateStatus,
		CheckpointCount:            readback.CheckpointCount,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		FinalResponseReason:        readback.FinalResponseReason,
		ExactNextAction:            readback.ExactNextAction,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		CompactTimeline:            compactTimeline,
		CommandTimelineBinding: AtlasRecommendationCommandTimelineBinding{
			Summary:                    compactTimeline,
			FirstExecutableNode:        readback.FirstExecutableNode,
			ExactNextAction:            readback.ExactNextAction,
			ContinuationContractReason: readback.ContinuationContract.Reason,
			ReturnGateStatus:           readback.ReturnGateStatus,
			NodeCompletionStatus:       nodeStatus,
			LeaseTimeStatus:            readback.LeaseTimeStatus,
			LeaseHealthStatus:          readback.LeaseHealthStatus,
			CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
			CheckpointCount:            readback.CheckpointCount,
			CompletedNodes:             readback.CompletedNodes,
			ReadyNodes:                 readback.ReadyNodes,
			TotalNodes:                 readback.TotalNodes,
			ElapsedMinutes:             readback.ElapsedMinutes,
			MinMinutes:                 minMinutes,
			MinMinutesMet:              readback.MinMinutesMet,
			FinalResponseAllowed:       readback.FinalResponseAllowed,
		},
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
	}
}

func BuildAtlasRecommendationPromoterReadback(readback AtlasRecommendationReadback) AtlasRecommendationPromoterReadback {
	reason := "Recommendation wave records no mutation authority promotion; RSI remains denied."
	if readback.FinalResponseAllowed {
		reason = "Recommendation wave may close its readback lease, but it does not promote mutation authority; RSI remains denied."
	}
	return AtlasRecommendationPromoterReadback{
		Schema:                     "ao.atlas.recommendation-promoter-readback.v0.1",
		Status:                     "no_promotion",
		MissionID:                  readback.MissionID,
		EvidenceRoot:               readback.EvidenceRoot,
		PromotionClaimed:           false,
		RSIRemainsDenied:           true,
		NoPromotionSummary:         "No mutation authority promotion claimed; RSI remains denied.",
		NextDeniedClass:            "RSI",
		Reason:                     reason,
		ElapsedMinutes:             readback.ElapsedMinutes,
		MinMinutesMet:              readback.MinMinutesMet,
		LeaseHealthStatus:          readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
	}
}

func BuildAtlasRecommendationFoundryRollup(readback AtlasRecommendationReadback) AtlasRecommendationFoundryRollup {
	nodeStatus := "nodes_in_progress"
	status := "in_progress"
	if readback.CompletedNodes == readback.TotalNodes && readback.ReadyNodes == 0 && readback.BlockedNodes == 0 && readback.FailedNodes == 0 {
		nodeStatus = "all_nodes_complete"
		status = "nodes_complete_lease_pending"
	}
	if readback.FinalResponseAllowed {
		status = "completed"
	}
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		nodeStatus = "blocked_or_failed_nodes_present"
		status = "blocked"
	}
	return AtlasRecommendationFoundryRollup{
		Schema:                     "ao.atlas.recommendation-foundry-rollup.v0.1",
		Status:                     status,
		MissionID:                  readback.MissionID,
		EvidenceRoot:               readback.EvidenceRoot,
		CompletedNodes:             readback.CompletedNodes,
		ReadyNodes:                 readback.ReadyNodes,
		BlockedNodes:               readback.BlockedNodes,
		FailedNodes:                readback.FailedNodes,
		TotalNodes:                 readback.TotalNodes,
		NodeCompletionStatus:       nodeStatus,
		LeaseCompletionStatus:      readback.LeaseTimeStatus,
		LeaseHealthStatus:          readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:  readback.CheckpointFreshnessStatus,
		ReturnGateStatus:           readback.ReturnGateStatus,
		CheckpointCount:            readback.CheckpointCount,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		ExactNextAction:            readback.ExactNextAction,
		ContinuationContractReason: readback.ContinuationContract.Reason,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
	}
}

func ValidateAtlasRecommendationClosureArtifacts(readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup) error {
	var errs []string
	if command.Schema != "ao.atlas.recommendation-command-readback.v0.1" {
		errs = append(errs, "command readback schema must be ao.atlas.recommendation-command-readback.v0.1")
	}
	if promoter.Schema != "ao.atlas.recommendation-promoter-readback.v0.1" {
		errs = append(errs, "promoter readback schema must be ao.atlas.recommendation-promoter-readback.v0.1")
	}
	if foundry.Schema != "ao.atlas.recommendation-foundry-rollup.v0.1" {
		errs = append(errs, "foundry rollup schema must be ao.atlas.recommendation-foundry-rollup.v0.1")
	}
	if command.MissionID != readback.MissionID {
		errs = append(errs, "command readback mission_id disagrees")
	}
	if command.Status != readback.Status {
		errs = append(errs, "command readback status disagrees")
	}
	if promoter.MissionID != readback.MissionID {
		errs = append(errs, "promoter readback mission_id disagrees")
	}
	if foundry.MissionID != readback.MissionID {
		errs = append(errs, "foundry rollup mission_id disagrees")
	}
	if command.CompletedNodes != readback.CompletedNodes || command.ReadyNodes != readback.ReadyNodes || command.TotalNodes != readback.TotalNodes {
		errs = append(errs, "command readback node counts disagree")
	}
	if foundry.CompletedNodes != readback.CompletedNodes || foundry.ReadyNodes != readback.ReadyNodes || foundry.TotalNodes != readback.TotalNodes {
		errs = append(errs, "foundry rollup node counts disagree")
	}
	if foundry.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "foundry rollup continuation_contract_reason disagrees")
	}
	if promoter.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "promoter readback continuation_contract_reason disagrees")
	}
	if command.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "command readback final_response_allowed disagrees")
	}
	if foundry.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "foundry rollup final_response_allowed disagrees")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && command.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "command readback return_gate_status disagrees")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && foundry.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "foundry rollup return_gate_status disagrees")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && command.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "command readback checkpoint_count disagrees")
	}
	if command.CommandTimelineBinding.Summary != command.CompactTimeline {
		errs = append(errs, "command timeline binding summary disagrees")
	}
	if command.CommandTimelineBinding.FirstExecutableNode != readback.FirstExecutableNode {
		errs = append(errs, "command timeline binding first_executable_node disagrees")
	}
	if command.CommandTimelineBinding.ExactNextAction != readback.ExactNextAction {
		errs = append(errs, "command timeline binding exact_next_action disagrees")
	}
	if command.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "command readback continuation_contract_reason disagrees")
	}
	if command.CommandTimelineBinding.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "command timeline binding continuation_contract_reason disagrees")
	}
	if command.CommandTimelineBinding.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "command timeline binding return_gate_status disagrees")
	}
	if command.CommandTimelineBinding.NodeCompletionStatus != command.NodeCompletionStatus {
		errs = append(errs, "command timeline binding node_completion_status disagrees")
	}
	if command.CommandTimelineBinding.LeaseTimeStatus != readback.LeaseTimeStatus {
		errs = append(errs, "command timeline binding lease_time_status disagrees")
	}
	if command.CommandTimelineBinding.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "command timeline binding checkpoint_count disagrees")
	}
	if command.CommandTimelineBinding.CompletedNodes != readback.CompletedNodes ||
		command.CommandTimelineBinding.ReadyNodes != readback.ReadyNodes ||
		command.CommandTimelineBinding.TotalNodes != readback.TotalNodes {
		errs = append(errs, "command timeline binding node counts disagree")
	}
	if command.CommandTimelineBinding.ElapsedMinutes != readback.ElapsedMinutes ||
		command.CommandTimelineBinding.MinMinutes != command.MinMinutes ||
		command.CommandTimelineBinding.MinMinutesMet != readback.MinMinutesMet {
		errs = append(errs, "command timeline binding lease timing disagrees")
	}
	if command.CommandTimelineBinding.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "command timeline binding final_response_allowed disagrees")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && foundry.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "foundry rollup checkpoint_count disagrees")
	}
	if command.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "command readback lease_health_status disagrees")
	}
	if command.CommandTimelineBinding.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "command timeline binding lease_health_status disagrees")
	}
	if promoter.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "promoter readback lease_health_status disagrees")
	}
	if foundry.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "foundry rollup lease_health_status disagrees")
	}
	if command.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "command readback checkpoint_freshness_status disagrees")
	}
	if command.CommandTimelineBinding.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "command timeline binding checkpoint_freshness_status disagrees")
	}
	if promoter.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "promoter readback checkpoint_freshness_status disagrees")
	}
	if foundry.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "foundry rollup checkpoint_freshness_status disagrees")
	}
	if foundry.Status == "completed" && !readback.FinalResponseAllowed {
		errs = append(errs, "foundry rollup completed while recommendation final response is denied")
	}
	if promoter.PromotionClaimed {
		errs = append(errs, "promoter readback must not claim promotion for recommendation wave")
	}
	if !promoter.RSIRemainsDenied {
		errs = append(errs, "promoter readback must keep RSI denied")
	}
	if promoter.NoPromotionSummary != "No mutation authority promotion claimed; RSI remains denied." {
		errs = append(errs, "promoter readback must include no-promotion summary")
	}
	if promoter.NextDeniedClass != "RSI" {
		errs = append(errs, "promoter readback next_denied_class must be RSI")
	}
	if command.SchedulesWork || command.ExecutesWork || command.ApprovesWork || command.ClaimsAuthorityAdvance {
		errs = append(errs, "command readback must not schedule, execute, approve, or claim authority advance")
	}
	if promoter.SchedulesWork || promoter.ExecutesWork || promoter.ApprovesWork || promoter.ClaimsAuthorityAdvance {
		errs = append(errs, "promoter readback must not schedule, execute, approve, or claim authority advance")
	}
	if foundry.SchedulesWork || foundry.ExecutesWork || foundry.ApprovesWork || foundry.ClaimsAuthorityAdvance {
		errs = append(errs, "foundry rollup must not schedule, execute, approve, or claim authority advance")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationReconciliationPacket(readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup) AtlasRecommendationReconciliationPacket {
	artifactsAgree := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry) == nil
	continuationReasonAgreement := command.ContinuationContractReason == readback.ContinuationContract.Reason &&
		promoter.ContinuationContractReason == readback.ContinuationContract.Reason &&
		foundry.ContinuationContractReason == readback.ContinuationContract.Reason
	status := "continuation_required"
	if !artifactsAgree {
		status = "blocked_stale_artifact"
	} else if readback.FinalResponseAllowed {
		status = "ready"
	}
	return AtlasRecommendationReconciliationPacket{
		Schema:                       "ao.atlas.recommendation-reconciliation-packet.v0.1",
		Status:                       status,
		MissionID:                    readback.MissionID,
		EvidenceRoot:                 readback.EvidenceRoot,
		FinalStateReconciliation:     buildRecommendationFinalStateReconciliation(readback, command, promoter, foundry, status),
		CompletedNodes:               readback.CompletedNodes,
		ReadyNodes:                   readback.ReadyNodes,
		BlockedNodes:                 readback.BlockedNodes,
		FailedNodes:                  readback.FailedNodes,
		TotalNodes:                   readback.TotalNodes,
		CheckpointCount:              readback.CheckpointCount,
		ReturnGateStatus:             readback.ReturnGateStatus,
		LeaseTimeStatus:              readback.LeaseTimeStatus,
		LeaseHealthStatus:            readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:    readback.CheckpointFreshnessStatus,
		StaleRouteDecisionStatus:     readback.StaleRouteDecisionStatus,
		FinalResponseAllowed:         readback.FinalResponseAllowed,
		FinalResponseReason:          readback.FinalResponseReason,
		ExactNextAction:              readback.ExactNextAction,
		ContinuationContractReason:   readback.ContinuationContract.Reason,
		CommandReturnGateStatus:      command.ReturnGateStatus,
		CommandContinuationReason:    command.ContinuationContractReason,
		CommandFinalResponseAllowed:  command.FinalResponseAllowed,
		PromoterStatus:               promoter.Status,
		PromoterContinuationReason:   promoter.ContinuationContractReason,
		PromotionClaimed:             promoter.PromotionClaimed,
		RSIRemainsDenied:             promoter.RSIRemainsDenied,
		FoundryStatus:                foundry.Status,
		FoundryReturnGateStatus:      foundry.ReturnGateStatus,
		FoundryContinuationReason:    foundry.ContinuationContractReason,
		FoundryNodeCompletionStatus:  foundry.NodeCompletionStatus,
		FoundryLeaseCompletionStatus: foundry.LeaseCompletionStatus,
		FoundryFinalResponseAllowed:  foundry.FinalResponseAllowed,
		ContinuationReasonAgreement:  continuationReasonAgreement,
		ArtifactsAgree:               artifactsAgree,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		ClaimsAuthorityAdvance:       false,
	}
}

func buildRecommendationFinalStateReconciliation(readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup, status string) AtlasFinalStateReconciliation {
	continuationReasonAgreement := command.ContinuationContractReason == readback.ContinuationContract.Reason &&
		promoter.ContinuationContractReason == readback.ContinuationContract.Reason &&
		foundry.ContinuationContractReason == readback.ContinuationContract.Reason
	return AtlasFinalStateReconciliation{
		ContractVersion:       "ao.atlas.final-state-reconciliation.v0.1",
		Status:                status,
		WorkgraphStatus:       readback.Status,
		FoundryRollupStatus:   foundry.Status,
		PromoterVerdictStatus: promoter.Status,
		CommandReadbackStatus: command.Status,
		ExactNextAction:       readback.ExactNextAction,
		ContinuationReason:    readback.ContinuationContract.Reason,
		ContinuationAgreement: continuationReasonAgreement,
		SchedulesWork:         false,
		ExecutesWork:          false,
		ApprovesWork:          false,
	}
}

func ValidateAtlasRecommendationReconciliationPacket(readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup, packet AtlasRecommendationReconciliationPacket) error {
	var errs []string
	if packet.Schema != "ao.atlas.recommendation-reconciliation-packet.v0.1" {
		errs = append(errs, "schema must be ao.atlas.recommendation-reconciliation-packet.v0.1")
	}
	if packet.MissionID != readback.MissionID {
		errs = append(errs, "reconciliation mission_id disagrees")
	}
	validateRecommendationFinalStateReconciliation(&errs, readback, command, promoter, foundry, packet)
	if packet.CompletedNodes != readback.CompletedNodes || packet.ReadyNodes != readback.ReadyNodes || packet.TotalNodes != readback.TotalNodes {
		errs = append(errs, "reconciliation node counts disagree")
	}
	if packet.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "reconciliation checkpoint_count disagrees")
	}
	if packet.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "reconciliation return_gate_status disagrees")
	}
	if packet.LeaseTimeStatus != readback.LeaseTimeStatus {
		errs = append(errs, "reconciliation lease_time_status disagrees")
	}
	if packet.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "reconciliation lease_health_status disagrees")
	}
	if packet.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "reconciliation checkpoint_freshness_status disagrees")
	}
	if packet.StaleRouteDecisionStatus != readback.StaleRouteDecisionStatus {
		errs = append(errs, "reconciliation stale_route_decision_status disagrees")
	}
	if packet.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "reconciliation final_response_allowed disagrees")
	}
	if packet.ExactNextAction != readback.ExactNextAction {
		errs = append(errs, "reconciliation exact_next_action disagrees")
	}
	if packet.ContinuationContractReason != readback.ContinuationContract.Reason {
		errs = append(errs, "reconciliation continuation_contract_reason disagrees")
	}
	if packet.CommandReturnGateStatus != command.ReturnGateStatus || packet.CommandFinalResponseAllowed != command.FinalResponseAllowed {
		errs = append(errs, "reconciliation command fields disagree")
	}
	if packet.CommandContinuationReason != command.ContinuationContractReason {
		errs = append(errs, "reconciliation command_continuation_contract_reason disagrees")
	}
	if packet.PromoterStatus != promoter.Status || packet.PromotionClaimed != promoter.PromotionClaimed || packet.RSIRemainsDenied != promoter.RSIRemainsDenied {
		errs = append(errs, "reconciliation promoter fields disagree")
	}
	if packet.PromoterContinuationReason != promoter.ContinuationContractReason {
		errs = append(errs, "reconciliation promoter_continuation_contract_reason disagrees")
	}
	if packet.FoundryStatus != foundry.Status ||
		packet.FoundryReturnGateStatus != foundry.ReturnGateStatus ||
		packet.FoundryNodeCompletionStatus != foundry.NodeCompletionStatus ||
		packet.FoundryLeaseCompletionStatus != foundry.LeaseCompletionStatus ||
		packet.FoundryFinalResponseAllowed != foundry.FinalResponseAllowed {
		errs = append(errs, "reconciliation foundry fields disagree")
	}
	if packet.FoundryContinuationReason != foundry.ContinuationContractReason {
		errs = append(errs, "reconciliation foundry_continuation_contract_reason disagrees")
	}
	expectedContinuationReasonAgreement := command.ContinuationContractReason == readback.ContinuationContract.Reason &&
		promoter.ContinuationContractReason == readback.ContinuationContract.Reason &&
		foundry.ContinuationContractReason == readback.ContinuationContract.Reason
	if packet.ContinuationReasonAgreement != expectedContinuationReasonAgreement {
		errs = append(errs, "reconciliation continuation_reason_agreement disagrees")
	}
	closureErr := ValidateAtlasRecommendationClosureArtifacts(readback, command, promoter, foundry)
	if closureErr == nil && !packet.ArtifactsAgree {
		errs = append(errs, "reconciliation artifacts_agree must be true when closure artifacts agree")
	}
	if closureErr == nil && !packet.ContinuationReasonAgreement {
		errs = append(errs, "reconciliation continuation_reason_agreement must be true when closure artifacts agree")
	}
	if closureErr == nil && packet.Status == "blocked_stale_artifact" {
		errs = append(errs, "reconciliation status blocked_stale_artifact requires stale closure artifacts")
	}
	if closureErr != nil && packet.ArtifactsAgree {
		errs = append(errs, "reconciliation artifacts_agree must be false when closure artifacts disagree")
	}
	if closureErr != nil && packet.Status != "blocked_stale_artifact" {
		errs = append(errs, "reconciliation status must be blocked_stale_artifact when closure artifacts disagree")
	}
	if packet.Status == "ready" && !packet.FinalResponseAllowed {
		errs = append(errs, "reconciliation ready status requires final_response_allowed")
	}
	if packet.SchedulesWork || packet.ExecutesWork || packet.ApprovesWork || packet.ClaimsAuthorityAdvance {
		errs = append(errs, "reconciliation packet must not schedule, execute, approve, or claim authority advance")
	}
	return joinErrors(errs)
}

func validateRecommendationFinalStateReconciliation(errs *[]string, readback AtlasRecommendationReadback, command AtlasRecommendationCommandReadback, promoter AtlasRecommendationPromoterReadback, foundry AtlasRecommendationFoundryRollup, packet AtlasRecommendationReconciliationPacket) {
	finalState := packet.FinalStateReconciliation
	if finalState.ContractVersion != "ao.atlas.final-state-reconciliation.v0.1" {
		*errs = append(*errs, "final_state_reconciliation.contract_version must be ao.atlas.final-state-reconciliation.v0.1")
	}
	if finalState.Status != packet.Status {
		*errs = append(*errs, "final_state_reconciliation.status must match reconciliation status")
	}
	if finalState.WorkgraphStatus != readback.Status {
		*errs = append(*errs, "final_state_reconciliation.workgraph_status disagrees")
	}
	if finalState.FoundryRollupStatus != foundry.Status {
		*errs = append(*errs, "final_state_reconciliation.foundry_rollup_status disagrees")
	}
	if finalState.PromoterVerdictStatus != promoter.Status {
		*errs = append(*errs, "final_state_reconciliation.promoter_verdict_status disagrees")
	}
	if finalState.CommandReadbackStatus != command.Status {
		*errs = append(*errs, "final_state_reconciliation.command_readback_status disagrees")
	}
	if finalState.ExactNextAction != readback.ExactNextAction {
		*errs = append(*errs, "final_state_reconciliation.exact_next_action disagrees")
	}
	if finalState.ContinuationReason != readback.ContinuationContract.Reason {
		*errs = append(*errs, "final_state_reconciliation.continuation_contract_reason disagrees")
	}
	expectedContinuationReasonAgreement := command.ContinuationContractReason == readback.ContinuationContract.Reason &&
		promoter.ContinuationContractReason == readback.ContinuationContract.Reason &&
		foundry.ContinuationContractReason == readback.ContinuationContract.Reason
	if finalState.ContinuationAgreement != expectedContinuationReasonAgreement {
		*errs = append(*errs, "final_state_reconciliation.continuation_reason_agreement disagrees")
	}
	if finalState.SchedulesWork || finalState.ExecutesWork || finalState.ApprovesWork {
		*errs = append(*errs, "final_state_reconciliation must not schedule, execute, or approve work")
	}
	if (readback.ReadyNodes > 0 || strings.TrimSpace(readback.ExactNextAction) != "") &&
		!readback.FinalResponseAllowed &&
		!oneOf(finalState.Status, "continuation_required", "blocked_stale_artifact") {
		*errs = append(*errs, "final_state_reconciliation must require continuation while ready nodes or exact next action remain")
	}
}

func ValidateAtlasRecommendationExecutionReadback(execution AtlasRecommendationExecutionReadback, readback AtlasRecommendationReadback) error {
	var errs []string
	if execution.Schema != "ao.atlas.long-recommendation-wave-execution.v0.3" {
		errs = append(errs, "schema must be ao.atlas.long-recommendation-wave-execution.v0.3")
	}
	requireField(&errs, "status", execution.Status)
	if execution.MissionID != readback.MissionID {
		errs = append(errs, "mission_id must match recommendation readback")
	}
	if execution.TotalRecommendationNodes != readback.TotalNodes {
		errs = append(errs, "total_recommendation_nodes must match recommendation readback total_nodes")
	}
	if execution.CompletedRecommendationNodes != readback.CompletedNodes {
		errs = append(errs, "completed_recommendation_nodes must match recommendation readback completed_nodes")
	}
	if execution.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "lease_health_status must match recommendation readback")
	}
	if execution.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "checkpoint_freshness_status must match recommendation readback")
	}
	if execution.GeneratedWorkgraph.TotalNodes != readback.TotalNodes {
		errs = append(errs, "generated_workgraph.total_nodes must match recommendation readback total_nodes")
	}
	if execution.GeneratedWorkgraph.ReadyNodes != readback.ReadyNodes {
		errs = append(errs, "generated_workgraph.ready_nodes must match recommendation readback ready_nodes")
	}
	if execution.GeneratedWorkgraph.ExecutableReadyNodes != readback.ExecutableReadyNodes {
		errs = append(errs, "generated_workgraph.executable_ready_nodes must match recommendation readback executable_ready_nodes")
	}
	if execution.GeneratedWorkgraph.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "generated_workgraph.final_response_allowed must match recommendation readback final_response_allowed")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && execution.GeneratedWorkgraph.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "generated_workgraph.return_gate_status must match recommendation readback return_gate_status")
	}
	if strings.TrimSpace(readback.ReturnGateStatus) != "" && execution.GeneratedWorkgraph.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "generated_workgraph.checkpoint_count must match recommendation readback checkpoint_count")
	}
	if execution.GeneratedWorkgraph.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "generated_workgraph.lease_health_status must match recommendation readback")
	}
	if execution.GeneratedWorkgraph.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "generated_workgraph.checkpoint_freshness_status must match recommendation readback")
	}
	requireField(&errs, "foundry_run_link_readiness_summary.status", execution.FoundryRunLinkReadinessSummary.Status)
	requireField(&errs, "foundry_run_link_readiness_summary.summary", execution.FoundryRunLinkReadinessSummary.Summary)
	if execution.FoundryRunLinkReadinessSummary.CompletedRunLinks != readback.CompletedNodes {
		errs = append(errs, "foundry run-link readiness completed_run_links must match recommendation readback completed_nodes")
	}
	if execution.FoundryRunLinkReadinessSummary.RequiredRunLinks != readback.TotalNodes {
		errs = append(errs, "foundry run-link readiness required_run_links must match recommendation readback total_nodes")
	}
	if execution.FoundryRunLinkReadinessSummary.MissingRunLinks != readback.TotalNodes-readback.CompletedNodes {
		errs = append(errs, "foundry run-link readiness missing_run_links must match remaining nodes")
	}
	if execution.FoundryRunLinkReadinessSummary.ReadyNodes != readback.ReadyNodes {
		errs = append(errs, "foundry run-link readiness ready_nodes must match recommendation readback ready_nodes")
	}
	if execution.FoundryRunLinkReadinessSummary.NextExecutableNode != readback.FirstExecutableNode {
		errs = append(errs, "foundry run-link readiness next_executable_node must match recommendation readback first_executable_node")
	}
	if execution.FoundryRunLinkReadinessSummary.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "foundry run-link readiness checkpoint_count must match recommendation readback checkpoint_count")
	}
	if execution.FoundryRunLinkReadinessSummary.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "foundry run-link readiness final_response_allowed must match recommendation readback final_response_allowed")
	}
	if execution.FoundryRunLinkReadinessSummary.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "foundry run-link readiness lease_health_status must match recommendation readback")
	}
	if execution.FoundryRunLinkReadinessSummary.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "foundry run-link readiness checkpoint_freshness_status must match recommendation readback")
	}
	if sourceDigest, ok := sourceArtifactDigest(execution.SourceArtifacts, "foundry_run_link_readiness_summary"); !ok {
		errs = append(errs, "source_artifacts must include foundry_run_link_readiness_summary")
	} else if sourceDigest != digestValue(execution.FoundryRunLinkReadinessSummary) {
		errs = append(errs, "foundry_run_link_readiness_summary source artifact digest disagrees")
	}
	if execution.Status == "completed" && !readback.FinalResponseAllowed {
		errs = append(errs, "status completed requires recommendation readback final_response_allowed")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationWorkgraphReadinessPacket(readback AtlasRecommendationReadback, options AtlasRecommendationWorkgraphReadinessPacketOptions) (AtlasRecommendationWorkgraphReadinessPacket, error) {
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasRecommendationWorkgraphReadinessPacket{}, err
	}
	waveDigest := readback.WaveDigest
	if strings.TrimSpace(options.WavePath) != "" {
		digest, err := digestFile(options.WavePath)
		if err != nil {
			return AtlasRecommendationWorkgraphReadinessPacket{}, err
		}
		waveDigest = digest
	}
	if strings.TrimSpace(waveDigest) == "" {
		waveDigest = digestValue(readback.MissionID + readback.SourceDigest)
	}
	workgraphDigest := readback.WorkgraphDigest
	if strings.TrimSpace(options.WorkgraphPath) != "" {
		digest, err := digestFile(options.WorkgraphPath)
		if err != nil {
			return AtlasRecommendationWorkgraphReadinessPacket{}, err
		}
		workgraphDigest = digest
	}
	if strings.TrimSpace(workgraphDigest) == "" {
		workgraphDigest = digestValue(readback.TotalNodes)
	}
	readbackDigest := digestValue(readback)
	if strings.TrimSpace(options.ReadbackPath) != "" {
		digest, err := digestFile(options.ReadbackPath)
		if err != nil {
			return AtlasRecommendationWorkgraphReadinessPacket{}, err
		}
		readbackDigest = digest
	}
	continueIfFastTarget := readback.TotalNodes
	if readback.Supervisor != nil && readback.Supervisor.ContinueIfFastTarget > 0 {
		continueIfFastTarget = readback.Supervisor.ContinueIfFastTarget
	}
	status := "continuation_required"
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		status = "blocked"
	}
	if readback.FinalResponseAllowed {
		status = "ready_for_final_response"
	}
	packet := AtlasRecommendationWorkgraphReadinessPacket{
		Schema:                          "ao.atlas.recommendation-workgraph-readiness-packet.v0.1",
		Status:                          status,
		MissionID:                       readback.MissionID,
		TargetInstance:                  readback.TargetInstance,
		EvidenceRoot:                    readback.EvidenceRoot,
		WaveDigest:                      waveDigest,
		WorkgraphDigest:                 workgraphDigest,
		ReadbackDigest:                  readbackDigest,
		TotalNodes:                      readback.TotalNodes,
		MinimumNodes:                    readback.MinimumNodes,
		NodeBudget:                      readback.TotalNodes,
		ContinueIfFastTarget:            continueIfFastTarget,
		CompletedNodes:                  readback.CompletedNodes,
		ReadyNodes:                      readback.ReadyNodes,
		BlockedNodes:                    readback.BlockedNodes,
		FailedNodes:                     readback.FailedNodes,
		ExecutableReadyNodes:            readback.ExecutableReadyNodes,
		FirstExecutableNode:             readback.FirstExecutableNode,
		LeaseHealthStatus:               readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:       readback.CheckpointFreshnessStatus,
		ReturnGateStatus:                readback.ReturnGateStatus,
		CheckpointCount:                 readback.CheckpointCount,
		EarlyReturnRiskStatus:           readback.EarlyReturnRiskStatus,
		ContinuationBudgetStatus:        recommendationContinuationBudgetStatus(readback, continueIfFastTarget),
		FinalResponseAllowed:            readback.FinalResponseAllowed,
		FinalResponseReason:             readback.FinalResponseReason,
		ExactNextAction:                 readback.ExactNextAction,
		OneExecutableMutationNodeActive: readback.ExecutableReadyNodes == 1,
		RefusesFinalResponse:            !readback.FinalResponseAllowed,
		SchedulesWork:                   false,
		ExecutesWork:                    false,
		ApprovesWork:                    false,
		ClaimsAuthorityAdvance:          false,
		RSIRemainsDenied:                true,
	}
	if err := ValidateAtlasRecommendationWorkgraphReadinessPacket(packet, readback); err != nil {
		return AtlasRecommendationWorkgraphReadinessPacket{}, err
	}
	return packet, nil
}

func recommendationContinuationBudgetStatus(readback AtlasRecommendationReadback, continueIfFastTarget int) string {
	if readback.FinalResponseAllowed {
		return "all_generated_nodes_complete"
	}
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		return "hard_blocker_requires_repair"
	}
	if readback.ReadyNodes > 0 && readback.CompletedNodes < readback.MinimumNodes {
		return "minimum_nodes_unmet_continue_to_40_node_budget"
	}
	if readback.ReadyNodes > 0 && readback.CompletedNodes >= readback.MinimumNodes && readback.CompletedNodes < continueIfFastTarget {
		return "minimum_met_continue_if_fast_budget_open"
	}
	if readback.ReadyNodes == 0 && !readback.MinMinutesMet {
		return "node_budget_complete_waiting_for_lease_evidence"
	}
	return "continuation_required"
}

func ValidateAtlasRecommendationWorkgraphReadinessPacket(packet AtlasRecommendationWorkgraphReadinessPacket, readback AtlasRecommendationReadback) error {
	var errs []string
	if packet.Schema != "ao.atlas.recommendation-workgraph-readiness-packet.v0.1" {
		errs = append(errs, "schema must be ao.atlas.recommendation-workgraph-readiness-packet.v0.1")
	}
	if !oneOf(packet.Status, "continuation_required", "ready_for_final_response", "blocked") {
		errs = append(errs, "status must be continuation_required, ready_for_final_response, or blocked")
	}
	if packet.MissionID != readback.MissionID {
		errs = append(errs, "mission_id must match recommendation readback")
	}
	if packet.TargetInstance != readback.TargetInstance {
		errs = append(errs, "target_instance must match recommendation readback")
	}
	for field, digest := range map[string]string{
		"wave_digest":      packet.WaveDigest,
		"workgraph_digest": packet.WorkgraphDigest,
		"readback_digest":  packet.ReadbackDigest,
	} {
		if !digestPattern.MatchString(digest) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if packet.TotalNodes != readback.TotalNodes {
		errs = append(errs, "total_nodes must match recommendation readback")
	}
	if packet.MinimumNodes != readback.MinimumNodes {
		errs = append(errs, "minimum_nodes must match recommendation readback")
	}
	if packet.NodeBudget != readback.TotalNodes {
		errs = append(errs, "node_budget must match recommendation readback total_nodes")
	}
	expectedContinueTarget := readback.TotalNodes
	if readback.Supervisor != nil && readback.Supervisor.ContinueIfFastTarget > 0 {
		expectedContinueTarget = readback.Supervisor.ContinueIfFastTarget
	}
	if packet.ContinueIfFastTarget != expectedContinueTarget {
		errs = append(errs, "continue_if_fast_target must match supervisor continue_if_fast_target")
	}
	if packet.CompletedNodes != readback.CompletedNodes {
		errs = append(errs, "completed_nodes must match recommendation readback")
	}
	if packet.ReadyNodes != readback.ReadyNodes {
		errs = append(errs, "ready_nodes must match recommendation readback")
	}
	if packet.BlockedNodes != readback.BlockedNodes {
		errs = append(errs, "blocked_nodes must match recommendation readback")
	}
	if packet.FailedNodes != readback.FailedNodes {
		errs = append(errs, "failed_nodes must match recommendation readback")
	}
	if packet.ExecutableReadyNodes != readback.ExecutableReadyNodes {
		errs = append(errs, "executable_ready_nodes must match recommendation readback")
	}
	if packet.FirstExecutableNode != readback.FirstExecutableNode {
		errs = append(errs, "first_executable_node must match recommendation readback")
	}
	if packet.LeaseHealthStatus != readback.LeaseHealthStatus {
		errs = append(errs, "lease_health_status must match recommendation readback")
	}
	if packet.CheckpointFreshnessStatus != readback.CheckpointFreshnessStatus {
		errs = append(errs, "checkpoint_freshness_status must match recommendation readback")
	}
	if packet.ReturnGateStatus != readback.ReturnGateStatus {
		errs = append(errs, "return_gate_status must match recommendation readback")
	}
	if packet.CheckpointCount != readback.CheckpointCount {
		errs = append(errs, "checkpoint_count must match recommendation readback")
	}
	if packet.EarlyReturnRiskStatus != readback.EarlyReturnRiskStatus {
		errs = append(errs, "early_return_risk_status must match recommendation readback")
	}
	expectedBudgetStatus := recommendationContinuationBudgetStatus(readback, expectedContinueTarget)
	if packet.ContinuationBudgetStatus != expectedBudgetStatus {
		errs = append(errs, "continuation_budget_status must match recommendation readback")
	}
	if packet.FinalResponseAllowed != readback.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must match recommendation readback")
	}
	if packet.FinalResponseReason != readback.FinalResponseReason {
		errs = append(errs, "final_response_reason must match recommendation readback")
	}
	if packet.ExactNextAction != readback.ExactNextAction {
		errs = append(errs, "exact_next_action must match recommendation readback")
	}
	if readback.ReadyNodes > 0 {
		if packet.ReturnGateStatus != "blocked_ready_nodes_remain" {
			errs = append(errs, "ready nodes require return_gate_status=blocked_ready_nodes_remain")
		}
		if !packet.OneExecutableMutationNodeActive {
			errs = append(errs, "ready nodes require one_executable_mutation_node_active=true")
		}
		if readback.FirstExecutableNode != "" && !strings.Contains(packet.ExactNextAction, readback.FirstExecutableNode) {
			errs = append(errs, "ready nodes require exact_next_action to name first_executable_node")
		}
		if packet.FinalResponseAllowed {
			errs = append(errs, "ready nodes require final_response_allowed=false")
		}
	}
	if readback.FinalResponseAllowed {
		if packet.Status != "ready_for_final_response" {
			errs = append(errs, "final_response_allowed requires status=ready_for_final_response")
		}
		if packet.RefusesFinalResponse {
			errs = append(errs, "final_response_allowed requires refuses_final_response=false")
		}
	} else {
		if packet.Status == "ready_for_final_response" {
			errs = append(errs, "status ready_for_final_response requires final_response_allowed=true")
		}
		if !packet.RefusesFinalResponse {
			errs = append(errs, "final_response_allowed=false requires refuses_final_response=true")
		}
	}
	if packet.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if packet.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if packet.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if packet.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !packet.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func BuildAtlasRecommendationExecutionReadback(readback AtlasRecommendationReadback) AtlasRecommendationExecutionReadback {
	status := "implementation_wave_completed_generated_workgraph_ready"
	if readback.Status == "in_progress" || readback.CompletedNodes > 0 {
		status = "in_progress"
	}
	if readback.Status == "blocked" {
		status = "blocked"
	}
	if readback.FinalResponseAllowed {
		status = "completed"
	}
	readinessStatus := "pending_first_run_link"
	if readback.CompletedNodes > 0 {
		readinessStatus = "partial_run_links_recorded"
	}
	if readback.CompletedNodes == readback.TotalNodes && readback.ReadyNodes == 0 && readback.BlockedNodes == 0 && readback.FailedNodes == 0 {
		readinessStatus = "all_required_run_links_recorded"
	}
	if readback.BlockedNodes > 0 || readback.FailedNodes > 0 {
		readinessStatus = "blocked_or_failed_run_links_need_repair"
	}
	runLinkSummary := AtlasRecommendationFoundryRunLinkReadinessSummary{
		Status:                    readinessStatus,
		Summary:                   fmt.Sprintf("%d/%d Foundry run-links recorded; ready_nodes=%d; next_executable_node=%s", readback.CompletedNodes, readback.TotalNodes, readback.ReadyNodes, readback.FirstExecutableNode),
		CompletedRunLinks:         readback.CompletedNodes,
		RequiredRunLinks:          readback.TotalNodes,
		MissingRunLinks:           readback.TotalNodes - readback.CompletedNodes,
		ReadyNodes:                readback.ReadyNodes,
		NextExecutableNode:        readback.FirstExecutableNode,
		LeaseHealthStatus:         readback.LeaseHealthStatus,
		CheckpointFreshnessStatus: readback.CheckpointFreshnessStatus,
		CheckpointCount:           readback.CheckpointCount,
		FinalResponseAllowed:      readback.FinalResponseAllowed,
	}
	return AtlasRecommendationExecutionReadback{
		Schema:                       "ao.atlas.long-recommendation-wave-execution.v0.3",
		Status:                       status,
		MissionID:                    readback.MissionID,
		EvidenceRoot:                 readback.EvidenceRoot,
		LeaseHealthStatus:            readback.LeaseHealthStatus,
		CheckpointFreshnessStatus:    readback.CheckpointFreshnessStatus,
		CompletedRecommendationNodes: readback.CompletedNodes,
		TotalRecommendationNodes:     readback.TotalNodes,
		GeneratedWorkgraph: AtlasRecommendationGeneratedWorkgraphReadback{
			TotalNodes:                readback.TotalNodes,
			ReadyNodes:                readback.ReadyNodes,
			ExecutableReadyNodes:      readback.ExecutableReadyNodes,
			FirstExecutableNode:       readback.FirstExecutableNode,
			LeaseHealthStatus:         readback.LeaseHealthStatus,
			CheckpointFreshnessStatus: readback.CheckpointFreshnessStatus,
			ReturnGateStatus:          readback.ReturnGateStatus,
			CheckpointCount:           readback.CheckpointCount,
			FinalResponseAllowed:      readback.FinalResponseAllowed,
			FinalResponseReason:       readback.FinalResponseReason,
		},
		FoundryRunLinkReadinessSummary: runLinkSummary,
		SourceArtifacts: []SourceRef{
			{Ref: "foundry_run_link_readiness_summary", Digest: digestValue(runLinkSummary)},
		},
	}
}

func sourceArtifactDigest(sources []SourceRef, ref string) (string, bool) {
	for _, source := range sources {
		if source.Ref == ref {
			return source.Digest, true
		}
	}
	return "", false
}

func CompleteAtlasRecommendationNodeWithRunLink(wave AtlasRecommendationWave, workgraph Workgraph, link RunLink, options AtlasRecommendationCompleteNodeOptions) (Workgraph, string, error) {
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		return Workgraph{}, "", err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return Workgraph{}, "", err
	}
	if err := ValidateRunLink(link); err != nil {
		return Workgraph{}, "", err
	}
	if wave.TargetInstance != workgraph.TargetInstance {
		return Workgraph{}, "", fmt.Errorf("target_instance mismatch between recommendation wave and workgraph")
	}
	if len(wave.Tasks) != len(workgraph.Nodes) {
		return Workgraph{}, "", fmt.Errorf("workgraph node count must match recommendation tasks")
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return Workgraph{}, "", err
	}
	executable, ok := state.NextReadyNode()
	if !ok {
		return Workgraph{}, "", fmt.Errorf("no executable recommendation node remains")
	}
	expectedNodeID := strings.TrimSpace(options.ExpectedNodeID)
	if expectedNodeID != "" && executable.ID != expectedNodeID {
		return Workgraph{}, "", fmt.Errorf("expected executable node %s, got %s", expectedNodeID, executable.ID)
	}
	if link.Status != "completed" {
		return Workgraph{}, "", fmt.Errorf("run-link status must be completed")
	}
	if link.TaskID != executable.FactoryTask.ID {
		return Workgraph{}, "", fmt.Errorf("run-link task_id must match executable node %s task %s", executable.ID, executable.FactoryTask.ID)
	}
	if err := validateRecommendationRunLinkEvidence(executable.FactoryTask, link, options.EvidenceRoot); err != nil {
		return Workgraph{}, "", err
	}
	return CompleteWorkgraph(workgraph, link)
}

func validateRecommendationRunLinkEvidence(task FactoryTask, link RunLink, evidenceRoot string) error {
	for _, key := range requiredRecommendationRunLinkEvidence(task) {
		path := strings.TrimSpace(link.Evidence[key])
		if path == "" {
			return fmt.Errorf("missing evidence %s", key)
		}
		if strings.TrimSpace(evidenceRoot) == "" {
			continue
		}
		clean := filepath.Clean(path)
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("evidence %s must stay inside evidence root", key)
		}
		if _, err := os.Stat(filepath.Join(evidenceRoot, clean)); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("evidence %s path does not exist: %s", key, filepath.ToSlash(clean))
			}
			return err
		}
	}
	return nil
}

func requiredRecommendationRunLinkEvidence(task FactoryTask) []string {
	seen := map[string]bool{}
	keys := []string{}
	add := func(key string) {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		keys = append(keys, key)
	}
	for _, gate := range task.RequiredGates {
		add(gate)
	}
	for _, key := range []string{
		"implementation_evidence",
		"foundry_import",
		"checkpoint_bundle",
	} {
		add(key)
	}
	return keys
}

func recommendationNodeEvidence(workgraph Workgraph) []AtlasRecommendationNodeEvidence {
	evidence := make([]AtlasRecommendationNodeEvidence, 0, len(workgraph.Nodes))
	for _, node := range workgraph.Nodes {
		evidence = append(evidence, AtlasRecommendationNodeEvidence{
			NodeID:                 node.ID,
			TaskID:                 node.FactoryTask.ID,
			Status:                 node.Status,
			NodeGate:               evidenceStatus(node.FactoryTask.RequiredGates, "node_gate"),
			CandidateRecord:        evidenceStatus(node.FactoryTask.RequiredGates, "candidate_record"),
			RollbackRecord:         evidenceStatus(node.FactoryTask.RequiredGates, "rollback_record"),
			ImplementationEvidence: "recorded",
			Tests:                  evidenceStatus(node.FactoryTask.RequiredGates, "tests"),
			Verification:           evidenceStatus(node.FactoryTask.RequiredGates, "verification"),
			PublicSafetyWording:    evidenceStatus(node.FactoryTask.RequiredGates, "sentinel_public_safety"),
			PromoterReadback:       evidenceStatus(node.FactoryTask.RequiredGates, "promoter_no_promotion"),
			CommandReadback:        evidenceStatus(node.FactoryTask.RequiredGates, "command_readback"),
			RequiredGates:          append([]string(nil), node.FactoryTask.RequiredGates...),
			VerificationCommands:   append([]string(nil), node.FactoryTask.Verification...),
		})
	}
	return evidence
}

func evidenceStatus(values []string, want string) string {
	for _, value := range values {
		if value == want {
			return "recorded"
		}
	}
	return "missing"
}

func featureDepthRecommendationReadback(tasks []AtlasRecommendationTask, limit int) []string {
	if limit <= 0 || limit > len(tasks) {
		limit = len(tasks)
	}
	items := make([]string, 0, limit)
	for _, task := range tasks[:limit] {
		items = append(items, task.ID+": "+task.Task)
	}
	return items
}

func atlasOwnedRecommendationTasks(tasks []AOMissionFeatureDepthTask, limit int) []AOMissionFeatureDepthTask {
	selected := []AOMissionFeatureDepthTask{}
	for _, task := range tasks {
		if task.Owner != "ao-atlas" {
			continue
		}
		selected = append(selected, task)
		if limit > 0 && len(selected) >= limit {
			break
		}
	}
	return selected
}

type AtlasRecommendationResumePromptOptions struct {
	EvidenceRoot   string
	LeaseStartPath string
	WorkgraphPath  string
	ReadbackPath   string
}

func BuildAtlasRecommendationResumePrompt(readback AtlasRecommendationReadback, options AtlasRecommendationResumePromptOptions) string {
	evidenceRoot := strings.TrimSpace(options.EvidenceRoot)
	if evidenceRoot == "" {
		evidenceRoot = readback.EvidenceRoot
	}
	leaseStartPath := filepath.ToSlash(strings.TrimSpace(options.LeaseStartPath))
	workgraphPath := filepath.ToSlash(strings.TrimSpace(options.WorkgraphPath))
	readbackPath := filepath.ToSlash(strings.TrimSpace(options.ReadbackPath))
	minMinutes := readback.ElapsedMinutes
	if readback.Supervisor != nil {
		minMinutes = readback.Supervisor.MinMinutes
	}
	nextNode := readback.FirstExecutableNode
	if strings.TrimSpace(nextNode) == "" {
		nextNode = "none"
	}

	var b strings.Builder
	b.WriteString("You are AO Atlas, continuing the AO Atlas long-run recommendation wave.\n\n")
	b.WriteString("Do not ask the operator for permission. Do not reset the lease clock. Load and preserve:\n\n")
	if evidenceRoot != "" {
		b.WriteString(fmt.Sprintf("- Evidence root: `%s`\n", filepath.ToSlash(evidenceRoot)))
	}
	if leaseStartPath != "" {
		b.WriteString(fmt.Sprintf("- Lease start: `%s`\n", leaseStartPath))
	}
	if workgraphPath != "" {
		b.WriteString(fmt.Sprintf("- Current workgraph: `%s`\n", workgraphPath))
	}
	if readbackPath != "" {
		b.WriteString(fmt.Sprintf("- Current readback: `%s`\n", readbackPath))
	}
	b.WriteString("\nCurrent status:\n")
	b.WriteString(fmt.Sprintf("- Completed nodes: %d / %d\n", readback.CompletedNodes, readback.TotalNodes))
	b.WriteString(fmt.Sprintf("- Ready nodes: %d\n", readback.ReadyNodes))
	b.WriteString(fmt.Sprintf("- Elapsed minutes at latest checkpoint: %d\n", readback.ElapsedMinutes))
	b.WriteString(fmt.Sprintf("- Minimum minutes: %d\n", minMinutes))
	b.WriteString(fmt.Sprintf("- `min_minutes_met=%t`\n", readback.MinMinutesMet))
	b.WriteString(fmt.Sprintf("- `final_response_allowed=%t`\n", readback.FinalResponseAllowed))
	b.WriteString(fmt.Sprintf("- Return gate: `%s`\n", readback.ReturnGateStatus))
	b.WriteString(fmt.Sprintf("- Early-return risk: `%s`\n", readback.EarlyReturnRiskStatus))
	b.WriteString(fmt.Sprintf("- Checkpoint count: %d\n", readback.CheckpointCount))
	b.WriteString(fmt.Sprintf("- Next executable node: `%s`\n\n", nextNode))
	b.WriteString("Goal:\n")
	b.WriteString("Continue the useful 2-3 hour Atlas-owned hardening wave. Execute exactly one bounded node at a time, preserving the original `started_at` from `lease-start.json`, until all ready work is handled or a true hard blocker remains after safe repair attempts.\n\n")
	b.WriteString("Exact next action:\n")
	b.WriteString(fmt.Sprintf("- %s\n\n", readback.ExactNextAction))
	b.WriteString("Blocked-node continuation:\n")
	b.WriteString("- If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair.\n\n")
	b.WriteString("Safety boundaries:\n")
	b.WriteString("- No provider calls.\n")
	b.WriteString("- No credential or token inspection.\n")
	b.WriteString("- No direct main mutation.\n")
	b.WriteString("- No release, deploy, publish, upload, or tag.\n")
	b.WriteString("- No dependency updates unless separately authorized.\n")
	b.WriteString("- No auth, policy, or config widening.\n")
	b.WriteString("- No hidden instruction mutation.\n")
	b.WriteString("- No broad RSI claim.\n")
	b.WriteString("- RSI remains denied.\n")
	b.WriteString("- Keep exactly one executable mutation node active at a time.\n\n")
	b.WriteString("Verification:\n")
	b.WriteString("- `go test ./... -count=1`\n")
	b.WriteString("- `go vet ./...`\n")
	b.WriteString("- `go build ./cmd/atlas`\n")
	b.WriteString("- `scripts/production-readiness.sh`\n")
	b.WriteString("- `scripts/atlas-foundry-roundtrip-smoke.sh`\n")
	b.WriteString("- Public-safety wording scan over changed docs and evidence.\n\n")
	b.WriteString("Final response is allowed only when the authoritative recommendation readback has `final_response_allowed=true`, the execution readback agrees, Command and Foundry summaries agree, Promoter records no promotion, verification passes, the repo is clean and synced, and no ready nodes or exact next actions remain.\n")
	b.WriteString("If `ready_nodes > 0` or `exact_next_action` is non-empty, do not produce a final response.\n")
	return b.String()
}

func buildAtlasRecommendationPrompt(wave AtlasRecommendationWave) string {
	var b strings.Builder
	minMinutes := wave.EstimatedMinutes
	maxMinutes := wave.EstimatedMinutes
	continueTarget := wave.NodeBudget
	returnOnlyWhen := fmt.Sprintf("all generated nodes complete, at least %d nodes complete, or a true hard blocker remains", wave.MinimumTasks)
	checkpointPolicy := "after each node or timed interval"
	if wave.Supervisor != nil {
		minMinutes = wave.Supervisor.MinMinutes
		maxMinutes = wave.Supervisor.MaxMinutes
		continueTarget = wave.Supervisor.ContinueIfFastTarget
		returnOnlyWhen = wave.Supervisor.ReturnOnlyWhen
		checkpointPolicy = wave.Supervisor.CheckpointPolicy
	}
	b.WriteString("You are AO Atlas, coordinating AO Foundry, AO Promoter, AO Command, and AO Architecture for a long-run Atlas-owned correction wave.\n\n")
	b.WriteString("Do not ask the operator for permission. Do not stop after one repo, one PR, one CI pass, one Foundry import, one route decision, one evidence artifact, or one short batch. Continue until the workgraph is completed with evidence, at least the lease minimum is met, or a true hard blocker remains after all safe repair, repack, and support work has been attempted.\n\n")
	b.WriteString("Current state:\n")
	b.WriteString(fmt.Sprintf("- Mission: %s.\n", wave.MissionID))
	b.WriteString(fmt.Sprintf("- Target instance: %s.\n", wave.TargetInstance))
	b.WriteString(fmt.Sprintf("- Generated Atlas-owned nodes: %d.\n", wave.TotalTasks))
	b.WriteString(fmt.Sprintf("- Lease minimum: %d nodes, %d to %d minutes.\n", wave.MinimumTasks, minMinutes, maxMinutes))
	b.WriteString(fmt.Sprintf("- Continue-if-fast target: %d nodes.\n", continueTarget))
	b.WriteString(fmt.Sprintf("- Final response allowed: %t, because %s.\n", wave.FinalResponseAllowed, wave.FinalResponseReason))
	b.WriteString(fmt.Sprintf("- Source digest: %s.\n\n", wave.SourceDigest))
	b.WriteString("Problem:\n")
	b.WriteString("- Recent AO Atlas/Mission recommendation prompts returned after short batches instead of sustaining 2-3 hour workgraphs.\n")
	b.WriteString("- Double the previous short batch when explicitly requested, and otherwise use the v0.2 2-3 hour supervisor default.\n")
	b.WriteString("- This continuation must behave like a long-run supervisor: Atlas owns sequencing, Foundry owns bounded implementation nodes, and Blueprint is used only for genuinely new requirements or authorization.\n\n")
	b.WriteString("Goal:\n")
	b.WriteString(fmt.Sprintf("- Target 2-3 hours and complete a durable AO Atlas long-run wave for %s.\n", wave.MissionID))
	b.WriteString(fmt.Sprintf("- Execute at least %d bounded Atlas nodes from the generated workgraph.\n", wave.MinimumTasks))
	b.WriteString(fmt.Sprintf("- Complete at least %d bounded implementation/evidence nodes before final response unless a true hard blocker remains.\n", wave.MinimumTasks))
	b.WriteString(fmt.Sprintf("- If the first %d nodes finish quickly and no blocker remains, continue through the %d-node continue-if-fast target.\n\n", wave.MinimumTasks, continueTarget))
	b.WriteString(fmt.Sprintf("Return only after all generated nodes complete, at least %d bounded Atlas nodes complete, or a true hard blocker remains after safe repair attempts.\n\n", wave.MinimumTasks))
	b.WriteString("Minimum work budget:\n")
	b.WriteString(fmt.Sprintf("- min_nodes: %d\n", wave.MinimumTasks))
	b.WriteString(fmt.Sprintf("- min_minutes: %d\n", minMinutes))
	b.WriteString(fmt.Sprintf("- max_minutes: %d\n", maxMinutes))
	b.WriteString(fmt.Sprintf("- max_iterations: %d\n", continueTarget))
	b.WriteString(fmt.Sprintf("- return_only_when: %s\n", returnOnlyWhen))
	b.WriteString(fmt.Sprintf("- checkpoint_policy: %s\n\n", checkpointPolicy))
	b.WriteString("Safety boundaries:\n")
	b.WriteString("- Keep exactly one executable mutation node active at a time.\n")
	b.WriteString("- Preserve no provider calls, no credential inspection, no direct main mutation, no release/deploy/publish/upload/tag, no dependency updates, no auth/policy/config widening, and no broad RSI claim.\n")
	b.WriteString("- RSI remains denied unless separate governed evidence proves otherwise.\n")
	b.WriteString("- Use existing repo auth only for normal PR, CI, and merge if available without exposing credentials.\n\n")
	b.WriteString("Required work:\n")
	for _, task := range wave.Tasks {
		b.WriteString(fmt.Sprintf("%s. %s\n", strings.TrimPrefix(task.ID, "next-"), task.Task))
	}
	b.WriteString("\nPer-node requirements:\n")
	b.WriteString("- Generate or validate node gate, candidate record, rollback record, implementation evidence, tests, verification command output, Sentinel/public-safety wording evidence where applicable, Promoter/no-promotion or promotion-readiness evidence where applicable, and Command/readback evidence where applicable.\n")
	b.WriteString("- Emit a Foundry import for exactly one active node at a time, execute the node, verify locally, record run-link evidence, complete the node in Atlas, evaluate the next stop gate, and continue.\n\n")
	b.WriteString("Regression tests:\n")
	b.WriteString("- Prove the recommendation wave defaults to at least 30 nodes and 120 minutes.\n")
	b.WriteString("- Prove the continue-if-fast target generates 40 bounded Atlas-owned tasks.\n")
	b.WriteString("- Prove mixed-owner default waves are rejected with exact readback.\n")
	b.WriteString("- Prove final response remains denied while ready nodes or exact next actions remain.\n\n")
	b.WriteString("Verification:\n")
	b.WriteString("- go test ./... -count=1\n")
	b.WriteString("- go vet ./...\n")
	b.WriteString("- go build ./cmd/atlas\n")
	b.WriteString("- scripts/production-readiness.sh\n")
	b.WriteString("- scripts/atlas-foundry-roundtrip-smoke.sh\n")
	b.WriteString("- Public-safety wording scan over changed docs and readbacks.\n\n")
	b.WriteString("Final response only after completion or true hard blocker:\n")
	b.WriteString("- Include `early_return_risk_status` in continuation prompts and treat any blocked risk status as final-response denial evidence.\n")
	b.WriteString("- If ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response.\n")
	b.WriteString("- If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair.\n")
	b.WriteString("- completed nodes / total nodes\n")
	b.WriteString("- list of node statuses\n")
	b.WriteString("- merged PRs by repo or local commits if remote lifecycle is blocked\n")
	b.WriteString("- evidence roots\n")
	b.WriteString("- final AO Atlas long-run supervisor status\n")
	b.WriteString("- Foundry rollup\n")
	b.WriteString("- Command readback\n")
	b.WriteString("- Feature Depth Recommendations, at least 10 tasks\n")
	b.WriteString("- verification results\n")
	b.WriteString("- clean/synced repo status\n")
	b.WriteString("- exact next action\n")
	return b.String()
}
