package atlas

type blueprintInstanceLoadResult struct {
	Missing  []string
	Blockers []string
}

func loadBlueprintInstance(paths BlueprintImportPaths, rules BlueprintCandidateRules, digests map[string]string) blueprintInstanceLoadResult {
	result := blueprintInstanceLoadResult{}
	var instance Instance
	if err := readJSONIfPossible(paths.InstancePath, &instance); err != nil {
		result.Missing = append(result.Missing, "stack_instance")
		result.Blockers = append(result.Blockers, "provide an AO Atlas stack instance")
		return result
	}
	if err := ValidateInstance(instance); err != nil {
		result.Missing = append(result.Missing, "stack_instance")
		result.Blockers = append(result.Blockers, "repair stack instance: "+err.Error())
		return result
	}
	if rules.TargetInstance != "" && instance.ID != rules.TargetInstance {
		result.Missing = append(result.Missing, "stack_instance_scope")
		result.Blockers = append(result.Blockers, "stack instance id must match candidate target_instance")
		return result
	}
	if digest, err := digestFile(paths.InstancePath); err == nil {
		digests["stack_instance"] = digest
	}
	return result
}
