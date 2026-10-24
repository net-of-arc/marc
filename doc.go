//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

// Package marc is a pure codec between one ArcNet CORE Markdown node file
// (https://github.com/net-of-arc/arcnet-spec/blob/main/spec/CORE.md) and a bag
// of (subject, predicate, object) triples.
//
// A Node is the file seen as data: its subject ID, its Class and a set of
// Triples whose Object is a reference (Ref) or a literal (Text, Number, Bool,
// Date, DateTime). Decode and Unmarshal read a node file; Encode and Marshal
// write the canonical file back. Decoding then encoding a canonical file
// reproduces it byte for byte, and encoding then decoding a node reproduces
// its triples.
//
//	node, err := marc.Unmarshal(src)
//	if err != nil {
//		return err
//	}
//	for t := range node.Triples() {
//		switch v := t.Object.(type) {
//		case marc.Ref:
//			fmt.Println(t.Predicate, "→", v.Target())
//		default:
//			fmt.Println(t.Predicate, v)
//		}
//	}
//
// Where each predicate renders (front-matter, prose, edge block or "## Label"
// block) is decided by its Property. The application supplies properties
// through a Schema in Config, such as the reference Properties map; when the
// schema is silent the encoder uses the property observed while decoding,
// then meta for literals and edge for references:
//
//	codec, err := marc.New(marc.Config{Schema: marc.Properties{
//		marc.MustPredicate("text"):        {Role: marc.RoleText, Default: true},
//		marc.MustPredicate("mentionedIn"): {Role: marc.RoleLink},
//	}})
//	b, err := codec.Marshal(node)
//
// The package has no side effects: it reads only the reader or bytes it is
// given, bounds input by Config.MaxSize, and reports failures as errors
// matchable with errors.Is against the exported Err constants.
package marc
