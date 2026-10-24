//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import "github.com/fogfish/faults"

const (
	// ErrMalformed is returned by the decoder when the input is not a CORE node:
	// missing or unterminated front-matter, a YAML syntax error, YAML anchors or
	// aliases, nested maps outside "@conflicts"/"@provenance", a missing "@id"
	// or "@type", a missing H1 title, a malformed "## Label" block or empty
	// input. The encoder returns it for the zero Node.
	ErrMalformed = faults.Type("malformed node")

	// ErrIdentityMismatch is returned by the decoder when the H1 title differs
	// from the "@id" of the node.
	ErrIdentityMismatch = faults.Type("title does not match @id")

	// ErrInvalidID is returned when an identity violates CORE §5, either on
	// construction with NewID or when an "@id", a link target or a provenance
	// document is decoded.
	ErrInvalidID = faults.Type("invalid identity")

	// ErrInvalidName is returned when a predicate is not camelCase or a class is
	// not UpperCamelCase, either on construction or while decoding.
	ErrInvalidName = faults.Type("invalid predicate or class name")

	// ErrInvalidObject is returned when a nil or zero object enters a Node, when
	// a text is not valid UTF-8, or when ParseNumber rejects a lexical form.
	ErrInvalidObject = faults.Type("invalid object")

	// ErrForeignSubject is returned by NewNode when a triple's subject differs
	// from the node's identity.
	ErrForeignSubject = faults.Type("triple subject differs from node identity")

	// ErrTooLarge is returned by the decoder when the input exceeds the
	// configured size limit.
	ErrTooLarge = faults.Type("input exceeds size limit")

	// ErrProseStructure is returned by the encoder when a prose value contains a
	// heading, a predicate bullet, a typed body-form edge, a bold label line or
	// an unterminated fence outside fenced code.
	ErrProseStructure = faults.Type("prose contains structural markup")

	// ErrInvalidProperty is returned when a Schema answers an inconsistent
	// Property, or when Properties.Validate finds the schema inconsistent.
	ErrInvalidProperty = faults.Type("invalid property")

	// ErrInvalidConfig is returned by New when the Config is invalid.
	ErrInvalidConfig = faults.Type("invalid config")
)

const (
	errReader = faults.Type("read failed")
	errWriter = faults.Type("write failed")
)

type detail string

// Error returns the element and position the detail names; never a payload.
func (d detail) Error() string { return string(d) }
