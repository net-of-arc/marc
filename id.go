//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// ID is the identity of a node, its "@id": the name of the subject every
// triple of the node is about.
//
// A non-zero ID always satisfies CORE §5: it is NFC-normalised, contains none
// of / \ : * ? " < > | . # ^ [ ], no control character, and neither begins nor
// ends with a space nor contains two consecutive spaces. IDs compare exactly
// after NFC; graph-level case folding is the caller's concern. The zero ID is
// invalid. ID is an immutable value, comparable, usable as a map key and safe
// for concurrent use.
type ID struct{ value string }

const forbiddenInID = "/\\:*?\"<>|.#^[]"

// NewID validates s against CORE §5 and returns its NFC-normalised identity.
// It returns ErrInvalidID when s is empty, not valid UTF-8, or breaks a CORE §5
// rule; the error never contains s.
func NewID(s string) (ID, error) {
	if !utf8.ValidString(s) {
		return ID{}, ErrInvalidID.With(detail("not UTF-8"))
	}
	s = norm.NFC.String(s)
	if !validID(s) {
		return ID{}, ErrInvalidID.With(detail("violates CORE §5"))
	}
	return ID{value: s}, nil
}

func validID(s string) bool {
	switch {
	case s == "":
		return false
	case strings.HasPrefix(s, " "), strings.HasSuffix(s, " "):
		return false
	case strings.Contains(s, "  "), strings.ContainsAny(s, forbiddenInID):
		return false
	}
	return !strings.ContainsFunc(s, unicode.IsControl)
}

// MustID is NewID that panics on an invalid identity. It is meant for
// package-level constants and tests only.
func MustID(s string) ID {
	id, err := NewID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the NFC-normalised identity as written in "@id" and the H1.
func (id ID) String() string { return id.value }

// IsZero reports whether id is the zero, invalid ID.
func (id ID) IsZero() bool { return id.value == "" }
