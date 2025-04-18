package ast_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
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
	private := "private"
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
	simpleRecord := types.NewRecord(types.RecordMap{
		"x": types.String("value"),
	})
	_ = ast.Forbid().
		When(
			ast.Value(simpleRecord).Access("x").Equal(ast.String("value")),
		).
		When(
			ast.Record(ast.Pairs{{Key: "x", Value: ast.Long(1).Add(ast.Context().Access("fooCount"))}}).
				Access("x").Equal(ast.Long(3)),
		).
		When(
			ast.Set(
				ast.Long(1),
				ast.Long(2).Add(ast.Long(3)).Multiply(ast.Long(4)),
				ast.Context().Access("fooCount"),
			).Contains(ast.Long(1)),
		)
}

func TestASTByTable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		out  ast.Policy
	}{
		{
			"permit",
			ast.Permit(),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"forbid",
			ast.Forbid(),
			ast.Policy{Effect: ast.EffectForbid, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"annotationPermit",
			ast.Annotation("key", "value").Permit(),
			ast.Policy{Annotations: []ast.AnnotationType{{Key: "key", Value: "value"}}, Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"annotationForbid",
			ast.Annotation("key", "value").Forbid(),
			ast.Policy{Annotations: []ast.AnnotationType{{Key: "key", Value: "value"}}, Effect: ast.EffectForbid, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"annotations",
			ast.Annotation("key", "value").Annotation("abc", "xyz").Permit(),
			ast.Policy{Annotations: []ast.AnnotationType{{Key: "key", Value: "value"}, {Key: "abc", Value: "xyz"}}, Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"policyAnnotate",
			ast.Permit().Annotate("key", "value"),
			ast.Policy{Annotations: []ast.AnnotationType{{Key: "key", Value: "value"}}, Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"when",
			ast.Permit().When(ast.True()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.True}}},
			},
		},
		{
			"unless",
			ast.Permit().Unless(ast.True()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionUnless, Body: ast.NodeValue{Value: types.True}}},
			},
		},
		{
			"scopePrincipalEq",
			ast.Permit().PrincipalEq(types.NewEntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeEq{Entity: types.NewEntityUID("T", "42")}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"scopePrincipalIn",
			ast.Permit().PrincipalIn(types.NewEntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeIn{Entity: types.NewEntityUID("T", "42")}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"scopePrincipalIs",
			ast.Permit().PrincipalIs("T"),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeIs{Type: types.EntityType("T")}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"scopePrincipalIsIn",
			ast.Permit().PrincipalIsIn("T", types.NewEntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeIsIn{Type: types.EntityType("T"), Entity: types.NewEntityUID("T", "42")}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"scopeActionEq",
			ast.Permit().ActionEq(types.NewEntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeEq{Entity: types.NewEntityUID("T", "42")}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"scopeActionIn",
			ast.Permit().ActionIn(types.NewEntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeIn{Entity: types.NewEntityUID("T", "42")}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"scopeActionInSet",
			ast.Permit().ActionInSet(types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeInSet{Entities: []types.EntityUID{types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43")}}, Resource: ast.ScopeTypeAll{}},
		},
		{
			"scopeResourceEq",
			ast.Permit().ResourceEq(types.NewEntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeEq{Entity: types.NewEntityUID("T", "42")}},
		},
		{
			"scopeResourceIn",
			ast.Permit().ResourceIn(types.NewEntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeIn{Entity: types.NewEntityUID("T", "42")}},
		},
		{
			"scopeResourceIs",
			ast.Permit().ResourceIs("T"),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeIs{Type: types.EntityType("T")}},
		},
		{
			"scopeResourceIsIn",
			ast.Permit().ResourceIsIn("T", types.NewEntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeIsIn{Type: types.EntityType("T"), Entity: types.NewEntityUID("T", "42")}},
		},
		{
			"variablePrincipal",
			ast.Permit().When(ast.Principal()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeVariable{Name: "principal"}}},
			},
		},
		{
			"variableAction",
			ast.Permit().When(ast.Action()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeVariable{Name: "action"}}},
			},
		},
		{
			"variableResource",
			ast.Permit().When(ast.Resource()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeVariable{Name: "resource"}}},
			},
		},
		{
			"variableContext",
			ast.Permit().When(ast.Context()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeVariable{Name: "context"}}},
			},
		},
		{
			"valueBoolFalse",
			ast.Permit().When(ast.Boolean(false)),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.False}}},
			},
		},
		{
			"valueBoolTrue",
			ast.Permit().When(ast.Boolean(true)),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.True}}},
			},
		},
		{
			"valueTrue",
			ast.Permit().When(ast.True()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.True}}},
			},
		},
		{
			"valueFalse",
			ast.Permit().When(ast.False()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.False}}},
			},
		},
		{
			"valueString",
			ast.Permit().When(ast.String("cedar")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.String("cedar")}}},
			},
		},
		{
			"valueLong",
			ast.Permit().When(ast.Long(42)),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.Long(42)}}},
			},
		},
		{
			"valueSet",
			ast.Permit().When(ast.Value(types.NewSet(types.Long(42), types.Long(43)))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.NewSet(types.Long(42), types.Long(43))}}},
			},
		},
		{
			"valueSetNodes",
			ast.Permit().When(ast.Set(ast.Long(42), ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeSet{Elements: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}, ast.NodeValue{Value: types.Long(43)}}}}},
			},
		},
		{
			"valueRecord",
			ast.Permit().When(ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(43)}))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.NewRecord(types.RecordMap{"key": types.Long(43)})}}},
			},
		},
		{
			"valueEntityUID",
			ast.Permit().When(ast.EntityUID("T", "42")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.NewEntityUID("T", "42")}}},
			},
		},
		{
			"valueIPAddr",
			ast.Permit().When(ast.IPAddr(netip.MustParsePrefix("127.0.0.1/16"))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeValue{Value: types.IPAddr(netip.MustParsePrefix("127.0.0.1/16"))}}},
			},
		},
		{
			"extensionCall",
			ast.Permit().When(ast.ExtensionCall("ip", ast.String("127.0.0.1"))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "ip", Args: []ast.IsNode{ast.NodeValue{Value: types.String("127.0.0.1")}}}}},
			}},
		{
			"opEquals",
			ast.Permit().When(ast.Long(42).Equal(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opNotEquals",
			ast.Permit().When(ast.Long(42).NotEqual(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeNotEquals{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opLessThan",
			ast.Permit().When(ast.Long(42).LessThan(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeLessThan{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opLessThanOrEqual",
			ast.Permit().When(ast.Long(42).LessThanOrEqual(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeLessThanOrEqual{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opGreaterThan",
			ast.Permit().When(ast.Long(42).GreaterThan(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opGreaterThanOrEqual",
			ast.Permit().When(ast.Long(42).GreaterThanOrEqual(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeGreaterThanOrEqual{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opLessThanExt",
			ast.Permit().When(ast.Long(42).DecimalLessThan(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "lessThan", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}, ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opLessThanOrEqualExt",
			ast.Permit().When(ast.Long(42).DecimalLessThanOrEqual(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "lessThanOrEqual", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}, ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opGreaterThanExt",
			ast.Permit().When(ast.Long(42).DecimalGreaterThan(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "greaterThan", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}, ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opGreaterThanOrEqualExt",
			ast.Permit().When(ast.Long(42).DecimalGreaterThanOrEqual(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "greaterThanOrEqual", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}, ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opLike",
			ast.Permit().When(ast.Long(42).Like(types.NewPattern(types.Wildcard{}))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeLike{Arg: ast.NodeValue{Value: types.Long(42)}, Value: types.NewPattern(types.Wildcard{})}}}},
		},
		{
			"opAnd",
			ast.Permit().When(ast.Long(42).And(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opOr",
			ast.Permit().When(ast.Long(42).Or(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeOr{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opNot",
			ast.Permit().When(ast.Not(ast.True())),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeNot{UnaryNode: ast.UnaryNode{Arg: ast.NodeValue{Value: types.True}}}}}},
		},
		{
			"opIf",
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.Long(42), ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeIfThenElse{If: ast.NodeValue{Value: types.True}, Then: ast.NodeValue{Value: types.Long(42)}, Else: ast.NodeValue{Value: types.Long(43)}}}}},
		},
		{
			"opPlus",
			ast.Permit().When(ast.Long(42).Add(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opMinus",
			ast.Permit().When(ast.Long(42).Subtract(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeSub{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opTimes",
			ast.Permit().When(ast.Long(42).Multiply(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeMult{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opNegate",
			ast.Permit().When(ast.Negate(ast.True())),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeNegate{UnaryNode: ast.UnaryNode{Arg: ast.NodeValue{Value: types.True}}}}}},
		},
		{
			"opIn",
			ast.Permit().When(ast.Long(42).In(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeIn{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opIs",
			ast.Permit().When(ast.Long(42).Is("T")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeIs{Left: ast.NodeValue{Value: types.Long(42)}, EntityType: types.EntityType("T")}}}},
		},
		{
			"opIsIn",
			ast.Permit().When(ast.Long(42).IsIn("T", ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeIsIn{NodeTypeIs: ast.NodeTypeIs{Left: ast.NodeValue{Value: types.Long(42)}, EntityType: types.EntityType("T")}, Entity: ast.NodeValue{Value: types.Long(43)}}}}},
		},
		{
			"opContains",
			ast.Permit().When(ast.Long(42).Contains(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeContains{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opContainsAll",
			ast.Permit().When(ast.Long(42).ContainsAll(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opContainsAny",
			ast.Permit().When(ast.Long(42).ContainsAny(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeContainsAny{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opGetTag",
			ast.Permit().When(ast.Long(42).GetTag(ast.String("key"))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.String("key")}}}}}},
		},
		{
			"opHasTag",
			ast.Permit().When(ast.Long(42).HasTag(ast.String("key"))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{Left: ast.NodeValue{Value: types.Long(42)}, Right: ast.NodeValue{Value: types.String("key")}}}}}},
		},
		{
			"opIsEmpty",
			ast.Permit().When(ast.Long(42).IsEmpty()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeIsEmpty{UnaryNode: ast.UnaryNode{Arg: ast.NodeValue{Value: types.Long(42)}}}}}},
		},
		{
			"opAccess",
			ast.Permit().When(ast.Long(42).Access("key")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{Arg: ast.NodeValue{Value: types.Long(42)}, Value: "key"}}}}},
		},
		{
			"opHas",
			ast.Permit().When(ast.Long(42).Has("key")),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{Arg: ast.NodeValue{Value: types.Long(42)}, Value: "key"}}}}},
		},
		{
			"opIsIpv4",
			ast.Permit().When(ast.Long(42).IsIpv4()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "isIpv4", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}}}}}},
		},
		{
			"opIsIpv6",
			ast.Permit().When(ast.Long(42).IsIpv6()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "isIpv6", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}}}}}},
		},
		{
			"opIsMulticast",
			ast.Permit().When(ast.Long(42).IsMulticast()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "isMulticast", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}}}}}},
		},
		{
			"opIsLoopback",
			ast.Permit().When(ast.Long(42).IsLoopback()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "isLoopback", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}}}}}},
		},
		{
			"opIsInRange",
			ast.Permit().When(ast.Long(42).IsInRange(ast.Long(43))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "isInRange", Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}, ast.NodeValue{Value: types.Long(43)}}}}}},
		},
		{
			"opOffset",
			ast.Permit().When(ast.Datetime(time.Time{}).Offset(ast.Duration(time.Duration(100)))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "offset", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDatetime(time.Time{})}, ast.NodeValue{Value: types.NewDuration(time.Duration(100))}}}}}},
		},
		{
			"opDurationSince",
			ast.Permit().When(ast.Datetime(time.Time{}).DurationSince(ast.Datetime(time.Time{}))),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "durationSince", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDatetime(time.Time{})}, ast.NodeValue{Value: types.NewDatetime(time.Time{})}}}}}},
		},
		{
			"opToDate",
			ast.Permit().When(ast.Datetime(time.Time{}).ToDate()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "toDate", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDatetime(time.Time{})}}}}}},
		},
		{
			"opToTime",
			ast.Permit().When(ast.Datetime(time.Time{}).ToTime()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "toTime", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDatetime(time.Time{})}}}}}},
		},
		{
			"opToDays",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToDays()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "toDays", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDuration(time.Duration(100))}}}}}},
		},
		{
			"opToHours",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToHours()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "toHours", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDuration(time.Duration(100))}}}}}},
		},
		{"opToMinutes",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToMinutes()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "toMinutes", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDuration(time.Duration(100))}}}}}},
		},
		{
			"opToSeconds",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToSeconds()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "toSeconds", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDuration(time.Duration(100))}}}}}},
		},
		{
			"opToMilliseconds",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToMilliseconds()),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeExtensionCall{Name: "toMilliseconds", Args: []ast.IsNode{ast.NodeValue{Value: types.NewDuration(time.Duration(100))}}}}}},
		},

		{
			"duplicateAnnotations",
			ast.Permit().Annotate("key", "value").Annotate("key", "value2"),
			ast.Policy{Annotations: []ast.AnnotationType{{Key: "key", Value: "value2"}}, Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{}},
		},

		{
			"valueRecordElements",
			ast.Permit().When(ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(42)}})),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeRecord{Elements: []ast.RecordElementNode{{Key: "key", Value: ast.NodeValue{Value: types.Long(42)}}}}}}},
		},
		{
			"duplicateValueRecordElements",
			ast.Permit().When(ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(42)}, {Key: "key", Value: ast.Long(43)}})),
			ast.Policy{Effect: ast.EffectPermit, Principal: ast.ScopeTypeAll{}, Action: ast.ScopeTypeAll{}, Resource: ast.ScopeTypeAll{},
				Conditions: []ast.ConditionType{{Condition: ast.ConditionWhen, Body: ast.NodeTypeRecord{Elements: []ast.RecordElementNode{{Key: "key", Value: ast.NodeValue{Value: types.Long(43)}}}}}}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.in, &tt.out)
		})
	}
}
