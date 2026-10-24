---

description: "Task list for 001-markdown-triples-codec (marc Library Constitution)"
---

# Tasks: Markdown ⇄ Triples Codec Core Types

**Input**: Design documents from `/specs/001-markdown-triples-codec/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/api.md](contracts/api.md), [contracts/format.md](contracts/format.md), [quickstart.md](quickstart.md)

**Tests**: Tests are MANDATORY (Principles VI, VIII). Every spec scenario has a black-box acceptance test,
and unit tests are written first and fail semantically before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Bugfix**: 2026-10-24 — BUG-001 Updated from bugfix patch (18 tasks reopened, T088–T099 added in Phase 9)

**Constitution**: `.specify/memory/constitution.md`. Phase 2 and Phase N are MANDATORY and MUST be retained
verbatim; only task descriptions are adapted to the feature (Governance → Task List Requirements).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Flat Go library per [ADR 001](../../adrs/001-system-architecture.md): root package `marc` at repository root
- Files by concept: `[concept].go` with `[concept]_test.go` (package `marc`, unit tests)
- Acceptance tests: `[feature]_test.go` (package `marc_test`, black-box); examples in `example_test.go`
- Fuzz tests in `fuzz_test.go` with corpora under `testdata/fuzz/`; benchmarks in `bench_test.go`
- Sub-packages only for independent vocabularies; shared non-API helpers under `internal/`

## Global coding rules (apply to every implementation task)

- No inline or block comments; GoDoc on every exported identifier, starting with its name and stating behaviour, errors, concurrency safety and ownership (Principles IV, XII)
- Functions ≤ 25 lines; no mutation of inputs; accept interfaces, return structs (Principle IV)
- Errors: wrap with `faults` (`ErrX.With(err)` / `faults.Safe1` for element names); never put payload values (identities, prose, scalars) into messages, only element names (`"@id"`, predicate name) and line numbers (Principles X, XII)
- Assertions only with `github.com/fogfish/it/v2` (`it.Then(t).Should(...)`); table-driven where parameterised; `t.Parallel()` in acceptance tests (Principles VI, VIII)
- No `os`, no stdout/stderr, no env, no mutable package-level variables (Principle X)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Module, CI/CD and fixtures (only what the feature requires)

- [X] T001 Update go.mod: change the `go` directive from `1.27.0` to `go 1.26.0` (research R11), then `go get github.com/fogfish/faults@latest go.yaml.in/yaml/v3@v3.0.4 golang.org/x/text@latest github.com/fogfish/it/v2@latest` and run `go mod tidy` so go.mod/go.sum list exactly these four direct requirements
- [X] T002 [P] Create .github/workflows/check-code.yml (name `check`, on `pull_request: [opened, synchronize]`, `ubuntu-latest`, `actions/checkout@v7` with `fetch-depth: 0`, `actions/setup-go@v6` with `go-version: "1.27.x"` pinned to the exact latest 1.27 patch). Add these steps in order: `test -z "$(gofmt -l .)"`; `go vet ./...`; `dominikh/staticcheck-action@v1.3.1` with `install-go: false` and a pinned `version`; `go mod tidy && git diff --exit-code go.mod go.sum`. Then add an apidiff step that runs only when `git describe --tags --abbrev=0` succeeds: install `golang.org/x/exp/cmd/apidiff` at a pinned pseudo-version and run `apidiff -m -incompatible` comparing the latest tag (checked out into a worktree) with HEAD, failing on any output (research R11, Principle XIV)
- [X] T003 [P] Create .github/workflows/check-test.yml (name `test`, on `pull_request: [opened, synchronize]`, a strategy matrix `go: ["1.26.6", "1.27.x-pinned"]`, `actions/checkout@v7`, `actions/setup-go@v6` with `go-version: ${{ matrix.go }}`). Steps: `CGO_ENABLED=0 go build ./...`, then `go test -race -coverprofile=profile.cov ./...`, then `shogo82148/actions-goveralls@v1` with `continue-on-error: true`, `path-to-profile: profile.cov`, run only for the latest matrix entry
- [X] T004 [P] Create .github/workflows/build.yml (name `build`, on `push: branches: [main]`, `actions/checkout@v7` with `fetch-depth: 0`, `actions/setup-go@v6` pinned to the latest 1.27 patch). Steps in this order:
  1. `go build ./...`
  2. `go test -race -coverprofile=profile.cov ./...`
  3. goveralls (`continue-on-error: true`)
  4. `go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...`, **before** tagging
  5. a shell step with `id: increment` that writes `major` if `${{ github.event.head_commit.message }}` contains `[major]`, `minor` if it contains `[minor]`, otherwise `patch`, to `$GITHUB_OUTPUT`
  6. `reecetech/version-increment@2023.10.2` with `scheme: semver` and `increment: ${{ steps.increment.outputs.increment }}`
  7. `git tag` and `git push origin` the version
  8. `goreleaser/goreleaser-action@v7` with `version: "~> v2"`, `args: release`, `GITHUB_TOKEN`
- [X] T005 [P] Create .goreleaser.yml (`version: 2`; `builds: [{ skip: true }]`; `changelog: { use: github, filters: { exclude: ["^docs:", "^test:", "^ci:"] } }`; `release: { github: { owner: net-of-arc, name: marc } }`) (Principle XIII)
- [X] T006 [P] Create the canonical fixture testdata/core-3-example.md. The bytes must be exactly as follows: each line below is one file line ending in `\n`, `⏎` marks an empty line, and the file ends with a single `\n`. Within the body, `conformsTo` sorts before `replaces` (canonical order, contracts/format.md §4).
  1. `---`
  2. `"@id": Transport Layer Security`
  3. `"@type": Entity`
  4. `category: script`
  5. `tags: [cryptography]`
  6. `---`
  7. `# Transport Layer Security`
  8. `⏎`
  9. `A cryptographic protocol that establishes an authenticated, confidential channel over an untrusted network.`
  10. `⏎`
  11. `- conformsTo:: [[RFC 8446]]`
  12. `- replaces:: [[SSL Protocol]]`
  13. `⏎`
  14. `## Mentioned In`
  15. `- mentionedIn:: [[rescorla-2026-tls13]]`
- [X] T007 [P] Create fixtures testdata/core-8-2-conflicts.md (the CORE §8.2 front-matter with `"@conflicts"` holding `category: [object, process]`, a `description` block-scalar alternative and `replaces: ["[[SSL 3.0]]"]`, plus an H1 `# Transport Layer Security`) and testdata/core-8-3-provenance.md (the CORE §8.3 front-matter with `"@provenance"` for `dierks-2008-tls12` and `rescorla-2018-tls13`, plus the H1). Both must be written in canonical form per contracts/format.md §1–§2
- [X] T008 [P] Create fixture testdata/prose-untyped-links.md: a canonical node `"@id": Example Node`, `"@type": Entity`, whose default prose is two paragraphs, the first containing `see [[RFC 8446]] for details`, followed by a labelled text predicate `**Key Findings**` with a one-line paragraph and a `## Claims` block holding two `- claims:: <statement>` bullets
- [X] T009 [P] Create fixtures testdata/forms/body-form.md (an edge written as the line `conformsTo:: [[RFC 8446]]` without a bullet) and testdata/forms/inline-form.md (prose containing `was [supports:: [[RFC 8446]]] standardized.`), plus testdata/canonical/ with at least 5 canonical nodes covering:
  - every literal kind (number `42`, bool `true`, date `2026-10-23`, datetime `2026-10-23T10:00:00Z`);
  - a quoted scalar (`"yes"`, `"a: b"`);
  - a `Many` meta with one value;
  - a link block with a non-derived label (`## Sources`);
  - an edge with display text `[[RFC 8446|TLS 1.3]]`
- [X] T010 [P] Create one fixture per error category in testdata/malformed/:
  - no-frontmatter.md
  - unterminated-frontmatter.md
  - missing-id.md
  - missing-type.md
  - missing-h1.md
  - h1-mismatch.md
  - invalid-id.md (`"@id": a/b`)
  - invalid-predicate.md (`Bad_Name: x`)
  - invalid-class.md (`"@type": entity`)
  - yaml-alias.md (an anchor `&a` with an alias `*a`)
  - nested-map.md (`meta: {k: v}`)
  - empty.md (0 bytes)

---

## Phase 2: Design Preconditions (MANDATORY)

**Purpose**: Constitution PRECONDITIONS that MUST be complete BEFORE implementation begins

**⚠️ CRITICAL**: No user story implementation can begin until this phase is complete

### Phase 2a: Domain Vocabulary & Glossary (Principle II)

- [X] T011 Rewrite ARCHITECTURE.md with three sections:
  - **Glossary**: term, definition, Go identifier for Subject/Identity → `ID`, Predicate → `Predicate`, Class → `Class`, Object → `Object` (variants `Ref`, `Text`, `Number`, `Bool`, `Date`, `DateTime`), Triple → `Triple`, Node/Bag → `Node`, Role → `Role`, Property → `Property`, Schema → `Schema`/`Properties`, Attribution → `Attribution`, Codec → `Codec`, Canonical form, Observed property. Take the definitions from data-model.md.
  - **Package list**: the single root package `marc`.
  - **NFRs**: Go ≥ 1.26 and the two latest releases; everything synchronous; every type immutable and safe for concurrent use, except that `Codec` is safe only when the consumer's `Schema` is; hot paths `Decode`/`Encode` < 1 ms for a 10 KB / 100-statement node; default input limit 4 MiB.
- [X] T012 Verify the repository has no existing type for any glossary concept (the repository holds only go.mod, README.md, ARCHITECTURE.md, adrs/); record "no duplicates" in the PR description

### Phase 2b: Public API Contract Design (Principle IX)

- [X] T013 Review specs/001-markdown-triples-codec/contracts/api.md against Principle IX:
  - no stutter;
  - `New(Config) (*Codec, error)`;
  - `Must*` only alongside `New*`;
  - no functional options;
  - every identifier traced to a story in the Traceability table.

  Record the result as accepted in the PR description.
- [X] T014 Confirm the `Config` contract in contracts/api.md: `Schema Schema` (nil ⇒ "roles inferred on decode, observed or fallback on encode; default text predicate `text`") and `MaxSize int` ("0 ⇒ 4 MiB (4 194 304 bytes)"; "`< 0` → `ErrInvalidConfig`"). No other fields.
- [X] T015 Create errors.go with the 10 exported constants exactly as in contracts/api.md §Errors: `ErrMalformed`, `ErrIdentityMismatch`, `ErrInvalidID`, `ErrInvalidName`, `ErrInvalidObject`, `ErrForeignSubject`, `ErrTooLarge`, `ErrProseStructure`, `ErrInvalidProperty` and `ErrInvalidConfig`, each a `faults.Type`. Give each a GoDoc comment naming when it is returned.
- [X] T016 Write down, as GoDoc requirements in contracts/api.md, the ownership and concurrency rule for every exported type:
  - `Node`, names and objects are immutable and safe for concurrent use;
  - `NewNode` consumes the sequence once and retains nothing;
  - `Unmarshal` does not retain its `[]byte`;
  - `Codec` keeps `Config.Schema` by reference and is safe for concurrent use when the schema is;
  - iterators yield copies.

### Phase 2c: External Interface Design (Principle VII, if applicable)

- [X] T017 Confirm the external interfaces: `io.Reader`/`[]byte` in, `io.Writer`/`[]byte` out, `iter.Seq`/`iter.Seq2` for sequences, and the consumer-implemented `Schema` interface (`Property(Class, Predicate) (Property, bool)` and `Predicate(Class, label string) (Predicate, bool)`, where `""` means the default-text predicate) declared in schema.go next to its consumer (ADR 003 §4)
- [X] T018 Move adrs/002-runtime-dependencies.md (`go.yaml.in/yaml/v3`, `golang.org/x/text`) from `Status: Proposed` to `Status: Accepted` after maintainer approval. Confirm that no technology SDK and no `os` file function is used. **Do not start Phase 3 until it is Accepted.**

### Phase 2d: Acceptance Test Design (Principle VIII)

Each test function starts with `// Scenario X.Y from specs/001-markdown-triples-codec/spec.md`, calls `t.Parallel()`, uses only exported identifiers, loads fixtures through `os.ReadFile("testdata/...")` (test code only), and asserts concrete values with `fogfish/it/v2`.

- [X] T019 ⚠️ Reopened [P] Write decode_test.go (package marc_test) for Scenarios 1.1–1.5 (reopened — BUG-001: AS 1.1 asserts document order (`category`, `tags`, `text`, `conformsTo`, `replaces`, `mentionedIn`), not canonical order; add AS 1.6):
  - **1.1**: `marc.Unmarshal(testdata/core-3-example.md)` returns `ID().String() == "Transport Layer Security"` and `Class().String() == "Entity"`, and `slices.Collect(n.Triples())` equals exactly these six triples in canonical order: `category Text("script")`, `conformsTo RefTo("RFC 8446")`, `mentionedIn RefTo("rescorla-2026-tls13")`, `replaces RefTo("SSL Protocol")`, `tags Text("cryptography")`, `text Text("A cryptographic protocol that establishes an authenticated, confidential channel over an untrusted network.")`.
  - **1.2**: inline input with `tags: [cryptography, networking]` yields two `tags` triples.
  - **1.3**: `marc.Select[marc.Ref](n.Values(MustPredicate("replaces")))` yields one `Ref` whose `Target().String() == "SSL Protocol"`.
  - **1.4**: `n.Values(MustPredicate("absent"))` yields nothing.
  - **1.5**: front-matter `n: 42`, `f: 1.5`, `b: true`, `d: 2026-10-23`, `s: 2026-10-23T10:00:00Z` yields the variants `Number` (`String()=="42"`), `Number` (`"1.5"`), `Bool(true)`, `Date` (`"2026-10-23"`) and `DateTime` (`"2026-10-23T10:00:00Z"`).
- [X] T020 ⚠️ Reopened [P] Write encode_test.go (package marc_test) for Scenarios 2.1–2.6 (reopened — BUG-001: AS 2.2 now expects insertion order `[zeta, alpha, mu]`; add AS 2.7 (multi-paragraph default and labelled text, sections per CORE §4)):
  - **2.1**: build the CORE §3 node with `NewNode` from the six triples, with `marc.New(marc.Config{Schema: marc.Properties{category: {Role: RoleMeta}, tags: {Role: RoleMeta, Many: true}, text: {Role: RoleText, Default: true}, replaces: {Role: RoleEdge}, conformsTo: {Role: RoleEdge}, mentionedIn: {Role: RoleLink}}})`. `Marshal` must equal the bytes of testdata/core-3-example.md.
  - **2.2**: `Many` meta values added in the order `zeta, alpha, mu` render as `[alpha, mu, zeta]`.
  - **2.3**: link `mentionedIn` with no label renders `## Mentioned In\n- mentionedIn:: [[X]]`.
  - **2.4**: link with `Label: "Sources"` renders `## Sources`.
  - **2.5**: a non-default text predicate `keyFindings` renders `**Key Findings**\n<paragraph>` after the default prose and before the edge block.
  - **2.6**: a node with no triples renders exactly `---\n"@id": X\n"@type": Entity\n---\n# X\n`.
- [X] T021 ⚠️ Reopened [P] Write roundtrip_test.go (package marc_test) for Scenarios 3.1–3.4 (reopened — BUG-001: AS 3.2 also compares order within each CORE §4 section):
  - **3.1**: for every file in testdata/canonical/, testdata/core-3-example.md, core-8-2-conflicts.md and core-8-3-provenance.md, `Marshal(Unmarshal(b)) == b` with the default codec (no schema).
  - **3.2**: a node built with `NewNode` plus a `Properties` schema survives `Unmarshal(Marshal(n))` with equal `slices.Collect(Triples())`.
  - **3.3**: an unknown predicate `fooBar` in labelled prose, in the edge block and in a link block is written back in the same position.
  - **3.4**: testdata/forms/body-form.md and inline-form.md each yield a triple `conformsTo`/`supports` → `RefTo("RFC 8446")`, and `Marshal` writes the list-form bullet.
- [X] T022 ⚠️ Reopened [P] Write prose_test.go (package marc_test) for Scenarios 4.1–4.3 using testdata/prose-untyped-links.md (reopened — BUG-001: AS 4.1/4.2 expect one `text` triple per paragraph in document order; add AS 4.4 (`Node.Links`) and AS 4.5 (`**Label**:` paragraph)):
  - **4.1**: the `text` value contains `[[RFC 8446]]` verbatim, and no triple has `RefTo("RFC 8446")` as its object.
  - **4.2**: the decoded `text` value holds both paragraphs separated by `\n\n` in the original order, and `Marshal` reproduces the file.
  - **4.3**: the `## Claims` block yields two `claims` triples of variant `Text`.
- [X] T023 [P] Write malformed_test.go (package marc_test) for Scenarios 5.1–5.7, using a table over testdata/malformed/ with expected `errors.Is` targets:
  - **5.1**: no-frontmatter, unterminated-frontmatter, missing-id, missing-type and empty → `ErrMalformed`.
  - **5.2**: h1-mismatch → `ErrIdentityMismatch`.
  - **5.3**: invalid-id → `ErrInvalidID`; also `NewID` with each of `/ \ : * ? " < > | . # ^ [ ]`, `" lead"`, `"trail "` and `"two  spaces"` → `ErrInvalidID`.
  - **5.4**: invalid-predicate and invalid-class → `ErrInvalidName`; also `NewPredicate("Bad")` and `NewClass("bad")` → `ErrInvalidName`.
  - **5.5**: `New(Config{MaxSize: 64})` followed by `Decode` of a 65-byte reader → `ErrTooLarge`, with a reader that counts bytes proving at most 65 bytes were read.
  - **5.6**: encoding a node whose `text` value contains the line `## Heading`, or the line `- p:: x`, → `ErrProseStructure`, while the same lines inside a ```` ``` ```` fence encode fine.
  - **5.7**: covered by FuzzDecode (T040); here, assert that 1 000 pseudo-random byte slices from a fixed seed return either a node or an error, never a panic.
  - Also yaml-alias and nested-map → `ErrMalformed`.
- [X] T024 [P] Write mergestate_test.go (package marc_test) for Scenarios 6.1–6.3:
  - **6.1**: core-8-2-conflicts.md → `Conflicts()` yields `(category, Text("object"))`, `(category, Text("process"))`, `(description, Text(...))` and `(replaces, RefTo("SSL 3.0"))`, and no triple has these objects.
  - **6.2**: core-8-3-provenance.md → `Provenance()` yields `Attribution{Document: dierks-2008-tls12, Predicate: category, Object: Text("script")}` and the remaining entries, and none of them appear in `Triples()`.
  - **6.3**: `Marshal(Unmarshal(b)) == b` for both fixtures.
- [X] T025 Add minimal compiling stubs for every identifier in contracts/api.md, with zero-value returns and `nil` errors where needed, in the files named by plan.md: id.go, name.go, object.go, triple.go, node.go, schema.go, config.go and codec.go. Confirm that `go vet ./...` compiles and `go test ./...` fails semantically: no `t.Skip()`, no "red phase" comments, no `TODO`.

### Phase 2e: Dependency, Compatibility & Graduation Review (Principles III, XIII, XIV)

- [X] T026 Check the ADR 001 graduation criteria and record "none met" in the PR description: one codec process, no interchangeable technologies, no SDK, one package, no composition step
- [X] T027 Classify the SemVer impact as MINOR (initial API, first tag `v0.1.0`; `apidiff` is skipped because no tag exists). Move adrs/003-public-api-shape.md (sealed `Object`, validated names, context-free codec, consumer-owned `Schema`, iterators) from `Status: Proposed` to `Status: Accepted` after maintainer approval
- [X] T028 Confirm that the `go` directive is `1.26.0` (lowered before the first release per research R11, so no MINOR note is needed), that there is a single module, and that go.mod has no `replace` directive

**Checkpoint**: Preconditions complete. Acceptance tests are red, the API contract is agreed, ADR 002/003 are Accepted, and user story implementation can begin.

---

## Phase 3: User Story 1 - Read a node file as a bag of triples (Priority: P1) 🎯 MVP

**Goal**: Decode one canonical Markdown node into a `Node` whose triples, predicates and typed objects can be iterated.

**Independent Test**: `go test -race -run TestDecode .` passes for Scenarios 1.1–1.5.

### Unit Tests for User Story 1 (write FIRST, MUST FAIL) ⚠️

- [X] T029 [P] [US1] Table-driven tests in id_test.go (package marc):
  - `NewID` accepts `Transport Layer Security`, `RFC 8446` and `rescorla-2026-tls13`;
  - it rejects empty input, each of `/ \ : * ? " < > | . # ^ [ ]`, a leading or trailing space, two consecutive spaces, `\n`, control characters and invalid UTF-8;
  - NFC: the decomposed `é` equals the composed `é`;
  - `MustID` panics on invalid input;
  - the zero `ID` reports `IsZero()`.
- [X] T030 [P] [US1] Table-driven tests in name_test.go:
  - `NewPredicate` accepts `^[a-z][A-Za-z0-9]*$` (`tags`, `mentionedIn`, `rfc8446`) and rejects `Tags`, `bad_name`, `@id` and the empty string;
  - `NewClass` accepts `^[A-Z][A-Za-z0-9]*$` and rejects `entity`;
  - `Predicate.Label()`: `mentionedIn` → `Mentioned In`, `keyFindings` → `Key Findings`, `tags` → `Tags`, `rfc8446` → `Rfc 8446`.
- [X] T031 [P] [US1] Table-driven tests in object_test.go:
  - equality keys: `Text(" a ")` equals `Text("a")`; NFC; per-line trailing whitespace is ignored; `RefTo(x).WithDisplay("d")` equals `RefTo(x)`; `Text("1")` differs from `ParseNumber("1")`; `DateTimeOf` at `+02:00` equals the same UTC instant;
  - `ParseNumber` rejects `abc` and the empty string with `ErrInvalidObject`;
  - `Int64` and `Float64` conversions;
  - canonical `String()` renderings per contracts/format.md §3;
  - canonical order: variant tag `b < d < n < r < s < t`, then dates by value, otherwise by code point;
  - `Select[Ref]` filters a mixed sequence.
- [X] T032 ⚠️ Reopened [P] [US1] Table-driven tests in node_test.go (reopened — BUG-001: `Triples()`, `Predicates()` and `Values()` follow natural (insertion) order, and a duplicate keeps its first position):
  - `NewNode` collapses duplicate triples (set semantics, FR-006);
  - it rejects a foreign subject (`ErrForeignSubject`), a zero predicate (`ErrInvalidName`), a nil object, a zero `Ref`, `Number`, `Date` or `DateTime` (`ErrInvalidObject`), and a zero id or class;
  - `Triples()`, `Predicates()` and `Values()` follow canonical order;
  - iteration stops early when `yield` returns false;
  - concurrent iteration from 8 goroutines passes under `-race`.
- [X] T033 [P] [US1] Tests in frontmatter_test.go for the read path:
  - splitting `---\n…\n---\n` from the body, including CRLF input;
  - `yaml.Node` scalar tags map to `Text`, `Number`, `Bool`, `Date` or `DateTime` as in Scenario 1.5;
  - a flow or block sequence yields one value per element and the observed property `{RoleMeta, Many: true}`;
  - a scalar yields `{RoleMeta}`;
  - a quoted `"[[T]]"` string yields `Ref`;
  - null and empty values yield no triple (research R10);
  - `"@id"` and `"@type"` are extracted and never become triples.
- [X] T034 ⚠️ Reopened [P] [US1] Tests in body_test.go for the line scanner on canonical input (reopened — BUG-001: unlabelled paragraphs become one `Text` per paragraph, not one joined value):
  - the H1 is extracted;
  - unlabelled paragraphs become one `Text` joined by `\n\n` on the default-text predicate (`text`, or `Schema.Predicate(class, "")`);
  - `- p:: [[T]]` and `- p:: [[T|d]]` bullets become `RoleEdge` triples;
  - a `## Label` block of ref bullets becomes `{RoleLink, Label: <heading if ≠ derived>}`.

### Implementation for User Story 1

- [X] T035 [P] [US1] Implement `ID`, `NewID`, `MustID`, `String` and `IsZero` in id.go: `type ID struct{ value string }`, NFC through `norm.NFC.String`, rejecting `/ \ : * ? " < > | . # ^ [ ]`, leading or trailing spaces, consecutive spaces, empty input, invalid UTF-8 and control characters with `ErrInvalidID` (the message must not contain the value)
- [X] T036 [P] [US1] Implement `Predicate` and `Class` (`NewX`, `MustX`, `String`, `IsZero`) and `Predicate.Label()` in name.go: camelCase is `^[a-z][A-Za-z0-9]*$` and UpperCamelCase is `^[A-Z][A-Za-z0-9]*$`, both validated without regexp in a ≤ 25-line function; the label splits at lower→upper and letter→digit boundaries and capitalises each word
- [X] T037 [US1] Implement the sealed `Object` interface (`String() string`, `key() string`) and its variants in object.go:
  - `Ref` with `RefTo`, `WithDisplay`, `Target` and `Display`;
  - `Text`;
  - `Number` with `ParseNumber` (accepting the YAML 1.2 core int/float lexical forms), `Int64` and `Float64`;
  - `Bool`;
  - `Date` with `DateOf` and `Time`;
  - `DateTime` with `DateTimeOf` (UTC) and `Time`;
  - `Select[T Object]`;
  - an unexported `compareObjects(a, b Object) int` implementing canonical order.

  Depends on T035.
- [X] T038 ⚠️ Reopened [US1] Implement `Triple` and `Attribution` in triple.go. Implement `Node` and `NewNode` in node.go: an internal map `Predicate → []Object` deduplicated by `key()` and kept sorted by `compareObjects`; validation as in T032; the accessors `ID`, `Class`, `Triples`, `Predicates` and `Values`, all canonical iterators. Depends on T037. (reopened — BUG-001: keep triples in insertion order, deduplicated by key; do not sort by `compareObjects`; iterators yield natural order)
- [X] T039 [US1] Implement in schema.go: `Role` with its constants and `String`; `Property`; the `Schema` interface; and `Properties` with `Property` (map lookup) and `Predicate(c, label)` (matching the explicit or derived label; `""` returns the `Default` entry). Add an unexported `validProperty(p Property) error` that returns `ErrInvalidProperty` when `Role` is out of range or `Default` is set on anything other than non-`Many` `RoleText`.
- [X] T040 [P] [US1] Write `FuzzDecode` in fuzz_test.go, seeded from every file under testdata/ and stored in testdata/fuzz/FuzzDecode/. The property: `Unmarshal` never panics and returns either a node or an error.
- [X] T041 [US1] Implement `Config`, its validation (`MaxSize < 0` → `ErrInvalidConfig` naming `MaxSize`; `0` → `4194304`) and `New` in config.go/codec.go. Implement the `Codec` struct holding the schema by reference, the package-level immutable default codec built from `New(Config{})`, and the methods/functions `Decode`/`Unmarshal`. Encode and Marshal stay as stubs until US2.
- [X] T042 [US1] Implement the front-matter read in frontmatter.go:
  - `splitFrontMatter([]byte) (yamlBytes, body []byte, error)` (accepting `\n` and `\r\n`);
  - `parseFrontMatter` over `yaml.Node`, mapping tags `!!str`→`Text` (or `Ref` when the value is `[[T]]`), `!!int`/`!!float`→`Number`, `!!bool`→`Bool`, `!!timestamp`→`Date`/`DateTime`, `!!null`→skip;
  - extraction of `"@id"`/`"@type"` → `ErrMalformed` when missing;
  - recording of the observed `Property`.

  Keep each function ≤ 25 lines.
- [X] T043 ⚠️ Reopened [US1] Implement the canonical-input line scanner in body.go: an H1 check (missing → `ErrMalformed`; ≠ `@id` → `ErrIdentityMismatch`), paragraph accumulation into the default text, edge bullets, `## Label` link blocks, and per-predicate observed properties with first-position-wins. Use `Schema.Predicate(class, label)` when a schema is configured. (reopened — BUG-001: paragraphs are split into separate values, not accumulated into one)
- [X] T044 [US1] Implement the assembly in decode.go: limit the reader (detailed in US5), split, parse front-matter, scan the body, then build the `Node` with its observed properties through an unexported constructor. Make Scenarios 1.1–1.5 green.
- [X] T045 [US1] Add `ExampleDecode` (printing the CORE §3 triples, using ≤ 3 library calls, SC-003) and `ExampleSelect` (printing only the reference targets) with `// Output:` blocks to example_test.go, and GoDoc for every identifier added in T035–T044

**Checkpoint**: `go test -race -run 'TestDecode|Example(Decode|Select)' .` is green, so the MVP reader is usable.

---

## Phase 4: User Story 2 - Write a bag of triples as a canonical node file (Priority: P1)

**Goal**: Encode a `Node` and a schema into canonical CORE §4 Markdown.

**Independent Test**: `go test -race -run TestEncode .` passes for Scenarios 2.1–2.6.

### Unit Tests for User Story 2 (write FIRST, MUST FAIL) ⚠️

- [X] T046 ⚠️ Reopened [P] [US2] Tests in frontmatter_test.go for the write path (reopened — BUG-001: meta predicates appear in natural order, not code-point order; flow-sequence values in natural order):
  - the `"@id"`/`"@type"` keys are quoted;
  - meta predicates appear in code-point order;
  - a single value renders as a scalar, while `Many` or more than one value renders as `[a, b]`;
  - `Text` stays plain when it reads back as the same `!!str`, and is otherwise double-quoted with JSON escapes. The quoted cases are `yes`, `true`, `1`, `~`, `null`, `2026-10-23`, `a: b`, `a #b`, `- x`, `[x]`, ` lead`, the empty string and `"q"`; the plain cases are `script`, `RFC 8446` and `cryptography`;
  - `Number`/`Bool`/`Date`/`DateTime` render unquoted;
  - a `Ref` renders as `"[[T]]"`.

  Include a property test: every rendered `Text` parsed back with `yaml.Unmarshal` equals the input.
- [X] T047 ⚠️ Reopened [P] [US2] Tests in encode_internal_test.go (package marc) (reopened — BUG-001: the edge block, labelled text and link blocks follow natural order, not (predicate, value) order):
  - the property lookup order is `Config.Schema` → observed → fallback (meta for literals, edge for `Ref`);
  - an invalid `Property` from the schema → `ErrInvalidProperty`;
  - section layout and blank-line rules per contracts/format.md §1;
  - the edge block is sorted by (predicate, value);
  - link blocks are sorted by predicate;
  - a zero `Node` → `ErrMalformed`.
- [X] T048 [P] [US2] Tests in schema_test.go for `Properties.Validate`: two `Default` entries, a `Default` with `RoleEdge`, a `Default` with `Many`, duplicate labels (an explicit `"Mentioned In"` colliding with the derived label of `mentionedIn`) and a `Role(9)` → `ErrInvalidProperty`; a valid schema → nil

### Implementation for User Story 2

- [X] T049 [US2] Implement `Properties.Validate()` in schema.go per T048
- [X] T050 [US2] Implement the canonical front-matter writer in frontmatter.go: `renderFrontMatter(n Node, props func(Predicate) Property) []byte` plus `renderScalar(Object) string` with the plain-vs-quoted rule of contracts/format.md §2, using JSON-style escaping for quoted scalars. `"@conflicts"`/`"@provenance"` are left for US6.
- [X] T051 ⚠️ Reopened [US2] Implement encode.go with one function per section, each ≤ 25 lines (reopened — BUG-001: every section is written in natural order; a text predicate writes one paragraph per value):
  - `resolveProperty` (schema → observed → fallback, validated through `validProperty`);
  - the H1;
  - the default prose;
  - labelled prose (`**Label**\n<text>`);
  - the edge block (`- p:: [[T]]` / `[[T|d]]`);
  - link and `Many`-text blocks (`## Label\n- p:: v`).

  Concatenate the sections with a single blank line between them and finish with exactly one trailing `\n`.
- [X] T052 [US2] Wire `Codec.Encode`/`Codec.Marshal` and the package-level `Encode`/`Marshal` in codec.go, wrapping writer errors with `faults`. Make Scenarios 2.1–2.6 green.
- [X] T053 [US2] Add `ExampleEncode` (building the CORE §3 node with `Properties` and printing the canonical file) to example_test.go, plus GoDoc for every identifier added in T049–T052, including the `Schema` concurrency and consistency contract

**Checkpoint**: US1 and US2 both pass independently.

---

## Phase 5: User Story 3 - Lossless round trip (Priority: P1)

**Goal**: decode∘encode is the identity on canonical files, encode∘decode is the identity on bags, unknown predicates keep their position, and the body and inline forms normalise to the list form.

**Independent Test**: `go test -race -run TestRoundTrip .` passes for Scenarios 3.1–3.4 and the generated corpus.

### Unit Tests for User Story 3 (write FIRST, MUST FAIL) ⚠️

- [X] T054 [P] [US3] Tests in body_test.go for the tolerant reader (contracts/format.md §5):
  - the body-form line `p:: [[T]]` becomes an edge;
  - the inline form `[p:: [[T]]]` inside prose becomes an edge while the prose stays verbatim;
  - `**Label**` followed by bullets inside the edge block is a grouping label and is ignored;
  - a predicate seen in two positions keeps the first observed property;
  - a link-block bullet whose predicate differs from the heading's is decoded under its own predicate.
- [X] T055 ⚠️ Reopened [P] [US3] Write `TestRoundTripGenerated` in roundtrip_test.go: a deterministic generator (fixed seed, `math/rand/v2`) of at least 1 000 valid nodes over random valid IDs, predicates, all object variants and a matching `Properties`. Assert `Unmarshal(Marshal(n))` has triples equal to `n` (SC-002). (reopened — BUG-001: also assert the relative order of triples within each CORE §4 section (SC-008))
- [X] T056 ⚠️ Reopened [P] [US3] Write `FuzzRoundTrip` in fuzz_test.go with its corpus in testdata/fuzz/FuzzRoundTrip/. The property: when `Unmarshal(b)` succeeds, `Marshal` succeeds or returns `ErrProseStructure`, and `Unmarshal(Marshal(n))` has the same triples and observed roles. (reopened — BUG-001: also assert per-section order is preserved)

### Implementation for User Story 3

- [X] T057 [US3] Extend body.go with body-form and inline-form recognition, edge-block grouping labels and the first-position-wins rule, so that T054 and Scenario 3.4 pass
- [X] T058 [US3] Ensure observed properties carry `Label` only when it differs from `Predicate.Label()`, and that the encoder uses the observed role for unknown predicates (Scenario 3.3). Fix any byte-level drift found by T021, T055 and T056 in frontmatter.go/encode.go.

**Checkpoint**: Scenarios 3.1–3.4 are green, and `go test -fuzz FuzzRoundTrip -fuzztime 60s .` finds nothing.

---

## Phase 6: User Story 4 - Prose and untyped links stay intact (Priority: P2)

**Goal**: Prose decodes into `Text` values with untyped links kept verbatim and paragraph order preserved, and statement lists decode to one triple per bullet.

**Independent Test**: `go test -race -run TestProse .` passes for Scenarios 4.1–4.3.

### Unit Tests for User Story 4 (write FIRST, MUST FAIL) ⚠️

- [X] T059 ⚠️ Reopened [P] [US4] Tests in body_test.go (reopened — BUG-001: multi-paragraph prose yields one `Text` per paragraph; fence blank lines do not split; add the `**Label**:` cases):
  - an untyped `[[T]]` produces no triple;
  - multi-paragraph prose is joined by `\n\n` with the original line breaks inside each paragraph preserved;
  - lines inside ```` ``` ```` or `~~~` fences that look like `# H`, `## H` or `- p:: x` stay prose;
  - a `## Label` block of non-link bullets becomes `{RoleText, Many: true}`, one `Text` per bullet;
  - `**Label**` followed by a paragraph becomes `{RoleText}` on `Schema.Predicate(class, label)` or on the camelCase join (`Key Findings` → `keyFindings`).

### Implementation for User Story 4

- [X] T060 ⚠️ Reopened [US4] Extend body.go with fence tracking, statement blocks (`Many` text) and labelled-prose resolution through the schema or the camelCase join, recording `Label` only when it differs from the derived label (reopened — BUG-001: add paragraph splitting and the `**Label**:` lead)
- [X] T061 ⚠️ Reopened [US4] Extend encode.go so that `Many` text predicates render as `## Label\n- p:: <statement>` blocks, ordered among the link blocks by predicate. Make Scenarios 4.1–4.3 green. (reopened — BUG-001: `Many` text blocks are ordered among the link blocks by natural order; a non-`Many` text predicate with several values renders as paragraphs, not bullets)

**Checkpoint**: Scenarios 4.1–4.3 are green, and US1–US3 stay green.

---

## Phase 7: User Story 5 - Reject malformed input with actionable errors (Priority: P2)

**Goal**: Every defect class returns a distinct, `errors.Is`-matchable error. Input is bounded and the decoder never panics.

**Independent Test**: `go test -race -run TestMalformed .` passes for Scenarios 5.1–5.7.

### Unit Tests for User Story 5 (write FIRST, MUST FAIL) ⚠️

- [X] T062 [P] [US5] Tests in decode_internal_test.go (package marc):
  - an `io.LimitReader(max+1)` read of more than `max` bytes → `ErrTooLarge`;
  - `Unmarshal` checks `len(b) > max` before parsing;
  - YAML anchors or aliases anywhere → `ErrMalformed`;
  - a mapping value outside `"@conflicts"`/`"@provenance"` → `ErrMalformed`;
  - an invalid predicate key → `ErrInvalidName`;
  - an invalid `"@type"` → `ErrInvalidName`;
  - an invalid link target → `ErrInvalidID`;
  - a `## Label` block containing a non-bullet line → `ErrMalformed`;
  - a bullet without `::` inside a block → `ErrMalformed`;
  - error messages do not contain the offending value (assert the message lacks the fixture's payload string).
- [X] T063 [P] [US5] Tests in encode_internal_test.go for `ErrProseStructure`: prose lines matching `^#{1,6} ` or `^- [a-z][A-Za-z0-9]*::` outside fences are rejected, while the same lines inside fences are accepted

### Implementation for User Story 5

- [X] T064 [US5] Implement the size limit in decode.go (`io.ReadAll(io.LimitReader(r, max+1))` → `ErrTooLarge` when it reads more than `max`; `Unmarshal` checks the length first), plus anchor/alias rejection and nested-map rejection in frontmatter.go
- [X] T065 [US5] Implement the prose structure guard in encode.go (fence-aware, returning `ErrProseStructure` naming the predicate) and the remaining malformed-block errors in body.go. Make Scenarios 5.1–5.7 green.

**Checkpoint**: Scenarios 5.1–5.7 are green, and `go test -fuzz FuzzDecode -fuzztime 60s .` finds nothing.

---

## Phase 8: User Story 6 - Merge-state keywords are carried, not lost (Priority: P3)

**Goal**: `"@conflicts"` and `"@provenance"` decode into structured data beside the triples and are written back unchanged in canonical order.

**Independent Test**: `go test -race -run TestMergeState .` passes for Scenarios 6.1–6.3.

### Unit Tests for User Story 6 (write FIRST, MUST FAIL) ⚠️

- [X] T066 [P] [US6] Tests in frontmatter_test.go:
  - `"@conflicts"` must be a map whose values are arrays of scalars (depth 2); `"@provenance"` must be a map of document → map of predicate → array (depth 3); deeper nesting → `ErrMalformed`;
  - the `"[[T]]"` string → `Ref`; a block scalar → multi-line `Text`; `sha256:…` → `Text`;
  - document keys are validated as `ID` (`ErrInvalidID`) and predicate keys as `Predicate` (`ErrInvalidName`);
  - rendering: block maps in key order, flow sequences in canonical order, multi-line `Text` as a `- |` block-sequence item.

### Implementation for User Story 6

- [X] T067 [US6] Extend frontmatter.go to parse and render `"@conflicts"` (stored as predicate → ordered objects) and `"@provenance"` (stored as ordered `Attribution`s). Never add their values to the triples.
- [X] T068 [US6] Implement `Node.Conflicts()` (`iter.Seq2[Predicate, Object]`) and `Node.Provenance()` (`iter.Seq[Attribution]`) in node.go, both in canonical order. Make Scenarios 6.1–6.3 green.

**Checkpoint**: All six stories are green under `-race`.

---

## Phase 9: Bugfix BUG-001 — natural order, paragraph values, `**Label**:`, untyped-link view

**Goal**: triples keep natural order (document order on decode, insertion order for `NewNode`) within the CORE §4 sections; each prose paragraph is one `Text`; a `**Label**:` paragraph is labelled prose; `Node.Links()` exposes untyped links without making them triples ([bugs/BUG-001.md](bugs/BUG-001.md), FR-007, FR-009, FR-016, FR-018, FR-020, FR-023, FR-026).

**Independent Test**: `go test -race ./...` passes, including Scenarios 1.6, 2.2 (revised), 2.7, 3.2 (revised), 4.1–4.5 and SC-008.

### Fixtures and tests (write FIRST, MUST FAIL) ⚠️

- [X] T088 [P] [US4] Add fixtures: testdata/canonical/multi-paragraph.md (two default paragraphs, a `**Note**` with two paragraphs, an edge block written `replaces` before `conformsTo`, `tags: [zeta, alpha]`, and a fenced block containing a blank line inside one paragraph) and testdata/forms/inline-label.md (a paragraph `**Note**: This is a note.` plus a paragraph `**Important** results follow.`). Re-check that every existing file under testdata/canonical/ is still canonical under natural order
- [X] T089 [P] [US1] Tests in node_test.go: `NewNode` with triples in the order `tags zeta, text b, tags alpha, text a, tags zeta` yields `Triples()` in that order minus the duplicate; `Predicates()` yields `tags, text`; `Values(tags)` yields `zeta, alpha`
- [X] T090 [P] [US4] Tests in body_test.go: two blank-line-separated paragraphs → two `Text` values on the default predicate; a blank line inside a ```` ``` ```` fence does not split; paragraphs after `**Note**` all belong to `note` until the next label or structure; `**Note**: text` → `note "text"` with `{RoleText}`; `**Note** text` (no colon) stays default prose; an inline typed edge in a paragraph still yields its edge triple
- [X] T091 [P] [US2] Tests in encode_internal_test.go: meta keys, flow-sequence values, labelled text, edge bullets and link blocks are written in natural order inside their CORE §4 sections, whatever the cross-section insertion order; several default-text values render as blank-line-separated paragraphs; a labelled predicate renders `**Label**` once followed by its paragraphs; a `Text` with a blank line outside a fence moves to front-matter; a paragraph whose first line begins with `**Label**:` → `ErrProseStructure`
- [X] T092 [P] [US4] Tests for `Node.Links()` in node_test.go: links from default and labelled text and from `Many` statements, paired with their predicate, in natural order; `[[T|d]]` yields `Ref` with display `d`; a repeated target within one predicate is listed once; links inside fences and inside `[p:: [[T]]]` are not listed; a node without text yields nothing; early `yield` stop

### Implementation

- [X] T093 [US1] Rework the builder in node.go to an insertion-ordered set: a dedup map from `predicate\x00key` plus one triple slice in arrival order; `Triples`, `Predicates` (first triple) and `Values` iterate it; keep `compareObjects`/`compareAttributions` only for `@conflicts`/`@provenance`. Depends on T089. Unblocks the reopened T038, T032
- [X] T094 [US4] Extend body.go: close a prose value at each blank line outside a fence (one `Text` per paragraph, in document order); recognise a paragraph lead `**Label**:` (label resolved as today, rest of the line trimmed starts the first paragraph); a bold lead without a colon stays prose. Keep functions ≤ 25 lines. Depends on T090, T093. Covers the reopened T043, T060
- [X] T095 [US2] Rework encode.go: sections follow node order instead of sorting; a non-`Many` text predicate renders one paragraph per value (default prose first, then `**Label**` + paragraphs); extend the body-renderability check (no blank line outside fences) and the prose-structure guard (`**Label**:` first line). Depends on T091, T093. Covers the reopened T051, T061, T046, T047
- [X] T096 [US4] Implement `Node.Links() iter.Seq2[Predicate, Ref]` in node.go (fence-aware scan of `Text` values, skipping inline typed edges, deduplicated per predicate and target), with GoDoc stating that links are derived and never triples, plus ownership and concurrency. Depends on T092, T093
- [X] T097 [US3] Make the reopened T019–T022, T055 and T056 green: update the expected orders and add Scenarios 1.6, 2.7, 4.4 and 4.5 (`// Scenario X.Y from specs/001-markdown-triples-codec/spec.md`); extend the generated round-trip corpus with multi-paragraph and labelled text and a per-section order assertion (SC-008). Depends on T093–T096
- [X] T098 [P] Update docs: `ExampleDecode` output order, add `ExampleNode_Links` to example_test.go, GoDoc of `Node`/`Triples`/`Encode` (natural order), doc.go and README quick start, ARCHITECTURE.md glossary (*Untyped link*, *Paragraph*, *Natural order*) and quickstart.md if it states canonical order. Depends on T097
- [X] T099 Re-run the benchmarks (T070) and confirm Decode/Encode stay < 1 ms/op; update the baseline in ARCHITECTURE.md. Depends on T097

**Checkpoint**: all 32 scenarios green under `-race`; reopened tasks T019–T022, T032, T034, T038, T043, T046, T047, T051, T055, T056, T059–T061 can be ticked; then re-verify T078 and T081.

---

## Phase N-1: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T069 [P] Rewrite README.md with the purpose (CORE Markdown ⇄ triples), `go get github.com/net-of-arc/marc`, a quick start (`Unmarshal`, iterate `Triples()`, a type switch on `Object`, `Marshal` with `Properties`), the supported Go versions (1.26, 1.27), a pkg.go.dev badge and link, and a link to CORE.md. Write doc.go with the package comment and the same quick start.
- [X] T070 [P] Write `BenchmarkDecode` and `BenchmarkEncode` in bench_test.go over a generated 10 KB / 100-statement fixture testdata/bench-100.md, with `b.ReportAllocs()`. Record ns/op and allocs/op as the baseline in ARCHITECTURE.md under hot paths. Confirm the time is < 1 ms/op (SC-006).
- [X] T071 (OPTIONAL, run only on explicit request; long fuzzing is CPU-heavy) Run `go test -fuzz FuzzDecode -fuzztime 10m .` and `go test -fuzz FuzzRoundTrip -fuzztime 10m .` (SC-004), and commit any new failing inputs, fixed, to testdata/fuzz/. Regular `go test` replays the seed corpora only. Last run: FuzzDecode 10m clean (69M execs); FuzzRoundTrip 5m, then the worker crashed under concurrent load, and the saved input passes on replay
- [X] T072 Refactor for readability across *.go: functions ≤ 25 lines, no inline comments, no input mutation, files by concept per plan.md
- [X] T073 Run every command in specs/001-markdown-triples-codec/quickstart.md §1–§6 and confirm the expected outcomes

---

## Phase N: Constitution Compliance Verification (MANDATORY)

**Purpose**: Constitution Implementation Phase checklist, verified per principle before the feature is complete

- [X] T074 Principle I: ARCHITECTURE.md (glossary, package list, NFRs, hot-path baseline) and ADR 002/003 (Accepted) reflect the design; ADR 001 is not contradicted
- [X] T075 Principle II: Glossary terms match the exported identifiers (`ID`, `Predicate`, `Class`, `Object`, `Ref`, `Text`, `Number`, `Bool`, `Date`, `DateTime`, `Triple`, `Node`, `Attribution`, `Role`, `Property`, `Schema`, `Properties`, `Codec`, `Config`); one concept, one name
- [X] T076 Principle III: Flat root package `marc` per ADR 001; no sub-packages, no `internal/`
- [X] T077 Principle IV: No inline/block comments except GoDoc; functions ≤ 25 lines; inputs not mutated; accept interfaces (`io.Reader`, `io.Writer`, `Schema`), return structs (`Node`, `*Codec`)
- [X] T078 ⚠️ Reopened Principle V: No catch-all packages; every exported identifier appears in the contracts/api.md Traceability table (reopened — BUG-001: `Node.Links` must appear in the contracts/api.md Traceability table)
- [X] T079 Principle VI: Unit tests were written first and changed minimally; `go test -race ./...` passes; `FuzzDecode` and `FuzzRoundTrip` exist with corpora; `Decode`/`Encode` are benchmarked; `fogfish/it/v2` is used exclusively
- [X] T080 Principle VII: The only runtime dependencies are `fogfish/faults`, `go.yaml.in/yaml/v3` and `golang.org/x/text` (ADR 002); no technology SDK; no `os` in non-test files
- [X] T081 ⚠️ Reopened Principle VIII: All 28 spec scenarios have a passing black-box acceptance test in package `marc_test` that changed minimally since Phase 2d (reopened — BUG-001: the spec now has 32 scenarios (adds 1.6, 2.7, 4.4, 4.5))
- [X] T082 Principle IX: Naming, `New(Config)`, typed `Config` with documented zero-value defaults, error and ownership rules hold; no functional options; the context-free codec matches ADR 003 §3
- [X] T083 Principle X: No `os.Exit`, stdout/stderr writes, environment reads, mutable globals or side-effecting `init()`; `CGO_ENABLED=0 go build ./...` succeeds
- [X] T084 Principle XI: No goroutines; input bounded by `MaxSize`; YAML aliases rejected; nesting bounded
- [X] T085 Principle XII: GoDoc on the package and on every exported identifier; examples for `Decode`, `Encode` and `Select`; errors via `fogfish/faults`, matched with `errors.Is`; messages free of payload values
- [X] T086 Principle XIII: `go mod tidy` leaves `go.mod`/`go.sum` unchanged; no `replace`, `vendor/` or `go.work`; the three workflows and .goreleaser.yml are present
- [X] T087 Principle XIV: `apidiff` is wired (no prior tag, so MINOR v0.1.0); `gofmt -l`, `go vet` and `staticcheck` are clean; `govulncheck` passes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies. T001 first; T002–T010 can then run in parallel.
- **Design Preconditions (Phase 2)**: Depends on Setup and BLOCKS all user stories. T018 and T027 (ADR acceptance) need maintainer approval. T025 (stubs) follows T015 and T019–T024.
- **US1 (Phase 3)**: Depends on Phase 2. It delivers the domain types (`ID`, names, `Object`, `Node`, `Schema`, `Codec`) that every later story uses.
- **US2 (Phase 4)**: Depends on US1's domain types (T035–T041), not on its decoder (T042–T044). It can start as soon as T041 is done.
- **US3 (Phase 5)**: Depends on US1 and US2, because a round trip needs both directions.
- **US4 (Phase 6)**: Depends on US1 (decode). Scenario 4.2's re-encode also needs US2.
- **US5 (Phase 7)**: Decode errors depend on US1, and `ErrProseStructure` depends on US2.
- **US6 (Phase 8)**: Depends on US1 and US2 (front-matter read and write).
- **Bugfix BUG-001 (Phase 9)**: Depends on US1–US6 being implemented. Order: T088–T092 (tests, parallel) → T093 → T094 ∥ T095 ∥ T096 → T097 → T098 ∥ T099. Reopens tasks in Phases 2d, 3, 4, 5, 6 and N; they close when T097 is green.
- **Polish (Phase N-1)**: Depends on all stories.
- **Compliance Verification (Phase N)**: The final gate.

### Story completion order

```text
Setup → Preconditions → US1 ─┬─→ US2 ─┬─→ US3
                             │        ├─→ US5
                             │        └─→ US6
                             └─→ US4 (needs US2 for 4.2)
```

### Within Each User Story

- Unit tests MUST be written and FAIL semantically before implementation
- Domain types before the functions that use them (T035/T036 → T037 → T038)
- GoDoc and examples in the same story as the API they document
- A story is complete (acceptance tests green under `-race`) before moving to the next priority

### Parallel Opportunities

- Phase 1: T002–T010 (workflows, goreleaser, fixtures) all touch different files
- Phase 2d: T019–T024 (six acceptance test files)
- US1: the unit tests T029–T034 in parallel; then T035 ∥ T036, and T040 alongside the implementation
- US2: T046 ∥ T047 ∥ T048
- After US2: US3, US5 and US6 touch mostly different functions and can proceed in parallel, coordinating edits to body.go, frontmatter.go and encode.go

### Parallel example: User Story 1

```text
Run together: T029 id_test.go · T030 name_test.go · T031 object_test.go · T032 node_test.go · T033 frontmatter_test.go · T034 body_test.go
Then together: T035 id.go · T036 name.go · T040 fuzz_test.go
Then sequentially: T037 → T038 → T039 → T041 → T042 → T043 → T044 → T045
```

### Parallel example: User Story 2

```text
Run together: T046 frontmatter_test.go (write path) · T047 encode_internal_test.go · T048 schema_test.go
Then: T049 → T050 → T051 → T052 → T053
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Design Preconditions (CRITICAL; blocks all stories; ADR 002/003 Accepted)
3. Complete Phase 3: User Story 1 (the read-only codec)
4. **STOP and VALIDATE**: Scenarios 1.1–1.5 pass under `-race`, and `ExampleDecode` prints the CORE §3 triples
5. Run Phase N compliance verification for the delivered scope

### Incremental Delivery

1. Setup + Preconditions → acceptance tests red, API contract agreed
2. US1 → read-only MVP
3. US2 → write; US3 → the round-trip guarantee (all three P1 stories together form the first releasable `v0.1.0`)
4. US4 + US5 → prose fidelity and hardening
5. US6 → merge-state passthrough
6. Each story adds value without breaking the previous stories or the public API

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Acceptance tests reference their scenario: `// Scenario X.Y from specs/001-markdown-triples-codec/spec.md`
- A scenario that cannot be tested through the public API is a missing API, not a reason to test internals
- Shell scripts MUST NOT be used to validate correctness
- Commit after each task or logical group
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
