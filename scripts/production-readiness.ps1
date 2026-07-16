$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Root

$Out = Join-Path "target" "production-readiness-windows"
$Bin = Join-Path $Out "atlas.exe"
New-Item -ItemType Directory -Force -Path $Out | Out-Null

$Checks = 0
function Pass($Name) {
    $script:Checks += 1
    Write-Output "ok $Name"
}

function Reject-LocalBuildArtifacts {
    $Artifacts = @(
        "atlas",
        "atlas.exe",
        (Join-Path "cmd" (Join-Path "atlas" "atlas")),
        (Join-Path "cmd" (Join-Path "atlas" "atlas.exe"))
    )
    foreach ($Artifact in $Artifacts) {
        git ls-files --error-unmatch $Artifact *> $null
        if ($LASTEXITCODE -eq 0) {
            throw "tracked build artifact present: $Artifact"
        }
        if (Test-Path $Artifact) {
            throw "local build artifact present: $Artifact"
        }
    }
}

function Assert-JsonSyntax($Path) {
    Get-Content -Raw $Path | ConvertFrom-Json | Out-Null
}

Reject-LocalBuildArtifacts
Pass "build-artifact-guard"

go test ./...
Pass "go-test"

go vet ./...
Pass "go-vet"

go build -o $Bin ./cmd/atlas
Pass "go-build"

$RequiredFiles = @(
    "README.md",
    "LICENSE",
    (Join-Path "docs" (Join-Path "sdd" "AO-ATLAS-PRD.md")),
    (Join-Path "docs" (Join-Path "sdd" "AO-ATLAS-ARCHITECTURE.md")),
    (Join-Path "docs" (Join-Path "sdd" "AO-ATLAS-CONTRACTS.md")),
    (Join-Path "schemas" "stack-instance.schema.json"),
    (Join-Path "schemas" "atlas-registry.schema.json"),
    (Join-Path "schemas" "workgraph.schema.json"),
    (Join-Path "examples" (Join-Path "valid" "stack-instance.json")),
    (Join-Path "examples" (Join-Path "valid" "atlas-registry.json")),
    (Join-Path "examples" (Join-Path "valid" "workgraph.json"))
)
foreach ($File in $RequiredFiles) {
    if (-not (Test-Path $File)) {
        throw "required file missing: $File"
    }
}
Pass "required-docs-and-contracts"

Get-ChildItem -Path "schemas", (Join-Path "examples" "valid"), (Join-Path "examples" "invalid") -Filter "*.json" -File | ForEach-Object {
    Assert-JsonSyntax $_.FullName
}
Pass "json-syntax"

& $Bin instance validate --instance (Join-Path "examples" (Join-Path "valid" "stack-instance.json")) | Out-Null
& $Bin instance doctor `
    --instance (Join-Path "examples" (Join-Path "valid" "stack-instance.json")) `
    --registry (Join-Path "examples" (Join-Path "valid" "atlas-registry.json")) `
    --out (Join-Path $Out "instance-doctor.json") | Out-Null
& $Bin intake validate --intake (Join-Path "examples" (Join-Path "valid" "intake.json")) | Out-Null
& $Bin workgraph validate --workgraph (Join-Path "examples" (Join-Path "valid" "workgraph.json")) | Out-Null
Pass "cli-smoke"

$Summary = @{
    schema_version = "ao.atlas.production-readiness-windows.v0.1"
    status = "passed"
    checks = $Checks
    shell = "pwsh"
    bash_required = $false
    release = $false
    tag = $false
    upload = $false
    deployment = $false
    provider_pilot = $false
    rsi_authorized = $false
}
$Summary | ConvertTo-Json -Depth 4 | Set-Content -Encoding UTF8 (Join-Path $Out "summary.json")
Pass "summary"
