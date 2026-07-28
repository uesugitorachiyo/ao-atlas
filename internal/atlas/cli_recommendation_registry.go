package atlas

import (
	"fmt"
	"io"
	"strings"
)

type missionRecommendationCommand struct {
	name             string
	run              func([]string, io.Writer) error
	commandClass     string
	recordsRunLedger bool
}

const (
	missionRecommendationCommandClassPlanningOnly    = "planning_only"
	missionRecommendationCommandClassMutationCapable = "mutation_capable"
)

func missionRecommendationCommandRegistry() []missionRecommendationCommand {
	return []missionRecommendationCommand{
		{name: "import", run: runMissionRecommendationsImport, commandClass: missionRecommendationCommandClassMutationCapable},
		{name: "export-next-wave", run: runMissionRecommendationsExportNextWave, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "export-refactoring-wave", run: runMissionRecommendationsExportRefactoringWave, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "next-track", run: runMissionRecommendationsNextTrack, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "consumed-ledger", run: runMissionRecommendationsConsumedLedger, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "track-registry", run: runMissionRecommendationsTrackRegistry, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "run-ledger", run: runMissionRecommendationsRunLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "run-ledger-rollup", run: runMissionRecommendationsRunLedgerRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "run-ledger-coverage-check", run: runMissionRecommendationsRunLedgerCoverageCheck, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "final-response-gates", run: runMissionRecommendationsFinalResponseGates, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "schema-registry", run: runMissionRecommendationsSchemaRegistry, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "schema-registry-health", run: runMissionRecommendationsSchemaRegistryHealth, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "schema-registry-coverage", run: runMissionRecommendationsSchemaRegistryCoverage, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
		{name: "schema-health-repair-prompt", run: runMissionRecommendationsSchemaHealthRepairPrompt, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "readback", run: runMissionRecommendationsReadback, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "readback-delta", run: runMissionRecommendationsReadbackDelta, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "readback-diff-fixture", run: runMissionRecommendationsReadbackDiffFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "stale-checkpoint-rejection", run: runMissionRecommendationsStaleCheckpointRejection, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "operator-summary-check", run: runMissionRecommendationsOperatorSummaryCheck, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "run-link-schema-coverage", run: runMissionRecommendationsRunLinkSchemaCoverage, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "schema-validator-drift", run: runMissionRecommendationsSchemaValidatorDrift, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "pr-ci-timing-summary", run: runMissionRecommendationsPRCITimingSummary, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "pr-ci-windows-threshold", run: runMissionRecommendationsPRCIWindowsThreshold, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "failed-check-replay", run: runMissionRecommendationsFailedCheckReplay, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-covenant-rejected-ticket-fixture", run: runMissionRecommendationsCommandCovenantRejectedTicketFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-covenant-quarantine-fixture", run: runMissionRecommendationsCommandCovenantQuarantineFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-ticket-byte-preservation-fixture", run: runMissionRecommendationsCommandTicketBytePreservationFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "ticket-digest-readback-binding-fixture", run: runMissionRecommendationsTicketDigestReadbackBindingFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "policy-hash-mismatch-rejection-fixture", run: runMissionRecommendationsPolicyHashMismatchRejectionFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "policy-version-replay-rejection-fixture", run: runMissionRecommendationsPolicyVersionReplayRejectionFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "covenant-evidence-digest-readback-fixture", run: runMissionRecommendationsCovenantEvidenceDigestReadbackFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-compact-rejection-reason-fixture", run: runMissionRecommendationsCommandCompactRejectionReasonFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "blueprint-ticket-schema-compatibility-ledger", run: runMissionRecommendationsBlueprintTicketSchemaCompatibilityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "atlas-ticket-schema-compatibility-ledger", run: runMissionRecommendationsAtlasTicketSchemaCompatibilityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-ticket-schema-compatibility-ledger", run: runMissionRecommendationsFoundryTicketSchemaCompatibilityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-ticket-schema-compatibility-ledger", run: runMissionRecommendationsCommandTicketSchemaCompatibilityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "covenant-ticket-schema-authority-ledger", run: runMissionRecommendationsCovenantTicketSchemaAuthorityLedger, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "policy-ticket-public-safety-scan", run: runMissionRecommendationsPolicyTicketPublicSafetyScan, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "merge-check-binding", run: runMissionRecommendationsMergeCheckBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "post-merge-branch-deletion-readback", run: runMissionRecommendationsPostMergeBranchDeletionReadback, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "stale-remote-branch-repair", run: runMissionRecommendationsStaleRemoteBranchRepair, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "local-main-sync-readback", run: runMissionRecommendationsLocalMainSyncReadback, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "branch-cleanup-handoff-summary", run: runMissionRecommendationsBranchCleanupHandoffSummary, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "compaction-resume-prompt", run: runMissionRecommendationsCompactionResumePrompt, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "compaction-resume-regression", run: runMissionRecommendationsCompactionResumeRegression, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "resume-denial-evidence", run: runMissionRecommendationsResumeDenialEvidence, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "public-safety-readback-binding", run: runMissionRecommendationsPublicSafetyReadbackBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "scoped-public-safety-scan", run: runMissionRecommendationsScopedPublicSafetyScan, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "authority-promotion-negative-fixtures", run: runMissionRecommendationsAuthorityPromotionNegativeFixtures, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "public-safety-coverage-rollup", run: runMissionRecommendationsPublicSafetyCoverageRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "promoter-no-promotion-rollup", run: runMissionRecommendationsPromoterNoPromotionRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-promoter-agreement-rollup", run: runMissionRecommendationsCommandPromoterAgreementRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "promoter-rollup-count-mismatch-regression", run: runMissionRecommendationsPromoterRollupCountMismatchRegression, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-promoter-disagreement-denial", run: runMissionRecommendationsCommandPromoterDisagreementDenial, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-import-readiness-binding", run: runMissionRecommendationsFoundryImportReadinessBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "run-link-digest-check", run: runMissionRecommendationsRunLinkDigestCheck, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-handoff-replay-fixture", run: runMissionRecommendationsFoundryHandoffReplayFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-terminal-status-examples", run: runMissionRecommendationsFoundryTerminalStatusExamples, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "mission-dashboard-closure-binding", run: runMissionRecommendationsMissionDashboardClosureBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "mission-dashboard-provenance-links", run: runMissionRecommendationsMissionDashboardProvenanceLinks, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "mission-dashboard-freshness-checks", run: runMissionRecommendationsMissionDashboardFreshnessChecks, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "mission-dashboard-compact-filters", run: runMissionRecommendationsMissionDashboardCompactFilters, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "bounded-signer-contract-fixture", run: runMissionRecommendationsBoundedSignerContractFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "canonical-contract-registry-manifest", run: runMissionRecommendationsCanonicalContractRegistryManifest, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "contract-compatibility-inventory", run: runMissionRecommendationsContractCompatibilityInventory, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "canonical-json-vectors", run: runMissionRecommendationsCanonicalJSONVectors, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "canonical-json-vector-smoke-checks", run: runMissionRecommendationsCanonicalJSONVectorSmokeChecks, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "sentinel-hosted-ci-workflow-fixture", run: runMissionRecommendationsSentinelHostedCIWorkflowFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "sentinel-signal-state-fixture", run: runMissionRecommendationsSentinelSignalStateFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "signed-assurance-dry-run-fixture", run: runMissionRecommendationsSignedAssuranceDryRunFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "promoter-no-activation-boundary-fixture", run: runMissionRecommendationsPromoterNoActivationBoundaryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "workspace-root-preflight-fixture", run: runMissionRecommendationsWorkspaceRootPreflightFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "bounded-execution-packet-fixture", run: runMissionRecommendationsBoundedExecutionPacketFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "forge-goalrun-evidence-fixture", run: runMissionRecommendationsForgeGoalRunEvidenceFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "execution-packet-regression-matrix", run: runMissionRecommendationsExecutionPacketRegressionMatrix, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "durable-state-migration-metadata", run: runMissionRecommendationsDurableStateMigrationMetadata, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "exactly-once-resume-accounting-fixture", run: runMissionRecommendationsExactlyOnceResumeAccountingFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "replayable-state-packet-fixture", run: runMissionRecommendationsReplayableStatePacketFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "indexed-event-query-fixture", run: runMissionRecommendationsIndexedEventQueryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "atomic-evidence-transition-fixture", run: runMissionRecommendationsAtomicEvidenceTransitionFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "local-backup-restore-fixture", run: runMissionRecommendationsLocalBackupRestoreFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-readback-adapter-boundary-fixture", run: runMissionRecommendationsCommandReadbackAdapterBoundaryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "compact-timeline-filter-fixture", run: runMissionRecommendationsCompactTimelineFilterFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "authority-readiness-inventory-fixture", run: runMissionRecommendationsAuthorityReadinessInventoryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "content-addressed-evidence-manifest-fixture", run: runMissionRecommendationsContentAddressedEvidenceManifestFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-evidence-size-boundary-fixture", run: runMissionRecommendationsFoundryEvidenceSizeBoundaryFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "evidence-catalog-index-export", run: runMissionRecommendationsEvidenceCatalogIndexExport, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "stack-restart-resume-rehearsal", run: runMissionRecommendationsStackRestartResumeRehearsal, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "repeated-task-result-ledger-fixture", run: runMissionRecommendationsRepeatedTaskResultLedgerFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "failure-injection-fuzzing-fixture", run: runMissionRecommendationsFailureInjectionFuzzingFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "local-platform-fixture", run: runMissionRecommendationsLocalPlatformFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "non-ao-replay-binding-fixture", run: runMissionRecommendationsNonAOReplayBindingFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "kill-restart-replay-fixture", run: runMissionRecommendationsKillRestartReplayFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "rollback-terminal-readback-fixture", run: runMissionRecommendationsRollbackTerminalReadbackFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "golden-path-readiness-matrix", run: runMissionRecommendationsGoldenPathReadinessMatrix, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-final-closure-rollup", run: runMissionRecommendationsMonth3FinalClosureRollup, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-final-readiness-report", run: runMissionRecommendationsMonth3FinalReadinessReport, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-terminal-digest-binding", run: runMissionRecommendationsMonth3TerminalDigestBinding, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-non-ao-dry-run-replay", run: runMissionRecommendationsMonth3NonAODryRunReplay, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-real-run-acceptance", run: runMissionRecommendationsMonth3RealRunAcceptance, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-control-plane-observer", run: runMissionRecommendationsMonth3ControlPlaneObserver, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-schema-owner-registry", run: runMissionRecommendationsMonth3SchemaOwnerRegistry, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-evidence-externalization", run: runMissionRecommendationsMonth3EvidenceExternalization, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-cross-repo-ci-matrix", run: runMissionRecommendationsMonth3CrossRepoCIMatrix, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-operator-dashboard-readback", run: runMissionRecommendationsMonth3OperatorDashboardReadback, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-restart-resume-soak", run: runMissionRecommendationsMonth3RestartResumeSoak, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-provider-model-provenance", run: runMissionRecommendationsMonth3ProviderModelProvenance, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-rollback-replay-negative", run: runMissionRecommendationsMonth3RollbackReplayNegative, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-architecture-source-truth", run: runMissionRecommendationsMonth3ArchitectureSourceTruth, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-no-promotion-rsi-matrix", run: runMissionRecommendationsMonth3NoPromotionRSIMatrix, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "month3-foundry-safe-next-work", run: runMissionRecommendationsMonth3FoundrySafeNextWork, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "blueprint-canonical-preservation-fixture", run: runMissionRecommendationsBlueprintCanonicalPreservationFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "foundry-canonical-import-fixture", run: runMissionRecommendationsFoundryCanonicalImportFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "command-covenant-field-parity-fixture", run: runMissionRecommendationsCommandCovenantFieldParityFixture, commandClass: missionRecommendationCommandClassPlanningOnly},
		{name: "complete-node", run: runMissionRecommendationsCompleteNode, commandClass: missionRecommendationCommandClassMutationCapable},
		{name: "resume", run: runMissionRecommendationsResume, commandClass: missionRecommendationCommandClassMutationCapable},
		{name: "validate-evidence", run: runMissionRecommendationsValidateEvidence, commandClass: missionRecommendationCommandClassPlanningOnly, recordsRunLedger: true},
	}
}

func missionRecommendationCommandNames() []string {
	commands := missionRecommendationCommandRegistry()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.name)
	}
	return names
}

func missionRecommendationRunLedgerCommandNames() []string {
	commands := missionRecommendationCommandRegistry()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		if command.recordsRunLedger {
			names = append(names, command.name)
		}
	}
	return names
}

func missionRecommendationPlanningOnlyCommandNames() []string {
	return missionRecommendationCommandNamesByClass(missionRecommendationCommandClassPlanningOnly)
}

func missionRecommendationMutationCapableCommandNames() []string {
	return missionRecommendationCommandNamesByClass(missionRecommendationCommandClassMutationCapable)
}

func missionRecommendationCommandNamesByClass(commandClass string) []string {
	commands := missionRecommendationCommandRegistry()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		if command.commandClass == commandClass {
			names = append(names, command.name)
		}
	}
	return names
}

func missionRecommendationsUsageError() error {
	return fmt.Errorf("mission recommendations requires %s", formatCommandList(missionRecommendationCommandNames()))
}

func formatCommandList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
	}
}
