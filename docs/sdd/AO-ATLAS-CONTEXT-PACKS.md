# AO Atlas Context Packs

Context packs are bounded packets for one factory task. They contain source
references, digests, summaries, assumptions, exclusions, missing-context reason
when repacked, and the missing-context protocol.

Validation fails when:

- the JSON exceeds its configured byte budget;
- a source digest is not `sha256:<64 lowercase hex>`;
- source references contain private or absolute local paths;
- summaries or context metadata contain machine-local path markers.

Context packs should summarize relevant material. They must not embed a whole
mission history or entire repo dump.

CLI:

```sh
atlas context-pack validate --pack <path>
atlas context-pack repack --task <factory-task> --run-link <run-link> --source-ref <ref> --source-digest <sha256> --out <path>
```

`context-pack repack` emits a new bounded context pack only when a blocked or
failed run link includes `needs_context` evidence. It does not copy source
repositories, widen scope by itself, schedule work, execute work, or call
providers. Repack records `missing_context_reason` so the next factory run can
see why the context changed without reading the whole prior mission. The public
`examples/valid/context-pack-needs-context-repack-demo.json` fixture shows the
needs-context repack shape.
