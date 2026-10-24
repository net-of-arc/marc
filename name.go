//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import "strings"

// Predicate is a camelCase property name (^[a-z][A-Za-z0-9]*$), the verb of a
// triple. Keywords such as "@id" can never be a Predicate. The zero Predicate
// is invalid. Predicate is an immutable value, comparable, usable as a map key
// and safe for concurrent use.
type Predicate struct{ value string }

// NewPredicate validates that s is camelCase and returns it as a Predicate.
// It returns ErrInvalidName otherwise; the error never contains s.
func NewPredicate(s string) (Predicate, error) {
	if !camelCase(s, isLower) {
		return Predicate{}, ErrInvalidName.With(detail("predicate is not camelCase"))
	}
	return Predicate{value: s}, nil
}

// MustPredicate is NewPredicate that panics on an invalid name. It is meant
// for package-level constants and tests only.
func MustPredicate(s string) Predicate {
	p, err := NewPredicate(s)
	if err != nil {
		panic(err)
	}
	return p
}

// String returns the predicate name.
func (p Predicate) String() string { return p.value }

// IsZero reports whether p is the zero, invalid Predicate.
func (p Predicate) IsZero() bool { return p.value == "" }

// Label returns the label derived from the name: the name split at
// lower→upper and letter→digit boundaries with each word capitalised, so
// "mentionedIn" becomes "Mentioned In" and "rfc8446" becomes "Rfc 8446".
func (p Predicate) Label() string {
	var sb strings.Builder
	for i := 0; i < len(p.value); i++ {
		c := p.value[i]
		switch {
		case i == 0:
			sb.WriteByte(toUpper(c))
		case labelBoundary(p.value[i-1], c):
			sb.WriteByte(' ')
			sb.WriteByte(toUpper(c))
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func labelBoundary(prev, c byte) bool {
	return (isLower(prev) && isUpper(c)) || (isLetter(prev) && isDigit(c))
}

func predicateOfLabel(label string) (Predicate, bool) {
	words := strings.Split(label, " ")
	var sb strings.Builder
	for i, word := range words {
		if word == "" || !camelCase(word, isAlphanumeric) {
			return Predicate{}, false
		}
		if i == 0 {
			sb.WriteString(strings.ToLower(word))
		} else {
			sb.WriteByte(toUpper(word[0]))
			sb.WriteString(word[1:])
		}
	}
	p, err := NewPredicate(sb.String())
	return p, err == nil
}

// Class is an UpperCamelCase class name (^[A-Z][A-Za-z0-9]*$), the "@type" of
// a node. The zero Class is invalid. Class is an immutable value, comparable,
// usable as a map key and safe for concurrent use.
type Class struct{ value string }

// NewClass validates that s is UpperCamelCase and returns it as a Class.
// It returns ErrInvalidName otherwise; the error never contains s.
func NewClass(s string) (Class, error) {
	if !camelCase(s, isUpper) {
		return Class{}, ErrInvalidName.With(detail("class is not UpperCamelCase"))
	}
	return Class{value: s}, nil
}

// MustClass is NewClass that panics on an invalid name. It is meant for
// package-level constants and tests only.
func MustClass(s string) Class {
	c, err := NewClass(s)
	if err != nil {
		panic(err)
	}
	return c
}

// String returns the class name.
func (c Class) String() string { return c.value }

// IsZero reports whether c is the zero, invalid Class.
func (c Class) IsZero() bool { return c.value == "" }

func comparePredicates(a, b Predicate) int { return strings.Compare(a.value, b.value) }

func camelCase(s string, head func(byte) bool) bool {
	if s == "" || !head(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isAlphanumeric(s[i]) {
			return false
		}
	}
	return true
}

func isLower(c byte) bool        { return 'a' <= c && c <= 'z' }
func isUpper(c byte) bool        { return 'A' <= c && c <= 'Z' }
func isDigit(c byte) bool        { return '0' <= c && c <= '9' }
func isLetter(c byte) bool       { return isLower(c) || isUpper(c) }
func isAlphanumeric(c byte) bool { return isLetter(c) || isDigit(c) }

func toUpper(c byte) byte {
	if isLower(c) {
		return c - 'a' + 'A'
	}
	return c
}
