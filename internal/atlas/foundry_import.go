package atlas

import (
	"encoding/json"
	"fmt"
	"path/filepath"
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
