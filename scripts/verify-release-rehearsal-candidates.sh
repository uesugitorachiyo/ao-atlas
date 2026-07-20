#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: verify-release-rehearsal-candidates.sh --candidates-dir <dir> --version <version> --tag <tag> --source-sha <sha> --approved-manifest-digest <digest> --plan-out <path>" >&2
}

fail() {
  echo "release rehearsal verifier: $*" >&2
  exit 1
}

hash_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

check_manifest() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum --check --strict --quiet SHA256SUMS
  else
    shasum -a 256 --check SHA256SUMS >/dev/null
  fi
}

candidates_dir=""
version=""
tag=""
source_sha=""
approved_manifest_digest=""
plan_out=""
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd "$script_dir/.." && pwd)
while [[ $# -gt 0 ]]; do
  case "$1" in
    --candidates-dir)
      candidates_dir="${2:-}"
      shift 2
      ;;
    --version)
      version="${2:-}"
      shift 2
      ;;
    --tag)
      tag="${2:-}"
      shift 2
      ;;
    --source-sha)
      source_sha="${2:-}"
      shift 2
      ;;
    --approved-manifest-digest)
      approved_manifest_digest="${2:-}"
      shift 2
      ;;
    --plan-out)
      plan_out="${2:-}"
      shift 2
      ;;
    *)
      usage
      fail "unknown argument $1"
      ;;
  esac
done

[[ -n "$candidates_dir" && -d "$candidates_dir" ]] || fail "candidates directory is required"
[[ "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]] || fail "version is invalid"
[[ "$tag" == "$version" ]] || fail "tag must exactly match version"
[[ "$source_sha" =~ ^[0-9a-f]{40}$ ]] || fail "source SHA is invalid"
[[ "$approved_manifest_digest" =~ ^sha256:[0-9a-f]{64}$ ]] || fail "approved manifest digest is invalid"
[[ -n "$plan_out" ]] || fail "plan output path is required"

expected_targets=(linux-x86_64 macos-x86_64 windows-x86_64)
work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT
find "$candidates_dir" -name candidate-summary.json -type f | sort > "$work_dir/summaries.txt"
observed_targets=" "

while IFS= read -r summary; do
  [[ -n "$summary" ]] || continue
  jq -e . "$summary" >/dev/null 2>&1 || fail "malformed candidate JSON: $summary"
  target=$(jq -er '.target_label' "$summary") || fail "candidate summary target is missing: $summary"
  if [[ " ${expected_targets[*]} " != *" $target "* ]]; then
    fail "unexpected candidate target: $target"
  fi
  if [[ "$observed_targets" == *" $target "* ]]; then
    fail "duplicate candidate: $target"
  fi
  observed_targets="${observed_targets}${target} "

  [[ "$(jq -r '.version' "$summary")" == "$version" ]] || fail "stale candidate version: $target"
  [[ "$(jq -r '.tag' "$summary")" == "$tag" ]] || fail "stale candidate tag: $target"
  [[ "$(jq -r '.source_sha' "$summary")" == "$source_sha" ]] || fail "stale candidate source: $target"
  [[ "$(jq -r '.approved_manifest_digest' "$summary")" == "$approved_manifest_digest" ]] ||
    fail "stale candidate manifest digest: $target"
  [[ "$(jq -r '.immutable' "$summary")" == "true" ]] || fail "stale candidate immutability: $target"

  case "$target" in
    linux-x86_64)
      expected_goos="linux"
      expected_goarch="amd64"
      expected_binary="ao-atlas"
      ;;
    macos-x86_64)
      expected_goos="darwin"
      expected_goarch="amd64"
      expected_binary="ao-atlas"
      ;;
    windows-x86_64)
      expected_goos="windows"
      expected_goarch="amd64"
      expected_binary="ao-atlas.exe"
      ;;
  esac
  expected_archive="ao-atlas-${version}-${target}.tar.gz"

  candidate_dir=$(dirname "$summary")
  archive=$(jq -er '.archive' "$summary") || fail "candidate archive is missing: $target"
  binary=$(jq -er '.binary' "$summary") || fail "candidate binary is missing: $target"
  [[ "$binary" == "$expected_binary" ]] || fail "executable target substitution: $target"
  [[ "$archive" == "$expected_archive" ]] || fail "archive target substitution: $target"
  [[ "$(jq -r '.goos' "$summary")" == "$expected_goos" &&
    "$(jq -r '.goarch' "$summary")" == "$expected_goarch" ]] ||
    fail "target architecture substitution: $target"
  jq -e \
    --arg archive "$expected_archive" \
    --arg binary "$expected_binary" \
    --arg target "$target" \
    '.schema_version == "ao.atlas.release-rehearsal-candidate.v0.3" and
      .immutable == true and
      .repository == "ao-atlas" and
      .entry_point == "ao-atlas" and
      .target_label == $target and
      .archive == $archive and
      .binary == $binary and
      .license_files == ["LICENSE"] and
      .provider_credentials_used == false and
      .help_readback == "help-readback.json" and
      .version_source_readback == "version-source-readback.json" and
      .provider_free_functional == "provider-free-functional-smoke.json" and
      .installed_archive_smoke == "installed-archive-smoke.json" and
      .provenance == "provenance.json" and
      .sbom == "sbom.spdx.json" and
      .signature_verification == "signature-verification.json" and
      (.go_version | type == "string" and length > 0) and
      (.runner_os | type == "string" and length > 0) and
      (.workflow_identity | type == "string" and test("^https://[^/]+/[^/]+/[^/]+/actions/runs/[0-9]+$"))' \
    "$summary" >/dev/null || fail "invalid candidate summary: $target"
  jq -e . "$candidate_dir/go-env.json" >/dev/null 2>&1 || fail "malformed candidate JSON: $target go-env.json"
  [[ "$(jq -r '.GOOS' "$candidate_dir/go-env.json")" == "$expected_goos" &&
    "$(jq -r '.GOARCH' "$candidate_dir/go-env.json")" == "$expected_goarch" ]] ||
    fail "target architecture substitution: $target"

  sidecar="${archive}.sha256"
  binary_sidecar="${binary}.sha256"
  [[ -f "$candidate_dir/$archive" ]] || fail "missing candidate archive: $target"
  [[ -f "$candidate_dir/$sidecar" ]] || fail "missing checksum sidecar: $target"
  [[ -f "$candidate_dir/$binary" ]] || fail "missing candidate binary: $target"
  [[ -f "$candidate_dir/$binary_sidecar" ]] || fail "missing binary checksum sidecar: $target"
  [[ -f "$candidate_dir/SHA256SUMS" ]] || fail "missing checksum manifest: $target"

  actual_archive_digest=$(hash_file "$candidate_dir/$archive")
  [[ "$(jq -r '.archive_sha256' "$summary")" == "sha256:$actual_archive_digest" ]] ||
    fail "substituted candidate archive: $target"

  expected_archive_line="$actual_archive_digest  $archive"
  printf '%s\n' "$expected_archive_line" > "$work_dir/$target.sidecar.expected"
  cmp -s "$candidate_dir/$sidecar" "$work_dir/$target.sidecar.expected" ||
    fail "checksum sidecar mismatch: $target"
  manifest_archive_lines=$(awk -v archive="$archive" '$2 == archive { print }' "$candidate_dir/SHA256SUMS")
  [[ "$manifest_archive_lines" == "$expected_archive_line" ]] ||
    fail "archive checksum manifest mismatch: $target"

  actual_binary_digest=$(hash_file "$candidate_dir/$binary")
  [[ "$(jq -r '.binary_sha256' "$summary")" == "sha256:$actual_binary_digest" ]] ||
    fail "substituted candidate binary: $target"
  binary_metadata=$(go version -m "$candidate_dir/$binary" 2>/dev/null) ||
    fail "executable target substitution: $target"
  binary_goos=$(printf '%s\n' "$binary_metadata" | awk '$1 == "build" && $2 ~ /^GOOS=/ { sub(/^GOOS=/, "", $2); print $2 }')
  binary_goarch=$(printf '%s\n' "$binary_metadata" | awk '$1 == "build" && $2 ~ /^GOARCH=/ { sub(/^GOARCH=/, "", $2); print $2 }')
  [[ "$binary_goos" == "$expected_goos" && "$binary_goarch" == "$expected_goarch" ]] ||
    fail "executable target substitution: $target"
  expected_binary_line="$actual_binary_digest  $binary"
  printf '%s\n' "$expected_binary_line" > "$work_dir/$target.binary-sidecar.expected"
  cmp -s "$candidate_dir/$binary_sidecar" "$work_dir/$target.binary-sidecar.expected" ||
    fail "binary checksum sidecar mismatch: $target"
  manifest_binary_lines=$(awk -v binary="$binary" '$2 == binary { print }' "$candidate_dir/SHA256SUMS")
  [[ "$manifest_binary_lines" == "$expected_binary_line" ]] ||
    fail "binary checksum manifest mismatch: $target"

  expected_archive_inventory=$(printf '%s\n' LICENSE "$binary" | sort)
  archive_member_file="$work_dir/$target.archive-members"
  archive_type_file="$work_dir/$target.archive-types"
  tar -tzf "$candidate_dir/$archive" > "$archive_member_file" 2>/dev/null ||
    fail "noncanonical archive entry: $target"
  archive_members=$(sort "$archive_member_file")
  [[ "$archive_members" == "$expected_archive_inventory" ]] ||
    fail "noncanonical archive entry: $target"
  tar -tvzf "$candidate_dir/$archive" > "$archive_type_file" 2>/dev/null ||
    fail "noncanonical archive entry: $target"
  [[ "$(wc -l < "$archive_type_file" | tr -d ' ')" == "2" ]] ||
    fail "noncanonical archive entry: $target"
  awk 'substr($1, 1, 1) != "-" { exit 1 }' "$archive_type_file" ||
    fail "noncanonical archive entry: $target"

  archive_extract_dir="$work_dir/archive-$target"
  mkdir -p "$archive_extract_dir"
  tar -xzf "$candidate_dir/$archive" -C "$archive_extract_dir" ||
    fail "archive target substitution: $target"
  archive_inventory=$(
    cd "$archive_extract_dir"
    find . -mindepth 1 -print | sed 's#^\./##' | sort
  )
  [[ "$archive_inventory" == "$expected_archive_inventory" &&
    -f "$archive_extract_dir/$binary" && ! -L "$archive_extract_dir/$binary" &&
    -f "$archive_extract_dir/LICENSE" && ! -L "$archive_extract_dir/LICENSE" ]] ||
    fail "archive target substitution: $target"
  [[ "$(hash_file "$archive_extract_dir/$binary")" == "$actual_binary_digest" ]] ||
    fail "archive target substitution: $target"

  (
    cd "$candidate_dir"
    check_manifest
  ) || fail "candidate checksum verification failed: $target"

  expected_inventory=(
    "$archive"
    "$sidecar"
    "$binary_sidecar"
    SHA256SUMS
    LICENSE
    archive-identity.txt
    candidate-summary.json
    go-env.json
    help-readback.json
    installed-archive-smoke.json
    provider-free-functional-smoke.json
    provenance.json
    sbom.spdx.json
    signature-verification.json
    version-source-readback.json
    "$binary"
  )
  actual_inventory=$(find "$candidate_dir" -maxdepth 1 -type f -exec basename {} \; | sort)
  expected_inventory_text=$(printf '%s\n' "${expected_inventory[@]}" | sort)
  [[ "$actual_inventory" == "$expected_inventory_text" ]] ||
    fail "unexpected candidate inventory: $target"

  [[ "$(jq -r '.help_readback' "$summary")" == "help-readback.json" ]] ||
    fail "missing help readback binding: $target"
  [[ "$(jq -r '.version_source_readback' "$summary")" == "version-source-readback.json" ]] ||
    fail "missing version source readback binding: $target"
  [[ "$(jq -r '.provider_free_functional' "$summary")" == "provider-free-functional-smoke.json" ]] ||
    fail "missing provider-free functional smoke binding: $target"
  [[ "$(jq -r '.installed_archive_smoke' "$summary")" == "installed-archive-smoke.json" ]] ||
    fail "missing installed archive smoke binding: $target"

  expected_help_command="$binary (no arguments)"
  expected_help_output='atlas <instance|intake|blueprint|mission|blueprint-request|workgraph|mutation-classes|factory-task|factory|context-pack|foundry|run-link> ...'
  jq -e \
    --arg command "$expected_help_command" \
    --arg output "$expected_help_output" \
    '.schema_version == "ao.atlas.release-rehearsal-help-readback.v0.1" and
      .status == "passed" and
      .classification == "help_readback_not_functional_smoke" and
      .documented_reference == "internal/atlas/cli.go:usage" and
      .command == $command and
      .expected_exit_code == 2 and
      .actual_exit_code == 2 and
      .output == $output' \
    "$candidate_dir/help-readback.json" >/dev/null 2>&1 ||
    fail "invalid help readback evidence: $target"

  expected_version_command="$binary --version"
  expected_candidate_identity="ao-atlas version=$version source_sha=$source_sha"
  jq -e \
    --arg command "$expected_version_command" \
    --arg identity "$expected_candidate_identity" \
    --arg source_sha "$source_sha" \
    --arg tag "$tag" \
    --arg version "$version" \
    '.schema_version == "ao.atlas.release-rehearsal-version-source-readback.v0.2" and
      .status == "passed" and
      .source == "workflow_dispatch.inputs.version" and
      .repository_version_source == "go.mod" and
      .module_path == "github.com/uesugitorachiyo/ao-atlas" and
      (.go_directive | type == "string" and test("^[0-9]+\\.[0-9]+([.][0-9]+)?$")) and
      .version == $version and
      .source_sha == $source_sha and
      .tag == $tag and
      .tag_matches_version == true and
      (.candidate_command | type == "string") and
      (.candidate_identity | type == "string")' \
    "$candidate_dir/version-source-readback.json" >/dev/null 2>&1 ||
    fail "invalid version source readback evidence: $target"
  [[ "$(jq -r '.candidate_command' "$candidate_dir/version-source-readback.json")" == "$expected_version_command" &&
    "$(jq -r '.candidate_identity' "$candidate_dir/version-source-readback.json")" == "$expected_candidate_identity" ]] ||
    fail "invalid built candidate version identity: $target"

  functional_fixture="examples/valid/workgraph.json"
  functional_command="workgraph validate --workgraph $functional_fixture"
  functional_fixture_digest="sha256:$(hash_file "$repo_root/$functional_fixture")"
  jq -e \
    --arg command "$functional_command" \
    --arg fixture "$functional_fixture" \
    --arg fixture_sha256 "$functional_fixture_digest" \
    '.schema_version == "ao.atlas.release-rehearsal-provider-free-functional-smoke.v0.1" and
      .status == "passed" and
      .command == $command and
      .fixture == $fixture and
      .fixture_sha256 == $fixture_sha256 and
      .installed_archive == true and
      .provider_credentials_used == false' \
    "$candidate_dir/provider-free-functional-smoke.json" >/dev/null 2>&1 ||
    fail "invalid provider-free functional smoke evidence: $target"

  jq -e \
    --arg archive "$archive" \
    --arg binary "$binary" \
    '.schema_version == "ao.atlas.release-rehearsal-installed-archive-smoke.v0.1" and
      .status == "passed" and
      .archive == $archive and
      .binary == $binary and
      .license_present == true' \
    "$candidate_dir/installed-archive-smoke.json" >/dev/null 2>&1 ||
    fail "invalid installed archive smoke evidence: $target"

  workflow_identity=$(jq -er '.workflow_identity' "$summary") ||
    fail "invalid candidate summary: $target"
  jq -e \
    --arg source_sha "$source_sha" \
    --arg target "$target" \
    --arg workflow_identity "$workflow_identity" \
    '.schema_version == "ao.atlas.release-rehearsal-provenance.v0.2" and
      .builder == "github-actions" and
      .source_sha == $source_sha and
      .target_label == $target and
      .workflow_identity == $workflow_identity' \
    "$candidate_dir/provenance.json" >/dev/null 2>&1 ||
    fail "invalid provenance evidence: $target"

  expected_namespace="https://github.com/uesugitorachiyo/ao-atlas/rehearsal/$source_sha/$target"
  jq -e \
    --arg namespace "$expected_namespace" \
    '.SPDXID == "SPDXRef-DOCUMENT" and
      .spdxVersion == "SPDX-2.3" and
      .name == "ao-atlas release rehearsal candidate" and
      .documentNamespace == $namespace and
      .creationInfo == {creators:["Tool: AO Atlas release rehearsal"],created:"1970-01-01T00:00:00Z"} and
      .packages == [{
        SPDXID:"SPDXRef-Package",
        name:"ao-atlas",
        versionInfo:"rehearsal",
        downloadLocation:"NOASSERTION",
        filesAnalyzed:false
      }]' \
    "$candidate_dir/sbom.spdx.json" >/dev/null 2>&1 ||
    fail "invalid SBOM evidence: $target"

  jq -e \
    '.schema_version == "ao.atlas.release-rehearsal-signature-verification.v0.1" and
      .signature_present == false and
      .verification_status == "not_applicable_rehearsal_no_signing_key"' \
    "$candidate_dir/signature-verification.json" >/dev/null 2>&1 ||
    fail "invalid signature verification evidence: $target"

  checksum_manifest_sha256="sha256:$(hash_file "$candidate_dir/SHA256SUMS")"
  help_readback_sha256="sha256:$(hash_file "$candidate_dir/help-readback.json")"
  installed_archive_smoke_sha256="sha256:$(hash_file "$candidate_dir/installed-archive-smoke.json")"
  provider_free_functional_sha256="sha256:$(hash_file "$candidate_dir/provider-free-functional-smoke.json")"
  provenance_sha256="sha256:$(hash_file "$candidate_dir/provenance.json")"
  sbom_sha256="sha256:$(hash_file "$candidate_dir/sbom.spdx.json")"
  signature_verification_sha256="sha256:$(hash_file "$candidate_dir/signature-verification.json")"
  version_source_readback_sha256="sha256:$(hash_file "$candidate_dir/version-source-readback.json")"

  jq -cS \
    --arg candidate_identity "$expected_candidate_identity" \
    --arg checksum_manifest_sha256 "$checksum_manifest_sha256" \
    --arg functional_fixture "$functional_fixture" \
    --arg functional_fixture_sha256 "$functional_fixture_digest" \
    --arg functional_smoke_command "$functional_command" \
    --arg help_command "$expected_help_command" \
    --arg help_readback_sha256 "$help_readback_sha256" \
    --arg installed_archive_smoke_sha256 "$installed_archive_smoke_sha256" \
    --arg provider_free_functional_sha256 "$provider_free_functional_sha256" \
    --arg provenance_sha256 "$provenance_sha256" \
    --arg sbom_sha256 "$sbom_sha256" \
    --arg signature_verification_sha256 "$signature_verification_sha256" \
    --arg version_command "$expected_version_command" \
    --arg version_source_readback_sha256 "$version_source_readback_sha256" \
    '{
      approved_manifest_digest,
      archive,
      archive_sha256,
      binary,
      binary_sha256,
      candidate_identity:$candidate_identity,
      checksum_manifest_sha256:$checksum_manifest_sha256,
      functional_fixture:$functional_fixture,
      functional_fixture_sha256:$functional_fixture_sha256,
      functional_smoke_command:$functional_smoke_command,
      goarch,
      goos,
      help_command:$help_command,
      help_readback,
      help_readback_sha256:$help_readback_sha256,
      installed_archive_smoke,
      installed_archive_smoke_sha256:$installed_archive_smoke_sha256,
      provider_free_functional,
      provider_free_functional_sha256:$provider_free_functional_sha256,
      provenance,
      provenance_sha256:$provenance_sha256,
      sbom,
      sbom_sha256:$sbom_sha256,
      signature_verification,
      signature_verification_sha256:$signature_verification_sha256,
      source_sha,
      target_label,
      tag,
      tag_matches_version:true,
      version,
      version_command:$version_command,
      version_source_readback,
      version_source_readback_sha256:$version_source_readback_sha256
    }' "$summary" > "$work_dir/$target.json"
done < "$work_dir/summaries.txt"

for target in "${expected_targets[@]}"; do
  [[ "$observed_targets" == *" $target "* ]] || fail "missing candidate: $target"
done

mkdir -p "$(dirname "$plan_out")"
jq -cS -s \
  --arg approved_manifest_digest "$approved_manifest_digest" \
  --arg source_sha "$source_sha" \
  --arg tag "$tag" \
  --arg version "$version" \
  '{schema_version:"ao.atlas.release-rehearsal-promotion-plan.v0.4",status:"dry_run_plan_ready",immutable:true,approved_manifest_digest:$approved_manifest_digest,source_sha:$source_sha,tag:$tag,version:$version,candidates:sort_by(.target_label)}' \
  "$work_dir"/*.json > "$plan_out"
[[ -s "$plan_out" ]] || fail "promotion plan was not written"
