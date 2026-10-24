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

type conflict struct {
	Predicate marc.Predicate
	Object    marc.Object
}

func TestMergeStateConflicts(t *testing.T) {
	// Scenario 6.1 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n, err := marc.Unmarshal(fixture(t, "testdata/core-8-2-conflicts.md"))
	it.Then(t).Must(it.Nil(err))

	var conflicts []conflict
	for p, o := range n.Conflicts() {
		conflicts = append(conflicts, conflict{p, o})
	}

	it.Then(t).Should(
		it.Seq(conflicts).Equal(
			conflict{marc.MustPredicate("category"), marc.Text("object")},
			conflict{marc.MustPredicate("category"), marc.Text("process")},
			conflict{marc.MustPredicate("description"), marc.Text("A protocol suite for securing transport connections, standardized by the IETF.\n")},
			conflict{marc.MustPredicate("replaces"), ref("SSL Version 3")},
		),
		it.Seq(slices.Collect(n.Triples())).Equal(
			triple(tls, "category", marc.Text("script")),
		),
	)
}

func TestMergeStateProvenance(t *testing.T) {
	// Scenario 6.2 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n, err := marc.Unmarshal(fixture(t, "testdata/core-8-3-provenance.md"))
	it.Then(t).Must(it.Nil(err))

	attribution := func(doc, p string, o marc.Object) marc.Attribution {
		return marc.Attribution{Document: marc.MustID(doc), Predicate: marc.MustPredicate(p), Object: o}
	}

	it.Then(t).Should(
		it.Seq(slices.Collect(n.Provenance())).Equal(
			attribution("dierks-2008-tls12", "category", marc.Text("script")),
			attribution("dierks-2008-tls12", "text", marc.Text("sha256:91ab03ce55f2a7d0")),
			attribution("rescorla-2018-tls13", "category", marc.Text("script")),
			attribution("rescorla-2018-tls13", "replaces", ref("SSL Protocol")),
			attribution("rescorla-2018-tls13", "text", marc.Text("sha256:4f1c2a9b0e7d3c18")),
		),
		it.Seq(slices.Collect(n.Triples())).Equal(
			triple(tls, "category", marc.Text("script")),
		),
	)
}

func TestMergeStateRoundTrip(t *testing.T) {
	// Scenario 6.3 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	for _, file := range []string{
		"testdata/core-8-2-conflicts.md",
		"testdata/core-8-3-provenance.md",
	} {
		src := fixture(t, file)
		n, err := marc.Unmarshal(src)
		it.Then(t).Must(it.Nil(err))

		b, err := marc.Marshal(n)
		it.Then(t).Must(it.Nil(err)).Should(
			it.Equal(string(b), string(src)),
		)
	}
}
