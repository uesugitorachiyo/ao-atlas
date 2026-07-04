#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

OUT="target/production-readiness"
BIN="$OUT/atlas"
mkdir -p "$OUT"

checks=0
pass() {
  checks=$((checks + 1))
  printf 'ok %s\n' "$1"
}

go test ./...
pass "go-test"

go vet ./...
pass "go-vet"

go build -o "$BIN" ./cmd/atlas
pass "go-build"

for script in scripts/*.sh; do
  bash -n "$script"
done
pass "script-syntax"

required_files=(
  README.md
  LICENSE
  docs/sdd/AO-ATLAS-PRD.md
  docs/sdd/AO-ATLAS-ARCHITECTURE.md
  docs/sdd/AO-ATLAS-CONTRACTS.md
  docs/sdd/AO-ATLAS-STACK-INSTANCES.md
  docs/sdd/AO-ATLAS-WORKGRAPH.md
  docs/sdd/AO-ATLAS-CONTEXT-PACKS.md
  docs/sdd/AO-ATLAS-FOUNDRY-HANDOFF.md
  docs/sdd/AO-ATLAS-IMPLEMENTATION-SLICES.md
  docs/sdd/AO-ATLAS-ACCEPTANCE-GATES.md
  docs/sdd/AO-ATLAS-SDD-HANDOFF.md
  docs/sdd/AO-ATLAS-LONG-RUN-RECOMMENDATIONS.md
  schemas/stack-instance.schema.json
  schemas/atlas-registry.schema.json
  schemas/instance-doctor.schema.json
  schemas/intake.schema.json
  schemas/mission-status.schema.json
  schemas/blueprint-request.schema.json
  schemas/blueprint-import.schema.json
  schemas/blueprint-candidate-selection.schema.json
  schemas/workgraph.schema.json
  schemas/workgraph-repair-plan.schema.json
  schemas/mutation-classes.schema.json
  schemas/factory-task.schema.json
  schemas/factory-materialization.schema.json
  schemas/context-pack.schema.json
  schemas/foundry-handoff.schema.json
  schemas/foundry-import.schema.json
  schemas/foundry-continuation-handoff.schema.json
  schemas/run-link.schema.json
  schemas/ao-mission-import.schema.json
  schemas/ao-mission-feature-depth-recommendations.schema.json
  schemas/recommendation-wave.schema.json
  schemas/recommendation-readback.schema.json
)
for file in "${required_files[@]}"; do
  test -s "$file"
done
pass "required-docs-and-contracts"

for file in schemas/*.json examples/valid/*.json examples/invalid/*.json; do
  jq -e . "$file" >/dev/null
done
pass "json-syntax"

"$BIN" instance validate --instance examples/valid/stack-instance.json >/dev/null
"$BIN" instance doctor --instance examples/valid/stack-instance.json --registry examples/valid/atlas-registry.json --out "$OUT/instance-doctor.json" >/dev/null
test -s "$OUT/instance-doctor.json"
"$BIN" instance doctor --instance examples/valid/stack-instance.json --json >/dev/null
"$BIN" intake validate --intake examples/valid/intake.json >/dev/null
"$BIN" mission status --intake examples/valid/intake.json --workgraph examples/valid/workgraph-completed.json --run-link examples/valid/run-link.json --out "$OUT/mission-status.json" >/dev/null
test -s "$OUT/mission-status.json"
"$BIN" mission status --intake examples/valid/intake.json --workgraph examples/valid/workgraph.json --run-link examples/valid/run-link-needs-context.json --json >/dev/null
"$BIN" mission import \
  --record examples/valid/ao-mission/mission-record.json \
  --command-status examples/valid/ao-mission/command-status.json \
  --artifact-manifest examples/valid/ao-mission/artifact-manifest.json \
  --route-history examples/valid/ao-mission/route-history.json \
  --scheduler-recovery examples/valid/ao-mission/scheduler-recovery-readback.json \
  --ledger-compaction examples/valid/ao-mission/ledger-compaction-readback.json \
  --out "$OUT/ao-mission-import.json" >/dev/null
test -s "$OUT/ao-mission-import.json"
"$BIN" mission recommendations import --recommendations examples/valid/ao-mission/feature-depth-recommendations.json --target-instance demo-stack --min-tasks 30 --node-budget 40 --min-minutes 120 --max-minutes 180 --continue-if-fast-target 40 --out "$OUT/mission-recommendations" >/dev/null
test -s "$OUT/mission-recommendations/recommendation-wave.json"
test -s "$OUT/mission-recommendations/recommendation-workgraph.json"
test -s "$OUT/mission-recommendations/recommendation-readback.json"
test -s "$OUT/mission-recommendations/next-recommended-prompt.md"
jq -e '.minimum_tasks == 30 and .total_tasks == 40 and .node_budget == 40 and .estimated_minutes == 120 and .supervisor.min_minutes == 120 and .supervisor.max_minutes == 180 and .supervisor.continue_if_fast_target == 40 and .final_response_allowed == false' "$OUT/mission-recommendations/recommendation-wave.json" >/dev/null
jq -e '(.nodes | length) == 40' "$OUT/mission-recommendations/recommendation-workgraph.json" >/dev/null
jq -e '.total_nodes == 40 and .minimum_nodes == 30 and .ready_nodes == 40 and .executable_ready_nodes == 1 and .final_response_allowed == false and .lease_health_status == "minimum_unmet" and .early_return_risk_status == "blocked_final_response_ready_nodes_remain"' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
grep -q "Target 2-3 hours" "$OUT/mission-recommendations/next-recommended-prompt.md"
"$BIN" mission recommendations readback --wave "$OUT/mission-recommendations/recommendation-wave.json" --workgraph "$OUT/mission-recommendations/recommendation-workgraph.json" --evidence-root target/production-readiness/mission-recommendations --out "$OUT/mission-recommendations/recommendation-readback-regenerated.json" >/dev/null
test -s "$OUT/mission-recommendations/recommendation-readback-regenerated.json"
"$BIN" workgraph validate --workgraph "$OUT/mission-recommendations/recommendation-workgraph.json" >/dev/null
"$BIN" foundry import --workgraph "$OUT/mission-recommendations/recommendation-workgraph.json" --instance examples/valid/stack-instance.json --node mission-recommendation-next-01 --out "$OUT/mission-recommendations-foundry-import" >/dev/null
test -s "$OUT/mission-recommendations-foundry-import/foundry-import.json"
node_evidence_dir="$OUT/mission-recommendations-node-01-evidence"
mkdir -p "$node_evidence_dir"
for key in node_gate candidate_record rollback_record implementation_evidence tests verification sentinel_public_safety promoter_no_promotion command_readback foundry_import checkpoint_bundle; do
  printf '{"status":"recorded","key":"%s"}\n' "$key" >"$node_evidence_dir/$key.json"
done
"$BIN" run-link attach \
  --task-id mission-recommendation-next-01-task \
  --status completed \
  --evidence node_gate="$node_evidence_dir/node_gate.json" \
  --evidence candidate_record="$node_evidence_dir/candidate_record.json" \
  --evidence rollback_record="$node_evidence_dir/rollback_record.json" \
  --evidence implementation_evidence="$node_evidence_dir/implementation_evidence.json" \
  --evidence tests="$node_evidence_dir/tests.json" \
  --evidence verification="$node_evidence_dir/verification.json" \
  --evidence sentinel_public_safety="$node_evidence_dir/sentinel_public_safety.json" \
  --evidence promoter_no_promotion="$node_evidence_dir/promoter_no_promotion.json" \
  --evidence command_readback="$node_evidence_dir/command_readback.json" \
  --evidence foundry_import="$node_evidence_dir/foundry_import.json" \
  --evidence checkpoint_bundle="$node_evidence_dir/checkpoint_bundle.json" \
  --out "$OUT/mission-recommendations-node-01-run-link.json" >/dev/null
"$BIN" mission recommendations complete-node \
  --wave "$OUT/mission-recommendations/recommendation-wave.json" \
  --workgraph "$OUT/mission-recommendations/recommendation-workgraph.json" \
  --run-link "$OUT/mission-recommendations-node-01-run-link.json" \
  --expected-node mission-recommendation-next-01 \
  --evidence-root . \
  --readback-evidence-root target/production-readiness/mission-recommendations \
  --out-workgraph "$OUT/mission-recommendations/recommendation-workgraph-after-node-01.json" \
  --out-readback "$OUT/mission-recommendations/recommendation-readback-after-node-01.json" \
  --out-execution-readback "$OUT/mission-recommendations/execution-readback-after-node-01.json" >/dev/null
test -s "$OUT/mission-recommendations/recommendation-workgraph-after-node-01.json"
test -s "$OUT/mission-recommendations/recommendation-readback-after-node-01.json"
test -s "$OUT/mission-recommendations/execution-readback-after-node-01.json"
jq -e '.completed_nodes == 1 and .ready_nodes == 39 and .first_executable_node == "mission-recommendation-next-02" and .final_response_allowed == false' "$OUT/mission-recommendations/recommendation-readback-after-node-01.json" >/dev/null
jq -e '.completed_recommendation_nodes == 1 and .generated_workgraph.ready_nodes == 39 and .generated_workgraph.executable_ready_nodes == 1 and .generated_workgraph.first_executable_node == "mission-recommendation-next-02" and .generated_workgraph.final_response_allowed == false' "$OUT/mission-recommendations/execution-readback-after-node-01.json" >/dev/null
"$BIN" blueprint-request validate --request examples/valid/blueprint-request.json >/dev/null
"$BIN" blueprint import \
  --pack examples/valid/blueprint-import-low-risk-code/blueprint-pack \
  --authorization examples/valid/blueprint-import-low-risk-code/build-authorization.json \
  --instance examples/valid/stack-instance.json \
  --mutation-classes examples/valid/mutation-classes.json \
  --out "$OUT/blueprint-import-low-risk-code" >/dev/null
test -s "$OUT/blueprint-import-low-risk-code/blueprint-import.json"
test -s "$OUT/blueprint-import-low-risk-code/workgraph.json"
test -s "$OUT/blueprint-import-low-risk-code/foundry-import/foundry-import.json"
test -s "$OUT/blueprint-import-low-risk-code/foundry-import/foundry-continuation-handoff.json"
test -s "$OUT/blueprint-import-low-risk-code/foundry-import/foundry-continuation-prompt.md"
grep -q "Move to AO Foundry" "$OUT/blueprint-import-low-risk-code/foundry-import/foundry-continuation-prompt.md"
grep -q "Run codex --yolo" "$OUT/blueprint-import-low-risk-code/foundry-import/foundry-continuation-prompt.md"
grep -q "Paste this prompt" "$OUT/blueprint-import-low-risk-code/foundry-import/foundry-continuation-prompt.md"
"$BIN" mutation-classes validate --model examples/valid/mutation-classes.json >/dev/null
"$BIN" factory-task validate --task examples/valid/factory-task.json >/dev/null
"$BIN" factory materialize --task examples/valid/factory-task.json --out "$OUT/factory-materialization" --dry-run >/dev/null
test -s "$OUT/factory-materialization/materialization.json"
"$BIN" workgraph validate --workgraph examples/valid/workgraph.json >/dev/null
"$BIN" workgraph validate --workgraph examples/valid/workgraph-large-stress.json >/dev/null
"$BIN" workgraph next --workgraph examples/valid/workgraph.json --json >/dev/null
"$BIN" workgraph materialize-next --workgraph examples/valid/workgraph.json --out "$OUT/workgraph-next-materialization" --dry-run >/dev/null
test -s "$OUT/workgraph-next-materialization/materialization.json"
"$BIN" workgraph complete --workgraph examples/valid/workgraph.json --run-link examples/valid/run-link.json --out "$OUT/workgraph-completed.json" >/dev/null
"$BIN" workgraph validate --workgraph "$OUT/workgraph-completed.json" >/dev/null
"$BIN" workgraph repair-plan --workgraph examples/valid/workgraph.json --run-link examples/invalid/run-link-blocked.json --out "$OUT/workgraph-repair-plan-blocked.json" >/dev/null
test -s "$OUT/workgraph-repair-plan-blocked.json"
"$BIN" workgraph repair-plan --workgraph examples/valid/workgraph.json --run-link examples/valid/run-link-failed.json --out "$OUT/workgraph-repair-plan-failed.json" >/dev/null
test -s "$OUT/workgraph-repair-plan-failed.json"
"$BIN" workgraph status --workgraph examples/valid/workgraph.json >/dev/null
"$BIN" context-pack validate --pack examples/valid/context-pack.json >/dev/null
"$BIN" context-pack validate --pack examples/valid/context-pack-repacked.json >/dev/null
"$BIN" context-pack repack \
  --task examples/valid/factory-task.json \
  --run-link examples/valid/run-link-needs-context.json \
  --source-ref docs/sdd/AO-ATLAS-CONTEXT-PACKS.md \
  --source-digest sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
  --budget 4096 \
  --out "$OUT/context-pack-repacked.json" >/dev/null
"$BIN" context-pack validate --pack "$OUT/context-pack-repacked.json" >/dev/null
"$BIN" foundry handoff emit --workgraph examples/valid/workgraph.json --out "$OUT/foundry-handoff.json" >/dev/null
"$BIN" foundry import --workgraph examples/valid/workgraph.json --instance examples/valid/stack-instance.json --out "$OUT/foundry-import" >/dev/null
test -s "$OUT/foundry-import/foundry-import.json"
test -s "$OUT/foundry-import/foundry-continuation-handoff.json"
test -s "$OUT/foundry-import/foundry-continuation-prompt.md"
grep -q "Move to AO Foundry" "$OUT/foundry-import/foundry-continuation-prompt.md"
grep -q "Run codex --yolo" "$OUT/foundry-import/foundry-continuation-prompt.md"
grep -q "Paste this prompt" "$OUT/foundry-import/foundry-continuation-prompt.md"
grep -q "Stop only on done, final denial, hard blocker, CI failure, unsafe scope drift, or kill switch." "$OUT/foundry-import/foundry-continuation-prompt.md"
if grep -Eq 'cat .*(foundry-import|foundry-continuation)' "$OUT/foundry-import/foundry-continuation-prompt.md"; then
  echo "Foundry continuation prompt used inspection-only cat action" >&2
  exit 1
fi
"$BIN" foundry import --workgraph examples/valid/workgraph.json --instance examples/valid/stack-instance.json --node readiness-ready --json >/dev/null
"$BIN" foundry import --workgraph examples/valid/workgraph-multiple-ready.json --instance examples/valid/stack-instance.json --out "$OUT/foundry-import-multiple" >/dev/null
test -s "$OUT/foundry-import-multiple/foundry-import.json"
"$BIN" foundry import --workgraph examples/valid/workgraph-large-stress.json --instance examples/valid/stack-instance.json --out "$OUT/foundry-import-large-stress" >/dev/null
test -s "$OUT/foundry-import-large-stress/foundry-import.json"
"$BIN" run-link validate --run-link examples/valid/run-link.json >/dev/null
"$BIN" run-link attach \
  --task-id atlas-readiness-task \
  --status completed \
  --evidence foundry=evidence/foundry/atlas-readiness.json \
  --evidence forge=evidence/forge/atlas-readiness.json \
  --evidence ao2=evidence/ao2/atlas-readiness.json \
  --out "$OUT/run-link-attached.json" >/dev/null
"$BIN" run-link validate --run-link "$OUT/run-link-attached.json" >/dev/null
pass "valid-fixtures"

if "$BIN" context-pack validate --pack examples/invalid/context-pack-bad-digest.json >/dev/null 2>&1; then
  echo "invalid context pack was accepted" >&2
  exit 1
fi
if "$BIN" instance doctor --instance examples/valid/stack-instance.json --registry examples/invalid/atlas-registry-parity-mismatch.json --out "$OUT/instance-doctor-mismatch.json" >/dev/null 2>&1; then
  echo "instance doctor accepted registry parity mismatch" >&2
  exit 1
fi
if "$BIN" instance doctor --instance examples/invalid/stack-instance-public-state-root.json --registry examples/valid/atlas-registry.json --out "$OUT/instance-doctor-public-state-root.json" >/dev/null 2>&1; then
  echo "instance doctor accepted public tracked state root" >&2
  exit 1
fi
if "$BIN" instance doctor --instance examples/valid/stack-instance.json --registry examples/invalid/atlas-registry-claims-authority.json --out "$OUT/instance-doctor-authority-claim.json" >/dev/null 2>&1; then
  echo "instance doctor accepted Atlas scheduling authority claim" >&2
  exit 1
fi
if "$BIN" workgraph validate --workgraph examples/invalid/workgraph-missing-dependency.json >/dev/null 2>&1; then
  echo "invalid workgraph was accepted" >&2
  exit 1
fi
if "$BIN" blueprint-request validate --request examples/invalid/blueprint-request-ready-status.json >/dev/null 2>&1; then
  echo "invalid blueprint request was accepted" >&2
  exit 1
fi
if "$BIN" mission recommendations import --recommendations examples/invalid/feature-depth-recommendations-shallow.json --target-instance demo-stack --min-tasks 20 --node-budget 20 --estimated-minutes 90 --out "$OUT/mission-recommendations-shallow" >/dev/null 2>&1; then
  echo "shallow Feature Depth recommendations were accepted" >&2
  exit 1
fi
if "$BIN" mission recommendations import --recommendations examples/invalid/feature-depth-recommendations-unsafe.json --target-instance demo-stack --min-tasks 20 --node-budget 20 --estimated-minutes 90 --out "$OUT/mission-recommendations-unsafe" >/dev/null 2>&1; then
  echo "unsafe Feature Depth recommendations were accepted" >&2
  exit 1
fi
if "$BIN" blueprint import --pack examples/invalid/blueprint-import-missing-authorization/blueprint-pack --instance examples/valid/stack-instance.json --mutation-classes examples/valid/mutation-classes.json --out "$OUT/blueprint-import-missing-auth" >/dev/null 2>&1; then
  echo "missing Blueprint authorization emitted ready import material" >&2
  exit 1
fi
test -s "$OUT/blueprint-import-missing-auth/blueprint-request.json"
if "$BIN" workgraph complete --workgraph examples/valid/workgraph.json --run-link examples/invalid/run-link-blocked.json --out "$OUT/blocked-completed.json" >/dev/null 2>&1; then
  echo "blocked run-link completed a workgraph" >&2
  exit 1
fi
if "$BIN" workgraph complete --workgraph examples/valid/workgraph.json --run-link examples/invalid/run-link-missing-node.json --out "$OUT/missing-node-completed.json" >/dev/null 2>&1; then
  echo "missing-node run-link completed a workgraph" >&2
  exit 1
fi
if "$BIN" workgraph complete --workgraph examples/invalid/workgraph-complete-incomplete-dependency.json --run-link examples/valid/run-link.json --out "$OUT/incomplete-dependency-completed.json" >/dev/null 2>&1; then
  echo "incomplete dependency completed a workgraph" >&2
  exit 1
fi
if "$BIN" workgraph complete --workgraph examples/valid/workgraph.json --run-link examples/valid/run-link.json --out examples/valid/workgraph.json >/dev/null 2>&1; then
  echo "same input/output workgraph completion was accepted" >&2
  exit 1
fi
if "$BIN" workgraph repair-plan --workgraph examples/valid/workgraph.json --run-link examples/valid/run-link.json --out "$OUT/completed-repair-plan.json" >/dev/null 2>&1; then
  echo "completed run-link emitted a repair plan" >&2
  exit 1
fi
if "$BIN" workgraph repair-plan --workgraph examples/valid/workgraph.json --run-link examples/invalid/run-link-missing-node-blocked.json --out "$OUT/missing-node-repair-plan.json" >/dev/null 2>&1; then
  echo "missing-node run-link emitted a repair plan" >&2
  exit 1
fi
if "$BIN" context-pack repack --task examples/valid/factory-task.json --run-link examples/valid/run-link.json --source-ref docs/sdd/AO-ATLAS-CONTEXT-PACKS.md --source-digest sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --budget 4096 --out "$OUT/completed-context-repack.json" >/dev/null 2>&1; then
  echo "completed run-link emitted a context repack" >&2
  exit 1
fi
if "$BIN" context-pack repack --task examples/valid/factory-task.json --run-link examples/invalid/run-link-blocked.json --source-ref docs/sdd/AO-ATLAS-CONTEXT-PACKS.md --source-digest sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --budget 4096 --out "$OUT/missing-context-repack.json" >/dev/null 2>&1; then
  echo "blocked run-link without needs_context emitted a context repack" >&2
  exit 1
fi
if "$BIN" foundry import --workgraph examples/valid/workgraph.json --instance examples/valid/stack-instance.json --out examples/valid/workgraph.json >/dev/null 2>&1; then
  echo "same input/output foundry import was accepted" >&2
  exit 1
fi
if "$BIN" foundry import --workgraph examples/invalid/workgraph-foundry-import-blocked-node.json --instance examples/valid/stack-instance.json --node blocked-node --out "$OUT/foundry-import-blocked" >/dev/null 2>&1; then
  echo "blocked node foundry import was accepted" >&2
  exit 1
fi
if "$BIN" foundry import --workgraph examples/invalid/workgraph-foundry-import-incomplete-dependency.json --instance examples/valid/stack-instance.json --node blocked-by-dependency --out "$OUT/foundry-import-incomplete-dependency" >/dev/null 2>&1; then
  echo "incomplete dependency foundry import was accepted" >&2
  exit 1
fi
if "$BIN" foundry import --workgraph examples/invalid/workgraph-foundry-import-missing-context.json --instance examples/valid/stack-instance.json --out "$OUT/foundry-import-missing-context" >/dev/null 2>&1; then
  echo "missing context pack foundry import was accepted" >&2
  exit 1
fi
if "$BIN" foundry import --workgraph examples/invalid/workgraph-foundry-import-unsafe-path.json --instance examples/valid/stack-instance.json --out "$OUT/foundry-import-unsafe-path" >/dev/null 2>&1; then
  echo "unsafe path foundry import was accepted" >&2
  exit 1
fi
if "$BIN" foundry import --workgraph examples/invalid/workgraph-foundry-import-missing-mutation-class.json --instance examples/valid/stack-instance.json --out "$OUT/foundry-import-missing-mutation-class" >/dev/null 2>&1; then
  echo "missing mutation class foundry import was accepted" >&2
  exit 1
fi
if "$BIN" foundry import --workgraph examples/invalid/workgraph-foundry-import-missing-required-gates.json --instance examples/valid/stack-instance.json --out "$OUT/foundry-import-missing-required-gates" >/dev/null 2>&1; then
  echo "missing required gates foundry import was accepted" >&2
  exit 1
fi
pass "invalid-fixtures-rejected"

while IFS= read -r execution_readback; do
  readback_dir="$(dirname "$execution_readback")"
  recommendation_readback="$readback_dir/recommendation-readback.json"
  if [ ! -s "$recommendation_readback" ]; then
    continue
  fi
  if ! jq -e 'has("completed_recommendation_nodes")' "$execution_readback" >/dev/null; then
    continue
  fi
  execution_completed="$(jq -r '.completed_recommendation_nodes' "$execution_readback")"
  recommendation_completed="$(jq -r '.completed_nodes' "$recommendation_readback")"
  execution_total="$(jq -r '.total_recommendation_nodes' "$execution_readback")"
  recommendation_total="$(jq -r '.total_nodes' "$recommendation_readback")"
  execution_status="$(jq -r '.status' "$execution_readback")"
  recommendation_final_allowed="$(jq -r '.final_response_allowed' "$recommendation_readback")"
  if [ "$execution_completed" != "$recommendation_completed" ]; then
    echo "execution readback $execution_readback claims completed_recommendation_nodes=$execution_completed but recommendation readback has completed_nodes=$recommendation_completed" >&2
    exit 1
  fi
  if [ "$execution_total" != "$recommendation_total" ]; then
    echo "execution readback $execution_readback total_recommendation_nodes does not match recommendation readback total_nodes" >&2
    exit 1
  fi
  if [ "$execution_status" = "completed" ] && [ "$recommendation_final_allowed" != "true" ]; then
    echo "execution readback $execution_readback cannot use status=completed while recommendation final_response_allowed is false" >&2
    exit 1
  fi
done < <(find docs/evidence -name execution-readback.json -type f)
pass "recommendation-ledger-consistency"

scan_file="$OUT/public-safety-files.txt"
find . \
  -path './.git' -prune -o \
  -path './target' -prune -o \
  -path './.atlas-local' -prune -o \
  -path './.atlas-state' -prune -o \
  -type f -print > "$scan_file"

patterns=(
  '/'"Users/"
  '/'"home/"
  '/'"tmp/"
  '/'"private/"
  "Downloads""/"
  "file:"'//'
  "OPENAI_""API_KEY"
  "ANTHROPIC_""API_KEY"
  "BEGIN ""PRIVATE KEY"
  "aws_""secret_access_key"
)

while IFS= read -r file; do
  for pattern in "${patterns[@]}"; do
    if grep -nF "$pattern" "$file" >/dev/null; then
      echo "public safety marker '$pattern' found in $file" >&2
      exit 1
    fi
  done
done < "$scan_file"
pass "public-safety-scan"

git diff --check
pass "git-diff-check"

cat > "$OUT/summary.json" <<JSON
{
  "schema_version": "ao.atlas.production-readiness.v0.1",
  "status": "ready",
  "score": "100/100",
  "checks": $checks
}
JSON

echo "status=ready"
echo "score=100/100"
echo "summary=$OUT/summary.json"
