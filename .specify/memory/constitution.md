<!--
Sync Impact Report
==================
Version change: (unversioned user draft) → 1.0.0 (initial adoption for marc)

Bump rationale: first adoption of the Library-edition constitution in this repository; no
prior ratified version of this project's constitution exists, so 1.0.0 is retained.

Modified principles: none (all fourteen principles I–XIV adopted verbatim as authored)

Non-semantic edits applied during adoption:
- Title placeholder `{Name}` resolved to `marc`; "the library" now bound to `marc`
- Relative links normalized to resolve from .specify/memory/ (ARCHITECTURE.md, adrs/,
  .specify/templates/, constitution.md); tasks-template link points to the project override
- Last Amended date set to the adoption date

Added sections: none
Removed sections: none

Templates requiring updates:
- ✅ .specify/templates/overrides/plan-template.md — new project override: Go library
  Technical Context, Constitution Check gates for Principles I–XIV, flat package layout (ADR 001)
- ✅ .specify/templates/overrides/tasks-template.md — new project override: mandatory
  Phase 2 (2a–2e) Design Preconditions and Phase N Constitution Compliance Verification,
  tests mandatory, Go paths
- ✅ .specify/templates/spec-template.md — compatible (Given/When/Then scenarios map 1:1 to
  acceptance tests); no override needed
- ✅ .specify/templates/checklist-template.md — compatible; no override needed
- ✅ .claude/skills/speckit-* — load constitution.md at runtime; no edits needed
- ✅ adrs/001-system-architecture.md — consistent with Principles III, VII, IX, XIII

Follow-up TODOs:
- ARCHITECTURE.md does not exist yet (Principle I requires glossary, package list, NFRs)
- go.mod, README.md content, and .github/workflows (build, check-code, check-test) are not
  yet present (Principles XII, XIII, Mandatory Libraries & Tooling)
-->

# marc Library Constitution

This constitution governs the design and engineering of a small, focused Go
library distributed as a single Go module and imported by other Go programs.
In this document, "the library" means `marc`, a semantic Markdown graph parser.

A library, in the sense of this document, solves one problem with one
vocabulary: a data structure, a codec, an algorithm, a protocol helper, a
thin client of a single service. It has no independent processes to
orchestrate and no interchangeable technologies underneath. For such code,
hexagonal architecture is overhead: the library is a flat package whose
exported API is the whole product. Libraries that outgrow this shape
graduate to the companion *capability* constitution ([ADR 001](../../adrs/001-system-architecture.md), "Graduation").

This constitution treats [Effective Go](https://go.dev/doc/effective_go),
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and the
[Go Modules Reference](https://go.dev/ref/mod) as authoritative for idiom and
module mechanics. Where this document is silent, those govern; where this
document is explicit, it MUST be followed.

The [Mandatory Libraries & Tooling](#mandatory-libraries--tooling) section
fixes the concrete toolchain. These are explicit, non-substitutable choices.

## Core Principles

### I. Architecture Documentation & Decision Records

Design and major decisions MUST be documented, and Architecture Decision Records are BINDING.

**Rules**:
- [ARCHITECTURE.md](../../ARCHITECTURE.md) is the single source of truth for the library's design; for a small library it MAY be short, but it MUST contain the glossary (Principle II), the package list, and the non-functional requirements (performance budget, concurrency guarantees, supported Go versions)
- When a feature changes the design, [ARCHITECTURE.md](../../ARCHITECTURE.md) MUST be updated in the same PR
- Major decisions (data format, algorithm choice, API shape, a new dependency) MUST be recorded in [adrs/](../../adrs/) as `NNN-decision-title.md` with Context, Decision, Consequences, Status (Proposed/Accepted/Deprecated/Superseded)
- **ADRs with status "Accepted" are BINDING**; deviation requires a new superseding ADR
- **Before implementing, read every ADR referenced in the plan's Constitution Check section**; when a plan contradicts an accepted ADR, the ADR takes precedence and the plan MUST be corrected
- **Resolving design tension by ignoring an ADR is FORBIDDEN**: raise the conflict and resolve it via a new ADR

**Rationale**: Even a small library accumulates decisions that are expensive to rediscover. Binding ADRs keep AI agents and contributors from silently re-deciding them.

### II. Domain-Driven Design & Glossary Management

The library's vocabulary MUST be explicit and consistent, because its exported identifiers *are* its domain model.

**Rules**:
- All domain terms MUST be listed in the Glossary section of [ARCHITECTURE.md](../../ARCHITECTURE.md) with definition and the Go identifier that implements it
- Ubiquitous language MUST be used consistently in package names, exported identifiers, GoDoc, and documentation — one concept, one name
- **Before defining a new type, check for an existing one** in the library; duplicate types for the same concept are forbidden
- Domain types MUST make illegal states unrepresentable where Go allows it (distinct named types, unexported fields with validated constructors, typed constants)

**Rationale**: Consumers learn a library through its exported identifiers and GoDoc; one name per concept is what makes it learnable.

### III. Flat Package Architecture

The library MUST use the simplest package structure that keeps its API coherent. [ADR 001](../../adrs/001-system-architecture.md) defines mandatory and binding principles of the library's structure.

**Rules**:
- **The root package is the library**: it is named after the module and holds the primary types and functions consumers import
- Files within a package are organized by concept (`codec.go`, `cursor.go`, `config.go`, `errors.go`), never by kind (`types.go`, `interfaces.go`, `utils.go`)
- **A public sub-package is introduced only for an independent vocabulary** that some consumers use without the rest (e.g., an optional encoding); it MUST NOT be introduced to group files
- Helpers shared by packages but not part of the API live under `internal/`
- There are no `port/`, `adapter/`, `service/`, `app/`, or `pkg/` directories; the hexagonal vocabulary is not used
- Logic MUST remain unit-testable without network, credentials, or real filesystem state: external effects reach the library through standard-library interfaces or small consumer-declared interfaces (Principle VII)
- **Graduation**: when the library meets the graduation criteria of ADR 001, the maintainers MUST decide, via an ADR, whether to adopt the capability constitution before adding the feature that triggered them

**Rationale**: A flat package keeps a small library easy to read, easy to document, and easy to import. Structure is added only when it carries information.

### IV. Functional Programming Style

Functional Programming Style is NON-NEGOTIABLE. All Go code MUST follow a functional-first design philosophy emphasizing clarity through structure rather than commentary.

**Rules**:
- **No Inline Comments**: Code MUST NOT include inline or block comments. The only exception is GoDoc comments, which are MANDATORY for every exported identifier and every package (Principle XII)
- **Self-Explanatory Naming**: All identifiers MUST be descriptive and unambiguous; avoid abbreviations unless universally understood
- **Small, Focused Functions**: Functions MUST be short (not more than 25 lines) and focused on a single responsibility
- **Composition Over Complexity**: Favor composing small functions; prefer passing data through transformations rather than mutating shared state
- **Immutability by Default**: Functions MUST avoid mutating inputs; prefer returning new values; isolate side effects at the boundary
- **Readability as a Requirement**: If naming and structure are insufficient to convey intent, the code must be rewritten rather than annotated
- **Segregation**: Go code MUST accept interfaces, return structs
- Where a measured hot path (Principle XI) requires in-place mutation or buffer reuse, it MUST be confined to an unexported function whose exported caller remains free of observable side effects

**Rationale**: Pure functions are testable without mocks, composable, and predictable under concurrency — properties that matter most in code embedded into processes its authors do not control.

### V. Code Quality & Simplicity

All contributors MUST follow SOLID principles and implement only what the current feature requires (YAGNI).

**Rules**:
- **Single Responsibility**: Functions, types, and packages MUST exhibit natural cohesion; catch-all packages (`utils`, `common`, `helpers`, `types`) are forbidden
- **Open/Closed**: Compose simple types via embedding; embedding in an exported struct exports the embedded type's methods — do it deliberately
- **Liskov Substitution / Interface Segregation**: depend on the narrowest interface that satisfies the need; prefer single-method interfaces
- **Dependency Inversion**: external effects are injected by the consumer, never instantiated inside the library
- **YAGNI for public API**: an identifier is exported only when a current spec scenario requires a consumer to use it. Unexporting later is a breaking change; exporting later is not

**Rationale**: Every exported identifier is a permanent promise; a small library stays small only if each promise is earned.

### VI. Test-Driven Development & Automated Testing

Code quality and correctness MUST be ensured through Test-Driven Development (TDD) and automated tests, not manual validation.

**Rules**:
- **Tests MUST be written FIRST and MUST FAIL before implementation begins** (red-green-refactor cycle)
- **"Starting red" means tests MUST compile AND fail semantically**: minimal stubs to make tests compile are ACCEPTABLE; skipping tests, marking them pending, or adding "red phase" comments (e.g., `// TODO: implement`) is FORBIDDEN
- Tests MUST drive design and MUST change as little as possible during implementation
- Automated tests MUST cover happy paths, error cases, and edge cases (nil/zero values, empty inputs, boundary sizes, canceled contexts, concurrent use)
- **`github.com/fogfish/it/v2` MUST be the sole assertion library**, used via its fluent `it.Then(t).Should(...)` style
- Table-driven tests MUST be used for parameterized scenarios
- **The full test suite MUST pass under the race detector** (`go test -race ./...`) in CI
- **Every function that parses, decodes, or deserializes untrusted bytes MUST have a native Go fuzz test**, with its seed corpus under `testdata/fuzz/`
- **Every hot path listed in [ARCHITECTURE.md](../../ARCHITECTURE.md) MUST have a benchmark** reporting allocations; PRs touching it MUST include a `benchstat` comparison
- **Bash/shell scripts MUST NOT be used to validate code correctness**
- Test coverage MUST be verifiable via `go test ./... -cover`

**Rationale**: TDD ensures tests verify requirements rather than retrofit implementation. The race detector, fuzzing, and benchmarks verify what a consumer cannot verify from outside.

### VII. Dependencies & External Interfaces

The library MUST be cheap to import: few dependencies, no hidden effects.

**Rules**:
- **Standard library first**: zero third-party runtime dependencies beyond [Mandatory Libraries & Tooling](#mandatory-libraries--tooling) is the target. **Every new direct dependency requires an ADR** stating why the standard library is insufficient, its license, maintenance status, and transitive footprint
- **The library MUST NOT depend on a technology SDK** (cloud provider, database driver, message broker). A library that needs one has become a capability or needs a nested adapter module — both are graduation signals (ADR 001)
- **External effects are expressed with standard-library interfaces** — `io.Reader`, `io.Writer`, `io/fs.FS`, `net/http.RoundTripper`/`*http.Client`, `context.Context` — supplied by the consumer
- Where no standard interface fits, the library declares a **small, consumer-implementable interface next to the function that needs it**, over library and standard-library types only; it is exported only because consumers must implement it
- The `os` package's file/directory functions MUST NOT be used; filesystem access goes through `io/fs` and caller-supplied writers
- Where deterministic behavior matters, time and randomness are injectable (a `Config` field accepting a clock or `rand.Source`); otherwise the library MUST NOT depend on wall-clock time
- Network calls, if any, MUST accept `context.Context`, use the consumer's `*http.Client` (never `http.DefaultClient`), and impose no hidden timeouts
- Credentials, if any, are passed in by the consumer; the library MUST NOT read them from the environment or log them

**Rationale**: For a small library, its dependency list is most of its cost to consumers. Standard-library interfaces make it composable with everything else in Go without introducing a single abstraction of its own.

### VIII. Acceptance Testing & Spec Traceability

All features MUST have acceptance tests where each scenario in `spec.md` maps 1:1 to a test, exercised through the public API.

**Rules**:
- **Every acceptance scenario in `specs/[NNN-feature-name]/spec.md` MUST have a corresponding acceptance test**, written BEFORE implementation begins
- **Acceptance tests MUST be black-box**: they live in the external test package (`package <name>_test`), colocated with the package they exercise, named after the feature (e.g., `decode_test.go`), and use exported identifiers only. A scenario a test cannot express through the public API is a missing API, not a reason to test internals
- External effects are supplied as in-memory values (`bytes.Buffer`, `strings.Reader`, `fstest.MapFS`, `httptest.Server`); acceptance tests MUST NOT require network access or credentials
- Test fixtures live in a colocated `testdata/` directory
- Acceptance tests MUST compile AND fail semantically initially, use `github.com/fogfish/it/v2` with concrete expected values, and MUST NOT contain `t.Skip()` or "red phase" comments
- Each test function SHOULD reference its scenario: `// Scenario X.Y from specs/[NNN-feature]/spec.md`
- **Every user story SHOULD additionally have a testable example** (`func ExampleDecode()` with an `// Output:` block) in `example_test.go`
- Acceptance tests MUST change minimally during implementation, MUST be independent of execution order, and SHOULD call `t.Parallel()`

**Rationale**: Black-box tests prove the library works for its consumer and leave internals free to change; examples make the same proof visible on pkg.go.dev.

### IX. Public API Design

The exported surface of the library is its product. It MUST be minimal, idiomatic, consistent, and hard to misuse.

**Rules**:
- **Minimal surface**: anything not needed by a consumer is unexported or under `internal/`
- **Naming**: package names are short, lower-case, singular nouns; exported identifiers MUST NOT stutter (`codec.New`, not `codec.NewCodec`)
- **Construction**: types that need setup are created by `New(...)` returning a concrete type (and an `error` if construction can fail); types that need no setup MUST have a useful zero value
- **Configuration is explicit and typed**:
  - **Required dependencies are positional parameters** of the constructor (e.g., `New(storage {port}, cfg Config) (*{Type}, error)`), so omitting one is a compile error
  - **Optional settings are fields of an exported `Config` struct** (`XConfig` when a package configures several types), passed by value; the constructor copies it, so later mutation by the caller has no effect
  - **The zero value of every field means its documented default**, and each field's GoDoc states that default; a zero `Config{}` MUST yield a working instance. When zero is a meaningful value distinct from the default, the field's GoDoc names an explicit sentinel (e.g., a negative duration disables a timeout); pointer fields used to signal "unset" are forbidden
  - The constructor **validates the whole `Config`** and returns an exported `faults` error (e.g., `ErrInvalidConfig`) naming the offending field; it MUST NOT silently clamp or ignore invalid values
  - Configuration is **immutable after construction**; reconfiguration requires a new instance or an explicit, documented method
  - `Config` MAY carry `json`/`yaml` struct tags so the consumer can load it from its own sources; the library MUST NOT load it itself (Principle X)
  - A **builder** MAY produce a `Config` when it has interdependent or conditionally required fields that a struct literal cannot express safely: each builder step returns a new value (Principle IV) and the terminal `Build() (Config, error)` validates. The constructor still accepts the `Config`; the builder is a convenience over it, never a replacement. The same builder shape applies to complex domain values (queries, requests, schemas)
  - **Functional options (`type Option func(*config)`, `WithX(...) Option`) are FORBIDDEN** in exported APIs, as are positional boolean parameters
- **`context.Context` is the first parameter** of every exported function that performs I/O, blocks, or may be canceled; contexts MUST NOT be stored in structs
- **Errors are values**: exported functions that can fail return `error` last and MUST NOT panic on any input a consumer can construct. Panics are reserved for programmer errors in unexported code and MUST NOT cross the public API (a `MustX` variant is allowed only alongside the error-returning `X`, for package-level initialization of constant inputs)
- **Interfaces are exported only when consumers implement them**; a single implementation is returned as a struct
- **Generics** are used when they remove duplication across types the consumer chooses, with type parameters constrained as narrowly as the algorithm allows
- **Sequences**: APIs yielding a potentially large or unbounded sequence SHOULD return an iterator (`iter.Seq`/`iter.Seq2`)
- **Ownership and concurrency**: GoDoc MUST state whether the library retains slices/maps passed in or returned, and whether each exported type is safe for concurrent use
- Exported identifiers MUST NOT expose types from `internal/` or from a third-party module that is not a deliberate part of the contract

**Rationale**: Consumers integrate against the API once and live with it for years.

### X. Library Citizenship & Side-Effect Hygiene

The library runs in a process it does not own. It MUST behave as a guest.

**Rules**:
- The library MUST NOT call `os.Exit`, `log.Fatal*`, or `log.Panic*`, and MUST NOT write to `os.Stdout` or `os.Stderr`
- The library MUST NOT read environment variables, command-line arguments, or flags; all configuration comes from the consumer
- **No mutable package-level state**: package-level variables are limited to constants, sentinel errors, and values immutable after initialization; `init()` MUST NOT perform I/O, start goroutines, or register global side effects
- The library MUST NOT modify process-wide settings (`http.DefaultClient`, `http.DefaultTransport`, the default `slog` logger, `GOMAXPROCS`, signal handlers)
- **Logging is opt-in and usually unnecessary**: a small library reports through return values and errors. If diagnostics are genuinely needed, they go through a `*slog.Logger` supplied as a `Config` field, where nil means a discarding logger
- The library MUST NOT collect telemetry or contact any network endpoint the consumer did not configure
- Secrets and payload data MUST NEVER appear in log records or error messages
- Cgo and `unsafe` MUST NOT be used without an accepted ADR; the library MUST build with `CGO_ENABLED=0`

**Rationale**: A library that exits, prints, reads the environment, or mutates globals makes its consumer's process unpredictable.

### XI. Concurrency, Resource Lifecycle & Performance

**Rules**:
- **Synchronous by default**: exported functions do their work and return. A library that starts goroutines MUST give each an owner (`Close` or a consumer-supplied context) and MUST verify in tests that none leaks
- Types that hold resources MUST implement `io.Closer`; `Close` MUST be idempotent
- Channels in exported signatures MUST be directional, with GoDoc stating who closes them
- Shared state MUST be protected by `sync` primitives or confined to one goroutine
- **Resource bounds are explicit**: buffers, caches, and reads of untrusted input (`io.LimitReader`) MUST have documented, consumer-configurable limits
- Hot paths, if any, are named in [ARCHITECTURE.md](../../ARCHITECTURE.md) with their budget and benchmarked (Principle VI); readability-compromising optimizations MUST be justified by a `benchstat` result
- `faults.Fast` MAY replace `faults.Type`/`faults.SafeN` on a documented hot path where a benchmark shows the error path matters

**Rationale**: Leaked goroutines, unbounded buffers, and undocumented thread-safety are defects consumers cannot fix themselves.

### XII. Documentation & Error Contract

The library MUST be self-documenting on pkg.go.dev, and its errors MUST be stable and matchable.

**Rules**:
- **Every package MUST have a package comment** (in `doc.go` once it exceeds a few lines) with purpose and a minimal usage snippet
- **Every exported identifier MUST have a GoDoc comment** starting with its name, stating behavior, errors returned, concurrency safety, and ownership of arguments
- Testable examples MUST exist for the library's primary entry points
- A `README.md` MUST document purpose, installation (`go get`), a quick-start example, supported Go versions, and a link to pkg.go.dev
- Documentation MUST be updated in the same PR as the API change it describes
- **Error context annotation MUST use `github.com/fogfish/faults`**, not ad hoc `fmt.Errorf("...: %w", err)` chains
- **Errors a consumer is expected to branch on are exported constants** (`const ErrMalformed = faults.Type("malformed input")`), documented, and part of the compatibility contract; all other errors are unexported
- Consumers match errors with `errors.Is`/`errors.As` only; GoDoc MUST NOT invite matching on error strings

**Rationale**: For a library, pkg.go.dev is the help system and the error set is part of the API.

### XIII. Modules & Release Engineering

**Rules**:
- The library is **a single Go module** whose path is the canonical import path of the repository; major versions ≥ 2 use the `/vN` suffix (Principle XIV). Needing a second module is a graduation signal (ADR 001)
- **The `go` directive declares the minimum supported Go version** and MUST NOT be raised without a MINOR release note; the library MUST support at least the two most recent Go releases
- Released `go.mod` MUST NOT contain `replace` directives; `go mod tidy` MUST leave `go.mod`/`go.sum` unchanged in CI; `vendor/` and `go.work` MUST NOT be committed
- Releases are Git tags with GitHub Releases generated in CI; tags MUST be created by CI from a documented increment rule (e.g., a marker in the merge commit message selecting major/minor/patch, defaulting to patch), never pushed manually
- Generated release notes MUST exclude non-consumer-facing commit categories (`^docs:`, `^test:`, `^ci:`)

**Rationale**: A clean single module with CI-driven tags makes every release reproducible.

### XIV. Versioning, Compatibility & Security

**Rules**:
- Versioning MUST follow **Semantic Versioning** applied to the exported Go API, including exported errors, documented behavior, and concurrency guarantees
- **Import compatibility rule**: breaking changes require a new major version with a new module path (`/v2`)
- **`Config` compatibility**: adding a field to an exported `Config` is MINOR; removing or renaming a field, changing its type, or changing the meaning of its zero value is MAJOR. `apidiff` does not detect a changed default, so reviewers MUST check it explicitly
- **`apidiff` MUST run on every PR** against the latest release tag; an incompatible change reported for a non-major release blocks merge
- `v0.x` versions MAY break between minor versions but MUST announce breaks in release notes
- **Deprecation precedes removal**: a `// Deprecated:` GoDoc paragraph naming the replacement for at least one minor release before the next major version
- A release published in error MUST be withdrawn with a `retract` directive, never by deleting or moving a tag
- **`govulncheck` MUST pass in CI before every release**
- **Untrusted input is hostile**: decoders and parsers MUST validate lengths and counts before allocating, bound recursion depth, and be fuzzed (Principle VI)
- Cryptography, if any, MUST use the standard library or `golang.org/x/crypto`

**Rationale**: Consumers upgrade a library because its version number tells them it is safe to; the API-diff gate makes that a checked property.

## Mandatory Libraries & Tooling

These libraries and tools are mandated for every project that adopts this constitution, identical to the capability constitution so that contributors move between libraries and capabilities with no change of idiom.

**Error Handling & Annotation**
- `github.com/fogfish/faults` MUST be used for error annotation (Principle XII). Expected error conditions are declared once as package-level constants using `faults.Type` or `faults.Safe1`–`faults.Safe5`; wrapping uses `.With(err[, args...])`; matching uses `errors.Is`. Ad hoc `fmt.Errorf("<context>: %w", err)` is forbidden. `faults.Fast` only on benchmarked hot paths (Principle XI).

**Logging**
- Standard library `log/slog`, supplied as a `Config` field where nil means `slog.New(slog.DiscardHandler)`, if the library logs at all (Principle X). No third-party logging library.

**Filesystem**
- Standard library `io/fs`, `io.Reader`, `io.Writer` only (Principle VII); paths follow `fs.ValidPath`. Tests use `testing/fstest.MapFS`.

**Testing**
- `github.com/fogfish/it/v2` is the sole assertion library (Principles VI, VIII).
- `go test -race -coverprofile=profile.cov ./...` MUST be runnable from CI; coverage MUST be published to a coverage-reporting service (e.g., Coveralls via `shogo82148/actions-goveralls`).
- Native Go fuzzing (`testing.F`) for decoders/parsers; `golang.org/x/perf/cmd/benchstat` for hot-path PRs.

**Static Analysis & Compatibility**
- `gofmt -l` (no output), `go vet`, and `staticcheck` MUST run on every pull request and block merge on failure.
- `apidiff` (`golang.org/x/exp/cmd/apidiff`) MUST run on every pull request in module mode with incompatible-only reporting (`apidiff -m -incompatible`), comparing the PR head against the latest release tag (Principle XIV). The `apidiff` version used in CI MUST be pinned.
- `govulncheck` (`golang.org/x/vuln/cmd/govulncheck`) MUST run before every release (Principle XIV).

**Continuous Integration (GitHub Actions)**
- **`build`** — on push to the default branch: build, `go test -race` with coverage, automatic SemVer tag (Principle XIII), GitHub Release
- **`check-code`** — on pull request: `gofmt`, `go vet`, `staticcheck`, `go mod tidy` diff, `apidiff`; required before merge
- **`check-test`** — on pull request: build and `go test -race` with coverage on a matrix of the minimum supported and latest stable Go versions; required before merge
- Go toolchain versions in CI MUST be pinned explicitly

**Rationale**: One toolchain for every library and capability under this governance.

## Development Requirements

### Compliance Checklist

**PRECONDITIONS (must complete BEFORE implementation begins)**:

- [ ] Domain terms identified and added to the [ARCHITECTURE.md](../../ARCHITECTURE.md) Glossary (Principle II)
- [ ] Public API contract designed: exported types, functions, `Config` fields with defaults, and error constants agreed before coding (Principle IX)
- [ ] External interfaces identified (standard-library interface or consumer-declared interface); no new dependency without an ADR (Principle VII)
- [ ] Graduation criteria checked; no criterion met, or an ADR decides otherwise (Principle III, ADR 001)
- [ ] **Acceptance tests written, black-box, for all spec scenarios; they compile and fail semantically (Principle VIII)**

**Implementation Phase**:

- [ ] [ARCHITECTURE.md](../../ARCHITECTURE.md) and ADRs reflect design changes (Principle I)
- [ ] Package structure follows ADR 001: flat root package, sub-packages only for independent vocabularies (Principle III)
- [ ] **Unit tests written FIRST, compile, fail semantically, and change minimally during implementation (Principle VI)**
- [ ] Tests pass under `-race`; decoders/parsers fuzzed; hot paths benchmarked (Principle VI)
- [ ] Tests use `github.com/fogfish/it/v2` exclusively (Principle VI)
- [ ] No new unjustified dependency, no technology SDK, no `os` file functions (Principle VII)
- [ ] Public API follows naming, construction, `Config`, context, error, and ownership rules; no functional options (Principle IX)
- [ ] No `os.Exit`, stdout/stderr writes, environment reads, mutable globals, or side-effecting `init()` (Principle X)
- [ ] Goroutines (if any) owned and leak-free; resources bounded; `Close` idempotent (Principle XI)
- [ ] GoDoc for every package and exported identifier; examples for primary entry points (Principle XII)
- [ ] Errors annotated via `github.com/fogfish/faults`; branchable errors exported and matched with `errors.Is` (Principle XII)
- [ ] `go mod tidy` clean, no `replace`; `apidiff` reports no incompatible change for a non-major release (Principles XIII, XIV)
- [ ] **All spec scenarios have passing acceptance tests that changed minimally (Principle VIII)**

### When Constraints Cannot Be Met

If **Principle I (ADRs)**, **Principle III (Flat Package Architecture)**, or **Principle VII (Dependencies & External Interfaces)** requires deviation:

1. STOP implementation immediately
2. Document why the rule cannot be followed
3. Draft a new ADR proposing the alternative — or proposing graduation to the capability constitution
4. Do NOT implement the deviation without an accepted ADR

If **Principle VI (TDD)** or **Principle VIII (Acceptance Testing)** cannot be satisfied:

1. Document the specific reason, treating an untestable scenario as a missing API first
2. Propose an alternative validation approach
3. Escalate to project maintainers for exception approval
4. Do NOT ship without tests mapped to spec scenarios

If **Principle IX (Public API Design)** or **Principle XIV (Versioning, Compatibility & Security)** cannot be satisfied:

1. STOP implementation immediately
2. Document the rule and the reason
3. Escalate to project maintainers and record the exception in an ADR
4. Do NOT release an incompatible API change under the same major version

## Governance

### Amendment Procedure

1. Propose amendment via PR to [.specify/memory/constitution.md](constitution.md)
2. Include rationale, impacted templates, and migration plan
3. Increment CONSTITUTION_VERSION per semantic versioning rules
4. Update the Sync Impact Report with changes
5. Propagate changes to dependent templates and documentation

### Versioning Policy

- **MAJOR**: Backward-incompatible governance/principle removals or redefinitions
- **MINOR**: New principle/section added or materially expanded guidance
- **PATCH**: Clarifications, wording, typo fixes, non-semantic refinements

### Compliance Review

- All PRs MUST verify compliance with this constitution; reviewers MUST challenge complexity and request justification when principles are violated
- Reviewers MUST verify adherence to accepted ADRs and the flat package structure, and check graduation criteria (Principles I, III)
- Reviewers MUST verify TDD, `-race`, fuzzing, and benchmark rules (Principle VI)
- Reviewers MUST verify every new dependency has an ADR and no technology SDK was introduced (Principle VII)
- Reviewers MUST verify black-box acceptance tests exist for all spec scenarios (Principle VIII)
- Reviewers MUST verify every new exported identifier is required and follows the API rules, including typed `Config` with documented zero-value defaults and no functional options (Principles V, IX)
- Reviewers MUST verify absence of process-level side effects and globals (Principle X)
- Reviewers MUST verify GoDoc, examples, and `fogfish/faults`-based error contract (Principle XII)
- Reviewers MUST verify `go.mod` hygiene and `apidiff` output (Principles XIII, XIV)
- Template files in [.specify/templates/](../templates/) provide execution workflows that enforce these principles

### Task List Requirements

Every feature's `tasks.md` file MUST include these mandatory sections from [tasks-template.md](../templates/overrides/tasks-template.md):

**MANDATORY SECTIONS** (cannot be omitted):
1. **Phase 2: Design Preconditions** - Constitution PRECONDITIONS implementation
   - Phase 2a: Domain Vocabulary & Glossary (Principle II)
   - Phase 2b: Public API Contract Design (Principle IX)
   - Phase 2c: External Interface Design (Principle VII, if applicable)
   - Phase 2d: Acceptance Test Design (Principle VIII)
   - Phase 2e: Dependency, Compatibility & Graduation Review (Principles III, XIII, XIV)

2. **Phase N: Constitution Compliance Verification** - Constitution Implementation Phase checklist, grouped by principle with explicit principle references

**OPTIONAL PHASES**:
- **Phase 0: Pre-implementation Refactoring** — a separate PR when the feature requires restructuring existing code; all tests MUST pass and `apidiff` MUST report no incompatible change afterwards

**CUSTOMIZABLE SECTIONS**: Phase 1 (Setup), Phase 3+ (User Stories from spec.md), Polish

**Enforcement**:
- AI agents generating tasks.md MUST retain Phase 2 and Phase N sections verbatim (adapting only task descriptions to the specific feature)
- Omitting mandatory sections violates this constitution and blocks feature completion

**Version**: 1.0.0 | **Ratified**: 2026-10-18 | **Last Amended**: 2026-10-23
