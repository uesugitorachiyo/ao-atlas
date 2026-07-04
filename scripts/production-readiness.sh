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

reject_local_absolute_paths() {
  local label="$1"
  shift
  local patterns=(
    '/'"Users/"
    '/'"home/"
    '/'"tmp/"
    '/'"private/"
    "Downloads""/"
    "file:"'//'
  )
  for file in "$@"; do
    test -s "$file"
    for pattern in "${patterns[@]}"; do
      if grep -nF "$pattern" "$file" >/dev/null; then
        echo "$label contains local absolute path marker '$pattern' in $file" >&2
        exit 1
      fi
    done
  done
}

reject_generated_recommendation_prompt_public_safety() {
  local prompt="$1"
  test -s "$prompt"
  reject_local_absolute_paths "generated recommendation prompt" "$prompt"
  local forbidden=(
    "RSI is proven"
    "RSI proven"
    "fully_unsupervised_complex_mutation is proven"
    "provider calls allowed"
    "credential inspection allowed"
    "direct main mutation allowed"
    "release deploy publish upload tag allowed"
  )
  for phrase in "${forbidden[@]}"; do
    if grep -niF "$phrase" "$prompt" >/dev/null; then
      echo "generated recommendation prompt contains unsafe wording '$phrase' in $prompt" >&2
      exit 1
    fi
  done
}

assert_schema_required_fields_present() {
  local schema="$1"
  local artifact="$2"
  local label="$3"
  test -s "$schema"
  test -s "$artifact"
  jq -e --slurpfile schema "$schema" '
    . as $artifact |
    ($schema[0].required // []) as $required |
    [ $required[] as $key | select(($artifact | has($key)) | not) ] as $missing |
    if ($missing | length) == 0 then true
    else error("missing required schema fields: \($missing | join(","))")
    end
  ' "$artifact" >/dev/null || {
    echo "$label does not cover required fields from $schema" >&2
    exit 1
  }
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
  schemas/recommendation-lease-start.schema.json
  schemas/recommendation-checkpoint-readback.schema.json
  schemas/recommendation-command-readback.schema.json
  schemas/recommendation-promoter-readback.schema.json
  schemas/recommendation-foundry-rollup.schema.json
  schemas/recommendation-reconciliation-packet.schema.json
  examples/valid/recommendation-wave-long-run-supervisor.json
  examples/valid/recommendation-lease-start-long-run.json
)
for file in "${required_files[@]}"; do
  test -s "$file"
done
grep -q "## Double-Size Operator Requests" docs/sdd/AO-ATLAS-LONG-RUN-RECOMMENDATIONS.md
grep -q "Ready Atlas-owned tasks skip Blueprint" docs/sdd/AO-ATLAS-LONG-RUN-RECOMMENDATIONS.md
grep -q "Foundry import ownership stays single-node" docs/sdd/AO-ATLAS-LONG-RUN-RECOMMENDATIONS.md
grep -q "A 20-node or" docs/sdd/AO-ATLAS-LONG-RUN-RECOMMENDATIONS.md
grep -q "90-minute wave is too small for this path" docs/sdd/AO-ATLAS-LONG-RUN-RECOMMENDATIONS.md
grep -q "prior wave was too short" README.md
grep -q "Foundry import ownership stays single-node" README.md
pass "required-docs-and-contracts"

for file in schemas/*.json examples/valid/*.json examples/invalid/*.json; do
  jq -e . "$file" >/dev/null
done
pass "json-syntax"

jq -e '.minimum_tasks == 30 and .node_budget == 40 and .estimated_minutes == 120 and .supervisor.min_nodes == 30 and .supervisor.min_minutes == 120 and .supervisor.max_minutes == 180 and .supervisor.continue_if_fast_target == 40 and .supervisor.checkpoint_policy == "after_each_node_or_timed_interval" and .final_response_allowed == false and .safe_to_execute == false and .schedules_work == false and .executes_work == false and .approves_work == false' examples/valid/recommendation-wave-long-run-supervisor.json >/dev/null
jq -e '.schema == "ao.atlas.recommendation-lease-start.v0.1" and .min_minutes == 120 and .max_minutes == 180 and .continue_if_fast_target == 40 and .checkpoint_policy == "after_each_node_or_timed_interval" and .final_response_allowed == false and .schedules_work == false and .executes_work == false and .approves_work == false and .mutates_repositories == false and .calls_providers == false and .claims_authority_advance == false' examples/valid/recommendation-lease-start-long-run.json >/dev/null
pass "recommendation-long-run-examples"

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
"$BIN" mission recommendations import --recommendations examples/valid/ao-mission/feature-depth-recommendations.json --target-instance demo-stack --min-tasks 30 --node-budget 40 --min-minutes 120 --max-minutes 180 --continue-if-fast-target 40 --started-at 2026-07-04T08:00:00-07:00 --out "$OUT/mission-recommendations" >/dev/null
test -s "$OUT/mission-recommendations/recommendation-wave.json"
test -s "$OUT/mission-recommendations/recommendation-workgraph.json"
test -s "$OUT/mission-recommendations/lease-start.json"
test -s "$OUT/mission-recommendations/recommendation-readback.json"
test -s "$OUT/mission-recommendations/next-recommended-prompt.md"
jq -e '.minimum_tasks == 30 and .total_tasks == 40 and .node_budget == 40 and .estimated_minutes == 120 and .supervisor.min_minutes == 120 and .supervisor.max_minutes == 180 and .supervisor.continue_if_fast_target == 40 and .final_response_allowed == false' "$OUT/mission-recommendations/recommendation-wave.json" >/dev/null
jq -e '(.tasks | length) == 40 and all(.tasks[]; (.source_task_digest | test("^sha256:[0-9a-f]{64}$")))' "$OUT/mission-recommendations/recommendation-wave.json" >/dev/null
jq -e '(.nodes | length) == 40' "$OUT/mission-recommendations/recommendation-workgraph.json" >/dev/null
jq -e 'all(.nodes[]; any(.factory_task.required_evidence[]; startswith("source_task_digest:sha256:")))' "$OUT/mission-recommendations/recommendation-workgraph.json" >/dev/null
jq -e '.schema == "ao.atlas.recommendation-lease-start.v0.1" and .started_at == "2026-07-04T08:00:00-07:00" and .min_minutes == 120 and .max_minutes == 180 and .final_response_allowed == false and .schedules_work == false and .executes_work == false and .approves_work == false' "$OUT/mission-recommendations/lease-start.json" >/dev/null
jq -e '.total_nodes == 40 and .minimum_nodes == 30 and .ready_nodes == 40 and .executable_ready_nodes == 1 and .checkpoint_count == 0 and .return_gate_status == "blocked_ready_nodes_remain" and .final_response_allowed == false and .lease_health_status == "minimum_unmet" and .early_return_risk_status == "blocked_final_response_ready_nodes_remain"' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
jq -e '.final_response_denial_gate == "deny_ready_nodes_or_exact_next_action_remain"' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
jq -e '.exact_next_action_readback.status == "continuation_required" and .exact_next_action_readback.action == .exact_next_action and .exact_next_action_readback.next_executable_node == .first_executable_node and .exact_next_action_readback.return_gate_status == .return_gate_status and .exact_next_action_readback.final_response_allowed == .final_response_allowed and .exact_next_action_readback.source == "recommendation_readback"' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
jq -e '([.command_timeline_placeholders[].slot] | sort) == ["checkpoint","exact_next_action","return_gate"] and all(.command_timeline_placeholders[]; .source == "recommendation_readback" and .status == "pending_command_timeline" and .required_before_final_response == true and (.summary | length) > 0)' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
jq -e '([.promoter_no_promotion_placeholders[].slot] | sort) == ["authority_advance","promotion_claim","rsi_boundary"] and all(.promoter_no_promotion_placeholders[]; .source == "recommendation_readback" and .status == "pending_promoter_no_promotion" and .required_before_final_response == true and (.summary | length) > 0)' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
jq -e '([.foundry_terminal_status_examples[].source_status] | sort) == ["blocked","completed","denied","promoted"] and any(.foundry_terminal_status_examples[]; .source_status == "promoted" and .normalized_status == "completed" and .terminal == true and .can_close_mission == true and (.required_readback | contains("RSI remains denied"))) and any(.foundry_terminal_status_examples[]; .source_status == "blocked" and .normalized_status == "blocked" and .terminal == true and .can_close_mission == false)' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
jq -e '([.foundry_denied_terminal_examples[].denial_reason] | sort) == ["forbidden_surface_or_rsi_claim","missing_node_evidence","missing_stop_gate_evidence"] and all(.foundry_denied_terminal_examples[]; .normalized_status == "denied" and .terminal == true and .can_close_mission == true and .requires_exact_missing_evidence == true and .rsi_remains_denied == true and .authority_advance_claimed == false) and any(.foundry_denied_terminal_examples[]; .denial_reason == "missing_node_evidence" and (.required_readback | contains("missing node id"))) and any(.foundry_denied_terminal_examples[]; .denial_reason == "forbidden_surface_or_rsi_claim" and (.required_readback | contains("RSI")))' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
jq -e '(.wave_digest | test("^sha256:[0-9a-f]{64}$")) and (.workgraph_digest | test("^sha256:[0-9a-f]{64}$"))' "$OUT/mission-recommendations/lease-start.json" >/dev/null
jq -e --slurpfile lease "$OUT/mission-recommendations/lease-start.json" '.wave_digest == $lease[0].wave_digest and .workgraph_digest == $lease[0].workgraph_digest' "$OUT/mission-recommendations/recommendation-readback.json" >/dev/null
pass "recommendation-import-artifact-binding"
grep -q "Target 2-3 hours" "$OUT/mission-recommendations/next-recommended-prompt.md"
grep -q "\`early_return_risk_status\`" "$OUT/mission-recommendations/next-recommended-prompt.md"
grep -q "If ready_nodes > 0 or exact_next_action is non-empty, do not produce a final response." "$OUT/mission-recommendations/next-recommended-prompt.md"
grep -q "If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair." "$OUT/mission-recommendations/next-recommended-prompt.md"
reject_generated_recommendation_prompt_public_safety "$OUT/mission-recommendations/next-recommended-prompt.md"
pass "recommendation-prompt-public-safety-scan"
"$BIN" mission recommendations readback --wave "$OUT/mission-recommendations/recommendation-wave.json" --workgraph "$OUT/mission-recommendations/recommendation-workgraph.json" --evidence-root target/production-readiness/mission-recommendations --out "$OUT/mission-recommendations/recommendation-readback-regenerated.json" >/dev/null
test -s "$OUT/mission-recommendations/recommendation-readback-regenerated.json"
"$BIN" workgraph validate --workgraph "$OUT/mission-recommendations/recommendation-workgraph.json" >/dev/null
completed_recommendation_workgraph="$OUT/mission-recommendations/recommendation-workgraph-completed.json"
jq '.nodes |= map(.status = "completed")' "$OUT/mission-recommendations/recommendation-workgraph.json" >"$completed_recommendation_workgraph"
"$BIN" mission recommendations readback \
  --wave "$OUT/mission-recommendations/recommendation-wave.json" \
  --workgraph "$completed_recommendation_workgraph" \
  --evidence-root target/production-readiness/mission-recommendations \
  --elapsed-minutes 22 \
  --started-at 2026-07-04T07:20:20-07:00 \
  --completed-at 2026-07-04T07:42:06-07:00 \
  --lease-timing-mode actual \
  --out "$OUT/mission-recommendations/recommendation-readback-completed-short.json" >/dev/null
jq -e '.completed_nodes == 40 and .ready_nodes == 0 and .checkpoint_count == 40 and .return_gate_status == "blocked_minimum_minutes_unmet" and .elapsed_minutes == 22 and .min_minutes_met == false and .lease_time_status == "minimum_minutes_unmet" and .final_response_allowed == false and .final_response_reason == "minimum lease minutes unmet"' "$OUT/mission-recommendations/recommendation-readback-completed-short.json" >/dev/null
"$BIN" mission recommendations readback \
  --wave "$OUT/mission-recommendations/recommendation-wave.json" \
  --workgraph "$completed_recommendation_workgraph" \
  --evidence-root target/production-readiness/mission-recommendations \
  --out "$OUT/mission-recommendations/recommendation-readback-completed-missing-timing.json" >/dev/null
jq -e '.completed_nodes == 40 and .ready_nodes == 0 and .checkpoint_count == 40 and .return_gate_status == "blocked_lease_timing_missing" and .min_minutes_met == false and .lease_time_status == "lease_timing_missing" and .final_response_allowed == false and .final_response_reason == "minimum lease timing evidence missing"' "$OUT/mission-recommendations/recommendation-readback-completed-missing-timing.json" >/dev/null
"$BIN" mission recommendations readback \
  --wave "$OUT/mission-recommendations/recommendation-wave.json" \
  --workgraph "$completed_recommendation_workgraph" \
  --evidence-root target/production-readiness/mission-recommendations \
  --elapsed-minutes 120 \
  --started-at 2026-07-04T07:20:00-07:00 \
  --completed-at 2026-07-04T09:20:00-07:00 \
  --lease-timing-mode actual \
  --out "$OUT/mission-recommendations/recommendation-readback-completed-lease-met.json" >/dev/null
jq -e '.completed_nodes == 40 and .ready_nodes == 0 and .checkpoint_count == 40 and .return_gate_status == "final_response_allowed" and .elapsed_minutes == 120 and .min_minutes_met == true and .lease_time_status == "minimum_minutes_met" and .final_response_allowed == true and .final_response_reason == "all generated nodes complete and no ready nodes remain"' "$OUT/mission-recommendations/recommendation-readback-completed-lease-met.json" >/dev/null
jq -e 'has("started_at") and has("completed_at")' "$OUT/mission-recommendations/recommendation-readback-completed-lease-met.json" >/dev/null
"$BIN" foundry import --workgraph "$OUT/mission-recommendations/recommendation-workgraph.json" --instance examples/valid/stack-instance.json --node mission-recommendation-next-01 --out "$OUT/mission-recommendations-foundry-import" >/dev/null
test -s "$OUT/mission-recommendations-foundry-import/foundry-import.json"
reject_local_absolute_paths "mission recommendations Foundry continuation" \
  "$OUT/mission-recommendations-foundry-import/foundry-continuation-handoff.json" \
  "$OUT/mission-recommendations-foundry-import/foundry-continuation-prompt.md"
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
  --lease-start "$OUT/mission-recommendations/lease-start.json" \
  --completed-at 2026-07-04T08:17:00-07:00 \
  --out-workgraph "$OUT/mission-recommendations/recommendation-workgraph-after-node-01.json" \
  --out-readback "$OUT/mission-recommendations/recommendation-readback-after-node-01.json" \
  --out-execution-readback "$OUT/mission-recommendations/execution-readback-after-node-01.json" \
  --out-checkpoint-readback "$OUT/mission-recommendations/checkpoint-readback-after-node-01.json" >/dev/null
test -s "$OUT/mission-recommendations/recommendation-workgraph-after-node-01.json"
test -s "$OUT/mission-recommendations/recommendation-readback-after-node-01.json"
test -s "$OUT/mission-recommendations/execution-readback-after-node-01.json"
test -s "$OUT/mission-recommendations/checkpoint-readback-after-node-01.json"
jq -e '.completed_nodes == 1 and .ready_nodes == 39 and .checkpoint_count == 1 and .return_gate_status == "blocked_ready_nodes_remain" and .first_executable_node == "mission-recommendation-next-02" and .started_at == "2026-07-04T08:00:00-07:00" and .elapsed_minutes == 17 and .final_response_allowed == false' "$OUT/mission-recommendations/recommendation-readback-after-node-01.json" >/dev/null
jq -e '.completed_recommendation_nodes == 1 and .lease_health_status == "minimum_unmet" and .checkpoint_freshness_status == "fresh_checkpoint_required_after_each_node_or_timed_interval" and .generated_workgraph.ready_nodes == 39 and .generated_workgraph.executable_ready_nodes == 1 and .generated_workgraph.first_executable_node == "mission-recommendation-next-02" and .generated_workgraph.lease_health_status == .lease_health_status and .generated_workgraph.checkpoint_freshness_status == .checkpoint_freshness_status and .generated_workgraph.checkpoint_count == 1 and .generated_workgraph.return_gate_status == "blocked_ready_nodes_remain" and .generated_workgraph.final_response_allowed == false and .foundry_run_link_readiness_summary.completed_run_links == 1 and .foundry_run_link_readiness_summary.required_run_links == 40 and .foundry_run_link_readiness_summary.next_executable_node == "mission-recommendation-next-02" and .foundry_run_link_readiness_summary.lease_health_status == .lease_health_status and .foundry_run_link_readiness_summary.checkpoint_freshness_status == .checkpoint_freshness_status and (.source_artifacts[] | select(.ref == "foundry_run_link_readiness_summary" and (.digest | test("^sha256:[0-9a-f]{64}$"))))' "$OUT/mission-recommendations/execution-readback-after-node-01.json" >/dev/null
jq -e '.schema == "ao.atlas.recommendation-checkpoint-readback.v0.1" and .completed_nodes == 1 and .ready_nodes == 39 and .elapsed_minutes == 17 and .min_minutes_met == false and .lease_health_status == "minimum_unmet" and .final_response_allowed == false' "$OUT/mission-recommendations/checkpoint-readback-after-node-01.json" >/dev/null
"$BIN" mission recommendations resume \
  --wave "$OUT/mission-recommendations/recommendation-wave.json" \
  --workgraph "$OUT/mission-recommendations/recommendation-workgraph-after-node-01.json" \
  --lease-start "$OUT/mission-recommendations/lease-start.json" \
  --completed-at 2026-07-04T08:25:00-07:00 \
  --evidence-root target/production-readiness/mission-recommendations \
  --out-readback "$OUT/mission-recommendations/recommendation-readback-resumed.json" \
  --out-execution-readback "$OUT/mission-recommendations/execution-readback-resumed.json" \
  --out-command-readback "$OUT/mission-recommendations/command-readback-resumed.json" \
  --out-promoter-readback "$OUT/mission-recommendations/promoter-readback-resumed.json" \
  --out-foundry-rollup "$OUT/mission-recommendations/foundry-rollup-resumed.json" \
  --out-reconciliation-packet "$OUT/mission-recommendations/reconciliation-packet-resumed.json" \
  --out-next-prompt "$OUT/mission-recommendations/next-recommended-prompt-resumed.md" >/dev/null
test -s "$OUT/mission-recommendations/recommendation-readback-resumed.json"
test -s "$OUT/mission-recommendations/execution-readback-resumed.json"
test -s "$OUT/mission-recommendations/command-readback-resumed.json"
test -s "$OUT/mission-recommendations/promoter-readback-resumed.json"
test -s "$OUT/mission-recommendations/foundry-rollup-resumed.json"
test -s "$OUT/mission-recommendations/reconciliation-packet-resumed.json"
test -s "$OUT/mission-recommendations/next-recommended-prompt-resumed.md"
jq -e '.started_at == "2026-07-04T08:00:00-07:00" and .elapsed_minutes == 25 and .checkpoint_count == 1 and .return_gate_status == "blocked_ready_nodes_remain" and .final_response_allowed == false' "$OUT/mission-recommendations/recommendation-readback-resumed.json" >/dev/null
jq -e '.exact_next_action_readback.status == "continuation_required" and .exact_next_action_readback.action == .exact_next_action and .exact_next_action_readback.next_executable_node == .first_executable_node and .exact_next_action_readback.return_gate_status == .return_gate_status and .exact_next_action_readback.final_response_allowed == .final_response_allowed and .exact_next_action_readback.source == "recommendation_readback"' "$OUT/mission-recommendations/recommendation-readback-resumed.json" >/dev/null
jq -e '([.command_timeline_placeholders[].slot] | sort) == ["checkpoint","exact_next_action","return_gate"] and all(.command_timeline_placeholders[]; .source == "recommendation_readback" and .status == "pending_command_timeline" and .required_before_final_response == true and (.summary | length) > 0)' "$OUT/mission-recommendations/recommendation-readback-resumed.json" >/dev/null
jq -e '([.promoter_no_promotion_placeholders[].slot] | sort) == ["authority_advance","promotion_claim","rsi_boundary"] and all(.promoter_no_promotion_placeholders[]; .source == "recommendation_readback" and .status == "pending_promoter_no_promotion" and .required_before_final_response == true and (.summary | length) > 0)' "$OUT/mission-recommendations/recommendation-readback-resumed.json" >/dev/null
jq -e '.lease_health_status == "minimum_unmet" and .checkpoint_freshness_status == "fresh_checkpoint_required_after_each_node_or_timed_interval" and .generated_workgraph.lease_health_status == .lease_health_status and .generated_workgraph.checkpoint_freshness_status == .checkpoint_freshness_status and .foundry_run_link_readiness_summary.completed_run_links == 1 and .foundry_run_link_readiness_summary.required_run_links == 40 and .foundry_run_link_readiness_summary.next_executable_node == "mission-recommendation-next-02" and .foundry_run_link_readiness_summary.lease_health_status == .lease_health_status and .foundry_run_link_readiness_summary.checkpoint_freshness_status == .checkpoint_freshness_status and (.source_artifacts[] | select(.ref == "foundry_run_link_readiness_summary" and (.digest | test("^sha256:[0-9a-f]{64}$"))))' "$OUT/mission-recommendations/execution-readback-resumed.json" >/dev/null
jq -e '.schema == "ao.atlas.recommendation-command-readback.v0.1" and .elapsed_minutes == 25 and .lease_time_status == "minimum_minutes_unmet" and .lease_health_status == "minimum_unmet" and .checkpoint_freshness_status == "fresh_checkpoint_required_after_each_node_or_timed_interval" and .checkpoint_count == 1 and .return_gate_status == "blocked_ready_nodes_remain" and .final_response_allowed == false and .claims_authority_advance == false and .command_timeline_binding.summary == .compact_timeline and .command_timeline_binding.exact_next_action == .exact_next_action and .command_timeline_binding.return_gate_status == .return_gate_status and .command_timeline_binding.lease_health_status == .lease_health_status and .command_timeline_binding.checkpoint_freshness_status == .checkpoint_freshness_status' "$OUT/mission-recommendations/command-readback-resumed.json" >/dev/null
jq -e '.schema == "ao.atlas.recommendation-promoter-readback.v0.1" and .lease_health_status == "minimum_unmet" and .checkpoint_freshness_status == "fresh_checkpoint_required_after_each_node_or_timed_interval" and .promotion_claimed == false and .rsi_remains_denied == true and .claims_authority_advance == false' "$OUT/mission-recommendations/promoter-readback-resumed.json" >/dev/null
jq -e '.no_promotion_summary == "No mutation authority promotion claimed; RSI remains denied." and .next_denied_class == "RSI"' "$OUT/mission-recommendations/promoter-readback-resumed.json" >/dev/null
jq -e '.schema == "ao.atlas.recommendation-foundry-rollup.v0.1" and .node_completion_status == "nodes_in_progress" and .lease_completion_status == "minimum_minutes_unmet" and .lease_health_status == "minimum_unmet" and .checkpoint_freshness_status == "fresh_checkpoint_required_after_each_node_or_timed_interval" and .checkpoint_count == 1 and .return_gate_status == "blocked_ready_nodes_remain" and .final_response_allowed == false and .claims_authority_advance == false' "$OUT/mission-recommendations/foundry-rollup-resumed.json" >/dev/null
jq -e '.schema == "ao.atlas.recommendation-reconciliation-packet.v0.1" and .status == "continuation_required" and .checkpoint_count == 1 and .return_gate_status == "blocked_ready_nodes_remain" and .lease_health_status == "minimum_unmet" and .checkpoint_freshness_status == "fresh_checkpoint_required_after_each_node_or_timed_interval" and .stale_route_decision_status == "fresh_atlas_supervises_foundry_owns_one_active_node" and .command_return_gate_status == "blocked_ready_nodes_remain" and .foundry_return_gate_status == "blocked_ready_nodes_remain" and .artifacts_agree == true and .promotion_claimed == false and .rsi_remains_denied == true and .claims_authority_advance == false' "$OUT/mission-recommendations/reconciliation-packet-resumed.json" >/dev/null
grep -q "Next executable node: \`mission-recommendation-next-02\`" "$OUT/mission-recommendations/next-recommended-prompt-resumed.md"
grep -q "Early-return risk: \`blocked_final_response_ready_nodes_remain\`" "$OUT/mission-recommendations/next-recommended-prompt-resumed.md"
grep -q "If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair." "$OUT/mission-recommendations/next-recommended-prompt-resumed.md"
grep -q "If \`ready_nodes > 0\` or \`exact_next_action\` is non-empty, do not produce a final response." "$OUT/mission-recommendations/next-recommended-prompt-resumed.md"
reject_generated_recommendation_prompt_public_safety "$OUT/mission-recommendations/next-recommended-prompt-resumed.md"
assert_schema_required_fields_present schemas/recommendation-readback.schema.json "$OUT/mission-recommendations/recommendation-readback-resumed.json" "recommendation readback"
assert_schema_required_fields_present schemas/recommendation-checkpoint-readback.schema.json "$OUT/mission-recommendations/checkpoint-readback-after-node-01.json" "recommendation checkpoint readback"
assert_schema_required_fields_present schemas/recommendation-command-readback.schema.json "$OUT/mission-recommendations/command-readback-resumed.json" "recommendation command readback"
assert_schema_required_fields_present schemas/recommendation-promoter-readback.schema.json "$OUT/mission-recommendations/promoter-readback-resumed.json" "recommendation promoter readback"
assert_schema_required_fields_present schemas/recommendation-foundry-rollup.schema.json "$OUT/mission-recommendations/foundry-rollup-resumed.json" "recommendation Foundry rollup"
assert_schema_required_fields_present schemas/recommendation-reconciliation-packet.schema.json "$OUT/mission-recommendations/reconciliation-packet-resumed.json" "recommendation reconciliation packet"
pass "recommendation-readback-schema-coverage"
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
reject_local_absolute_paths "Blueprint Foundry continuation" \
  "$OUT/blueprint-import-low-risk-code/foundry-import/foundry-continuation-handoff.json" \
  "$OUT/blueprint-import-low-risk-code/foundry-import/foundry-continuation-prompt.md"
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
reject_local_absolute_paths "Foundry continuation" \
  "$OUT/foundry-import/foundry-continuation-handoff.json" \
  "$OUT/foundry-import/foundry-continuation-prompt.md"
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
if "$BIN" mission recommendations import --recommendations examples/invalid/feature-depth-recommendations-shallow.json --target-instance demo-stack --min-tasks 20 --node-budget 20 --estimated-minutes 90 --out "$OUT/mission-recommendations-shallow" >"$OUT/mission-recommendations-shallow.out" 2>&1; then
  echo "shallow Feature Depth recommendations were accepted" >&2
  exit 1
fi
grep -q "at least 20 tasks" "$OUT/mission-recommendations-shallow.out"
if "$BIN" mission recommendations import --recommendations examples/invalid/feature-depth-recommendations-unsafe.json --target-instance demo-stack --min-tasks 20 --node-budget 20 --estimated-minutes 90 --out "$OUT/mission-recommendations-unsafe" >"$OUT/mission-recommendations-unsafe.out" 2>&1; then
  echo "unsafe Feature Depth recommendations were accepted" >&2
  exit 1
fi
grep -q "safe_to_execute must be false" "$OUT/mission-recommendations-unsafe.out"
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
  recommendation_min_minutes_met="$(jq -r '.min_minutes_met // false' "$recommendation_readback")"
  recommendation_elapsed_minutes="$(jq -r '.elapsed_minutes // 0' "$recommendation_readback")"
  recommendation_min_minutes="$(jq -r '.supervisor.min_minutes // 0' "$recommendation_readback")"
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
  if [ "$recommendation_final_allowed" = "true" ]; then
    if [ "$recommendation_min_minutes_met" != "true" ]; then
      echo "recommendation readback $recommendation_readback allows final response without min_minutes_met=true" >&2
      exit 1
    fi
    if [ "$recommendation_min_minutes" -gt 0 ] && [ "$recommendation_elapsed_minutes" -lt "$recommendation_min_minutes" ]; then
      echo "recommendation readback $recommendation_readback allows final response with elapsed_minutes=$recommendation_elapsed_minutes below supervisor.min_minutes=$recommendation_min_minutes" >&2
      exit 1
    fi
  fi
done < <(find docs/evidence -name execution-readback.json -type f)
pass "recommendation-ledger-consistency"

lease_resume_root="docs/evidence/ao-atlas-lease-resume-wave-v01"
lease_resume_readback="$lease_resume_root/recommendation-readback.json"
lease_resume_synthesis="$lease_resume_root/final-synthesis.json"
lease_resume_prompt="$lease_resume_root/next-recommended-prompt.md"
if [ -f "$lease_resume_readback" ]; then
  test -s "$lease_resume_synthesis"
  test -s "$lease_resume_prompt"
  jq -e --slurpfile synthesis "$lease_resume_synthesis" '
    .total_nodes == $synthesis[0].total_nodes and
    .completed_nodes == $synthesis[0].completed_nodes and
    .ready_nodes == $synthesis[0].ready_nodes and
    .checkpoint_count == $synthesis[0].checkpoint_count and
    .elapsed_minutes == $synthesis[0].elapsed_minutes and
    .return_gate_status == $synthesis[0].return_gate_status and
    .final_response_allowed == $synthesis[0].final_response_allowed and
    .exact_next_action == $synthesis[0].exact_next_action and
    .final_response_allowed == false and
    .return_gate_status == "blocked_ready_nodes_remain" and
    .ready_nodes > 0 and
    (.exact_next_action | length) > 0
  ' "$lease_resume_readback" >/dev/null
  lease_resume_completed_nodes="$(jq -r '.completed_nodes' "$lease_resume_readback")"
  lease_resume_total_nodes="$(jq -r '.total_nodes' "$lease_resume_readback")"
  lease_resume_ready_nodes="$(jq -r '.ready_nodes' "$lease_resume_readback")"
  lease_resume_elapsed_minutes="$(jq -r '.elapsed_minutes' "$lease_resume_readback")"
  lease_resume_checkpoint_count="$(jq -r '.checkpoint_count' "$lease_resume_readback")"
  lease_resume_next_action="$(jq -r '.exact_next_action' "$lease_resume_readback")"
  lease_resume_node_suffix="$(printf "%02d" "$lease_resume_completed_nodes")"
  lease_resume_workgraph="$lease_resume_root/nodes/mission-recommendation-next-$lease_resume_node_suffix/workgraph-after.json"
  test -s "$lease_resume_workgraph"
  grep -qF "Current workgraph: \`$lease_resume_workgraph\`" "$lease_resume_prompt"
  grep -qF "Completed nodes: $lease_resume_completed_nodes / $lease_resume_total_nodes" "$lease_resume_prompt"
  grep -qF "Ready nodes: $lease_resume_ready_nodes" "$lease_resume_prompt"
  grep -qF "Elapsed minutes at latest checkpoint: $lease_resume_elapsed_minutes" "$lease_resume_prompt"
  grep -qF "Checkpoint count: $lease_resume_checkpoint_count" "$lease_resume_prompt"
  grep -qF "Early-return risk: \`blocked_final_response_ready_nodes_remain\`" "$lease_resume_prompt"
  grep -qF "$lease_resume_next_action" "$lease_resume_prompt"
  grep -qF 'If a node becomes blocked or failed, record the exact blocked node id, missing evidence or stop gate, safe repair or repack action, and resume from the latest checkpoint after repair.' "$lease_resume_prompt"
  grep -qF 'If `ready_nodes > 0` or `exact_next_action` is non-empty, do not produce a final response.' "$lease_resume_prompt"
  reject_generated_recommendation_prompt_public_safety "$lease_resume_prompt"
fi
pass "lease-resume-wave-public-safety-readback"

scan_file="$OUT/public-safety-files.txt"
find . \
  -path './.git' -prune -o \
  -path './target' -prune -o \
  -path './.atlas-local' -prune -o \
  -path './.atlas-state' -prune -o \
  -path './atlas' -prune -o \
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
