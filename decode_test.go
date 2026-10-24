//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"os"
	"slices"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/net-of-arc/marc"
)

func fixture(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func triple(id, p string, o marc.Object) marc.Triple {
	return marc.Triple{Subject: marc.MustID(id), Predicate: marc.MustPredicate(p), Object: o}
}

func ref(id string) marc.Ref { return marc.RefTo(marc.MustID(id)) }

const tls = "Transport Layer Security"

func TestDecodeCoreExample(t *testing.T) {
	// Scenario 1.1 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n, err := marc.Unmarshal(fixture(t, "testdata/core-3-example.md"))

	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(n.ID().String(), tls),
		it.Equal(n.Class().String(), "Entity"),
		it.Seq(slices.Collect(n.Triples())).Equal(
			triple(tls, "category", marc.Text("script")),
			triple(tls, "tags", marc.Text("cryptography")),
			triple(tls, "text", marc.Text("A cryptographic protocol that establishes an authenticated, confidential channel over an untrusted network.")),
			triple(tls, "conformsTo", ref("RFC 8446")),
			triple(tls, "replaces", ref("SSL Protocol")),
			triple(tls, "mentionedIn", ref("rescorla-2026-tls13")),
		),
	)
}

func TestDecodeNaturalOrder(t *testing.T) {
	// Scenario 1.6 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	src := "---\n\"@id\": X\n\"@type\": Entity\ntags: [zeta, alpha]\n---\n# X\n\n- replaces:: [[B]]\n- conformsTo:: [[A]]\n"
	n, err := marc.Unmarshal([]byte(src))

	it.Then(t).Must(it.Nil(err)).Should(
		it.Seq(slices.Collect(n.Triples())).Equal(
			triple("X", "tags", marc.Text("zeta")),
			triple("X", "tags", marc.Text("alpha")),
			triple("X", "replaces", ref("B")),
			triple("X", "conformsTo", ref("A")),
		),
		it.Seq(slices.Collect(n.Predicates())).Equal(
			marc.MustPredicate("tags"), marc.MustPredicate("replaces"), marc.MustPredicate("conformsTo"),
		),
	)
}

func TestDecodeMultiValued(t *testing.T) {
	// Scenario 1.2 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	src := "---\n\"@id\": X\n\"@type\": Entity\ntags: [cryptography, networking]\n---\n# X\n"
	n, err := marc.Unmarshal([]byte(src))

	it.Then(t).Must(it.Nil(err)).Should(
		it.Seq(slices.Collect(n.Triples())).Equal(
			triple("X", "tags", marc.Text("cryptography")),
			triple("X", "tags", marc.Text("networking")),
		),
	)
}

func TestDecodeReference(t *testing.T) {
	// Scenario 1.3 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n, err := marc.Unmarshal(fixture(t, "testdata/core-3-example.md"))
	it.Then(t).Must(it.Nil(err))

	refs := slices.Collect(marc.Select[marc.Ref](n.Values(marc.MustPredicate("replaces"))))

	it.Then(t).Must(it.Equal(len(refs), 1)).Should(
		it.Equal(refs[0].Target().String(), "SSL Protocol"),
	)
}

func TestDecodeAbsentPredicate(t *testing.T) {
	// Scenario 1.4 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	n, err := marc.Unmarshal(fixture(t, "testdata/core-3-example.md"))

	it.Then(t).Must(it.Nil(err)).Should(
		it.Seq(slices.Collect(n.Values(marc.MustPredicate("absent")))).BeEmpty(),
	)
}

func TestDecodeLiteralKinds(t *testing.T) {
	// Scenario 1.5 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	src := "---\n\"@id\": X\n\"@type\": Entity\nn: 42\nf: 1.5\nb: true\nd: 2026-10-23\ns: 2026-10-23T10:00:00Z\n---\n# X\n"
	n, err := marc.Unmarshal([]byte(src))
	it.Then(t).Must(it.Nil(err))

	value := func(p string) marc.Object {
		for o := range n.Values(marc.MustPredicate(p)) {
			return o
		}
		return marc.Text("")
	}

	it.Then(t).Should(
		it.TypeOf[marc.Number](value("n")),
		it.Equal(value("n").String(), "42"),
		it.TypeOf[marc.Number](value("f")),
		it.Equal(value("f").String(), "1.5"),
		it.Equiv(value("b"), marc.Object(marc.Bool(true))),
		it.TypeOf[marc.Date](value("d")),
		it.Equal(value("d").String(), "2026-10-23"),
		it.TypeOf[marc.DateTime](value("s")),
		it.Equal(value("s").String(), "2026-10-23T10:00:00Z"),
	)
}
