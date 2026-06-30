package atlas

import "fmt"

// RepairNextComplexMissionNode consumes completed Foundry evidence and advances
// the next dependency-cleared complex_repo_mutation node from blocked to ready.
// It does not schedule or execute work; it only repairs Atlas-owned workgraph
// evidence so Foundry can gate exactly one node.
func RepairNextComplexMissionNode(workgraph Workgraph, completedLink RunLink) (Workgraph, WorkgraphNode, bool, error) {
	continued, _, err := CompleteWorkgraph(workgraph, completedLink)
	if err != nil {
		return Workgraph{}, WorkgraphNode{}, false, err
	}
	if ready, ok, err := singleReadyComplexNode(continued); ok || err != nil {
		return continued, ready, ok, err
	}

	statusByID := map[string]string{}
	for _, node := range continued.Nodes {
		statusByID[node.ID] = node.Status
	}
	for i, node := range continued.Nodes {
		if node.Status != "blocked" {
			continue
		}
		if node.FactoryTask.MutationClass != "complex_repo_mutation" {
			continue
		}
		if !dependenciesCompleted(node, statusByID) {
			continue
		}
		repaired := continued
		repaired.Nodes = append([]WorkgraphNode(nil), continued.Nodes...)
		repaired.Nodes[i].Status = "ready"
		repaired.Nodes[i].Blockers = nil
		repaired.Nodes[i].FactoryTask.RequiredEvidence = bindSafeToExecuteTrue(node.FactoryTask.RequiredEvidence)
		if err := ValidateWorkgraph(repaired); err != nil {
			return Workgraph{}, WorkgraphNode{}, false, err
		}
		return repaired, repaired.Nodes[i], true, nil
	}
	return continued, WorkgraphNode{}, false, nil
}

func singleReadyComplexNode(workgraph Workgraph) (WorkgraphNode, bool, error) {
	readyNodes := []WorkgraphNode{}
	for _, node := range workgraph.Nodes {
		if node.Status != "ready" {
			continue
		}
		readyNodes = append(readyNodes, node)
	}
	if len(readyNodes) == 0 {
		return WorkgraphNode{}, false, nil
	}
	if len(readyNodes) > 1 {
		return WorkgraphNode{}, false, fmt.Errorf("exactly one ready executable node is allowed")
	}
	ready := readyNodes[0]
	if ready.FactoryTask.MutationClass != "complex_repo_mutation" {
		return WorkgraphNode{}, false, fmt.Errorf("ready node mutation_class must be complex_repo_mutation")
	}
	return ready, true, nil
}

func dependenciesCompleted(node WorkgraphNode, statusByID map[string]string) bool {
	for _, dep := range node.Dependencies {
		if statusByID[dep] != "completed" {
			return false
		}
	}
	return true
}

func bindSafeToExecuteTrue(evidence []string) []string {
	bound := make([]string, 0, len(evidence)+1)
	sawTrue := false
	for _, item := range evidence {
		switch item {
		case "safe_to_execute:false":
			if !sawTrue {
				bound = append(bound, "safe_to_execute:true")
				sawTrue = true
			}
		case "safe_to_execute:true":
			if !sawTrue {
				bound = append(bound, item)
				sawTrue = true
			}
		default:
			bound = append(bound, item)
		}
	}
	if !sawTrue {
		bound = append(bound, "safe_to_execute:true")
	}
	return bound
}
