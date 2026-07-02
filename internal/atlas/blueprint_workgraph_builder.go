package atlas

func buildBlueprintWorkgraph(rules BlueprintCandidateRules, task FactoryTask) (Workgraph, error) {
	workgraph := Workgraph{
		ContractVersion: WorkgraphContract,
		ID:              rules.WorkgraphID,
		TargetInstance:  rules.TargetInstance,
		Nodes: []WorkgraphNode{{
			ID:           rules.CandidateID + "-node",
			Status:       "ready",
			FactoryTask:  task,
			Dependencies: []string{},
			Blockers:     []string{},
			StitchTask:   false,
		}},
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return Workgraph{}, err
	}
	return workgraph, nil
}
