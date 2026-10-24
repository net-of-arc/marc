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

func TestNewPredicate(t *testing.T) {
	for _, s := range []string{"tags", "mentionedIn", "rfc8446", "a"} {
		p, err := NewPredicate(s)
		it.Then(t).Should(it.Nil(err), it.Equal(p.String(), s))
	}
	for _, s := range []string{"Tags", "bad_name", "@id", "", "1abc", "has space", "é"} {
		_, err := NewPredicate(s)
		it.Then(t).Should(it.True(errors.Is(err, ErrInvalidName)))
	}
}

func TestNewClass(t *testing.T) {
	for _, s := range []string{"Entity", "Property", "A1"} {
		c, err := NewClass(s)
		it.Then(t).Should(it.Nil(err), it.Equal(c.String(), s))
	}
	for _, s := range []string{"entity", "", "Bad-Name", "1A"} {
		_, err := NewClass(s)
		it.Then(t).Should(it.True(errors.Is(err, ErrInvalidName)))
	}
}

func TestMustNamesPanic(t *testing.T) {
	it.Then(t).Should(
		it.Fail(func() { MustPredicate("Bad") }),
		it.Fail(func() { MustClass("bad") }),
	)
}

func TestNamesZero(t *testing.T) {
	it.Then(t).Should(
		it.True(Predicate{}.IsZero()),
		it.True(Class{}.IsZero()),
		it.True(!MustPredicate("tags").IsZero()),
		it.True(!MustClass("Entity").IsZero()),
	)
}

func TestPredicateLabel(t *testing.T) {
	for p, label := range map[string]string{
		"mentionedIn": "Mentioned In",
		"keyFindings": "Key Findings",
		"tags":        "Tags",
		"rfc8446":     "Rfc 8446",
		"aBC":         "A BC",
	} {
		it.Then(t).Should(it.Equal(MustPredicate(p).Label(), label))
	}
}

func TestLabelJoinInvertsLabel(t *testing.T) {
	for _, s := range []string{"mentionedIn", "keyFindings", "tags", "rfc8446", "aBC", "abc1D", "aB1", "x9y"} {
		p, ok := predicateOfLabel(MustPredicate(s).Label())
		it.Then(t).Should(it.True(ok), it.Equal(p.String(), s))
	}
	for _, s := range []string{"", "Key  Findings", " Key", "Key-Findings", "123 Foo"} {
		_, ok := predicateOfLabel(s)
		it.Then(t).Should(it.True(!ok))
	}
}
