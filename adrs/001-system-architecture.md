# ADR 001 - Principles of System Architecture (Library Edition)

**Status**: Accepted
**Date**: 2026-10-18

## Context

We are aiming for the simplest maintainable structure for a small, focused Go library — one problem, one vocabulary, one module — imported by other Go programs, as required by the project constitution (Principles III, VII, IX, XIII).

For a small library principles of screaming, hexagonal, and onion architecture is overhead: there are no use-cases to decouple, no technologies to swap, and no composition to perform. Its rings, ports, and composition root would be empty indirections that make the code harder to read and the API harder to find.

A small library still needs the *properties* the hexagon protects — testability without network or credentials, independence from any specific technology, a stable public surface — and Go's standard library already provides the abstractions that deliver them (`io.Reader`, `io.Writer`, `io/fs.FS`, `http.RoundTripper`, `context.Context`, `iter.Seq`).

## Decision

1. **The root package is the library.** It is named after the module (`github.com/{org}/{name}` → `package {name}`) and holds the primary types and functions consumers import.
2. **Files are organized by concept**, one concept per file (`{name}.go` for the central type, then `codec.go`, `cursor.go`, `config.go`, `errors.go`, …), never by kind (`types.go`, `interfaces.go`, `utils.go`).
3. **Public sub-packages exist only for independent vocabularies** — something some consumers use without the rest (an optional encoding, a test helper such as `{name}test`). A sub-package is never introduced to group files.
4. **`internal/` holds helpers** shared between the root package and its sub-packages that are not part of the API.
5. **External effects use standard-library interfaces**, supplied by the consumer. Where none fits, the library declares a small interface next to the function that consumes it, exported only because consumers must implement it. There are no `port/`, `adapter/`, `service/`, `app/`, or `pkg/` directories.
6. **One module.** The library has a single `go.mod` with the standard library and the constitution's mandated libraries as its only target dependencies.

### Graduation

A library MUST be reconsidered for the capability constitution (with its edition of ADR 001) via an ADR when **any** of the following becomes true:

- it performs **three or more independent processes** that share state but should not share code (e.g., write, compact, recover);
- it drives **two or more interchangeable technologies** for the same concern (e.g., in-memory and cloud storage backends);
- it needs a **technology SDK** or another heavy dependency that not every consumer should inherit — i.e., it needs a nested module;
- its **public API spans more than a handful of packages**, or consumers need to supply implementations of several interfaces to use it;
- it needs a **composition step**: constructing the library requires wiring several internal components together.

Graduation is a deliberate decision recorded in an ADR, not an incremental drift: the package layout, test pattern (`sut()` over a composition root), and module structure all change together. Because principle numbering is identical across both constitutions, specs, plans, and ADRs remain valid.

## Neglected

- **Hexagonal/onion architecture** (the capability edition): rejected for small libraries; its rings and ports would be empty indirections.
- **[Standard Go Project Layout](https://github.com/golang-standards/project-layout?tab=readme-ov-file)**: its `pkg/`, `cmd/`, `api/` directories address applications and add path segments to every import without adding information to a library.
- **One package per type**: fragments one vocabulary into many imports and forces exported identifiers that should be unexported.

## To Achieve

A library that is read top-to-bottom in one package, documented on one pkg.go.dev page, imported with one `go get`, and composable with the rest of the Go ecosystem through the standard library's own interfaces.

## Accepting

All code of the library lives in one namespace, so internal cohesion is maintained by file organization and naming discipline rather than by package boundaries. We accept this because the library is small by definition; when that discipline stops being sufficient, the graduation criteria apply.

## Implementation notes

### Domain

The exported types of the root package are the domain vocabulary. Each term in the [ARCHITECTURE.md](../ARCHITECTURE.md) glossary maps to one exported identifier. Types make illegal states unrepresentable where Go allows it: named types over primitives, unexported fields behind validating constructors, typed constants for enumerations.

### External interfaces

| Concern            | Expressed as                                                                           |
| ------------------ | -------------------------------------------------------------------------------------- |
| byte streams       | `io.Reader`, `io.Writer` (and `io.ReaderAt`, `io.WriterTo` …)                          |
| files, directories | `io/fs.FS`, `fs.ReadDirFS`, `fs.StatFS`; tests use `fstest.MapFS`                      |
| HTTP               | consumer-supplied `*http.Client`; tests use `httptest.Server`                          |
| cancellation       | `context.Context` as the first parameter                                               |
| sequences          | `iter.Seq[T]`, `iter.Seq2[K, V]`                                                       |
| time, randomness   | a `Config` field accepting a clock function or `rand.Source`, when determinism matters |
| diagnostics        | a `Config` field accepting `*slog.Logger`, only if genuinely needed                    |
| anything else      | a small interface declared next to its consuming function                              |

### Test helpers for consumers

When consumers need to test their own code against the library (e.g., a fake implementation, a fixture builder), the library MAY ship a public `{name}test` sub-package, following the standard library's `httptest`/`fstest` convention. It is part of the public API and versioned with it.

### Project Layout

```
root/                          # module github.com/{org}/{name}
├── doc.go                     # package {name}: overview and quick start
├── {name}.go                  # central domain type(s) and primary functions
├── {name}_test.go             # unit tests (package {name})
├── conceptX.go                # one concept per file
├── conceptX_test.go
├── config.go                  # Config struct, documented zero-value defaults, validation
├── errors.go                  # exported error constants (faults.Type)
├── featureX_test.go           # black-box acceptance tests (package {name}_test)
├── example_test.go            # testable examples rendered on pkg.go.dev
├── fuzz_test.go               # fuzz tests for decoders/parsers
├── bench_test.go              # benchmarks for hot paths
├── testdata/                  # fixtures and fuzz corpora
├── vocabularyY/               # OPTIONAL: independent public vocabulary
│   └── ...
├── {name}test/                # OPTIONAL: public test helpers for consumers
│   └── ...
├── internal/                  # OPTIONAL: helpers shared by the packages above
│   └── ...
├── adrs/                      # architecture decision records
├── ARCHITECTURE.md            # glossary, package list, NFRs, hot paths
├── README.md
├── LICENSE
├── .github/workflows/         # build, check-code, check-test
├── go.mod
└── go.sum
```
