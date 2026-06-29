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
  schemas/stack-instance.schema.json
  schemas/atlas-registry.schema.json
  schemas/instance-doctor.schema.json
  schemas/intake.schema.json
  schemas/mission-status.schema.json
  schemas/blueprint-request.schema.json
  schemas/workgraph.schema.json
  schemas/workgraph-repair-plan.schema.json
  schemas/factory-task.schema.json
  schemas/factory-materialization.schema.json
  schemas/context-pack.schema.json
  schemas/foundry-handoff.schema.json
  schemas/foundry-import.schema.json
  schemas/run-link.schema.json
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
"$BIN" blueprint-request validate --request examples/valid/blueprint-request.json >/dev/null
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
pass "invalid-fixtures-rejected"

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
