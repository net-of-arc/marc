//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

// Triple is one statement (Subject, Predicate, Object) about a node. It is a
// plain value; its validity is checked when it enters a Node through NewNode,
// not when the struct is built.
type Triple struct {
	Subject   ID
	Predicate Predicate
	Object    Object
}

// Attribution is one "@provenance" entry: Document contributed Object as a
// value of Predicate to the node. Attributions are carried beside the triples
// and are never triples themselves.
type Attribution struct {
	Document  ID
	Predicate Predicate
	Object    Object
}
