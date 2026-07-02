package atlas

import (
	"fmt"
	"path/filepath"
	"strings"
)

func buildBlueprintIntake(rules BlueprintCandidateRules) Intake {
	return Intake{
		ContractVersion: IntakeContract,
		ID:              rules.ProjectID + "-intake",
		TargetInstance:  rules.TargetInstance,
		BroadPrompt:     rules.Objective + " The work must pass through AO Atlas before AO Foundry gates and must remain dry-run readback only.",
		InstructionRefs: []string{"implementation-spec.md", "quality-profile.md", "candidate-rules.json"},
		FolderRefs:      []string{"excluded/"},
		Constraints:     append([]string{"Blueprint -> Atlas -> Foundry is mandatory"}, rules.SafetyLimits...),
	}
}

func buildBlueprintContextPack(packPath, candidateRulesPath string, rules BlueprintCandidateRules, digests map[string]string) (ContextPack, error) {
	sourceRefs := []SourceRef{}
	for _, ref := range rules.ContextRefs {
		key := contextDigestKey(ref)
		digest := digests[key]
		sourceRef := filepath.ToSlash(ref)
		if digest == "" {
			path := filepath.Join(packPath, ref)
			var err error
			digest, err = digestFile(path)
			if err != nil {
				return ContextPack{}, fmt.Errorf("context ref %s: %w", ref, err)
			}
		}
		if key == "candidate_rules" && strings.TrimSpace(candidateRulesPath) != "" {
			sourceRef = publicArtifactRef(candidateRulesPath)
		}
		sourceRefs = append(sourceRefs, SourceRef{Ref: sourceRef, Digest: digest})
	}
	pack := ContextPack{
		ContractVersion: ContextPackContract,
		ID:              rules.CandidateID + "-context-pack",
		TaskID:          rules.CandidateID + "-task",
		BudgetBytes:     8192,
		SourceRefs:      sourceRefs,
		Summaries: []string{
			"Blueprint pack is authorized for Atlas compile into Foundry import material.",
			"Mutation class " + rules.MutationClass + " remains dry-run/readback only.",
		},
		Assumptions:     []string{"AO Blueprint owns requirements authorization.", "AO Atlas does not schedule or execute work."},
		Exclusions:      []string{"live code mutation", "provider calls", "private local paths outside excluded placeholders"},
		MissingProtocol: "Return to AO Blueprint for missing requirements or authorization before widening scope.",
	}
	if err := ValidateContextPack(pack, 0); err != nil {
		return ContextPack{}, err
	}
	return pack, nil
}

func buildBlueprintFactoryTask(rules BlueprintCandidateRules, pack ContextPack) FactoryTask {
	return FactoryTask{
		ContractVersion:   FactoryTaskContract,
		ID:                rules.CandidateID + "-task",
		Objective:         rules.Objective,
		TargetFactoryRepo: rules.TargetFactoryRepo,
		FactoryFolder:     rules.FactoryFolder,
		MutationClass:     rules.MutationClass,
		Acceptance:        append([]string(nil), rules.Acceptance...),
		NonGoals:          append([]string(nil), rules.NonGoals...),
		WriteScope:        append([]string(nil), rules.WriteScope...),
		RequiredGates:     append([]string(nil), rules.RequiredGates...),
		RollbackScope:     append([]string(nil), rules.RollbackScope...),
		Verification:      append([]string(nil), rules.Verification...),
		RequiredEvidence:  append([]string(nil), rules.RequiredEvidence...),
		SafetyLimits:      append([]string(nil), rules.SafetyLimits...),
		AuthorityBoundary: rules.AuthorityBoundary,
		DependencyRefs:    append([]string(nil), rules.DependencyRefs...),
		ContextPackRefs:   []string{filepath.ToSlash(filepath.Join("context-packs", pack.ID+".json"))},
	}
}

func buildBlueprintCandidateSelection(rules BlueprintCandidateRules, node WorkgraphNode, digests map[string]string) BlueprintCandidateSelection {
	return BlueprintCandidateSelection{
		ContractVersion:     BlueprintCandidateSelectionContract,
		ID:                  rules.CandidateID + "-selection",
		ProjectID:           rules.ProjectID,
		Status:              "ready",
		SelectedCandidateID: rules.CandidateID,
		MutationClass:       rules.MutationClass,
		TargetFactoryRepo:   rules.TargetFactoryRepo,
		WorkgraphID:         rules.WorkgraphID,
		NodeID:              node.ID,
		TaskID:              node.FactoryTask.ID,
		RequiredGates:       append([]string(nil), rules.RequiredGates...),
		RequiredEvidence:    append([]string(nil), rules.RequiredEvidence...),
		SafetyLimits:        append([]string(nil), rules.SafetyLimits...),
		Digests:             copyStringMap(digests),
		SchedulesWork:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		SafeToExecute:       false,
		LiveExecutionProven: false,
	}
}
