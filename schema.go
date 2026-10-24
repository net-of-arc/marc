//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"maps"
	"slices"
	"strings"
)

// Role is the position a predicate renders to (CORE §4).
type Role uint8

const (
	// RoleMeta renders a predicate as a front-matter scalar or array. It is
	// the zero Role.
	RoleMeta Role = iota
	// RoleText renders a predicate as prose in the body: the unlabelled
	// default text, a **Label** paragraph, or with Many a "## Label" block of
	// statements.
	RoleText
	// RoleEdge renders a predicate as bullets of the heterogeneous edge block.
	RoleEdge
	// RoleLink renders a predicate as its own "## Label" block of bullets.
	RoleLink
)

// String returns the CORE name of the role: meta, text, edge or link.
func (r Role) String() string {
	switch r {
	case RoleMeta:
		return "meta"
	case RoleText:
		return "text"
	case RoleEdge:
		return "edge"
	case RoleLink:
		return "link"
	}
	return "invalid"
}

// Property is the rendering rule for one predicate.
//
// Label is the display label of a **Label** paragraph or a "## Label" block;
// "" means the label derived from the predicate name. Default marks the
// class's default (unlabelled) text predicate and is legal only with RoleText
// and Many false. Many renders a meta predicate as a YAML array even when it
// holds one value, and a text predicate as a "## Label" block of one
// statement bullet per value. A Label must be a single line without "*" and
// without surrounding spaces. Property is a comparable value.
type Property struct {
	Role    Role
	Label   string
	Default bool
	Many    bool
}

// Schema is the application-owned vocabulary that tells the codec how each
// predicate renders on nodes of a class. Applications implement it over their
// own schema types; the codec keeps it by reference and never copies or
// mutates it. An implementation must be safe for concurrent use whenever the
// Codec using it is, and must keep its two methods consistent: Predicate(c,
// "") names the predicate whose Property(c, ·).Default is true, and labels
// are unique per class. Each Property it answers is validated when the codec
// obtains it (ErrInvalidProperty).
type Schema interface {
	// Property answers how p renders on nodes of class c; false makes the
	// encoder fall back to the observed property, then to meta for literals
	// and edge for references.
	Property(c Class, p Predicate) (Property, bool)

	// Predicate answers which predicate a **Label** paragraph names on nodes
	// of class c; label "" asks for the default-text predicate. False makes the
	// decoder use the camelCase join of the label words, or "text" for "".
	Predicate(c Class, label string) (Predicate, bool)
}

// Properties is the reference, class-agnostic Schema: a map from predicate to
// its Property. It is used as a read-only value; the codec never mutates it,
// so it is safe for concurrent use as long as the caller does not mutate it
// either.
type Properties map[Predicate]Property

// Property returns the entry for p, ignoring the class.
func (s Properties) Property(c Class, p Predicate) (Property, bool) {
	prop, ok := s[p]
	return prop, ok
}

// Predicate returns the predicate whose explicit or derived label equals
// label, or the Default entry when label is "". It ignores the class.
func (s Properties) Predicate(c Class, label string) (Predicate, bool) {
	for p, prop := range s {
		if (label == "" && prop.Default) || (label != "" && labelOf(p, prop) == label) {
			return p, true
		}
	}
	return Predicate{}, false
}

// Validate reports ErrInvalidProperty, naming the predicate, when an entry is
// invalid, when more than one entry is Default, or when two entries share a
// label after derivation. It returns nil for a consistent schema.
func (s Properties) Validate() error {
	labels := map[string]struct{}{}
	defaults := 0
	for _, p := range slices.SortedFunc(maps.Keys(s), comparePredicates) {
		if err := validProperty(p, s[p]); err != nil {
			return err
		}
		if _, dup := labels[labelOf(p, s[p])]; dup || (s[p].Default && defaults > 0) {
			return ErrInvalidProperty.With(detail(p.value))
		}
		labels[labelOf(p, s[p])] = struct{}{}
		defaults += btoi(s[p].Default)
	}
	return nil
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func labelOf(p Predicate, prop Property) string {
	if prop.Label != "" {
		return prop.Label
	}
	return p.Label()
}

func validProperty(p Predicate, prop Property) error {
	switch {
	case prop.Role > RoleLink:
		return ErrInvalidProperty.With(detail(p.value))
	case prop.Default && (prop.Role != RoleText || prop.Many):
		return ErrInvalidProperty.With(detail(p.value))
	case strings.ContainsAny(prop.Label, "*\n\r") || strings.TrimSpace(prop.Label) != prop.Label:
		return ErrInvalidProperty.With(detail(p.value))
	}
	return nil
}
