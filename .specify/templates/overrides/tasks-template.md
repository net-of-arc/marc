---

description: "Task list template for feature implementation (marc Library Constitution)"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are MANDATORY (Principles VI, VIII). Every spec scenario has a black-box acceptance test,
and unit tests are written first and fail semantically before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

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

<!--
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.

  The /speckit-tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements and Constitution Check from plan.md
  - Domain terms from data-model.md
  - Exported API from contracts/

  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment

  Phase 2 (2a–2e) and Phase N MUST be kept; adapt only their task descriptions.
  DO NOT keep sample tasks of other phases in the generated tasks.md file.
  ============================================================================
-->

## Phase 0: Pre-implementation Refactoring (OPTIONAL)

**Purpose**: Restructure existing code the feature depends on, delivered as a separate PR

- [ ] T000 [Refactoring task with exact file path]
- [ ] T00X Verify `go test -race ./...` passes and `apidiff -m -incompatible` reports no incompatible change

**Checkpoint**: Refactoring merged; behavior and public API unchanged

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Module and tooling initialization (only what the feature requires)

- [ ] T001 [Setup task, e.g., add `go.mod` entry for `github.com/fogfish/it/v2` test dependency]
- [ ] T002 [P] [Setup task, e.g., create `testdata/` fixtures for the feature]

---

## Phase 2: Design Preconditions (MANDATORY)

**Purpose**: Constitution PRECONDITIONS that MUST be complete BEFORE implementation begins

**⚠️ CRITICAL**: No user story implementation can begin until this phase is complete

### Phase 2a: Domain Vocabulary & Glossary (Principle II)

- [ ] T003 Identify domain terms for [feature] and add them with definitions and Go identifiers to the Glossary in ARCHITECTURE.md
- [ ] T004 Verify no existing type already models [concept]; reuse instead of duplicating

### Phase 2b: Public API Contract Design (Principle IX)

- [ ] T005 Define exported types and functions for [feature] in specs/[###-feature]/contracts/ (no stutter, minimal surface)
- [ ] T006 Define `Config` fields with documented zero-value defaults and validation errors (no functional options)
- [ ] T007 Define exported error constants (`faults.Type`) consumers branch on, in errors.go contract
- [ ] T008 Document ownership of arguments/results and concurrency safety for each exported type

### Phase 2c: External Interface Design (Principle VII, if applicable)

- [ ] T009 Select the standard-library interface for each external effect (`io.Reader`, `io/fs.FS`, `iter.Seq`, …) or declare a small consumer-implementable interface next to its consuming function
- [ ] T010 Confirm no new dependency, technology SDK, or `os` file function; otherwise draft ADR in adrs/

### Phase 2d: Acceptance Test Design (Principle VIII)

- [ ] T011 [P] Write black-box acceptance test for Scenario 1.1 in [feature]_test.go (package marc_test), using `fogfish/it/v2` with concrete expected values
- [ ] T012 [P] Write black-box acceptance test for Scenario 1.2 in [feature]_test.go
- [ ] T013 Add minimal stubs so acceptance tests compile; confirm `go test ./...` fails semantically (no `t.Skip()`, no "red phase" comments)

### Phase 2e: Dependency, Compatibility & Graduation Review (Principles III, XIII, XIV)

- [ ] T014 Check ADR 001 graduation criteria; record an ADR if any criterion is met
- [ ] T015 Classify SemVer impact (PATCH/MINOR/MAJOR) of the planned API and `Config` changes
- [ ] T016 Confirm `go` directive, single module, and no `replace` directives remain unchanged

**Checkpoint**: Preconditions complete — acceptance tests are red, API contract agreed, user story implementation can begin

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [Acceptance tests for Scenarios 1.x pass through the public API]

### Unit Tests for User Story 1 (write FIRST, MUST FAIL) ⚠️

- [ ] T017 [P] [US1] Table-driven unit tests for [concept] in [concept]_test.go (happy path, errors, nil/zero/empty, boundaries)
- [ ] T018 [P] [US1] Fuzz test for [parser/decoder] in fuzz_test.go with seed corpus in testdata/fuzz/ (if it parses untrusted bytes)

### Implementation for User Story 1

- [ ] T019 [P] [US1] Implement [domain type] in [concept].go
- [ ] T020 [US1] Implement [function] in [concept].go (depends on T019)
- [ ] T021 [US1] Implement `Config` validation in config.go returning `ErrInvalidConfig` naming the field
- [ ] T022 [US1] Add GoDoc for new exported identifiers and `ExampleX` in example_test.go

**Checkpoint**: User Story 1 acceptance tests pass under `-race`

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [Acceptance tests for Scenarios 2.x pass through the public API]

### Unit Tests for User Story 2 (write FIRST, MUST FAIL) ⚠️

- [ ] T023 [P] [US2] Unit tests for [concept] in [concept]_test.go

### Implementation for User Story 2

- [ ] T024 [US2] Implement [function] in [concept].go
- [ ] T025 [US2] Add GoDoc and `ExampleX` in example_test.go

**Checkpoint**: User Stories 1 AND 2 both pass independently

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N-1: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] TXXX [P] Update README.md (purpose, `go get`, quick start, supported Go versions) and doc.go
- [ ] TXXX [P] Benchmarks for hot paths in bench_test.go with `benchstat` comparison (if ARCHITECTURE.md names hot paths)
- [ ] TXXX Refactor for readability: functions ≤ 25 lines, no inline comments, no input mutation

---

## Phase N: Constitution Compliance Verification (MANDATORY)

**Purpose**: Constitution Implementation Phase checklist, verified per principle before the feature is complete

- [ ] TXXX Principle I: ARCHITECTURE.md and ADRs reflect design changes; no accepted ADR contradicted
- [ ] TXXX Principle II: Glossary terms match exported identifiers; one concept, one name
- [ ] TXXX Principle III: Flat root package per ADR 001; sub-packages only for independent vocabularies
- [ ] TXXX Principle IV: No inline/block comments except GoDoc; functions ≤ 25 lines; inputs not mutated; accept interfaces, return structs
- [ ] TXXX Principle V: No catch-all packages; every exported identifier required by a spec scenario
- [ ] TXXX Principle VI: Unit tests written first and changed minimally; `go test -race ./...` passes; parsers fuzzed; hot paths benchmarked; `fogfish/it/v2` exclusively
- [ ] TXXX Principle VII: No unjustified dependency, no technology SDK, no `os` file functions
- [ ] TXXX Principle VIII: Every spec scenario has a passing black-box acceptance test that changed minimally
- [ ] TXXX Principle IX: Naming, construction, typed `Config` with zero-value defaults, context-first, error, and ownership rules; no functional options
- [ ] TXXX Principle X: No `os.Exit`, stdout/stderr writes, environment reads, mutable globals, or side-effecting `init()`; builds with `CGO_ENABLED=0`
- [ ] TXXX Principle XI: Goroutines (if any) owned and leak-free; resources bounded; `Close` idempotent
- [ ] TXXX Principle XII: GoDoc for every package and exported identifier; examples for entry points; errors via `fogfish/faults`, matched with `errors.Is`
- [ ] TXXX Principle XIII: `go mod tidy` leaves `go.mod`/`go.sum` unchanged; no `replace`, `vendor/`, or `go.work`
- [ ] TXXX Principle XIV: `apidiff -m -incompatible` matches the SemVer impact; `gofmt -l`, `go vet`, `staticcheck` clean

---

## Dependencies & Execution Order

### Phase Dependencies

- **Refactoring (Phase 0)**: Optional; separate PR merged before Phase 1
- **Setup (Phase 1)**: No dependencies - can start immediately
- **Design Preconditions (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Phase 2 completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase N-1)**: Depends on all desired user stories being complete
- **Compliance Verification (Phase N)**: Final gate; feature is incomplete until every item passes

### Within Each User Story

- Unit tests MUST be written and FAIL semantically before implementation
- Domain types before functions that use them
- GoDoc and examples in the same story as the API they document
- Story complete (acceptance tests green under `-race`) before moving to next priority

### Parallel Opportunities

- Tasks marked [P] touch different files and can run in parallel
- Acceptance tests for different scenarios in Phase 2d can be written in parallel
- Once Phase 2 completes, user stories can start in parallel

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Design Preconditions (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: User Story 1 acceptance tests pass under `-race`
5. Run Phase N compliance verification for the delivered scope

### Incremental Delivery

1. Setup + Design Preconditions → acceptance tests red, API contract agreed
2. Add User Story 1 → acceptance tests green → MVP
3. Add User Story 2 → acceptance tests green
4. Each story adds value without breaking previous stories or the public API

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Acceptance tests reference their scenario: `// Scenario X.Y from specs/[###-feature]/spec.md`
- A scenario that cannot be tested through the public API is a missing API, not a reason to test internals
- Shell scripts MUST NOT be used to validate correctness
- Commit after each task or logical group
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
