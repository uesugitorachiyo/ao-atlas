package atlas

import "fmt"

type WorkgraphNodeState struct {
	NodeID               string
	TaskID               string
	Status               string
	Dependencies         []string
	DependenciesComplete bool
	ExecutableReady      bool
}

type WorkgraphState struct {
	Workgraph              Workgraph
	NodeCounts             map[string]int
	ReadyTaskIDs           []string
	ExecutableReadyNodeIDs []string
	Nodes                  []WorkgraphNodeState
	nodeStateByID          map[string]WorkgraphNodeState
}

func BuildWorkgraphState(workgraph Workgraph) (WorkgraphState, error) {
	if err := ValidateWorkgraph(workgraph); err != nil {
		return WorkgraphState{}, err
	}
	statusByID := map[string]string{}
	for _, node := range workgraph.Nodes {
		statusByID[node.ID] = node.Status
	}
	state := WorkgraphState{
		Workgraph:     workgraph,
		NodeCounts:    map[string]int{"ready": 0, "blocked": 0, "completed": 0, "failed": 0},
		nodeStateByID: map[string]WorkgraphNodeState{},
	}
	for _, node := range workgraph.Nodes {
		state.NodeCounts[node.Status]++
		dependenciesComplete := node.Status == "ready"
		if dependenciesComplete {
			for _, dep := range node.Dependencies {
				if statusByID[dep] != "completed" {
					dependenciesComplete = false
					break
				}
			}
		}
		nodeState := WorkgraphNodeState{
			NodeID:               node.ID,
			TaskID:               node.FactoryTask.ID,
			Status:               node.Status,
			Dependencies:         append([]string(nil), node.Dependencies...),
			DependenciesComplete: dependenciesComplete,
			ExecutableReady:      node.Status == "ready" && dependenciesComplete,
		}
		state.Nodes = append(state.Nodes, nodeState)
		state.nodeStateByID[node.ID] = nodeState
		if node.Status == "ready" {
			state.ReadyTaskIDs = append(state.ReadyTaskIDs, node.FactoryTask.ID)
		}
		if nodeState.ExecutableReady {
			state.ExecutableReadyNodeIDs = append(state.ExecutableReadyNodeIDs, node.ID)
		}
	}
	return state, nil
}

func (state WorkgraphState) NodeState(nodeID string) (WorkgraphNodeState, bool) {
	if state.nodeStateByID != nil {
		node, ok := state.nodeStateByID[nodeID]
		return node, ok
	}
	for _, node := range state.Nodes {
		if node.NodeID == nodeID {
			return node, true
		}
	}
	return WorkgraphNodeState{}, false
}

func (state WorkgraphState) NextReadyNode() (WorkgraphNode, bool) {
	for _, node := range state.Workgraph.Nodes {
		nodeState, ok := state.NodeState(node.ID)
		if ok && nodeState.ExecutableReady {
			return node, true
		}
	}
	return WorkgraphNode{}, false
}

func (state WorkgraphState) CompleteWithRunLink(link RunLink) (Workgraph, string, error) {
	if err := ValidateRunLink(link); err != nil {
		return Workgraph{}, "", err
	}
	if link.Status != "completed" {
		return Workgraph{}, "", fmt.Errorf("run-link status must be completed")
	}
	matchedTask := false
	for i, node := range state.Workgraph.Nodes {
		if node.FactoryTask.ID != link.TaskID {
			continue
		}
		matchedTask = true
		nodeState, ok := state.NodeState(node.ID)
		if !ok || !nodeState.ExecutableReady {
			if ok && nodeState.Status == "ready" && !nodeState.DependenciesComplete {
				return Workgraph{}, "", fmt.Errorf("matching node dependencies must be completed")
			}
			continue
		}
		updated := state.Workgraph
		updated.Nodes = append([]WorkgraphNode(nil), state.Workgraph.Nodes...)
		updated.Nodes[i].Status = "completed"
		if err := ValidateWorkgraph(updated); err != nil {
			return Workgraph{}, "", err
		}
		return updated, node.ID, nil
	}
	if !matchedTask {
		return Workgraph{}, "", fmt.Errorf("no matching workgraph node for run-link task_id %q", link.TaskID)
	}
	return Workgraph{}, "", fmt.Errorf("no dependency-ready matching workgraph node for run-link task_id %q", link.TaskID)
}

func (state WorkgraphState) MissingHandoffs(runLinks map[string]string) []string {
	missing := []string{}
	for _, taskID := range state.ReadyTaskIDs {
		if _, ok := runLinks[taskID]; !ok {
			missing = append(missing, taskID)
		}
	}
	return missing
}
