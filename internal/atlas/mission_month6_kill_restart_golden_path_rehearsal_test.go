package atlas

import (
	"path/filepath"
	"testing"
)

func TestMonth6KillRestartGoldenPathRehearsalFixtureBindsDryRunRestarts(t *testing.T) {
	root := repoRoot(t)
	fixturePath := filepath.Join(root, "docs", "evidence", "ao-stack-month6-beta-launch-v01", "nodes", "month6-recommendation-15-kill-restart-golden-path-rehearsal", "kill-restart-golden-path-rehearsal.json")

	fixture := mustLoadJSON[AtlasMonth6KillRestartGoldenPathRehearsal](t, fixturePath)
	if err := ValidateAtlasMonth6KillRestartGoldenPathRehearsal(fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Status != "kill_restart_rehearsal_bound" ||
		fixture.SourceRecommendationRank != 15 ||
		fixture.RehearsalCount != 3 ||
		len(fixture.Rehearsals) != 3 ||
		!fixture.FixtureOnly ||
		!fixture.DryRunOnly ||
		!fixture.KilledRunReplayed ||
		!fixture.RestartReadbackBound ||
		!fixture.NoLostEvidence ||
		fixture.DuplicateMutationDetected ||
		fixture.FalseCompletionDetected ||
		fixture.ProviderCallsAllowed ||
		fixture.CredentialUseAllowed ||
		fixture.LiveMutationAllowed ||
		fixture.ReleaseOrPublishAllowed ||
		fixture.ApprovalGranted ||
		!fixture.NoPromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied ||
		fixture.SafeToExecute ||
		fixture.ExecutesWork ||
		fixture.ApprovesWork ||
		fixture.MutatesRepositories {
		t.Fatalf("Month 6 kill-restart rehearsal fixture lost safety state: %#v", fixture)
	}
	for _, rehearsal := range fixture.Rehearsals {
		if !rehearsal.KilledAfterCheckpoint ||
			!rehearsal.EventIndexRebuilt ||
			!rehearsal.ResumeSelectedSameNextNode ||
			!rehearsal.NoLostEvidence ||
			rehearsal.DuplicateMutationDetected ||
			rehearsal.FalseCompletionDetected ||
			rehearsal.ProviderCallsAllowed ||
			rehearsal.SafeToExecute ||
			rehearsal.ExecutesWork ||
			rehearsal.MutatesRepository {
			t.Fatalf("unsafe kill-restart rehearsal row: %#v", rehearsal)
		}
	}
	validator, err := validateRecommendationEvidenceTypedFile(fixturePath, AtlasMonth6KillRestartGoldenPathRehearsalContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:month6-kill-restart-golden-path-rehearsal" {
		t.Fatalf("expected typed Month 6 kill-restart rehearsal validator, got %s", validator)
	}
}

func TestMonth6KillRestartGoldenPathRehearsalRejectsProviderExecution(t *testing.T) {
	fixture := validMonth6KillRestartGoldenPathRehearsalForTest()
	fixture.ProviderCallsAllowed = true
	fixture.Rehearsals[0].ProviderCallsAllowed = true
	if err := ValidateAtlasMonth6KillRestartGoldenPathRehearsal(fixture); err == nil {
		t.Fatal("expected provider execution to be rejected")
	}
}

func validMonth6KillRestartGoldenPathRehearsalForTest() AtlasMonth6KillRestartGoldenPathRehearsal {
	return AtlasMonth6KillRestartGoldenPathRehearsal{
		Schema:                   AtlasMonth6KillRestartGoldenPathRehearsalContract,
		NodeID:                   "month6-recommendation-15-kill-restart-golden-path-rehearsal",
		Status:                   "kill_restart_rehearsal_bound",
		SourceRecommendationRank: 15,
		SourceRecommendationTask: "Run kill-restart golden path rehearsal without provider execution",
		SafetyGate:               "planning_only_no_provider_no_release",
		SourceRehearsalRef:       "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-03-golden-path-dry-run-rehearsals/golden-path-dry-run-rehearsals.json",
		InterruptionMarker:       "after_foundry_import_before_ao2_execution",
		RestartResumeMarker:      "mission_event_index_rebuilt_before_next_node_selection",
		RehearsalCount:           3,
		Rehearsals: []AtlasMonth6KillRestartGoldenPathRehearsalRow{
			validMonth6KillRestartGoldenPathRehearsalRowForTest("external-non-ao-cli-sample"),
			validMonth6KillRestartGoldenPathRehearsalRowForTest("external-non-ao-library-sample"),
			validMonth6KillRestartGoldenPathRehearsalRowForTest("external-non-ao-docs-sample"),
		},
		FixtureOnly:               true,
		DryRunOnly:                true,
		KilledRunReplayed:         true,
		RestartReadbackBound:      true,
		NoLostEvidence:            true,
		DuplicateMutationDetected: false,
		FalseCompletionDetected:   false,
		NoPromotionRequested:      true,
		ClaimsAuthorityAdvance:    false,
		RSIRemainsDenied:          true,
		SafeToExecute:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		MutatesRepositories:       false,
		ProviderCallsAllowed:      false,
		CredentialUseAllowed:      false,
		LiveMutationAllowed:       false,
		ReleaseOrPublishAllowed:   false,
		ApprovalGranted:           false,
	}
}

func validMonth6KillRestartGoldenPathRehearsalRowForTest(id string) AtlasMonth6KillRestartGoldenPathRehearsalRow {
	return AtlasMonth6KillRestartGoldenPathRehearsalRow{
		ID:                         id,
		KilledAfterCheckpoint:      true,
		RestartReadbackRef:         "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-15-kill-restart-golden-path-rehearsal/restart-readbacks/" + id + ".json",
		RollbackReceiptRef:         "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-03-golden-path-dry-run-rehearsals/rollback-receipts/" + id + ".json",
		CommandReadbackRef:         "docs/evidence/ao-stack-month6-beta-launch-v01/nodes/month6-recommendation-03-golden-path-dry-run-rehearsals/command-readbacks/" + id + ".json",
		EventIndexRebuilt:          true,
		ResumeSelectedSameNextNode: true,
		NoLostEvidence:             true,
		DuplicateMutationDetected:  false,
		FalseCompletionDetected:    false,
		ProviderCallsAllowed:       false,
		SafeToExecute:              false,
		ExecutesWork:               false,
		MutatesRepository:          false,
	}
}
