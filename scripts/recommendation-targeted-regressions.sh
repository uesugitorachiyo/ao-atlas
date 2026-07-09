#!/usr/bin/env bash
set -euo pipefail

suite="${1:-all}"

run_go_suite() {
  local name="$1"
  local regex="$2"
  echo "suite=$name"
  go test ./internal/atlas -run "$regex" -count=1
}

list_suites() {
  cat <<'SUITES'
validator-boundaries
fixture-builders
compact-dashboard
command-ledger
all
SUITES
}

case "$suite" in
  list)
    list_suites
    ;;
  validator-boundaries)
    run_go_suite "$suite" 'Test(PromoterNoPromotionRollupValidatorRejectsPromotionAndRSIBoundaryDrift|CommandPromoterAgreementRollupValidatorRejectsPromotionAndRSIBoundaryDrift)'
    ;;
  fixture-builders)
    run_go_suite "$suite" 'TestRecommendationTestFixtureBuildersCoverWaveNodeAndReadbackDomains'
    ;;
  compact-dashboard)
    run_go_suite "$suite" 'TestMissionDashboardCompactFilters(RenderCompletedWaveWithNoReadyNodes|IncludeTrackCIAndCleanupStateRows|CarrySchemaHealthStatusWhenReadbackHasIt|ClassifySchemaHealthFilterStates)'
    ;;
  command-ledger)
    run_go_suite "$suite" 'TestRecommendation(RunLedgerOutputStatusClassificationCoversPassReadyFailedBlocked|RunLedgerRollupBindsOperatorSummaryWithoutSelfReference|RunLedgerRetryFixturePackCoversRetriesAndResumedSessions)'
    ;;
  all)
    "$0" validator-boundaries
    "$0" fixture-builders
    "$0" compact-dashboard
    "$0" command-ledger
    ;;
  *)
    echo "unknown recommendation targeted regression suite: $suite" >&2
    echo "available suites:" >&2
    list_suites >&2
    exit 2
    ;;
esac
