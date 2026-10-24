//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"cmp"
	"iter"
	"math"
	"strconv"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
	"golang.org/x/text/unicode/norm"
)

// Object is the value of a triple. It is a sealed sum type: the only
// implementations are Ref, Text, Number, Bool, Date and DateTime, and no type
// outside this package can implement it. Clients branch on the variant with a
// type switch or filter a sequence with Select.
//
// String returns the canonical rendering of the value. Two objects are equal
// when their CORE §8.5 equality keys are equal: references by target ignoring
// display text, text by NFC with surrounding and per-line trailing whitespace
// removed, dates and datetimes by value, other literals by trimmed written
// form; objects of different variants are never equal. Every variant is an
// immutable value and safe for concurrent use.
type Object interface {
	String() string
	key() string
}

// Ref is a reference to another node: its target identity plus optional
// presentation-only display text, written [[Target]] or [[Target|display]].
// The zero Ref is invalid and is rejected by NewNode.
type Ref struct {
	target  ID
	display string
}

// RefTo returns a reference to target without display text.
func RefTo(target ID) Ref { return Ref{target: target} }

// WithDisplay returns a copy of r carrying the display text. Display text
// does not take part in equality.
func (r Ref) WithDisplay(text string) Ref { return Ref{target: r.target, display: text} }

// Target returns the identity of the referenced node.
func (r Ref) Target() ID { return r.target }

// Display returns the display text, or "" when there is none.
func (r Ref) Display() string { return r.display }

// String renders r as [[Target]] or [[Target|display]].
func (r Ref) String() string {
	if r.display == "" {
		return "[[" + r.target.value + "]]"
	}
	return "[[" + r.target.value + "|" + r.display + "]]"
}

func (r Ref) key() string { return "r:" + r.target.value }

// Text is a string literal, and the value of prose and statement predicates.
// Every valid UTF-8 string is a legal Text.
type Text string

// String returns the text verbatim.
func (t Text) String() string { return string(t) }

func (t Text) key() string {
	lines := strings.Split(norm.NFC.String(string(t)), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t\r")
	}
	return "t:" + strings.TrimSpace(strings.Join(lines, "\n"))
}

// Bool is a boolean literal.
type Bool bool

// String renders b as true or false.
func (b Bool) String() string { return strconv.FormatBool(bool(b)) }

func (b Bool) key() string { return "b:" + b.String() }

// Number is a numeric literal kept in its exact written (lexical) form. The
// zero Number is invalid and is rejected by NewNode.
type Number struct{ lexical string }

// ParseNumber accepts any YAML 1.2 core integer or float lexical form, such
// as 42, -7, 1.5, 1e3, 0x1F, 0o17 or .inf, after trimming surrounding
// whitespace. It returns ErrInvalidObject for anything else; the error never
// contains s.
func ParseNumber(s string) (Number, error) {
	s = strings.TrimSpace(s)
	if !numberLexical(s) {
		return Number{}, ErrInvalidObject.With(detail("not a number"))
	}
	return Number{lexical: s}, nil
}

func numberLexical(s string) bool {
	if s == "" || strings.ContainsAny(s, "\n\r\t ") {
		return false
	}
	var doc yaml.Node
	if yaml.Unmarshal([]byte(s), &doc) != nil || len(doc.Content) != 1 {
		return false
	}
	n := doc.Content[0]
	tag := n.ShortTag()
	return n.Kind == yaml.ScalarNode && n.Style == 0 && n.Value == s && (tag == "!!int" || tag == "!!float")
}

// Int64 returns n as an int64 when its lexical form is an integer that fits.
func (n Number) Int64() (int64, bool) {
	i, err := strconv.ParseInt(strings.ReplaceAll(n.lexical, "_", ""), 0, 64)
	return i, err == nil
}

// Float64 returns n as a float64, including the YAML forms .inf, -.inf and
// .nan.
func (n Number) Float64() (float64, bool) {
	if i, ok := n.Int64(); ok {
		return float64(i), true
	}
	switch strings.ToLower(strings.TrimLeft(n.lexical, "+")) {
	case ".inf":
		return math.Inf(1), true
	case "-.inf":
		return math.Inf(-1), true
	case ".nan":
		return math.NaN(), true
	}
	f, err := strconv.ParseFloat(strings.ReplaceAll(n.lexical, "_", ""), 64)
	return f, err == nil
}

// String returns the lexical form.
func (n Number) String() string { return n.lexical }

func (n Number) key() string { return "n:" + n.lexical }

// Date is a civil calendar date. The zero Date is invalid and is rejected by
// NewNode.
type Date struct{ t time.Time }

// DateOf returns the date of t, keeping t's year, month and day in t's own
// location.
func DateOf(t time.Time) Date {
	return Date{t: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

// Time returns the date as midnight UTC.
func (d Date) Time() time.Time { return d.t }

// String renders d as YYYY-MM-DD.
func (d Date) String() string { return d.t.Format(time.DateOnly) }

func (d Date) key() string { return "d:" + d.String() }

// DateTime is an instant, normalised to UTC. The zero DateTime is invalid and
// is rejected by NewNode.
type DateTime struct{ t time.Time }

// DateTimeOf returns the instant t normalised to UTC.
func DateTimeOf(t time.Time) DateTime { return DateTime{t: t.UTC()} }

// Time returns the instant in UTC.
func (d DateTime) Time() time.Time { return d.t }

// String renders d as RFC 3339 in UTC, with fractional seconds only when they
// are non-zero.
func (d DateTime) String() string { return d.t.Format(time.RFC3339Nano) }

func (d DateTime) key() string { return "s:" + d.String() }

// Select filters seq down to the objects of variant T, typed as T. It yields
// lazily and stops when the consumer stops.
func Select[T Object](seq iter.Seq[Object]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for o := range seq {
			if v, ok := o.(T); ok && !yield(v) {
				return
			}
		}
	}
}

func compareObjects(a, b Object) int {
	if c := cmp.Compare(variantTag(a), variantTag(b)); c != 0 {
		return c
	}
	switch x := a.(type) {
	case Date:
		return x.t.Compare(b.(Date).t)
	case DateTime:
		return x.t.Compare(b.(DateTime).t)
	case Text:
		return strings.Compare(norm.NFC.String(string(x)), norm.NFC.String(string(b.(Text))))
	}
	return strings.Compare(a.String(), b.String())
}

func variantTag(o Object) byte {
	switch o.(type) {
	case Bool:
		return 'b'
	case Date:
		return 'd'
	case Number:
		return 'n'
	case Ref:
		return 'r'
	case DateTime:
		return 's'
	}
	return 't'
}

func zeroObject(o Object) bool {
	switch v := o.(type) {
	case nil:
		return true
	case Ref:
		return v.target.IsZero()
	case Number:
		return v.lexical == ""
	case Date:
		return v.t.IsZero()
	case DateTime:
		return v.t.IsZero()
	}
	return false
}
