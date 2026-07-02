package atlas

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FoundryImportBuilder assembles the Atlas-owned Foundry import projection from
// a validated workgraph. It does not schedule, approve, or execute work.
type FoundryImportBuilder struct {
	Workgraph       Workgraph
	SelectedNodes   []string
	SourceArtifacts []SourceRef
}

func (b FoundryImportBuilder) Build() (FoundryImport, error) {
	return buildFoundryImportForNodes(b.Workgraph, b.SelectedNodes, b.SourceArtifacts)
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
	return FoundryImportBuilder{Workgraph: workgraph}.Build()
}

func BuildFoundryImportForNodes(workgraph Workgraph, selectedNodes []string, sourceArtifacts []SourceRef) (FoundryImport, error) {
	return FoundryImportBuilder{
		Workgraph:       workgraph,
		SelectedNodes:   selectedNodes,
		SourceArtifacts: sourceArtifacts,
	}.Build()
}

func buildFoundryImportForNodes(workgraph Workgraph, selectedNodes []string, sourceArtifacts []SourceRef) (FoundryImport, error) {
	if err := ValidateWorkgraph(workgraph); err != nil {
		return FoundryImport{}, err
	}
	statusByID := map[string]string{}
	for _, node := range workgraph.Nodes {
		statusByID[node.ID] = node.Status
	}
	selected := map[string]bool{}
	for _, nodeID := range selectedNodes {
		if strings.TrimSpace(nodeID) != "" {
			selected[nodeID] = true
		}
	}
	for selectedID := range selected {
		found := false
		for _, node := range workgraph.Nodes {
			if node.ID != selectedID {
				continue
			}
			found = true
			if node.Status != "ready" {
				return FoundryImport{}, fmt.Errorf("selected node %s must be ready", selectedID)
			}
			break
		}
		if !found {
			return FoundryImport{}, fmt.Errorf("selected node %s was not found", selectedID)
		}
	}
	fixtures := []FoundryImportTaskFixture{}
	for _, node := range workgraph.Nodes {
		if len(selected) > 0 && !selected[node.ID] {
			continue
		}
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
			return FoundryImport{}, fmt.Errorf("selected node %s has incomplete dependencies", node.ID)
		}
		task := node.FactoryTask
		if err := ValidateFoundryReadyTaskAuthorityMetadata(task); err != nil {
			return FoundryImport{}, fmt.Errorf("ready node %s authority metadata: %w", node.ID, err)
		}
		fixture := FoundryImportTaskFixture{
			NodeID:            node.ID,
			TaskID:            task.ID,
			Path:              filepath.ToSlash(filepath.Join("tasks", task.ID+".json")),
			MutationClass:     task.MutationClass,
			WriteScope:        append([]string(nil), task.WriteScope...),
			RollbackScope:     append([]string(nil), task.RollbackScope...),
			RequiredGates:     append([]string(nil), task.RequiredGates...),
			RequiredEvidence:  append([]string(nil), task.RequiredEvidence...),
			AuthorityBoundary: task.AuthorityBoundary,
			Task:              task,
			TaskHash:          digestFactoryTask(task),
		}
		fixtures = append(fixtures, fixture)
	}
	if len(sourceArtifacts) == 0 {
		sourceArtifacts = []SourceRef{{Ref: "generated", Digest: digestFoundryImportSources(workgraph)}}
	}
	foundryImport := FoundryImport{
		ContractVersion: FoundryImportContract,
		ID:              workgraph.ID + "-foundry-import",
		WorkgraphID:     workgraph.ID,
		TargetInstance:  workgraph.TargetInstance,
		Status:          "ready_for_foundry_fixture_import",
		SourceArtifacts: sourceArtifacts,
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

func digestFoundryImportSources(workgraph Workgraph) string {
	data, _ := json.Marshal(workgraph)
	return DigestBytes(data)
}

func digestFactoryTask(task FactoryTask) string {
	data, err := json.Marshal(task)
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

// FoundryContinuationHandoffBuilder creates the operator-ready Foundry
// continuation prompt and validates that the import still matches the workgraph.
type FoundryContinuationHandoffBuilder struct {
	Workgraph     Workgraph
	FoundryImport FoundryImport
	Inputs        FoundryContinuationHandoffInputs
}

type FoundryContinuationHandoffInputs struct {
	BlueprintPackPath               string
	AtlasImportPath                 string
	WorkgraphPath                   string
	FoundryImportPath               string
	MissionContinuationEvidencePath string
}

func BuildFoundryContinuationHandoff(workgraph Workgraph, foundryImport FoundryImport, inputs FoundryContinuationHandoffInputs) (FoundryContinuationHandoff, error) {
	return FoundryContinuationHandoffBuilder{
		Workgraph:     workgraph,
		FoundryImport: foundryImport,
		Inputs:        inputs,
	}.Build()
}

func (b FoundryContinuationHandoffBuilder) Build() (FoundryContinuationHandoff, error) {
	if err := ValidateWorkgraph(b.Workgraph); err != nil {
		return FoundryContinuationHandoff{}, err
	}
	if err := ValidateFoundryImport(b.FoundryImport); err != nil {
		return FoundryContinuationHandoff{}, err
	}
	if err := ValidateFoundryImportMatchesWorkgraph(b.Workgraph, b.FoundryImport); err != nil {
		return FoundryContinuationHandoff{}, err
	}
	state, err := BuildWorkgraphState(b.Workgraph)
	if err != nil {
		return FoundryContinuationHandoff{}, err
	}
	firstSafeNode := ""
	if len(b.FoundryImport.Tasks) > 0 {
		firstSafeNode = b.FoundryImport.Tasks[0].NodeID
	}
	classBoundary := foundryContinuationClassBoundary(b.FoundryImport)
	handoff := FoundryContinuationHandoff{
		ContractVersion:                 FoundryContinuationHandoffContract,
		ID:                              b.FoundryImport.ID + "-continuation-handoff",
		TargetFolder:                    foundryContinuationTargetFolder(),
		Command:                         "codex --yolo",
		BlueprintPackPath:               slashOrNotProvided(b.Inputs.BlueprintPackPath),
		AtlasImportPath:                 slashOrNotProvided(b.Inputs.AtlasImportPath),
		WorkgraphPath:                   slashOrNotProvided(b.Inputs.WorkgraphPath),
		FoundryImportPath:               slashOrNotProvided(b.Inputs.FoundryImportPath),
		MissionContinuationEvidencePath: slashOrNotProvided(b.Inputs.MissionContinuationEvidencePath),
		FirstSafeNode:                   firstSafeNode,
		TotalNodeCount:                  len(b.Workgraph.Nodes),
		CompletedNodeCount:              state.NodeCounts["completed"],
		BlockedNodeCount:                state.NodeCounts["blocked"],
		ReadyNodeCount:                  state.NodeCounts["ready"],
		ClassBoundary:                   classBoundary,
		StopConditions: []string{
			"done",
			"final denial",
			"hard blocker",
			"CI failure",
			"unsafe scope drift",
			"kill switch",
		},
		SafetyProhibitions: []string{
			"Atlas must not execute live mutation",
			"no direct main mutation",
			"no release deploy publish upload tag provider call credential use dependency update auth policy widening secret env exposure or config expansion",
			"do not claim fully_unsupervised_complex_mutation or RSI is proven",
			"do not claim complex_repo_mutation is live-proven unless downstream evidence proves it",
		},
		SchedulesWork: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	}
	handoff.NextRecommendedAction = "Move to " + handoff.TargetFolder + "; Run codex --yolo; Paste this prompt"
	handoff.Prompt = buildFoundryContinuationPrompt(handoff)
	if err := ValidateFoundryContinuationHandoff(handoff); err != nil {
		return FoundryContinuationHandoff{}, err
	}
	return handoff, nil
}

func WriteFoundryContinuationPrompt(path string, handoff FoundryContinuationHandoff) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(foundryContinuationPromptMarkdown(handoff)), 0o644)
}

func foundryContinuationTargetFolder() string {
	return strings.Join([]string{"", "Users", "torachiyouesugi", "Documents", "public", "ao-foundry"}, "/")
}

func slashOrNotProvided(path string) string {
	if strings.TrimSpace(path) == "" {
		return "not_provided"
	}
	return filepath.ToSlash(path)
}

func foundryContinuationClassBoundary(foundryImport FoundryImport) string {
	classes := map[string]bool{}
	for _, task := range foundryImport.Tasks {
		if strings.TrimSpace(task.MutationClass) != "" {
			classes[task.MutationClass] = true
		}
	}
	if len(classes) == 0 {
		return "Atlas import only; Foundry must preserve Atlas no-execution boundary"
	}
	names := make([]string, 0, len(classes))
	for name := range classes {
		names = append(names, name)
	}
	sort.Strings(names)
	return "Atlas import only for " + strings.Join(names, ", ") + "; Foundry must preserve Atlas no-execution boundary"
}

func buildFoundryContinuationPrompt(handoff FoundryContinuationHandoff) string {
	var b strings.Builder
	b.WriteString("You are AO Foundry. Continue from the AO Atlas first-phase handoff.\n\n")
	b.WriteString("Source artifacts:\n")
	b.WriteString("- Blueprint pack: " + handoff.BlueprintPackPath + "\n")
	b.WriteString("- Atlas import: " + handoff.AtlasImportPath + "\n")
	b.WriteString("- Atlas workgraph: " + handoff.WorkgraphPath + "\n")
	b.WriteString("- Foundry import: " + handoff.FoundryImportPath + "\n")
	if handoff.MissionContinuationEvidencePath != "not_provided" {
		b.WriteString("- Mission continuation evidence: " + handoff.MissionContinuationEvidencePath + "\n")
	}
	b.WriteString("\nCurrent Atlas readback:\n")
	b.WriteString(fmt.Sprintf("- first safe node: %s\n", handoff.FirstSafeNode))
	b.WriteString(fmt.Sprintf("- total nodes: %d\n", handoff.TotalNodeCount))
	b.WriteString(fmt.Sprintf("- completed nodes: %d\n", handoff.CompletedNodeCount))
	b.WriteString(fmt.Sprintf("- ready nodes: %d\n", handoff.ReadyNodeCount))
	b.WriteString(fmt.Sprintf("- blocked nodes: %d\n", handoff.BlockedNodeCount))
	b.WriteString("- class boundary: " + handoff.ClassBoundary + "\n\n")
	b.WriteString("Required continuation behavior:\n")
	b.WriteString("- Move to AO Foundry.\n")
	b.WriteString("- Run codex --yolo.\n")
	b.WriteString("- Paste this prompt.\n")
	b.WriteString("- Import and validate the Foundry import.\n")
	b.WriteString("- do not stop after import validation.\n")
	b.WriteString("- do not stop after one gate artifact.\n")
	b.WriteString("- do not stop after one node.\n")
	b.WriteString("- Continue until all generated slices/tasks/nodes are consumed or a true hard blocker remains.\n")
	b.WriteString("- If evidence/schema/readback support is missing and can be safely implemented, implement it with PR/CI/merge.\n\n")
	b.WriteString("Hard safety prohibitions:\n")
	for _, prohibition := range handoff.SafetyProhibitions {
		b.WriteString("- " + prohibition + "\n")
	}
	b.WriteString("- fully_unsupervised_complex_mutation remains denied.\n")
	b.WriteString("- RSI remains denied.\n\n")
	b.WriteString("Stop conditions:\n")
	for _, condition := range handoff.StopConditions {
		b.WriteString("- " + condition + "\n")
	}
	b.WriteString("\nStop only on done, final denial, hard blocker, CI failure, unsafe scope drift, or kill switch.\n")
	return b.String()
}

func foundryContinuationPromptMarkdown(handoff FoundryContinuationHandoff) string {
	var b strings.Builder
	b.WriteString("# AO Foundry Continuation Handoff\n\n")
	b.WriteString("Move to AO Foundry:\n\n")
	b.WriteString("```sh\n")
	b.WriteString("cd " + handoff.TargetFolder + "\n")
	b.WriteString("codex --yolo\n")
	b.WriteString("```\n\n")
	b.WriteString("Paste this prompt:\n\n")
	b.WriteString("```text\n")
	b.WriteString(handoff.Prompt)
	b.WriteString("```\n")
	return b.String()
}

func ValidateFoundryImportMatchesWorkgraph(workgraph Workgraph, foundryImport FoundryImport) error {
	if foundryImport.WorkgraphID != workgraph.ID {
		return fmt.Errorf("foundry import workgraph_id must match workgraph")
	}
	if foundryImport.TargetInstance != workgraph.TargetInstance {
		return fmt.Errorf("foundry import target_instance must match workgraph")
	}
	nodes := map[string]WorkgraphNode{}
	for _, node := range workgraph.Nodes {
		nodes[node.ID] = node
	}
	for _, task := range foundryImport.Tasks {
		node, ok := nodes[task.NodeID]
		if !ok {
			return fmt.Errorf("foundry import node_id must match a workgraph node")
		}
		if node.Status != "ready" {
			return fmt.Errorf("foundry import node_id must reference a ready workgraph node")
		}
		if task.TaskID != node.FactoryTask.ID || task.Task.ID != node.FactoryTask.ID {
			return fmt.Errorf("foundry import task_id must match ready workgraph node task")
		}
		if task.TaskHash != digestFactoryTask(node.FactoryTask) {
			return fmt.Errorf("foundry import task_hash must match ready workgraph node task")
		}
	}
	return nil
}
