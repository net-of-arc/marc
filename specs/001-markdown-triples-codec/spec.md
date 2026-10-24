# Feature Specification: Markdown ⇄ Triples Codec Core Types

**Feature Branch**: `001-markdown-triples-codec`

**Created**: 2026-10-23

**Status**: Draft

**Bugfix**: 2026-10-24 — BUG-001 Text values are per paragraph, inline `**Label**:` paragraphs are labelled text, order is natural (document/insertion order within the CORE §4 sections), and untyped links are exposed as a derived view (not triples)

**Input**: User description: "The library is pure Markdown to/from subject-predicate-object triples codec. The markdown format is governed by https://github.com/net-of-arc/arcnet-spec/blob/main/spec/CORE.md . Implement the core types to be used as a public interface between the client application and this library so that client operates with Markdown files as bag of triples."

## Context

The ArcNet CORE format ([CORE.md](https://github.com/net-of-arc/arcnet-spec/blob/main/spec/CORE.md)) stores a knowledge graph as plain Markdown files. Each file is a **node** describing exactly one **subject**; underneath the Markdown, a node is a **bag of predicates about that subject** — one `(subject, predicate, object)` triple per value. CORE states that Markdown is a lossless serialization of that bag, up to canonical form (CORE §3).

This feature delivers the public vocabulary through which a client application reads a node file as a bag of triples and writes a bag of triples back as a node file, without knowing anything about Markdown or YAML. The vocabulary (triples, subjects, predicates, objects, the node, and predicate roles) is the contract; decoding and encoding are the two operations over it.

The **actor** throughout is the *client developer*: a programmer whose application (an ingestion tool, an agent, an indexer, a merge driver) consumes this library.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read a node file as a bag of triples (Priority: P1)

A client developer hands the library the content of one Markdown node file and receives the node's subject together with every statement the file makes about it, as triples. The developer can iterate the triples, look up all values of a given predicate, and tell whether each object is a literal value or a reference to another node — without parsing Markdown or YAML themselves.

**Why this priority**: Reading is the entry point for every downstream use (indexing, querying, merging, validation). A read-only codec is already a viable product.

**Independent Test**: Decode the CORE §3 example node ("Transport Layer Security") and compare the resulting bag with the triple listing given in CORE §3.

**Acceptance Scenarios**:

1. **Given** the CORE §3 example node, **When** the client decodes it, **Then** the node's subject is `Transport Layer Security`, its type is `Entity`, and the bag contains exactly these statements: `category "script"`, `tags "cryptography"`, `text "A cryptographic protocol that establishes an authenticated, confidential channel over an untrusted network."`, `replaces → SSL Protocol`, `conformsTo → RFC 8446`, `mentionedIn → rescorla-2026-tls13`.
2. **Given** a node whose front-matter holds a multi-valued predicate `tags: [cryptography, networking]`, **When** the client decodes it, **Then** the bag contains two separate triples, one per value, sharing the same subject and predicate.
3. **Given** a decoded node, **When** the client asks for the values of predicate `replaces`, **Then** it receives the single object `SSL Protocol`, identified as a reference to another node rather than a literal.
4. **Given** a decoded node, **When** the client asks for the values of a predicate the node does not carry, **Then** it receives an empty result, not an error.
5. **Given** a node with a front-matter value that is a number, a boolean, a date or a datetime, **When** the client decodes it, **Then** the corresponding object reports its literal kind (number, boolean, date, datetime) along with its exact written form.
6. **Given** a node whose front-matter holds `tags: [zeta, alpha]` and whose edge block lists `replaces` before `conformsTo`, **When** the client decodes it, **Then** `Triples()` yields the front-matter triples, then the text triples, then the edge triples, then the link triples, and within each section in the order they are written (`tags "zeta"` before `tags "alpha"`, `replaces` before `conformsTo`); `Predicates()` and `Values()` follow the same order. *(BUG-001)*

---

### User Story 2 - Write a bag of triples as a canonical node file (Priority: P1)

A client developer builds a node in memory — a subject, a type, and a set of triples — and asks the library to render it as a Markdown node file that conforms to CORE §4: YAML front-matter, H1 title equal to the subject, default prose, labelled prose, a heterogeneous edge block, and one `## Label` block per link-role predicate. The developer tells the library which role each predicate plays; the library decides placement mechanically.

**Why this priority**: Writing is the other half of the codec; together with Story 1 it enables round-tripping, which is the defining promise of the format.

**Independent Test**: Build the CORE §3 node from triples plus a role table (`category`, `tags` → meta; `text` → text, default; `replaces`, `conformsTo` → edge; `mentionedIn` → link) and compare the output byte-for-byte with the canonical form of the CORE §3 example.

**Acceptance Scenarios**:

1. **Given** the CORE §3 triples and their roles, **When** the client encodes the node, **Then** the output is the CORE §3 example file in canonical form, with `"@id"` and `"@type"` written as quoted front-matter keys.
2. ~~**Given** a node with a multi-valued predicate whose values were added in the order `zeta, alpha, mu`, **When** the client encodes it, **Then** the values are written in lexicographic order `alpha, mu, zeta`, so that independent writers produce identical files.~~ *(superseded by BUG-001: CORE §3 asks only for a stable order; the natural order is the stable order)* **Given** a node with a multi-valued predicate whose values were added in the order `zeta, alpha, mu`, **When** the client encodes it, **Then** the values are written in the order they were added, `zeta, alpha, mu`, and encoding the same node again yields identical bytes.
3. **Given** a link-role predicate `mentionedIn` with no declared label, **When** the client encodes the node, **Then** its block heading is `## Mentioned In` (camelCase split and capitalized) and each target is written as `- mentionedIn:: [[Target]]`.
4. **Given** a link-role predicate that declares the label `Sources`, **When** the client encodes the node, **Then** its block heading is `## Sources`.
5. **Given** a text-role predicate that is not the default text, **When** the client encodes the node, **Then** it appears as a bold `**Label**` line followed by its paragraph, after the default prose and before the edge block.
6. **Given** a node with no triples beyond its subject and type, **When** the client encodes it, **Then** the output contains front-matter with `"@id"` and `"@type"` and an H1 title, and nothing else.
7. **Given** a node whose default text predicate holds the values `abc` and `def` (added in that order) and whose labelled text predicate `note` holds two values, **When** the client encodes it, **Then** the default prose is written as the paragraphs `abc` and `def` separated by a blank line, followed by `**Note**` on its own line and its two paragraphs; the sections still follow CORE §4 (meta → text → edge → link) whatever order the triples were added in. *(BUG-001)*

---

### User Story 3 - Lossless round trip (Priority: P1)

A client developer relies on the guarantee that decoding and re-encoding preserves meaning: Markdown → triples → Markdown reproduces any canonical file exactly, and triples → Markdown → triples reproduces the original bag exactly, including predicates the client's vocabulary does not know.

**Why this priority**: Round-trip fidelity is what lets clients treat files as data. Without it, every write risks silently dropping statements, which CORE forbids ("Consumers MUST preserve unknown predicates").

**Independent Test**: For a corpus of canonical node files, decode then encode each one and compare bytes; for a corpus of generated bags, encode then decode and compare bags as sets.

**Acceptance Scenarios**:

1. **Given** any canonical node file, **When** the client decodes it and encodes the result without supplying any role information, **Then** the output is byte-identical to the input.
2. **Given** any bag of triples with roles for each predicate, **When** the client encodes and then decodes it, **Then** the resulting bag equals the original as a set of triples, and within each CORE §4 section the triples keep their original relative order. *(order clause: BUG-001)*
3. **Given** a node containing a predicate unknown to the client, **When** the client decodes and re-encodes it, **Then** the unknown predicate is written back in the same position (front-matter, labelled prose, edge block or link block) it was found in.
4. **Given** a non-canonical file using the body form `conformsTo:: [[RFC 8446]]` or the inline form `[supports:: [[RFC 8446]]]`, **When** the client decodes it, **Then** each yields a triple identical to the one the list form would yield, and re-encoding writes the list form.

---

### User Story 4 - Prose and untyped links stay intact (Priority: P2)

A client developer reads prose predicates as text values that keep their untyped `[[wiki-links]]` inline, and is assured those untyped links never become triples of their own.

**Why this priority**: Prose carries most of a node's human value; mangling it or inventing statements from it would corrupt both the reading experience and the graph.

**Independent Test**: Decode a node whose default prose contains `see [[RFC 8446]] for details` and inspect the bag.

**Acceptance Scenarios**:

1. **Given** default prose containing an untyped link `[[RFC 8446]]`, **When** the client decodes the node, **Then** the bag holds ~~one text triple whose value contains~~ a text triple for the paragraph that contains `[[RFC 8446]]` verbatim, and no triple whose object is a reference to `RFC 8446` arises from that prose. *(BUG-001: one triple per paragraph)*
2. **Given** default prose spanning several paragraphs, **When** the client decodes ~~and re-encodes~~ the node, **Then** ~~the paragraphs and their order are preserved~~ each paragraph is a separate text triple, yielded in document order, and re-encoding reproduces the paragraphs in that order. *(BUG-001: was one value holding all paragraphs)*
3. **Given** a text-role predicate whose values are a list of statements (a `## Label` block of `predicate:: statement` bullets), **When** the client decodes it, **Then** each bullet becomes one text triple.
4. **Given** default prose containing the untyped links `[[RFC 8446]]` and `[[SSL Protocol|SSL]]` and the inline typed edge `[supports:: [[RFC 5246]]]`, **When** the client asks the node for its untyped links, **Then** it receives `(text, RFC 8446)` and `(text, SSL Protocol)` in document order, the bag still holds no triple for them, and `supports → RFC 5246` is a triple and is not listed among the untyped links. *(BUG-001)*
5. **Given** a paragraph `**Note**: This is a note.`, **When** the client decodes the node, **Then** the bag holds the text triple `note "This is a note."` with the role of labelled text, exactly as for `**Note**` on its own line followed by `This is a note.`; a paragraph `**Note** This is a note.` (no colon) stays part of the surrounding prose; re-encoding writes `**Note**` on its own line. *(BUG-001)*

---

### User Story 5 - Reject malformed input with actionable errors (Priority: P2)

A client developer feeding untrusted or hand-edited files receives a distinguishable error for each class of defect, so the application can report or skip the file, and the library never crashes or hangs on hostile input.

**Why this priority**: Graph files are hand-edited and AI-generated; clients must be able to separate "this file is broken" from "my code is broken".

**Independent Test**: Decode a set of deliberately malformed files and check each yields the expected error category.

**Acceptance Scenarios**:

1. **Given** a file with no front-matter, or front-matter lacking `"@id"` or `"@type"`, **When** the client decodes it, **Then** it receives a "malformed node" error naming the missing element.
2. **Given** a file whose H1 title differs from its `"@id"`, **When** the client decodes it, **Then** it receives an identity-mismatch error.
3. **Given** a subject containing any of `/ \ : * ? " < > | . # ^ [ ]`, a leading or trailing space, or two consecutive spaces, **When** the client decodes such a file or tries to construct such a subject, **Then** it receives an invalid-identity error.
4. **Given** a predicate name that is not camelCase or a type that is not UpperCamelCase, **When** the client constructs it or decodes a file containing it, **Then** it receives an invalid-name error.
5. **Given** input larger than the configured size limit, **When** the client decodes it, **Then** it receives a size-limit error without the whole input being read into memory.
6. **Given** a bag whose prose value contains a Markdown heading line or a line beginning with a `predicate::` bullet, **When** the client encodes it, **Then** it receives an error instead of producing a file that would be re-read as structure.
7. **Given** arbitrary bytes, **When** the client decodes them, **Then** the library returns either a node or an error and never crashes, loops, or exhausts memory.

---

### User Story 6 - Merge-state keywords are carried, not lost (Priority: P3)

A client developer reading graph nodes that carry `"@conflicts"` or `"@provenance"` can access them as structured data distinct from the triples, and writing the node back preserves them.

**Why this priority**: These keywords are produced by graph tooling (CORE §8.2–8.3). A codec that dropped them would destroy review and retraction state, but most clients only need triples, so it ranks below the core flows.

**Independent Test**: Decode the CORE §8.2 and §8.3 example front-matter and re-encode.

**Acceptance Scenarios**:

1. **Given** a node with an `"@conflicts"` map, **When** the client decodes it, **Then** the open alternatives are available per predicate and none of them appear as triples in the bag.
2. **Given** a node with an `"@provenance"` map, **When** the client decodes it, **Then** the attributions are available as document → predicate → values, and none of them appear as triples.
3. **Given** a node carrying either keyword, **When** the client decodes and re-encodes it, **Then** both keywords are written back unchanged, as quoted keys, in canonical order.

---

### Edge Cases

- A predicate appears in more than one position in the same file (e.g. both in front-matter and as an edge bullet): its values are merged into one set in the bag; on re-encode without role information, the predicate keeps the position of its first occurrence.
- Duplicate values of the same predicate (same value written twice, or the same edge target with different display text `[[Target|text]]`): the bag holds one triple — edges compare by target, literals by trimmed NFC-normalized text.
- Edge display text `[[Target|text]]`: the object is the target node; the display text is retained as presentation-only data and does not affect equality.
- A link-role block contains a bullet whose predicate differs from the block's other bullets: decoded as written (the bullet's own predicate wins; the heading carries no data).
- Front-matter contains a nested map other than `"@conflicts"`/`"@provenance"`: rejected as malformed, since meta values are scalars or arrays of scalars only.
- Empty file, or file containing only front-matter: rejected as malformed for the empty file; front-matter without body is rejected because the H1 title is mandatory.
- Front-matter key `@id` written bare instead of quoted: accepted on read for robustness if the parser can read it unambiguously; always written quoted.
- Fenced code blocks inside prose containing lines that look like headings or `predicate::` bullets: treated as prose, not structure.
- Unicode subjects that differ only by normalization form: treated as the same subject after NFC normalization.
- An empty bag passed with an empty subject: rejected at construction, since every node needs a valid subject and type.
- Paragraph boundaries are blank lines outside fenced code: a fenced code block with blank lines inside it stays within one paragraph value. *(BUG-001)*
- A text value that cannot be written as one paragraph (it contains a blank line outside fenced code) is not body-renderable and moves to front-matter like any other value its role cannot render; a paragraph value whose first line begins with `**Label**:` would be re-read as a label and is refused with the prose-structure error. *(BUG-001)*
- Paragraphs following a `**Label**` line (or a `**Label**:` paragraph) belong to that label until the next label or structural element. *(BUG-001)*
- The same untyped link written twice in the text of one predicate is listed once, at its first occurrence; untyped links inside fenced code are not listed. *(BUG-001)*
- Merge-state keywords (`@conflicts`, `@provenance`) are not attributes and keep their canonical order (CORE §8.5). *(BUG-001)*

## Requirements *(mandatory)*

### Functional Requirements

**Core vocabulary**

- **FR-001**: The library MUST expose a **triple** as an immutable value of subject, predicate and object.
- **FR-002**: The library MUST expose a **subject identity** type that can only hold a value satisfying CORE §5 (forbidden characters, no leading/trailing or double spaces), NFC-normalized; construction from an invalid string MUST fail with an invalid-identity error.
- **FR-003**: The library MUST expose a **predicate** name type that only holds camelCase names, and a **class** name type that only holds UpperCamelCase names; invalid names MUST fail construction.
- **FR-004**: The library MUST expose an **object** that is exactly one of: a *literal* (with its kind — text, number, boolean, date, datetime — and its exact written form) or a *reference* to another node's identity (with optional, presentation-only display text).
- **FR-005**: The library MUST expose a **node** as a subject identity, a class, a set of triples about that subject, and optional merge-state keywords (`@conflicts`, `@provenance`). `@id`, `@type`, `@conflicts` and `@provenance` MUST NOT appear as triples.
- **FR-006**: A node's triples MUST behave as a **set**: adding an equal triple twice yields one triple. Equality follows CORE §8.5 — literals by trimmed NFC text (dates/datetimes by value), references by target identity ignoring display text, prose after whitespace normalization.
- **FR-007**: Clients MUST be able to enumerate all triples of a node, enumerate the predicates in use, and retrieve all objects of a given predicate. Enumeration order MUST be deterministic ~~(canonical order, CORE §8.5)~~ and natural: the order in which triples entered the node — document order for a decoded node (meta → text → edge → link, as written within each section), insertion order for a node built in memory. A duplicate triple keeps the position of its first occurrence; predicates are ordered by their first triple. *(BUG-001)*
- **FR-008**: The library MUST expose the four predicate **roles** — meta, text, edge, link — plus, per predicate, an optional display label and, for text, whether it is the class's default text and whether its values are a list of statements.

**Decoding (Markdown → triples)**

- **FR-009**: The library MUST decode one Markdown node file into a node, mapping front-matter keys to meta triples, unlabelled prose to the default-text predicate, bold-labelled prose to its predicate, edge bullets (list, body and inline forms) to reference triples, and `## Label` blocks to the triples of the bullets they contain. Each prose paragraph MUST become its own text triple. A paragraph that begins with `**Label**:` MUST be decoded as labelled prose of that label's predicate, with the text after the colon as its first paragraph; a bold lead without a colon stays prose. *(paragraph and colon-label clauses: BUG-001)*
- **FR-010**: Decoding MUST record, for every predicate found, the role and label implied by the position it was found in, so that the node can be re-encoded without external role information.
- **FR-011**: Decoding MUST NOT produce triples from untyped links inside prose; the links stay inside the text value verbatim.
- **FR-026**: The library MUST expose, as a derived read-only view of a node and not as triples, the untyped links (`[[Target]]`, `[[Target|text]]`, CORE §7) found in its text values, each paired with the predicate of the text value that carries it, in node order, once per predicate and target, excluding fenced code and inline typed edges. *(BUG-001)*
- **FR-012**: Decoding MUST preserve every predicate, including predicates unknown to the client.
- **FR-013**: The predicate name used for unlabelled default prose MUST be configurable by the client; when not configured it defaults to `text`.
- **FR-014**: Decoding MUST reject: missing or unparseable front-matter, missing `"@id"` or `"@type"`, H1 not equal to `"@id"`, invalid identities or names, nested front-matter maps other than the merge-state keywords — each with a distinct, matchable error category.
- **FR-015**: Decoding MUST enforce a client-configurable maximum input size and bounded nesting, with a documented default.

**Encoding (triples → Markdown)**

- **FR-016**: The library MUST encode a node into a Markdown file laid out per CORE §4: front-matter (quoted `@`-keys, then meta predicates), H1 title equal to the subject, default prose, labelled prose, one heterogeneous edge block in list form, then one `## Label` block per link-role predicate and per list-of-statements text predicate. Each value of a text predicate is written as one paragraph; a labelled text predicate is written as its `**Label**` line followed by its paragraphs. *(paragraph clause: BUG-001)*
- **FR-017**: The client MUST be able to supply role information for predicates at encode time; for predicates without client-supplied roles, the library MUST use the role recorded at decode time; for predicates with neither, it MUST default to meta for literals and edge for references.
- **FR-018**: Encoding MUST produce canonical form: ~~values of multi-valued predicates in lexicographic order of their canonical rendering (dates by value)~~ predicates and their values written in node order (FR-007) within each CORE §4 section, edges in list form, ~~deterministic ordering of predicates within each block~~, so two encodings of ~~equal nodes~~ the same node, or of equal nodes whose triples entered in the same order, are byte-identical. *(BUG-001: CORE §3 requires only a stable order)*
- **FR-019**: When a predicate has no declared label, encoding MUST derive it by splitting the predicate name at camelCase boundaries and capitalizing each word.
- **FR-020**: Encoding MUST refuse prose values containing ATX heading lines or lines beginning with a `predicate::` bullet (outside fenced code), or beginning with a `**Label**:` lead, returning a matchable error. *(`**Label**:` clause: BUG-001)*
- **FR-021**: Encoding MUST write `@conflicts` and `@provenance`, when present, back into front-matter as quoted keys in canonical order.

**Round trip & robustness**

- **FR-022**: For every canonical node file, decode followed by encode (with no client role information) MUST reproduce the file byte-for-byte.
- **FR-023**: For every valid node, encode followed by decode MUST reproduce an equal node (same subject, class, triple set and merge-state keywords), with the relative order of triples within each CORE §4 section preserved. *(order clause: BUG-001)*
- **FR-024**: The library MUST NOT crash on any input; every failure MUST be reported as a returned error whose category the client can test for without inspecting message text.
- **FR-025**: Decoding and encoding MUST be free of side effects: no filesystem, network, environment or global state access; the client supplies the input stream and output destination.

### Key Entities

- **Subject (Identity)**: the `@id` of a node — the name of the thing described; validated against CORE §5, NFC-normalized; equal to the file basename and H1 (basename check is the client's concern).
- **Class**: the `@type` of a node; UpperCamelCase name.
- **Predicate**: a camelCase property name; the "verb" of a triple.
- **Object**: the value of a triple — either a *literal* (kind + written form) or a *reference* to another node's identity (+ optional display text).
- **Triple**: one statement `(subject, predicate, object)`.
- **Node (Bag)**: one Markdown file seen as data — subject, class, set of triples, optional merge-state keywords, and the observed placement of each predicate.
- **Role**: where a predicate renders — `meta`, `text`, `edge`, `link` — with optional label and text sub-attributes (default text, list of statements).
- **Merge-state keywords**: `@conflicts` (predicate → open alternatives) and `@provenance` (document → predicate → values); carried beside triples, never as triples.
- **Untyped link**: a bare `[[Target]]` inside a text value; part of that value, never a triple, exposed as a derived view (FR-026). *(BUG-001)*
- **Paragraph**: the unit of a prose text value; one paragraph, one triple. *(BUG-001)*

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of the canonical example nodes in CORE.md (§3, §8.2, §8.3) round-trip Markdown → triples → Markdown byte-identically.
- **SC-002**: 100% of a generated corpus of at least 1,000 random valid nodes round-trip triples → Markdown → triples to an equal bag.
- **SC-003**: A client developer can decode a node file and list its triples using no more than 3 library calls, demonstrated by a published runnable example.
- **SC-004**: Continuous random-input testing of the decoder for at least 10 minutes produces zero crashes, hangs, or unbounded memory growth.
- **SC-005**: Every malformed-input category in User Story 5 is distinguishable by the client without reading error message text (100% of categories testable by category).
- **SC-006**: Decoding a typical node file (≤ 10 KB, ≤ 100 statements) completes in under 1 millisecond on commodity hardware, so a graph of 10,000 nodes can be read in seconds.
- **SC-007**: Every acceptance scenario in this specification has a passing automated test exercising only the library's public surface.
- **SC-008**: For 100% of the generated round-trip corpus (SC-002), the relative order of triples within each CORE §4 section survives triples → Markdown → triples. *(BUG-001)*

## Assumptions

- **Scope is a single node file.** Graph-level concerns — folder layout, graph-wide uniqueness, alias resolution, link resolution, schema loading, merge/apply/combine/retract (CORE §6, §8.4–8.11), git integration (§9) and the Document Patch exchange format (§10) — are out of scope for this feature and may build on these types later.
- **Roles come from the client.** The library does not read `_schema/Property` nodes; the client supplies roles (from its schema or hard-coded) when encoding. On decode, roles are inferred from position, which CORE §4 makes unambiguous (the mandatory `## Label` heading separates link blocks from the edge block).
- **Merge-state keywords are carried, not interpreted.** `@conflicts` and `@provenance` are decoded into structured data and written back, but the library applies no merge semantics and does not validate them against candidate sets.
- **Literal kinds** are inferred from the front-matter scalar as written (text, number, boolean, date, datetime); prose and statement values are text. Universal timestamps (creation/modification time) are ordinary meta predicates with datetime literals.
- **Default text predicate** is named `text` unless the client configures otherwise, matching the CORE §3 illustration.
- **Inline bold labels** are recognised only with a colon (`**Note**: text`); the decoder does not consult the schema to decide whether a bold lead is a label. *(BUG-001)*
- **Natural order** satisfies CORE §3: the written order of values "carries no meaning" and a producer "SHOULD write them in a stable order"; document/insertion order is that stable order. Lexicographic order is not imposed. *(BUG-001)*
- **Bold-label → predicate mapping on decode**: a bold `**Label**` paragraph maps to the predicate whose label it is when the client supplies one; otherwise to the camelCase join of the label words (`**Key Findings**` → `keyFindings`).
- **Display text** of edges is retained for presentation but is not part of triple equality (CORE §7).
- **File basename vs. `@id`**: the library receives content, not paths; checking that the basename matches `@id` is left to the client.
- **Default input size limit** is a documented, generous value (on the order of a few megabytes) adjustable by the client.
