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

func TestNewIDAccepts(t *testing.T) {
	for _, s := range []string{"Transport Layer Security", "RFC 8446", "rescorla-2026-tls13", "Object (entity)", "Café"} {
		id, err := NewID(s)
		it.Then(t).Should(it.Nil(err), it.Equal(id.String(), s))
	}
}

func TestNewIDRejects(t *testing.T) {
	for _, s := range []string{
		"", "a/b", `a\b`, "a:b", "a*b", "a?b", `a"b`, "a<b", "a>b", "a|b",
		"a.b", "a#b", "a^b", "a[b", "a]b", " lead", "trail ", "two  spaces",
		"line\nbreak", "tab\tch", "nul\x00", "bad\xffutf8",
	} {
		_, err := NewID(s)
		it.Then(t).Should(it.True(errors.Is(err, ErrInvalidID)))
	}
}

func TestNewIDErrorOmitsValue(t *testing.T) {
	_, err := NewID("secret/value")
	it.Then(t).ShouldNot(it.String(err.Error()).Contain("secret"))
}

func TestNewIDNormalisesNFC(t *testing.T) {
	decomposed, errD := NewID("Café")
	composed, errC := NewID("Café")
	it.Then(t).Should(it.Nil(errD), it.Nil(errC), it.Equal(decomposed, composed))
}

func TestMustIDPanics(t *testing.T) {
	it.Then(t).Should(it.Fail(func() { MustID("a/b") }))
}

func TestIDZero(t *testing.T) {
	it.Then(t).Should(
		it.True(ID{}.IsZero()),
		it.True(!MustID("X").IsZero()),
	)
}
