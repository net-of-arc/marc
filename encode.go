//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"bytes"
	"strings"
)

type layout struct {
	meta     []entry
	prose    []section
	labelled []section
	edges    []section
	blocks   []section
}

type section struct {
	label     string
	predicate Predicate
	values    []Object
}

func (c *Codec) encode(buf *bytes.Buffer, n Node) error {
	if n.id.IsZero() || n.class.IsZero() {
		return ErrMalformed.With(detail("zero node"))
	}
	var l layout
	for _, set := range n.facts {
		prop, err := c.property(n, set)
		if err != nil {
			return err
		}
		if err := l.place(set, prop); err != nil {
			return err
		}
	}
	buf.WriteString(l.render(n))
	return nil
}

func (c *Codec) property(n Node, set valueSet) (Property, error) {
	if c.schema != nil {
		if prop, ok := c.schema.Property(n.class, set.predicate); ok {
			return prop, validProperty(set.predicate, prop)
		}
	}
	if prop, ok := n.observed[set.predicate]; ok {
		return prop, nil
	}
	if refs, _ := partition(set.values, isRef); len(refs) > 0 {
		return Property{Role: RoleEdge}, nil
	}
	return Property{Role: RoleMeta}, nil
}

func (l *layout) place(set valueSet, prop Property) error {
	rest, err := l.placeBody(set, prop)
	if err != nil || len(rest) == 0 {
		return err
	}
	many := prop.Role == RoleMeta && prop.Many
	l.meta = append(l.meta, entry{predicate: set.predicate, property: Property{Many: many}, values: rest})
	return nil
}

func (l *layout) placeBody(set valueSet, prop Property) ([]Object, error) {
	label := labelOf(set.predicate, prop)
	switch prop.Role {
	case RoleEdge:
		refs, rest := partition(set.values, isRef)
		l.edges = appendSection(l.edges, section{predicate: set.predicate, values: refs})
		return rest, nil
	case RoleLink:
		refs, rest := partition(set.values, isRef)
		l.blocks = appendSection(l.blocks, section{label: label, predicate: set.predicate, values: refs})
		return rest, nil
	case RoleText:
		return l.placeText(set, prop, label)
	}
	return set.values, nil
}

// placeText places the Text values of a text-role predicate. A Many
// predicate (an explicit list of statements) renders as a "## Label" block
// of bullets; otherwise every renderable value is one paragraph, in node
// order, under the default prose or a single "**Label**" (BUG-001: a
// predicate is no longer forced into a bullet block merely because it has
// more than one value).
func (l *layout) placeText(set valueSet, prop Property, label string) ([]Object, error) {
	texts, rest := partition(set.values, isText)
	if prop.Many {
		bullets, others := partition(texts, bulletable)
		l.blocks = appendSection(l.blocks, section{label: label, predicate: set.predicate, values: bullets})
		return append(rest, others...), nil
	}
	paragraphs, others := partition(texts, func(o Object) bool { return proseRenderable(o.String()) })
	for _, p := range paragraphs {
		if proseStructure(p.String()) {
			return nil, ErrProseStructure.With(detail(set.predicate.value))
		}
	}
	if len(paragraphs) == 0 {
		return append(rest, others...), nil
	}
	prose := section{label: label, predicate: set.predicate, values: paragraphs}
	if prop.Default && len(l.prose) == 0 {
		l.prose = append(l.prose, prose)
	} else {
		l.labelled = append(l.labelled, prose)
	}
	return append(rest, others...), nil
}

func appendSection(sections []section, s section) []section {
	if len(s.values) == 0 {
		return sections
	}
	return append(sections, s)
}

func partition(values []Object, keep func(Object) bool) ([]Object, []Object) {
	var yes, no []Object
	for _, o := range values {
		if keep(o) {
			yes = append(yes, o)
		} else {
			no = append(no, o)
		}
	}
	return yes, no
}

func isRef(o Object) bool {
	_, ok := o.(Ref)
	return ok
}

func isText(o Object) bool {
	_, ok := o.(Text)
	return ok
}

func bulletable(o Object) bool {
	s := o.String()
	return s != "" && strings.TrimSpace(s) == s && !strings.ContainsAny(s, "\n\r") && !refLexical(s)
}

// proseRenderable reports whether s can be written as one paragraph: no
// blank line outside fenced code (BUG-001: a value with an internal blank
// line would be re-read as more than one paragraph), no "\r", not empty.
func proseRenderable(s string) bool {
	if s == "" || strings.Contains(s, "\r") {
		return false
	}
	fence := ""
	for _, line := range strings.Split(s, "\n") {
		switch {
		case fence != "":
			if closesFence(line, fence) {
				fence = ""
			}
		case fenceMarker(line) != "":
			fence = fenceMarker(line)
		case strings.TrimSpace(line) == "":
			return false
		}
	}
	return fence == ""
}

func proseStructure(text string) bool {
	fence := ""
	for _, line := range strings.Split(text, "\n") {
		switch {
		case fence != "":
			if closesFence(line, fence) {
				fence = ""
			}
		case fenceMarker(line) != "":
			fence = fenceMarker(line)
		case structuralLine(line):
			return true
		}
	}
	return fence != ""
}

func structuralLine(line string) bool {
	_, _, isBullet := bullet(line)
	_, _, isEdge := bodyEdge(line)
	_, isLabel := labelText(line)
	_, _, isLead := boldLeadText(line)
	return atxHeading(line) || isBullet || isEdge || isLabel || isLead
}

func (l layout) render(n Node) string {
	var sb strings.Builder
	renderFrontMatter(&sb, n, l.meta)
	sb.WriteString("# " + n.id.value + "\n")
	for _, s := range l.prose {
		writeParagraphs(&sb, "\n", s.values)
	}
	for _, s := range l.labelled {
		writeParagraphs(&sb, "\n**"+s.label+"**\n", s.values)
	}
	if len(l.edges) > 0 {
		sb.WriteString("\n")
		renderBullets(&sb, l.edges)
	}
	for _, s := range l.blocks {
		sb.WriteString("\n## " + s.label + "\n")
		renderBullets(&sb, []section{s})
	}
	return sb.String()
}

// writeParagraphs writes one or more paragraph values with lead preceding
// the first and a blank line between each pair (BUG-001).
func writeParagraphs(sb *strings.Builder, lead string, values []Object) {
	sb.WriteString(lead + values[0].String() + "\n")
	for _, v := range values[1:] {
		sb.WriteString("\n" + v.String() + "\n")
	}
}

func renderBullets(sb *strings.Builder, sections []section) {
	for _, s := range sections {
		for _, o := range s.values {
			sb.WriteString("- " + s.predicate.value + ":: " + o.String() + "\n")
		}
	}
}
