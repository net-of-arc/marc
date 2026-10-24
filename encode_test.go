//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/net-of-arc/marc"
)

func node(t *testing.T, id string, triples ...marc.Triple) marc.Node {
	t.Helper()
	n, err := marc.NewNode(marc.MustID(id), marc.MustClass("Entity"), slices.Values(triples))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func codec(t *testing.T, schema marc.Schema) *marc.Codec {
	t.Helper()
	c, err := marc.New(marc.Config{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func coreSchema() marc.Properties {
	return marc.Properties{
		marc.MustPredicate("category"):    {Role: marc.RoleMeta},
		marc.MustPredicate("tags"):        {Role: marc.RoleMeta, Many: true},
		marc.MustPredicate("text"):        {Role: marc.RoleText, Default: true},
		marc.MustPredicate("replaces"):    {Role: marc.RoleEdge},
		marc.MustPredicate("conformsTo"):  {Role: marc.RoleEdge},
		marc.MustPredicate("mentionedIn"): {Role: marc.RoleLink},
	}
}

// coreTriples lists the CORE §3 statements in the same order as
// testdata/core-3-example.md: meta, then default text, then edges, then
// the link block (BUG-001: encoding keeps natural, not canonical, order).
func coreTriples() iter.Seq[marc.Triple] {
	return slices.Values([]marc.Triple{
		triple(tls, "category", marc.Text("script")),
		triple(tls, "tags", marc.Text("cryptography")),
		triple(tls, "text", marc.Text("A cryptographic protocol that establishes an authenticated, confidential channel over an untrusted network.")),
		triple(tls, "conformsTo", ref("RFC 8446")),
		triple(tls, "replaces", ref("SSL Protocol")),
		triple(tls, "mentionedIn", ref("rescorla-2026-tls13")),
	})
}

func TestEncodeCoreExample(t *testing.T) {
	// Scenario 2.1 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n, err := marc.NewNode(marc.MustID(tls), marc.MustClass("Entity"), coreTriples())
	it.Then(t).Must(it.Nil(err))

	b, err := codec(t, coreSchema()).Marshal(n)

	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(string(b), string(fixture(t, "testdata/core-3-example.md"))),
	)
}

func TestEncodeInsertionOrderValues(t *testing.T) {
	// Scenario 2.2 from specs/001-markdown-triples-codec/spec.md (BUG-001: natural, not lexicographic, order)
	t.Parallel()

	n := node(t, "X",
		triple("X", "tags", marc.Text("zeta")),
		triple("X", "tags", marc.Text("alpha")),
		triple("X", "tags", marc.Text("mu")),
	)
	schema := marc.Properties{marc.MustPredicate("tags"): {Role: marc.RoleMeta, Many: true}}

	b, err := codec(t, schema).Marshal(n)

	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\ntags: [zeta, alpha, mu]\n---\n# X\n"),
	)
}

func TestEncodeDerivedLinkLabel(t *testing.T) {
	// Scenario 2.3 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n := node(t, "X", triple("X", "mentionedIn", ref("Y")))
	schema := marc.Properties{marc.MustPredicate("mentionedIn"): {Role: marc.RoleLink}}

	b, err := codec(t, schema).Marshal(n)

	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\n## Mentioned In\n- mentionedIn:: [[Y]]\n"),
	)
}

func TestEncodeDeclaredLinkLabel(t *testing.T) {
	// Scenario 2.4 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n := node(t, "X", triple("X", "citedBy", ref("Y")))
	schema := marc.Properties{marc.MustPredicate("citedBy"): {Role: marc.RoleLink, Label: "Sources"}}

	b, err := codec(t, schema).Marshal(n)

	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\n## Sources\n- citedBy:: [[Y]]\n"),
	)
}

func TestEncodeLabelledProse(t *testing.T) {
	// Scenario 2.5 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n := node(t, "X",
		triple("X", "text", marc.Text("Default prose.")),
		triple("X", "keyFindings", marc.Text("A finding.")),
		triple("X", "replaces", ref("Y")),
	)
	schema := marc.Properties{
		marc.MustPredicate("text"):        {Role: marc.RoleText, Default: true},
		marc.MustPredicate("keyFindings"): {Role: marc.RoleText},
		marc.MustPredicate("replaces"):    {Role: marc.RoleEdge},
	}

	b, err := codec(t, schema).Marshal(n)

	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\nDefault prose.\n\n**Key Findings**\nA finding.\n\n- replaces:: [[Y]]\n"),
	)
}

func TestEncodeMultiParagraphText(t *testing.T) {
	// Scenario 2.7 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n := node(t, "X",
		triple("X", "text", marc.Text("abc")),
		triple("X", "text", marc.Text("def")),
		triple("X", "note", marc.Text("First.")),
		triple("X", "note", marc.Text("Second.")),
	)
	schema := marc.Properties{
		marc.MustPredicate("text"): {Role: marc.RoleText, Default: true},
		marc.MustPredicate("note"): {Role: marc.RoleText},
	}

	b, err := codec(t, schema).Marshal(n)

	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\nabc\n\ndef\n\n**Note**\nFirst.\n\nSecond.\n"),
	)
}

func TestEncodeEmptyNode(t *testing.T) {
	// Scenario 2.6 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	b, err := marc.Marshal(node(t, "X"))

	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n"),
	)
}
