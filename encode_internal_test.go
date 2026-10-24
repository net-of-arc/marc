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
	"testing"

	"github.com/fogfish/it/v2"
)

func nodeOf(t *testing.T, triples ...Triple) Node {
	t.Helper()
	n, err := NewNode(MustID("X"), MustClass("Entity"), slices.Values(triples))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func marshal(t *testing.T, schema Schema, n Node) (string, error) {
	t.Helper()
	c, err := New(Config{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Marshal(n)
	return string(b), err
}

const head = "---\n\"@id\": X\n\"@type\": Entity\n"

func TestEncodePropertyLookupOrder(t *testing.T) {
	decoded, err := Unmarshal([]byte(head + "---\n# X\n\n- p:: [[Y]]\n"))
	it.Then(t).Must(it.Nil(err))

	fromObserved, errObserved := marshal(t, nil, decoded)
	fromSchema, errSchema := marshal(t, Properties{MustPredicate("p"): {Role: RoleMeta}}, decoded)
	fallbackRef, _ := marshal(t, nil, nodeOf(t, tripleOf("X", "p", RefTo(MustID("Y")))))
	fallbackLiteral, _ := marshal(t, nil, nodeOf(t, tripleOf("X", "p", Text("v"))))

	it.Then(t).Should(
		it.Nil(errObserved), it.Nil(errSchema),
		it.Equal(fromObserved, head+"---\n# X\n\n- p:: [[Y]]\n"),
		it.Equal(fromSchema, head+"p: \"[[Y]]\"\n---\n# X\n"),
		it.Equal(fallbackRef, head+"---\n# X\n\n- p:: [[Y]]\n"),
		it.Equal(fallbackLiteral, head+"p: v\n---\n# X\n"),
	)
}

func TestEncodeInvalidProperty(t *testing.T) {
	n := nodeOf(t, tripleOf("X", "p", Text("v")))
	for _, prop := range []Property{
		{Role: Role(7)},
		{Role: RoleEdge, Default: true},
		{Role: RoleText, Default: true, Many: true},
	} {
		_, err := marshal(t, Properties{MustPredicate("p"): prop}, n)
		it.Then(t).Should(it.True(errors.Is(err, ErrInvalidProperty)))
	}
}

func TestEncodeLayout(t *testing.T) {
	n := nodeOf(t,
		tripleOf("X", "zLink", RefTo(MustID("B"))),
		tripleOf("X", "aLink", RefTo(MustID("A"))),
		tripleOf("X", "replaces", RefTo(MustID("B"))),
		tripleOf("X", "conformsTo", RefTo(MustID("Z"))),
		tripleOf("X", "conformsTo", RefTo(MustID("A")).WithDisplay("a")),
		tripleOf("X", "claims", Text("second")),
		tripleOf("X", "claims", Text("first")),
		tripleOf("X", "notes", Text("Notes.")),
		tripleOf("X", "text", Text("Prose.")),
		tripleOf("X", "count", mustNumber("3")),
	)
	schema := Properties{
		MustPredicate("zLink"):      {Role: RoleLink},
		MustPredicate("aLink"):      {Role: RoleLink, Label: "Alpha"},
		MustPredicate("replaces"):   {Role: RoleEdge},
		MustPredicate("conformsTo"): {Role: RoleEdge},
		MustPredicate("claims"):     {Role: RoleText, Many: true},
		MustPredicate("notes"):      {Role: RoleText},
		MustPredicate("text"):       {Role: RoleText, Default: true},
	}

	b, err := marshal(t, schema, n)

	it.Then(t).Must(it.Nil(err)).Should(it.Equal(b, head+"count: 3\n---\n# X\n\nProse.\n\n**Notes**\nNotes.\n\n"+
		"- conformsTo:: [[A|a]]\n- conformsTo:: [[Z]]\n- replaces:: [[B]]\n\n"+
		"## Alpha\n- aLink:: [[A]]\n\n## Claims\n- claims:: first\n- claims:: second\n\n## Z Link\n- zLink:: [[B]]\n"))
}

func TestEncodeSpillsToMeta(t *testing.T) {
	n := nodeOf(t,
		tripleOf("X", "replaces", RefTo(MustID("Y"))),
		tripleOf("X", "replaces", mustNumber("1")),
		tripleOf("X", "text", Text("\nleading blank line")),
		tripleOf("X", "claims", Text("multi\nline")),
	)
	schema := Properties{
		MustPredicate("replaces"): {Role: RoleEdge},
		MustPredicate("text"):     {Role: RoleText, Default: true},
		MustPredicate("claims"):   {Role: RoleText, Many: true},
	}

	b, err := marshal(t, schema, n)

	it.Then(t).Must(it.Nil(err)).Should(it.Equal(b, head+
		"claims: \"multi\\nline\"\nreplaces: 1\ntext: \"\\nleading blank line\"\n---\n# X\n\n- replaces:: [[Y]]\n"))
}

func TestEncodeZeroNode(t *testing.T) {
	_, err := Marshal(Node{})
	it.Then(t).Should(it.True(errors.Is(err, ErrMalformed)))
}

func TestEncodeProseStructure(t *testing.T) {
	for text, rejected := range map[string]bool{
		"# Title":                 true,
		"a\n###### Deep":          true,
		"a\n- p:: x":              true,
		"a\n- p::x":               true,
		"a\nconformsTo:: [[Y]]":   true,
		"a\n**Label**\nb":         true,
		"a\n```\nunterminated":    true,
		"a\n```\n# Title\n```\nb": false,
		"a\n~~~\n- p:: x\n~~~":    false,
		"a\n#hashtag":             false,
		"a\n- item":               false,
		"a\nnote:: plain":         false,
		"**Bold** in a sentence.": false,
	} {
		it.Then(t).Should(it.Equal(proseStructure(text), rejected))
	}
}
