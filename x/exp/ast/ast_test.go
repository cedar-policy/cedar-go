package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// These tests mostly verify that policy ASTs compile
func TestAst(t *testing.T) {
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
	_ = ast.Permit().
		Annotate("example", "one").
		PrincipalEq(johnny).
		ActionIn(sow, cast).
		When(ast.True()).
		Unless(ast.False())

	// @example("two")
	// forbid (principal, action, resource)
	// when { resource.tags.contains("private") }
	// unless { resource in principal.allowed_resources };
	private := types.String("private")
	_ = ast.Forbid().Annotate("example", "two").
		When(
			ast.Resource().Access("tags").Contains(ast.String(private)),
		).
		Unless(
			ast.Resource().In(ast.Principal().Access("allowed_resources")),
		)

	// forbid (principal, action, resource)
	// when { resource[context.resourceField] == "specialValue" }
	// when { {x: "value"}.x == "value" }
	// when { {x: 1 + context.fooCount}.x == 3 }
	// when { [1, 2 + 3, context.fooCount].contains(1) };
	simpleRecord := types.Record{
		"x": types.String("value"),
	}
	_ = ast.Forbid().
		When(
			ast.Resource().AccessNode(
				ast.Context().Access("resourceField"),
			).Equals(ast.String("specialValue")),
		).
		When(
			ast.Record(simpleRecord).Access("x").Equals(ast.String("value")),
		).
		When(
			ast.RecordNodes(map[types.String]ast.Node{
				"x": ast.Long(1).Plus(ast.Context().Access("fooCount")),
			}).Access("x").Equals(ast.Long(3)),
		).
		When(
			ast.SetNodes([]ast.Node{
				ast.Long(1),
				ast.Long(2).Plus(ast.Long(3)),
				ast.Context().Access("fooCount"),
			}).Contains(ast.Long(1)),
		)
}
