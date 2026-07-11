package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3RestartResumeSoak(nodeID, exactlyOncePath, killRestartPath, sourceReadbackPath, dashboardReadbackPath string) (AtlasMonth3RestartResumeSoak, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3RestartResumeSoak{}, fmt.Errorf("node id is required")
	}
	exactlyOnce, err := LoadJSON[AtlasExactlyOnceResumeAccountingFixture](exactlyOncePath)
	if err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	if err := ValidateAtlasExactlyOnceResumeAccountingFixture(exactlyOnce); err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	killRestart, err := LoadJSON[AtlasKillRestartReplayFixture](killRestartPath)
	if err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	if err := ValidateAtlasKillRestartReplayFixture(killRestart); err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	dashboard, err := LoadJSON[AtlasMonth3OperatorDashboardReadback](dashboardReadbackPath)
	if err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	if err := ValidateAtlasMonth3OperatorDashboardReadback(dashboard); err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	digests, err := digestMonth3RestartResumeInputs(exactlyOncePath, killRestartPath, sourceReadbackPath, dashboardReadbackPath)
	if err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	scenarios := append(month3RestartResumeScenarioNames(exactlyOnce), "kill_restart_replay")
	fixture := AtlasMonth3RestartResumeSoak{
		Schema:                     AtlasMonth3RestartResumeSoakContract,
		NodeID:                     nodeID,
		Status:                     "restart_resume_soak_ready",
		ExactlyOncePath:            publicArtifactRef(exactlyOncePath),
		ExactlyOnceDigest:          digests[0],
		KillRestartPath:            publicArtifactRef(killRestartPath),
		KillRestartDigest:          digests[1],
		SourceReadbackPath:         publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:       digests[2],
		DashboardReadbackPath:      publicArtifactRef(dashboardReadbackPath),
		DashboardReadbackDigest:    digests[3],
		Scenarios:                  scenarios,
		ScenarioCount:              len(scenarios),
		ExactlyOnceAccountingBound: exactlyOnce.ExactlyOnceNodeAccounting && !exactlyOnce.DuplicateHandoffDoubleCountAllowed,
		KillRestartReplayBound:     killRestart.KilledRunReplayed && killRestart.RestartReadbackBound,
		CheckpointRecoveryBound:    readback.CompletedNodes == 10 && dashboard.ReadyWorkVisible && dashboard.BlockerCount == 0,
		NoLostEvidence:             killRestart.NoLostEvidence,
		DuplicateMutationDetected:  killRestart.DuplicateMutationDetected,
		FalseCompletionDetected:    killRestart.FalseCompletionDetected,
		CompletedNodes:             readback.CompletedNodes,
		ReadyNodes:                 readback.ReadyNodes,
		NextExecutableNode:         readback.FirstExecutableNode,
		FinalResponseAllowed:       readback.FinalResponseAllowed,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     exactlyOnce.ClaimsAuthorityAdvance || killRestart.ClaimsAuthorityAdvance || dashboard.ClaimsAuthorityAdvance,
		RSIRemainsDenied:           exactlyOnce.RSIRemainsDenied && killRestart.RSIRemainsDenied && dashboard.RSIRemainsDenied && readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if !fixture.ExactlyOnceAccountingBound || !fixture.KillRestartReplayBound || !fixture.CheckpointRecoveryBound || !fixture.NoLostEvidence || fixture.DuplicateMutationDetected || fixture.FalseCompletionDetected {
		fixture.Status = "restart_resume_soak_failed"
	}
	if err := ValidateAtlasMonth3RestartResumeSoak(fixture); err != nil {
		return AtlasMonth3RestartResumeSoak{}, err
	}
	return fixture, nil
}

func digestMonth3RestartResumeInputs(paths ...string) ([]string, error) {
	digests := make([]string, 0, len(paths))
	for _, path := range paths {
		digest, err := digestTextFileWithNormalizedLineEndings(path)
		if err != nil {
			return nil, err
		}
		digests = append(digests, digest)
	}
	return digests, nil
}

func month3RestartResumeScenarioNames(fixture AtlasExactlyOnceResumeAccountingFixture) []string {
	names := make([]string, 0, len(fixture.Scenarios))
	for _, scenario := range fixture.Scenarios {
		names = append(names, scenario.Name)
	}
	return names
}

func ValidateAtlasMonth3RestartResumeSoak(fixture AtlasMonth3RestartResumeSoak) error {
	var errs []string
	requireContract(&errs, "month3_restart_resume_soak", fixture.Schema, AtlasMonth3RestartResumeSoakContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if !oneOf(fixture.Status, "restart_resume_soak_ready", "restart_resume_soak_failed") {
		errs = append(errs, "status must be restart_resume_soak_ready or restart_resume_soak_failed")
	}
	for field, value := range map[string]string{
		"exactly_once_path":       fixture.ExactlyOncePath,
		"kill_restart_path":       fixture.KillRestartPath,
		"source_readback_path":    fixture.SourceReadbackPath,
		"dashboard_readback_path": fixture.DashboardReadbackPath,
		"next_executable_node":    fixture.NextExecutableNode,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"exactly_once_digest":       fixture.ExactlyOnceDigest,
		"kill_restart_digest":       fixture.KillRestartDigest,
		"source_readback_digest":    fixture.SourceReadbackDigest,
		"dashboard_readback_digest": fixture.DashboardReadbackDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if fixture.ScenarioCount != len(fixture.Scenarios) || fixture.ScenarioCount != 4 {
		errs = append(errs, "scenario_count must match four restart/resume scenarios")
	}
	if !fixture.ExactlyOnceAccountingBound {
		errs = append(errs, "exactly_once_accounting_bound must be true")
	}
	if !fixture.KillRestartReplayBound {
		errs = append(errs, "kill_restart_replay_bound must be true")
	}
	if !fixture.CheckpointRecoveryBound {
		errs = append(errs, "checkpoint_recovery_bound must be true")
	}
	if !fixture.NoLostEvidence {
		errs = append(errs, "no_lost_evidence must be true")
	}
	if fixture.DuplicateMutationDetected {
		errs = append(errs, "duplicate_mutation_detected must be false")
	}
	if fixture.FalseCompletionDetected {
		errs = append(errs, "false_completion_detected must be false")
	}
	if fixture.CompletedNodes != 10 || fixture.ReadyNodes <= 0 {
		errs = append(errs, "readback counts must show node 10 completed with ready work remaining")
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3RestartResumeSoak(path string, fixture AtlasMonth3RestartResumeSoak) error {
	return WriteJSON(path, fixture)
}
