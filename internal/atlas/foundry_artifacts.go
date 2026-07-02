package atlas

import "fmt"

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

// FoundryContinuationHandoffBuilder creates the operator-ready Foundry
// continuation prompt and validates that the import still matches the workgraph.
type FoundryContinuationHandoffBuilder struct {
	Workgraph     Workgraph
	FoundryImport FoundryImport
	Inputs        FoundryContinuationHandoffInputs
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
	firstSafeNode := ""
	if len(b.FoundryImport.Tasks) > 0 {
		firstSafeNode = b.FoundryImport.Tasks[0].NodeID
	}
	counts := map[string]int{}
	for _, node := range b.Workgraph.Nodes {
		counts[node.Status]++
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
		CompletedNodeCount:              counts["completed"],
		BlockedNodeCount:                counts["blocked"],
		ReadyNodeCount:                  counts["ready"],
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
