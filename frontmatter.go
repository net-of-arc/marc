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
	"fmt"
	"strings"
	"time"
	"unicode"

	"go.yaml.in/yaml/v3"
)

type header struct {
	id         ID
	class      Class
	meta       []entry
	conflicts  []entry
	provenance []Attribution
}

type entry struct {
	predicate Predicate
	property  Property
	values    []Object
}

const (
	keyID         = "@id"
	keyType       = "@type"
	keyConflicts  = "@conflicts"
	keyProvenance = "@provenance"
)

func splitFrontMatter(src []byte) ([]byte, []byte, error) {
	text := bytes.ReplaceAll(bytes.TrimPrefix(src, []byte("\xef\xbb\xbf")), []byte("\r\n"), []byte("\n"))
	rest, ok := bytes.CutPrefix(text, []byte("---\n"))
	switch {
	case !ok:
		return nil, nil, ErrMalformed.With(detail("front-matter is missing"))
	case bytes.Equal(rest, []byte("---")):
		return nil, nil, nil
	case bytes.HasPrefix(rest, []byte("---\n")):
		return nil, rest[4:], nil
	}
	if i := bytes.Index(rest, []byte("\n---\n")); i >= 0 {
		return rest[:i+1], rest[i+5:], nil
	}
	if bytes.HasSuffix(rest, []byte("\n---")) {
		return rest[:len(rest)-3], nil, nil
	}
	return nil, nil, ErrMalformed.With(detail("front-matter is not terminated"))
}

func parseFrontMatter(front []byte) (header, error) {
	root, err := yamlMapping(front)
	if err != nil {
		return header{}, err
	}
	var h header
	for i := 0; i+1 < len(root.Content); i += 2 {
		if err := h.field(root.Content[i], root.Content[i+1]); err != nil {
			return header{}, err
		}
	}
	return h, h.complete()
}

func yamlMapping(front []byte) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(front, &doc); err != nil {
		return nil, ErrMalformed.With(detail("front-matter is not valid YAML"))
	}
	if doc.Kind == 0 {
		return &yaml.Node{Kind: yaml.MappingNode}, nil
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, ErrMalformed.With(detail("front-matter is not a mapping"))
	}
	return doc.Content[0], rejectAliases(doc.Content[0])
}

func rejectAliases(n *yaml.Node) error {
	if n.Kind == yaml.AliasNode || n.Anchor != "" {
		return malformedAt(n, "YAML anchor or alias")
	}
	for _, child := range n.Content {
		if err := rejectAliases(child); err != nil {
			return err
		}
	}
	return nil
}

func malformedAt(n *yaml.Node, what string) error {
	return ErrMalformed.With(detail(fmt.Sprintf("%s at front-matter line %d", what, n.Line)))
}

func (h *header) field(key, value *yaml.Node) error {
	if key.Kind != yaml.ScalarNode {
		return malformedAt(key, "non-scalar key")
	}
	var err error
	switch key.Value {
	case keyID:
		h.id, err = identity(value)
	case keyType:
		h.class, err = class(value)
	case keyConflicts:
		h.conflicts, err = parseConflicts(value)
	case keyProvenance:
		h.provenance, err = parseProvenance(value)
	default:
		h.meta, err = appendEntry(h.meta, key, value, RoleMeta)
	}
	return err
}

func (h *header) complete() error {
	if h.id.IsZero() {
		return ErrMalformed.With(detail(`"@id" is missing`))
	}
	if h.class.IsZero() {
		return ErrMalformed.With(detail(`"@type" is missing`))
	}
	return nil
}

func identity(n *yaml.Node) (ID, error) {
	if n.Kind != yaml.ScalarNode {
		return ID{}, malformedAt(n, `non-scalar "@id"`)
	}
	return NewID(n.Value)
}

func class(n *yaml.Node) (Class, error) {
	if n.Kind != yaml.ScalarNode {
		return Class{}, malformedAt(n, `non-scalar "@type"`)
	}
	return NewClass(n.Value)
}

func appendEntry(entries []entry, key, value *yaml.Node, role Role) ([]entry, error) {
	p, err := NewPredicate(key.Value)
	if err != nil {
		return nil, ErrInvalidName.With(detail(fmt.Sprintf("key at front-matter line %d", key.Line)))
	}
	values, many, err := atoms(value)
	if err != nil || len(values) == 0 {
		return entries, err
	}
	prop := Property{Role: role, Many: many}
	return append(entries, entry{predicate: p, property: prop, values: values}), nil
}

func atoms(n *yaml.Node) ([]Object, bool, error) {
	switch n.Kind {
	case yaml.ScalarNode:
		o, err := atom(n)
		return compact(o), false, err
	case yaml.SequenceNode:
		values, err := sequence(n)
		return values, true, err
	}
	return nil, false, malformedAt(n, "nested value")
}

func sequence(n *yaml.Node) ([]Object, error) {
	values := make([]Object, 0, len(n.Content))
	for _, item := range n.Content {
		if item.Kind != yaml.ScalarNode {
			return nil, malformedAt(item, "nested value")
		}
		o, err := atom(item)
		if err != nil {
			return nil, err
		}
		values = append(values, compact(o)...)
	}
	return values, nil
}

func compact(o Object) []Object {
	if o == nil {
		return nil
	}
	return []Object{o}
}

func atom(n *yaml.Node) (Object, error) {
	switch n.ShortTag() {
	case "!!null":
		return nil, nil
	case "!!str":
		return textOrRef(n.Value)
	case "!!int", "!!float":
		return number(n)
	case "!!bool":
		return boolean(n)
	case "!!timestamp":
		return timestamp(n)
	}
	return nil, malformedAt(n, "unsupported YAML tag")
}

func number(n *yaml.Node) (Object, error) {
	if n.Style&yaml.TaggedStyle != 0 {
		return ParseNumber(n.Value)
	}
	return Number{lexical: n.Value}, nil
}

func boolean(n *yaml.Node) (Object, error) {
	switch n.Value {
	case "true", "True", "TRUE":
		return Bool(true), nil
	case "false", "False", "FALSE":
		return Bool(false), nil
	}
	return nil, ErrInvalidObject.With(detail(fmt.Sprintf("boolean at front-matter line %d", n.Line)))
}

func timestamp(n *yaml.Node) (Object, error) {
	var t time.Time
	if err := n.Decode(&t); err != nil || t.IsZero() {
		return nil, ErrInvalidObject.With(detail(fmt.Sprintf("timestamp at front-matter line %d", n.Line)))
	}
	if strings.ContainsAny(n.Value, "Tt ") {
		return DateTimeOf(t), nil
	}
	return DateOf(t), nil
}

func textOrRef(s string) (Object, error) {
	if !refLexical(s) {
		return Text(s), nil
	}
	return parseRef(s)
}

func refLexical(s string) bool {
	if len(s) < 4 || !strings.HasPrefix(s, "[[") || !strings.HasSuffix(s, "]]") {
		return false
	}
	inner := s[2 : len(s)-2]
	target, _, _ := strings.Cut(inner, "|")
	return !strings.Contains(inner, "[[") && !strings.Contains(inner, "]]") && !strings.ContainsAny(target, "[]")
}

func parseRef(s string) (Object, error) {
	target, display, _ := strings.Cut(s[2:len(s)-2], "|")
	id, err := NewID(target)
	if err != nil {
		return nil, err
	}
	return RefTo(id).WithDisplay(display), nil
}

func parseConflicts(n *yaml.Node) ([]entry, error) {
	if n.Kind != yaml.MappingNode {
		return nil, malformedAt(n, `"@conflicts" is not a mapping`)
	}
	var entries []entry
	for i := 0; i+1 < len(n.Content); i += 2 {
		var err error
		if entries, err = appendEntry(entries, n.Content[i], n.Content[i+1], RoleMeta); err != nil {
			return nil, err
		}
	}
	for i := range entries {
		entries[i].property = Property{}
	}
	return entries, nil
}

func parseProvenance(n *yaml.Node) ([]Attribution, error) {
	if n.Kind != yaml.MappingNode {
		return nil, malformedAt(n, `"@provenance" is not a mapping`)
	}
	var attributions []Attribution
	for i := 0; i+1 < len(n.Content); i += 2 {
		doc, err := NewID(n.Content[i].Value)
		if err != nil {
			return nil, err
		}
		if attributions, err = appendAttributions(attributions, doc, n.Content[i+1]); err != nil {
			return nil, err
		}
	}
	return attributions, nil
}

func appendAttributions(attributions []Attribution, doc ID, n *yaml.Node) ([]Attribution, error) {
	if n.Kind != yaml.MappingNode {
		return nil, malformedAt(n, `"@provenance" document is not a mapping`)
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		entries, err := appendEntry(nil, n.Content[i], n.Content[i+1], RoleMeta)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			for _, o := range e.values {
				attributions = append(attributions, Attribution{Document: doc, Predicate: e.predicate, Object: o})
			}
		}
	}
	return attributions, nil
}

func renderFrontMatter(sb *strings.Builder, n Node, meta []entry) {
	sb.WriteString("---\n")
	sb.WriteString(`"@id": ` + yamlString(n.id.value) + "\n")
	sb.WriteString(`"@type": ` + yamlString(n.class.value) + "\n")
	for _, e := range meta {
		sb.WriteString(e.predicate.value + ": " + metaValue(e) + "\n")
	}
	renderConflicts(sb, n.conflicts)
	renderProvenance(sb, n.provenance)
	sb.WriteString("---\n")
}

func metaValue(e entry) string {
	if len(e.values) == 1 && !e.property.Many {
		return yamlScalar(e.values[0])
	}
	return flowSequence(e.values)
}

func flowSequence(values []Object) string {
	scalars := make([]string, len(values))
	for i, o := range values {
		scalars[i] = yamlScalar(o)
	}
	return "[" + strings.Join(scalars, ", ") + "]"
}

func renderConflicts(sb *strings.Builder, conflicts []valueSet) {
	if len(conflicts) == 0 {
		return
	}
	sb.WriteString(`"@conflicts":` + "\n")
	for _, set := range conflicts {
		renderValues(sb, "  ", set.predicate.value, set.values)
	}
}

func renderProvenance(sb *strings.Builder, provenance []Attribution) {
	if len(provenance) == 0 {
		return
	}
	sb.WriteString(`"@provenance":` + "\n")
	for i, a := range provenance {
		if i == 0 || provenance[i-1].Document != a.Document {
			sb.WriteString("  " + yamlString(a.Document.value) + ":\n")
		}
		if i == 0 || provenance[i-1].Document != a.Document || provenance[i-1].Predicate != a.Predicate {
			renderValues(sb, "    ", a.Predicate.value, attributedObjects(provenance[i:], a))
		}
	}
}

func attributedObjects(attributions []Attribution, head Attribution) []Object {
	var values []Object
	for _, a := range attributions {
		if a.Document != head.Document || a.Predicate != head.Predicate {
			break
		}
		values = append(values, a.Object)
	}
	return values
}

func renderValues(sb *strings.Builder, indent, key string, values []Object) {
	if !hasMultiline(values) {
		sb.WriteString(indent + key + ": " + flowSequence(values) + "\n")
		return
	}
	sb.WriteString(indent + key + ":\n")
	for _, o := range values {
		sb.WriteString(indent + "  - " + blockItem(o, indent+"    ") + "\n")
	}
}

func hasMultiline(values []Object) bool {
	for _, o := range values {
		if t, ok := o.(Text); ok && strings.Contains(string(t), "\n") {
			return true
		}
	}
	return false
}

func blockItem(o Object, indent string) string {
	t, ok := o.(Text)
	if !ok || !strings.Contains(string(t), "\n") || !blockable(string(t)) {
		return yamlScalar(o)
	}
	body, chomp := strings.TrimSuffix(string(t), "\n"), "-"
	switch {
	case strings.HasSuffix(string(t), "\n\n"):
		chomp = "+"
	case strings.HasSuffix(string(t), "\n"):
		chomp = ""
	}
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return "|" + chomp + "\n" + strings.Join(lines, "\n")
}

func blockable(s string) bool {
	first := strings.TrimLeft(s, "\n")
	if strings.TrimSpace(s) == "" || strings.HasPrefix(first, " ") || strings.HasPrefix(first, "\t") {
		return false
	}
	return !strings.ContainsFunc(s, func(r rune) bool { return r != '\n' && r != '\t' && !unicode.IsPrint(r) })
}

func yamlScalar(o Object) string {
	switch v := o.(type) {
	case Text:
		return yamlString(string(v))
	case Ref:
		return quoted(v.String())
	}
	return o.String()
}

func yamlString(s string) string {
	if plainSafe(s) {
		return s
	}
	return quoted(s)
}

func plainSafe(s string) bool {
	switch {
	case s == "", strings.TrimSpace(s) != s, reservedWord(s):
		return false
	case strings.ContainsAny(s, ":#,[]{}\"'?") || strings.ContainsRune("-|>!&*%@`", rune(s[0])):
		return false
	case strings.ContainsFunc(s, func(r rune) bool { return !unicode.IsPrint(r) }):
		return false
	case strings.ContainsRune("0123456789+.~<=", rune(s[0])):
		return resolvesToString(s)
	}
	return true
}

func reservedWord(s string) bool {
	switch strings.ToLower(s) {
	case "true", "false", "null", "~", "yes", "no", "on", "off", "y", "n":
		return true
	}
	return false
}

func resolvesToString(s string) bool {
	var doc yaml.Node
	if yaml.Unmarshal([]byte(s), &doc) != nil || len(doc.Content) != 1 {
		return false
	}
	n := doc.Content[0]
	return n.Kind == yaml.ScalarNode && n.Style == 0 && n.ShortTag() == "!!str" && n.Value == s
}

func quoted(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range s {
		sb.WriteString(escape(r))
	}
	sb.WriteByte('"')
	return sb.String()
}

func escape(r rune) string {
	switch {
	case r == '"':
		return `\"`
	case r == '\\':
		return `\\`
	case r == '\n':
		return `\n`
	case r == '\t':
		return `\t`
	case r == '\r':
		return `\r`
	case unicode.IsPrint(r):
		return string(r)
	case r > 0xFFFF:
		return fmt.Sprintf(`\U%08X`, r)
	}
	return fmt.Sprintf(`\u%04X`, r)
}
