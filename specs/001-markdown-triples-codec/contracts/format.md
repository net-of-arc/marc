# Contract: Canonical Node Rendering and Decoding Grammar

**Normative source**: [CORE.md](https://github.com/net-of-arc/arcnet-spec/blob/main/spec/CORE.md) §3, §4, §7, §8.2, §8.3, §8.5. This contract fixes the byte-level choices CORE leaves to the producer, so that encoding is deterministic (FR-018) and decode∘encode is the identity on canonical files (FR-022).

**Canonical file** := any byte sequence that `Encode` can produce.

## 1. Layout

```text
---\n
"@id": <scalar>\n
"@type": <Class>\n
<meta predicates, code-point order>\n
"@conflicts":\n   (only if non-empty)
"@provenance":\n  (only if non-empty)
---\n
# <ID>\n
[\n<default prose>\n]
[\n**<Label>**\n<prose>\n]*           text predicates (non-default, non-Many), predicate order
[\n- <p>:: [[<T>]]\n ...]             edge block: (predicate, value) order, one bullet per triple
[\n## <Label>\n- <p>:: <value>\n ...]* link and Many-text blocks, predicate order
```

- Line endings are `\n`, and the file ends with exactly one `\n`.
- Sections are separated by exactly one blank line, and there is no blank line between `---` and the H1 (as in CORE §3).
- There is no blank line between a `## Label` heading and its first bullet.
- Empty sections are omitted entirely.

## 2. Front-matter values

| Value | Rendering |
|---|---|
| one value, `Many == false` | `key: <scalar>` |
| `Many == true`, or more than one value | flow sequence `key: [<scalar>, <scalar>]`, values in canonical order |
| `Text` scalar | plain if reading it back yields the same `!!str` in both block and flow context: not empty, no leading or trailing space, no non-printable character, none of `: # , [ ] { } " ' ?` anywhere, no indicator (`- \| > ! & * % @` and backtick) at the start, not a YAML 1.1/1.2 boolean or null word (`true`, `yes`, `on`, `null`, `~`, … in any case), and, when it starts with a digit or one of `+ . ~ < =`, resolving to `!!str` (so `1`, `2026-10-23`, `.inf`, `<<` are quoted). Otherwise double-quoted with JSON-style escapes (`\"`, `\\`, `\n`, `\t`, `\r`, `\uXXXX`) |
| `Number`, `Bool`, `Date`, `DateTime` | the variant's canonical rendering, unquoted |
| `Ref` in meta | double-quoted `"[[Target]]"` |
| `@`-keys | always double-quoted (CORE §4) |
| `@conflicts` | block map, predicate order; each value a flow sequence using the rules above; multi-line `Text` as a block-sequence item with a `|` block scalar |
| `@provenance` | block map document → block map predicate → flow sequence |

## 3. Canonical object rendering (`Object.String()`)

`Ref` → `[[T]]` / `[[T|display]]` · `Text` → its text · `Number` → lexical · `Bool` → `true`/`false` · `Date` → `2006-01-02` · `DateTime` → RFC 3339 with `Z`, fractional seconds only when non-zero.

## 4. Canonical order

- Predicates: by code point of the name.
- Values of one predicate: by variant tag (`b < d < n < r < s < t`), then within a variant by value for `Date`/`DateTime`, by code point of the canonical rendering otherwise.

## 5. Decoding grammar (tolerant reader)

The decoder accepts the canonical file and these non-canonical inputs:

| Input | Decoded as |
|---|---|
| `p:: [[T]]` line without a bullet (body form) | edge triple |
| `[p:: [[T]]]` inside prose (inline form) | edge triple, with the prose kept verbatim |
| `[[T]]` inside prose (untyped link) | nothing; stays in the prose |
| `**Label**` followed by `- p:: …` bullets inside the edge block | a grouping label; ignored |
| block-style YAML sequences, single-quoted scalars | the same values as the flow/double-quoted forms |
| bare `@id:` key | accepted if the YAML parser reads it |
| null or empty value | no triple |
| front-matter string `"[[T]]"` | `Ref` to `T` (meta-role edge value) |
| CRLF line endings | treated as LF |
| lines inside ```` ``` ```` / `~~~` fences | prose, never structure |

These are rejected: YAML anchors or aliases, nested maps outside `@conflicts`/`@provenance`, a `## Label` block with a non-bullet line, a bullet without `::`, a link target that is an invalid `ID`.

## 6. Observed properties

| Found as | Recorded `Property` |
|---|---|
| front-matter scalar | `{Role: RoleMeta}` |
| front-matter sequence | `{Role: RoleMeta, Many: true}` |
| unlabelled paragraphs | `{Role: RoleText, Default: true}` on `Schema.Predicate(class, "")`, otherwise `text` |
| `**Label**` + paragraph | `{Role: RoleText, Label: <label if ≠ derived>}` on `Schema.Predicate(class, label)`, otherwise the camelCase join of the label words |
| edge-block bullet / body / inline form | `{Role: RoleEdge}` |
| `## Label` block of `[[T]]` bullets | `{Role: RoleLink, Label: <heading if ≠ derived>}` |
| `## Label` block of non-link bullets | `{Role: RoleText, Many: true, Label: <heading if ≠ derived>}` |

`## Label` headings are resolved to predicates in the same way (`Schema.Predicate(class, heading)`), but the bullets' own `p::` prefix always wins: the heading carries no data (CORE §4).

When a predicate appears in several positions, the first position wins (spec Edge Cases).

## 7. Values a role cannot render

The encoder never drops a value. A value its predicate's role cannot represent in the body is written to the front-matter as a meta value of that predicate instead, and reads back unchanged:

| Role | Rendered in the body | Moved to front-matter |
|---|---|---|
| `edge`, `link` | `Ref` values | every other variant |
| `text`, not `Many`, one value | one `Text` that is non-blank, has no leading or trailing blank line and no `\r` | any other `Text`, every other variant |
| `text` with `Many`, or `text` with more than one value | single-line, trimmed, non-empty `Text` not of the form `[[T]]`, as `## Label` statement bullets | any other `Text`, every other variant |

A body-renderable prose value that contains structure outside fenced code is rejected with `ErrProseStructure`: an ATX heading line (`#`…`######`), a `- p::` bullet, a typed body-form edge `p:: [[T]]`, a bold label line `**Label**`, or an unterminated fence.

A `Text` of the form `[[T]]` cannot be represented as text anywhere: it decodes as a `Ref`. Input that is not valid UTF-8 is rejected with `ErrMalformed`.
