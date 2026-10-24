# Implementation Plan: Markdown ⇄ Triples Codec Core Types

**Branch**: `001-markdown-triples-codec` | **Date**: 2026-10-23 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-markdown-triples-codec/spec.md`

**Bugfix**: 2026-10-24 — BUG-001 Updated from bugfix patch (natural order, paragraph-per-value text, `**Label**:` decoding, derived untyped-link view)

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.
It is the project override of the core plan template and enforces `.specify/memory/constitution.md`.

## Summary

Deliver the root package `marc` as a pure, side-effect-free codec between one ArcNet CORE Markdown node file and a **bag of triples**:

- **Codec functions.** `Decode(io.Reader)`, `Unmarshal([]byte)`, `Encode(io.Writer, Node)` and `Marshal(Node)` are pure functions, also available on a configured `*Codec` from `New(Config)`.
- **Core types.** These are the public contract:
  - `ID`, `Predicate` and `Class` are validated opaque name types. `ID` is its own type, not a generic one.
  - `Object` is a **sealed-interface sum type** with the variants `Ref`, `Text`, `Number`, `Bool`, `Date` and `DateTime`, plus the generic `Select[T]` for typed filtering (research R1).
  - `Triple` is a plain struct.
  - `Node` is immutable and is iterated with `iter.Seq[Triple]`, `iter.Seq[Object]` and `iter.Seq[Predicate]`.
  - *(BUG-001)* `Node` keeps triples in **natural order** (document order on decode, insertion order for `NewNode`), deduplicated with the first occurrence winning; the encoder writes that order within each CORE §4 section. Canonical order (`compareObjects`) remains only for `@conflicts`/`@provenance`. Untyped links in text values are exposed by `Node.Links() iter.Seq2[Predicate, Ref]`, a derived view, never triples (FR-026).
  - Placement is driven by a **consumer-owned `Schema` interface** (`Property(Class, Predicate)` and `Predicate(Class, label)`) that returns `Property` and `Role` values; `Properties` is the reference map implementation (research R4).
- **Parsing.**
  - Front-matter is read with `go.yaml.in/yaml/v3` and written by hand to canonical rules ([contracts/format.md](contracts/format.md)).
  - The body uses a line scanner. *(BUG-001)* It splits prose into one `Text` per paragraph (blank lines outside fences) and reads a `**Label**:` paragraph lead as a label.
  - NFC normalisation uses `golang.org/x/text`.
- **CI/CD.** GitHub Actions adapted from fogfish/arcnet-cli (`build`, `check-code`, `check-test`) and extended to the constitution's mandatory tooling (research R11).

## Technical Context

**Language/Version**: Go 1.26 minimum (`go` directive lowered from `1.27.0` to `1.26.0`, see research R11); CI matrix `1.26.6` and the latest pinned `1.27.x` (Principle XIII)

**Module**: single Go module `github.com/net-of-arc/marc`, root package `marc` (fixed, Principles III, XIII)

**Primary Dependencies**: standard library; `github.com/fogfish/faults` (fixed); `go.yaml.in/yaml/v3` and `golang.org/x/text` → [ADR 002](../../adrs/002-runtime-dependencies.md) (Principle VII). Test-only: `github.com/fogfish/it/v2`.

**Testing**: `go test -race ./...` with `github.com/fogfish/it/v2` (fixed); `FuzzDecode` and `FuzzRoundTrip` with a seed corpus in `testdata/fuzz/`; `BenchmarkDecode` and `BenchmarkEncode` with `-benchmem`; a property-style generated round-trip test (≥ 1,000 nodes) (Principle VI)

**External Interfaces**: `io.Reader`, `io.Writer`, `[]byte`, `iter.Seq`, `iter.Seq2`, `time.Time`; no filesystem, network or context (codec is pure; [ADR 003](../../adrs/003-public-api-shape.md) §3) (Principle VII)

**Public API Surface**: as listed in full in [contracts/api.md](contracts/api.md) (Principle IX):
- **Names**: `ID`, `Predicate`, `Class` with `New*`/`Must*`.
- **Object**: `Object` (sealed); `Ref` + `RefTo`; `Text`; `Number` + `ParseNumber`; `Bool`; `Date` + `DateOf`; `DateTime` + `DateTimeOf`; `Select[T]`.
- **Bag**: `Triple`, `Attribution`, `Node` + `NewNode`; `Node.Links` (BUG-001, FR-026).
- **Placement**: `Role` (`RoleMeta`/`RoleText`/`RoleEdge`/`RoleLink`), `Property` (value struct), `Schema` (consumer-implemented interface), `Properties` (reference implementation with `Validate`).
- **Codec**: `Config{Schema; MaxSize}` (zero means no schema and a 4 MiB limit); `Codec` + `New`; package-level `Decode`/`Unmarshal`/`Encode`/`Marshal`.
- **Errors**: 10 `faults.Type` constants (`ErrMalformed`, `ErrIdentityMismatch`, `ErrInvalidID`, `ErrInvalidName`, `ErrInvalidObject`, `ErrForeignSubject`, `ErrTooLarge`, `ErrProseStructure`, `ErrInvalidProperty`, `ErrInvalidConfig`).

**Concurrency & Resources**: fully synchronous, no goroutines. `Node`, all names and all objects are immutable and safe for concurrent use. `Codec` is immutable and safe for concurrent use when the consumer's `Schema` is (the codec keeps it by reference and never copies it). Input is bounded by `Config.MaxSize` (read through `io.LimitReader(max+1)`, and `len` is checked before `Unmarshal` parses). YAML anchors and aliases are rejected, and nesting is bounded by the node model (Principle XI)

**Performance Goals**: hot paths `Decode` and `Encode`, each under 1 ms for a 10 KB / 100-statement node (SC-006). Named in ARCHITECTURE.md with the benchmark baseline (Principles VI, XI)

**Compatibility**: initial API on v0, MINOR (first tag `v0.1.0`); `apidiff` is skipped because no previous tag exists. Lowering the `go` directive happens before any release, so it is not a compatibility event (Principle XIV)

**Graduation Check**: none of the ADR 001 criteria are met. There is one process (a codec), no interchangeable technologies, no SDK, one package and no composition step (Principle III)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Read every ADR referenced below before planning; an accepted ADR overrides this plan (Principle I).

**Referenced ADRs**: [ADR 001 - Principles of System Architecture (Library Edition)](../../adrs/001-system-architecture.md), [ADR 002 - Runtime Dependencies](../../adrs/002-runtime-dependencies.md) (Proposed), [ADR 003 - Public API Shape](../../adrs/003-public-api-shape.md) (Proposed)

| # | Principle | Gate | Status |
|---|-----------|------|--------|
| I | Architecture Documentation & Decision Records | ARCHITECTURE.md (currently a 2-line stub) gets a glossary, package list and NFRs in this feature; ADR 002 (dependencies) and ADR 003 (API shape) are drafted and must be **Accepted before implementation**; ADR 001 is respected | PASS (conditional on accepting ADR 002/003) |
| II | Domain-Driven Design & Glossary | Glossary terms with their Go identifiers are listed in [data-model.md](data-model.md); one name per concept (`Property` mirrors CORE `_schema/Property`; `Ref` is used rather than "link", to avoid clashing with the `link` role); illegal states unrepresentable through opaque validated types and a sealed `Object` | PASS |
| III | Flat Package Architecture | Root package `marc` only, files by concept (see Project Structure), no sub-packages, no `internal/` needed initially | PASS |
| IV | Functional Programming Style | Immutable values; the decoder is a composition of small pure steps (split front-matter → parse YAML → scan body → assemble node); the encoder renders sections independently and concatenates them; ≤ 25-line functions enforced in review; no inline comments | PASS |
| V | Code Quality & Simplicity | Every exported identifier is traced to a spec story ([contracts/api.md](contracts/api.md) Traceability). Deferred (YAGNI): constructing nodes with `@conflicts`/`@provenance`, graph-level identity folding | PASS |
| VI | TDD & Automated Testing | Tests first; `-race`; fuzz tests `FuzzDecode` and `FuzzRoundTrip`; benchmarks on the declared hot paths; `fogfish/it/v2` only | PASS |
| VII | Dependencies & External Interfaces | Two new runtime dependencies, each justified in ADR 002; no technology SDK; no `os` usage; effects arrive only through `io.Reader`/`io.Writer` | PASS (conditional on ADR 002) |
| VIII | Acceptance Testing & Spec Traceability | 6 stories / 28 scenarios → black-box tests in `package marc_test` (file map in [quickstart.md](quickstart.md) §2); `ExampleDecode`, `ExampleEncode` and `ExampleSelect` | PASS |
| IX | Public API Design | `New(Config)` with zero-value defaults and validation of `MaxSize` (`ErrInvalidConfig`); `Schema` is exported as an interface because consumers implement it, and each `Property` it returns is validated at use (`ErrInvalidProperty`); no functional options; `Must*` only alongside `New*`; no panics across the API; ownership and concurrency stated in GoDoc. **No `context.Context`** on codec functions: justified as pure in-memory transformation in ADR 003 §3 | PASS (conditional on ADR 003) |
| X | Library Citizenship | No `os`, no stdout/stderr, no env access, no logging; the package-level default `Codec` is immutable after init; `init()` does no I/O; CGO-free build verified in CI | PASS |
| XI | Concurrency, Resources & Performance | Synchronous; no goroutines; no `io.Closer` needed; untrusted input bounded by `MaxSize`; anchors and aliases rejected | PASS |
| XII | Documentation & Error Contract | `doc.go` with a quick start; GoDoc on every identifier; examples; README rewritten; errors as `faults.Type` constants, wrapped with `.With`; messages never include payload values | PASS |
| XIII | Modules & Release Engineering | Single module; `go` directive lowered to 1.26 to satisfy the "two most recent releases" rule (pre-release, so it needs no MINOR note); CI-driven tags from a commit-message marker; release notes exclude `^docs:`/`^test:`/`^ci:` | PASS |
| XIV | Versioning, Compatibility & Security | First release is v0.1.0; `apidiff` is wired in `check-code` and activates once a tag exists; `govulncheck` runs before tagging; the decoder validates sizes before allocating and is fuzzed. ADR 003 notes that new `Object` variants are effectively breaking | PASS |

No FAIL. The conditional PASSes depend on the maintainer moving ADR 002 and ADR 003 from **Proposed** to **Accepted** (Phase 2e of tasks).

### Post-design re-check (after Phase 1)

The design artifacts ([data-model.md](data-model.md), [contracts/api.md](contracts/api.md), [contracts/format.md](contracts/format.md)) were re-checked against the table above:
- **Principle II**: the `Link` constructor was renamed to `RefTo` to avoid clashing with the CORE `link` role.
- **Principle IX**: `Config` was cut to two fields by folding "default text predicate" into `Property.Default`, which removes a redundant setting.
- **Principles V, IX (review feedback)**: `Schema` changed from a library-owned `map[Predicate]Property` to a consumer-implemented, class-keyed interface, because applications own their vocabulary. `Property` stays a value struct, and `Properties` is the reference implementation. As a side effect, the "one `Codec` per class" limitation disappears. Whole-schema validation can no longer run in `New`; it moves to per-`Property` checks at use and to `Properties.Validate`. This is recorded in ADR 003 §4.

All gates remain PASS (conditional on the ADRs).

### Bugfix re-check (BUG-001, 2026-10-24)

- **Principle V/IX**: one exported method is added, `Node.Links`, traced to FR-026 / Story 4 in [contracts/api.md](contracts/api.md).
- **Principle XIV**: iteration and encoding order change from canonical to natural, and multiple text values now render as paragraphs instead of `## Label` bullets. Pre-release (no tag), so this is not a compatibility event; it is recorded in the GoDoc of `Node` and `Encode`.
- **Principle XI**: the ordered builder keeps the dedup map plus one slice in insertion order; no new allocation class on the hot path.

All gates remain PASS.

## Project Structure

### Documentation (this feature)

```text
specs/001-markdown-triples-codec/
├── spec.md
├── plan.md              # this file
├── research.md          # Phase 0: R1 sum type … R11 CI/CD
├── data-model.md        # Phase 1: entities, invariants, errors
├── quickstart.md        # Phase 1: validation guide
├── contracts/
│   ├── api.md           # exported Go API
│   └── format.md        # canonical rendering & decoding grammar
├── checklists/requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
./                               # module github.com/net-of-arc/marc, package marc
├── doc.go                       # package comment, quick start
├── id.go / id_test.go           # ID: validation (CORE §5), NFC
├── name.go / name_test.go       # Predicate, Class, derived label
├── object.go / object_test.go   # Object sum type, variants, equality keys, canonical order, Select
├── triple.go                    # Triple, Attribution
├── node.go / node_test.go       # Node, NewNode, iterators, set semantics
├── schema.go / schema_test.go   # Role, Property, Schema, validation
├── config.go / config_test.go   # Config, defaults, validation
├── codec.go                     # Codec, New, package-level Decode/Unmarshal/Encode/Marshal
├── frontmatter.go / _test.go    # split + YAML read (yaml.Node) + canonical YAML write
├── body.go / body_test.go       # line scanner: H1, prose, labels, bullets, blocks, fences, inline form
├── decode.go                    # assembly: front-matter + body → Node (observed properties)
├── encode.go                    # Node + schema → canonical bytes
├── errors.go                    # exported faults.Type constants
├── decode_test.go               # Story 1 (package marc_test)
├── encode_test.go               # Story 2
├── roundtrip_test.go            # Story 3 + generated round trip (SC-002)
├── prose_test.go                # Story 4
├── malformed_test.go            # Story 5
├── mergestate_test.go           # Story 6
├── example_test.go              # ExampleDecode, ExampleEncode, ExampleSelect
├── fuzz_test.go                 # FuzzDecode, FuzzRoundTrip
├── bench_test.go                # BenchmarkDecode, BenchmarkEncode
├── testdata/                    # CORE fixtures, canonical/, forms/, malformed/, fuzz/
├── .goreleaser.yml              # library mode: builds skipped, changelog filters
├── .github/workflows/
│   ├── build.yml                # push to main: race tests, coverage, govulncheck, tag, release
│   ├── check-code.yml           # PR: gofmt, vet, staticcheck, mod tidy, apidiff
│   └── check-test.yml           # PR: matrix 1.26/1.27, race, coverage, CGO-free build
├── ARCHITECTURE.md              # updated: glossary, package list, NFRs, hot paths
├── README.md                    # updated: purpose, install, quick start, Go versions, pkg.go.dev
└── go.mod                       # go 1.26.0; faults, yaml, x/text; it/v2 (test)
```

**Structure Decision**: This conforms to ADR 001: one flat root package, files organised by concept (identity, names, object, node, schema, front-matter, body, codec), and no sub-packages or `internal/`. Splitting `frontmatter.go` and `body.go` from `decode.go`/`encode.go` keeps every function under 25 lines and lets the two grammars be unit-tested separately. A `marctest` helper package is not needed: building nodes with `NewNode` is enough for consumer tests.

## Complexity Tracking

| Violation | Principle | Why Needed | Simpler Alternative Rejected Because | ADR |
|-----------|-----------|------------|-------------------------------------|-----|
| Two third-party runtime dependencies | VII | YAML front-matter parsing and Unicode NFC are required by CORE and absent from the standard library | A hand-written YAML subset parser is a large attack surface that diverges from other graph tools; skipping NFC violates CORE §5 | [ADR 002](../../adrs/002-runtime-dependencies.md) |
| Natural (insertion) order replaces lexicographic canonical order for triples (BUG-001) | XIV (determinism) | The client's document order is meaningful to readers; CORE §3 requires only a stable order | Lexicographic order reshuffles paragraphs and bullets, losing the order the author wrote | — |
| Codec functions take no `context.Context` | IX (interpretation) | Pure transformation over caller-owned, size-bounded input; nothing to cancel | Adding `ctx` diverges from `encoding/json` idiom without adding behaviour | [ADR 003](../../adrs/003-public-api-shape.md) |
