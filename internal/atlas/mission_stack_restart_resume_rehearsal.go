package atlas

import "fmt"

func BuildAtlasStackRestartResumeRehearsal() (AtlasStackRestartResumeRehearsal, error) {
	components := []AtlasStackRestartResumeComponent{
		{
			Repo:           "ao-mission",
			Role:           "mission_ledger_recovery",
			ResumeSignal:   "checkpoint_fresh",
			ReplayRequired: true,
			ReadyForResume: true,
		},
		{
			Repo:           "ao-atlas",
			Role:           "workgraph_next_node_recovery",
			ResumeSignal:   "exact_next_action_present",
			ReplayRequired: true,
			ReadyForResume: true,
		},
		{
			Repo:           "ao-foundry",
			Role:           "safe_next_work_recovery",
			ResumeSignal:   "single_active_node_bound",
			ReplayRequired: true,
			ReadyForResume: true,
		},
	}
	fixture := AtlasStackRestartResumeRehearsal{
		Schema:                    AtlasStackRestartResumeRehearsalContract,
		Status:                    "restart_resume_rehearsal_ready",
		Wave:                      "ao-stack-month6-recommendations",
		Components:                components,
		ComponentCount:            len(components),
		MissionCheckpointBound:    true,
		AtlasWorkgraphBound:       true,
		FoundrySafeNextWorkBound:  true,
		NoLostEvidence:            true,
		SingleActiveNodePreserved: true,
		FinalResponseAllowed:      false,
		SchedulesWork:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		ClaimsAuthorityAdvance:    false,
		RSIRemainsDenied:          true,
	}
	if err := ValidateAtlasStackRestartResumeRehearsal(fixture); err != nil {
		return AtlasStackRestartResumeRehearsal{}, err
	}
	return fixture, nil
}

func ValidateAtlasStackRestartResumeRehearsal(fixture AtlasStackRestartResumeRehearsal) error {
	var errs []string
	requireContract(&errs, "stack_restart_resume_rehearsal", fixture.Schema, AtlasStackRestartResumeRehearsalContract)
	if fixture.Status != "restart_resume_rehearsal_ready" {
		errs = append(errs, "status must be restart_resume_rehearsal_ready")
	}
	requireField(&errs, "wave", fixture.Wave)
	if fixture.ComponentCount != len(fixture.Components) || fixture.ComponentCount != 3 {
		errs = append(errs, "component_count must match three stack components")
	}
	seen := map[string]bool{}
	for i, component := range fixture.Components {
		prefix := fmt.Sprintf("components[%d]", i)
		requireField(&errs, prefix+".repo", component.Repo)
		requireField(&errs, prefix+".role", component.Role)
		requireField(&errs, prefix+".resume_signal", component.ResumeSignal)
		seen[component.Repo] = true
		if !component.ReplayRequired {
			errs = append(errs, prefix+".replay_required must be true")
		}
		if !component.ReadyForResume {
			errs = append(errs, prefix+".ready_for_resume must be true")
		}
	}
	for _, repo := range []string{"ao-mission", "ao-atlas", "ao-foundry"} {
		if !seen[repo] {
			errs = append(errs, "missing restart/resume component "+repo)
		}
	}
	if !fixture.MissionCheckpointBound {
		errs = append(errs, "mission_checkpoint_bound must be true")
	}
	if !fixture.AtlasWorkgraphBound {
		errs = append(errs, "atlas_workgraph_bound must be true")
	}
	if !fixture.FoundrySafeNextWorkBound {
		errs = append(errs, "foundry_safe_next_work_bound must be true")
	}
	if !fixture.NoLostEvidence {
		errs = append(errs, "no_lost_evidence must be true")
	}
	if !fixture.SingleActiveNodePreserved {
		errs = append(errs, "single_active_node_preserved must be true")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
