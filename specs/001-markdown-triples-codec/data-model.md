# Data Model: Markdown ⇄ Triples Codec Core Types

**Feature**: [spec.md](spec.md) · **Research**: [research.md](research.md) · **API contract**: [contracts/api.md](contracts/api.md)

**Bugfix**: 2026-10-24 — BUG-001 `facts` are an insertion-ordered set; one `Text` per paragraph; `Links()` derived view

Every entity is an immutable value. Constructors validate; a non-zero value of every type is legal. Nothing here has a lifecycle or a state transition: the only "transition" is a pure function from one value to another (decode, encode).

## Entity map

```text
                  ┌──────────── Node ─────────────┐
                  │ id        ID                  │
                  │ class     Class               │
                  │ facts     [Triple] ordered set│──── Triples() ──▶ iter.Seq[Triple]
                  │                               │──── Links() ────▶ iter.Seq2[Predicate, Ref]  (derived from Text values, BUG-001)
                  │ observed  Predicate → Property│
                  │ conflicts Predicate → [Object]│──── Conflicts() ─▶ iter.Seq2[Predicate, Object]
                  │ provenance [Attribution]      │──── Provenance() ▶ iter.Seq[Attribution]
                  └───────────────────────────────┘
Triple = (Subject ID, Predicate, Object)

Object (sealed) ─┬─ Ref       { target ID, display string }
                 ├─ Text      string
                 ├─ Number    { lexical string }
                 ├─ Bool      bool
                 ├─ Date      { civil date }
                 └─ DateTime  { UTC instant }

Schema (interface, app-owned)
  Property(Class, Predicate)  → Property{ Role, Label, Default, Many }   encoder
  Predicate(Class, label)     → Predicate   ("" = default text)          decoder
Properties = Predicate → Property   reference Schema, class-agnostic
Codec  = { schema (by reference), maxSize }   built by New(Config)
```

## Entities

### ID — node identity (`@id`)

| Aspect | Rule | Source |
|---|---|---|
| Representation | opaque, comparable, map-key safe | R2 |
| Normalisation | NFC applied on construction | CORE §5, R7 |
| Forbidden | `/ \ : * ? " < > \| . # ^ [ ]`, leading or trailing space, two consecutive spaces, empty, invalid UTF-8, control characters, newlines | CORE §5, FR-002 |
| Zero value | `IsZero() == true`; rejected by `NewNode` and the encoder | Principle IX |
| Error | `ErrInvalidID` | FR-002 |

### Predicate — property name

camelCase: `^[a-z][A-Za-z0-9]*$`. Error: `ErrInvalidName`. The zero value is invalid. The derived label splits the name at lower→upper and letter→digit boundaries and capitalises each word (`mentionedIn` → `Mentioned In`).

### Class — node type (`@type`)

UpperCamelCase: `^[A-Z][A-Za-z0-9]*$`. Error: `ErrInvalidName`. The zero value is invalid.

### Object — sealed sum type

| Variant | Payload | Constructed by | Equality key (CORE §8.5, R8) | Canonical rendering |
|---|---|---|---|---|
| `Ref` | target `ID`, display text | `RefTo(ID)`, `Ref.WithDisplay(string)` | `r:` + target | `[[Target]]` or `[[Target\|display]]` |
| `Text` | string | conversion `Text(s)` | `t:` + NFC, per-line right-trim, overall trim | plain/quoted scalar (meta) or verbatim prose |
| `Number` | lexical form | `ParseNumber(string)` | `n:` + trimmed lexical | lexical form |
| `Bool` | bool | conversion `Bool(b)` | `b:true` / `b:false` | `true` / `false` |
| `Date` | civil date | `DateOf(time.Time)` (keeps Y-M-D) | `d:` + `YYYY-MM-DD` | `YYYY-MM-DD` |
| `DateTime` | instant | `DateTimeOf(time.Time)` (normalised to UTC) | `s:` + UTC RFC 3339 nano | RFC 3339 in UTC (`Z`) |

Invariants:
- The set of variants is closed (unexported sealing method `key()`).
- A zero `Ref`, `Number`, `Date` or `DateTime` is rejected by `NewNode` with `ErrInvalidObject`.
- `Text` values that come from front-matter are single-line; multi-line `Text` is prose.
- A prose `Text` is one paragraph: it contains no blank line outside fenced code (BUG-001). A `Text` that does is not body-renderable and moves to front-matter (format.md §7).

### Triple

`{Subject ID; Predicate Predicate; Object Object}`: a plain comparable-by-field value with exported fields. Validity is checked when it enters a `Node`, not when the struct is built:
- the subject must equal the node's `ID`, otherwise `ErrForeignSubject`;
- the predicate must be non-zero, otherwise `ErrInvalidName`;
- the object must be non-nil and non-zero, otherwise `ErrInvalidObject`.

### Node — one Markdown file as a bag of triples

| Field | Meaning | Rules |
|---|---|---|
| `id` | the subject | non-zero |
| `class` | the `@type` | non-zero |
| `facts` | ~~predicate → set of objects~~ insertion-ordered set of triples (BUG-001) | set semantics by equality key (FR-006), first occurrence keeps its position (FR-007); no keyword predicates (`@id`, `@type`, `@conflicts`, `@provenance` cannot be expressed as `Predicate` because `@` fails camelCase) |
| `observed` | predicate → `Property` seen while decoding | empty for nodes built by `NewNode`; used by the encoder when the schema has no entry (FR-010, FR-017) |
| `conflicts` | predicate → open alternatives | decode-only in this feature; written back verbatim in canonical order (FR-021) |
| `provenance` | list of `Attribution` | decode-only in this feature; written back in canonical order (FR-021) |

Behaviour:
- `Triples()`, `Values(p)`, `Predicates()` iterate in ~~canonical order~~ natural order: document order for a decoded node, insertion order for `NewNode`; predicates by their first triple (FR-007, BUG-001).
- `Links()` yields `(predicate, Ref)` for every untyped `[[T]]` / `[[T|d]]` inside the node's `Text` values, in natural order, once per (predicate, target), skipping fenced code and inline typed edges `[p:: [[T]]]`. It is derived on demand and never stored as triples (FR-026, BUG-001).
- `Conflicts()` and `Provenance()` keep canonical order (CORE §8.5).
- `Values(p)` on an absent predicate yields nothing (Story 1, scenario 4).
- A node is immutable after construction and safe for concurrent use. Iterators read shared state but never mutate it.

### Attribution — one `@provenance` entry

`{Document ID; Predicate Predicate; Object Object}`. A `text`-role digest (`sha256:…`) is carried as `Text`; an edge value `"[[X]]"` decodes to `Ref`.

### Property — rendering rule for one predicate

| Field | Type | Zero value means |
|---|---|---|
| `Role` | `Role` (`RoleMeta`, `RoleText`, `RoleEdge`, `RoleLink`) | `RoleMeta` is the zero value; when no property is known at all, the encoder falls back to meta for literals and edge for refs (FR-017) |
| `Label` | string | the label derived from the predicate name |
| `Default` | bool | not the default text; only legal with `RoleText` and `Many == false` |
| `Many` | bool | meta: scalar when there is one value; text: a single (prose) value. `true` → meta renders as an array, text renders as a `## Label` block of `p:: statement` bullets |

A `Property` is validated **each time the codec obtains one** from a `Schema` (`ErrInvalidProperty`, naming the predicate): `Role` within range; `Default` only on `RoleText` without `Many`.

### Schema — consumer-implemented interface

| Method | Used by | Contract |
|---|---|---|
| `Property(c Class, p Predicate) (Property, bool)` | encoder | how `p` renders on nodes of class `c`; `false` → fall back to the observed property, then to the default role |
| `Predicate(c Class, label string) (Predicate, bool)` | decoder | the predicate a `**Label**` / `## Label` names on class `c`; `label == ""` asks for the class's default-text predicate; `false` → camelCase join of the label (or `text` for `""`) |

The codec does not copy the schema or take ownership of it. The implementation must be safe for concurrent use whenever the `Codec` is, and must keep the two methods consistent: `Predicate(c, "")` names the predicate whose `Property(c, ·).Default` is true, and labels are unique per class.

### Properties — reference `Schema`

`map[Predicate]Property`, class-agnostic.
- `Property` is a map lookup.
- `Predicate(c, label)` matches the entry whose explicit or derived label equals `label`, or the `Default` entry for `""`.
- `Validate() error` (`ErrInvalidProperty`) checks that there is at most one `Default`, that labels are unique after derivation, and that each property is valid.

It is used by the examples and tests, and by applications with a static vocabulary.

### Config

| Field | Type | Zero value / default | Validation |
|---|---|---|---|
| `Schema` | `Schema` (interface) | nil: roles inferred on decode, observed or fallback on encode; default text predicate `text` | none at `New` (validated per `Property` at use) |
| `MaxSize` | `int` | 4 MiB (4 194 304 bytes) | `< 0` → `ErrInvalidConfig` |

### Codec

Built by `New(Config)`. It holds the consumer's `Schema` (by reference) and the size limit. The codec itself is immutable; it is safe for concurrent use when the supplied `Schema` is. The package-level functions use `New(Config{})`, which has no schema and is built once at package initialisation.

## Errors (`faults.Type` constants)

| Constant | Raised when | FR |
|---|---|---|
| `ErrMalformed` | no or unterminated front-matter, YAML syntax error, anchors/aliases, nested maps outside the keywords, missing `"@id"`/`"@type"`, missing H1, a non-bullet line in a `## Label` block, input that is not UTF-8, empty input | FR-014 |
| `ErrIdentityMismatch` | H1 ≠ `"@id"` | FR-014 |
| `ErrInvalidID` | `NewID` fails, or an invalid `"@id"`/link target is decoded | FR-002, FR-014 |
| `ErrInvalidName` | an invalid predicate or class (constructed or decoded) | FR-003, FR-014 |
| `ErrInvalidObject` | a nil or zero-variant object entering a node, a text that is not valid UTF-8, a display text containing a line break, `[[` or `]]`, or an invalid number lexical form | FR-004 |
| `ErrForeignSubject` | a triple's subject ≠ the node's id | FR-005 |
| `ErrTooLarge` | input exceeds `MaxSize` | FR-015 |
| `ErrProseStructure` | encoding prose that contains, outside fenced code, a heading line, a `p::` bullet, a typed body-form edge, a bold label line, or an unterminated fence | FR-020 |
| `ErrInvalidProperty` | a `Schema` returns an inconsistent `Property`, or `Properties.Validate` fails | FR-008, FR-017 |
| `ErrInvalidConfig` | `New` rejects the `Config` | Principle IX |

Error messages name the element (`"@id"`, the predicate name, the line number) and never include payload values (Principle X).
