//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"cmp"
	"iter"
	"slices"
	"strings"
	"unicode/utf8"
)

// Node is one Markdown node file seen as data: a subject identity, a class, a
// set of triples about the subject, the merge-state keywords "@conflicts" and
// "@provenance", and the Property observed for each predicate while decoding.
//
// Triples form a set under CORE §8.5 equality; every iterator yields in
// canonical order (predicates by code point, then values by variant and
// value). The zero Node is invalid: it has no triples and the encoder rejects
// it with ErrMalformed. Node is immutable and safe for concurrent use; its
// iterators read shared state, never mutate it, and yield immutable values.
type Node struct {
	id         ID
	class      Class
	facts      []valueSet
	observed   map[Predicate]Property
	conflicts  []valueSet
	provenance []Attribution
}

type valueSet struct {
	predicate Predicate
	values    []Object
}

// NewNode builds a node from its identity, its class and a sequence of
// triples. It consumes triples exactly once and retains no reference to the
// sequence. Equal triples collapse into one; the first display text of equal
// references wins. A node built this way has no observed properties and no
// merge-state keywords.
//
// It returns ErrInvalidID for a zero id, ErrInvalidName for a zero class or
// predicate, ErrForeignSubject for a triple about another subject, and
// ErrInvalidObject for a nil or zero object, a text that is not valid UTF-8,
// or a display text containing a line break, "[[" or "]]".
func NewNode(id ID, class Class, triples iter.Seq[Triple]) (Node, error) {
	if err := validHead(id, class); err != nil {
		return Node{}, err
	}
	b := newBuilder()
	for t := range triples {
		if err := validTriple(id, t); err != nil {
			return Node{}, err
		}
		b.add(t.Predicate, t.Object)
	}
	return b.node(id, class), nil
}

func validHead(id ID, class Class) error {
	if id.IsZero() {
		return ErrInvalidID.With(detail(`"@id"`))
	}
	if class.IsZero() {
		return ErrInvalidName.With(detail(`"@type"`))
	}
	return nil
}

func validTriple(id ID, t Triple) error {
	switch {
	case t.Subject != id:
		return ErrForeignSubject.With(detail("subject"))
	case t.Predicate.IsZero():
		return ErrInvalidName.With(detail("predicate"))
	case zeroObject(t.Object) || !validPayload(t.Object):
		return ErrInvalidObject.With(detail(t.Predicate.value))
	}
	return nil
}

func validPayload(o Object) bool {
	switch v := o.(type) {
	case Text:
		return utf8.ValidString(string(v))
	case Ref:
		return utf8.ValidString(v.display) && !strings.ContainsAny(v.display, "\n\r") &&
			!strings.Contains(v.display, "[[") && !strings.Contains(v.display, "]]")
	}
	return true
}

// ID returns the subject identity of the node.
func (n Node) ID() ID { return n.id }

// Class returns the "@type" of the node.
func (n Node) Class() Class { return n.class }

// Triples yields every triple of the node in canonical order.
func (n Node) Triples() iter.Seq[Triple] {
	return func(yield func(Triple) bool) {
		for _, set := range n.facts {
			for _, o := range set.values {
				if !yield(Triple{Subject: n.id, Predicate: set.predicate, Object: o}) {
					return
				}
			}
		}
	}
}

// Predicates yields every predicate in use, in code-point order.
func (n Node) Predicates() iter.Seq[Predicate] {
	return func(yield func(Predicate) bool) {
		for _, set := range n.facts {
			if !yield(set.predicate) {
				return
			}
		}
	}
}

// Values yields the objects of predicate p in canonical order; it yields
// nothing when the node does not carry p.
func (n Node) Values(p Predicate) iter.Seq[Object] {
	return slices.Values(lookup(n.facts, p))
}

// Property returns the Property observed for p while decoding: the role and
// label implied by the position p was first found in. It reports false for a
// predicate that was not decoded, including every predicate of a node built
// with NewNode.
func (n Node) Property(p Predicate) (Property, bool) {
	prop, ok := n.observed[p]
	return prop, ok
}

// Conflicts yields the open alternatives recorded under "@conflicts", as
// (predicate, alternative) pairs in canonical order. They are not triples.
func (n Node) Conflicts() iter.Seq2[Predicate, Object] {
	return func(yield func(Predicate, Object) bool) {
		for _, set := range n.conflicts {
			for _, o := range set.values {
				if !yield(set.predicate, o) {
					return
				}
			}
		}
	}
}

// Provenance yields the attributions recorded under "@provenance", ordered by
// document, predicate and value. They are not triples.
func (n Node) Provenance() iter.Seq[Attribution] {
	return slices.Values(n.provenance)
}

func lookup(sets []valueSet, p Predicate) []Object {
	i, ok := slices.BinarySearchFunc(sets, p, func(s valueSet, p Predicate) int {
		return cmp.Compare(s.predicate.value, p.value)
	})
	if !ok {
		return nil
	}
	return sets[i].values
}

type builder struct {
	sets     map[Predicate][]Object
	seen     map[string]struct{}
	observed map[Predicate]Property
}

func newBuilder() *builder {
	return &builder{
		sets:     map[Predicate][]Object{},
		seen:     map[string]struct{}{},
		observed: map[Predicate]Property{},
	}
}

func (b *builder) add(p Predicate, o Object) bool {
	k := p.value + "\x00" + o.key()
	if _, ok := b.seen[k]; ok {
		return false
	}
	b.seen[k] = struct{}{}
	b.sets[p] = append(b.sets[p], o)
	return true
}

func (b *builder) observe(p Predicate, prop Property) {
	if _, ok := b.observed[p]; !ok {
		b.observed[p] = prop
	}
}

func (b *builder) valueSets() []valueSet {
	sets := make([]valueSet, 0, len(b.sets))
	for p, values := range b.sets {
		slices.SortFunc(values, compareObjects)
		sets = append(sets, valueSet{predicate: p, values: values})
	}
	slices.SortFunc(sets, func(a, b valueSet) int { return comparePredicates(a.predicate, b.predicate) })
	return sets
}

func (b *builder) node(id ID, class Class) Node {
	return Node{id: id, class: class, facts: b.valueSets(), observed: b.observed}
}

func compareAttributions(a, b Attribution) int {
	return cmp.Or(
		cmp.Compare(a.Document.value, b.Document.value),
		cmp.Compare(a.Predicate.value, b.Predicate.value),
		compareObjects(a.Object, b.Object),
	)
}
