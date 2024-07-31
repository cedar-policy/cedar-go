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

var parseTests = []struct {
	Text           string
	ExpectedPolicy *Policy
}{
	{
		`permit (
			principal,
			action,
			resource
		);`,
		Permit(),
	},
	{
		`forbid (
			principal,
			action,
			resource
		);`,
		Forbid(),
	},
	{
		`@foo("bar")
		permit (
			principal,
			action,
			resource
		);`,
		Annotation("foo", "bar").Permit(),
	},
	{
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
		`permit (
			principal == User::"johnny",
			action == Action::"sow",
			resource == Crop::"apple"
		);`,
		Permit().PrincipalEq(johnny).ActionEq(sow).ResourceEq(apple),
	},
	{
		`permit (
			principal is User,
			action,
			resource is Crop
		);`,
		Permit().PrincipalIs("User").ResourceIs("Crop"),
	},
	{
		`permit (
			principal is User in Group::"folkHeroes",
			action,
			resource is Crop in Genus::"malus"
		);`,
		Permit().PrincipalIsIn("User", folkHeroes).ResourceIsIn("Crop", malus),
	},
	{
		`permit (
			principal in Group::"folkHeroes",
			action in ActionType::"farming",
			resource in Genus::"malus"
		);`,
		Permit().PrincipalIn(folkHeroes).ActionIn(farming).ResourceIn(malus),
	},
	{
		`permit (
			principal,
			action in [ActionType::"farming", ActionType::"forestry"],
			resource
		);`,
		Permit().ActionIn(farming, forestry),
	},
}

func TestParse(t *testing.T) {
	t.Parallel()
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
