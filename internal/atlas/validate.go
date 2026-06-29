package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	InstanceContract               = "ao.atlas.stack-instance.v0.1"
	AtlasRegistryContract          = "ao.atlas.foundry-registry.v0.1"
	InstanceDoctorContract         = "ao.atlas.instance-doctor.v0.1"
	IntakeContract                 = "ao.atlas.intake.v0.1"
	MissionStatusContract          = "ao.atlas.mission-status.v0.1"
	WorkgraphContract              = "ao.atlas.workgraph.v0.1"
	WorkgraphRepairPlanContract    = "ao.atlas.workgraph-repair-plan.v0.1"
	FactoryTaskContract            = "ao.atlas.factory-task.v0.1"
	FactoryMaterializationContract = "ao.atlas.factory-materialization.v0.1"
	ContextPackContract            = "ao.atlas.context-pack.v0.1"
	FoundryHandoffContract         = "ao.atlas.foundry-handoff.v0.1"
	FoundryImportContract          = "ao.atlas.foundry-import.v0.1"
	RunLinkContract                = "ao.atlas.run-link.v0.1"
	BlueprintRequestContract       = "ao.atlas.blueprint-request.v0.1"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var driveAbsPattern = regexp.MustCompile(`^[A-Za-z]:[\\/]`)

func LoadJSON[T any](path string) (T, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	return value, nil
}

func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ValidateInstance(instance Instance) error {
	var errs []string
	requireContract(&errs, "instance", instance.ContractVersion, InstanceContract)
	requireField(&errs, "id", instance.ID)
	requireField(&errs, "state_root", instance.StateRoot)
	requireField(&errs, "toolchain_root", instance.ToolchainRoot)
	if len(instance.Roots) == 0 {
		errs = append(errs, "roots must not be empty")
	}
	requiredRoots := []string{"mission", "workgraph", "context", "evidence", "worktree"}
	for _, key := range requiredRoots {
		requireField(&errs, "roots."+key, instance.Roots[key])
	}
	checkPublicPathMap(&errs, instance.Roots)
	checkPublicPath(&errs, "state_root", instance.StateRoot, false)
	checkPublicPath(&errs, "toolchain_root", instance.ToolchainRoot, false)
	return joinErrors(errs)
}

func ValidateAtlasRegistry(registry AtlasRegistry) error {
	var errs []string
	requireContract(&errs, "atlas_registry", registry.ContractVersion, AtlasRegistryContract)
	requireField(&errs, "instance_id", registry.InstanceID)
	requireField(&errs, "toolchain_root", registry.ToolchainRoot)
	if len(registry.Roots) == 0 {
		errs = append(errs, "roots must not be empty")
	}
	checkPublicPathMap(&errs, registry.Roots)
	checkPublicPath(&errs, "toolchain_root", registry.ToolchainRoot, false)
	if registry.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if registry.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	return joinErrors(errs)
}

func ValidateInstanceDoctorReport(report InstanceDoctorReport) error {
	var errs []string
	requireContract(&errs, "instance_doctor", report.ContractVersion, InstanceDoctorContract)
	requireField(&errs, "instance_id", report.InstanceID)
	if report.Status != "ready" {
		errs = append(errs, "status must be ready")
	}
	required := []string{"instance", "registry_parity", "ignored_local_state", "worktree_hygiene"}
	for _, key := range required {
		if report.Checks[key] != "passed" {
			errs = append(errs, "checks."+key+" must be passed")
		}
	}
	if report.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if report.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	return joinErrors(errs)
}

func BuildInstanceDoctorReport(instance Instance, registry AtlasRegistry) (InstanceDoctorReport, error) {
	if err := ValidateInstance(instance); err != nil {
		return InstanceDoctorReport{}, err
	}
	if err := ValidateAtlasRegistry(registry); err != nil {
		return InstanceDoctorReport{}, err
	}
	if registry.InstanceID != instance.ID {
		return InstanceDoctorReport{}, fmt.Errorf("registry instance_id must match instance id")
	}
	if registry.ToolchainRoot != instance.ToolchainRoot {
		return InstanceDoctorReport{}, fmt.Errorf("registry toolchain_root must match instance toolchain_root")
	}
	for key, value := range instance.Roots {
		if registry.Roots[key] != value {
			return InstanceDoctorReport{}, fmt.Errorf("registry roots.%s must match instance roots.%s", key, key)
		}
	}
	if !strings.HasPrefix(instance.StateRoot, ".atlas-local/") && !strings.HasPrefix(instance.StateRoot, ".atlas-state/") {
		return InstanceDoctorReport{}, fmt.Errorf("state_root must be under ignored Atlas local state")
	}
	worktree := strings.TrimSpace(instance.Roots["worktree"])
	if worktree == "" || worktree == "." || worktree == ".." {
		return InstanceDoctorReport{}, fmt.Errorf("worktree root must be a bounded instance path")
	}
	report := InstanceDoctorReport{
		ContractVersion: InstanceDoctorContract,
		InstanceID:      instance.ID,
		Status:          "ready",
		Checks: map[string]string{
			"instance":            "passed",
			"registry_parity":     "passed",
			"ignored_local_state": "passed",
			"worktree_hygiene":    "passed",
		},
		SchedulesWork: false,
		ExecutesWork:  false,
	}
	if err := ValidateInstanceDoctorReport(report); err != nil {
		return InstanceDoctorReport{}, err
	}
	return report, nil
}

func ValidateIntake(intake Intake) (BlueprintRequest, error) {
	var errs []string
	requireContract(&errs, "intake", intake.ContractVersion, IntakeContract)
	requireField(&errs, "id", intake.ID)
	checkPublicStrings(&errs, "instruction_refs", intake.InstructionRefs, false)
	checkPublicStrings(&errs, "folder_refs", intake.FolderRefs, false)
	missing := []string{}
	if strings.TrimSpace(intake.TargetInstance) == "" {
		missing = append(missing, "target_instance")
	}
	if len(strings.Fields(intake.BroadPrompt)) < 8 {
		missing = append(missing, "broad_prompt_detail")
	}
	if len(intake.Constraints) == 0 {
		missing = append(missing, "constraints")
	}
	if len(missing) > 0 {
		return BlueprintRequest{
			ContractVersion: BlueprintRequestContract,
			IntakeID:        intake.ID,
			Status:          "blueprint_required",
			Missing:         missing,
			Reason:          "AO Atlas cannot compile underspecified intake into a ready workgraph without AO Blueprint clarification.",
		}, joinErrors(errs)
	}
	return BlueprintRequest{}, joinErrors(errs)
}

func ValidateMissionStatus(status MissionStatus) error {
	var errs []string
	requireContract(&errs, "mission_status", status.ContractVersion, MissionStatusContract)
	requireField(&errs, "intake_id", status.IntakeID)
	requireField(&errs, "workgraph_id", status.WorkgraphID)
	requireField(&errs, "target_instance", status.TargetInstance)
	if !oneOf(status.CompletionStatus, "completed", "blocked", "in_progress") {
		errs = append(errs, "completion_status must be completed, blocked, or in_progress")
	}
	for _, key := range []string{"ready", "blocked", "completed"} {
		if _, ok := status.NodeCounts[key]; !ok {
			errs = append(errs, "node_counts."+key+" must be present")
		}
	}
	if len(status.NextActions) == 0 {
		errs = append(errs, "next_actions must not be empty")
	}
	checkPublicStrings(&errs, "next_actions", status.NextActions, true)
	if status.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if status.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	return joinErrors(errs)
}

func BuildMissionStatus(intake Intake, workgraph Workgraph, links []RunLink) (MissionStatus, error) {
	if _, err := ValidateIntake(intake); err != nil {
		return MissionStatus{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return MissionStatus{}, err
	}
	nodeCounts := map[string]int{"ready": 0, "blocked": 0, "completed": 0}
	for _, node := range workgraph.Nodes {
		nodeCounts[node.Status]++
	}
	runLinks := map[string]string{}
	anyBlocked := false
	anyIncomplete := false
	for _, link := range links {
		if err := ValidateRunLink(link); err != nil {
			return MissionStatus{}, err
		}
		runLinks[link.TaskID] = link.Status
		if oneOf(link.Status, "blocked", "failed") {
			anyBlocked = true
		}
		if link.Status != "completed" {
			anyIncomplete = true
		}
	}
	completion := "in_progress"
	nextActions := []string{"continue ready factory tasks through Foundry"}
	if anyBlocked || nodeCounts["blocked"] > 0 {
		completion = "blocked"
		nextActions = []string{"emit repair plan or context repack for blocked task"}
	} else if nodeCounts["ready"] == 0 && !anyIncomplete {
		completion = "completed"
		nextActions = []string{"record completion readback and keep artifacts public-safe"}
	}
	status := MissionStatus{
		ContractVersion:  MissionStatusContract,
		IntakeID:         intake.ID,
		WorkgraphID:      workgraph.ID,
		TargetInstance:   workgraph.TargetInstance,
		CompletionStatus: completion,
		NodeCounts:       nodeCounts,
		RunLinks:         runLinks,
		NextActions:      nextActions,
		SchedulesWork:    false,
		ExecutesWork:     false,
	}
	if err := ValidateMissionStatus(status); err != nil {
		return MissionStatus{}, err
	}
	return status, nil
}

func ValidateBlueprintRequest(request BlueprintRequest) error {
	var errs []string
	requireContract(&errs, "blueprint_request", request.ContractVersion, BlueprintRequestContract)
	requireField(&errs, "intake_id", request.IntakeID)
	if request.Status != "blueprint_required" {
		errs = append(errs, "status must be blueprint_required")
	}
	requireList(&errs, "missing", request.Missing)
	requireField(&errs, "reason", request.Reason)
	checkPublicStrings(&errs, "missing", request.Missing, true)
	checkPublicPath(&errs, "reason", request.Reason, true)
	return joinErrors(errs)
}

func ValidateWorkgraph(workgraph Workgraph) error {
	var errs []string
	requireContract(&errs, "workgraph", workgraph.ContractVersion, WorkgraphContract)
	requireField(&errs, "id", workgraph.ID)
	requireField(&errs, "target_instance", workgraph.TargetInstance)
	if len(workgraph.Nodes) == 0 {
		errs = append(errs, "nodes must not be empty")
	}
	seen := map[string]WorkgraphNode{}
	for i, node := range workgraph.Nodes {
		field := fmt.Sprintf("nodes[%d]", i)
		requireField(&errs, field+".id", node.ID)
		if _, ok := seen[node.ID]; ok {
			errs = append(errs, field+".id must be unique")
		}
		seen[node.ID] = node
		if !oneOf(node.Status, "ready", "blocked", "completed") {
			errs = append(errs, field+".status must be ready, blocked, or completed")
		}
		if node.Status == "blocked" && len(node.Blockers) == 0 {
			errs = append(errs, field+".blockers must explain blocked state")
		}
		if err := ValidateFactoryTask(node.FactoryTask); err != nil {
			errs = append(errs, field+".factory_task: "+err.Error())
		}
	}
	for _, node := range workgraph.Nodes {
		for _, dep := range node.Dependencies {
			if _, ok := seen[dep]; !ok {
				errs = append(errs, "dependency "+dep+" does not reference an existing node")
			}
		}
	}
	return joinErrors(errs)
}

func ValidateWorkgraphRepairPlan(plan WorkgraphRepairPlan) error {
	var errs []string
	requireContract(&errs, "workgraph_repair_plan", plan.ContractVersion, WorkgraphRepairPlanContract)
	requireField(&errs, "id", plan.ID)
	requireField(&errs, "task_id", plan.TaskID)
	if plan.Status != "repair_required" {
		errs = append(errs, "status must be repair_required")
	}
	if !oneOf(plan.SourceRunLinkStatus, "blocked", "failed") {
		errs = append(errs, "source_run_link_status must be blocked or failed")
	}
	requireField(&errs, "reason", plan.Reason)
	if len(plan.RepairTasks) == 0 {
		errs = append(errs, "repair_tasks must not be empty")
	}
	for i, task := range plan.RepairTasks {
		if err := ValidateFactoryTask(task); err != nil {
			errs = append(errs, fmt.Sprintf("repair_tasks[%d]: %s", i, err.Error()))
		}
	}
	if plan.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if plan.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if plan.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	return joinErrors(errs)
}

func ValidateFactoryTask(task FactoryTask) error {
	var errs []string
	requireContract(&errs, "factory_task", task.ContractVersion, FactoryTaskContract)
	requireField(&errs, "id", task.ID)
	requireField(&errs, "objective", task.Objective)
	requireField(&errs, "target_factory_repo", task.TargetFactoryRepo)
	requireField(&errs, "factory_folder", task.FactoryFolder)
	requireList(&errs, "acceptance_criteria", task.Acceptance)
	requireList(&errs, "non_goals", task.NonGoals)
	requireList(&errs, "write_scope", task.WriteScope)
	requireList(&errs, "verification_commands", task.Verification)
	requireList(&errs, "required_evidence", task.RequiredEvidence)
	requireList(&errs, "safety_limits", task.SafetyLimits)
	checkPublicPath(&errs, "target_factory_repo", task.TargetFactoryRepo, false)
	checkPublicPath(&errs, "factory_folder", task.FactoryFolder, false)
	checkPublicStrings(&errs, "write_scope", task.WriteScope, false)
	return joinErrors(errs)
}

func ValidateFactoryMaterialization(materialization FactoryMaterialization) error {
	var errs []string
	requireContract(&errs, "factory_materialization", materialization.ContractVersion, FactoryMaterializationContract)
	requireField(&errs, "task_id", materialization.TaskID)
	if materialization.Mode != "dry_run" {
		errs = append(errs, "mode must be dry_run")
	}
	requireField(&errs, "output_root", materialization.OutputRoot)
	if strings.ContainsAny(materialization.OutputRoot, `/\`) {
		errs = append(errs, "output_root must not record a local path")
	}
	requireList(&errs, "files", materialization.Files)
	checkPublicStrings(&errs, "files", materialization.Files, true)
	if materialization.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if materialization.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if !digestPattern.MatchString(materialization.TaskDigest) {
		errs = append(errs, "task_digest must be sha256:<64 hex>")
	}
	return joinErrors(errs)
}

func ValidateContextPack(pack ContextPack, budgetOverride int) error {
	var errs []string
	requireContract(&errs, "context_pack", pack.ContractVersion, ContextPackContract)
	requireField(&errs, "id", pack.ID)
	requireField(&errs, "task_id", pack.TaskID)
	if pack.BudgetBytes <= 0 {
		errs = append(errs, "budget_bytes must be positive")
	}
	budget := pack.BudgetBytes
	if budgetOverride > 0 {
		budget = budgetOverride
	}
	data, _ := json.Marshal(pack)
	if len(data) > budget {
		errs = append(errs, fmt.Sprintf("context pack exceeds budget: %d > %d bytes", len(data), budget))
	}
	if len(pack.SourceRefs) == 0 {
		errs = append(errs, "source_refs must not be empty")
	}
	for i, ref := range pack.SourceRefs {
		checkPublicPath(&errs, fmt.Sprintf("source_refs[%d].ref", i), ref.Ref, true)
		if !digestPattern.MatchString(ref.Digest) {
			errs = append(errs, fmt.Sprintf("source_refs[%d].digest must be sha256:<64 hex>", i))
		}
	}
	requireList(&errs, "summaries", pack.Summaries)
	requireField(&errs, "missing_context_protocol", pack.MissingProtocol)
	checkPublicStrings(&errs, "summaries", pack.Summaries, true)
	checkPublicStrings(&errs, "assumptions", pack.Assumptions, true)
	checkPublicStrings(&errs, "exclusions", pack.Exclusions, true)
	return joinErrors(errs)
}

func BuildContextRepack(task FactoryTask, link RunLink, sourceRef, sourceDigest string, budget int) (ContextPack, error) {
	if err := ValidateFactoryTask(task); err != nil {
		return ContextPack{}, err
	}
	if err := ValidateRunLink(link); err != nil {
		return ContextPack{}, err
	}
	if task.ID != link.TaskID {
		return ContextPack{}, fmt.Errorf("run-link task_id must match factory task id")
	}
	if !oneOf(link.Status, "blocked", "failed") {
		return ContextPack{}, fmt.Errorf("run-link status must be blocked or failed")
	}
	if strings.TrimSpace(link.Evidence["needs_context"]) == "" {
		return ContextPack{}, fmt.Errorf("run-link evidence must include needs_context")
	}
	pack := ContextPack{
		ContractVersion: ContextPackContract,
		ID:              task.ID + "-context-repack",
		TaskID:          task.ID,
		BudgetBytes:     budget,
		SourceRefs:      []SourceRef{{Ref: sourceRef, Digest: sourceDigest}},
		Summaries:       []string{"Repacked bounded context requested by a needs_context run-link."},
		Assumptions:     []string{"Only referenced public-safe sources are included."},
		Exclusions:      []string{"whole source repositories", "private local state", "credentials", "provider transcripts"},
		MissingProtocol: "Ask AO Blueprint or the operator for missing requirements before widening scope.",
	}
	if err := ValidateContextPack(pack, 0); err != nil {
		return ContextPack{}, err
	}
	return pack, nil
}

func ValidateFoundryHandoff(handoff FoundryHandoff) error {
	var errs []string
	requireContract(&errs, "foundry_handoff", handoff.ContractVersion, FoundryHandoffContract)
	requireField(&errs, "id", handoff.ID)
	requireField(&errs, "target_instance", handoff.TargetInstance)
	if handoff.Status != "ready_for_foundry" {
		errs = append(errs, "status must be ready_for_foundry")
	}
	if len(handoff.Tasks) == 0 {
		errs = append(errs, "tasks must not be empty")
	}
	for i, task := range handoff.Tasks {
		requireField(&errs, fmt.Sprintf("tasks[%d].id", i), task.ID)
		requireField(&errs, fmt.Sprintf("tasks[%d].objective", i), task.Objective)
		checkPublicPath(&errs, fmt.Sprintf("tasks[%d].target_factory_repo", i), task.TargetFactoryRepo, false)
		checkPublicPath(&errs, fmt.Sprintf("tasks[%d].factory_folder", i), task.FactoryFolder, false)
	}
	return joinErrors(errs)
}

func ValidateFoundryImport(foundryImport FoundryImport) error {
	var errs []string
	requireContract(&errs, "foundry_import", foundryImport.ContractVersion, FoundryImportContract)
	requireField(&errs, "id", foundryImport.ID)
	requireField(&errs, "workgraph_id", foundryImport.WorkgraphID)
	requireField(&errs, "target_instance", foundryImport.TargetInstance)
	if foundryImport.Status != "ready_for_foundry_fixture_import" {
		errs = append(errs, "status must be ready_for_foundry_fixture_import")
	}
	if len(foundryImport.Tasks) == 0 {
		errs = append(errs, "tasks must not be empty")
	}
	if foundryImport.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if foundryImport.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if foundryImport.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	seenPaths := map[string]bool{}
	for i, fixture := range foundryImport.Tasks {
		prefix := fmt.Sprintf("tasks[%d]", i)
		requireField(&errs, prefix+".node_id", fixture.NodeID)
		requireField(&errs, prefix+".task_id", fixture.TaskID)
		requireField(&errs, prefix+".path", fixture.Path)
		checkPublicPath(&errs, prefix+".path", fixture.Path, true)
		if strings.Contains(filepath.Clean(fixture.Path), "..") {
			errs = append(errs, prefix+".path must stay inside the import output")
		}
		if seenPaths[fixture.Path] {
			errs = append(errs, prefix+".path must be unique")
		}
		seenPaths[fixture.Path] = true
		if !digestPattern.MatchString(fixture.TaskHash) {
			errs = append(errs, prefix+".task_digest must be sha256:<64 hex>")
		}
		if err := ValidateFactoryTask(fixture.Task); err != nil {
			errs = append(errs, prefix+".task: "+err.Error())
		}
		if fixture.TaskID != fixture.Task.ID {
			errs = append(errs, prefix+".task_id must match task.id")
		}
	}
	return joinErrors(errs)
}

func ValidateRunLink(link RunLink) error {
	var errs []string
	requireContract(&errs, "run_link", link.ContractVersion, RunLinkContract)
	requireField(&errs, "task_id", link.TaskID)
	if !oneOf(link.Status, "planned", "running", "completed", "blocked", "failed") {
		errs = append(errs, "status must be planned, running, completed, blocked, or failed")
	}
	if len(link.Evidence) == 0 {
		errs = append(errs, "evidence must not be empty")
	}
	checkPublicPathMapStrict(&errs, link.Evidence)
	if !digestPattern.MatchString(link.Digest) {
		errs = append(errs, "digest must be sha256:<64 hex>")
	}
	return joinErrors(errs)
}

func BuildRunLink(taskID, status string, evidence map[string]string) (RunLink, error) {
	link := RunLink{
		ContractVersion: RunLinkContract,
		TaskID:          taskID,
		Status:          status,
		Evidence:        evidence,
	}
	link.Digest = digestRunLink(link)
	if err := ValidateRunLink(link); err != nil {
		return RunLink{}, err
	}
	return link, nil
}

func NextReadyNode(workgraph Workgraph) (WorkgraphNode, bool) {
	statusByID := map[string]string{}
	for _, node := range workgraph.Nodes {
		statusByID[node.ID] = node.Status
	}
	for _, node := range workgraph.Nodes {
		if node.Status != "ready" {
			continue
		}
		ok := true
		for _, dep := range node.Dependencies {
			if statusByID[dep] != "completed" {
				ok = false
				break
			}
		}
		if ok {
			return node, true
		}
	}
	return WorkgraphNode{}, false
}

func CompleteWorkgraph(workgraph Workgraph, link RunLink) (Workgraph, string, error) {
	if err := ValidateWorkgraph(workgraph); err != nil {
		return Workgraph{}, "", err
	}
	if err := ValidateRunLink(link); err != nil {
		return Workgraph{}, "", err
	}
	if link.Status != "completed" {
		return Workgraph{}, "", fmt.Errorf("run-link status must be completed")
	}
	statusByID := map[string]string{}
	for _, node := range workgraph.Nodes {
		statusByID[node.ID] = node.Status
	}
	for i, node := range workgraph.Nodes {
		if node.FactoryTask.ID != link.TaskID {
			continue
		}
		for _, dep := range node.Dependencies {
			if statusByID[dep] != "completed" {
				return Workgraph{}, "", fmt.Errorf("matching node dependencies must be completed")
			}
		}
		updated := workgraph
		updated.Nodes = append([]WorkgraphNode(nil), workgraph.Nodes...)
		updated.Nodes[i].Status = "completed"
		if err := ValidateWorkgraph(updated); err != nil {
			return Workgraph{}, "", err
		}
		return updated, node.ID, nil
	}
	return Workgraph{}, "", fmt.Errorf("no matching workgraph node for run-link task_id %q", link.TaskID)
}

func BuildWorkgraphRepairPlan(workgraph Workgraph, link RunLink) (WorkgraphRepairPlan, error) {
	if err := ValidateWorkgraph(workgraph); err != nil {
		return WorkgraphRepairPlan{}, err
	}
	if err := ValidateRunLink(link); err != nil {
		return WorkgraphRepairPlan{}, err
	}
	if !oneOf(link.Status, "blocked", "failed") {
		return WorkgraphRepairPlan{}, fmt.Errorf("run-link status must be blocked or failed")
	}
	for _, node := range workgraph.Nodes {
		if node.FactoryTask.ID != link.TaskID {
			continue
		}
		source := node.FactoryTask
		repair := FactoryTask{
			ContractVersion:   FactoryTaskContract,
			ID:                "repair-" + source.ID,
			Objective:         "Repair blocked Atlas factory task: " + source.Objective,
			TargetFactoryRepo: source.TargetFactoryRepo,
			FactoryFolder:     source.FactoryFolder + "-repair",
			Acceptance:        []string{"a follow-up run-link for " + source.ID + " validates with status completed"},
			NonGoals:          []string{"do not schedule work from Atlas", "do not execute work from Atlas", "do not approve work from Atlas"},
			WriteScope:        append([]string(nil), source.WriteScope...),
			Verification:      append([]string(nil), source.Verification...),
			RequiredEvidence:  append([]string(nil), source.RequiredEvidence...),
			SafetyLimits:      append(append([]string(nil), source.SafetyLimits...), "repair plan is readback only"),
			DependencyRefs:    []string{source.ID},
		}
		plan := WorkgraphRepairPlan{
			ContractVersion:     WorkgraphRepairPlanContract,
			ID:                  workgraph.ID + "-" + source.ID + "-repair-plan",
			TaskID:              source.ID,
			Status:              "repair_required",
			SourceRunLinkStatus: link.Status,
			Reason:              "run-link did not complete the task; emit a bounded repair task for Foundry scheduling",
			RepairTasks:         []FactoryTask{repair},
			SchedulesWork:       false,
			ExecutesWork:        false,
			ApprovesWork:        false,
		}
		if err := ValidateWorkgraphRepairPlan(plan); err != nil {
			return WorkgraphRepairPlan{}, err
		}
		return plan, nil
	}
	return WorkgraphRepairPlan{}, fmt.Errorf("no matching workgraph node for run-link task_id %q", link.TaskID)
}

func BuildFoundryHandoff(workgraph Workgraph) FoundryHandoff {
	tasks := []FoundryTaskEntry{}
	for _, node := range workgraph.Nodes {
		if node.Status != "ready" {
			continue
		}
		task := node.FactoryTask
		tasks = append(tasks, FoundryTaskEntry{
			ID:                task.ID,
			Objective:         task.Objective,
			TargetFactoryRepo: task.TargetFactoryRepo,
			FactoryFolder:     task.FactoryFolder,
			Verification:      task.Verification,
			RequiredEvidence:  task.RequiredEvidence,
		})
	}
	return FoundryHandoff{
		ContractVersion: FoundryHandoffContract,
		ID:              workgraph.ID + "-foundry-handoff",
		TargetInstance:  workgraph.TargetInstance,
		Status:          "ready_for_foundry",
		Tasks:           tasks,
	}
}

func BuildFoundryImport(workgraph Workgraph) (FoundryImport, error) {
	if err := ValidateWorkgraph(workgraph); err != nil {
		return FoundryImport{}, err
	}
	statusByID := map[string]string{}
	for _, node := range workgraph.Nodes {
		statusByID[node.ID] = node.Status
	}
	fixtures := []FoundryImportTaskFixture{}
	for _, node := range workgraph.Nodes {
		if node.Status != "ready" {
			continue
		}
		ready := true
		for _, dep := range node.Dependencies {
			if statusByID[dep] != "completed" {
				ready = false
				break
			}
		}
		if !ready {
			continue
		}
		task := node.FactoryTask
		fixture := FoundryImportTaskFixture{
			NodeID:   node.ID,
			TaskID:   task.ID,
			Path:     filepath.ToSlash(filepath.Join("tasks", task.ID+".json")),
			Task:     task,
			TaskHash: digestFactoryTask(task),
		}
		fixtures = append(fixtures, fixture)
	}
	foundryImport := FoundryImport{
		ContractVersion: FoundryImportContract,
		ID:              workgraph.ID + "-foundry-import",
		WorkgraphID:     workgraph.ID,
		TargetInstance:  workgraph.TargetInstance,
		Status:          "ready_for_foundry_fixture_import",
		Tasks:           fixtures,
		SchedulesWork:   false,
		ExecutesWork:    false,
		ApprovesWork:    false,
	}
	if err := ValidateFoundryImport(foundryImport); err != nil {
		return FoundryImport{}, err
	}
	return foundryImport, nil
}

func DigestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestFactoryTask(task FactoryTask) string {
	data, err := json.Marshal(task)
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

func requireContract(errs *[]string, name, got, want string) {
	if got != want {
		*errs = append(*errs, name+" contract_version must be "+want)
	}
}

func requireField(errs *[]string, field, value string) {
	if strings.TrimSpace(value) == "" {
		*errs = append(*errs, field+" must not be empty")
	}
}

func requireList(errs *[]string, field string, values []string) {
	if len(values) == 0 {
		*errs = append(*errs, field+" must not be empty")
	}
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			*errs = append(*errs, fmt.Sprintf("%s[%d] must not be empty", field, i))
		}
	}
}

func checkPublicPathMap(errs *[]string, values map[string]string) {
	for key, value := range values {
		checkPublicPath(errs, key, value, false)
	}
}

func checkPublicPathMapStrict(errs *[]string, values map[string]string) {
	for key, value := range values {
		requireField(errs, key, value)
		checkPublicPath(errs, key, value, true)
	}
}

func checkPublicStrings(errs *[]string, field string, values []string, rejectAbsolute bool) {
	for i, value := range values {
		checkPublicPath(errs, fmt.Sprintf("%s[%d]", field, i), value, rejectAbsolute)
	}
}

func checkPublicPath(errs *[]string, field, value string, rejectAbsolute bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	normalized := strings.ReplaceAll(value, "\\", "/")
	lower := strings.ToLower(normalized)
	for _, marker := range []string{
		"/" + "users/",
		"/" + "home/",
		"/" + "tmp/",
		"/" + "var/folders/",
		"downloads/",
		"file:" + "//",
		".ssh/",
		".aws/",
		".config/",
	} {
		if strings.Contains(lower, marker) {
			*errs = append(*errs, field+" contains a private or machine-local path")
			return
		}
	}
	if rejectAbsolute && (strings.HasPrefix(normalized, "/") || driveAbsPattern.MatchString(value)) {
		*errs = append(*errs, field+" must not be an absolute local path")
	}
}

func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func joinErrors(errs []string) error {
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

func DefaultInstance(id, stateRoot, toolchainRoot string) Instance {
	cleanState := filepath.ToSlash(filepath.Clean(stateRoot))
	root := func(name string) string {
		return filepath.ToSlash(filepath.Join(cleanState, id, name))
	}
	if runtime.GOOS == "windows" {
		cleanState = strings.ReplaceAll(cleanState, "\\", "/")
	}
	return Instance{
		ContractVersion: InstanceContract,
		ID:              id,
		StateRoot:       cleanState,
		ToolchainRoot:   filepath.ToSlash(filepath.Clean(toolchainRoot)),
		Roots: map[string]string{
			"mission":   root("mission"),
			"workgraph": root("workgraph"),
			"context":   root("context"),
			"evidence":  root("evidence"),
			"worktree":  root("worktree"),
		},
	}
}

func digestRunLink(link RunLink) string {
	payload := struct {
		ContractVersion string            `json:"contract_version"`
		TaskID          string            `json:"task_id"`
		Status          string            `json:"status"`
		Evidence        map[string]string `json:"evidence"`
	}{
		ContractVersion: link.ContractVersion,
		TaskID:          link.TaskID,
		Status:          link.Status,
		Evidence:        link.Evidence,
	}
	data, _ := json.Marshal(payload)
	return DigestBytes(data)
}
