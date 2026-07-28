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

func NextReadyNode(workgraph Workgraph) (WorkgraphNode, bool) {
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return WorkgraphNode{}, false
	}
	return state.NextReadyNode()
}

func CompleteWorkgraph(workgraph Workgraph, link RunLink) (Workgraph, string, error) {
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return Workgraph{}, "", err
	}
	return state.CompleteWithRunLink(link)
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
			ContextPackRefs:   append([]string(nil), source.ContextPackRefs...),
		}
		plan := WorkgraphRepairPlan{
			ContractVersion:     WorkgraphRepairPlanContract,
			ID:                  workgraph.ID + "-" + source.ID + "-repair-plan",
			TaskID:              source.ID,
			Status:              "repair_required",
			SourceRunLinkStatus: link.Status,
			Reason:              "run-link status " + link.Status + " did not complete the task; emit a bounded repair task for Foundry scheduling",
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
