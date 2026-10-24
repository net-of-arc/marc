//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/fogfish/it/v2"
)

func tripleOf(id, p string, o Object) Triple {
	return Triple{Subject: MustID(id), Predicate: MustPredicate(p), Object: o}
}

func TestNewNodeSetSemantics(t *testing.T) {
	n, err := NewNode(MustID("X"), MustClass("Entity"), slices.Values([]Triple{
		tripleOf("X", "tags", Text("b")),
		tripleOf("X", "tags", Text(" b ")),
		tripleOf("X", "tags", Text("a")),
		tripleOf("X", "edge", RefTo(MustID("Y")).WithDisplay("first")),
		tripleOf("X", "edge", RefTo(MustID("Y")).WithDisplay("second")),
	}))

	it.Then(t).Must(it.Nil(err)).Should(
		it.Seq(slices.Collect(n.Triples())).Equal(
			tripleOf("X", "edge", RefTo(MustID("Y")).WithDisplay("first")),
			tripleOf("X", "tags", Text("a")),
			tripleOf("X", "tags", Text("b")),
		),
		it.Seq(slices.Collect(n.Predicates())).Equal(MustPredicate("edge"), MustPredicate("tags")),
		it.Seq(slices.Collect(n.Values(MustPredicate("tags")))).Equal(Text("a"), Text("b")),
	)
}

func TestNewNodeRejects(t *testing.T) {
	id, class := MustID("X"), MustClass("Entity")
	for name, tc := range map[string]struct {
		id      ID
		class   Class
		triple  Triple
		expectd error
	}{
		"foreign subject": {id, class, tripleOf("Y", "p", Text("v")), ErrForeignSubject},
		"zero predicate":  {id, class, Triple{Subject: id, Object: Text("v")}, ErrInvalidName},
		"nil object":      {id, class, Triple{Subject: id, Predicate: MustPredicate("p")}, ErrInvalidObject},
		"zero ref":        {id, class, tripleOf("X", "p", Ref{}), ErrInvalidObject},
		"zero number":     {id, class, tripleOf("X", "p", Number{}), ErrInvalidObject},
		"zero date":       {id, class, tripleOf("X", "p", Date{}), ErrInvalidObject},
		"zero datetime":   {id, class, tripleOf("X", "p", DateTime{}), ErrInvalidObject},
		"invalid utf8":    {id, class, tripleOf("X", "p", Text("\xff")), ErrInvalidObject},
		"display break":   {id, class, tripleOf("X", "p", RefTo(MustID("Y")).WithDisplay("a\nb")), ErrInvalidObject},
		"display link":    {id, class, tripleOf("X", "p", RefTo(MustID("Y")).WithDisplay("a]]b")), ErrInvalidObject},
		"zero id":         {ID{}, class, tripleOf("X", "p", Text("v")), ErrInvalidID},
		"zero class":      {id, Class{}, tripleOf("X", "p", Text("v")), ErrInvalidName},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewNode(tc.id, tc.class, slices.Values([]Triple{tc.triple}))
			it.Then(t).Should(it.True(errors.Is(err, tc.expectd)))
		})
	}
}

func TestNodeIteratorsStopEarly(t *testing.T) {
	n, err := NewNode(MustID("X"), MustClass("Entity"), slices.Values([]Triple{
		tripleOf("X", "a", Text("1")), tripleOf("X", "a", Text("2")), tripleOf("X", "b", Text("3")),
	}))
	it.Then(t).Must(it.Nil(err))

	count := 0
	for range n.Triples() {
		count++
		break
	}
	values := 0
	for range n.Values(MustPredicate("a")) {
		values++
		break
	}
	predicates := 0
	for range n.Predicates() {
		predicates++
		break
	}
	it.Then(t).Should(it.Equal(count, 1), it.Equal(values, 1), it.Equal(predicates, 1))
}

func TestNodeConcurrentIteration(t *testing.T) {
	n, err := NewNode(MustID("X"), MustClass("Entity"), slices.Values([]Triple{
		tripleOf("X", "a", Text("1")), tripleOf("X", "b", RefTo(MustID("Y"))),
	}))
	it.Then(t).Must(it.Nil(err))

	var wg sync.WaitGroup
	counts := make([]int, 8)
	for i := range counts {
		wg.Go(func() {
			for range n.Triples() {
				counts[i]++
			}
		})
	}
	wg.Wait()
	it.Then(t).Should(it.Seq(counts).Equal(2, 2, 2, 2, 2, 2, 2, 2))
}

func TestZeroNodeIsEmpty(t *testing.T) {
	var n Node
	it.Then(t).Should(
		it.True(n.ID().IsZero()),
		it.Seq(slices.Collect(n.Triples())).BeEmpty(),
	)
}
