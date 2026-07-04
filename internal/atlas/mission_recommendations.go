package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AtlasRecommendationWaveOptions struct {
	RecommendationsPath string
	TargetInstance      string
	MinTasks            int
	NodeBudget          int
	EstimatedMinutes    int
}

type AtlasRecommendationWaveResult struct {
	Wave      AtlasRecommendationWave
	Workgraph Workgraph
	Prompt    string
}

func BuildAtlasRecommendationWave(options AtlasRecommendationWaveOptions) (AtlasRecommendationWaveResult, error) {
	minTasks := options.MinTasks
	if minTasks <= 0 {
		minTasks = 20
	}
	nodeBudget := options.NodeBudget
	if nodeBudget <= 0 {
		nodeBudget = minTasks
	}
	estimatedMinutes := options.EstimatedMinutes
	if estimatedMinutes <= 0 {
		estimatedMinutes = 90
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
	if len(selected) < minTasks {
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
		Tasks:            tasks,
		SafeToExecute:    false,
		SchedulesWork:    false,
		ExecutesWork:     false,
		ApprovesWork:     false,
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
	b.WriteString("You are AO Atlas, continuing the AO Atlas long recommendation wave.\n\n")
	b.WriteString("Double the previous short batch: target about 90 minutes of useful Atlas work.\n")
	b.WriteString(fmt.Sprintf("Execute at least %d bounded Atlas nodes from mission %s before final response.\n", wave.MinimumTasks, wave.MissionID))
	b.WriteString("Return only after all generated nodes complete, at least 20 bounded Atlas nodes complete, or a true hard blocker remains after safe repair attempts.\n\n")
	b.WriteString("Required behavior:\n")
	b.WriteString("- Keep exactly one executable mutation node active at a time.\n")
	b.WriteString("- Preserve no provider calls, no credential inspection, no direct main mutation, no release/deploy/publish/upload/tag, no dependency updates, no auth/policy/config widening, and no broad RSI claim.\n")
	b.WriteString("- For each node, record gate, candidate, rollback, tests, verification, public-safety scan, promoter/no-promotion readback, and command/readback evidence.\n")
	b.WriteString("- Run go test ./... -count=1, go vet ./..., go build ./cmd/atlas, scripts/production-readiness.sh, and scripts/atlas-foundry-roundtrip-smoke.sh before completion.\n\n")
	b.WriteString("Recommended nodes:\n")
	for _, task := range wave.Tasks {
		b.WriteString(fmt.Sprintf("- %s: %s\n", task.ID, task.Task))
	}
	return b.String()
}
