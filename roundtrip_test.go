//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/net-of-arc/marc"
)

func TestRoundTripCanonical(t *testing.T) {
	// Scenario 3.1 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	files, err := filepath.Glob("testdata/canonical/*.md")
	it.Then(t).Must(it.Nil(err), it.True(len(files) >= 5))
	files = append(files,
		"testdata/core-3-example.md",
		"testdata/core-8-2-conflicts.md",
		"testdata/core-8-3-provenance.md",
		"testdata/prose-untyped-links.md",
	)

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			src := fixture(t, file)
			n, err := marc.Unmarshal(src)
			it.Then(t).Must(it.Nil(err))

			b, err := marc.Marshal(n)
			it.Then(t).Must(it.Nil(err)).Should(
				it.Equal(string(b), string(src)),
			)
		})
	}
}

func TestRoundTripBag(t *testing.T) {
	// Scenario 3.2 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n, err := marc.NewNode(marc.MustID(tls), marc.MustClass("Entity"), coreTriples())
	it.Then(t).Must(it.Nil(err))

	b, err := codec(t, coreSchema()).Marshal(n)
	it.Then(t).Must(it.Nil(err))

	m, err := marc.Unmarshal(b)
	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(m.ID(), n.ID()),
		it.Equal(m.Class(), n.Class()),
		it.Seq(slices.Collect(m.Triples())).Equal(slices.Collect(n.Triples())...),
	)
}

func TestRoundTripUnknownPredicate(t *testing.T) {
	// Scenario 3.3 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	head := "---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\n"
	for name, body := range map[string]string{
		"labelled prose": "**Foo Bar**\nSome text about foo.\n",
		"edge block":     "- fooBar:: [[Y]]\n",
		"link block":     "## Foo Bar\n- fooBar:: [[Y]]\n",
		"front-matter":   "",
	} {
		t.Run(name, func(t *testing.T) {
			src := head + body
			if body == "" {
				src = "---\n\"@id\": X\n\"@type\": Entity\nfooBar: value\n---\n# X\n"
			}
			n, err := marc.Unmarshal([]byte(src))
			it.Then(t).Must(it.Nil(err))

			b, err := marc.Marshal(n)
			it.Then(t).Must(it.Nil(err)).Should(
				it.Equal(string(b), src),
			)
		})
	}
}

func TestRoundTripEdgeForms(t *testing.T) {
	// Scenario 3.4 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	for file, predicate := range map[string]string{
		"testdata/forms/body-form.md":   "conformsTo",
		"testdata/forms/inline-form.md": "supports",
	} {
		t.Run(predicate, func(t *testing.T) {
			n, err := marc.Unmarshal(fixture(t, file))
			it.Then(t).Must(it.Nil(err)).Should(
				it.Seq(slices.Collect(n.Values(marc.MustPredicate(predicate)))).Equal(ref("RFC 8446")),
			)

			b, err := marc.Marshal(n)
			it.Then(t).Must(it.Nil(err)).Should(
				it.String(string(b)).Contain("\n- " + predicate + ":: [[RFC 8446]]\n"),
			)
		})
	}
}
