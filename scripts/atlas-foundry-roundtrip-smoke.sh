#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

OUT="target/atlas-foundry-roundtrip"
FOUNDRY_ROOT="${AO_FOUNDRY_ROOT:-$ROOT/../ao-foundry}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out)
      OUT="${2:?missing --out value}"
      shift 2
      ;;
    --foundry-root)
      FOUNDRY_ROOT="${2:?missing --foundry-root value}"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 2
      ;;
  esac
done

case "$OUT" in
  /*|[A-Za-z]:*)
    echo "--out must be a relative public-safe path" >&2
    exit 1
    ;;
esac

if [[ ! -d "$FOUNDRY_ROOT" ]]; then
  echo "AO Foundry repo not found: $FOUNDRY_ROOT" >&2
  exit 1
fi

mkdir -p "$OUT"
ATLAS_BIN="$OUT/atlas"
HANDOFF="$OUT/atlas-foundry-handoff.json"
IMPORT_DIR="$OUT/atlas-foundry-import"
IMPORT_PACKET="$IMPORT_DIR/foundry-import.json"
FOUNDRY_VALIDATE="$OUT/foundry-registry-validate.txt"
FOUNDRY_IMPORT_VALIDATE="$OUT/foundry-import-validate.txt"
RUN_LINK="$OUT/run-link.json"
SUMMARY="$OUT/summary.json"

go build -o "$ATLAS_BIN" ./cmd/atlas

"$ATLAS_BIN" foundry handoff emit \
  --workgraph examples/valid/workgraph.json \
  --out "$HANDOFF" > "$OUT/atlas-handoff-emit.txt"

"$ATLAS_BIN" foundry import \
  --workgraph examples/valid/workgraph.json \
  --out "$IMPORT_DIR" > "$OUT/atlas-foundry-import.txt"

(
  cd "$FOUNDRY_ROOT"
  go run ./cmd/foundry registry validate \
    --registry examples/registry/atlas-demo.foundry-registry.json
) > "$FOUNDRY_VALIDATE"

(
  cd "$FOUNDRY_ROOT"
  go run ./cmd/foundry atlas import validate \
    --import "$ROOT/$IMPORT_PACKET"
) > "$FOUNDRY_IMPORT_VALIDATE"

"$ATLAS_BIN" run-link attach \
  --task-id atlas-readiness-task \
  --status completed \
  --evidence atlas="$HANDOFF" \
  --evidence atlas_import="$IMPORT_PACKET" \
  --evidence foundry="$FOUNDRY_VALIDATE" \
  --evidence foundry_import_validation="$FOUNDRY_IMPORT_VALIDATE" \
  --out "$RUN_LINK" > "$OUT/run-link-attach.txt"

"$ATLAS_BIN" run-link validate --run-link "$RUN_LINK" > "$OUT/run-link-validate.txt"

cat > "$SUMMARY" <<JSON
{
  "schema_version": "ao.atlas.foundry-roundtrip-smoke.v0.1",
  "status": "ready",
  "mode": "fixture_only_readback",
  "atlas_handoff": "$HANDOFF",
  "atlas_foundry_import": "$IMPORT_PACKET",
  "foundry_validation": "$FOUNDRY_VALIDATE",
  "foundry_import_validation": "$FOUNDRY_IMPORT_VALIDATE",
  "run_link": "$RUN_LINK",
  "schedules_work": false,
  "executes_work": false,
  "approves_work": false
}
JSON

echo "status=ready"
echo "mode=fixture_only_readback"
echo "summary=$SUMMARY"
