package atlas

import "fmt"

func BuildAtlasReplayableStatePacketFixture() (AtlasReplayableStatePacketFixture, error) {
	states := []AtlasReplayableStatePacketState{
		{Name: "handoff", CountsAsCompleted: false, Replayable: true},
		{Name: "active", CountsAsCompleted: false, Replayable: true},
		{Name: "completed", CountsAsCompleted: true, Replayable: true},
		{Name: "denied", CountsAsCompleted: false, Replayable: true},
	}
	fixture := AtlasReplayableStatePacketFixture{
		Schema:                   AtlasReplayableStatePacketFixtureContract,
		Status:                   "replayable_state_packet_ready",
		States:                   states,
		StateCount:               len(states),
		Replayable:               true,
		HandoffCountsAsCompleted: false,
		SchedulesWork:            false,
		ExecutesWork:             false,
		ApprovesWork:             false,
		ClaimsAuthorityAdvance:   false,
		RSIRemainsDenied:         true,
	}
	if err := ValidateAtlasReplayableStatePacketFixture(fixture); err != nil {
		return AtlasReplayableStatePacketFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasReplayableStatePacketFixture(fixture AtlasReplayableStatePacketFixture) error {
	var errs []string
	requireContract(&errs, "replayable_state_packet_fixture", fixture.Schema, AtlasReplayableStatePacketFixtureContract)
	if fixture.Status != "replayable_state_packet_ready" {
		errs = append(errs, "status must be replayable_state_packet_ready")
	}
	if fixture.StateCount != len(fixture.States) {
		errs = append(errs, "state_count must match states")
	}
	if fixture.StateCount != 4 {
		errs = append(errs, "state_count must be 4")
	}
	if !fixture.Replayable {
		errs = append(errs, "replayable must be true")
	}
	if fixture.HandoffCountsAsCompleted {
		errs = append(errs, "handoff_counts_as_completed must be false")
	}
	seen := map[string]bool{}
	for i, state := range fixture.States {
		prefix := fmt.Sprintf("states[%d]", i)
		requireField(&errs, prefix+".name", state.Name)
		if !state.Replayable {
			errs = append(errs, prefix+".replayable must be true")
		}
		if state.Name == "handoff" && state.CountsAsCompleted {
			errs = append(errs, prefix+".counts_as_completed must be false for handoff")
		}
		if state.Name == "completed" && !state.CountsAsCompleted {
			errs = append(errs, prefix+".counts_as_completed must be true for completed")
		}
		seen[state.Name] = true
	}
	for _, required := range []string{"handoff", "active", "completed", "denied"} {
		if !seen[required] {
			errs = append(errs, "states must include "+required)
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
