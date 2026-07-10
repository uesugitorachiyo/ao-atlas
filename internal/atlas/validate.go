package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	InstanceContract                                           = "ao.atlas.stack-instance.v0.1"
	AtlasRegistryContract                                      = "ao.atlas.foundry-registry.v0.1"
	InstanceDoctorContract                                     = "ao.atlas.instance-doctor.v0.1"
	IntakeContract                                             = "ao.atlas.intake.v0.1"
	MissionStatusContract                                      = "ao.atlas.mission-status.v0.1"
	AOMissionImportContract                                    = "ao.atlas.ao-mission-import.v0.1"
	AOMissionWorkgraphMetadataContract                         = "ao.atlas.ao-mission-workgraph-metadata.v0.1"
	AOMissionFinalSynthesisReadbackContract                    = "ao.atlas.ao-mission-final-synthesis-readback.v0.1"
	AtlasRecommendationWaveContract                            = "ao.atlas.recommendation-wave.v0.1"
	AtlasRecommendationReadbackContract                        = "ao.atlas.recommendation-readback.v0.1"
	AtlasMissionReadbackDeltaContract                          = "ao.atlas.mission-readback-delta.v0.1"
	AtlasMissionReadbackDiffFixtureContract                    = "ao.atlas.mission-readback-diff-fixture.v0.1"
	AtlasMissionStaleCheckpointRejectionContract               = "ao.atlas.mission-stale-checkpoint-rejection.v0.1"
	AtlasMissionOperatorSummaryCheckContract                   = "ao.atlas.mission-operator-summary-check.v0.1"
	AtlasNodeCommandReadbackContract                           = "ao.atlas.command-readback.v0.1"
	AtlasNodePromoterNoPromotionContract                       = "ao.atlas.promoter-no-promotion.v0.1"
	AtlasNodeSentinelPublicSafetyContract                      = "ao.atlas.sentinel-public-safety.v0.1"
	AtlasRunLinkSchemaCoverageContract                         = "ao.atlas.run-link-schema-coverage.v0.1"
	AtlasSchemaValidatorDriftContract                          = "ao.atlas.schema-validator-drift.v0.1"
	AtlasPRCITimingLedgerContract                              = "ao.atlas.pr-ci-timing-ledger.v0.1"
	AtlasPRCITimingSummaryContract                             = "ao.atlas.pr-ci-timing-summary.v0.1"
	AtlasPRCINormalizedRowContract                             = "ao.atlas.pr-ci-normalized-row.v0.1"
	AtlasPRCIWindowsThresholdEvidenceContract                  = "ao.atlas.pr-ci-windows-threshold-evidence.v0.1"
	AtlasP0BWindowsCIWaitTelemetryContract                     = "ao.atlas.p0b-windows-ci-wait-telemetry.v0.1"
	AtlasP0BRollbackAuditContract                              = "ao.atlas.p0b-rollback-audit.v0.1"
	AtlasP0BCommandPromoterAgreementContract                   = "ao.atlas.p0b-command-promoter-agreement.v0.1"
	AtlasFailedCheckReplayInputContract                        = "ao.atlas.failed-check-replay-input.v0.1"
	AtlasFailedCheckReplayFixtureContract                      = "ao.atlas.failed-check-replay-fixture.v0.1"
	AtlasCommandCovenantRejectedTicketInputContract            = "ao.atlas.command-covenant-rejected-ticket-input.v0.1"
	AtlasCommandCovenantRejectedTicketFixtureContract          = "ao.atlas.command-covenant-rejected-ticket-fixture.v0.1"
	AtlasCommandCovenantQuarantineInputContract                = "ao.atlas.command-covenant-quarantine-input.v0.1"
	AtlasCommandCovenantQuarantineFixtureContract              = "ao.atlas.command-covenant-quarantine-fixture.v0.1"
	AtlasCommandTicketBytePreservationInputContract            = "ao.atlas.command-ticket-byte-preservation-input.v0.1"
	AtlasCommandTicketBytePreservationFixtureContract          = "ao.atlas.command-ticket-byte-preservation-fixture.v0.1"
	AtlasTicketDigestReadbackBindingInputContract              = "ao.atlas.ticket-digest-readback-binding-input.v0.1"
	AtlasTicketDigestReadbackBindingFixtureContract            = "ao.atlas.ticket-digest-readback-binding-fixture.v0.1"
	AtlasPolicyHashMismatchRejectionInputContract              = "ao.atlas.policy-hash-mismatch-rejection-input.v0.1"
	AtlasPolicyHashMismatchRejectionFixtureContract            = "ao.atlas.policy-hash-mismatch-rejection-fixture.v0.1"
	AtlasPolicyVersionReplayRejectionInputContract             = "ao.atlas.policy-version-replay-rejection-input.v0.1"
	AtlasPolicyVersionReplayRejectionFixtureContract           = "ao.atlas.policy-version-replay-rejection-fixture.v0.1"
	AtlasCovenantEvidenceDigestReadbackInputContract           = "ao.atlas.covenant-evidence-digest-readback-input.v0.1"
	AtlasCovenantEvidenceDigestReadbackFixtureContract         = "ao.atlas.covenant-evidence-digest-readback-fixture.v0.1"
	AtlasCommandCompactRejectionReasonInputContract            = "ao.atlas.command-compact-rejection-reason-input.v0.1"
	AtlasCommandCompactRejectionReasonFixtureContract          = "ao.atlas.command-compact-rejection-reason-fixture.v0.1"
	AtlasBlueprintTicketSchemaCompatibilityLedgerInputContract = "ao.atlas.blueprint-ticket-schema-compatibility-ledger-input.v0.1"
	AtlasBlueprintTicketSchemaCompatibilityLedgerContract      = "ao.atlas.blueprint-ticket-schema-compatibility-ledger.v0.1"
	AtlasTicketSchemaCompatibilityLedgerInputContract          = "ao.atlas.atlas-ticket-schema-compatibility-ledger-input.v0.1"
	AtlasTicketSchemaCompatibilityLedgerContract               = "ao.atlas.atlas-ticket-schema-compatibility-ledger.v0.1"
	AtlasFoundryTicketSchemaCompatibilityLedgerInputContract   = "ao.atlas.foundry-ticket-schema-compatibility-ledger-input.v0.1"
	AtlasFoundryTicketSchemaCompatibilityLedgerContract        = "ao.atlas.foundry-ticket-schema-compatibility-ledger.v0.1"
	AtlasCommandTicketSchemaCompatibilityLedgerInputContract   = "ao.atlas.command-ticket-schema-compatibility-ledger-input.v0.1"
	AtlasCommandTicketSchemaCompatibilityLedgerContract        = "ao.atlas.command-ticket-schema-compatibility-ledger.v0.1"
	AtlasCovenantTicketSchemaAuthorityLedgerInputContract      = "ao.atlas.covenant-ticket-schema-authority-ledger-input.v0.1"
	AtlasCovenantTicketSchemaAuthorityLedgerContract           = "ao.atlas.covenant-ticket-schema-authority-ledger.v0.1"
	AtlasPolicyTicketPublicSafetyScanInputContract             = "ao.atlas.policy-ticket-public-safety-scan-input.v0.1"
	AtlasPolicyTicketPublicSafetyScanContract                  = "ao.atlas.policy-ticket-public-safety-scan.v0.1"
	AtlasP0BPRCILedgerContract                                 = "ao.atlas.p0b-pr-ci-ledger.v0.1"
	AtlasP0CReadinessCriteriaContract                          = "ao.atlas.p0c-readiness-criteria.v0.1"
	AtlasP0CMissionFoundryHandoffCheckContract                 = "ao.atlas.p0c-mission-foundry-handoff-check.v0.1"
	AtlasMergeCheckBindingInputContract                        = "ao.atlas.merge-check-binding-input.v0.1"
	AtlasMergeCheckBindingContract                             = "ao.atlas.merge-check-binding.v0.1"
	AtlasPostMergeBranchDeletionReadbackContract               = "ao.atlas.post-merge-branch-deletion-readback.v0.1"
	AtlasStaleRemoteBranchRepairInputContract                  = "ao.atlas.stale-remote-branch-repair-input.v0.1"
	AtlasStaleRemoteBranchRepairContract                       = "ao.atlas.stale-remote-branch-repair.v0.1"
	AtlasLocalMainSyncReadbackInputContract                    = "ao.atlas.local-main-sync-readback-input.v0.1"
	AtlasLocalMainSyncReadbackContract                         = "ao.atlas.local-main-sync-readback.v0.1"
	AtlasBranchCleanupHandoffSummaryContract                   = "ao.atlas.branch-cleanup-handoff-summary.v0.1"
	AtlasCompactionResumePromptContract                        = "ao.atlas.compaction-resume-prompt.v0.1"
	AtlasCompactionResumeRegressionContract                    = "ao.atlas.compaction-resume-regression.v0.1"
	AtlasResumeDenialEvidenceContract                          = "ao.atlas.resume-denial-evidence.v0.1"
	AtlasPublicSafetyReadbackBindingContract                   = "ao.atlas.public-safety-readback-binding.v0.1"
	AtlasScopedPublicSafetyScanContract                        = "ao.atlas.scoped-public-safety-scan.v0.1"
	WorkgraphContract                                          = "ao.atlas.workgraph.v0.1"
	WorkgraphRepairPlanContract                                = "ao.atlas.workgraph-repair-plan.v0.1"
	FactoryTaskContract                                        = "ao.atlas.factory-task.v0.1"
	FactoryMaterializationContract                             = "ao.atlas.factory-materialization.v0.1"
	ContextPackContract                                        = "ao.atlas.context-pack.v0.1"
	FoundryHandoffContract                                     = "ao.atlas.foundry-handoff.v0.1"
	FoundryImportContract                                      = "ao.atlas.foundry-import.v0.1"
	FoundryContinuationHandoffContract                         = "ao.atlas.foundry-continuation-handoff.v0.1"
	RunLinkContract                                            = "ao.atlas.run-link.v0.1"
	BlueprintRequestContract                                   = "ao.atlas.blueprint-request.v0.1"
	BlueprintImportContract                                    = "ao.atlas.blueprint-import.v0.1"
	BlueprintCandidateRulesContract                            = "ao.atlas.blueprint-candidate-rules.v0.1"
	BlueprintCandidateSelectionContract                        = "ao.atlas.blueprint-candidate-selection.v0.1"
	MutationClassModelContract                                 = "ao.atlas.mutation-classes.v0.1"
	LowRiskCodeDenialAuditContract                             = "ao.atlas.low-risk-code-denial-audit.v0.1"
)

const AtlasAuthorityPromotionNegativeFixturesContract = "ao.atlas.authority-promotion-negative-fixtures.v0.1"
const AtlasPublicSafetyCoverageRollupContract = "ao.atlas.public-safety-coverage-rollup.v0.1"
const AtlasPromoterNoPromotionRollupContract = "ao.atlas.promoter-no-promotion-rollup.v0.1"
const AtlasCommandPromoterAgreementRollupContract = "ao.atlas.command-promoter-agreement-rollup.v0.1"
const AtlasPromoterRollupCountMismatchRegressionContract = "ao.atlas.promoter-rollup-count-mismatch-regression.v0.1"
const AtlasCommandPromoterDisagreementDenialContract = "ao.atlas.command-promoter-disagreement-denial.v0.1"
const AtlasFoundryImportReadinessBindingContract = "ao.atlas.foundry-import-readiness-binding.v0.1"
const AtlasRunLinkDigestCheckContract = "ao.atlas.run-link-digest-check.v0.1"
const AtlasFoundryHandoffReplayFixtureContract = "ao.atlas.foundry-handoff-replay-fixture.v0.1"
const AtlasFoundryTerminalStatusExamplesContract = "ao.atlas.foundry-terminal-status-examples.v0.1"
const AtlasMissionDashboardClosureBindingContract = "ao.atlas.mission-dashboard-closure-binding.v0.1"
const AtlasMissionDashboardProvenanceLinksContract = "ao.atlas.mission-dashboard-provenance-links.v0.1"
const AtlasMissionDashboardFreshnessChecksContract = "ao.atlas.mission-dashboard-freshness-checks.v0.1"
const AtlasMissionDashboardCompactFiltersContract = "ao.atlas.mission-dashboard-compact-filters.v0.1"
const AOMissionRefactoringRecommendationsContract = "ao.mission.refactoring-recommendations.v0.1"
const AtlasRecommendationNextTrackDecisionContract = "ao.atlas.recommendation-next-track-decision.v0.1"
const AtlasConsumedRecommendationLedgerContract = "ao.atlas.consumed-recommendation-ledger.v0.1"
const AtlasRecommendationTrackRegistryContract = "ao.atlas.recommendation-track-registry.v0.1"
const AtlasRecommendationCommandRunLedgerContract = "ao.atlas.recommendation-command-run-ledger.v0.1"
const AtlasRecommendationCommandRunLedgerRollupContract = "ao.atlas.recommendation-command-run-ledger-rollup.v0.1"
const AtlasRecommendationRunLedgerCoverageCheckContract = "ao.atlas.recommendation-run-ledger-coverage-check.v0.1"
const AtlasRecommendationRunLedgerOperatorSummaryBindingContract = "ao.atlas.recommendation-run-ledger-operator-summary-binding.v0.1"
const AtlasRecommendationRunLedgerRetryFixturePackContract = "ao.atlas.recommendation-run-ledger-retry-fixture-pack.v0.1"
const AtlasRecommendationFinalResponseGatesContract = "ao.atlas.recommendation-final-response-gates.v0.1"
const AtlasRecommendationStaleReadbackTrackFixtureContract = "ao.atlas.recommendation-stale-readback-track-fixture.v0.1"
const AtlasRecommendationEvidenceSchemaRegistryContract = "ao.atlas.recommendation-evidence-schema-registry.v0.1"
const AtlasRecommendationEvidenceSchemaRegistryCoverageContract = "ao.atlas.recommendation-evidence-schema-registry-coverage.v0.1"
const AtlasSchemaHealthRepairPromptContract = "ao.atlas.schema-health-repair-prompt.v0.1"

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var driveAbsPattern = regexp.MustCompile(`^[A-Za-z]:[\\/]`)

func LoadJSON[T any](path string) (T, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	return value, nil
}

func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ValidateInstance(instance Instance) error {
	var errs []string
	requireContract(&errs, "instance", instance.ContractVersion, InstanceContract)
	requireField(&errs, "id", instance.ID)
	requireField(&errs, "state_root", instance.StateRoot)
	requireField(&errs, "toolchain_root", instance.ToolchainRoot)
	if len(instance.Roots) == 0 {
		errs = append(errs, "roots must not be empty")
	}
	requiredRoots := []string{"mission", "workgraph", "context", "evidence", "worktree"}
	for _, key := range requiredRoots {
		requireField(&errs, "roots."+key, instance.Roots[key])
	}
	checkPublicPathMap(&errs, instance.Roots)
	checkPublicPath(&errs, "state_root", instance.StateRoot, false)
	checkPublicPath(&errs, "toolchain_root", instance.ToolchainRoot, false)
	return joinErrors(errs)
}

func ValidateAtlasRegistry(registry AtlasRegistry) error {
	var errs []string
	requireContract(&errs, "atlas_registry", registry.ContractVersion, AtlasRegistryContract)
	requireField(&errs, "instance_id", registry.InstanceID)
	requireField(&errs, "toolchain_root", registry.ToolchainRoot)
	if len(registry.Roots) == 0 {
		errs = append(errs, "roots must not be empty")
	}
	checkPublicPathMap(&errs, registry.Roots)
	checkPublicPath(&errs, "toolchain_root", registry.ToolchainRoot, false)
	if registry.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if registry.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if registry.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	return joinErrors(errs)
}

func ValidateInstanceDoctorReport(report InstanceDoctorReport) error {
	var errs []string
	requireContract(&errs, "instance_doctor", report.ContractVersion, InstanceDoctorContract)
	requireField(&errs, "instance_id", report.InstanceID)
	if !oneOf(report.Status, "ready", "blocked", "failed") {
		errs = append(errs, "status must be ready, blocked, or failed")
	}
	required := []string{"instance", "registry_parity", "ignored_local_state", "worktree_hygiene", "shared_toolchain", "authority_boundary"}
	for _, key := range required {
		if strings.TrimSpace(report.Checks[key]) == "" {
			errs = append(errs, "checks."+key+" must be present")
		}
	}
	if report.Status == "ready" {
		for _, key := range required {
			if report.Checks[key] != "passed" {
				errs = append(errs, "checks."+key+" must be passed when status is ready")
			}
		}
		if strings.TrimSpace(report.FirstFailingCheck) != "" {
			errs = append(errs, "first_failing_check must be empty when status is ready")
		}
		if len(report.BlockingNextActions) != 0 {
			errs = append(errs, "blocking_next_actions must be empty when status is ready")
		}
	} else {
		requireField(&errs, "first_failing_check", report.FirstFailingCheck)
		requireList(&errs, "blocking_next_actions", report.BlockingNextActions)
	}
	checkPublicStrings(&errs, "blocking_next_actions", report.BlockingNextActions, true)
	checkPublicStrings(&errs, "maintenance_suggestions", report.MaintenanceSuggestions, true)
	if report.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if report.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if report.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	return joinErrors(errs)
}

func BuildInstanceDoctorReport(instance Instance, registry AtlasRegistry) (InstanceDoctorReport, error) {
	report := diagnoseInstanceDoctor(instance, registry)
	if err := ValidateInstanceDoctorReport(report); err != nil {
		return report, err
	}
	if report.Status != "ready" {
		return report, errors.New(report.BlockingNextActions[0])
	}
	return report, nil
}

func diagnoseInstanceDoctor(instance Instance, registry AtlasRegistry) InstanceDoctorReport {
	report := InstanceDoctorReport{
		ContractVersion:        InstanceDoctorContract,
		InstanceID:             firstNonEmpty(instance.ID, registry.InstanceID, "unknown-instance"),
		Status:                 "ready",
		Checks:                 map[string]string{},
		BlockingNextActions:    []string{},
		MaintenanceSuggestions: []string{},
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
	}
	fail := func(check, message string) {
		report.Checks[check] = "failed"
		if report.FirstFailingCheck == "" {
			report.FirstFailingCheck = check
			report.BlockingNextActions = append(report.BlockingNextActions, message)
		}
		report.Status = "failed"
	}
	if err := ValidateInstance(instance); err != nil {
		fail("instance", err.Error())
	} else {
		report.Checks["instance"] = "passed"
	}
	if err := ValidateAtlasRegistry(registry); err != nil {
		if registry.SchedulesWork || registry.ExecutesWork || registry.ApprovesWork {
			fail("authority_boundary", err.Error())
		} else {
			fail("registry_parity", err.Error())
		}
	}
	if report.Checks["authority_boundary"] == "" {
		report.Checks["authority_boundary"] = "passed"
	}
	if report.Checks["registry_parity"] == "" {
		if registry.InstanceID != instance.ID {
			fail("registry_parity", "registry instance_id must match instance id")
		}
		if registry.ToolchainRoot != instance.ToolchainRoot {
			fail("registry_parity", "registry toolchain_root must match instance toolchain_root")
		}
		for key, value := range instance.Roots {
			if registry.Roots[key] != value {
				fail("registry_parity", fmt.Sprintf("registry roots.%s must match instance roots.%s", key, key))
			}
		}
		if report.Checks["registry_parity"] == "" {
			report.Checks["registry_parity"] = "passed"
		}
	}
	if !strings.HasPrefix(instance.StateRoot, ".atlas-local/") && !strings.HasPrefix(instance.StateRoot, ".atlas-state/") {
		fail("ignored_local_state", "state_root must be under ignored Atlas local state")
	} else if !rootsUnderStateRoot(instance) {
		fail("ignored_local_state", "mission/workgraph/context/evidence/worktree roots must remain under state_root")
	} else {
		report.Checks["ignored_local_state"] = "passed"
	}
	worktree := strings.TrimSpace(instance.Roots["worktree"])
	if worktree == "" || worktree == "." || worktree == ".." {
		fail("worktree_hygiene", "worktree root must be a bounded instance path")
	} else {
		report.Checks["worktree_hygiene"] = "passed"
	}
	if strings.HasPrefix(instance.ToolchainRoot, ".atlas-local/") ||
		strings.HasPrefix(instance.ToolchainRoot, ".atlas-state/") ||
		instance.ToolchainRoot == instance.StateRoot ||
		instance.ToolchainRoot == worktree {
		fail("shared_toolchain", "toolchain_root must point to a shared AO toolchain, not copied instance state")
	} else {
		report.Checks["shared_toolchain"] = "passed"
	}
	if report.Status == "ready" {
		report.MaintenanceSuggestions = append(report.MaintenanceSuggestions, "Keep generated instance state under ignored Atlas local roots.")
	}
	return report
}

func ValidateIntake(intake Intake) (BlueprintRequest, error) {
	var errs []string
	requireContract(&errs, "intake", intake.ContractVersion, IntakeContract)
	requireField(&errs, "id", intake.ID)
	checkPublicStrings(&errs, "instruction_refs", intake.InstructionRefs, false)
	checkPublicStrings(&errs, "folder_refs", intake.FolderRefs, false)
	missing := []string{}
	if strings.TrimSpace(intake.TargetInstance) == "" {
		missing = append(missing, "target_instance")
	}
	if len(strings.Fields(intake.BroadPrompt)) < 8 {
		missing = append(missing, "broad_prompt_detail")
	}
	if len(intake.Constraints) == 0 {
		missing = append(missing, "constraints")
	}
	if len(missing) > 0 {
		return BlueprintRequest{
			ContractVersion: BlueprintRequestContract,
			IntakeID:        intake.ID,
			Status:          "blueprint_required",
			Missing:         missing,
			Reason:          "AO Atlas cannot compile underspecified intake into a ready workgraph without AO Blueprint clarification.",
		}, joinErrors(errs)
	}
	return BlueprintRequest{}, joinErrors(errs)
}

func ValidateMissionStatus(status MissionStatus) error {
	var errs []string
	requireContract(&errs, "mission_status", status.ContractVersion, MissionStatusContract)
	requireField(&errs, "intake_id", status.IntakeID)
	requireField(&errs, "workgraph_id", status.WorkgraphID)
	requireField(&errs, "target_instance", status.TargetInstance)
	if !oneOf(status.CompletionStatus, "completed", "blocked", "in_progress") {
		errs = append(errs, "completion_status must be completed, blocked, or in_progress")
	}
	for _, key := range []string{"ready", "blocked", "completed", "failed"} {
		if _, ok := status.NodeCounts[key]; !ok {
			errs = append(errs, "node_counts."+key+" must be present")
		}
	}
	checkPublicStrings(&errs, "missing_context_packs", status.MissingContextPacks, true)
	checkPublicStrings(&errs, "missing_handoffs", status.MissingHandoffs, true)
	requireField(&errs, "next_recommended_action", status.NextRecommendedAction)
	checkPublicPath(&errs, "next_recommended_action", status.NextRecommendedAction, true)
	if len(status.NextActions) == 0 {
		errs = append(errs, "next_actions must not be empty")
	}
	checkPublicStrings(&errs, "next_actions", status.NextActions, true)
	if status.AuthorityLadder != nil {
		validateAuthorityLadderStatus(&errs, *status.AuthorityLadder)
	}
	requireField(&errs, "final_response_reason", status.FinalResponseReason)
	if status.FinalStateReconciliation == nil {
		errs = append(errs, "final_state_reconciliation must be present")
	} else {
		validateAtlasFinalStateReconciliation(&errs, *status.FinalStateReconciliation)
	}
	if status.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if status.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	return joinErrors(errs)
}

func validateAtlasFinalStateReconciliation(errs *[]string, packet AtlasFinalStateReconciliation) {
	requireContract(errs, "final_state_reconciliation", packet.ContractVersion, "ao.atlas.final-state-reconciliation.v0.1")
	requireField(errs, "final_state_reconciliation.status", packet.Status)
	requireField(errs, "final_state_reconciliation.workgraph_status", packet.WorkgraphStatus)
	requireField(errs, "final_state_reconciliation.foundry_rollup_status", packet.FoundryRollupStatus)
	requireField(errs, "final_state_reconciliation.promoter_verdict_status", packet.PromoterVerdictStatus)
	requireField(errs, "final_state_reconciliation.command_readback_status", packet.CommandReadbackStatus)
	requireField(errs, "final_state_reconciliation.exact_next_action", packet.ExactNextAction)
	checkPublicPath(errs, "final_state_reconciliation.exact_next_action", packet.ExactNextAction, true)
	if packet.SchedulesWork {
		*errs = append(*errs, "final_state_reconciliation.schedules_work must be false")
	}
	if packet.ExecutesWork {
		*errs = append(*errs, "final_state_reconciliation.executes_work must be false")
	}
	if packet.ApprovesWork {
		*errs = append(*errs, "final_state_reconciliation.approves_work must be false")
	}
}

func validateAuthorityLadderStatus(errs *[]string, ladder AuthorityLadderStatus) {
	requireField(errs, "authority_ladder.current_class", ladder.CurrentClass)
	requireField(errs, "authority_ladder.next_class", ladder.NextClass)
	requireList(errs, "authority_ladder.proven_live_classes", ladder.ProvenLiveClasses)
	requireList(errs, "authority_ladder.blockers", ladder.Blockers)
	requireList(errs, "authority_ladder.required_evidence", ladder.RequiredEvidence)
	if len(ladder.DeniedHigherClasses) == 0 {
		*errs = append(*errs, "authority_ladder.denied_higher_classes must not be empty")
	}
	if len(ladder.DoNotAdvanceGates) == 0 {
		*errs = append(*errs, "authority_ladder.do_not_advance_gates must not be empty")
	}
	checkPublicStrings(errs, "authority_ladder.proven_live_classes", ladder.ProvenLiveClasses, true)
	checkPublicStrings(errs, "authority_ladder.dry_run_ready_classes", ladder.DryRunReadyClasses, true)
	checkPublicStrings(errs, "authority_ladder.blockers", ladder.Blockers, true)
	checkPublicStrings(errs, "authority_ladder.required_evidence", ladder.RequiredEvidence, true)
	checkPublicStrings(errs, "authority_ladder.do_not_advance_gates", ladder.DoNotAdvanceGates, true)
	checkPublicPathMapStrict(errs, ladder.DeniedHigherClasses)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func ValidateBlueprintRequest(request BlueprintRequest) error {
	var errs []string
	requireContract(&errs, "blueprint_request", request.ContractVersion, BlueprintRequestContract)
	requireField(&errs, "intake_id", request.IntakeID)
	if request.Status != "blueprint_required" {
		errs = append(errs, "status must be blueprint_required")
	}
	requireList(&errs, "missing", request.Missing)
	requireField(&errs, "reason", request.Reason)
	checkPublicStrings(&errs, "missing", request.Missing, true)
	checkPublicPath(&errs, "reason", request.Reason, true)
	return joinErrors(errs)
}

func ValidateWorkgraph(workgraph Workgraph) error {
	var errs []string
	requireContract(&errs, "workgraph", workgraph.ContractVersion, WorkgraphContract)
	requireField(&errs, "id", workgraph.ID)
	requireField(&errs, "target_instance", workgraph.TargetInstance)
	if len(workgraph.Nodes) == 0 {
		errs = append(errs, "nodes must not be empty")
	}
	seen := map[string]WorkgraphNode{}
	for i, node := range workgraph.Nodes {
		field := fmt.Sprintf("nodes[%d]", i)
		requireField(&errs, field+".id", node.ID)
		if _, ok := seen[node.ID]; ok {
			errs = append(errs, field+".id must be unique")
		}
		seen[node.ID] = node
		if !oneOf(node.Status, "ready", "blocked", "completed") {
			errs = append(errs, field+".status must be ready, blocked, or completed")
		}
		if node.Status == "blocked" && len(node.Blockers) == 0 {
			errs = append(errs, field+".blockers must explain blocked state")
		}
		if err := ValidateFactoryTask(node.FactoryTask); err != nil {
			errs = append(errs, field+".factory_task: "+err.Error())
		}
	}
	for _, node := range workgraph.Nodes {
		for _, dep := range node.Dependencies {
			if _, ok := seen[dep]; !ok {
				errs = append(errs, "dependency "+dep+" does not reference an existing node")
			}
		}
	}
	return joinErrors(errs)
}

func ValidateWorkgraphRepairPlan(plan WorkgraphRepairPlan) error {
	var errs []string
	requireContract(&errs, "workgraph_repair_plan", plan.ContractVersion, WorkgraphRepairPlanContract)
	requireField(&errs, "id", plan.ID)
	requireField(&errs, "task_id", plan.TaskID)
	if plan.Status != "repair_required" {
		errs = append(errs, "status must be repair_required")
	}
	if !oneOf(plan.SourceRunLinkStatus, "blocked", "failed") {
		errs = append(errs, "source_run_link_status must be blocked or failed")
	}
	requireField(&errs, "reason", plan.Reason)
	if len(plan.RepairTasks) == 0 {
		errs = append(errs, "repair_tasks must not be empty")
	}
	for i, task := range plan.RepairTasks {
		if err := ValidateFactoryTask(task); err != nil {
			errs = append(errs, fmt.Sprintf("repair_tasks[%d]: %s", i, err.Error()))
		}
	}
	if plan.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if plan.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if plan.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	return joinErrors(errs)
}

func ValidateMutationClassModel(model MutationClassModel) error {
	var errs []string
	requireContract(&errs, "mutation_class_model", model.ContractVersion, MutationClassModelContract)
	requireField(&errs, "id", model.ID)
	if len(model.Classes) == 0 {
		errs = append(errs, "classes must not be empty")
	}
	if model.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if model.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if model.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	required := requiredMutationClassNames()
	seen := map[string]bool{}
	for i, class := range model.Classes {
		prefix := fmt.Sprintf("classes[%d]", i)
		requireField(&errs, prefix+".name", class.Name)
		if !required[class.Name] {
			errs = append(errs, prefix+".name must be one of the required mutation classes")
		}
		if seen[class.Name] {
			errs = append(errs, prefix+".name must be unique")
		}
		seen[class.Name] = true
		requireList(&errs, prefix+".allowed_paths", class.AllowedPaths)
		requireList(&errs, prefix+".forbidden_paths", class.ForbiddenPaths)
		if class.MaxFiles <= 0 {
			errs = append(errs, prefix+".max_files must be positive")
		}
		requireList(&errs, prefix+".required_gates", class.RequiredGates)
		requireList(&errs, prefix+".rollback_requirements", class.RollbackRequirements)
		requireList(&errs, prefix+".ci_requirements", class.CIRequirements)
		requireList(&errs, prefix+".promotion_requirements", class.PromotionRequirements)
		checkPublicStrings(&errs, prefix+".allowed_paths", class.AllowedPaths, true)
		checkPublicStrings(&errs, prefix+".forbidden_paths", class.ForbiddenPaths, true)
		checkPublicStrings(&errs, prefix+".required_gates", class.RequiredGates, true)
		checkPublicStrings(&errs, prefix+".rollback_requirements", class.RollbackRequirements, true)
		checkPublicStrings(&errs, prefix+".ci_requirements", class.CIRequirements, true)
		checkPublicStrings(&errs, prefix+".promotion_requirements", class.PromotionRequirements, true)
	}
	for name := range required {
		if !seen[name] {
			errs = append(errs, "classes must include "+name)
		}
	}
	return joinErrors(errs)
}

func requiredMutationClassNames() map[string]bool {
	return map[string]bool{
		"docs_only_single_file": true,
		"docs_only_multi_file":  true,
		"docs_config_only":      true,
		"test_only":             true,
		"low_risk_code":         true,
		"multi_repo_low_risk":   true,
		"complex_repo_mutation": true,
	}
}

func ValidateLowRiskCodeDenialAudit(audit LowRiskCodeDenialAudit) error {
	var errs []string
	requireContract(&errs, "low_risk_code_denial_audit", audit.SchemaVersion, LowRiskCodeDenialAuditContract)
	if audit.Status != "blocked" {
		errs = append(errs, "status must be blocked")
	}
	if audit.MutationClass != "low_risk_code" {
		errs = append(errs, "mutation_class must be low_risk_code")
	}
	if audit.CurrentProvenLiveClass != "test_only" {
		errs = append(errs, "current_proven_live_class must be test_only")
	}
	if audit.NextDeniedClass != "low_risk_code" {
		errs = append(errs, "next_denied_class must be low_risk_code")
	}
	if !audit.SafeToRequest {
		errs = append(errs, "safe_to_request must be true for dry-run continuation")
	}
	if audit.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	requireList(&errs, "missing_policy_evidence", audit.MissingPolicyEvidence)
	requireList(&errs, "missing_rollback_evidence", audit.MissingRollbackEvidence)
	requireList(&errs, "missing_sentinel_promoter_evidence", audit.MissingSentinelPromoterEvidence)
	requireList(&errs, "ci_requirements", audit.CIRequirements)
	requireField(&errs, "sentinel_state", audit.SentinelState)
	requireField(&errs, "promoter_state", audit.PromoterState)
	requireField(&errs, "exact_next_action", audit.ExactNextAction)
	requireField(&errs, "denial_reason", audit.DenialReason)
	for _, want := range []struct {
		field  string
		values []string
		item   string
	}{
		{"missing_policy_evidence", audit.MissingPolicyEvidence, "policy:low_risk_code_live_promotion"},
		{"missing_policy_evidence", audit.MissingPolicyEvidence, "command_readback:low_risk_code_live"},
		{"missing_rollback_evidence", audit.MissingRollbackEvidence, "rollback_proof:low_risk_code_live"},
		{"missing_sentinel_promoter_evidence", audit.MissingSentinelPromoterEvidence, "sentinel_clear:low_risk_code_live"},
		{"missing_sentinel_promoter_evidence", audit.MissingSentinelPromoterEvidence, "promoter_promotion:low_risk_code_live"},
		{"ci_requirements", audit.CIRequirements, "ci_passed:low_risk_code_live"},
	} {
		if !containsValue(want.values, want.item) {
			errs = append(errs, want.field+" must include "+want.item)
		}
	}
	if audit.ExactNextAction != "build_low_risk_code_promotion_prerequisites" {
		errs = append(errs, "exact_next_action must be build_low_risk_code_promotion_prerequisites")
	}
	if audit.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if audit.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if audit.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if audit.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	if audit.CallsProviders {
		errs = append(errs, "calls_providers must be false")
	}
	if audit.ReleaseOrPublishAllowed {
		errs = append(errs, "release_or_publish_allowed must be false")
	}
	if audit.FullyUnsupervisedMutationClaimed {
		errs = append(errs, "fully_unsupervised_mutation_claimed must be false")
	}
	checkPublicStrings(&errs, "missing_policy_evidence", audit.MissingPolicyEvidence, true)
	checkPublicStrings(&errs, "missing_rollback_evidence", audit.MissingRollbackEvidence, true)
	checkPublicStrings(&errs, "missing_sentinel_promoter_evidence", audit.MissingSentinelPromoterEvidence, true)
	checkPublicStrings(&errs, "ci_requirements", audit.CIRequirements, true)
	return joinErrors(errs)
}

func ValidateFactoryTask(task FactoryTask) error {
	var errs []string
	requireContract(&errs, "factory_task", task.ContractVersion, FactoryTaskContract)
	requireField(&errs, "id", task.ID)
	requireField(&errs, "objective", task.Objective)
	requireField(&errs, "target_factory_repo", task.TargetFactoryRepo)
	requireField(&errs, "factory_folder", task.FactoryFolder)
	requireList(&errs, "acceptance_criteria", task.Acceptance)
	requireList(&errs, "non_goals", task.NonGoals)
	requireList(&errs, "write_scope", task.WriteScope)
	requireList(&errs, "verification_commands", task.Verification)
	requireList(&errs, "required_evidence", task.RequiredEvidence)
	requireList(&errs, "safety_limits", task.SafetyLimits)
	checkPublicPath(&errs, "target_factory_repo", task.TargetFactoryRepo, false)
	checkPublicPath(&errs, "factory_folder", task.FactoryFolder, false)
	if strings.TrimSpace(task.MutationClass) != "" && !requiredMutationClassNames()[task.MutationClass] {
		errs = append(errs, "mutation_class must be one of the required mutation classes")
	}
	checkPublicStrings(&errs, "write_scope", task.WriteScope, false)
	checkPublicStrings(&errs, "required_gates", task.RequiredGates, true)
	checkPublicStrings(&errs, "rollback_scope", task.RollbackScope, true)
	checkPublicStrings(&errs, "context_pack_refs", task.ContextPackRefs, true)
	checkPublicPath(&errs, "authority_boundary", task.AuthorityBoundary, true)
	return joinErrors(errs)
}

func ValidateFoundryReadyTaskAuthorityMetadata(task FactoryTask) error {
	var errs []string
	if strings.TrimSpace(task.MutationClass) == "" {
		errs = append(errs, "mutation_class must not be empty")
	} else if !requiredMutationClassNames()[task.MutationClass] {
		errs = append(errs, "mutation_class must be one of the required mutation classes")
	}
	requireList(&errs, "write_scope", task.WriteScope)
	requireList(&errs, "rollback_scope", task.RollbackScope)
	requireList(&errs, "required_gates", task.RequiredGates)
	requireList(&errs, "required_evidence", task.RequiredEvidence)
	requireField(&errs, "authority_boundary", task.AuthorityBoundary)
	checkPublicStrings(&errs, "write_scope", task.WriteScope, true)
	checkPublicStrings(&errs, "rollback_scope", task.RollbackScope, true)
	checkPublicStrings(&errs, "required_gates", task.RequiredGates, true)
	checkPublicStrings(&errs, "required_evidence", task.RequiredEvidence, true)
	checkPublicPath(&errs, "authority_boundary", task.AuthorityBoundary, true)
	return joinErrors(errs)
}

func ValidateFactoryMaterialization(materialization FactoryMaterialization) error {
	var errs []string
	requireContract(&errs, "factory_materialization", materialization.ContractVersion, FactoryMaterializationContract)
	requireField(&errs, "task_id", materialization.TaskID)
	if materialization.Mode != "dry_run" {
		errs = append(errs, "mode must be dry_run")
	}
	requireField(&errs, "output_root", materialization.OutputRoot)
	if strings.ContainsAny(materialization.OutputRoot, `/\`) {
		errs = append(errs, "output_root must not record a local path")
	}
	requireList(&errs, "files", materialization.Files)
	checkPublicStrings(&errs, "files", materialization.Files, true)
	if materialization.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if materialization.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if !digestPattern.MatchString(materialization.TaskDigest) {
		errs = append(errs, "task_digest must be sha256:<64 hex>")
	}
	return joinErrors(errs)
}

func ValidateContextPack(pack ContextPack, budgetOverride int) error {
	var errs []string
	requireContract(&errs, "context_pack", pack.ContractVersion, ContextPackContract)
	requireField(&errs, "id", pack.ID)
	requireField(&errs, "task_id", pack.TaskID)
	if pack.BudgetBytes <= 0 {
		errs = append(errs, "budget_bytes must be positive")
	}
	budget := pack.BudgetBytes
	if budgetOverride > 0 {
		budget = budgetOverride
	}
	data, _ := json.Marshal(pack)
	if len(data) > budget {
		errs = append(errs, fmt.Sprintf("context pack exceeds budget: %d > %d bytes", len(data), budget))
	}
	if len(pack.SourceRefs) == 0 {
		errs = append(errs, "source_refs must not be empty")
	}
	for i, ref := range pack.SourceRefs {
		checkPublicPath(&errs, fmt.Sprintf("source_refs[%d].ref", i), ref.Ref, true)
		if !digestPattern.MatchString(ref.Digest) {
			errs = append(errs, fmt.Sprintf("source_refs[%d].digest must be sha256:<64 hex>", i))
		}
	}
	requireList(&errs, "summaries", pack.Summaries)
	requireField(&errs, "missing_context_protocol", pack.MissingProtocol)
	checkPublicStrings(&errs, "summaries", pack.Summaries, true)
	checkPublicStrings(&errs, "assumptions", pack.Assumptions, true)
	checkPublicStrings(&errs, "exclusions", pack.Exclusions, true)
	if strings.TrimSpace(pack.MissingContextReason) != "" {
		checkPublicPath(&errs, "missing_context_reason", pack.MissingContextReason, false)
	}
	return joinErrors(errs)
}

func BuildContextRepack(task FactoryTask, link RunLink, sourceRef, sourceDigest string, budget int) (ContextPack, error) {
	if err := ValidateFactoryTask(task); err != nil {
		return ContextPack{}, err
	}
	if err := ValidateRunLink(link); err != nil {
		return ContextPack{}, err
	}
	if task.ID != link.TaskID {
		return ContextPack{}, fmt.Errorf("run-link task_id must match factory task id")
	}
	if !oneOf(link.Status, "blocked", "failed") {
		return ContextPack{}, fmt.Errorf("run-link status must be blocked or failed")
	}
	if strings.TrimSpace(link.Evidence["needs_context"]) == "" {
		return ContextPack{}, fmt.Errorf("run-link evidence must include needs_context")
	}
	pack := ContextPack{
		ContractVersion:      ContextPackContract,
		ID:                   task.ID + "-context-repack",
		TaskID:               task.ID,
		BudgetBytes:          budget,
		SourceRefs:           []SourceRef{{Ref: sourceRef, Digest: sourceDigest}},
		Summaries:            []string{"Repacked bounded context requested by a needs_context run-link."},
		Assumptions:          []string{"Only referenced public-safe sources are included."},
		Exclusions:           []string{"whole source repositories", "private local state", "credentials", "provider transcripts"},
		MissingContextReason: "run-link evidence needs_context=" + link.Evidence["needs_context"],
		MissingProtocol:      "Ask AO Blueprint or the operator for missing requirements before widening scope.",
	}
	if err := ValidateContextPack(pack, 0); err != nil {
		return ContextPack{}, err
	}
	return pack, nil
}

func ValidateFoundryHandoff(handoff FoundryHandoff) error {
	var errs []string
	requireContract(&errs, "foundry_handoff", handoff.ContractVersion, FoundryHandoffContract)
	requireField(&errs, "id", handoff.ID)
	requireField(&errs, "target_instance", handoff.TargetInstance)
	if handoff.Status != "ready_for_foundry" {
		errs = append(errs, "status must be ready_for_foundry")
	}
	if len(handoff.Tasks) == 0 {
		errs = append(errs, "tasks must not be empty")
	}
	for i, task := range handoff.Tasks {
		requireField(&errs, fmt.Sprintf("tasks[%d].id", i), task.ID)
		requireField(&errs, fmt.Sprintf("tasks[%d].objective", i), task.Objective)
		checkPublicPath(&errs, fmt.Sprintf("tasks[%d].target_factory_repo", i), task.TargetFactoryRepo, false)
		checkPublicPath(&errs, fmt.Sprintf("tasks[%d].factory_folder", i), task.FactoryFolder, false)
	}
	return joinErrors(errs)
}

func ValidateFoundryImport(foundryImport FoundryImport) error {
	var errs []string
	requireContract(&errs, "foundry_import", foundryImport.ContractVersion, FoundryImportContract)
	requireField(&errs, "id", foundryImport.ID)
	requireField(&errs, "workgraph_id", foundryImport.WorkgraphID)
	requireField(&errs, "target_instance", foundryImport.TargetInstance)
	if foundryImport.Status != "ready_for_foundry_fixture_import" {
		errs = append(errs, "status must be ready_for_foundry_fixture_import")
	}
	if len(foundryImport.SourceArtifacts) == 0 {
		errs = append(errs, "source_artifacts must not be empty")
	}
	for i, source := range foundryImport.SourceArtifacts {
		checkPublicPath(&errs, fmt.Sprintf("source_artifacts[%d].ref", i), source.Ref, true)
		if !digestPattern.MatchString(source.Digest) {
			errs = append(errs, fmt.Sprintf("source_artifacts[%d].digest must be sha256:<64 hex>", i))
		}
	}
	if len(foundryImport.Tasks) == 0 {
		errs = append(errs, "tasks must not be empty")
	}
	if foundryImport.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if foundryImport.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if foundryImport.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	seenPaths := map[string]bool{}
	for i, fixture := range foundryImport.Tasks {
		prefix := fmt.Sprintf("tasks[%d]", i)
		requireField(&errs, prefix+".node_id", fixture.NodeID)
		requireField(&errs, prefix+".task_id", fixture.TaskID)
		requireField(&errs, prefix+".path", fixture.Path)
		requireField(&errs, prefix+".mutation_class", fixture.MutationClass)
		requireList(&errs, prefix+".write_scope", fixture.WriteScope)
		requireList(&errs, prefix+".rollback_scope", fixture.RollbackScope)
		requireList(&errs, prefix+".required_gates", fixture.RequiredGates)
		requireList(&errs, prefix+".required_evidence", fixture.RequiredEvidence)
		requireField(&errs, prefix+".authority_boundary", fixture.AuthorityBoundary)
		checkPublicPath(&errs, prefix+".path", fixture.Path, true)
		checkPublicPath(&errs, prefix+".mutation_class", fixture.MutationClass, true)
		checkPublicStrings(&errs, prefix+".write_scope", fixture.WriteScope, true)
		checkPublicStrings(&errs, prefix+".rollback_scope", fixture.RollbackScope, true)
		checkPublicStrings(&errs, prefix+".required_gates", fixture.RequiredGates, true)
		checkPublicStrings(&errs, prefix+".required_evidence", fixture.RequiredEvidence, true)
		checkPublicPath(&errs, prefix+".authority_boundary", fixture.AuthorityBoundary, true)
		if strings.Contains(filepath.Clean(fixture.Path), "..") {
			errs = append(errs, prefix+".path must stay inside the import output")
		}
		if seenPaths[fixture.Path] {
			errs = append(errs, prefix+".path must be unique")
		}
		seenPaths[fixture.Path] = true
		if !digestPattern.MatchString(fixture.TaskHash) {
			errs = append(errs, prefix+".task_digest must be sha256:<64 hex>")
		}
		if err := ValidateFactoryTask(fixture.Task); err != nil {
			errs = append(errs, prefix+".task: "+err.Error())
		}
		if err := ValidateFoundryReadyTaskAuthorityMetadata(fixture.Task); err != nil {
			errs = append(errs, prefix+".task authority metadata: "+err.Error())
		}
		if fixture.TaskID != fixture.Task.ID {
			errs = append(errs, prefix+".task_id must match task.id")
		}
		if fixture.MutationClass != fixture.Task.MutationClass {
			errs = append(errs, prefix+".mutation_class must match task.mutation_class")
		}
		if !equalStringSlices(fixture.WriteScope, fixture.Task.WriteScope) {
			errs = append(errs, prefix+".write_scope must match task.write_scope")
		}
		if !equalStringSlices(fixture.RollbackScope, fixture.Task.RollbackScope) {
			errs = append(errs, prefix+".rollback_scope must match task.rollback_scope")
		}
		if !equalStringSlices(fixture.RequiredGates, fixture.Task.RequiredGates) {
			errs = append(errs, prefix+".required_gates must match task.required_gates")
		}
		if !equalStringSlices(fixture.RequiredEvidence, fixture.Task.RequiredEvidence) {
			errs = append(errs, prefix+".required_evidence must match task.required_evidence")
		}
		if fixture.AuthorityBoundary != fixture.Task.AuthorityBoundary {
			errs = append(errs, prefix+".authority_boundary must match task.authority_boundary")
		}
	}
	return joinErrors(errs)
}

func ValidateFoundryContinuationHandoff(handoff FoundryContinuationHandoff) error {
	var errs []string
	requireContract(&errs, "foundry_continuation_handoff", handoff.ContractVersion, FoundryContinuationHandoffContract)
	requireField(&errs, "id", handoff.ID)
	requireField(&errs, "target_folder", handoff.TargetFolder)
	requireField(&errs, "command", handoff.Command)
	requireField(&errs, "next_recommended_action", handoff.NextRecommendedAction)
	requireField(&errs, "prompt", handoff.Prompt)
	requireField(&errs, "blueprint_pack_path", handoff.BlueprintPackPath)
	requireField(&errs, "atlas_import_path", handoff.AtlasImportPath)
	requireField(&errs, "workgraph_path", handoff.WorkgraphPath)
	requireField(&errs, "foundry_import_path", handoff.FoundryImportPath)
	requireField(&errs, "first_safe_node", handoff.FirstSafeNode)
	requireField(&errs, "class_boundary", handoff.ClassBoundary)
	requireList(&errs, "stop_conditions", handoff.StopConditions)
	requireList(&errs, "safety_prohibitions", handoff.SafetyProhibitions)
	if handoff.Command != "codex --yolo" {
		errs = append(errs, "command must be codex --yolo")
	}
	checkPublicPath(&errs, "target_folder", handoff.TargetFolder, true)
	checkPublicPath(&errs, "next_recommended_action", handoff.NextRecommendedAction, true)
	for _, required := range []string{
		"Move to AO Foundry",
		"Run codex --yolo",
		"Paste this prompt",
		"do not stop after import validation",
		"do not stop after one gate artifact",
		"do not stop after one node",
		"Continue until all generated slices/tasks/nodes are consumed or a true hard blocker remains",
		"Atlas must not execute live mutation",
	} {
		if !strings.Contains(handoff.Prompt, required) {
			errs = append(errs, "prompt must include "+required)
		}
	}
	if strings.Contains(handoff.NextRecommendedAction, "cat ") {
		errs = append(errs, "next_recommended_action must not use cat as the primary action")
	}
	if handoff.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if handoff.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if handoff.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if handoff.TotalNodeCount < 1 {
		errs = append(errs, "total_node_count must be positive")
	}
	if handoff.ReadyNodeCount < 1 {
		errs = append(errs, "ready_node_count must be positive")
	}
	return joinErrors(errs)
}

func ValidateRunLink(link RunLink) error {
	var errs []string
	requireContract(&errs, "run_link", link.ContractVersion, RunLinkContract)
	requireField(&errs, "task_id", link.TaskID)
	if !oneOf(link.Status, "planned", "running", "completed", "blocked", "failed") {
		errs = append(errs, "status must be planned, running, completed, blocked, or failed")
	}
	if len(link.Evidence) == 0 {
		errs = append(errs, "evidence must not be empty")
	}
	checkPublicPathMapStrict(&errs, link.Evidence)
	if !digestPattern.MatchString(link.Digest) {
		errs = append(errs, "digest must be sha256:<64 hex>")
	}
	return joinErrors(errs)
}

func BuildRunLink(taskID, status string, evidence map[string]string) (RunLink, error) {
	link := RunLink{
		ContractVersion: RunLinkContract,
		TaskID:          taskID,
		Status:          status,
		Evidence:        evidence,
	}
	link.Digest = digestRunLink(link)
	if err := ValidateRunLink(link); err != nil {
		return RunLink{}, err
	}
	return link, nil
}

func NextReadyNode(workgraph Workgraph) (WorkgraphNode, bool) {
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return WorkgraphNode{}, false
	}
	return state.NextReadyNode()
}

func CompleteWorkgraph(workgraph Workgraph, link RunLink) (Workgraph, string, error) {
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return Workgraph{}, "", err
	}
	return state.CompleteWithRunLink(link)
}

func BuildWorkgraphRepairPlan(workgraph Workgraph, link RunLink) (WorkgraphRepairPlan, error) {
	if err := ValidateWorkgraph(workgraph); err != nil {
		return WorkgraphRepairPlan{}, err
	}
	if err := ValidateRunLink(link); err != nil {
		return WorkgraphRepairPlan{}, err
	}
	if !oneOf(link.Status, "blocked", "failed") {
		return WorkgraphRepairPlan{}, fmt.Errorf("run-link status must be blocked or failed")
	}
	for _, node := range workgraph.Nodes {
		if node.FactoryTask.ID != link.TaskID {
			continue
		}
		source := node.FactoryTask
		repair := FactoryTask{
			ContractVersion:   FactoryTaskContract,
			ID:                "repair-" + source.ID,
			Objective:         "Repair blocked Atlas factory task: " + source.Objective,
			TargetFactoryRepo: source.TargetFactoryRepo,
			FactoryFolder:     source.FactoryFolder + "-repair",
			Acceptance:        []string{"a follow-up run-link for " + source.ID + " validates with status completed"},
			NonGoals:          []string{"do not schedule work from Atlas", "do not execute work from Atlas", "do not approve work from Atlas"},
			WriteScope:        append([]string(nil), source.WriteScope...),
			Verification:      append([]string(nil), source.Verification...),
			RequiredEvidence:  append([]string(nil), source.RequiredEvidence...),
			SafetyLimits:      append(append([]string(nil), source.SafetyLimits...), "repair plan is readback only"),
			DependencyRefs:    []string{source.ID},
			ContextPackRefs:   append([]string(nil), source.ContextPackRefs...),
		}
		plan := WorkgraphRepairPlan{
			ContractVersion:     WorkgraphRepairPlanContract,
			ID:                  workgraph.ID + "-" + source.ID + "-repair-plan",
			TaskID:              source.ID,
			Status:              "repair_required",
			SourceRunLinkStatus: link.Status,
			Reason:              "run-link status " + link.Status + " did not complete the task; emit a bounded repair task for Foundry scheduling",
			RepairTasks:         []FactoryTask{repair},
			SchedulesWork:       false,
			ExecutesWork:        false,
			ApprovesWork:        false,
		}
		if err := ValidateWorkgraphRepairPlan(plan); err != nil {
			return WorkgraphRepairPlan{}, err
		}
		return plan, nil
	}
	return WorkgraphRepairPlan{}, fmt.Errorf("no matching workgraph node for run-link task_id %q", link.TaskID)
}

func DigestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func requireContract(errs *[]string, name, got, want string) {
	if got != want {
		*errs = append(*errs, name+" contract_version must be "+want)
	}
}

func requireField(errs *[]string, field, value string) {
	if strings.TrimSpace(value) == "" {
		*errs = append(*errs, field+" must not be empty")
	}
}

func requireList(errs *[]string, field string, values []string) {
	if len(values) == 0 {
		*errs = append(*errs, field+" must not be empty")
	}
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			*errs = append(*errs, fmt.Sprintf("%s[%d] must not be empty", field, i))
		}
	}
}

func checkPublicPathMap(errs *[]string, values map[string]string) {
	for key, value := range values {
		checkPublicPath(errs, key, value, false)
	}
}

func checkPublicPathMapStrict(errs *[]string, values map[string]string) {
	for key, value := range values {
		requireField(errs, key, value)
		checkPublicPath(errs, key, value, true)
	}
}

func checkPublicStrings(errs *[]string, field string, values []string, rejectAbsolute bool) {
	for i, value := range values {
		checkPublicPath(errs, fmt.Sprintf("%s[%d]", field, i), value, rejectAbsolute)
	}
}

func checkPublicPath(errs *[]string, field, value string, rejectAbsolute bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	normalized := strings.ReplaceAll(value, "\\", "/")
	lower := strings.ToLower(normalized)
	for _, marker := range []string{
		"/" + "users/",
		"/" + "home/",
		"/" + "tmp/",
		"/" + "var/folders/",
		"downloads/",
		"file:" + "//",
		".ssh/",
		".aws/",
		".config/",
	} {
		if strings.Contains(lower, marker) {
			*errs = append(*errs, field+" contains a private or machine-local path")
			return
		}
	}
	if rejectAbsolute && (strings.HasPrefix(normalized, "/") || driveAbsPattern.MatchString(value)) {
		*errs = append(*errs, field+" must not be an absolute local path")
	}
}

func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func containsValue(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func joinErrors(errs []string) error {
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

func DefaultInstance(id, stateRoot, toolchainRoot string) Instance {
	cleanState := filepath.ToSlash(filepath.Clean(stateRoot))
	root := func(name string) string {
		return filepath.ToSlash(filepath.Join(cleanState, id, name))
	}
	if runtime.GOOS == "windows" {
		cleanState = strings.ReplaceAll(cleanState, "\\", "/")
	}
	return Instance{
		ContractVersion: InstanceContract,
		ID:              id,
		StateRoot:       cleanState,
		ToolchainRoot:   filepath.ToSlash(filepath.Clean(toolchainRoot)),
		Roots: map[string]string{
			"mission":   root("mission"),
			"workgraph": root("workgraph"),
			"context":   root("context"),
			"evidence":  root("evidence"),
			"worktree":  root("worktree"),
		},
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func rootsUnderStateRoot(instance Instance) bool {
	stateRoot := strings.TrimSuffix(filepath.ToSlash(filepath.Clean(instance.StateRoot)), "/")
	if stateRoot == "." || stateRoot == "" {
		return false
	}
	for _, key := range []string{"mission", "workgraph", "context", "evidence", "worktree"} {
		root := filepath.ToSlash(filepath.Clean(instance.Roots[key]))
		if root != stateRoot && !strings.HasPrefix(root, stateRoot+"/") {
			return false
		}
	}
	return true
}

func digestRunLink(link RunLink) string {
	payload := struct {
		ContractVersion string            `json:"contract_version"`
		TaskID          string            `json:"task_id"`
		Status          string            `json:"status"`
		Evidence        map[string]string `json:"evidence"`
	}{
		ContractVersion: link.ContractVersion,
		TaskID:          link.TaskID,
		Status:          link.Status,
		Evidence:        link.Evidence,
	}
	data, _ := json.Marshal(payload)
	return DigestBytes(data)
}
