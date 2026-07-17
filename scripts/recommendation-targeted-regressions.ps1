param(
    [string]$Suite = "all"
)

$ErrorActionPreference = "Stop"

function Invoke-GoSuite {
    param(
        [string]$Name,
        [string]$Regex
    )

    Write-Output "suite=$Name"
    & go test ./internal/atlas -run $Regex -count=1
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

function Write-SuiteList {
    @(
        "validator-boundaries"
        "fixture-builders"
        "compact-dashboard"
        "command-ledger"
        "all"
    ) | Write-Output
}

Set-Location (Resolve-Path (Join-Path $PSScriptRoot ".."))

switch ($Suite) {
    "list" {
        Write-SuiteList
    }
    "validator-boundaries" {
        Invoke-GoSuite $Suite 'Test(PromoterNoPromotionRollupValidatorRejectsPromotionAndRSIBoundaryDrift|CommandPromoterAgreementRollupValidatorRejectsPromotionAndRSIBoundaryDrift)'
    }
    "fixture-builders" {
        Invoke-GoSuite $Suite 'TestRecommendationTestFixtureBuildersCoverWaveNodeAndReadbackDomains'
    }
    "compact-dashboard" {
        Invoke-GoSuite $Suite 'TestMissionDashboardCompactFilters(RenderCompletedWaveWithNoReadyNodes|IncludeTrackCIAndCleanupStateRows|CarrySchemaHealthStatusWhenReadbackHasIt|ClassifySchemaHealthFilterStates)'
    }
    "command-ledger" {
        Invoke-GoSuite $Suite 'TestRecommendation(RunLedgerOutputStatusClassificationCoversPassReadyFailedBlocked|RunLedgerRollupBindsOperatorSummaryWithoutSelfReference|RunLedgerRetryFixturePackCoversRetriesAndResumedSessions)'
    }
    "all" {
        foreach ($Name in @("validator-boundaries", "fixture-builders", "compact-dashboard", "command-ledger")) {
            & $PSCommandPath $Name
            if ($LASTEXITCODE -ne 0) {
                exit $LASTEXITCODE
            }
        }
    }
    default {
        [Console]::Error.WriteLine("unknown recommendation targeted regression suite: $Suite")
        [Console]::Error.WriteLine("available suites:")
        Write-SuiteList | ForEach-Object { [Console]::Error.WriteLine($_) }
        exit 2
    }
}
