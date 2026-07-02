package atlas

type blueprintMutationModelLoadResult struct {
	Missing  []string
	Blockers []string
}

func loadBlueprintMutationModel(paths BlueprintImportPaths, rules BlueprintCandidateRules, digests map[string]string) blueprintMutationModelLoadResult {
	result := blueprintMutationModelLoadResult{}
	var mutationModel MutationClassModel
	if err := readJSONIfPossible(paths.MutationClassesPath, &mutationModel); err != nil {
		result.Missing = append(result.Missing, "mutation_class_model")
		result.Blockers = append(result.Blockers, "provide the AO Atlas mutation class model")
		return result
	}
	if err := ValidateMutationClassModel(mutationModel); err != nil {
		result.Missing = append(result.Missing, "mutation_class_model")
		result.Blockers = append(result.Blockers, "repair mutation class model: "+err.Error())
		return result
	}
	if !mutationModelIncludes(mutationModel, rules.MutationClass) {
		result.Missing = append(result.Missing, "mutation_class_scope")
		result.Blockers = append(result.Blockers, "mutation class model must include "+rules.MutationClass)
		return result
	}
	if digest, err := digestFile(paths.MutationClassesPath); err == nil {
		digests["mutation_class_model"] = digest
	}
	return result
}
