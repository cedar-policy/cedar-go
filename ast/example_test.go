package ast_test

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/types"
)

// This example shows a basic programmatic AST construction via the Permit() builder:
func Example() {
	johnny := types.NewEntityUID("FolkHeroes", "johnnyChapman")
	sow := types.NewEntityUID("Action", "sow")
	cast := types.NewEntityUID("Action", "cast")
	midwest := types.NewEntityUID("Locations::USA::Regions", "midwest")

	policy := ast.Permit().
		PrincipalEq(johnny).
		ActionInSet(sow, cast).
		ResourceIs("Crops::Apple").
		When(ast.Context().Access("location").In(ast.Value(midwest))).
		Unless(ast.Context().Access("season").Equal(ast.String("winter")))

	fmt.Println(string(policy.MarshalCedar()))

	// Output:
	// permit (
	//     principal == FolkHeroes::"johnnyChapman",
	//     action in [Action::"sow", Action::"cast"],
	//     resource is Crops::Apple
	// )
	// when { context.location in Locations::USA::Regions::"midwest" }
	// unless { context.season == "winter" };
}

// To programmatically create policies with annotations, use the Annotation() builder:
func Example_annotation() {
	policy := ast.Annotation("example1", "value").
		Annotation("example2", "").
		Forbid()

	fmt.Println(string(policy.MarshalCedar()))

	// Output:
	// @example1("value")
	// @example2("")
	// forbid ( principal, action, resource );
}

// This example shows how precedence can be expressed using the AST builder syntax:
func Example_precedence() {
	// The argument passed to .Add() is the entire right-hand side of the expression, so 1 + 5 is evaluated with
	// higher precedence than the subsequent multiplication by 10.
	policy := ast.Permit().
		When(ast.Long(1).Add(ast.Long(5)).Multiply(ast.Long(10)).Equal(ast.Long(60)))

	fmt.Println(string(policy.MarshalCedar()))

	// Output:
	// permit ( principal, action, resource )
	// when { (1 + 5) * 10 == 60 };
}

// Extension functions can be explicitly called by using the appropriate builder with the ExtensionCall suffix. This
// example demonstrates the use of DecimalExtensionCall():
func Example_explicitExtensionCall() {
	policy := ast.Forbid().
		When(
			ast.Resource().Access("angleRadians").DecimalGreaterThan(
				ast.DecimalExtensionCall(ast.String("3.1415")),
			),
		)

	fmt.Println(string(policy.MarshalCedar()))

	// Output:
	// forbid ( principal, action, resource )
	// when { resource.angleRadians.greaterThan(decimal("3.1415")) };
}

func ExampleRecord() {
	// Literal records can be constructed and passed via the ast.Value() builder
	literalRecord := types.NewRecord(types.RecordMap{
		"x": types.String("value1"),
		"y": types.String("value2"),
	})

	// Records with internal expressions are constructed via the ast.Record() builder
	exprRecord := ast.Record(ast.Pairs{
		{
			Key:   "x",
			Value: ast.Long(1).Add(ast.Context().Access("fooCount")),
		},
		{
			Key:   "y",
			Value: ast.Long(8),
		},
	})

	policy := ast.Forbid().
		When(
			ast.Value(literalRecord).Access("x").Equal(ast.String("value1")),
		).
		When(
			exprRecord.Access("x").Equal(ast.Long(3)),
		)

	fmt.Println(string(policy.MarshalCedar()))

	// Output:
	// forbid ( principal, action, resource )
	// when { {"x":"value1", "y":"value2"}.x == "value1" }
	// when { {"x":(1 + context.fooCount), "y":8}.x == 3 };
}
