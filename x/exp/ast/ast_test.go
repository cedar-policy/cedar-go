package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/types"
)

// These tests mostly verify that policy ASTs compile
func TestAst(t *testing.T) {
	t.Parallel()

	johnny := types.EntityUID{"user", "johnny"}
	sow := types.EntityUID{"Action", "sow"}
	cast := types.EntityUID{"Action", "cast"}

	_ = ast.Permit().
		Annotate("example", "one").
		PrincipalEq(johnny).
		ActionIn(sow, cast).
		When(ast.True()).
		Unless(ast.False())

	_ = ast.Forbid().
		Annotate("example", "two").
		PrincipalEq(johnny).
		ResourceIn(types.EntityUID{"Classification", "Poisonous"})

	private := types.String("private")

	_ = ast.Forbid().Annotate("example", "three").
		When(
			// TODO: It's a little annoying that we have to wrap private in ast.String here.
			ast.Resource().Access("tags").Contains(ast.String(private)),
		).
		Unless(
			ast.Resource().In(ast.Principal().Access("allowed_resources")),
		)
}
