package atlas

func ValidateBlueprintCandidateRules(rules BlueprintCandidateRules) error {
	var errs []string
	if rules.SchemaVersion != BlueprintCandidateRulesContract {
		errs = append(errs, "schema_version must be "+BlueprintCandidateRulesContract)
	}
	requireField(&errs, "project_id", rules.ProjectID)
	requireField(&errs, "target_instance", rules.TargetInstance)
	requireField(&errs, "workgraph_id", rules.WorkgraphID)
	requireField(&errs, "candidate_id", rules.CandidateID)
	requireField(&errs, "mutation_class", rules.MutationClass)
	if !requiredMutationClassNames()[rules.MutationClass] {
		errs = append(errs, "mutation_class must be one of the required mutation classes")
	}
	requireField(&errs, "target_factory_repo", rules.TargetFactoryRepo)
	requireField(&errs, "factory_folder", rules.FactoryFolder)
	requireField(&errs, "objective", rules.Objective)
	requireList(&errs, "acceptance_criteria", rules.Acceptance)
	requireList(&errs, "non_goals", rules.NonGoals)
	requireList(&errs, "write_scope", rules.WriteScope)
	requireList(&errs, "rollback_scope", rules.RollbackScope)
	requireList(&errs, "required_gates", rules.RequiredGates)
	requireList(&errs, "verification_commands", rules.Verification)
	requireList(&errs, "required_evidence", rules.RequiredEvidence)
	requireList(&errs, "safety_limits", rules.SafetyLimits)
	requireField(&errs, "authority_boundary", rules.AuthorityBoundary)
	requireList(&errs, "context_refs", rules.ContextRefs)
	checkPublicPath(&errs, "target_factory_repo", rules.TargetFactoryRepo, false)
	checkPublicPath(&errs, "factory_folder", rules.FactoryFolder, false)
	checkPublicStrings(&errs, "write_scope", rules.WriteScope, true)
	checkPublicStrings(&errs, "rollback_scope", rules.RollbackScope, true)
	checkPublicStrings(&errs, "required_gates", rules.RequiredGates, true)
	checkPublicStrings(&errs, "verification_commands", rules.Verification, true)
	checkPublicStrings(&errs, "required_evidence", rules.RequiredEvidence, true)
	checkPublicStrings(&errs, "safety_limits", rules.SafetyLimits, true)
	checkPublicStrings(&errs, "context_refs", rules.ContextRefs, true)
	return joinErrors(errs)
}
