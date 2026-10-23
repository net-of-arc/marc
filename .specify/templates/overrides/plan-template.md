# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]

**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.
It is the project override of the core plan template and enforces `.specify/memory/constitution.md`.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the feature. Values marked (fixed) are mandated by the constitution and MUST NOT change
  without an accepted ADR.
-->

**Language/Version**: Go [minimum supported version from `go.mod` `go` directive]; CI matrix of the minimum supported and latest stable Go releases (Principle XIII)

**Module**: single Go module `github.com/net-of-arc/marc`, root package `marc` (fixed, Principles III, XIII)

**Primary Dependencies**: standard library; `github.com/fogfish/faults` (fixed); any new direct dependency → ADR reference [ADR NNN or N/A] (Principle VII)

**Testing**: `go test -race ./...` with `github.com/fogfish/it/v2` (fixed); native fuzzing for decoders/parsers; benchmarks + `benchstat` for hot paths (Principle VI)

**External Interfaces**: [standard-library interfaces used, e.g., `io.Reader`, `io/fs.FS`, `iter.Seq`; or consumer-declared interface; or N/A] (Principle VII)

**Public API Surface**: [new/changed exported identifiers, `Config` fields with zero-value defaults, exported error constants] (Principle IX)

**Concurrency & Resources**: [synchronous / goroutine owner; concurrency safety of exported types; input size limits] (Principle XI)

**Performance Goals**: [hot paths named in ARCHITECTURE.md and their budget, or N/A] (Principles VI, XI)

**Compatibility**: [SemVer impact: PATCH / MINOR / MAJOR; expected `apidiff` result] (Principle XIV)

**Graduation Check**: [none of the ADR 001 graduation criteria met, or reference to the ADR deciding otherwise] (Principle III)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Read every ADR referenced below before planning; an accepted ADR overrides this plan (Principle I).

**Referenced ADRs**: [ADR 001 - Principles of System Architecture (Library Edition)](../../adrs/001-system-architecture.md), [others]

| # | Principle | Gate | Status |
|---|-----------|------|--------|
| I | Architecture Documentation & Decision Records | ARCHITECTURE.md update identified; major decisions have ADRs; no accepted ADR contradicted | [PASS/FAIL/N/A] |
| II | Domain-Driven Design & Glossary | New domain terms listed with Go identifiers; no duplicate type for an existing concept | [PASS/FAIL/N/A] |
| III | Flat Package Architecture | Root package `marc`; files by concept; sub-package only for an independent vocabulary; no `port/`, `adapter/`, `service/`, `app/`, `pkg/` | [PASS/FAIL/N/A] |
| IV | Functional Programming Style | No inline comments; functions ≤ 25 lines; no input mutation; accept interfaces, return structs | [PASS/FAIL/N/A] |
| V | Code Quality & Simplicity | No catch-all packages; every new exported identifier required by a spec scenario | [PASS/FAIL/N/A] |
| VI | TDD & Automated Testing | Tests first, red semantically; `-race`; fuzz tests for parsers/decoders; benchmarks for hot paths; `fogfish/it/v2` only | [PASS/FAIL/N/A] |
| VII | Dependencies & External Interfaces | Standard library first; no technology SDK; no `os` file functions; effects supplied by consumer | [PASS/FAIL/N/A] |
| VIII | Acceptance Testing & Spec Traceability | Each spec scenario → black-box test in `package marc_test`; testable examples per user story | [PASS/FAIL/N/A] |
| IX | Public API Design | `New(...)` + typed `Config` with zero-value defaults; no functional options; `context.Context` first; no panics across API; ownership & concurrency documented | [PASS/FAIL/N/A] |
| X | Library Citizenship | No `os.Exit`, stdout/stderr, env reads, mutable globals, side-effecting `init()`; `CGO_ENABLED=0` | [PASS/FAIL/N/A] |
| XI | Concurrency, Resources & Performance | Goroutines owned and leak-tested; `Close` idempotent; untrusted input bounded | [PASS/FAIL/N/A] |
| XII | Documentation & Error Contract | GoDoc for every package/exported identifier; errors via `fogfish/faults`; branchable errors exported constants | [PASS/FAIL/N/A] |
| XIII | Modules & Release Engineering | Single module; no `replace`; `go mod tidy` clean; `go` directive unchanged or MINOR note | [PASS/FAIL/N/A] |
| XIV | Versioning, Compatibility & Security | SemVer impact matches `apidiff`; deprecation before removal; input validated before allocation | [PASS/FAIL/N/A] |

Any FAIL MUST be resolved or justified in Complexity Tracking; FAIL on Principles I, III, VII, IX, or XIV
requires an accepted ADR before implementation (constitution: "When Constraints Cannot Be Met").

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command): exported Go API contract
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the tree below with the concrete files this feature adds or changes,
  following ADR 001. Delete OPTIONAL entries that the feature does not use.
-->

```text
./                             # module github.com/net-of-arc/marc, package marc
├── doc.go                     # package comment and quick start
├── marc.go                    # central domain type(s) and primary functions
├── marc_test.go               # unit tests (package marc)
├── [concept].go               # one concept per file
├── [concept]_test.go
├── config.go                  # Config struct, documented zero-value defaults, validation
├── errors.go                  # exported error constants (faults.Type)
├── [feature]_test.go          # black-box acceptance tests (package marc_test)
├── example_test.go            # testable examples
├── fuzz_test.go               # fuzz tests for parsers/decoders
├── bench_test.go              # benchmarks for hot paths
├── testdata/                  # fixtures and fuzz corpora (testdata/fuzz/)
├── [vocabulary]/              # OPTIONAL: independent public vocabulary
├── marctest/                  # OPTIONAL: public test helpers for consumers
└── internal/                  # OPTIONAL: helpers shared by the packages above
```

**Structure Decision**: [Document the files added or changed and confirm conformance with ADR 001]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Principle | Why Needed | Simpler Alternative Rejected Because | ADR |
|-----------|-----------|------------|-------------------------------------|-----|
| [e.g., new third-party dependency] | VII | [current need] | [why the standard library is insufficient] | [ADR NNN] |
