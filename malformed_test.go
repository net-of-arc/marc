//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"bytes"
	"errors"
	"io"
	"math/rand/v2"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/net-of-arc/marc"
)

func TestMalformedFiles(t *testing.T) {
	// Scenario 5.1 from specs/001-markdown-triples-codec/spec.md
	// Scenario 5.2 from specs/001-markdown-triples-codec/spec.md
	// Scenario 5.3 from specs/001-markdown-triples-codec/spec.md
	// Scenario 5.4 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	for file, expected := range map[string]error{
		"no-frontmatter.md":           marc.ErrMalformed,
		"unterminated-frontmatter.md": marc.ErrMalformed,
		"missing-id.md":               marc.ErrMalformed,
		"missing-type.md":             marc.ErrMalformed,
		"missing-h1.md":               marc.ErrMalformed,
		"empty.md":                    marc.ErrMalformed,
		"yaml-alias.md":               marc.ErrMalformed,
		"nested-map.md":               marc.ErrMalformed,
		"h1-mismatch.md":              marc.ErrIdentityMismatch,
		"invalid-id.md":               marc.ErrInvalidID,
		"invalid-predicate.md":        marc.ErrInvalidName,
		"invalid-class.md":            marc.ErrInvalidName,
	} {
		t.Run(file, func(t *testing.T) {
			_, err := marc.Unmarshal(fixture(t, "testdata/malformed/"+file))
			it.Then(t).Should(it.True(errors.Is(err, expected)))
		})
	}
}

func TestMalformedIdentity(t *testing.T) {
	// Scenario 5.3 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	for _, s := range []string{
		"a/b", `a\b`, "a:b", "a*b", "a?b", `a"b`, "a<b", "a>b", "a|b",
		"a.b", "a#b", "a^b", "a[b", "a]b", " lead", "trail ", "two  spaces",
	} {
		_, err := marc.NewID(s)
		it.Then(t).Should(it.True(errors.Is(err, marc.ErrInvalidID)))
	}
}

func TestMalformedNames(t *testing.T) {
	// Scenario 5.4 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	_, errPredicate := marc.NewPredicate("Bad")
	_, errClass := marc.NewClass("bad")

	it.Then(t).Should(
		it.True(errors.Is(errPredicate, marc.ErrInvalidName)),
		it.True(errors.Is(errClass, marc.ErrInvalidName)),
	)
}

type countingReader struct {
	r    io.Reader
	read int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.read += n
	return n, err
}

func TestMalformedTooLarge(t *testing.T) {
	// Scenario 5.5 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	c, err := marc.New(marc.Config{MaxSize: 64})
	it.Then(t).Must(it.Nil(err))

	r := &countingReader{r: bytes.NewReader(bytes.Repeat([]byte("a"), 4096))}
	_, err = c.Decode(r)

	it.Then(t).Should(
		it.True(errors.Is(err, marc.ErrTooLarge)),
		it.LessOrEqual(r.read, 65),
	)
}

func TestMalformedProseStructure(t *testing.T) {
	// Scenario 5.6 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	schema := marc.Properties{marc.MustPredicate("text"): {Role: marc.RoleText, Default: true}}

	for prose, expected := range map[string]error{
		"Intro.\n## Heading\nMore.":           marc.ErrProseStructure,
		"Intro.\n- p:: x\nMore.":              marc.ErrProseStructure,
		"Intro.\n```\n## Heading\n```\nMore.": nil,
		"Intro.\n```\n- p:: x\n```\nMore.":    nil,
	} {
		n := node(t, "X", triple("X", "text", marc.Text(prose)))
		_, err := codec(t, schema).Marshal(n)
		if expected == nil {
			it.Then(t).Should(it.Nil(err))
		} else {
			it.Then(t).Should(it.True(errors.Is(err, expected)))
		}
	}
}

func TestMalformedArbitraryBytes(t *testing.T) {
	// Scenario 5.7 from specs/001-markdown-triples-codec/spec.md
	t.Parallel()

	rnd := rand.New(rand.NewPCG(5, 7))
	alphabet := []byte("---\n\"@id\": X\n\"@type\": Entity\n# X[]:-*&|>{}`~ab\r\t")
	for range 1000 {
		b := make([]byte, rnd.IntN(256))
		for i := range b {
			if rnd.IntN(2) == 0 {
				b[i] = alphabet[rnd.IntN(len(alphabet))]
			} else {
				b[i] = byte(rnd.IntN(256))
			}
		}
		n, err := marc.Unmarshal(b)
		it.Then(t).Should(it.True(err != nil || !n.ID().IsZero()))
	}
}
