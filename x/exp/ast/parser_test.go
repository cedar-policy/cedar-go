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
		{
			"trivial conditions",
			`permit (principal, action, resource)
			when { true }
			unless { false };`,
			Permit().When(Boolean(true)).Unless(Boolean(false)),
		},
		{
			"not operator",
			`permit (principal, action, resource)
			when { !true };`,
			Permit().When(Not(Boolean(true))),
		},
		{
			"negate operator",
			`permit (principal, action, resource)
			when { -1 };`,
			Permit().When(Negate(Long(1))),
		},
		{
			"variable member",
			`permit (principal, action, resource)
			when { context.boolValue };`,
			Permit().When(Context().Access("boolValue")),
		},
		{
			"contains method call",
			`permit (principal, action, resource)
			when { context.strings.contains("foo") };`,
			Permit().When(Context().Access("strings").Contains(String("foo"))),
		},
		{
			"containsAll method call",
			`permit (principal, action, resource)
			when { context.strings.containsAll(["foo"]) };`,
			Permit().When(Context().Access("strings").ContainsAll(SetNodes(String("foo")))),
		},
		{
			"containsAny method call",
			`permit (principal, action, resource)
			when { context.strings.containsAny(["foo"]) };`,
			Permit().When(Context().Access("strings").ContainsAny(SetNodes(String("foo")))),
		},
		{
			"extension method call",
			`permit (principal, action, resource)
			when { context.sourceIP.isIpv4() };`,
			Permit().When(Context().Access("sourceIP").IsIpv4()),
		},
		{
			"multiplication",
			`permit (principal, action, resource)
			when { 42 * 2 };`,
			Permit().When(Long(42).Times(Long(2))),
		},
		{
			"addition",
			`permit (principal, action, resource)
			when { 42 + 2 };`,
			Permit().When(Long(42).Plus(Long(2))),
		},
		{
			"subtraction",
			`permit (principal, action, resource)
			when { 42 - 2 };`,
			Permit().When(Long(42).Minus(Long(2))),
		},
		{
			"less than",
			`permit (principal, action, resource)
			when { 2 < 42 };`,
			Permit().When(Long(2).LessThan(Long(42))),
		},
		{
			"less than or equal",
			`permit (principal, action, resource)
			when { 2 <= 42 };`,
			Permit().When(Long(2).LessThanOrEqual(Long(42))),
		},
		{
			"greater than",
			`permit (principal, action, resource)
			when { 2 > 42 };`,
			Permit().When(Long(2).GreaterThan(Long(42))),
		},
		{
			"greater than or equal",
			`permit (principal, action, resource)
			when { 2 >= 42 };`,
			Permit().When(Long(2).GreaterThanOrEqual(Long(42))),
		},
		{
			"equal",
			`permit (principal, action, resource)
			when { 2 == 42 };`,
			Permit().When(Long(2).Equals(Long(42))),
		},
		{
			"not equal",
			`permit (principal, action, resource)
			when { 2 != 42 };`,
			Permit().When(Long(2).NotEquals(Long(42))),
		},
		{
			"in",
			`permit (principal, action, resource)
			when { principal in Group::"folkHeroes" };`,
			Permit().When(Principal().In(Entity(folkHeroes))),
		},
		{
			"has ident",
			`permit (principal, action, resource)
			when { principal has firstName };`,
			Permit().When(Principal().Has("firstName")),
		},
		{
			"has string",
			`permit (principal, action, resource)
			when { principal has "firstName" };`,
			Permit().When(Principal().Has("firstName")),
		},
		// N.B. Most pattern parsing tests can be found in pattern_test.go
		{
			"like no wildcards",
			`permit (principal, action, resource)
			when { principal.firstName like "johnny" };`,
			Permit().When(Principal().Access("firstName").Like(testutil.Must(PatternFromCedar("johnny")))),
		},
		{
			"like escaped asterisk",
			`permit (principal, action, resource)
			when { principal.firstName like "joh\*nny" };`,
			Permit().When(Principal().Access("firstName").Like(testutil.Must(PatternFromCedar(`joh\*nny`)))),
		},
		{
			"like wildcard",
			`permit (principal, action, resource)
			when { principal.firstName like "*" };`,
			Permit().When(Principal().Access("firstName").Like(testutil.Must(PatternFromCedar("*")))),
		},
		{
			"is",
			`permit (principal, action, resource)
			when { principal is User };`,
			Permit().When(Principal().Is("User")),
		},
		{
			"is in",
			`permit (principal, action, resource)
			when { principal is User in Group::"folkHeroes" };`,
			Permit().When(Principal().IsIn("User", Entity(folkHeroes))),
		},
		{
			"is in",
			`permit (principal, action, resource)
			when { principal is User in Group::"folkHeroes" };`,
			Permit().When(Principal().IsIn("User", Entity(folkHeroes))),
		},
		{
			"and",
			`permit (principal, action, resource)
			when { true && false };`,
			Permit().When(True().And(False())),
		},
		{
			"or",
			`permit (principal, action, resource)
			when { true || false };`,
			Permit().When(True().Or(False())),
		},
		{
			"if then else",
			`permit (principal, action, resource)
			when { if true then true else false };`,
			Permit().When(If(True(), True(), False())),
		},
		{
			"and over or precedence",
			`permit (principal, action, resource)
			when { true && false || true && true };`,
			Permit().When(True().And(False()).Or(True().And(True()))),
		},
		{
			"rel over and precedence",
			`permit (principal, action, resource)
			when { 1 < 2 && true };`,
			Permit().When(Long(1).LessThan(Long(2)).And(True())),
		},
		{
			"add over rel precedence",
			`permit (principal, action, resource)
			when { 1 + 1 < 3 };`,
			Permit().When(Long(1).Plus(Long(1)).LessThan(Long(3))),
		},
		{
			"mult over add precedence",
			`permit (principal, action, resource)
			when { 2 * 3 + 4 == 10 };`,
			Permit().When(Long(2).Times(Long(3)).Plus(Long(4)).Equals(Long(10))),
		},
		{
			"unary over mult precedence",
			`permit (principal, action, resource)
			when { -2 * 3 == -6 };`,
			Permit().When(Negate(Long(2)).Times(Long(3)).Equals(Negate(Long(6)))),
		},
		{
			"member over unary precedence",
			`permit (principal, action, resource)
			when { -context.num };`,
			Permit().When(Negate(Context().Access("num"))),
		},
		{
			"member over unary precedence",
			`permit (principal, action, resource)
			when { -context.num };`,
			Permit().When(Negate(Context().Access("num"))),
		},
		{
			"parens over unary precedence",
			`permit (principal, action, resource)
			when { -(2 + 3) == -5 };`,
			Permit().When(Negate(Long(2).Plus(Long(3))).Equals(Negate(Long(5)))),
		},
	}

	for _, tt := range parseTests {
		t.Run(tt.Name, func(t *testing.T) {
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
