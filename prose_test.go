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
	"strings"
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

func defaultText(n marc.Node) string {
	for o := range n.Values(marc.MustPredicate("text")) {
		return o.String()
	}
	return ""
}

func TestProseUntypedLink(t *testing.T) {
	// Scenario 4.1 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n := proseNode(t)
	link := 0
	for tr := range n.Triples() {
		if r, ok := tr.Object.(marc.Ref); ok && r.Target().String() == "RFC 8446" {
			link++
		}
	}

	it.Then(t).Should(
		it.String(defaultText(n)).Contain("see [[RFC 8446]] for details"),
		it.Equal(link, 0),
	)
}

func TestProseParagraphs(t *testing.T) {
	// Scenario 4.2 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	src := fixture(t, "testdata/prose-untyped-links.md")
	n := proseNode(t)
	paragraphs := strings.Split(defaultText(n), "\n\n")

	b, err := marc.Marshal(n)

	it.Then(t).Must(it.Nil(err), it.Equal(len(paragraphs), 2)).Should(
		it.String(paragraphs[0]).HavePrefix("The first paragraph"),
		it.String(paragraphs[0]).HaveSuffix("It continues on a second line."),
		it.Equal(paragraphs[1], "The second paragraph closes the introduction."),
		it.Equal(string(b), string(src)),
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
