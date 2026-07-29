# Canonical Terminal Index

`atlas terminal-index` reconciles immutable mission evidence through an explicit,
digest-bound lineage manifest. It never edits source evidence.

The input contract is `ao.canonical-terminal-index-input.v1`. It requires a mission
identity, a deterministic RFC 3339 generation timestamp, `evidence_root="."`, and
an ordered list of artifacts with unique roles, relative paths, and SHA-256
digests. The producer accepts at most 128 regular files, 1 MiB per file, and
16 MiB in total. It rejects duplicate JSON keys, malformed JSON, symlinks,
absolute paths, traversal, digest changes, identity drift, and non-monotonic
lineage.

The output contract is `ao.canonical-terminal-index.v1`. It records every source
artifact and digest, the canonical terminal reference, node counts, lease
minimum/target/maximum/elapsed values, completion observation, readiness,
return-gate state, conflict codes, exact next action, and explicit safety
boundaries. Its digest covers the complete index except the digest field itself.

Completion and readiness are separate. A completed node wave that exceeds its
maximum lease is recorded as completed but not ready. Ready, blocked, or failed
work; a remaining exact action; stale duration state; unsafe flags; or a lease
outside its allowed window all force `readiness_passed=false` and
`final_response_allowed=false`.

```sh
go run ./cmd/atlas terminal-index build \
  --root docs/evidence/example \
  --manifest docs/evidence/example/canonical-terminal-index-input.json \
  --out docs/evidence/example/canonical-terminal-index.json

go run ./cmd/atlas terminal-index verify \
  --root docs/evidence/example \
  --index docs/evidence/example/canonical-terminal-index.json
```
