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
	"errors"
	"strings"
	"testing"

	"github.com/fogfish/it/v2"
)

func TestDecodeSizeLimit(t *testing.T) {
	c, err := New(Config{MaxSize: 32})
	it.Then(t).Must(it.Nil(err))

	src := bytes.Repeat([]byte("x"), 33)
	_, errDecode := c.Decode(bytes.NewReader(src))
	_, errUnmarshal := c.Unmarshal(src)
	_, errFits := c.Unmarshal(src[:32])

	it.Then(t).Should(
		it.True(errors.Is(errDecode, ErrTooLarge)),
		it.True(errors.Is(errUnmarshal, ErrTooLarge)),
		it.True(errors.Is(errFits, ErrMalformed)),
	)
}

func TestConfigValidation(t *testing.T) {
	_, errNegative := New(Config{MaxSize: -1})
	c, errZero := New(Config{})
	it.Then(t).Should(
		it.True(errors.Is(errNegative, ErrInvalidConfig)),
		it.String(errNegative.Error()).Contain("MaxSize"),
		it.Nil(errZero),
		it.Equal(c.maxSize, 4194304),
	)
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestDecodeReaderError(t *testing.T) {
	_, err := Decode(failingReader{})
	it.Then(t).Should(it.True(err != nil), it.String(err.Error()).Contain("boom"))
}

func TestDecodeRejects(t *testing.T) {
	head := "---\n\"@id\": X\n\"@type\": Entity\n"
	for src, expected := range map[string]error{
		head + "a: &x secret\nb: *x\n---\n# X\n":         ErrMalformed,
		head + "k: [&x secret]\n---\n# X\n":              ErrMalformed,
		head + "meta: {k: secret}\n---\n# X\n":           ErrMalformed,
		head + "Bad: secret\n---\n# X\n":                 ErrInvalidName,
		"---\n\"@id\": X\n\"@type\": secret\n---\n# X\n": ErrInvalidName,
		head + "---\n# X\n\n- p:: [[secret/value]]\n":    ErrInvalidID,
		head + "---\n# X\n\n## Block\nsecret line\n":     ErrMalformed,
		head + "---\n# X\n\n## Block\n- secret line\n":   ErrMalformed,
		head + "---\n# X\n\n- p::\n":                     ErrMalformed,
		head + "---\n# X\n\nsecret \xff\n":               ErrMalformed,
	} {
		_, err := Unmarshal([]byte(src))
		it.Then(t).Should(it.True(errors.Is(err, expected)))
		it.Then(t).ShouldNot(it.String(err.Error()).Contain("secret"))
	}
}

func TestDecodeErrorNamesLine(t *testing.T) {
	_, err := Unmarshal([]byte("---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\n## Block\nnot a bullet\n"))
	it.Then(t).Should(it.True(errors.Is(err, ErrMalformed)), it.String(err.Error()).Contain("line 8"))
}

func TestUnmarshalDoesNotRetainInput(t *testing.T) {
	src := []byte("---\n\"@id\": X\n\"@type\": Entity\ntag: value\n---\n# X\n\nProse.\n")
	n, err := Unmarshal(src)
	it.Then(t).Must(it.Nil(err))

	copy(src, strings.Repeat("z", len(src)))
	b, err := Marshal(n)
	it.Then(t).Must(it.Nil(err)).Should(
		it.Equal(string(b), "---\n\"@id\": X\n\"@type\": Entity\ntag: value\n---\n# X\n\nProse.\n"),
	)
}
