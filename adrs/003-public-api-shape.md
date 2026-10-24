# ADR 003 - Public API Shape: Sealed Object Sum Type, Validated Names, Context-Free Codec

**Status**: Accepted
**Date**: 2026-10-23

## Context

`marc` exposes Markdown nodes as bags of `(subject, predicate, object)` triples ([spec 001](../specs/001-markdown-triples-codec/spec.md)). Three API-shape decisions are expensive to change after release (Principle XIV), so they are recorded here. The design research is in [research.md](../specs/001-markdown-triples-codec/research.md) R1–R3.

## Decision

1. **Object is a sealed interface sum type.**
   - `Object` declares `String() string` and an unexported `key() string`, so no type outside `marc` can implement it.
   - The variants are `Ref`, `Text`, `Number`, `Bool`, `Date` and `DateTime`, each a concrete type that carries only its own payload.
   - Clients branch with a type switch. The generic `Select[T Object](iter.Seq[Object]) iter.Seq[T]` gives typed filtering.
   - Variants with invariants (`Ref`, `Number`, `Date`, `DateTime`) keep unexported fields behind constructors.
2. **Identity and names are validated opaque types.** `ID`, `Predicate` and `Class` are structs with one unexported field, built through `NewX(string) (X, error)` and `MustX(string) X`. A non-zero value has always passed validation (CORE §5 for `ID`, camelCase for `Predicate`, UpperCamelCase for `Class`).
3. **The codec is context-free.**
   - `Decode(io.Reader)`, `Unmarshal([]byte)`, `Encode(io.Writer, Node)` and `Marshal(Node)` are pure transformations, also available as methods on a configured `*Codec` from `New(Config)`.
   - They take no `context.Context`. The codec performs no I/O of its own: it consumes a caller-owned reader, bounded by `Config.MaxSize`, and its CPU work is linear in that bound. Cancellation is the reader owner's responsibility, the same contract as `encoding/json` and `encoding/xml`.
   - This is the project's reading of Principle IX ("context.Context is the first parameter of every exported function that performs I/O, blocks, or may be canceled") for codecs.
4. **The schema is consumer-owned.**
   - `Schema` is an interface the application implements: `Property(Class, Predicate) (Property, bool)` for encoding, and `Predicate(Class, label string) (Predicate, bool)` for decoding, where `""` names the class's default-text predicate.
   - The codec keeps it by reference and never copies or mutates it.
   - `Property{Role, Label, Default, Many}` stays a value struct: the library's vocabulary for the answer, which application schema types produce inside their `Property` method.
   - `Properties` (`map[Predicate]Property`) is the reference implementation, with `Validate()`.
   - Each `Property` is validated at use (`ErrInvalidProperty`), because an interface cannot be validated as a whole at `New`.
5. **Sequences are iterators.** `Node.Triples`, `Node.Values`, `Node.Predicates` and `Node.Provenance` return `iter.Seq`; `Node.Conflicts` returns `iter.Seq2`. `NewNode` consumes an `iter.Seq[Triple]`.

## Consequences

- Adding a new `Object` variant is a breaking change for clients' type switches in practice, even though `apidiff` reports it as compatible. Reviewers must treat it as MAJOR, or as MINOR with a release-note warning while the module is on v0.
- Every object value is boxed in an interface (one small allocation). This is accepted against the SC-006 budget and verified by benchmark.
- Compiler-checked exhaustiveness is not available. GoDoc lists the closed set of variants.
- If streaming decode of very large inputs is ever needed, a context-aware API is added alongside; the existing functions remain.
- Adding a method to `Schema` breaks every application implementation, so the interface is frozen at two methods. New needs are met by new `Property` fields (MINOR) or by an optional interface detected with a type assertion.
- Schema inconsistencies (duplicate labels, a default-text predicate that disagrees with `Property.Default`) can only be caught by the implementation; the reference `Properties` catches them in `Validate`.

## Neglected

- Tagged struct (`Kind` + union of fields): illegal states are representable.
- Visitor / `Match(cases)`: unidiomatic, and still not exhaustive at compile time.
- Generic `Object[T]`: cannot hold mixed kinds in one node.
- `type ID string`: conversion bypasses validation.
- `ctx` on codec functions: there is no I/O or goroutine to cancel, and it would diverge from standard-library codecs.
- `type Schema map[Predicate]Property` copied into `Config`: forces applications to snapshot and convert their vocabulary, and cannot follow a growing one.
- `Property` as an interface: adds nil and mutability hazards and dynamic dispatch for no gain.
