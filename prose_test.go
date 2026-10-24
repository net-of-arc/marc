//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"slices"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/net-of-arc/marc"
)

func proseNode(t *testing.T) marc.Node {
	t.Helper()
	n, err := marc.Unmarshal(fixture(t, "testdata/prose-untyped-links.md"))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func defaultTexts(n marc.Node) []string {
	var out []string
	for o := range n.Values(marc.MustPredicate("text")) {
		out = append(out, o.String())
	}
	return out
}

func TestProseUntypedLink(t *testing.T) {
	// Scenario 4.1 from specs/001-markdown-triples-codec/spec.md (BUG-001:
	// one text triple per paragraph)
	t.Parallel()

	n := proseNode(t)
	link := 0
	for tr := range n.Triples() {
		if r, ok := tr.Object.(marc.Ref); ok && r.Target().String() == "RFC 8446" {
			link++
		}
	}

	it.Then(t).Should(
		it.String(defaultTexts(n)[0]).Contain("see [[RFC 8446]] for details"),
		it.Equal(link, 0),
	)
}

func TestProseParagraphs(t *testing.T) {
	// Scenario 4.2 from specs/001-markdown-triples-codec/spec.md (BUG-001:
	// each paragraph is its own Text value, in document order)
	t.Parallel()

	src := fixture(t, "testdata/prose-untyped-links.md")
	n := proseNode(t)
	paragraphs := defaultTexts(n)

	b, err := marc.Marshal(n)

	it.Then(t).Must(it.Nil(err), it.Equal(len(paragraphs), 2)).Should(
		it.String(paragraphs[0]).HavePrefix("The first paragraph"),
		it.String(paragraphs[0]).HaveSuffix("It continues on a second line."),
		it.Equal(paragraphs[1], "The second paragraph closes the introduction."),
		it.Equal(string(b), string(src)),
	)
}

func TestProseLinks(t *testing.T) {
	// Scenario 4.4 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n := proseNode(t)
	var got []struct {
		Predicate marc.Predicate
		Target    string
	}
	for p, r := range n.Links() {
		got = append(got, struct {
			Predicate marc.Predicate
			Target    string
		}{p, r.Target().String()})
	}

	it.Then(t).Should(it.Equal(len(got), 1), it.Equal(got[0].Predicate.String(), "text"), it.Equal(got[0].Target, "RFC 8446"))
}

func TestProseColonLead(t *testing.T) {
	// Scenario 4.5 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n, err := marc.Unmarshal([]byte(
		"---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\n**Note**: This is a note.\n"))
	it.Then(t).Must(it.Nil(err))

	b, err := marc.Marshal(n)
	it.Then(t).Must(it.Nil(err)).Should(
		it.Seq(slices.Collect(n.Values(marc.MustPredicate("note")))).Equal(marc.Text("This is a note.")),
		it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\n**Note**\nThis is a note.\n"),
	)
}

func TestProseStatements(t *testing.T) {
	// Scenario 4.3 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n := proseNode(t)

	it.Then(t).Should(
		it.Seq(slices.Collect(n.Values(marc.MustPredicate("claims")))).Equal(
			marc.Text("Handshake latency drops to one round trip."),
			marc.Text("Static RSA key exchange is removed."),
		),
	)
}
