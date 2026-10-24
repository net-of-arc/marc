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
// Triples form a set under CORE §8.5 equality; Triples, Predicates and Values
// yield in natural order — the order triples entered the node (document
// order for a decoded node, the sequence order given to NewNode otherwise),
// with a duplicate keeping the position of its first occurrence. Conflicts
// and Provenance keep canonical order (CORE §8.5), since they are merge
// state, not attributes. The zero Node is invalid: it has no triples and the
// encoder rejects it with ErrMalformed. Node is immutable and safe for
// concurrent use; its iterators read shared state, never mutate it, and yield
// immutable values.
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

// Triples yields every triple of the node in natural order.
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

// Predicates yields every predicate in use, in natural order: each one at
// the position of its first triple.
func (n Node) Predicates() iter.Seq[Predicate] {
	return func(yield func(Predicate) bool) {
		for _, set := range n.facts {
			if !yield(set.predicate) {
				return
			}
		}
	}
}

// Values yields the objects of predicate p in natural order; it yields
// nothing when the node does not carry p.
func (n Node) Values(p Predicate) iter.Seq[Object] {
	return slices.Values(lookup(n.facts, p))
}

// Links yields the untyped links ([[Target]] or [[Target|display]], CORE §7)
// found in the node's Text values, each paired with the predicate of the
// value that carries it, in natural order. A link inside fenced code, or
// inside the typed inline form "[p:: [[T]]]", is not yielded. The same
// (predicate, target) pair is yielded once, at its first occurrence. Links
// are derived on demand and never appear as triples (FR-026).
func (n Node) Links() iter.Seq2[Predicate, Ref] {
	return func(yield func(Predicate, Ref) bool) {
		seen := map[string]struct{}{}
		for _, set := range n.facts {
			for _, o := range set.values {
				t, ok := o.(Text)
				if !ok {
					continue
				}
				for _, ref := range untypedLinks(string(t)) {
					key := set.predicate.value + "\x00" + ref.target.value
					if _, dup := seen[key]; dup {
						continue
					}
					seen[key] = struct{}{}
					if !yield(set.predicate, ref) {
						return
					}
				}
			}
		}
	}
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
	for _, s := range sets {
		if s.predicate == p {
			return s.values
		}
	}
	return nil
}

type builder struct {
	sets     map[Predicate][]Object
	order    []Predicate
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
	if _, ok := b.sets[p]; !ok {
		b.order = append(b.order, p)
	}
	b.sets[p] = append(b.sets[p], o)
	return true
}

func (b *builder) observe(p Predicate, prop Property) {
	if _, ok := b.observed[p]; !ok {
		b.observed[p] = prop
	}
}

// valueSets returns the facts in natural order: each predicate at the
// position of its first value, each predicate's values in the order they
// were added.
func (b *builder) valueSets() []valueSet {
	sets := make([]valueSet, 0, len(b.order))
	for _, p := range b.order {
		sets = append(sets, valueSet{predicate: p, values: b.sets[p]})
	}
	return sets
}

// canonicalValueSets returns the facts in CORE §8.5 canonical order. It is
// used for merge-state keywords ("@conflicts"), never for triples.
func (b *builder) canonicalValueSets() []valueSet {
	sets := make([]valueSet, 0, len(b.order))
	for _, p := range b.order {
		values := slices.Clone(b.sets[p])
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
