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
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"go.yaml.in/yaml/v3"
)

func TestSplitFrontMatter(t *testing.T) {
	for src, expected := range map[string][2]string{
		"---\na: 1\n---\n# X\n":         {"a: 1\n", "# X\n"},
		"---\r\na: 1\r\n---\r\n# X\r\n": {"a: 1\n", "# X\n"},
		"---\n---\n# X\n":               {"", "# X\n"},
		"---\na: 1\n---":                {"a: 1\n", ""},
		"\ufeff---\na: 1\n---\nbody\n":  {"a: 1\n", "body\n"},
	} {
		front, body, err := splitFrontMatter([]byte(src))
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(string(front), expected[0]),
			it.Equal(string(body), expected[1]),
		)
	}
	for _, src := range []string{"", "# X\n", "---\na: 1\n", "--- \na: 1\n---\n"} {
		_, _, err := splitFrontMatter([]byte(src))
		it.Then(t).Should(it.True(errors.Is(err, ErrMalformed)))
	}
}

func parseHeader(t *testing.T, yaml string) header {
	t.Helper()
	h, err := parseFrontMatter([]byte("\"@id\": X\n\"@type\": Entity\n" + yaml))
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestParseFrontMatterKeywords(t *testing.T) {
	h := parseHeader(t, "")
	it.Then(t).Should(
		it.Equal(h.id, MustID("X")),
		it.Equal(h.class, MustClass("Entity")),
		it.Seq(h.meta).BeEmpty(),
	)
}

func TestParseFrontMatterScalars(t *testing.T) {
	h := parseHeader(t, "n: 42\nf: 1.5\nb: true\nd: 2026-10-23\ns: 2026-10-23T12:00:00+02:00\nt: script\nr: \"[[Target|shown]]\"\nq: '42'\nz: ~\ne:\n")

	it.Then(t).Should(it.Seq(h.meta).Equal(
		entry{MustPredicate("n"), Property{Role: RoleMeta}, []Object{mustNumber("42")}},
		entry{MustPredicate("f"), Property{Role: RoleMeta}, []Object{mustNumber("1.5")}},
		entry{MustPredicate("b"), Property{Role: RoleMeta}, []Object{Bool(true)}},
		entry{MustPredicate("d"), Property{Role: RoleMeta}, []Object{DateOf(time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC))}},
		entry{MustPredicate("s"), Property{Role: RoleMeta}, []Object{DateTimeOf(time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC))}},
		entry{MustPredicate("t"), Property{Role: RoleMeta}, []Object{Text("script")}},
		entry{MustPredicate("r"), Property{Role: RoleMeta}, []Object{RefTo(MustID("Target")).WithDisplay("shown")}},
		entry{MustPredicate("q"), Property{Role: RoleMeta}, []Object{Text("42")}},
	))
}

func TestParseFrontMatterSequences(t *testing.T) {
	h := parseHeader(t, "flow: [a, b]\nblock:\n  - x\n  - 1\none: [only]\nempty: []\n")

	it.Then(t).Should(it.Seq(h.meta).Equal(
		entry{MustPredicate("flow"), Property{Role: RoleMeta, Many: true}, []Object{Text("a"), Text("b")}},
		entry{MustPredicate("block"), Property{Role: RoleMeta, Many: true}, []Object{Text("x"), mustNumber("1")}},
		entry{MustPredicate("one"), Property{Role: RoleMeta, Many: true}, []Object{Text("only")}},
	))
}

func TestParseFrontMatterBareKeyword(t *testing.T) {
	h, err := parseFrontMatter([]byte("'@id': X\n'@type': Entity\n"))
	it.Then(t).Should(it.Nil(err), it.Equal(h.id, MustID("X")))
}

func TestParseFrontMatterRejects(t *testing.T) {
	for src, expected := range map[string]error{
		"\"@type\": Entity\n":                             ErrMalformed,
		"\"@id\": X\n":                                    ErrMalformed,
		"\"@id\": X\n\"@type\": Entity\nk: [a\n":          ErrMalformed,
		"- a\n- b\n":                                      ErrMalformed,
		"\"@id\": X\n\"@type\": Entity\nk: &a v\n":        ErrMalformed,
		"\"@id\": X\n\"@type\": Entity\nk: [[a]]\n":       ErrMalformed,
		"\"@id\": X\n\"@type\": Entity\nk: !foo v\n":      ErrMalformed,
		"\"@id\": X\n\"@type\": Entity\nk: !!int v\n":     ErrInvalidObject,
		"\"@id\": X\n\"@type\": Entity\nBad: v\n":         ErrInvalidName,
		"\"@id\": X\n\"@type\": Entity\n\"@x\": v\n":      ErrInvalidName,
		"\"@id\": a/b\n\"@type\": Entity\n":               ErrInvalidID,
		"\"@id\": X\n\"@type\": entity\n":                 ErrInvalidName,
		"\"@id\": X\n\"@type\": Entity\nr: \"[[a/b]]\"\n": ErrInvalidID,
	} {
		_, err := parseFrontMatter([]byte(src))
		it.Then(t).Should(it.True(errors.Is(err, expected)))
	}
}

func TestParseFrontMatterMergeState(t *testing.T) {
	h := parseHeader(t, "\"@conflicts\":\n  category: [object, process]\n  description:\n    - |\n      A text.\n\"@provenance\":\n  doc-1:\n    replaces: [\"[[Y]]\"]\n    text: [\"sha256:91ab\"]\n")

	it.Then(t).Should(
		it.Seq(h.meta).BeEmpty(),
		it.Seq(h.conflicts).Equal(
			entry{MustPredicate("category"), Property{}, []Object{Text("object"), Text("process")}},
			entry{MustPredicate("description"), Property{}, []Object{Text("A text.\n")}},
		),
		it.Seq(h.provenance).Equal(
			Attribution{MustID("doc-1"), MustPredicate("replaces"), RefTo(MustID("Y"))},
			Attribution{MustID("doc-1"), MustPredicate("text"), Text("sha256:91ab")},
		),
	)
}

func TestParseFrontMatterMergeStateRejects(t *testing.T) {
	head := "\"@id\": X\n\"@type\": Entity\n"
	for src, expected := range map[string]error{
		"\"@conflicts\": [a]\n":                          ErrMalformed,
		"\"@conflicts\":\n  p: [[a]]\n":                  ErrMalformed,
		"\"@conflicts\":\n  p:\n    k: v\n":              ErrMalformed,
		"\"@conflicts\":\n  Bad: [a]\n":                  ErrInvalidName,
		"\"@provenance\":\n  doc: [a]\n":                 ErrMalformed,
		"\"@provenance\":\n  a/b:\n    p: [a]\n":         ErrInvalidID,
		"\"@provenance\":\n  doc:\n    p:\n      k: v\n": ErrMalformed,
		"\"@provenance\":\n  doc:\n    Bad: [a]\n":       ErrInvalidName,
	} {
		_, err := parseFrontMatter([]byte(head + src))
		it.Then(t).Should(it.True(errors.Is(err, expected)))
	}
}

func TestYAMLStringQuoting(t *testing.T) {
	for _, s := range []string{"script", "RFC 8446", "cryptography", "Café au lait", "a-b", "x.y", "dierks-2008-tls12", "2026-review"} {
		it.Then(t).Should(it.Equal(yamlString(s), s))
	}
	for s, rendered := range map[string]string{
		"yes":         `"yes"`,
		"true":        `"true"`,
		"1":           `"1"`,
		"~":           `"~"`,
		"null":        `"null"`,
		"2026-10-23":  `"2026-10-23"`,
		"a: b":        `"a: b"`,
		"a #b":        `"a #b"`,
		"- x":         `"- x"`,
		"[x]":         `"[x]"`,
		" lead":       `" lead"`,
		"":            `""`,
		`"q"`:         `"\"q\""`,
		"sha256:91ab": `"sha256:91ab"`,
		"tab\there":   `"tab\there"`,
		"bell\a":      `"bell\u0007"`,
		"<<":          `"<<"`,
		".inf":        `".inf"`,
	} {
		it.Then(t).Should(it.Equal(yamlString(s), rendered))
	}
}

func TestYAMLScalarVariants(t *testing.T) {
	it.Then(t).Should(
		it.Equal(yamlScalar(mustNumber("42")), "42"),
		it.Equal(yamlScalar(Bool(false)), "false"),
		it.Equal(yamlScalar(DateOf(time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC))), "2026-10-23"),
		it.Equal(yamlScalar(DateTimeOf(time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC))), "2026-10-23T10:00:00Z"),
		it.Equal(yamlScalar(RefTo(MustID("T"))), `"[[T]]"`),
		it.Equal(yamlScalar(RefTo(MustID("T")).WithDisplay(`say "hi"`)), `"[[T|say \"hi\"]]"`),
	)
}

func TestYAMLStringReadsBack(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2))
	alphabet := []rune("ab Z09:#,-[]{}\"'\\|>!&*%@`~?.+<=\t\n\r\x00\u00a0\u2028é🙂")
	for range 5000 {
		runes := make([]rune, rnd.IntN(12))
		for i := range runes {
			runes[i] = alphabet[rnd.IntN(len(alphabet))]
		}
		s := string(runes)
		var back struct {
			V string
			F []string
		}
		err := yaml.Unmarshal([]byte("v: "+yamlString(s)+"\nf: [x, "+yamlString(s)+"]\n"), &back)
		it.Then(t).Should(it.Nil(err), it.Equal(back.V, s), it.Seq(back.F).Equal("x", s))
	}
}

func TestRenderFrontMatter(t *testing.T) {
	n, err := NewNode(MustID("X"), MustClass("Entity"), slices.Values([]Triple{
		tripleOf("X", "tags", Text("b")),
		tripleOf("X", "tags", Text("a")),
		tripleOf("X", "one", Text("v")),
		tripleOf("X", "many", Text("v")),
	}))
	it.Then(t).Must(it.Nil(err))

	var sb strings.Builder
	renderFrontMatter(&sb, n, []entry{
		{MustPredicate("many"), Property{Many: true}, []Object{Text("v")}},
		{MustPredicate("one"), Property{}, []Object{Text("v")}},
		{MustPredicate("tags"), Property{}, []Object{Text("a"), Text("b")}},
	})

	it.Then(t).Should(it.Equal(sb.String(), "---\n\"@id\": X\n\"@type\": Entity\nmany: [v]\none: v\ntags: [a, b]\n---\n"))
}

func TestRenderMergeState(t *testing.T) {
	n, err := Unmarshal([]byte("---\n\"@id\": X\n\"@type\": Entity\n" +
		"\"@provenance\":\n  zeta:\n    p: [b, a]\n  alpha:\n    q: [x]\n    p: [\"[[Y]]\"]\n" +
		"\"@conflicts\":\n  z: [b, a]\n  a:\n    - plain\n    - |-\n      two\n      lines\n    - |+\n      kept\n\n---\n# X\n"))
	it.Then(t).Must(it.Nil(err))

	b, err := Marshal(n)
	it.Then(t).Must(it.Nil(err)).Should(it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\n"+
		"\"@conflicts\":\n  a:\n    - |+\n      kept\n\n    - plain\n    - |-\n      two\n      lines\n  z: [a, b]\n"+
		"\"@provenance\":\n  alpha:\n    p: [\"[[Y]]\"]\n    q: [x]\n  zeta:\n    p: [a, b]\n---\n# X\n"))

	back, err := Unmarshal(b)
	it.Then(t).Must(it.Nil(err)).Should(
		it.Equiv(back.conflicts, n.conflicts),
		it.Equiv(back.provenance, n.provenance),
		it.Equiv(n.conflicts[0].values, []Object{Text("kept\n\n"), Text("plain"), Text("two\nlines")}),
	)
}

func TestBlockItemFallsBackToQuoted(t *testing.T) {
	it.Then(t).Should(
		it.Equal(blockItem(Text(" indented\nline"), "    "), `" indented\nline"`),
		it.Equal(blockItem(Text("cr\r\nline"), "    "), `"cr\r\nline"`),
		it.Equal(blockItem(Text("a\nb\n"), "  "), "|\n  a\n  b"),
	)
}
