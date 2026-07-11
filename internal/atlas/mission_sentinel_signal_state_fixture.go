package atlas

import "fmt"

func BuildAtlasSentinelSignalStateFixture() (AtlasSentinelSignalStateFixture, error) {
	signals := []string{"ci", "runtime", "policy", "evidence_freshness"}
	states := []string{"stale", "pending", "pass", "failure"}
	var matrix []AtlasSentinelSignalStateMatrixRow
	for _, signal := range signals {
		for _, state := range states {
			matrix = append(matrix, AtlasSentinelSignalStateMatrixRow{
				Signal:  signal,
				State:   state,
				Verdict: sentinelSignalStateVerdict(state),
			})
		}
	}
	fixture := AtlasSentinelSignalStateFixture{
		Schema:                 AtlasSentinelSignalStateFixtureContract,
		Status:                 "signal_states_ready",
		Signals:                signals,
		States:                 states,
		SignalCount:            len(signals),
		StateCount:             len(states),
		Matrix:                 matrix,
		MatrixCount:            len(matrix),
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasSentinelSignalStateFixture(fixture); err != nil {
		return AtlasSentinelSignalStateFixture{}, err
	}
	return fixture, nil
}

func sentinelSignalStateVerdict(state string) string {
	switch state {
	case "pass":
		return "allow_continue"
	case "pending":
		return "wait"
	case "stale":
		return "refresh_required"
	default:
		return "block"
	}
}

func ValidateAtlasSentinelSignalStateFixture(fixture AtlasSentinelSignalStateFixture) error {
	var errs []string
	requireContract(&errs, "sentinel_signal_state_fixture", fixture.Schema, AtlasSentinelSignalStateFixtureContract)
	if fixture.Status != "signal_states_ready" {
		errs = append(errs, "status must be signal_states_ready")
	}
	if fixture.SignalCount != len(fixture.Signals) {
		errs = append(errs, "signal_count must match signals")
	}
	if fixture.StateCount != len(fixture.States) {
		errs = append(errs, "state_count must match states")
	}
	if fixture.SignalCount != 4 || !containsAll(fixture.Signals, []string{"ci", "runtime", "policy", "evidence_freshness"}) {
		errs = append(errs, "signals must cover ci, runtime, policy, and evidence_freshness")
	}
	if fixture.StateCount != 4 || !containsAll(fixture.States, []string{"stale", "pending", "pass", "failure"}) {
		errs = append(errs, "states must cover stale, pending, pass, and failure")
	}
	if fixture.MatrixCount != len(fixture.Matrix) {
		errs = append(errs, "matrix_count must match matrix")
	}
	if fixture.MatrixCount != fixture.SignalCount*fixture.StateCount {
		errs = append(errs, "matrix_count must cover every signal/state pair")
	}
	seen := map[string]bool{}
	for i, row := range fixture.Matrix {
		prefix := fmt.Sprintf("matrix[%d]", i)
		requireField(&errs, prefix+".signal", row.Signal)
		requireField(&errs, prefix+".state", row.State)
		requireField(&errs, prefix+".verdict", row.Verdict)
		if !containsAll(fixture.Signals, []string{row.Signal}) {
			errs = append(errs, prefix+".signal must be declared in signals")
		}
		if !containsAll(fixture.States, []string{row.State}) {
			errs = append(errs, prefix+".state must be declared in states")
		}
		key := row.Signal + "\x00" + row.State
		if seen[key] {
			errs = append(errs, prefix+".signal/state pair must be unique")
		}
		seen[key] = true
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
