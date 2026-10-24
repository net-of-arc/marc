//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type scope struct {
	id     ID
	class  Class
	schema Schema
	line   int
}

type scanner struct {
	scope
	entries   []entry
	prose     map[int][]string
	target    int
	predicate Predicate
	textProp  Property
	active    bool
	fence     string
	inBlock   bool
	heading   string
}

var (
	inlineEdge    = regexp.MustCompile(`\[([a-z][A-Za-z0-9]*)::\s*(\[\[[^\[\]]*\]\])\]`)
	untypedLinkRe = regexp.MustCompile(`\[\[[^\[\]]*\]\]`)
)

func scanBody(body string, sc scope) ([]entry, error) {
	lines := strings.Split(body, "\n")
	start, err := sc.title(lines)
	if err != nil {
		return nil, err
	}
	s := &scanner{scope: sc, prose: map[int][]string{}, target: -1}
	for i := start; i < len(lines); i++ {
		if err := s.scanLine(lines, i); err != nil {
			return nil, err
		}
	}
	return s.finish(), nil
}

func (sc scope) at(i int, what string) error {
	return detail(fmt.Sprintf("%s at line %d", what, sc.line+i))
}

func (sc scope) title(lines []string) (int, error) {
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		title, ok := strings.CutPrefix(line, "# ")
		if !ok {
			return 0, ErrMalformed.With(sc.at(i, "H1 title is missing"))
		}
		if norm.NFC.String(strings.TrimSpace(title)) != sc.id.value {
			return 0, ErrIdentityMismatch.With(sc.at(i, "H1 title"))
		}
		return i + 1, nil
	}
	return 0, ErrMalformed.With(detail("H1 title is missing"))
}

func (s *scanner) scanLine(lines []string, i int) error {
	switch {
	case s.fence != "":
		s.appendProse(lines[i])
		if closesFence(lines[i], s.fence) {
			s.fence = ""
		}
		return nil
	case s.inBlock:
		return s.blockLine(lines[i], i)
	}
	return s.bodyLine(lines, i)
}

func (s *scanner) blockLine(line string, i int) error {
	if strings.TrimSpace(line) == "" {
		return nil
	}
	if heading, ok := headingOf(line); ok {
		s.heading = heading
		return nil
	}
	p, value, ok := bullet(line)
	if !ok || value == "" {
		return ErrMalformed.With(s.at(i, "line in a ## block"))
	}
	o, err := textOrRef(value)
	if err != nil {
		return err
	}
	prop := Property{Role: RoleText, Many: true, Label: s.blockLabel(p)}
	if _, isRef := o.(Ref); isRef {
		prop = Property{Role: RoleLink, Label: prop.Label}
	}
	s.entries = append(s.entries, entry{predicate: p, property: prop, values: []Object{o}})
	return nil
}

func (s *scanner) blockLabel(p Predicate) string {
	if s.heading == p.Label() {
		return ""
	}
	return s.heading
}

func (s *scanner) bodyLine(lines []string, i int) error {
	line := lines[i]
	if marker := fenceMarker(line); marker != "" {
		s.fence = marker
		s.appendProse(line)
		return nil
	}
	if heading, ok := headingOf(line); ok {
		s.closeParagraph()
		s.active = false
		s.inBlock, s.heading = true, heading
		return nil
	}
	if p, value, ok := bullet(line); ok {
		s.closeParagraph()
		s.active = false
		return s.edge(p, value, i)
	}
	if p, value, ok := bodyEdge(line); ok {
		s.closeParagraph()
		s.active = false
		return s.edge(p, value, i)
	}
	if s.wholeLabel(lines, i) {
		return nil
	}
	if label, rest, p, ok := s.leadLabel(line); ok {
		s.openLabel(label, p)
		if rest == "" {
			return nil
		}
		return s.proseLine(rest)
	}
	return s.proseLine(line)
}

// wholeLabel recognises a line that is exactly "**Label**": a grouping
// label when the next line is a bullet (ignored, carries no data, CORE §4),
// otherwise the start of that label's text.
func (s *scanner) wholeLabel(lines []string, i int) bool {
	text, ok := labelText(lines[i])
	if !ok {
		return false
	}
	if i+1 < len(lines) {
		if _, _, next := bullet(lines[i+1]); next {
			return true
		}
	}
	p, ok := s.resolveLabel(text)
	if !ok {
		return false
	}
	s.openLabel(text, p)
	return true
}

// leadLabel recognises a paragraph lead "**Label**: rest" (BUG-001). A bold
// lead without a colon is never a label, regardless of the schema.
func (s *scanner) leadLabel(line string) (string, string, Predicate, bool) {
	text, rest, ok := boldLeadText(line)
	if !ok {
		return "", "", Predicate{}, false
	}
	p, ok := s.resolveLabel(text)
	return text, rest, p, ok
}

func (s *scanner) resolveLabel(text string) (Predicate, bool) {
	if s.schema != nil {
		if p, ok := s.schema.Predicate(s.class, text); ok && !p.IsZero() {
			return p, true
		}
	}
	return predicateOfLabel(text)
}

func (s *scanner) openLabel(label string, p Predicate) {
	s.closeParagraph()
	if label == p.Label() {
		label = ""
	}
	s.predicate, s.textProp, s.active = p, Property{Role: RoleText, Label: label}, true
}

func (s *scanner) edge(p Predicate, value string, i int) error {
	if value == "" {
		return ErrMalformed.With(s.at(i, "bullet without value"))
	}
	o, err := textOrRef(value)
	if err != nil {
		return err
	}
	s.entries = append(s.entries, entry{predicate: p, property: Property{Role: RoleEdge}, values: []Object{o}})
	return nil
}

func (s *scanner) slot(p Predicate, prop Property) int {
	s.entries = append(s.entries, entry{predicate: p, property: prop})
	return len(s.entries) - 1
}

func (s *scanner) proseLine(line string) error {
	if strings.TrimSpace(line) == "" {
		s.closeParagraph()
		return nil
	}
	s.appendProse(line)
	for _, m := range inlineEdge.FindAllStringSubmatch(line, -1) {
		o, err := parseRef(m[2])
		if err != nil {
			return err
		}
		s.entries = append(s.entries, entry{predicate: Predicate{value: m[1]}, property: Property{Role: RoleEdge}, values: []Object{o}})
	}
	return nil
}

// appendProse feeds one line into the paragraph currently being
// accumulated, opening the default-text paragraph lazily when no label is
// active (BUG-001: one Text per paragraph).
func (s *scanner) appendProse(line string) {
	if s.target < 0 {
		if !s.active {
			s.predicate, s.textProp, s.active = s.defaultText(), Property{Role: RoleText, Default: true}, true
		}
		s.target = s.slot(s.predicate, s.textProp)
	}
	s.prose[s.target] = append(s.prose[s.target], line)
}

// closeParagraph finalises the paragraph currently being accumulated, if
// any, into one Text entry. The active predicate (default or labelled)
// stays in effect, so the next paragraph after a blank line reopens under
// the same predicate (BUG-001).
func (s *scanner) closeParagraph() {
	if s.target < 0 {
		return
	}
	if text := trimBlankLines(s.prose[s.target]); text != "" {
		s.entries[s.target].values = []Object{Text(text)}
	}
	s.target = -1
}

func (s *scanner) defaultText() Predicate {
	if s.schema != nil {
		if p, ok := s.schema.Predicate(s.class, ""); ok && !p.IsZero() {
			return p
		}
	}
	return Predicate{value: "text"}
}

func (s *scanner) finish() []entry {
	s.closeParagraph()
	return s.entries
}

func trimBlankLines(lines []string) string {
	first, last := 0, len(lines)-1
	for first <= last && strings.TrimSpace(lines[first]) == "" {
		first++
	}
	for last >= first && strings.TrimSpace(lines[last]) == "" {
		last--
	}
	return strings.Join(lines[first:last+1], "\n")
}

func headingOf(line string) (string, bool) {
	heading, ok := strings.CutPrefix(line, "## ")
	return strings.TrimSpace(heading), ok
}

func bullet(line string) (Predicate, string, bool) {
	rest, ok := strings.CutPrefix(line, "- ")
	if !ok {
		return Predicate{}, "", false
	}
	name, value, ok := strings.Cut(rest, "::")
	p, err := NewPredicate(name)
	return p, strings.TrimSpace(value), ok && err == nil
}

func bodyEdge(line string) (Predicate, string, bool) {
	name, value, ok := strings.Cut(line, "::")
	if !ok {
		return Predicate{}, "", false
	}
	p, err := NewPredicate(name)
	value = strings.TrimSpace(value)
	return p, value, err == nil && refLexical(value)
}

func labelText(line string) (string, bool) {
	if len(line) <= 4 || !strings.HasPrefix(line, "**") || !strings.HasSuffix(line, "**") {
		return "", false
	}
	text := line[2 : len(line)-2]
	return text, !strings.Contains(text, "*") && strings.TrimSpace(text) == text
}

// boldLeadText recognises a paragraph lead "**Label**: rest", the inline
// form of a bold label (BUG-001). A bold lead without a colon right after
// the closing "**" is never a label, whatever the schema says.
func boldLeadText(line string) (string, string, bool) {
	if !strings.HasPrefix(line, "**") {
		return "", "", false
	}
	tail := line[2:]
	end := strings.Index(tail, "**")
	if end <= 0 {
		return "", "", false
	}
	label := tail[:end]
	if strings.Contains(label, "*") || strings.TrimSpace(label) != label {
		return "", "", false
	}
	rest, ok := strings.CutPrefix(tail[end+2:], ":")
	if !ok {
		return "", "", false
	}
	return label, strings.TrimSpace(rest), true
}

// untypedLinks scans text for bare [[Target]] / [[Target|display]] links
// (CORE §7), skipping fenced code and the typed inline form
// "[p:: [[T]]]" (BUG-001). A link whose target is not a valid ID is
// silently dropped: it is a best-effort, derived view, not a triple.
func untypedLinks(text string) []Ref {
	var refs []Ref
	fence := ""
	for _, line := range strings.Split(text, "\n") {
		switch {
		case fence != "":
			if closesFence(line, fence) {
				fence = ""
			}
			continue
		case fenceMarker(line) != "":
			fence = fenceMarker(line)
			continue
		}
		typed := inlineEdge.FindAllStringIndex(line, -1)
		for _, m := range untypedLinkRe.FindAllStringIndex(line, -1) {
			if withinAny(typed, m) {
				continue
			}
			if o, err := parseRef(line[m[0]:m[1]]); err == nil {
				refs = append(refs, o.(Ref))
			}
		}
	}
	return refs
}

func withinAny(spans [][]int, m []int) bool {
	for _, s := range spans {
		if m[0] >= s[0] && m[1] <= s[1] {
			return true
		}
	}
	return false
}

func fenceMarker(line string) string {
	trimmed := strings.TrimLeft(line, " ")
	switch {
	case len(line)-len(trimmed) > 3:
		return ""
	case strings.HasPrefix(trimmed, "```"):
		return "```"
	case strings.HasPrefix(trimmed, "~~~"):
		return "~~~"
	}
	return ""
}

func closesFence(line, marker string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, marker) && strings.Trim(trimmed, marker[:1]) == ""
}

func atxHeading(line string) bool {
	hashes := len(line) - len(strings.TrimLeft(line, "#"))
	return hashes >= 1 && hashes <= 6 && (len(line) == hashes || line[hashes] == ' ')
}
