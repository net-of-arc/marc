# Contract: Exported Go API of `github.com/net-of-arc/marc`

**Feature**: [../spec.md](../spec.md) · **Data model**: [../data-model.md](../data-model.md) · **Format**: [format.md](format.md)

This is the complete exported surface introduced by this feature, shown as signatures. GoDoc for each identifier must state its behaviour, the errors it returns, whether it is safe for concurrent use and who owns its arguments (Principle XII). An identifier not listed here is unexported.

## Identity and names — `id.go`, `name.go`

```go
type ID struct{ /* unexported */ }
func NewID(s string) (ID, error)          // ErrInvalidID
func MustID(s string) ID                  // panics; constants & tests only
func (id ID) String() string
func (id ID) IsZero() bool

type Predicate struct{ /* unexported */ }
func NewPredicate(s string) (Predicate, error) // ErrInvalidName
func MustPredicate(s string) Predicate
func (p Predicate) String() string
func (p Predicate) Label() string              // derived label, e.g. "Mentioned In"
func (p Predicate) IsZero() bool

type Class struct{ /* unexported */ }
func NewClass(s string) (Class, error)         // ErrInvalidName
func MustClass(s string) Class
func (c Class) String() string
func (c Class) IsZero() bool
```

## Object sum type — `object.go`

```go
type Object interface {
    String() string   // canonical rendering (format.md §3)
    key() string      // sealed: equality key, CORE §8.5
}

type Ref struct{ /* target ID, display string */ }
func RefTo(target ID) Ref
func (r Ref) WithDisplay(text string) Ref     // returns a copy
func (r Ref) Target() ID
func (r Ref) Display() string

type Text string
type Bool bool

type Number struct{ /* lexical string */ }
func ParseNumber(s string) (Number, error)    // ErrInvalidObject
func (n Number) Int64() (int64, bool)
func (n Number) Float64() (float64, bool)

type Date struct{ /* civil date */ }
func DateOf(t time.Time) Date                 // keeps t's year, month, day
func (d Date) Time() time.Time                // midnight UTC

type DateTime struct{ /* UTC instant */ }
func DateTimeOf(t time.Time) DateTime         // normalised to UTC
func (d DateTime) Time() time.Time

func Select[T Object](seq iter.Seq[Object]) iter.Seq[T]
```

How a client handles objects dynamically:

```go
for o := range node.Values(p) {
    switch v := o.(type) {
    case marc.Ref:      follow(v.Target())
    case marc.Text:     show(string(v))
    case marc.Number:   n, _ := v.Float64(); plot(n)
    case marc.Bool, marc.Date, marc.DateTime:
    }
}
for ref := range marc.Select[marc.Ref](node.Values(p)) { follow(ref.Target()) }
```

## Triple, Node and Attribution — `triple.go`, `node.go`

```go
type Triple struct {
    Subject   ID
    Predicate Predicate
    Object    Object
}

type Attribution struct {
    Document  ID
    Predicate Predicate
    Object    Object
}

type Node struct{ /* unexported */ }
func NewNode(id ID, class Class, triples iter.Seq[Triple]) (Node, error)
    // ErrInvalidID, ErrInvalidName, ErrInvalidObject, ErrForeignSubject.
    // Consumes triples once; retains no reference to the sequence. Duplicates collapse (set).
func (n Node) ID() ID
func (n Node) Class() Class
func (n Node) Triples() iter.Seq[Triple]                 // natural order (BUG-001; was canonical)
func (n Node) Predicates() iter.Seq[Predicate]           // natural order: by first triple (BUG-001; was canonical)
func (n Node) Values(p Predicate) iter.Seq[Object]       // natural order; empty if absent (BUG-001; was canonical)
func (n Node) Links() iter.Seq2[Predicate, Ref]          // untyped [[T]] in Text values; derived, not triples (BUG-001, FR-026)
func (n Node) Property(p Predicate) (Property, bool)     // as observed while decoding
func (n Node) Conflicts() iter.Seq2[Predicate, Object]   // @conflicts, flattened
func (n Node) Provenance() iter.Seq[Attribution]         // @provenance, flattened
```

`Node` is immutable and safe for concurrent use; the zero `Node` is invalid and the encoder rejects it with `ErrMalformed`.

*Natural order* (BUG-001): the order triples entered the node — document order on decode, `NewNode` sequence order otherwise — with duplicates keeping their first position. `Conflicts` and `Provenance` stay in canonical order.

## Rendering rules — `schema.go`

```go
type Role uint8
const (
    RoleMeta Role = iota
    RoleText
    RoleEdge
    RoleLink
)
func (r Role) String() string

type Property struct {
    Role    Role
    Label   string // "" ⇒ Predicate.Label()
    Default bool   // default (unlabelled) text predicate
    Many    bool   // meta: YAML array; text: "## Label" block of statements
}

type Schema interface {
    Property(c Class, p Predicate) (Property, bool)    // encoder lookup
    Predicate(c Class, label string) (Predicate, bool) // decoder lookup; "" ⇒ default-text predicate
}

type Properties map[Predicate]Property                 // reference Schema; ignores class
func (s Properties) Property(c Class, p Predicate) (Property, bool)
func (s Properties) Predicate(c Class, label string) (Predicate, bool)
func (s Properties) Validate() error                   // ErrInvalidProperty
```

`Schema` is implemented by the application, which owns its vocabulary. The codec neither copies nor mutates it; the implementation must be safe for concurrent use whenever the `Codec` is. Each `Property` it returns is validated at use (`ErrInvalidProperty`).

An application adapts its own schema types like this:

```go
type vocabulary struct{ /* app-owned schema nodes */ }

func (v *vocabulary) Property(c marc.Class, p marc.Predicate) (marc.Property, bool) {
    def, ok := v.lookup(c, p)
    if !ok {
        return marc.Property{}, false
    }
    return marc.Property{Role: def.role, Label: def.label, Many: def.cardinality > 1}, true
}
```

## Codec — `codec.go`, `config.go`

```go
type Config struct {
    Schema  Schema // nil ⇒ infer roles on decode; observed/fallback on encode; default text "text"
    MaxSize int    // bytes; 0 ⇒ 4 MiB; negative ⇒ ErrInvalidConfig
}

type Codec struct{ /* unexported */ }
func New(cfg Config) (*Codec, error)                     // ErrInvalidConfig; keeps cfg.Schema by reference
func (c *Codec) Decode(r io.Reader) (Node, error)
func (c *Codec) Unmarshal(b []byte) (Node, error)        // does not retain b
func (c *Codec) Encode(w io.Writer, n Node) error
func (c *Codec) Marshal(n Node) ([]byte, error)

func Decode(r io.Reader) (Node, error)                   // default Codec
func Unmarshal(b []byte) (Node, error)
func Encode(w io.Writer, n Node) error
func Marshal(n Node) ([]byte, error)
```

Decode errors: `ErrTooLarge`, `ErrMalformed`, `ErrIdentityMismatch`, `ErrInvalidID`, `ErrInvalidName`, `ErrInvalidObject`, plus the reader's own error wrapped. Encode errors: `ErrMalformed` (zero node), `ErrProseStructure`, `ErrInvalidProperty`, plus the writer's own error wrapped. No function panics on any input a client can construct, except the `Must*` constructors.

## Errors — `errors.go`

```go
const (
    ErrMalformed        = faults.Type("malformed node")
    ErrIdentityMismatch = faults.Type("title does not match @id")
    ErrInvalidID        = faults.Type("invalid identity")
    ErrInvalidName      = faults.Type("invalid predicate or class name")
    ErrInvalidObject    = faults.Type("invalid object")
    ErrForeignSubject   = faults.Type("triple subject differs from node identity")
    ErrTooLarge         = faults.Type("input exceeds size limit")
    ErrProseStructure   = faults.Type("prose contains structural markup")
    ErrInvalidProperty  = faults.Type("invalid property")
    ErrInvalidConfig    = faults.Type("invalid config")
)
```

Clients match errors with `errors.Is` only.

## Traceability

| Spec item | API |
|---|---|
| Story 1 / FR-001, 004, 007, 009 | `Decode`, `Unmarshal`, `Node.Triples/Values/Predicates`, `Object` variants, `Select` |
| Story 2 / FR-008, 016–019 | `NewNode`, `Encode`, `Marshal`, `Schema`, `Properties`, `Property`, `Role` |
| Story 3 / FR-010, 012, 022, 023 | `Node.Property`, observed properties, canonical encoder |
| Story 4 / FR-011 | `Text` prose values |
| Story 4 / FR-026 (BUG-001) | `Node.Links` |
| Story 5 / FR-002, 003, 014, 015, 020, 024 | `NewID`, `NewPredicate`, `NewClass`, `Config.MaxSize`, error constants |
| Story 6 / FR-005, 021 | `Node.Conflicts`, `Node.Provenance`, `Attribution` |

## Ownership and concurrency (GoDoc requirements)

Each rule below must appear in the GoDoc of the identifier it concerns (Principle IX):

- `ID`, `Predicate`, `Class`, every `Object` variant and `Node` are immutable values and safe for concurrent use.
- `NewNode` consumes the triple sequence exactly once and retains no reference to it.
- `Unmarshal` (package function and method) does not retain its `[]byte` argument; `Decode` does not retain the reader.
- `Codec` keeps `Config.Schema` by reference, never copies or mutates it, and is safe for concurrent use when the schema is.
- `Node` iterators (`Triples`, `Predicates`, `Values`, `Links`, `Conflicts`, `Provenance`) yield copies of immutable values; the caller cannot mutate the node through them.
