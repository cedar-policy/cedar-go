package parser_test

import (
	"bytes"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
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

func TestParsePolicy(t *testing.T) {
	t.Parallel()
	parseTests := []struct {
		Name           string
		Text           string
		ExpectedPolicy *ast.Policy
	}{
		{
			"permit any scope",
			`permit ( principal, action, resource );`,
			ast.Permit(),
		},
		{
			"forbid any scope",
			`forbid ( principal, action, resource );`,
			ast.Forbid(),
		},
		{
			"one annotation",
			`@foo("bar")
permit ( principal, action, resource );`,
			ast.Annotation("foo", "bar").Permit(),
		},
		{
			"two annotations",
			`@foo("bar")
@baz("quux")
permit ( principal, action, resource );`,
			ast.Annotation("foo", "bar").Annotation("baz", "quux").Permit(),
		},
		{
			"scope eq",
			`permit (
    principal == User::"johnny",
    action == Action::"sow",
    resource == Crop::"apple"
);`,
			ast.Permit().PrincipalEq(johnny).ActionEq(sow).ResourceEq(apple),
		},
		{
			"scope is",
			`permit (
    principal is User,
    action,
    resource is Crop
);`,
			ast.Permit().PrincipalIs("User").ResourceIs("Crop"),
		},
		{
			"scope is in",
			`permit (
    principal is User in Group::"folkHeroes",
    action,
    resource is Crop in Genus::"malus"
);`,
			ast.Permit().PrincipalIsIn("User", folkHeroes).ResourceIsIn("Crop", malus),
		},
		{
			"scope in",
			`permit (
    principal in Group::"folkHeroes",
    action in ActionType::"farming",
    resource in Genus::"malus"
);`,
			ast.Permit().PrincipalIn(folkHeroes).ActionIn(farming).ResourceIn(malus),
		},
		{
			"scope action in entities",
			`permit (
    principal,
    action in [ActionType::"farming", ActionType::"forestry"],
    resource
);`,
			ast.Permit().ActionInSet(farming, forestry),
		},
		{
			"trivial conditions",
			`permit ( principal, action, resource )
when { true }
unless { false };`,
			ast.Permit().When(ast.Boolean(true)).Unless(ast.Boolean(false)),
		},
		{
			"not operator",
			`permit ( principal, action, resource )
when { !true };`,
			ast.Permit().When(ast.Not(ast.Boolean(true))),
		},
		{
			"multiple not operators",
			`permit ( principal, action, resource )
when { !!true };`,
			ast.Permit().When(ast.Not(ast.Not(ast.Boolean(true)))),
		},
		{
			"negate operator",
			`permit ( principal, action, resource )
when { -1 };`,
			ast.Permit().When(ast.Long(-1)),
		},
		{
			"negate operator context",
			`permit ( principal, action, resource )
when { -context };`,
			ast.Permit().When(ast.Negate(ast.Context())),
		},
		{
			"mutliple negate operators",
			`permit ( principal, action, resource )
when { !--1 };`,
			ast.Permit().When(ast.Not(ast.Negate(ast.Long(-1)))),
		},
		{
			"variable member",
			`permit ( principal, action, resource )
when { context.boolValue };`,
			ast.Permit().When(ast.Context().Access("boolValue")),
		},
		{
			"variable member via []",
			`permit ( principal, action, resource )
when { context["2legit2quit"] };`,
			ast.Permit().When(ast.Context().Access("2legit2quit")),
		},
		{
			"contains method call",
			`permit ( principal, action, resource )
when { context.strings.contains("foo") };`,
			ast.Permit().When(ast.Context().Access("strings").Contains(ast.String("foo"))),
		},
		{
			"containsAll method call",
			`permit ( principal, action, resource )
when { context.strings.containsAll(["foo"]) };`,
			ast.Permit().When(ast.Context().Access("strings").ContainsAll(ast.SetNodes(ast.String("foo")))),
		},
		{
			"containsAny method call",
			`permit ( principal, action, resource )
when { context.strings.containsAny(["foo"]) };`,
			ast.Permit().When(ast.Context().Access("strings").ContainsAny(ast.SetNodes(ast.String("foo")))),
		},
		{
			"extension method call",
			`permit ( principal, action, resource )
when { context.sourceIP.isIpv4() };`,
			ast.Permit().When(ast.Context().Access("sourceIP").IsIpv4()),
		},
		{
			"multiplication",
			`permit ( principal, action, resource )
when { 42 * 2 };`,
			ast.Permit().When(ast.Long(42).Times(ast.Long(2))),
		},
		{
			"multiple multiplication",
			`permit ( principal, action, resource )
when { 42 * 2 * 1 };`,
			ast.Permit().When(ast.Long(42).Times(ast.Long(2)).Times(ast.Long(1))),
		},
		{
			"addition",
			`permit ( principal, action, resource )
when { 42 + 2 };`,
			ast.Permit().When(ast.Long(42).Plus(ast.Long(2))),
		},
		{
			"multiple addition",
			`permit ( principal, action, resource )
when { 42 + 2 + 1 };`,
			ast.Permit().When(ast.Long(42).Plus(ast.Long(2)).Plus(ast.Long(1))),
		},
		{
			"subtraction",
			`permit ( principal, action, resource )
when { 42 - 2 };`,
			ast.Permit().When(ast.Long(42).Minus(ast.Long(2))),
		},
		{
			"multiple subtraction",
			`permit ( principal, action, resource )
when { 42 - 2 - 1 };`,
			ast.Permit().When(ast.Long(42).Minus(ast.Long(2)).Minus(ast.Long(1))),
		},
		{
			"mixed addition and subtraction",
			`permit ( principal, action, resource )
when { 42 - 2 + 1 };`,
			ast.Permit().When(ast.Long(42).Minus(ast.Long(2)).Plus(ast.Long(1))),
		},
		{
			"less than",
			`permit ( principal, action, resource )
when { 2 < 42 };`,
			ast.Permit().When(ast.Long(2).LessThan(ast.Long(42))),
		},
		{
			"less than or equal",
			`permit ( principal, action, resource )
when { 2 <= 42 };`,
			ast.Permit().When(ast.Long(2).LessThanOrEqual(ast.Long(42))),
		},
		{
			"greater than",
			`permit ( principal, action, resource )
when { 2 > 42 };`,
			ast.Permit().When(ast.Long(2).GreaterThan(ast.Long(42))),
		},
		{
			"greater than or equal",
			`permit ( principal, action, resource )
when { 2 >= 42 };`,
			ast.Permit().When(ast.Long(2).GreaterThanOrEqual(ast.Long(42))),
		},
		{
			"equal",
			`permit ( principal, action, resource )
when { 2 == 42 };`,
			ast.Permit().When(ast.Long(2).Equals(ast.Long(42))),
		},
		{
			"not equal",
			`permit ( principal, action, resource )
when { 2 != 42 };`,
			ast.Permit().When(ast.Long(2).NotEquals(ast.Long(42))),
		},
		{
			"in",
			`permit ( principal, action, resource )
when { principal in Group::"folkHeroes" };`,
			ast.Permit().When(ast.Principal().In(ast.EntityUID(folkHeroes))),
		},
		{
			"has ident",
			`permit ( principal, action, resource )
when { principal has firstName };`,
			ast.Permit().When(ast.Principal().Has("firstName")),
		},
		{
			"has string",
			`permit ( principal, action, resource )
when { principal has "1stName" };`,
			ast.Permit().When(ast.Principal().Has("1stName")),
		},
		// N.B. Most pattern parsing tests can be found in pattern_test.go
		{
			"like no wildcards",
			`permit ( principal, action, resource )
when { principal.firstName like "johnny" };`,
			ast.Permit().When(ast.Principal().Access("firstName").Like(testutil.Must(types.ParsePattern("johnny")))),
		},
		{
			"like escaped asterisk",
			`permit ( principal, action, resource )
when { principal.firstName like "joh\*nny" };`,
			ast.Permit().When(ast.Principal().Access("firstName").Like(testutil.Must(types.ParsePattern(`joh\*nny`)))),
		},
		{
			"like wildcard",
			`permit ( principal, action, resource )
when { principal.firstName like "*" };`,
			ast.Permit().When(ast.Principal().Access("firstName").Like(testutil.Must(types.ParsePattern("*")))),
		},
		{
			"is",
			`permit ( principal, action, resource )
when { principal is User };`,
			ast.Permit().When(ast.Principal().Is("User")),
		},
		{
			"is in",
			`permit ( principal, action, resource )
when { principal is User in Group::"folkHeroes" };`,
			ast.Permit().When(ast.Principal().IsIn("User", ast.EntityUID(folkHeroes))),
		},
		{
			"and",
			`permit ( principal, action, resource )
when { true && false };`,
			ast.Permit().When(ast.True().And(ast.False())),
		},
		{
			"multiple and",
			`permit ( principal, action, resource )
when { true && false && true };`,
			ast.Permit().When(ast.True().And(ast.False()).And(ast.True())),
		},
		{
			"or",
			`permit ( principal, action, resource )
when { true || false };`,
			ast.Permit().When(ast.True().Or(ast.False())),
		},
		{
			"multiple or",
			`permit ( principal, action, resource )
when { true || false || true };`,
			ast.Permit().When(ast.True().Or(ast.False()).Or(ast.True())),
		},
		{
			"if then else",
			`permit ( principal, action, resource )
when { if true then true else false };`,
			ast.Permit().When(ast.If(ast.True(), ast.True(), ast.False())),
		},
		{
			"ip extension function",
			`permit ( principal, action, resource )
when { ip("1.2.3.4") == ip("2.3.4.5") };`,
			ast.Permit().When(
				ast.ExtensionCall("ip", ast.String("1.2.3.4")).Equals(
					ast.ExtensionCall("ip", ast.String("2.3.4.5")),
				),
			),
		},
		{
			"decimal extension function",
			`permit ( principal, action, resource )
when { decimal("12.34") == decimal("23.45") };`,
			ast.Permit().When(
				ast.ExtensionCall("decimal", ast.String("12.34")).Equals(ast.ExtensionCall("decimal", ast.String("23.45"))),
			),
		},
		{
			"and over or precedence",
			`permit ( principal, action, resource )
when { true && false || true && true };`,
			ast.Permit().When(ast.True().And(ast.False()).Or(ast.True().And(ast.True()))),
		},
		{
			"rel over and precedence",
			`permit ( principal, action, resource )
when { 1 < 2 && true };`,
			ast.Permit().When(ast.Long(1).LessThan(ast.Long(2)).And(ast.True())),
		},
		{
			"add over rel precedence",
			`permit ( principal, action, resource )
when { 1 + 1 < 3 };`,
			ast.Permit().When(ast.Long(1).Plus(ast.Long(1)).LessThan(ast.Long(3))),
		},
		{
			"mult over add precedence (rhs add)",
			`permit ( principal, action, resource )
when { 2 * 3 + 4 == 10 };`,
			ast.Permit().When(ast.Long(2).Times(ast.Long(3)).Plus(ast.Long(4)).Equals(ast.Long(10))),
		},
		{
			"mult over add precedence (lhs add)",
			`permit ( principal, action, resource )
when { 2 + 3 * 4 == 14 };`,
			ast.Permit().When(ast.Long(2).Plus(ast.Long(3).Times(ast.Long(4))).Equals(ast.Long(14))),
		},
		{
			"unary over mult precedence",
			`permit ( principal, action, resource )
when { -2 * 3 == -6 };`,
			ast.Permit().When(ast.Long(-2).Times(ast.Long(3)).Equals(ast.Long(-6))),
		},
		{
			"member over unary precedence",
			`permit ( principal, action, resource )
when { -context.num };`,
			ast.Permit().When(ast.Negate(ast.Context().Access("num"))),
		},
		{
			"parens over unary precedence",
			`permit ( principal, action, resource )
when { -(2 + 3) == -5 };`,
			ast.Permit().When(ast.Negate(ast.Long(2).Plus(ast.Long(3))).Equals(ast.Long(-5))),
		},
		{
			"multiple parenthesized operations",
			`permit ( principal, action, resource )
when { (2 + 3 + 4) * 5 == 18 };`,
			ast.Permit().When(ast.Long(2).Plus(ast.Long(3)).Plus(ast.Long(4)).Times(ast.Long(5)).Equals(ast.Long(18))),
		},
		{
			"parenthesized if",
			`permit ( principal, action, resource )
when { (if true then 2 else 3 * 4) == 2 };`,
			ast.Permit().When(ast.If(ast.True(), ast.Long(2), ast.Long(3).Times(ast.Long(4))).Equals(ast.Long(2))),
		},
		{
			"parenthesized if with trailing mult",
			`permit ( principal, action, resource )
when { (if true then 2 else 3) * 4 == 8 };`,
			ast.Permit().When(ast.If(ast.True(), ast.Long(2), ast.Long(3)).Times(ast.Long(4)).Equals(ast.Long(8))),
		},
	}

	for _, tt := range parseTests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var policy parser.Policy
			testutil.OK(t, policy.UnmarshalCedar([]byte(tt.Text)))
			policy.Position = ast.Position{}
			testutil.Equals(t, policy, parser.Policy{*tt.ExpectedPolicy})

			var buf bytes.Buffer
			policy.MarshalCedar(&buf)
			testutil.Equals(t, buf.String(), tt.Text)
		})
	}
}

func TestParsePolicySet(t *testing.T) {
	t.Parallel()
	t.Run("single policy", func(t *testing.T) {
		policyStr := []byte(`permit (
			principal,
			action,
			resource
		);`)

		var policies parser.PolicySet
		testutil.OK(t, policies.UnmarshalCedar(policyStr))

		expectedPolicy := ast.Permit()
		expectedPolicy.Position = ast.Position{Offset: 0, Line: 1, Column: 1}
		testutil.Equals(t, &policies[0].Policy, expectedPolicy)
	})
	t.Run("two policies", func(t *testing.T) {
		policyStr := []byte(`permit (
			principal,
			action,
			resource
		);
		forbid (
			principal,
			action,
			resource
		);`)
		var policies parser.PolicySet
		testutil.OK(t, policies.UnmarshalCedar(policyStr))

		expectedPolicy0 := ast.Permit()
		expectedPolicy0.Position = ast.Position{Offset: 0, Line: 1, Column: 1}
		testutil.Equals(t, &policies[0].Policy, expectedPolicy0)

		expectedPolicy1 := ast.Forbid()
		expectedPolicy1.Position = ast.Position{Offset: 53, Line: 6, Column: 3}
		testutil.Equals(t, &policies[1].Policy, expectedPolicy1)
	})
}
