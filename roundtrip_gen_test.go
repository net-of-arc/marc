//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/net-of-arc/marc"
)

type generator struct{ *rand.Rand }

func (g generator) pick(s string) string {
	r := []rune(s)
	return string(r[g.IntN(len(r))])
}

func (g generator) word(alphabet string, max int) string {
	var sb strings.Builder
	for range 1 + g.IntN(max) {
		sb.WriteString(g.pick(alphabet))
	}
	return sb.String()
}

func (g generator) words(n int) string {
	ws := make([]string, 1+g.IntN(n))
	for i := range ws {
		ws[i] = g.word("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789é", 8)
	}
	return strings.Join(ws, " ")
}

func (g generator) id() marc.ID {
	return marc.MustID(g.word("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-()é", 6) + " " + g.word("abcXYZ0123", 4))
}

func (g generator) predicate() marc.Predicate {
	return marc.MustPredicate(g.word("abcdefghijklmnopqrstuvwxyz", 6) + g.word("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789", 6))
}

func (g generator) ref() marc.Object {
	r := marc.RefTo(g.id())
	if g.IntN(3) == 0 {
		return r.WithDisplay(g.words(3))
	}
	return r
}

func (g generator) literal() marc.Object {
	switch g.IntN(7) {
	case 0:
		n, _ := marc.ParseNumber(strconv.Itoa(g.IntN(2000) - 1000))
		return n
	case 1:
		n, _ := marc.ParseNumber(strconv.FormatFloat(g.NormFloat64()*1e3, 'g', -1, 64))
		return n
	case 2:
		return marc.Bool(g.IntN(2) == 0)
	case 3:
		return marc.DateOf(time.Unix(g.Int64N(4e9), 0))
	case 4:
		return marc.DateTimeOf(time.Unix(g.Int64N(4e9), g.Int64N(1e9)))
	case 5:
		return g.ref()
	}
	return marc.Text(g.word("ab Z09:#,-[]{}\"'\\|>!&*%@`~?.+<=\t\n\u00a0é🙂", 16))
}

func (g generator) prose() marc.Object {
	paragraphs := make([]string, 1+g.IntN(3))
	for i := range paragraphs {
		paragraphs[i] = g.words(6) + "\n" + g.words(6) + " [[" + g.id().String() + "]]."
	}
	return marc.Text(strings.Join(paragraphs, "\n\n"))
}

func (g generator) values(prop marc.Property) []marc.Object {
	values := make([]marc.Object, 1+g.IntN(3))
	for i := range values {
		switch {
		case prop.Role == marc.RoleEdge || prop.Role == marc.RoleLink:
			values[i] = g.ref()
		case prop.Role == marc.RoleText && prop.Many:
			values[i] = marc.Text(g.words(5))
		case prop.Role == marc.RoleText:
			return []marc.Object{g.prose()}
		default:
			values[i] = g.literal()
		}
	}
	return values
}

func (g generator) property() marc.Property {
	switch g.IntN(6) {
	case 0:
		return marc.Property{Role: marc.RoleMeta, Many: true}
	case 1:
		return marc.Property{Role: marc.RoleText}
	case 2:
		return marc.Property{Role: marc.RoleText, Many: true}
	case 3:
		return marc.Property{Role: marc.RoleEdge}
	case 4:
		return marc.Property{Role: marc.RoleLink}
	}
	return marc.Property{Role: marc.RoleMeta}
}

func (g generator) node() (marc.Node, marc.Properties, error) {
	id := g.id()
	schema := marc.Properties{marc.MustPredicate("text"): {Role: marc.RoleText, Default: true}}
	triples := []marc.Triple{{Subject: id, Predicate: marc.MustPredicate("text"), Object: g.prose()}}
	for range g.IntN(8) {
		p, prop := g.predicate(), g.property()
		if _, exists := schema[p]; exists {
			continue
		}
		schema[p] = prop
		for _, o := range g.values(prop) {
			triples = append(triples, marc.Triple{Subject: id, Predicate: p, Object: o})
		}
	}
	n, err := marc.NewNode(id, marc.MustClass("Entity"), slices.Values(triples))
	return n, schema, err
}

func TestRoundTripGenerated(t *testing.T) {
	// Scenario 3.2 from specs/001-markdown-triples-codec/spec.md (SC-002)
	t.Parallel()

	g := generator{rand.New(rand.NewPCG(2026, 10))}
	for range 1000 {
		n, schema, err := g.node()
		it.Then(t).Must(it.Nil(err), it.Nil(schema.Validate()))

		c := codec(t, schema)
		b, err := c.Marshal(n)
		it.Then(t).Must(it.Nil(err))

		m, err := c.Unmarshal(b)
		it.Then(t).Must(it.Nil(err)).Should(
			it.Equal(m.ID(), n.ID()),
			it.Seq(slices.Collect(m.Triples())).Equal(slices.Collect(n.Triples())...),
		)
	}
}
