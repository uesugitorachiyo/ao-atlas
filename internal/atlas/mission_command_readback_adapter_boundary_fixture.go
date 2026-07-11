package atlas

import "fmt"

func BuildAtlasCommandReadbackAdapterBoundaryFixture() (AtlasCommandReadbackAdapterBoundaryFixture, error) {
	adapters := []AtlasCommandReadbackAdapterBoundaryAdapter{
		{Name: "mission_lifecycle_readback", DelegatesDecisionTo: "ao-mission", DomainDecisionReadOnly: true},
		{Name: "control_plane_observer_readback", DelegatesDecisionTo: "ao2-control-plane", DomainDecisionReadOnly: true},
	}
	fixture := AtlasCommandReadbackAdapterBoundaryFixture{
		Schema:                    AtlasCommandReadbackAdapterBoundaryFixtureContract,
		Status:                    "command_readback_adapter_boundary_ready",
		Adapters:                  adapters,
		AdapterCount:              len(adapters),
		DelegatesDomainDecisions:  true,
		DuplicatesDomainDecisions: false,
		PresentationOnly:          true,
		SchedulesWork:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		ClaimsAuthorityAdvance:    false,
		RSIRemainsDenied:          true,
	}
	if err := ValidateAtlasCommandReadbackAdapterBoundaryFixture(fixture); err != nil {
		return AtlasCommandReadbackAdapterBoundaryFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasCommandReadbackAdapterBoundaryFixture(fixture AtlasCommandReadbackAdapterBoundaryFixture) error {
	var errs []string
	requireContract(&errs, "command_readback_adapter_boundary_fixture", fixture.Schema, AtlasCommandReadbackAdapterBoundaryFixtureContract)
	if fixture.Status != "command_readback_adapter_boundary_ready" {
		errs = append(errs, "status must be command_readback_adapter_boundary_ready")
	}
	if fixture.AdapterCount != len(fixture.Adapters) {
		errs = append(errs, "adapter_count must match adapters")
	}
	if fixture.AdapterCount != 2 {
		errs = append(errs, "adapter_count must be 2")
	}
	if !fixture.DelegatesDomainDecisions {
		errs = append(errs, "delegates_domain_decisions must be true")
	}
	if fixture.DuplicatesDomainDecisions {
		errs = append(errs, "duplicates_domain_decisions must be false")
	}
	if !fixture.PresentationOnly {
		errs = append(errs, "presentation_only must be true")
	}
	seen := map[string]bool{}
	for i, adapter := range fixture.Adapters {
		prefix := fmt.Sprintf("adapters[%d]", i)
		requireField(&errs, prefix+".name", adapter.Name)
		requireField(&errs, prefix+".delegates_decision_to", adapter.DelegatesDecisionTo)
		if !adapter.DomainDecisionReadOnly {
			errs = append(errs, prefix+".domain_decision_read_only must be true")
		}
		seen[adapter.DelegatesDecisionTo] = true
	}
	for _, required := range []string{"ao-mission", "ao2-control-plane"} {
		if !seen[required] {
			errs = append(errs, "adapters must delegate to "+required)
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
