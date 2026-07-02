package atlas

type blueprintCompileSourceInputs struct {
	Paths      BlueprintImportPaths
	Record     BlueprintImport
	Digests    map[string]string
	PackDigest string
	PackErr    error
}

type blueprintCompileSourceLoad struct {
	Record     BlueprintImport
	Rules      BlueprintCandidateRules
	Missing    []string
	Blockers   []string
	AuthDigest string
}

func loadBlueprintCompileSources(inputs blueprintCompileSourceInputs) blueprintCompileSourceLoad {
	record := inputs.Record
	missing := []string{}
	blockers := []string{}
	if inputs.PackErr != nil {
		missing = append(missing, "blueprint_pack")
		blockers = append(blockers, "provide a readable AO Blueprint pack")
	}

	rulesResult := loadBlueprintCandidateRules(inputs.Paths, record, inputs.Digests)
	rules := rulesResult.Rules
	record = rulesResult.Record
	missing = append(missing, rulesResult.Missing...)
	blockers = append(blockers, rulesResult.Blockers...)

	requiredResult := loadBlueprintRequiredArtifacts(inputs.Paths, inputs.Digests)
	missing = append(missing, requiredResult.Missing...)
	blockers = append(blockers, requiredResult.Blockers...)

	instanceResult := loadBlueprintInstance(inputs.Paths, rules, inputs.Digests)
	missing = append(missing, instanceResult.Missing...)
	blockers = append(blockers, instanceResult.Blockers...)

	mutationModelResult := loadBlueprintMutationModel(inputs.Paths, rules, inputs.Digests)
	missing = append(missing, mutationModelResult.Missing...)
	blockers = append(blockers, mutationModelResult.Blockers...)

	authorizationResult := loadBlueprintAuthorization(inputs.Paths, record, rules, inputs.PackDigest, inputs.Digests)
	record = authorizationResult.Record
	missing = append(missing, authorizationResult.Missing...)
	blockers = append(blockers, authorizationResult.Blockers...)

	return blueprintCompileSourceLoad{
		Record:     record,
		Rules:      rules,
		Missing:    missing,
		Blockers:   blockers,
		AuthDigest: authorizationResult.AuthDigest,
	}
}
