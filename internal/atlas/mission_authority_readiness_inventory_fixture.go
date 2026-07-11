package atlas

import "fmt"

func BuildAtlasAuthorityReadinessInventoryFixture() (AtlasAuthorityReadinessInventoryFixture, error) {
	inputs := []AtlasAuthorityReadinessInventoryInput{
		{Name: "stack_lockfile", Source: "stack-lockfile", Required: true},
		{Name: "contract_owner_registry", Source: "contract-owner-inputs", Required: true},
	}
	sections := []AtlasAuthorityReadinessInventorySection{
		{Name: "current_authority_statement", Generated: true, SourceInput: "stack_lockfile", CurrentTruth: true},
		{Name: "readiness_inventory", Generated: true, SourceInput: "contract_owner_registry", CurrentTruth: true},
	}
	fixture := AtlasAuthorityReadinessInventoryFixture{
		Schema:                     AtlasAuthorityReadinessInventoryFixtureContract,
		Status:                     "authority_readiness_inventory_ready",
		Inputs:                     inputs,
		InputCount:                 len(inputs),
		Sections:                   sections,
		SectionCount:               len(sections),
		GeneratedFromInputs:        true,
		CopiedCampaignProseAllowed: false,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
		RSIRemainsDenied:           true,
	}
	if err := ValidateAtlasAuthorityReadinessInventoryFixture(fixture); err != nil {
		return AtlasAuthorityReadinessInventoryFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasAuthorityReadinessInventoryFixture(fixture AtlasAuthorityReadinessInventoryFixture) error {
	var errs []string
	requireContract(&errs, "authority_readiness_inventory_fixture", fixture.Schema, AtlasAuthorityReadinessInventoryFixtureContract)
	if fixture.Status != "authority_readiness_inventory_ready" {
		errs = append(errs, "status must be authority_readiness_inventory_ready")
	}
	if fixture.InputCount != len(fixture.Inputs) {
		errs = append(errs, "input_count must match inputs")
	}
	if fixture.InputCount != 2 {
		errs = append(errs, "input_count must be 2")
	}
	if fixture.SectionCount != len(fixture.Sections) {
		errs = append(errs, "section_count must match sections")
	}
	if fixture.SectionCount != 2 {
		errs = append(errs, "section_count must be 2")
	}
	if !fixture.GeneratedFromInputs {
		errs = append(errs, "generated_from_inputs must be true")
	}
	if fixture.CopiedCampaignProseAllowed {
		errs = append(errs, "copied_campaign_prose_allowed must be false")
	}
	seenInputs := map[string]bool{}
	for i, input := range fixture.Inputs {
		prefix := fmt.Sprintf("inputs[%d]", i)
		requireField(&errs, prefix+".name", input.Name)
		requireField(&errs, prefix+".source", input.Source)
		if !input.Required {
			errs = append(errs, prefix+".required must be true")
		}
		seenInputs[input.Name] = true
	}
	for _, required := range []string{"stack_lockfile", "contract_owner_registry"} {
		if !seenInputs[required] {
			errs = append(errs, "inputs must include "+required)
		}
	}
	for i, section := range fixture.Sections {
		prefix := fmt.Sprintf("sections[%d]", i)
		requireField(&errs, prefix+".name", section.Name)
		requireField(&errs, prefix+".source_input", section.SourceInput)
		if !section.Generated {
			errs = append(errs, prefix+".generated must be true")
		}
		if !section.CurrentTruth {
			errs = append(errs, prefix+".current_truth must be true")
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
