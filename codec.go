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
	"io"
)

// Codec decodes CORE Markdown nodes into Nodes and encodes Nodes into
// canonical CORE Markdown. A Codec is immutable; it keeps Config.Schema by
// reference and is safe for concurrent use whenever that Schema is. The
// package-level Decode, Unmarshal, Encode and Marshal use a Codec built from
// the zero Config.
type Codec struct {
	schema  Schema
	maxSize int
}

var std = &Codec{maxSize: defaultMaxSize}

// New builds a Codec from cfg. It returns ErrInvalidConfig, naming the field,
// when cfg is invalid.
func New(cfg Config) (*Codec, error) {
	size, err := cfg.maxSize()
	if err != nil {
		return nil, err
	}
	return &Codec{schema: cfg.Schema, maxSize: size}, nil
}

// Decode reads one node file from r, reading at most MaxSize+1 bytes, and
// decodes it. It does not retain r. It returns ErrTooLarge when the input
// exceeds MaxSize, the reader's own error wrapped, or any error of Unmarshal.
func (c *Codec) Decode(r io.Reader) (Node, error) {
	b, err := io.ReadAll(io.LimitReader(r, int64(c.maxSize)+1))
	if err != nil {
		return Node{}, errReader.With(err)
	}
	return c.Unmarshal(b)
}

// Unmarshal decodes one node file. It does not retain b. It returns
// ErrTooLarge, ErrMalformed, ErrIdentityMismatch, ErrInvalidID,
// ErrInvalidName or ErrInvalidObject; messages name the offending element
// and line, never its value.
func (c *Codec) Unmarshal(b []byte) (Node, error) {
	if len(b) > c.maxSize {
		return Node{}, ErrTooLarge.With(detail("input"))
	}
	return c.decode(b)
}

// Encode writes the canonical node file of n to w. It returns ErrMalformed
// for the zero Node, ErrProseStructure or ErrInvalidProperty, or the writer's
// own error wrapped.
func (c *Codec) Encode(w io.Writer, n Node) error {
	b, err := c.Marshal(n)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return errWriter.With(err)
	}
	return nil
}

// Marshal returns the canonical node file of n, with the errors of Encode.
func (c *Codec) Marshal(n Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := c.encode(&buf, n); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode decodes one node file from r with the default Codec.
func Decode(r io.Reader) (Node, error) { return std.Decode(r) }

// Unmarshal decodes one node file with the default Codec. It does not retain b.
func Unmarshal(b []byte) (Node, error) { return std.Unmarshal(b) }

// Encode writes the canonical node file of n to w with the default Codec.
func Encode(w io.Writer, n Node) error { return std.Encode(w, n) }

// Marshal returns the canonical node file of n with the default Codec.
func Marshal(n Node) ([]byte, error) { return std.Marshal(n) }
