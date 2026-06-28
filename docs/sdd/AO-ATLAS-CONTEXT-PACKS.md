# AO Atlas Context Packs

Context packs are bounded packets for one factory task. They contain source
references, digests, summaries, assumptions, exclusions, and the missing-context
protocol.

Validation fails when:

- the JSON exceeds its configured byte budget;
- a source digest is not `sha256:<64 lowercase hex>`;
- source references contain private or absolute local paths;
- summaries or context metadata contain machine-local path markers.

Context packs should summarize relevant material. They must not embed a whole
mission history or entire repo dump.

