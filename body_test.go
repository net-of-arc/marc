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
	"testing"

	"github.com/fogfish/it/v2"
)

func scan(t *testing.T, body string, schema Schema) []entry {
	t.Helper()
	entries, err := scanBody(body, scope{id: MustID("X"), class: MustClass("Entity"), schema: schema, line: 5})
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func textEntry(p string, prop Property, s string) entry {
	return entry{MustPredicate(p), prop, []Object{Text(s)}}
}

func refEntry(p string, prop Property, target string) entry {
	return entry{MustPredicate(p), prop, []Object{RefTo(MustID(target))}}
}

var (
	defaultProse = Property{Role: RoleText, Default: true}
	edgeProperty = Property{Role: RoleEdge}
)

func TestScanBodyTitle(t *testing.T) {
	for body, expected := range map[string]error{
		"# X\n":        nil,
		"\n\n# X\n":    nil,
		"":             ErrMalformed,
		"prose\n# X\n": ErrMalformed,
		"# Y\n":        ErrIdentityMismatch,
		"# X \n":       nil,
		"## X\n":       ErrMalformed,
		"#X\n":         ErrMalformed,
		"# Café\n":    ErrIdentityMismatch,
	} {
		_, err := scanBody(body, scope{id: MustID("X"), class: MustClass("Entity"), line: 5})
		if expected == nil {
			it.Then(t).Should(it.Nil(err))
		} else {
			it.Then(t).Should(it.True(errors.Is(err, expected)))
		}
	}
}

func TestScanBodyDefaultProse(t *testing.T) {
	entries := scan(t, "# X\n\nFirst line\nsecond line.\n\nSecond paragraph.\n", nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("text", defaultProse, "First line\nsecond line.\n\nSecond paragraph."),
	))
}

type summarySchema struct{}

func (summarySchema) Property(Class, Predicate) (Property, bool) { return Property{}, false }
func (summarySchema) Predicate(c Class, label string) (Predicate, bool) {
	if label == "" {
		return MustPredicate("summary"), true
	}
	return Predicate{}, false
}

func TestScanBodyDefaultProseFromSchema(t *testing.T) {
	entries := scan(t, "# X\n\nSome prose.\n", summarySchema{})
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("summary", defaultProse, "Some prose."),
	))
}

func TestScanBodyEdges(t *testing.T) {
	entries := scan(t, "# X\n\n- conformsTo:: [[RFC 8446]]\n- replaces:: [[SSL Protocol|SSL]]\n", nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		refEntry("conformsTo", edgeProperty, "RFC 8446"),
		entry{MustPredicate("replaces"), edgeProperty, []Object{RefTo(MustID("SSL Protocol")).WithDisplay("SSL")}},
	))
}

func TestScanBodyLinkBlocks(t *testing.T) {
	entries := scan(t, "# X\n\n## Mentioned In\n- mentionedIn:: [[A]]\n- mentionedIn:: [[B]]\n\n## Sources\n- citedBy:: [[C]]\n", nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		refEntry("mentionedIn", Property{Role: RoleLink}, "A"),
		refEntry("mentionedIn", Property{Role: RoleLink}, "B"),
		refEntry("citedBy", Property{Role: RoleLink, Label: "Sources"}, "C"),
	))
}

func TestScanBodyTolerantForms(t *testing.T) {
	entries := scan(t, "# X\n\nIt was [supports:: [[RFC 8446]]] adopted.\n\nconformsTo:: [[RFC 5246]]\n\n**Standards**\n- obsoletes:: [[RFC 2246]]\n", nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("text", defaultProse, "It was [supports:: [[RFC 8446]]] adopted."),
		refEntry("supports", edgeProperty, "RFC 8446"),
		refEntry("conformsTo", edgeProperty, "RFC 5246"),
		refEntry("obsoletes", edgeProperty, "RFC 2246"),
	))
}

func TestScanBodyFirstPositionWins(t *testing.T) {
	n, err := Unmarshal([]byte("---\n\"@id\": X\n\"@type\": Entity\nreplaces: \"[[A]]\"\n---\n# X\n\n- replaces:: [[B]]\n"))
	prop, _ := n.Property(MustPredicate("replaces"))
	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(prop, Property{Role: RoleMeta}),
		it.Equal(len(lookup(n.facts, MustPredicate("replaces"))), 2),
	)
}

func TestScanBodyForeignBulletInBlock(t *testing.T) {
	entries := scan(t, "# X\n\n## Mentioned In\n- mentionedIn:: [[A]]\n- citedBy:: [[B]]\n", nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		refEntry("mentionedIn", Property{Role: RoleLink}, "A"),
		refEntry("citedBy", Property{Role: RoleLink, Label: "Mentioned In"}, "B"),
	))
}

func TestScanBodyUntypedLinks(t *testing.T) {
	entries := scan(t, "# X\n\nSee [[RFC 8446]] and [[Other|shown]].\n", nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("text", defaultProse, "See [[RFC 8446]] and [[Other|shown]]."),
	))
}

func TestScanBodyFencedProse(t *testing.T) {
	body := "# X\n\nIntro.\n\n```markdown\n# Not a title\n## Not a block\n- p:: [[Not]]\n**Not**\n```\n\n~~~\n- q:: x\n~~~\n"
	entries := scan(t, body, nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("text", defaultProse, "Intro.\n\n```markdown\n# Not a title\n## Not a block\n- p:: [[Not]]\n**Not**\n```\n\n~~~\n- q:: x\n~~~"),
	))
}

func TestScanBodyStatements(t *testing.T) {
	entries := scan(t, "# X\n\n## Claims\n- claims:: One.\n- claims:: Two.\n\n## Notable Claims\n- claims:: Three.\n", nil)
	many := Property{Role: RoleText, Many: true}
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("claims", many, "One."),
		textEntry("claims", many, "Two."),
		textEntry("claims", Property{Role: RoleText, Many: true, Label: "Notable Claims"}, "Three."),
	))
}

func TestScanBodyLabelledProse(t *testing.T) {
	entries := scan(t, "# X\n\nDefault.\n\n**Key Findings**\nFirst line\nsecond line.\n\nMore.\n\n**Key findings**\nAgain.\n", nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("text", defaultProse, "Default."),
		textEntry("keyFindings", Property{Role: RoleText}, "First line\nsecond line.\n\nMore."),
		textEntry("keyFindings", Property{Role: RoleText, Label: "Key findings"}, "Again."),
	))
}

type labelSchema struct{}

func (labelSchema) Property(Class, Predicate) (Property, bool) { return Property{}, false }
func (labelSchema) Predicate(c Class, label string) (Predicate, bool) {
	if label == "Summary Notes" {
		return MustPredicate("notes"), true
	}
	return Predicate{}, false
}

func TestScanBodyLabelFromSchema(t *testing.T) {
	entries := scan(t, "# X\n\n**Summary Notes**\nText.\n", labelSchema{})
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("notes", Property{Role: RoleText, Label: "Summary Notes"}, "Text."),
	))
}

func TestScanBodyUnresolvableLabelIsProse(t *testing.T) {
	entries := scan(t, "# X\n\n**123 Go!**\nText.\n", nil)
	it.Then(t).Should(it.Seq(entries).Equal(
		textEntry("text", defaultProse, "**123 Go!**\nText."),
	))
}
