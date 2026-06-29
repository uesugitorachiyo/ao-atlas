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
  schemas/intake.schema.json
  schemas/blueprint-request.schema.json
  schemas/workgraph.schema.json
  schemas/factory-task.schema.json
  schemas/factory-materialization.schema.json
  schemas/context-pack.schema.json
  schemas/foundry-handoff.schema.json
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
"$BIN" intake validate --intake examples/valid/intake.json >/dev/null
"$BIN" blueprint-request validate --request examples/valid/blueprint-request.json >/dev/null
"$BIN" factory-task validate --task examples/valid/factory-task.json >/dev/null
"$BIN" factory materialize --task examples/valid/factory-task.json --out "$OUT/factory-materialization" --dry-run >/dev/null
test -s "$OUT/factory-materialization/materialization.json"
"$BIN" workgraph validate --workgraph examples/valid/workgraph.json >/dev/null
"$BIN" workgraph next --workgraph examples/valid/workgraph.json --json >/dev/null
"$BIN" workgraph materialize-next --workgraph examples/valid/workgraph.json --out "$OUT/workgraph-next-materialization" --dry-run >/dev/null
test -s "$OUT/workgraph-next-materialization/materialization.json"
"$BIN" workgraph status --workgraph examples/valid/workgraph.json >/dev/null
"$BIN" context-pack validate --pack examples/valid/context-pack.json >/dev/null
"$BIN" foundry handoff emit --workgraph examples/valid/workgraph.json --out "$OUT/foundry-handoff.json" >/dev/null
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
if "$BIN" workgraph validate --workgraph examples/invalid/workgraph-missing-dependency.json >/dev/null 2>&1; then
  echo "invalid workgraph was accepted" >&2
  exit 1
fi
if "$BIN" blueprint-request validate --request examples/invalid/blueprint-request-ready-status.json >/dev/null 2>&1; then
  echo "invalid blueprint request was accepted" >&2
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
