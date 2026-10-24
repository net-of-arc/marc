# Research: Markdown ⇄ Triples Codec Core Types

**Feature**: [spec.md](spec.md) · **Plan**: [plan.md](plan.md) · **Date**: 2026-10-23

Each section resolves one unknown from the plan's Technical Context. Format: Decision, Rationale, Alternatives considered.

---

## R1. Type-safe sum type for the triple's object

The object of a triple is either a reference to another node or a literal of one of five kinds (text, number, boolean, date, datetime) (FR-004). Clients must be able to branch on the variant at run time and get a typed value back, with illegal states (e.g. "a reference that is also a number", "a reference to an invalid identity") unrepresentable.

Go has no native sum types. The designs below were evaluated against five criteria: (a) the set of variants is closed to the outside world, (b) each variant carries only its own payload, (c) the client handles values with ordinary Go (`switch`, generics), (d) zero/garbage values cannot slip through, (e) cost on the decode hot path.

| Design | (a) closed | (b) own payload | (c) idiom | (d) no illegal state | (e) cost |
|---|---|---|---|---|---|
| **1. Sealed interface** — `interface { String() string; key() string }` with an unexported method; one concrete type per variant; type switch | ✅ unexported method blocks outside implementations | ✅ | ✅ `switch v := o.(type)` | ✅ variant constructors validate | one interface box per value |
| 2. Tagged struct — `struct { Kind; Str string; Num float64; Ref ID; ... }` | ❌ any field combination constructible | ❌ union of all fields | ⚠️ `switch o.Kind` + read the "right" field by convention | ❌ `Kind=Ref` with empty `Ref` is legal | no boxing |
| 3. Visitor / `Match(cases)` — interface with `Accept(Visitor)` or a `Cases{OnRef, OnText...}` struct of funcs | ✅ | ✅ | ⚠️ unidiomatic; `Cases` literals may omit fields, so exhaustiveness is still not compile-checked | ✅ | closure call per value |
| 4. Generic `Object[T]` | — | ✅ | ❌ a node holds objects of mixed kinds; `[]Object[any]` loses the type | — | — |
| 5. `any` + documented set of types | ❌ | ✅ | ⚠️ | ❌ any value accepted | boxing |

Exhaustiveness: no design gives compiler-checked exhaustiveness in Go. Linters (`BurntSushi/go-sumtype`, `nishanths/exhaustive`) can approximate it, but go-sumtype requires a `//sumtype:decl` marker comment and the constitution forbids non-GoDoc comments (Principle IV). Clients can still use these linters in their own code; the library guarantees the closed set via sealing, and documents it.

**Decision**: **Design 1, sealed interface**, with these variants: `Ref`, `Text`, `Number`, `Bool`, `Date`, `DateTime`. Also provide one generic helper, `Select[T Object](iter.Seq[Object]) iter.Seq[T]`, which filters a value sequence down to one variant with static typing (`for r := range marc.Select[marc.Ref](n.Values(replaces))`).

- The sealing method is `key() string`: it carries real behaviour (the CORE §8.5 equality key, prefixed with a variant tag), so it is not just an empty marker.
- `Text` and `Bool` are named basic types (`type Text string`, `type Bool bool`): every value of them is legal, so there is no need for an opaque struct.
- `Number`, `Date`, `DateTime` and `Ref` have invariants (a valid lexical number, a calendar date, a UTC instant, a valid target identity). They hold unexported fields and have validating constructors, so their zero value is the only possible illegal state. `NewNode` rejects zero values, so they can never get into a node.

**Rationale**: This is the only design that satisfies (a), (b) and (d) together, and it is the idiom used by the standard library for closed variants (`go/ast.Expr`, `go/constant.Value`, `encoding/xml.Token`). The boxing cost is one small allocation per value. For a ≤ 100-statement node it is well inside the SC-006 budget, and the benchmark will confirm it.

**Alternatives considered**: Designs 2–5 above. Design 3 was rejected even as a complement: `Select` plus a type switch covers the same need with less API.

---

## R2. Node identity type

**Decision**: `type ID struct{ value string }`, built through `NewID(string) (ID, error)` (validates CORE §5 and applies NFC normalisation) and `MustID(string) ID` (for constants and tests, per Principle IX). Methods: `String()`, `IsZero()`. The type is comparable and usable as a map key. The same pattern is used for `Predicate` (camelCase) and `Class` (UpperCamelCase).

**Rationale**: The user asked for a dedicated identity type. A named string (`type ID string`) would let `marc.ID("a/b")` bypass validation, which violates Principle II ("make illegal states unrepresentable"). A struct with an unexported field means every non-zero `ID` has passed validation.

**Alternatives considered**: `type ID string` (rejected: unvalidated conversion); a generic `Name[K kind]` shared by ID, Predicate and Class (rejected: the three have different validation rules, and a generic identity type is exactly what the user asked to avoid).

CORE §5 case-insensitive uniqueness is a graph-level concern and out of scope (spec Assumptions). `ID` keeps the written case and compares exactly after NFC.

---

## R3. Codec API shape: pure functions over `io.Reader` and `[]byte`

**Decision**:

- Package-level pure functions with default configuration: `Decode(io.Reader) (Node, error)`, `Unmarshal([]byte) (Node, error)`, `Encode(io.Writer, Node) error`, `Marshal(Node) ([]byte, error)`.
- A configured `*Codec` from `New(Config) (*Codec, error)` with the same four methods.
- The package-level functions delegate to an immutable default `Codec` built at package initialisation. That is a value immutable after init, which Principle X allows.

**No `context.Context`**. Principle IX asks for `ctx` on exported functions that "perform I/O, block, or may be canceled". The codec does no I/O of its own. It reads from a reader the caller owns, limited to `Config.MaxSize`, and runs CPU-bound work bounded by that size. Cancellation belongs to the caller's reader (closing it ends the read), the same contract as `encoding/json.Decoder` and `encoding/xml`. Because this interprets Principle IX, it is recorded in [ADR 003](../../adrs/003-public-api-shape.md).

**Rationale**: This matches the user's request ("pure function that operates with `io.Reader` and `[]byte`") and the encoding/json naming that Go developers already know (Marshal/Unmarshal for bytes, Encode/Decode for streams).

**Alternatives considered**: Functional options (forbidden by Principle IX); passing `Config` to every package-level function (rejected: it re-validates the config on every call); exposing only `Codec` (rejected: the common case would need a constructor call).

---

## R4. Where predicate roles come from

The application owns its vocabulary. It may load it from `_schema/Property` nodes, a database or generated code, and the vocabulary can grow at run time (CORE §4: "a vocabulary that grows over time"). The library must therefore consume a schema, not define one.

**Decision**:

- **`Schema` is a consumer-implemented interface** of two methods, both keyed by class (CORE makes the default text, and in principle any placement, a statement of the class):

  ```go
  type Schema interface {
      Property(c Class, p Predicate) (Property, bool)    // encoder: how p renders on nodes of class c
      Predicate(c Class, label string) (Predicate, bool) // decoder: which predicate a label names; "" = default text
  }
  ```

  It is supplied through `Config.Schema` and is **not copied**. An implementation must be safe for concurrent use whenever the `Codec` is. A nil `Schema` means roles are inferred on decode, observed or fallback values are used on encode, and the default text predicate is `text`.
- **`Property` stays a value struct**: `Property{Role, Label, Default, Many}` is the library's own answer about one predicate, not the application's schema type. The application's richer types (cardinality, merge, range…) produce it inside their `Property` method.
  - `Role` is one of `RoleMeta`, `RoleText`, `RoleEdge`, `RoleLink`.
  - `Label` is empty to mean the derived camelCase label.
  - `Default` marks the class's default (unlabelled) text predicate. The encoder uses it to place prose first, without a label.
  - `Many` marks a meta predicate rendered as a YAML array, or a text predicate rendered as a `## Label` block of statements.
- **`Properties` (`map[Predicate]Property`) is the reference implementation**. It ignores the class, answers `Predicate` by matching each entry's (explicit or derived) label, answers `""` with its `Default` entry, and offers `Validate() error` for static checks (at most one `Default`, unique labels). It serves the examples, the tests and applications with a small static vocabulary.
- **Validation moves from `New` to use.** An interface cannot be enumerated, so a `Property` is checked every time the codec obtains one: a `Role` out of range, or `Default` on anything other than non-`Many` `RoleText`, fails with `ErrInvalidProperty`. Whole-schema consistency (duplicate labels, agreement between `Predicate(c, "")` and `Property.Default`) is the implementation's responsibility; `Properties` guarantees it by construction plus `Validate`.
- **Lookup order is unchanged.** Decoding records the `Property` it observed for every predicate in the `Node`. When encoding a predicate, the codec looks for its `Property` in this order: `Config.Schema`, then the observed property, then a default (meta for literals, edge for refs).

**Rationale**:

- The interface is exported because consumers implement it (Principle IX) and stays narrow, with one method per direction (Principle V).
- `Property` as a value keeps the hot path free of extra dynamic dispatch, keeps `Node.Property` immutable and comparable, and avoids a nil case.
- `Many` exists because byte-identical round trip (FR-022) must tell `tags: [x]` apart from `tags: x`. In CORE that difference comes from the predicate's cardinality, which lives in the schema, so the codec has to record it. Folding "list of statements" (a `union` text predicate) into the same flag keeps `Property` to four fields.
- Keying by `Class` removes the earlier "one `Codec` per class" limitation. The decoder knows the class from `"@type"` before it scans the body.

**Alternatives considered**:

- `type Schema map[Predicate]Property` owned by `Config`, copied and validated in `New`: rejected because it forces the application to snapshot and convert its vocabulary, and cannot follow a growing one. It survives as the `Properties` reference implementation.
- `Property` as an interface: rejected because it adds nil and mutability hazards and dynamic calls for no gain, since adaptation already happens in `Schema.Property`.
- A third `Schema` method for the default text, or for enumeration: rejected because the empty label covers the default, and enumeration would only serve up-front validation, which a dynamic schema cannot promise anyway.

---

## R5. YAML front-matter parsing: dependency choice

**Decision**: Use `go.yaml.in/yaml/v3` (the YAML organisation's maintained continuation of `gopkg.in/yaml.v3`, MIT/Apache-2.0, no transitive dependencies) for **reading** front-matter only, through its `yaml.Node` tree. That tree exposes each scalar's resolved tag (`!!str`, `!!int`, `!!float`, `!!bool`, `!!timestamp`, `!!null`), its written form and its key order, which gives exactly the literal kinds FR-004 needs. **Writing** front-matter is hand-rendered from canonical rules ([contracts/format.md](contracts/format.md)), so the output is byte-stable no matter how the YAML encoder chooses its style.

Hardening rules applied on top of the parser:
- anchors and aliases are rejected (`ErrMalformed`), which removes alias-expansion bombs;
- only scalars and sequences of scalars are allowed, except under `"@conflicts"` (depth 2) and `"@provenance"` (depth 3), which bounds nesting (FR-015);
- the input is size-checked before parsing (FR-015).

Recorded in [ADR 002](../../adrs/002-runtime-dependencies.md).

**Rationale**: The standard library has no YAML parser. A hand-written YAML subset parser would have to reimplement plain/quoted/block scalar rules and flow sequences correctly: that is a large attack surface and a source of divergence from every other tool reading the graph. `gopkg.in/yaml.v3` was archived in 2025; `go.yaml.in/yaml/v3` is the drop-in successor under active maintenance.

**Alternatives considered**: `gopkg.in/yaml.v3` (archived); `github.com/goccy/go-yaml` (larger, more transitive surface); `go.yaml.in/yaml/v4` (release candidate at the time of writing, so not chosen for a v0 dependency yet); a hand-written subset parser (rejected as above).

---

## R6. Markdown body parsing

**Decision**: A hand-written, line-oriented scanner. No Markdown library.

**Rationale**: CORE §4 defines the body as a fixed, line-level grammar: an H1, paragraphs, `**Label**` lines, `predicate::` bullets, `## Label` headings, plus the inline form `[p:: [[T]]]`. A general CommonMark AST (e.g. goldmark) discards the exact source bytes that a byte-identical round trip needs, and would add a dependency with an ADR for no benefit. Fenced code blocks are tracked by the scanner so that structure-looking lines inside them stay prose (spec Edge Cases).

**Alternatives considered**: `github.com/yuin/goldmark` (rejected: lossy for round trip, a dependency); a regex-only approach (rejected: fenced-code state needs a scanner anyway).

**Revision (BUG-001, 2026-10-24)**: the scanner closes a prose value at each blank line outside a fence, so every paragraph is its own `Text`. A paragraph whose first line starts with `**Label**:` opens labelled prose for that label, as an own-line `**Label**` does, and the rest of the line (trimmed) begins its first paragraph. A bold lead without a colon stays prose; the schema is not consulted for this decision. Untyped `[[T]]` links are found by a fence-aware scan of the text values that skips inline typed edges `[p:: [[T]]]`; the result is a derived view (`Node.Links`), not triples.

---

## R7. Unicode NFC normalisation

**Decision**: `golang.org/x/text/unicode/norm` (part of the Go project, BSD-3-Clause, zero non-x transitive dependencies). Covered by ADR 002.

**Rationale**: CORE §5 and §8.5 require NFC. The standard library has no normaliser, and `golang.org/x/text` is the canonical implementation.

**Alternatives considered**: Skipping normalisation (rejected: violates CORE identity rules); a vendored normalisation table (rejected: a maintenance burden).

---

## R8. Value equality, prose and canonical order

**Decision**:

- **Equality keys**, computed by `Object.key()`:
  - `Text`: NFC, trailing whitespace removed from each line, then trimmed.
  - `Number`: trimmed lexical form.
  - `Bool`: `true`/`false`.
  - `Date`: `YYYY-MM-DD`.
  - `DateTime`: the UTC instant.
  - `Ref`: the target `ID`, ignoring display text.
  - Each key is prefixed by a variant tag, so `Text("1")` ≠ `Number(1)`.
- **Prose**: the default text and each labelled text predicate decode to **one** `Text` value that holds all their paragraphs, separated by a blank line. Paragraph order is therefore preserved even though values form a set (Story 4, scenario 2). A `Many` text predicate yields one `Text` per bullet.
- **Canonical order** ([contracts/format.md](contracts/format.md)):
  - predicates in code-point order within each section;
  - values ordered by their canonical rendering, except `Date`/`DateTime`, which are ordered by value;
  - when the same predicate holds values of mixed variants, they are ordered by variant tag first.

**Rationale**: This follows CORE §8.5 directly and makes `Node.Triples()` deterministic (FR-007, FR-018).

**Revision (BUG-001, 2026-10-24)** — supersedes the *Prose* and *Canonical order* bullets above for triples:
- **Prose**: ~~one `Text` value holding all paragraphs~~ → one `Text` per paragraph. The blob existed only to keep paragraph order inside a sorted set; with natural order it is unnecessary. Paragraph equality is unchanged (the `Text` key), so a repeated paragraph collapses to its first occurrence (CORE §8.4 `append`: "every paragraph not yet present").
- **Order**: ~~canonical order~~ → **natural order**: triples keep the order they entered the node (document order on decode, insertion order for `NewNode`); predicates are ordered by their first triple. The encoder writes that order within each CORE §4 section. CORE §3 asks only for a *stable* order, and natural order is stable: encoding the same node twice gives the same bytes.
- **Canonical order** (`compareObjects`, `compareAttributions`) is kept only for `@conflicts` and `@provenance`, which are merge state, not attributes.

**Alternatives considered (BUG-001)**: keep canonical order and add a separate "document order" iterator (rejected: two orders for one bag, and the encoder would still reshuffle the author's order); order only text values (rejected: the reader expects edges and meta values in the order written as well).

---

## R9. The inline and body edge forms when decoding

**Decision**: The **body form** (`p:: [[T]]` on its own line, without a bullet) is read as an edge triple. The **inline form** (`[p:: [[T]]]` inside prose) produces an edge triple, and the prose keeps the text verbatim. On encode, the triple is written in list form in the edge block. Re-decoding then sees the same triple twice (in the prose and in the list), and set semantics collapse it to one.

**Rationale**: CORE §7 lets a round trip normalise to the list form, while §4 says prose is never altered. Keeping the prose verbatim and adding the list bullet satisfies both and stays stable over repeated round trips.

---

## R10. Null and empty front-matter values

**Decision**: A front-matter key whose value is null or an empty sequence produces no triple and no observed property. This is non-canonical input, so byte identity is not required for it.

**Rationale**: Hand-edited files often carry `tags:` placeholders. Rejecting them would be hostile, and a bag cannot represent a predicate that has no value.

---

## R11. Go version and CI/CD

**Decision**:

- **Go version**: Lower the `go` directive from `1.27.0` to `1.26.0`. Principle XIII requires support for "at least the two most recent Go releases" (1.26 and 1.27 as of 2026-10). The project has no release yet, so this is not a compatibility event. The `check-test` matrix pins `1.26.6` and `1.27.x` (the exact latest 1.27 patch is pinned when the workflow is written).
- **Workflows**: Adapted from [fogfish/arcnet-cli](https://github.com/fogfish/arcnet-cli/tree/main/.github/workflows), keeping its three-file layout and action versions (`actions/checkout@v7`, `actions/setup-go@v6`, `dominikh/staticcheck-action@v1.3.1`, `shogo82148/actions-goveralls@v1`, `reecetech/version-increment@2023.10.2`, `goreleaser/goreleaser-action@v7`). Deltas required by the constitution:

| Workflow | Example | Change for `marc` |
|---|---|---|
| `check-code.yml` | staticcheck only | + `gofmt -l` (no output), `go vet`, `go mod tidy` diff check, `apidiff -m -incompatible` against the latest tag (pinned `golang.org/x/exp/cmd/apidiff` version; skipped while no tag exists) |
| `check-test.yml` | single Go version, no race | matrix `[1.26.6, 1.27.x]`, `go test -race -coverprofile=profile.cov ./...`, `CGO_ENABLED=0 go build ./...` (CGO-free build, Principle X); coveralls on the latest version only |
| `build.yml` | tag → govulncheck → goreleaser | `go test -race`; **govulncheck moved before tagging** and pinned to a version; increment selected from the merge commit message (`[major]`, `[minor]`, default patch) and passed to `version-increment`; goreleaser in library mode |
| `.goreleaser.yml` (new) | — | `builds: - skip: true`; changelog from GitHub, excluding `^docs:`, `^test:`, `^ci:` (Principle XIII) |

**Rationale**: This keeps the example's structure, which the user asked to follow, and closes the gaps against the Mandatory Libraries & Tooling section.

**Alternatives considered**: `softprops/action-gh-release` instead of goreleaser (rejected to stay aligned with the example; goreleaser's library mode handles changelog filtering natively).
