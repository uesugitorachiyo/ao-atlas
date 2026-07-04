package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
				RequiredEvidence:  []string{"source_digest:" + wave.SourceDigest, "source_recommendation:" + item.ID},
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
	return os.WriteFile(filepath.Join(outDir, "next-recommended-prompt.md"), []byte(result.Prompt), 0o644)
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
