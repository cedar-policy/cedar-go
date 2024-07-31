package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

var johnny = types.EntityUID{
	Type: "User",
	ID:   "johnny",
}
var folkHeroes = types.EntityUID{
	Type: "Group",
	ID:   "folkHeroes",
}
var sow = types.EntityUID{
	Type: "Action",
	ID:   "sow",
}
var farming = types.EntityUID{
	Type: "ActionType",
	ID:   "farming",
}
var forestry = types.EntityUID{
	Type: "ActionType",
	ID:   "forestry",
}
var apple = types.EntityUID{
	Type: "Crop",
	ID:   "apple",
}
var malus = types.EntityUID{
	Type: "Genus",
	ID:   "malus",
}

func TestParse(t *testing.T) {
	t.Parallel()
	parseTests := []struct {
		Name           string
		Text           string
		ExpectedPolicy *Policy
	}{
		{
			"permit any scope",
			`permit (
			principal,
			action,
			resource
		);`,
			Permit(),
		},
		{
			"forbid any scope",
			`forbid (
			principal,
			action,
			resource
		);`,
			Forbid(),
		},
		{
			"one annotation",
			`@foo("bar")
		permit (
			principal,
			action,
			resource
		);`,
			Annotation("foo", "bar").Permit(),
		},
		{
			"two annotations",
			`@foo("bar")
		@baz("quux")
		permit (
			principal,
			action,
			resource
		);`,
			Annotation("foo", "bar").Annotation("baz", "quux").Permit(),
		},
		{
			"scope eq",
			`permit (
			principal == User::"johnny",
			action == Action::"sow",
			resource == Crop::"apple"
		);`,
			Permit().PrincipalEq(johnny).ActionEq(sow).ResourceEq(apple),
		},
		{
			"scope is",
			`permit (
			principal is User,
			action,
			resource is Crop
		);`,
			Permit().PrincipalIs("User").ResourceIs("Crop"),
		},
		{
			"scope is in",
			`permit (
			principal is User in Group::"folkHeroes",
			action,
			resource is Crop in Genus::"malus"
		);`,
			Permit().PrincipalIsIn("User", folkHeroes).ResourceIsIn("Crop", malus),
		},
		{
			"scope in",
			`permit (
			principal in Group::"folkHeroes",
			action in ActionType::"farming",
			resource in Genus::"malus"
		);`,
			Permit().PrincipalIn(folkHeroes).ActionIn(farming).ResourceIn(malus),
		},
		{
			"scope action in entities",
			`permit (
			principal,
			action in [ActionType::"farming", ActionType::"forestry"],
			resource
		);`,
			Permit().ActionInSet(farming, forestry),
		},
	}

	for _, tt := range parseTests {
		t.Run(tt.Text, func(t *testing.T) {
			t.Parallel()

			tokens, err := Tokenize([]byte(tt.Text))
			testutil.OK(t, err)

			parser := newParser(tokens)

			policy, err := policyFromCedar(&parser)
			testutil.OK(t, err)

			testutil.Equals(t, policy, tt.ExpectedPolicy)
		})
	}
}
