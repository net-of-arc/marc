//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"fmt"
	"slices"

	"github.com/net-of-arc/marc"
)

const coreExample = `---
"@id": Transport Layer Security
"@type": Entity
category: script
tags: [cryptography]
---
# Transport Layer Security

A cryptographic protocol that establishes an authenticated, confidential channel over an untrusted network.

- conformsTo:: [[RFC 8446]]
- replaces:: [[SSL Protocol]]

## Mentioned In
- mentionedIn:: [[rescorla-2026-tls13]]
`

func ExampleDecode() {
	node, err := marc.Unmarshal([]byte(coreExample))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s a %s\n", node.ID(), node.Class())
	for t := range node.Triples() {
		fmt.Printf("  %s %s\n", t.Predicate, t.Object)
	}

	// Output:
	// Transport Layer Security a Entity
	//   category script
	//   tags cryptography
	//   text A cryptographic protocol that establishes an authenticated, confidential channel over an untrusted network.
	//   conformsTo [[RFC 8446]]
	//   replaces [[SSL Protocol]]
	//   mentionedIn [[rescorla-2026-tls13]]
}

func ExampleSelect() {
	node, err := marc.Unmarshal([]byte(coreExample))
	if err != nil {
		panic(err)
	}

	for p := range node.Predicates() {
		for ref := range marc.Select[marc.Ref](node.Values(p)) {
			fmt.Printf("%s → %s\n", p, ref.Target())
		}
	}

	// Output:
	// conformsTo → RFC 8446
	// replaces → SSL Protocol
	// mentionedIn → rescorla-2026-tls13
}

func ExampleNode_Links() {
	node, err := marc.Unmarshal([]byte(
		"---\n\"@id\": X\n\"@type\": Entity\n---\n# X\n\nSee [[RFC 8446]] for details.\n"))
	if err != nil {
		panic(err)
	}

	for p, ref := range node.Links() {
		fmt.Printf("%s → %s\n", p, ref.Target())
	}

	// Output:
	// text → RFC 8446
}

func ExampleEncode() {
	id := marc.MustID("Transport Layer Security")
	triples := []marc.Triple{
		{Subject: id, Predicate: marc.MustPredicate("category"), Object: marc.Text("script")},
		{Subject: id, Predicate: marc.MustPredicate("tags"), Object: marc.Text("cryptography")},
		{Subject: id, Predicate: marc.MustPredicate("text"), Object: marc.Text("A cryptographic protocol.")},
		{Subject: id, Predicate: marc.MustPredicate("replaces"), Object: marc.RefTo(marc.MustID("SSL Protocol"))},
		{Subject: id, Predicate: marc.MustPredicate("mentionedIn"), Object: marc.RefTo(marc.MustID("rescorla-2026-tls13"))},
	}
	node, err := marc.NewNode(id, marc.MustClass("Entity"), slices.Values(triples))
	if err != nil {
		panic(err)
	}

	codec, err := marc.New(marc.Config{
		Schema: marc.Properties{
			marc.MustPredicate("category"):    {Role: marc.RoleMeta},
			marc.MustPredicate("tags"):        {Role: marc.RoleMeta, Many: true},
			marc.MustPredicate("text"):        {Role: marc.RoleText, Default: true},
			marc.MustPredicate("replaces"):    {Role: marc.RoleEdge},
			marc.MustPredicate("mentionedIn"): {Role: marc.RoleLink},
		},
	})
	if err != nil {
		panic(err)
	}

	b, err := codec.Marshal(node)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(b))

	// Output:
	// ---
	// "@id": Transport Layer Security
	// "@type": Entity
	// category: script
	// tags: [cryptography]
	// ---
	// # Transport Layer Security
	//
	// A cryptographic protocol.
	//
	// - replaces:: [[SSL Protocol]]
	//
	// ## Mentioned In
	// - mentionedIn:: [[rescorla-2026-tls13]]
}
