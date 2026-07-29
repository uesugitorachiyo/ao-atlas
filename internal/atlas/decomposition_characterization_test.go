package atlas

import (
	"bytes"
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestDecompositionCharacterizationPreservesWorkgraphJSONBytes(t *testing.T) {
	workgraph := Workgraph{
		ContractVersion: WorkgraphContract,
		ID:              "wg-characterization",
		TargetInstance:  "demo",
		Nodes: []WorkgraphNode{{
			ID:     "node-1",
			Status: "ready",
			FactoryTask: FactoryTask{
				ContractVersion: FactoryTaskContract,
				ID:              "task-1",
			},
			Dependencies: []string{},
			Blockers:     []string{},
			StitchTask:   true,
		}},
	}

	got, err := json.Marshal(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"contract_version":"ao.atlas.workgraph.v0.1","id":"wg-characterization","target_instance":"demo","nodes":[{"id":"node-1","status":"ready","factory_task":{"contract_version":"ao.atlas.factory-task.v0.1","id":"task-1","objective":"","target_factory_repo":"","factory_folder":"","acceptance_criteria":null,"non_goals":null,"write_scope":null,"verification_commands":null,"required_evidence":null,"safety_limits":null,"dependency_refs":null,"context_pack_refs":null},"dependencies":[],"blockers":[],"stitch_task":true}]}`
	if string(got) != want {
		t.Fatalf("workgraph JSON bytes changed\nwant: %s\n got: %s", want, got)
	}
}

func TestDecompositionCharacterizationPreservesRecommendationJSONBytes(t *testing.T) {
	wave := AtlasRecommendationWave{
		ContractVersion:  AtlasRecommendationWaveContract,
		MissionID:        "mission-characterization",
		TargetInstance:   "demo",
		Status:           "ready",
		SourceDigest:     "sha256:" + strings.Repeat("a", 64),
		MinimumTasks:     1,
		TotalTasks:       1,
		NodeBudget:       1,
		EstimatedMinutes: 15,
		Tasks: []AtlasRecommendationTask{{
			ID:     "recommendation-1",
			Owner:  "ao-atlas",
			Task:   "Preserve behavior.",
			NodeID: "node-1",
			TaskID: "task-1",
			RequiredGates: []string{
				"go test ./...",
			},
		}},
		NextRecommendedPrompt: "Continue.",
		FinalResponseReason:   "work remains",
	}

	got, err := json.Marshal(wave)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"contract_version":"ao.atlas.recommendation-wave.v0.1","mission_id":"mission-characterization","target_instance":"demo","status":"ready","source_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","minimum_tasks":1,"total_tasks":1,"node_budget":1,"estimated_minutes":15,"tasks":[{"id":"recommendation-1","owner":"ao-atlas","task":"Preserve behavior.","node_id":"node-1","task_id":"task-1","mutation_class":"","source_task_digest":"","target_factory_repo":"","factory_folder":"","required_gates":["go test ./..."],"verification_commands":null,"safety_limits":null}],"next_recommended_prompt":"Continue.","final_response_allowed":false,"final_response_reason":"work remains","promoter_readback_status":"","command_readback_status":"","public_safety_scan_status":"","safe_to_execute":false,"schedules_work":false,"executes_work":false,"approves_work":false}`
	if string(got) != want {
		t.Fatalf("recommendation JSON bytes changed\nwant: %s\n got: %s", want, got)
	}
}

func TestDecompositionCharacterizationPreservesWorkgraphReadinessAndTransitionOrder(t *testing.T) {
	workgraph := fixtureWorkgraph()
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(state.ReadyTaskIDs, ","), "factory-task,factory-task"; got != want {
		t.Fatalf("ready task ordering changed: got %q want %q", got, want)
	}
	if got, want := strings.Join(state.ExecutableReadyNodeIDs, ","), "task-ready,task-ready-2"; got != want {
		t.Fatalf("executable readiness ordering changed: got %q want %q", got, want)
	}
	next, ok := state.NextReadyNode()
	if !ok || next.ID != "task-ready" {
		t.Fatalf("next ready node changed: ok=%t node=%+v", ok, next)
	}

	link, err := BuildRunLink("factory-task", "completed", map[string]string{
		"verification": "evidence/verification.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, completedNodeID, err := state.CompleteWithRunLink(link)
	if err != nil {
		t.Fatal(err)
	}
	if completedNodeID != "task-ready" || updated.Nodes[1].Status != "completed" || updated.Nodes[3].Status != "ready" {
		t.Fatalf("completion transition or first-match ordering changed: completed=%q nodes=%+v", completedNodeID, updated.Nodes)
	}
}

func TestDecompositionCharacterizationPreservesOrderedCommandCatalogs(t *testing.T) {
	const wantRoot = "instance,intake,blueprint,mission,blueprint-request,workgraph,mutation-classes,factory-task,factory,context-pack,foundry,run-link,terminal-index"
	if got := strings.Join(rootCommandNames(), ","); got != wantRoot {
		t.Fatalf("root command registry ordering changed: got %q want %q", got, wantRoot)
	}
	var stdout, stderr bytes.Buffer
	if code := Run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("root usage exit code changed: %d", code)
	}
	if stdout.String() != "" {
		t.Fatalf("root usage unexpectedly wrote stdout: %q", stdout.String())
	}
	if got, want := stderr.String(), "atlas <"+strings.ReplaceAll(wantRoot, ",", "|")+"> ...\n"; got != want {
		t.Fatalf("root command ordering changed\nwant: %q\n got: %q", want, got)
	}

	const wantRecommendations = "import,export-next-wave,export-refactoring-wave,next-track,consumed-ledger,track-registry,run-ledger,run-ledger-rollup,run-ledger-coverage-check,final-response-gates,schema-registry,schema-registry-health,schema-registry-coverage,schema-health-repair-prompt,readback,readback-delta,readback-diff-fixture,stale-checkpoint-rejection,operator-summary-check,run-link-schema-coverage,schema-validator-drift,pr-ci-timing-summary,pr-ci-windows-threshold,failed-check-replay,command-covenant-rejected-ticket-fixture,command-covenant-quarantine-fixture,command-ticket-byte-preservation-fixture,ticket-digest-readback-binding-fixture,policy-hash-mismatch-rejection-fixture,policy-version-replay-rejection-fixture,covenant-evidence-digest-readback-fixture,command-compact-rejection-reason-fixture,blueprint-ticket-schema-compatibility-ledger,atlas-ticket-schema-compatibility-ledger,foundry-ticket-schema-compatibility-ledger,command-ticket-schema-compatibility-ledger,covenant-ticket-schema-authority-ledger,policy-ticket-public-safety-scan,merge-check-binding,post-merge-branch-deletion-readback,stale-remote-branch-repair,local-main-sync-readback,branch-cleanup-handoff-summary,compaction-resume-prompt,compaction-resume-regression,resume-denial-evidence,public-safety-readback-binding,scoped-public-safety-scan,authority-promotion-negative-fixtures,public-safety-coverage-rollup,promoter-no-promotion-rollup,command-promoter-agreement-rollup,promoter-rollup-count-mismatch-regression,command-promoter-disagreement-denial,foundry-import-readiness-binding,run-link-digest-check,foundry-handoff-replay-fixture,foundry-terminal-status-examples,mission-dashboard-closure-binding,mission-dashboard-provenance-links,mission-dashboard-freshness-checks,mission-dashboard-compact-filters,bounded-signer-contract-fixture,canonical-contract-registry-manifest,contract-compatibility-inventory,canonical-json-vectors,canonical-json-vector-smoke-checks,sentinel-hosted-ci-workflow-fixture,sentinel-signal-state-fixture,signed-assurance-dry-run-fixture,promoter-no-activation-boundary-fixture,workspace-root-preflight-fixture,bounded-execution-packet-fixture,forge-goalrun-evidence-fixture,execution-packet-regression-matrix,durable-state-migration-metadata,exactly-once-resume-accounting-fixture,replayable-state-packet-fixture,indexed-event-query-fixture,atomic-evidence-transition-fixture,local-backup-restore-fixture,command-readback-adapter-boundary-fixture,compact-timeline-filter-fixture,authority-readiness-inventory-fixture,content-addressed-evidence-manifest-fixture,foundry-evidence-size-boundary-fixture,evidence-catalog-index-export,stack-restart-resume-rehearsal,repeated-task-result-ledger-fixture,failure-injection-fuzzing-fixture,local-platform-fixture,non-ao-replay-binding-fixture,kill-restart-replay-fixture,rollback-terminal-readback-fixture,golden-path-readiness-matrix,month3-final-closure-rollup,month3-final-readiness-report,month3-terminal-digest-binding,month3-non-ao-dry-run-replay,month3-real-run-acceptance,month3-control-plane-observer,month3-schema-owner-registry,month3-evidence-externalization,month3-cross-repo-ci-matrix,month3-operator-dashboard-readback,month3-restart-resume-soak,month3-provider-model-provenance,month3-rollback-replay-negative,month3-architecture-source-truth,month3-no-promotion-rsi-matrix,month3-foundry-safe-next-work,blueprint-canonical-preservation-fixture,foundry-canonical-import-fixture,command-covenant-field-parity-fixture,complete-node,resume,validate-evidence"
	if got := strings.Join(missionRecommendationCommandNames(), ","); got != wantRecommendations {
		t.Fatalf("recommendation command ordering changed\nwant: %s\n got: %s", wantRecommendations, got)
	}
}

func TestDecompositionCharacterizationBindsCommandsToHandlers(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		const want = `instance=runInstance
intake=runIntake
blueprint=runBlueprint
mission=runMission
blueprint-request=runBlueprintRequest
workgraph=runWorkgraph
mutation-classes=runMutationClasses
factory-task=runFactoryTask
factory=runFactory
context-pack=runContextPack
foundry=runFoundry
run-link=runRunLink
terminal-index=runTerminalIndex`
		got := make([]string, 0, len(rootCommandRegistry()))
		for _, command := range rootCommandRegistry() {
			got = append(got, command.name+"="+decompositionHandlerIdentity(t, command.run))
		}
		if joined := strings.Join(got, "\n"); joined != want {
			t.Fatalf("root command handler bindings changed\nwant:\n%s\n\ngot:\n%s", want, joined)
		}
	})

	t.Run("recommendations", func(t *testing.T) {
		const want = `import=runMissionRecommendationsImport
export-next-wave=runMissionRecommendationsExportNextWave
export-refactoring-wave=runMissionRecommendationsExportRefactoringWave
next-track=runMissionRecommendationsNextTrack
consumed-ledger=runMissionRecommendationsConsumedLedger
track-registry=runMissionRecommendationsTrackRegistry
run-ledger=runMissionRecommendationsRunLedger
run-ledger-rollup=runMissionRecommendationsRunLedgerRollup
run-ledger-coverage-check=runMissionRecommendationsRunLedgerCoverageCheck
final-response-gates=runMissionRecommendationsFinalResponseGates
schema-registry=runMissionRecommendationsSchemaRegistry
schema-registry-health=runMissionRecommendationsSchemaRegistryHealth
schema-registry-coverage=runMissionRecommendationsSchemaRegistryCoverage
schema-health-repair-prompt=runMissionRecommendationsSchemaHealthRepairPrompt
readback=runMissionRecommendationsReadback
readback-delta=runMissionRecommendationsReadbackDelta
readback-diff-fixture=runMissionRecommendationsReadbackDiffFixture
stale-checkpoint-rejection=runMissionRecommendationsStaleCheckpointRejection
operator-summary-check=runMissionRecommendationsOperatorSummaryCheck
run-link-schema-coverage=runMissionRecommendationsRunLinkSchemaCoverage
schema-validator-drift=runMissionRecommendationsSchemaValidatorDrift
pr-ci-timing-summary=runMissionRecommendationsPRCITimingSummary
pr-ci-windows-threshold=runMissionRecommendationsPRCIWindowsThreshold
failed-check-replay=runMissionRecommendationsFailedCheckReplay
command-covenant-rejected-ticket-fixture=runMissionRecommendationsCommandCovenantRejectedTicketFixture
command-covenant-quarantine-fixture=runMissionRecommendationsCommandCovenantQuarantineFixture
command-ticket-byte-preservation-fixture=runMissionRecommendationsCommandTicketBytePreservationFixture
ticket-digest-readback-binding-fixture=runMissionRecommendationsTicketDigestReadbackBindingFixture
policy-hash-mismatch-rejection-fixture=runMissionRecommendationsPolicyHashMismatchRejectionFixture
policy-version-replay-rejection-fixture=runMissionRecommendationsPolicyVersionReplayRejectionFixture
covenant-evidence-digest-readback-fixture=runMissionRecommendationsCovenantEvidenceDigestReadbackFixture
command-compact-rejection-reason-fixture=runMissionRecommendationsCommandCompactRejectionReasonFixture
blueprint-ticket-schema-compatibility-ledger=runMissionRecommendationsBlueprintTicketSchemaCompatibilityLedger
atlas-ticket-schema-compatibility-ledger=runMissionRecommendationsAtlasTicketSchemaCompatibilityLedger
foundry-ticket-schema-compatibility-ledger=runMissionRecommendationsFoundryTicketSchemaCompatibilityLedger
command-ticket-schema-compatibility-ledger=runMissionRecommendationsCommandTicketSchemaCompatibilityLedger
covenant-ticket-schema-authority-ledger=runMissionRecommendationsCovenantTicketSchemaAuthorityLedger
policy-ticket-public-safety-scan=runMissionRecommendationsPolicyTicketPublicSafetyScan
merge-check-binding=runMissionRecommendationsMergeCheckBinding
post-merge-branch-deletion-readback=runMissionRecommendationsPostMergeBranchDeletionReadback
stale-remote-branch-repair=runMissionRecommendationsStaleRemoteBranchRepair
local-main-sync-readback=runMissionRecommendationsLocalMainSyncReadback
branch-cleanup-handoff-summary=runMissionRecommendationsBranchCleanupHandoffSummary
compaction-resume-prompt=runMissionRecommendationsCompactionResumePrompt
compaction-resume-regression=runMissionRecommendationsCompactionResumeRegression
resume-denial-evidence=runMissionRecommendationsResumeDenialEvidence
public-safety-readback-binding=runMissionRecommendationsPublicSafetyReadbackBinding
scoped-public-safety-scan=runMissionRecommendationsScopedPublicSafetyScan
authority-promotion-negative-fixtures=runMissionRecommendationsAuthorityPromotionNegativeFixtures
public-safety-coverage-rollup=runMissionRecommendationsPublicSafetyCoverageRollup
promoter-no-promotion-rollup=runMissionRecommendationsPromoterNoPromotionRollup
command-promoter-agreement-rollup=runMissionRecommendationsCommandPromoterAgreementRollup
promoter-rollup-count-mismatch-regression=runMissionRecommendationsPromoterRollupCountMismatchRegression
command-promoter-disagreement-denial=runMissionRecommendationsCommandPromoterDisagreementDenial
foundry-import-readiness-binding=runMissionRecommendationsFoundryImportReadinessBinding
run-link-digest-check=runMissionRecommendationsRunLinkDigestCheck
foundry-handoff-replay-fixture=runMissionRecommendationsFoundryHandoffReplayFixture
foundry-terminal-status-examples=runMissionRecommendationsFoundryTerminalStatusExamples
mission-dashboard-closure-binding=runMissionRecommendationsMissionDashboardClosureBinding
mission-dashboard-provenance-links=runMissionRecommendationsMissionDashboardProvenanceLinks
mission-dashboard-freshness-checks=runMissionRecommendationsMissionDashboardFreshnessChecks
mission-dashboard-compact-filters=runMissionRecommendationsMissionDashboardCompactFilters
bounded-signer-contract-fixture=runMissionRecommendationsBoundedSignerContractFixture
canonical-contract-registry-manifest=runMissionRecommendationsCanonicalContractRegistryManifest
contract-compatibility-inventory=runMissionRecommendationsContractCompatibilityInventory
canonical-json-vectors=runMissionRecommendationsCanonicalJSONVectors
canonical-json-vector-smoke-checks=runMissionRecommendationsCanonicalJSONVectorSmokeChecks
sentinel-hosted-ci-workflow-fixture=runMissionRecommendationsSentinelHostedCIWorkflowFixture
sentinel-signal-state-fixture=runMissionRecommendationsSentinelSignalStateFixture
signed-assurance-dry-run-fixture=runMissionRecommendationsSignedAssuranceDryRunFixture
promoter-no-activation-boundary-fixture=runMissionRecommendationsPromoterNoActivationBoundaryFixture
workspace-root-preflight-fixture=runMissionRecommendationsWorkspaceRootPreflightFixture
bounded-execution-packet-fixture=runMissionRecommendationsBoundedExecutionPacketFixture
forge-goalrun-evidence-fixture=runMissionRecommendationsForgeGoalRunEvidenceFixture
execution-packet-regression-matrix=runMissionRecommendationsExecutionPacketRegressionMatrix
durable-state-migration-metadata=runMissionRecommendationsDurableStateMigrationMetadata
exactly-once-resume-accounting-fixture=runMissionRecommendationsExactlyOnceResumeAccountingFixture
replayable-state-packet-fixture=runMissionRecommendationsReplayableStatePacketFixture
indexed-event-query-fixture=runMissionRecommendationsIndexedEventQueryFixture
atomic-evidence-transition-fixture=runMissionRecommendationsAtomicEvidenceTransitionFixture
local-backup-restore-fixture=runMissionRecommendationsLocalBackupRestoreFixture
command-readback-adapter-boundary-fixture=runMissionRecommendationsCommandReadbackAdapterBoundaryFixture
compact-timeline-filter-fixture=runMissionRecommendationsCompactTimelineFilterFixture
authority-readiness-inventory-fixture=runMissionRecommendationsAuthorityReadinessInventoryFixture
content-addressed-evidence-manifest-fixture=runMissionRecommendationsContentAddressedEvidenceManifestFixture
foundry-evidence-size-boundary-fixture=runMissionRecommendationsFoundryEvidenceSizeBoundaryFixture
evidence-catalog-index-export=runMissionRecommendationsEvidenceCatalogIndexExport
stack-restart-resume-rehearsal=runMissionRecommendationsStackRestartResumeRehearsal
repeated-task-result-ledger-fixture=runMissionRecommendationsRepeatedTaskResultLedgerFixture
failure-injection-fuzzing-fixture=runMissionRecommendationsFailureInjectionFuzzingFixture
local-platform-fixture=runMissionRecommendationsLocalPlatformFixture
non-ao-replay-binding-fixture=runMissionRecommendationsNonAOReplayBindingFixture
kill-restart-replay-fixture=runMissionRecommendationsKillRestartReplayFixture
rollback-terminal-readback-fixture=runMissionRecommendationsRollbackTerminalReadbackFixture
golden-path-readiness-matrix=runMissionRecommendationsGoldenPathReadinessMatrix
month3-final-closure-rollup=runMissionRecommendationsMonth3FinalClosureRollup
month3-final-readiness-report=runMissionRecommendationsMonth3FinalReadinessReport
month3-terminal-digest-binding=runMissionRecommendationsMonth3TerminalDigestBinding
month3-non-ao-dry-run-replay=runMissionRecommendationsMonth3NonAODryRunReplay
month3-real-run-acceptance=runMissionRecommendationsMonth3RealRunAcceptance
month3-control-plane-observer=runMissionRecommendationsMonth3ControlPlaneObserver
month3-schema-owner-registry=runMissionRecommendationsMonth3SchemaOwnerRegistry
month3-evidence-externalization=runMissionRecommendationsMonth3EvidenceExternalization
month3-cross-repo-ci-matrix=runMissionRecommendationsMonth3CrossRepoCIMatrix
month3-operator-dashboard-readback=runMissionRecommendationsMonth3OperatorDashboardReadback
month3-restart-resume-soak=runMissionRecommendationsMonth3RestartResumeSoak
month3-provider-model-provenance=runMissionRecommendationsMonth3ProviderModelProvenance
month3-rollback-replay-negative=runMissionRecommendationsMonth3RollbackReplayNegative
month3-architecture-source-truth=runMissionRecommendationsMonth3ArchitectureSourceTruth
month3-no-promotion-rsi-matrix=runMissionRecommendationsMonth3NoPromotionRSIMatrix
month3-foundry-safe-next-work=runMissionRecommendationsMonth3FoundrySafeNextWork
blueprint-canonical-preservation-fixture=runMissionRecommendationsBlueprintCanonicalPreservationFixture
foundry-canonical-import-fixture=runMissionRecommendationsFoundryCanonicalImportFixture
command-covenant-field-parity-fixture=runMissionRecommendationsCommandCovenantFieldParityFixture
complete-node=runMissionRecommendationsCompleteNode
resume=runMissionRecommendationsResume
validate-evidence=runMissionRecommendationsValidateEvidence`
		got := make([]string, 0, len(missionRecommendationCommandRegistry()))
		for _, command := range missionRecommendationCommandRegistry() {
			got = append(got, command.name+"="+decompositionHandlerIdentity(t, command.run))
		}
		if joined := strings.Join(got, "\n"); joined != want {
			t.Fatalf("recommendation command handler bindings changed\nwant:\n%s\n\ngot:\n%s", want, joined)
		}
	})
}

func decompositionHandlerIdentity(t *testing.T, handler any) string {
	t.Helper()
	value := reflect.ValueOf(handler)
	if value.Kind() != reflect.Func || value.IsNil() {
		t.Fatalf("registry handler must be a non-nil function, got %T", handler)
	}
	function := runtime.FuncForPC(value.Pointer())
	if function == nil {
		t.Fatalf("runtime function identity unavailable for %T", handler)
	}
	name := function.Name()
	if separator := strings.LastIndex(name, "."); separator >= 0 {
		return name[separator+1:]
	}
	return name
}

func TestDecompositionCharacterizationPreservesRepresentativeCLIStreams(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "version",
			args:       []string{"--version"},
			wantCode:   0,
			wantStdout: "ao-atlas version=dev source_sha=unknown\n",
		},
		{
			name:       "unknown root",
			args:       []string{"unknown"},
			wantCode:   1,
			wantStderr: "error: unknown command \"unknown\"\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(test.args, &stdout, &stderr)
			if code != test.wantCode || stdout.String() != test.wantStdout || stderr.String() != test.wantStderr {
				t.Fatalf("CLI contract changed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}
