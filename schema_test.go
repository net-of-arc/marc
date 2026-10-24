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

func TestPropertiesValidate(t *testing.T) {
	for name, schema := range map[string]Properties{
		"two defaults": {
			MustPredicate("text"):    {Role: RoleText, Default: true},
			MustPredicate("summary"): {Role: RoleText, Default: true},
		},
		"default edge":      {MustPredicate("text"): {Role: RoleEdge, Default: true}},
		"default many":      {MustPredicate("text"): {Role: RoleText, Default: true, Many: true}},
		"role out of range": {MustPredicate("text"): {Role: Role(9)}},
		"duplicate label": {
			MustPredicate("mentionedIn"): {Role: RoleLink},
			MustPredicate("citedBy"):     {Role: RoleLink, Label: "Mentioned In"},
		},
		"label with break": {MustPredicate("notes"): {Role: RoleText, Label: "A\nB"}},
		"label with star":  {MustPredicate("notes"): {Role: RoleText, Label: "**A"}},
	} {
		t.Run(name, func(t *testing.T) {
			it.Then(t).Should(it.True(errors.Is(schema.Validate(), ErrInvalidProperty)))
		})
	}
}

func TestPropertiesValidateAccepts(t *testing.T) {
	schema := Properties{
		MustPredicate("text"):        {Role: RoleText, Default: true},
		MustPredicate("tags"):        {Role: RoleMeta, Many: true},
		MustPredicate("mentionedIn"): {Role: RoleLink},
		MustPredicate("citedBy"):     {Role: RoleLink, Label: "Sources"},
	}
	it.Then(t).Should(it.Nil(schema.Validate()))
}

func TestPropertiesLookup(t *testing.T) {
	schema := Properties{
		MustPredicate("text"):        {Role: RoleText, Default: true},
		MustPredicate("citedBy"):     {Role: RoleLink, Label: "Sources"},
		MustPredicate("keyFindings"): {Role: RoleText},
	}
	entity := MustClass("Entity")

	prop, ok := schema.Property(entity, MustPredicate("citedBy"))
	_, missing := schema.Property(entity, MustPredicate("absent"))
	byDefault, okDefault := schema.Predicate(entity, "")
	byLabel, okLabel := schema.Predicate(entity, "Sources")
	byDerived, okDerived := schema.Predicate(entity, "Key Findings")
	_, okUnknown := schema.Predicate(entity, "Unknown")

	it.Then(t).Should(
		it.True(ok), it.Equal(prop.Label, "Sources"), it.True(!missing),
		it.True(okDefault), it.Equal(byDefault, MustPredicate("text")),
		it.True(okLabel), it.Equal(byLabel, MustPredicate("citedBy")),
		it.True(okDerived), it.Equal(byDerived, MustPredicate("keyFindings")),
		it.True(!okUnknown),
	)
}

func TestRoleString(t *testing.T) {
	it.Then(t).Should(
		it.Equal(RoleMeta.String(), "meta"),
		it.Equal(RoleText.String(), "text"),
		it.Equal(RoleEdge.String(), "edge"),
		it.Equal(RoleLink.String(), "link"),
		it.Equal(Role(9).String(), "invalid"),
	)
}
