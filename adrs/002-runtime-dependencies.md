# ADR 002 - Runtime Dependencies: YAML Front-Matter and Unicode Normalization

**Status**: Accepted
**Date**: 2026-10-23

## Context

`marc` decodes ArcNet CORE nodes (see [spec 001](../specs/001-markdown-triples-codec/spec.md)). Every node starts with YAML front-matter, and CORE §5/§8.5 require Unicode NFC normalization for identities and value equality. The Go standard library provides neither a YAML parser nor a Unicode normalizer. Principle VII requires an ADR for every new direct dependency.

## Decision

1. **`go.yaml.in/yaml/v3`** — used for *reading* front-matter only, through its `yaml.Node` tree (resolved scalar tags, written forms, key order). Front-matter is *written* by `marc` itself under the canonical rules of [contracts/format.md](../specs/001-markdown-triples-codec/contracts/format.md), so the output never depends on the encoder's style choices.
   - License: MIT and Apache-2.0. Maintained by the YAML organization as the continuation of the archived `gopkg.in/yaml.v3`. No transitive dependencies.
   - Hardening: input is size-limited before parsing; anchors and aliases are rejected; nesting is bounded by the node model (scalars and sequences, maps only under `@conflicts`/`@provenance`).
2. **`golang.org/x/text`** — package `unicode/norm` only, for NFC.
   - License: BSD-3-Clause. Maintained by the Go team. No non-`golang.org/x` transitive dependencies.

No other runtime dependency is introduced. The Markdown body is parsed by a hand-written line scanner (CORE §4 is a line grammar, and a CommonMark AST is lossy for byte-identical round trips).

## Consequences

- Consumers inherit two well-maintained, permissively licensed modules. `govulncheck` covers both in CI.
- A future switch to `go.yaml.in/yaml/v4` is internal: no YAML type appears in the exported API (Principle IX).
- The canonical front-matter writer must be kept consistent with the YAML parser's resolution rules. Fuzzing the round trip (Principle VI) guards this.

## Neglected

- `gopkg.in/yaml.v3`: archived, no longer maintained.
- `github.com/goccy/go-yaml`: larger surface, more transitive code.
- A hand-written YAML subset parser: reimplements quoting and scalar resolution, a large attack surface, and divergent from every other tool reading the graph.
- `github.com/yuin/goldmark` for the body: loses source bytes, so it is not needed for a line grammar.
