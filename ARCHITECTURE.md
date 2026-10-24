# Architecture

`marc` is a pure codec between one [ArcNet CORE](https://github.com/net-of-arc/arcnet-spec/blob/main/spec/CORE.md) Markdown node file and a bag of `(subject, predicate, object)` triples. Decisions are recorded in [adrs/](adrs/): [ADR 001](adrs/001-system-architecture.md) (flat library layout), [ADR 002](adrs/002-runtime-dependencies.md) (runtime dependencies), [ADR 003](adrs/003-public-api-shape.md) (public API shape).

## Glossary

| Term | Definition | Go identifier |
|---|---|---|
| Subject / Identity | The `@id` of a node: the name of the thing it describes. Validated against CORE §5 and NFC-normalised. | `ID` |
| Predicate | A camelCase property name; the "verb" of a triple. | `Predicate` |
| Class | The `@type` of a node; an UpperCamelCase name. | `Class` |
| Object | The value of a triple: exactly one of a reference or a literal. A sealed sum type. | `Object` |
| Reference | An object naming another node's identity, with optional presentation-only display text. | `Ref` |
| Text literal | A string literal; also the value of prose and statement predicates. | `Text` |
| Number literal | A numeric literal kept in its exact lexical form. | `Number` |
| Boolean literal | `true` or `false`. | `Bool` |
| Date literal | A civil calendar date. | `Date` |
| Datetime literal | An instant, normalised to UTC. | `DateTime` |
| Triple | One statement `(subject, predicate, object)`. | `Triple` |
| Node / Bag | One Markdown file seen as data: subject, class, a set of triples, merge-state keywords and the observed placement of each predicate. | `Node` |
| Attribution | One `@provenance` entry: a document contributed a value of a predicate. | `Attribution` |
| Role | Where a predicate renders: `meta`, `text`, `edge` or `link`. | `Role` |
| Property | The rendering rule for one predicate: role, label, default-text flag and list flag. | `Property` |
| Schema | The application-owned vocabulary that answers rendering rules per class. `Properties` is the class-agnostic reference implementation. | `Schema`, `Properties` |
| Codec | A configured encoder/decoder (schema and input limit). | `Codec`, `Config` |
| Canonical form | The exact bytes the encoder produces; decoding then encoding a canonical file reproduces it byte for byte. | — |
| Observed property | The `Property` implied by the position a predicate was found in while decoding; the encoder falls back to it when the schema is silent. | `Node.Property` |
| Natural order | The order triples entered a node — document order on decode, call order to `NewNode` otherwise — kept within each CORE §4 section on encode; replaces a lexicographic canonical order for triples (BUG-001). | `Node.Triples`, `Node.Predicates`, `Node.Values` |
| Paragraph | The unit of a text value; one paragraph, one `Text`, split at blank lines outside fenced code (BUG-001). | `Node.Values` on a `RoleText` predicate |
| Untyped link | A bare `[[Target]]` inside a text value; part of that value, never a triple; exposed as a derived, read-only view (BUG-001). | `Node.Links` |

## Packages

| Package | Purpose |
|---|---|
| `github.com/net-of-arc/marc` | The whole library: names, objects, nodes, schema, codec. No sub-packages, no `internal/`. |

## Non-functional requirements

- **Go versions**: `go 1.26` minimum; the two latest Go releases (1.26, 1.27) are tested in CI.
- **Concurrency**: everything is synchronous; the library starts no goroutines. Every name, object and `Node` is immutable and safe for concurrent use. A `Codec` is immutable and safe for concurrent use when the consumer's `Schema` is, because it keeps the schema by reference.
- **Resources**: decoding reads at most `Config.MaxSize` bytes (default 4 MiB = 4 194 304 bytes). YAML anchors and aliases are rejected; front-matter nesting is bounded to the node model.
- **Side effects**: none. No filesystem, network, environment, logging or mutable package state.

### Hot paths

Benchmarked by `BenchmarkDecode` and `BenchmarkEncode` in bench_test.go over `testdata/bench-100.md` (9.4 KB, 100 statements). Baseline on Apple M2 Pro, Go 1.27, 2026-10-24 (BUG-001: fixture's default prose merged into one paragraph to keep the statement count at 100 under natural order):

| Path | Budget | ns/op | B/op | allocs/op |
|---|---|---|---|---|
| `Decode` | < 1 ms/op | 73 910 | 114 353 | 912 |
| `Encode` | < 1 ms/op | 16 921 | 65 040 | 210 |
