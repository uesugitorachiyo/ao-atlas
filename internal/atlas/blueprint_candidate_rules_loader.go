package atlas

import (
	"path/filepath"
	"strings"
)

type blueprintCandidateRulesLoadResult struct {
	Rules     BlueprintCandidateRules
	Record    BlueprintImport
	RulesPath string
	Missing   []string
	Blockers  []string
}

func loadBlueprintCandidateRules(paths BlueprintImportPaths, record BlueprintImport, digests map[string]string) blueprintCandidateRulesLoadResult {
	var rules BlueprintCandidateRules
	rulesPath := filepath.Join(paths.PackPath, "candidate-rules.json")
	if strings.TrimSpace(paths.CandidateRulesPath) != "" {
		rulesPath = paths.CandidateRulesPath
	}
	result := blueprintCandidateRulesLoadResult{
		Rules:     rules,
		Record:    record,
		RulesPath: rulesPath,
	}
	if err := readJSONIfPossible(rulesPath, &rules); err != nil {
		result.Missing = append(result.Missing, "candidate_rules")
		result.Blockers = append(result.Blockers, "add candidate-rules.json to the Blueprint pack")
		return result
	}
	if err := ValidateBlueprintCandidateRules(rules); err != nil {
		result.Missing = append(result.Missing, "candidate_rules")
		result.Blockers = append(result.Blockers, "repair candidate-rules.json: "+err.Error())
		return result
	}
	record.ID = rules.WorkgraphID + "-blueprint-import"
	record.ProjectID = rules.ProjectID
	record.TargetInstance = rules.TargetInstance
	record.WorkgraphID = rules.WorkgraphID
	record.MutationClass = rules.MutationClass
	record.SafetyLimits = append([]string(nil), rules.SafetyLimits...)
	if digest, err := digestFile(rulesPath); err == nil {
		digests["candidate_rules"] = digest
	}
	result.Rules = rules
	result.Record = record
	return result
}
