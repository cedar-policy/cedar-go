package ast_test

import (
	"net/netip"
	"testing"

	"github.com/cedar-policy/cedar-go/ast"
	internalast "github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

// These tests serve mostly as examples of how to translate from Cedar text into programmatic AST construction. They
// don't verify anything.
func TestAstExamples(t *testing.T) {
	t.Parallel()

	johnny := types.NewEntityUID("User", "johnny")
	sow := types.NewEntityUID("Action", "sow")
	cast := types.NewEntityUID("Action", "cast")

	// @example("one")
	// permit (
	//     principal == User::"johnny"
	//     action in [Action::"sow", Action::"cast"]
	//     resource
	// )
	// when { true }
	// unless { false };
	_ = ast.Annotation("example", "one").
		Permit().
		PrincipalIsIn("User", johnny).
		ActionInSet(sow, cast).
		When(ast.True()).
		Unless(ast.False())

	// @example("two")
	// forbid (principal, action, resource)
	// when { resource.tags.contains("private") }
	// unless { resource in principal.allowed_resources };
	private := types.String("private")
	_ = ast.Annotation("example", "two").
		Forbid().
		When(
			ast.Resource().Access("tags").Contains(ast.String(private)),
		).
		Unless(
			ast.Resource().In(ast.Principal().Access("allowed_resources")),
		)

	// forbid (principal, action, resource)
	// when { {x: "value"}.x == "value" }
	// when { {x: 1 + context.fooCount}.x == 3 }
	// when { [1, (2 + 3) * 4, context.fooCount].contains(1) };
	simpleRecord := types.Record{
		"x": types.String("value"),
	}
	_ = ast.Forbid().
		When(
			ast.Record(simpleRecord).Access("x").Equals(ast.String("value")),
		).
		When(
			ast.RecordNodes(map[types.String]ast.Node{
				"x": ast.Long(1).Plus(ast.Context().Access("fooCount")),
			}).Access("x").Equals(ast.Long(3)),
		).
		When(
			ast.SetNodes(
				ast.Long(1),
				ast.Long(2).Plus(ast.Long(3)).Times(ast.Long(4)),
				ast.Context().Access("fooCount"),
			).Contains(ast.Long(1)),
		)
}

func TestASTByTable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		out  *internalast.Policy
	}{
		{"permit", ast.Permit(), internalast.Permit()},
		{"forbid", ast.Forbid(), internalast.Forbid()},
		{"annotationPermit", ast.Annotation("key", "value").Permit(), internalast.Annotation("key", "value").Permit()},
		{"annotationForbid", ast.Annotation("key", "value").Forbid(), internalast.Annotation("key", "value").Forbid()},
		{"annotations", ast.Annotation("key", "value").Annotation("abc", "xyz").Permit(), internalast.Annotation("key", "value").Annotation("abc", "xyz").Permit()},
		{"policyAnnotate", ast.Permit().Annotate("key", "value"), internalast.Permit().Annotate("key", "value")},
		{"when", ast.Permit().When(ast.True()), internalast.Permit().When(internalast.True())},
		{"unless", ast.Permit().Unless(ast.True()), internalast.Permit().Unless(internalast.True())},
		{"scopePrincipalEq", ast.Permit().PrincipalEq(types.NewEntityUID("T", "42")), internalast.Permit().PrincipalEq(types.NewEntityUID("T", "42"))},
		{"scopePrincipalIn", ast.Permit().PrincipalIn(types.NewEntityUID("T", "42")), internalast.Permit().PrincipalIn(types.NewEntityUID("T", "42"))},
		{"scopePrincipalIs", ast.Permit().PrincipalIs("T"), internalast.Permit().PrincipalIs("T")},
		{"scopePrincipalIsIn", ast.Permit().PrincipalIsIn("T", types.NewEntityUID("T", "42")), internalast.Permit().PrincipalIsIn("T", types.NewEntityUID("T", "42"))},
		{"scopeActionEq", ast.Permit().ActionEq(types.NewEntityUID("T", "42")), internalast.Permit().ActionEq(types.NewEntityUID("T", "42"))},
		{"scopeActionIn", ast.Permit().ActionIn(types.NewEntityUID("T", "42")), internalast.Permit().ActionIn(types.NewEntityUID("T", "42"))},
		{"scopeActionInSet", ast.Permit().ActionInSet(types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43")), internalast.Permit().ActionInSet(types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43"))},
		{"scopeResourceEq", ast.Permit().ResourceEq(types.NewEntityUID("T", "42")), internalast.Permit().ResourceEq(types.NewEntityUID("T", "42"))},
		{"scopeResourceIn", ast.Permit().ResourceIn(types.NewEntityUID("T", "42")), internalast.Permit().ResourceIn(types.NewEntityUID("T", "42"))},
		{"scopeResourceIs", ast.Permit().ResourceIs("T"), internalast.Permit().ResourceIs("T")},
		{"scopeResourceIsIn", ast.Permit().ResourceIsIn("T", types.NewEntityUID("T", "42")), internalast.Permit().ResourceIsIn("T", types.NewEntityUID("T", "42"))},
		{"variablePrincipal", ast.Permit().When(ast.Principal()), internalast.Permit().When(internalast.Principal())},
		{"variableAction", ast.Permit().When(ast.Action()), internalast.Permit().When(internalast.Action())},
		{"variableResource", ast.Permit().When(ast.Resource()), internalast.Permit().When(internalast.Resource())},
		{"variableContext", ast.Permit().When(ast.Context()), internalast.Permit().When(internalast.Context())},
		{"valueBoolFalse", ast.Permit().When(ast.Boolean(false)), internalast.Permit().When(internalast.Boolean(false))},
		{"valueBoolTrue", ast.Permit().When(ast.Boolean(true)), internalast.Permit().When(internalast.Boolean(true))},
		{"valueTrue", ast.Permit().When(ast.True()), internalast.Permit().When(internalast.True())},
		{"valueFalse", ast.Permit().When(ast.False()), internalast.Permit().When(internalast.False())},
		{"valueString", ast.Permit().When(ast.String("cedar")), internalast.Permit().When(internalast.String("cedar"))},
		{"valueLong", ast.Permit().When(ast.Long(42)), internalast.Permit().When(internalast.Long(42))},
		{"valueSet", ast.Permit().When(ast.Set(types.Set{types.Long(42), types.Long(43)})), internalast.Permit().When(internalast.Set(types.Set{types.Long(42), types.Long(43)}))},
		{"valueSetNodes", ast.Permit().When(ast.SetNodes(ast.Long(42), ast.Long(43))), internalast.Permit().When(internalast.SetNodes(internalast.Long(42), internalast.Long(43)))},
		{"valueRecord", ast.Permit().When(ast.Record(types.Record{"key": types.Long(43)})), internalast.Permit().When(internalast.Record(types.Record{"key": types.Long(43)}))},
		{"valueRecordNodes", ast.Permit().When(ast.RecordNodes(map[types.String]ast.Node{"key": ast.Long(42)})), internalast.Permit().When(internalast.RecordNodes(map[types.String]internalast.Node{"key": internalast.Long(42)}))},
		{"valueRecordElements", ast.Permit().When(ast.RecordElements(ast.RecordElement{Key: "key", Value: ast.Long(42)})), internalast.Permit().When(internalast.RecordElements(internalast.RecordElement{Key: "key", Value: internalast.Long(42)}))},
		{"valueEntityUID", ast.Permit().When(ast.EntityUID(types.NewEntityUID("T", "42"))), internalast.Permit().When(internalast.EntityUID(types.NewEntityUID("T", "42")))},
		{"valueDecimal", ast.Permit().When(ast.Decimal(420000)), internalast.Permit().When(internalast.Decimal(420000))},
		{"valueIPAddr", ast.Permit().When(ast.IPAddr(types.IPAddr(netip.MustParsePrefix("127.0.0.1/16")))), internalast.Permit().When(internalast.IPAddr(types.IPAddr(netip.MustParsePrefix("127.0.0.1/16"))))},
		{"extensionCall", ast.Permit().When(ast.ExtensionCall("ip", ast.String("127.0.0.1"))), internalast.Permit().When(internalast.ExtensionCall("ip", internalast.String("127.0.0.1")))},
		{"opEquals", ast.Permit().When(ast.Long(42).Equals(ast.Long(43))), internalast.Permit().When(internalast.Long(42).Equals(internalast.Long(43)))},
		{"opNotEquals", ast.Permit().When(ast.Long(42).NotEquals(ast.Long(43))), internalast.Permit().When(internalast.Long(42).NotEquals(internalast.Long(43)))},
		{"opLessThan", ast.Permit().When(ast.Long(42).LessThan(ast.Long(43))), internalast.Permit().When(internalast.Long(42).LessThan(internalast.Long(43)))},
		{"opLessThanOrEqual", ast.Permit().When(ast.Long(42).LessThanOrEqual(ast.Long(43))), internalast.Permit().When(internalast.Long(42).LessThanOrEqual(internalast.Long(43)))},
		{"opGreaterThan", ast.Permit().When(ast.Long(42).GreaterThan(ast.Long(43))), internalast.Permit().When(internalast.Long(42).GreaterThan(internalast.Long(43)))},
		{"opGreaterThanOrEqual", ast.Permit().When(ast.Long(42).GreaterThanOrEqual(ast.Long(43))), internalast.Permit().When(internalast.Long(42).GreaterThanOrEqual(internalast.Long(43)))},
		{"opLessThanExt", ast.Permit().When(ast.Long(42).LessThanExt(ast.Long(43))), internalast.Permit().When(internalast.Long(42).LessThanExt(internalast.Long(43)))},
		{"opLessThanOrEqualExt", ast.Permit().When(ast.Long(42).LessThanOrEqualExt(ast.Long(43))), internalast.Permit().When(internalast.Long(42).LessThanOrEqualExt(internalast.Long(43)))},
		{"opGreaterThanExt", ast.Permit().When(ast.Long(42).GreaterThanExt(ast.Long(43))), internalast.Permit().When(internalast.Long(42).GreaterThanExt(internalast.Long(43)))},
		{"opGreaterThanOrEqualExt", ast.Permit().When(ast.Long(42).GreaterThanOrEqualExt(ast.Long(43))), internalast.Permit().When(internalast.Long(42).GreaterThanOrEqualExt(internalast.Long(43)))},
		{"opLike", ast.Permit().When(ast.Long(42).Like(types.Pattern{})), internalast.Permit().When(internalast.Long(42).Like(types.Pattern{}))},
		{"opAnd", ast.Permit().When(ast.Long(42).And(ast.Long(43))), internalast.Permit().When(internalast.Long(42).And(internalast.Long(43)))},
		{"opOr", ast.Permit().When(ast.Long(42).Or(ast.Long(43))), internalast.Permit().When(internalast.Long(42).Or(internalast.Long(43)))},
		{"opNot", ast.Permit().When(ast.Not(ast.True())), internalast.Permit().When(internalast.Not(internalast.True()))},
		{"opIf", ast.Permit().When(ast.If(ast.True(), ast.Long(42), ast.Long(43))), internalast.Permit().When(internalast.If(internalast.True(), internalast.Long(42), internalast.Long(43)))},
		{"opPlus", ast.Permit().When(ast.Long(42).Plus(ast.Long(43))), internalast.Permit().When(internalast.Long(42).Plus(internalast.Long(43)))},
		{"opMinus", ast.Permit().When(ast.Long(42).Minus(ast.Long(43))), internalast.Permit().When(internalast.Long(42).Minus(internalast.Long(43)))},
		{"opTimes", ast.Permit().When(ast.Long(42).Times(ast.Long(43))), internalast.Permit().When(internalast.Long(42).Times(internalast.Long(43)))},
		{"opNegate", ast.Permit().When(ast.Negate(ast.True())), internalast.Permit().When(internalast.Negate(internalast.True()))},
		{"opIn", ast.Permit().When(ast.Long(42).In(ast.Long(43))), internalast.Permit().When(internalast.Long(42).In(internalast.Long(43)))},
		{"opIs", ast.Permit().When(ast.Long(42).Is(types.Path("T"))), internalast.Permit().When(internalast.Long(42).Is(types.Path("T")))},
		{"opIsIn", ast.Permit().When(ast.Long(42).IsIn(types.Path("T"), ast.Long(43))), internalast.Permit().When(internalast.Long(42).IsIn(types.Path("T"), internalast.Long(43)))},
		{"opContains", ast.Permit().When(ast.Long(42).Contains(ast.Long(43))), internalast.Permit().When(internalast.Long(42).Contains(internalast.Long(43)))},
		{"opContainsAll", ast.Permit().When(ast.Long(42).ContainsAll(ast.Long(43))), internalast.Permit().When(internalast.Long(42).ContainsAll(internalast.Long(43)))},
		{"opContainsAny", ast.Permit().When(ast.Long(42).ContainsAny(ast.Long(43))), internalast.Permit().When(internalast.Long(42).ContainsAny(internalast.Long(43)))},
		{"opAccess", ast.Permit().When(ast.Long(42).Access("key")), internalast.Permit().When(internalast.Long(42).Access("key"))},
		{"opHas", ast.Permit().When(ast.Long(42).Has("key")), internalast.Permit().When(internalast.Long(42).Has("key"))},
		{"opIsIpv4", ast.Permit().When(ast.Long(42).IsIpv4()), internalast.Permit().When(internalast.Long(42).IsIpv4())},
		{"opIsIpv6", ast.Permit().When(ast.Long(42).IsIpv6()), internalast.Permit().When(internalast.Long(42).IsIpv6())},
		{"opIsMulticast", ast.Permit().When(ast.Long(42).IsMulticast()), internalast.Permit().When(internalast.Long(42).IsMulticast())},
		{"opIsLoopback", ast.Permit().When(ast.Long(42).IsLoopback()), internalast.Permit().When(internalast.Long(42).IsLoopback())},
		{"opIsInRange", ast.Permit().When(ast.Long(42).IsInRange(ast.Long(43))), internalast.Permit().When(internalast.Long(42).IsInRange(internalast.Long(43)))},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, (*internalast.Policy)(tt.in), tt.out)
		})
	}
}
