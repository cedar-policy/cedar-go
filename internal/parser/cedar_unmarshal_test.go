package parser_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
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
			"reserved keyword annotation key",
			`@is("bar")
permit ( principal, action, resource );`,
			ast.Annotation("is", "bar").Permit(),
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
			ast.Permit().When(ast.Context().Access("strings").ContainsAll(ast.Set(ast.String("foo")))),
		},
		{
			"containsAny method call",
			`permit ( principal, action, resource )
when { context.strings.containsAny(["foo"]) };`,
			ast.Permit().When(ast.Context().Access("strings").ContainsAny(ast.Set(ast.String("foo")))),
		},
		{
			"isEmpty method call",
			`permit ( principal, action, resource )
when { context.strings.isEmpty() };`,
			ast.Permit().When(ast.Context().Access("strings").IsEmpty()),
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
			ast.Permit().When(ast.Long(42).Multiply(ast.Long(2))),
		},
		{
			"principal has tag",
			`permit ( principal, action, resource )
when { principal.hasTag("blue") };`,
			ast.Permit().When(ast.Principal().HasTag(ast.String("blue"))),
		},
		{
			"resource has tag",
			`permit ( principal, action, resource )
when { resource.hasTag("blue") };`,
			ast.Permit().When(ast.Resource().HasTag(ast.String("blue"))),
		},
		{
			"principal tag equals value",
			`permit ( principal, action, resource )
when { principal.hasTag("blue") && principal.getTag("blue") == "green" };`,
			ast.Permit().When(ast.Principal().HasTag(ast.String("blue")).And(ast.Principal().GetTag(ast.String("blue")).Equal(ast.String("green")))),
		},
		{
			"principal tag has attribute",
			`permit ( principal, action, resource )
when { principal.hasTag("blue") && principal.getTag("blue") has attr };`,
			ast.Permit().When(ast.Principal().HasTag(ast.String("blue")).And(ast.Principal().GetTag(ast.String("blue")).Has("attr"))),
		},
		{
			"principal has tag from context",
			`permit ( principal, action, resource )
when { principal.hasTag(context.request_ip) };`,
			ast.Permit().When(ast.Principal().HasTag(ast.Context().Access("request_ip"))),
		},
		{
			"multiple multiplication",
			`permit ( principal, action, resource )
when { 42 * 2 * 1 };`,
			ast.Permit().When(ast.Long(42).Multiply(ast.Long(2)).Multiply(ast.Long(1))),
		},
		{
			"addition",
			`permit ( principal, action, resource )
when { 42 + 2 };`,
			ast.Permit().When(ast.Long(42).Add(ast.Long(2))),
		},
		{
			"multiple addition",
			`permit ( principal, action, resource )
when { 42 + 2 + 1 };`,
			ast.Permit().When(ast.Long(42).Add(ast.Long(2)).Add(ast.Long(1))),
		},
		{
			"subtraction",
			`permit ( principal, action, resource )
when { 42 - 2 };`,
			ast.Permit().When(ast.Long(42).Subtract(ast.Long(2))),
		},
		{
			"multiple subtraction",
			`permit ( principal, action, resource )
when { 42 - 2 - 1 };`,
			ast.Permit().When(ast.Long(42).Subtract(ast.Long(2)).Subtract(ast.Long(1))),
		},
		{
			"mixed addition and subtraction",
			`permit ( principal, action, resource )
when { 42 - 2 + 1 };`,
			ast.Permit().When(ast.Long(42).Subtract(ast.Long(2)).Add(ast.Long(1))),
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
			ast.Permit().When(ast.Long(2).Equal(ast.Long(42))),
		},
		{
			"not equal",
			`permit ( principal, action, resource )
when { 2 != 42 };`,
			ast.Permit().When(ast.Long(2).NotEqual(ast.Long(42))),
		},
		{
			"in",
			`permit ( principal, action, resource )
when { principal in Group::"folkHeroes" };`,
			ast.Permit().When(ast.Principal().In(ast.Value(folkHeroes))),
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
		// N.B. Most pattern parsing tests can be found in types/pattern_test.go
		{
			"like no wildcards",
			`permit ( principal, action, resource )
when { principal.firstName like "johnny" };`,
			ast.Permit().When(ast.Principal().Access("firstName").Like(types.NewPattern(types.String("johnny")))),
		},
		{
			"like escaped asterisk",
			`permit ( principal, action, resource )
when { principal.firstName like "joh\*nny" };`,
			ast.Permit().When(ast.Principal().Access("firstName").Like(types.NewPattern(types.String("joh*nny")))),
		},
		{
			"like wildcard",
			`permit ( principal, action, resource )
when { principal.firstName like "*" };`,
			ast.Permit().When(ast.Principal().Access("firstName").Like(types.NewPattern(types.Wildcard{}))),
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
			ast.Permit().When(ast.Principal().IsIn("User", ast.Value(folkHeroes))),
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
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.True(), ast.False())),
		},
		{
			"ip extension function",
			`permit ( principal, action, resource )
when { ip("1.2.3.4") == ip("2.3.4.5") };`,
			ast.Permit().When(
				ast.ExtensionCall("ip", ast.String("1.2.3.4")).Equal(
					ast.ExtensionCall("ip", ast.String("2.3.4.5")),
				),
			),
		},
		{
			"decimal extension function",
			`permit ( principal, action, resource )
when { decimal("12.34") == decimal("23.45") };`,
			ast.Permit().When(
				ast.ExtensionCall("decimal", ast.String("12.34")).Equal(ast.ExtensionCall("decimal", ast.String("23.45"))),
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
			ast.Permit().When(ast.Long(1).Add(ast.Long(1)).LessThan(ast.Long(3))),
		},
		{
			"mult over add precedence (rhs add)",
			`permit ( principal, action, resource )
when { 2 * 3 + 4 == 10 };`,
			ast.Permit().When(ast.Long(2).Multiply(ast.Long(3)).Add(ast.Long(4)).Equal(ast.Long(10))),
		},
		{
			"mult over add precedence (lhs add)",
			`permit ( principal, action, resource )
when { 2 + 3 * 4 == 14 };`,
			ast.Permit().When(ast.Long(2).Add(ast.Long(3).Multiply(ast.Long(4))).Equal(ast.Long(14))),
		},
		{
			"unary over mult precedence",
			`permit ( principal, action, resource )
when { -2 * 3 == -6 };`,
			ast.Permit().When(ast.Long(-2).Multiply(ast.Long(3)).Equal(ast.Long(-6))),
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
			ast.Permit().When(ast.Negate(ast.Long(2).Add(ast.Long(3))).Equal(ast.Long(-5))),
		},
		{
			"multiple parenthesized operations",
			`permit ( principal, action, resource )
when { (2 + 3 + 4) * 5 == 18 };`,
			ast.Permit().When(ast.Long(2).Add(ast.Long(3)).Add(ast.Long(4)).Multiply(ast.Long(5)).Equal(ast.Long(18))),
		},
		{
			"parenthesized if",
			`permit ( principal, action, resource )
when { (if true then 2 else 3 * 4) == 2 };`,
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.Long(2), ast.Long(3).Multiply(ast.Long(4))).Equal(ast.Long(2))),
		},
		{
			"parenthesized if with trailing mult",
			`permit ( principal, action, resource )
when { (if true then 2 else 3) * 4 == 8 };`,
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.Long(2), ast.Long(3)).Multiply(ast.Long(4)).Equal(ast.Long(8))),
		},
	}

	for _, tt := range parseTests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var policy parser.Policy
			testutil.OK(t, policy.UnmarshalCedar([]byte(tt.Text)))
			policy.Position = ast.Position{}
			testutil.Equals(t, &policy, (*parser.Policy)(tt.ExpectedPolicy))

			var buf bytes.Buffer
			policy.MarshalCedar(&buf)
			testutil.Equals(t, buf.String(), tt.Text)
		})
	}
}

func TestParsePolicySetErrors(t *testing.T) {
	t.Parallel()
	parseTests := []struct {
		Name          string
		Text          string
		ExpectedError string
	}{
		{
			"not-extension-function",
			"permit ( principal, action, resource ) when { not_an_extension_fn() };",
			"parse error at <input>:1:67 \")\": `not_an_extension_fn` is not a function",
		},
		{
			"extension-function-is-method",
			"permit ( principal, action, resource ) when { isIpv4() };",
			"parse error at <input>:1:54 \")\": `isIpv4` is a method, not a function",
		},
		{
			"not-extension-method",
			"permit ( principal, action, resource ) when { context.not_an_extension_method() };",
			"parse error at <input>:1:81 \"}\": `not_an_extension_method` is not a method",
		},
		{
			"extension-method-is-function",
			"permit ( principal, action, resource ) when { context.ip() };",
			"parse error at <input>:1:60 \"}\": `ip` is a function, not a method",
		},
	}

	for _, tt := range parseTests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			var policy parser.Policy
			err := policy.UnmarshalCedar([]byte(tt.Text))
			testutil.Error(t, err)
			testutil.Equals(t, err.Error(), tt.ExpectedError)
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

		var policies parser.PolicySlice
		testutil.OK(t, policies.UnmarshalCedar(policyStr))

		expectedPolicy := ast.Permit()
		expectedPolicy.Position = ast.Position{Offset: 0, Line: 1, Column: 1}
		testutil.Equals(t, policies[0], (*parser.Policy)(expectedPolicy))
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
		var policies parser.PolicySlice
		testutil.OK(t, policies.UnmarshalCedar(policyStr))

		expectedPolicy0 := ast.Permit()
		expectedPolicy0.Position = ast.Position{Offset: 0, Line: 1, Column: 1}
		testutil.Equals(t, policies[0], (*parser.Policy)(expectedPolicy0))

		expectedPolicy1 := ast.Forbid()
		expectedPolicy1.Position = ast.Position{Offset: 53, Line: 6, Column: 3}
		testutil.Equals(t, policies[1], (*parser.Policy)(expectedPolicy1))
	})
}

func TestParseApproximateErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		in              string
		outErrSubstring string
	}{
		{"unexpectedEffect", "!", "unexpected effect"},
		{"nul", "\x00", "invalid character"},
		{"notTerminated", `"`, "literal not terminated"},
		{"principalBadIsIn", `permit (principal is T in error);`, "got ) want ::"},
		{"principalBadIn", `permit (principal in error);`, "got ) want ::"},
		{"resourceBadEq", `permit (principal, action, resource == error);`, "got ) want ::"},
		{"resourceBadIsIn1", `permit (principal, action, resource is "error");`, "expected ident"},
		{"resourceBadIsIn1", `permit (principal, action, resource is T in error);`, "got ) want ::"},
		{"resourceBadIn", `permit (principal, action, resource in error);`, "got ) want ::"},
		{"unlessCondition", `permit (principal, action, resource) unless {`, "invalid primary"},
		{"or", `permit (principal, action, resource) when { true ||`, "invalid primary"},
		{"and", `permit (principal, action, resource) when { true &&`, "invalid primary"},
		{"isPath", `permit (principal, action, resource) when { context is`, "expected ident"},
		{"isIn", `permit (principal, action, resource) when { context is T in`, "invalid primary"},
		{"mult", `permit (principal, action, resource) when { 42 *`, "invalid primary"},
		{"parens", `permit (principal, action, resource) when { (42}`, "got } want )"},
		{"func", `permit (principal, action, resource) when { ip(}`, "invalid primary"},
		{"args", `permit (principal, action, resource) when { ip(42 42)`, "got 42 want ,"},
		{"dupeKey", `permit (principal, action, resource) when { {k:42,k:43}`, "duplicate key"},
		{"reservedKeywordAsRecordKey", `permit (principal, action, resource) when { {false:43} }`, "expected ident or string"},
		{"reservedKeywordAsHas", `permit (principal, action, resource) when { {} has false }`, "expected ident or string"},
		{"reservedKeywordAsEntityType", `permit (principal == false::"42", action, resource)`, "expected ident"},
		{"reservedKeywordAsAttributeAccess", `permit (principal, action, resource) when { context.false }`, "expected ident"},
		{"invalidPrimary", `permit (principal, action, resource) when { foobar }`, "invalid primary"},
		{"unexpectedTokenInEntityOrExtFun", `permit (principal, action, resource) when { A::B 42 }`, "unexpected token"},
		{"unexpectedZeroArgMethodArity", `permit (principal, action, resource) when { context.set.isEmpty("foo") }`, "isEmpty expects no arguments"},
		{"unexpectedOneArgMethodArity", `permit (principal, action, resource) when { context.set.contains() }`, "contains expects one argument"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var pol parser.Policy
			err := pol.UnmarshalCedar([]byte(tt.in))
			testutil.Error(t, err)
			testutil.FatalIf(t, !strings.Contains(err.Error(), tt.outErrSubstring), "got %v want %v", err.Error(), tt.outErrSubstring)
		})
	}
}

func TestPolicySliceErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		in              string
		outErrSubstring string
	}{
		{"notTerminated", `"`, "literal not terminated"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var pol parser.PolicySlice
			err := pol.UnmarshalCedar([]byte(tt.in))
			testutil.FatalIf(t, !strings.Contains(err.Error(), tt.outErrSubstring), "got %v want %v", err.Error(), tt.outErrSubstring)
		})
	}
}

func TestReservedNamesInEntityPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		out  *ast.Policy
		err  func(testutil.TB, error)
	}{
		{"action", `permit (principal, action, resource) when { action::"test" };`,
			ast.Permit().When(ast.EntityUID("action", "test")), testutil.OK}, // Rust: OK
		{"context", `permit (principal, action, resource) when { context::"test" };`,
			ast.Permit().When(ast.EntityUID("context", "test")), testutil.OK}, // Rust: OK
		{"else", `permit (principal, action, resource) when { else::"test" };`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"false", `permit (principal, action, resource) when { false::"test" };`,
			nil, testutil.Error}, // Rust: unexpected token
		{"forbid", `permit (principal, action, resource) when { forbid::"test" };`,
			ast.Permit().When(ast.EntityUID("forbid", "test")), testutil.OK}, // Rust: OK
		{"has", `permit (principal, action, resource) when { has::"test" };`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"if", `permit (principal, action, resource) when { if::"test" };`,
			nil, testutil.Error}, // Rust: unexpected token
		{"in", `permit (principal, action, resource) when { in::"test" };`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"is", `permit (principal, action, resource) when { is::"test" };`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"like", `permit (principal, action, resource) when { like::"test" };`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"permit", `permit (principal, action, resource) when { permit::"test" };`,
			ast.Permit().When(ast.EntityUID("permit", "test")), testutil.OK}, // Rust: OK
		{"principal", `permit (principal, action, resource) when { principal::"test" };`,
			ast.Permit().When(ast.EntityUID("principal", "test")), testutil.OK}, // Rust: OK
		{"resource", `permit (principal, action, resource) when { resource::"test" };`,
			ast.Permit().When(ast.EntityUID("resource", "test")), testutil.OK}, // Rust: OK
		{"then", `permit (principal, action, resource) when { then::"test" };`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"true", `permit (principal, action, resource) when { true::"test" };`,
			nil, testutil.Error}, // Rust: unexpected token
		{"unless", `permit (principal, action, resource) when { unless::"test" };`,
			ast.Permit().When(ast.EntityUID("unless", "test")), testutil.OK}, // Rust: OK
		{"when", `permit (principal, action, resource) when { when::"test" };`,
			ast.Permit().When(ast.EntityUID("when", "test")), testutil.OK}, // Rust: OK

		{"scope/action", `permit (principal == action::"test", action, resource);`,
			ast.Permit().PrincipalEq(types.NewEntityUID("action", "test")), testutil.OK}, // Rust: OK
		{"scope/context", `permit (principal == context::"test", action, resource);`,
			ast.Permit().PrincipalEq(types.NewEntityUID("context", "test")), testutil.OK}, // Rust: OK
		{"scope/else", `permit (principal == else::"test", action, resource);`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"scope/false", `permit (principal == false::"test", action, resource);`,
			nil, testutil.Error}, // Rust: unexpected token
		{"scope/forbid", `permit (principal == forbid::"test", action, resource);`,
			ast.Permit().PrincipalEq(types.NewEntityUID("forbid", "test")), testutil.OK}, // Rust: OK
		{"scope/has", `permit (principal == has::"test", action, resource);`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"scope/if", `permit (principal == if::"test", action, resource);`,
			nil, testutil.Error}, // Rust: unexpected token
		{"scope/in", `permit (principal == in::"test", action, resource);`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"scope/is", `permit (principal == is::"test", action, resource);`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"scope/like", `permit (principal == like::"test", action, resource);`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"scope/permit", `permit (principal == permit::"test", action, resource);`,
			ast.Permit().PrincipalEq(types.NewEntityUID("permit", "test")), testutil.OK}, // Rust: OK
		{"scope/principal", `permit (principal == principal::"test", action, resource);`,
			ast.Permit().PrincipalEq(types.NewEntityUID("principal", "test")), testutil.OK}, // Rust: OK
		{"scope/resource", `permit (principal == resource::"test", action, resource);`,
			ast.Permit().PrincipalEq(types.NewEntityUID("resource", "test")), testutil.OK}, // Rust: OK
		{"scope/then", `permit (principal == then::"test", action, resource);`,
			nil, testutil.Error}, // Rust: this identifier is reserved and cannot be used
		{"scope/true", `permit (principal == true::"test", action, resource);`,
			nil, testutil.Error}, // Rust: unexpected token
		{"scope/unless", `permit (principal == unless::"test", action, resource);`,
			ast.Permit().PrincipalEq(types.NewEntityUID("unless", "test")), testutil.OK}, // Rust: OK
		{"scope/when", `permit (principal == when::"test", action, resource);`,
			ast.Permit().PrincipalEq(types.NewEntityUID("when", "test")), testutil.OK}, // Rust: OK
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out parser.Policy
			err := out.UnmarshalCedar([]byte(tt.in))
			out.Position = ast.Position{}
			tt.err(t, err)
			if err == nil {
				testutil.Equals(t, &out, (*parser.Policy)(tt.out))
			}
		})
	}
}
