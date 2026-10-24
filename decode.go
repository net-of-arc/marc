//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"bytes"
	"slices"
	"unicode/utf8"
)

func (c *Codec) decode(src []byte) (Node, error) {
	if !utf8.Valid(src) {
		return Node{}, ErrMalformed.With(detail("input is not UTF-8"))
	}
	front, body, err := splitFrontMatter(src)
	if err != nil {
		return Node{}, err
	}
	h, err := parseFrontMatter(front)
	if err != nil {
		return Node{}, err
	}
	sc := scope{id: h.id, class: h.class, schema: c.schema, line: bytes.Count(front, []byte("\n")) + 3}
	entries, err := scanBody(string(body), sc)
	if err != nil {
		return Node{}, err
	}
	return assemble(h, entries), nil
}

func assemble(h header, entries []entry) Node {
	facts := newBuilder()
	for _, e := range slices.Concat(h.meta, entries) {
		addEntry(facts, e)
	}
	n := facts.node(h.id, h.class)
	conflicts := newBuilder()
	for _, e := range h.conflicts {
		addEntry(conflicts, e)
	}
	n.conflicts = conflicts.canonicalValueSets()
	n.provenance = canonicalProvenance(h.provenance)
	return n
}

func addEntry(b *builder, e entry) {
	if len(e.values) == 0 {
		return
	}
	b.observe(e.predicate, e.property)
	for _, o := range e.values {
		b.add(e.predicate, o)
	}
}

func canonicalProvenance(attributions []Attribution) []Attribution {
	sorted := slices.SortedFunc(slices.Values(attributions), compareAttributions)
	return slices.CompactFunc(sorted, func(a, b Attribution) bool {
		return a.Document == b.Document && a.Predicate == b.Predicate && a.Object.key() == b.Object.key()
	})
}
