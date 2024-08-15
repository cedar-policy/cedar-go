package eval

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestCompile(t *testing.T) {
	t.Parallel()
	e := Compile(ast.Permit())
	res, err := e.Eval(nil)
	testutil.OK(t, err)
	testutil.Equals(t, res, types.Value(types.True))
}

func TestPolicyToNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		out  ast.Node
	}{
		{
			"permit",
			ast.Permit(),
			ast.True().And(ast.True().And(ast.True())),
		},
		{
			"eqs",

			ast.Permit().
				PrincipalEq(types.NewEntityUID("Account", "principal")).
				ActionEq(types.NewEntityUID("Action", "test")).
				ResourceEq(types.NewEntityUID("Resource", "database")),

			ast.Principal().Equals(ast.EntityUID(types.NewEntityUID("Account", "principal"))).And(
				ast.Action().Equals(ast.EntityUID(types.NewEntityUID("Action", "test"))).And(
					ast.Resource().Equals(ast.EntityUID(types.NewEntityUID("Resource", "database"))),
				),
			),
		},

		{
			"conds",

			ast.Permit().
				When(ast.Long(123)).
				Unless(ast.Long(456)),

			ast.True().And(ast.True().And(ast.True().And(ast.Long(123).And(ast.Not(ast.Long(456)))))),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := policyToNode(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestScopeToNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		scope ast.NodeTypeVariable
		in    ast.IsScopeNode
		out   ast.Node
	}{
		{
			"all",
			ast.NewPrincipalNode(),
			ast.ScopeTypeAll{},
			ast.True(),
		},
		{
			"eq",
			ast.NewPrincipalNode(),
			ast.ScopeTypeEq{Entity: types.NewEntityUID("T", "42")},
			ast.Principal().Equals(ast.EntityUID(types.NewEntityUID("T", "42"))),
		},
		{
			"in",
			ast.NewPrincipalNode(),
			ast.ScopeTypeIn{Entity: types.NewEntityUID("T", "42")},
			ast.Principal().In(ast.EntityUID(types.NewEntityUID("T", "42"))),
		},
		{
			"inSet",
			ast.NewActionNode(),
			ast.ScopeTypeInSet{Entities: []types.EntityUID{types.NewEntityUID("T", "42")}},
			ast.Action().In(ast.SetDeprecated(types.Set{types.NewEntityUID("T", "42")})),
		},
		{
			"is",
			ast.NewResourceNode(),
			ast.ScopeTypeIs{Type: "T"},
			ast.Resource().Is("T"),
		},
		{
			"isIn",
			ast.NewResourceNode(),
			ast.ScopeTypeIsIn{Type: "T", Entity: types.NewEntityUID("T", "42")},
			ast.Resource().IsIn("T", ast.EntityUID(types.NewEntityUID("T", "42"))),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := scopeToNode(tt.scope, tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestScopeToNodePanic(t *testing.T) {
	t.Parallel()
	testutil.AssertPanic(t, func() {
		_ = scopeToNode(ast.NewPrincipalNode(), ast.ScopeNode{})
	})
}
