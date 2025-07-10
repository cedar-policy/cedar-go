package eval

import (
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestCompile(t *testing.T) {
	t.Parallel()
	e := Compile(ast.Permit())
	res, err := e.Eval(Env{})
	testutil.OK(t, err)
	testutil.Equals(t, res, types.True)
}

func TestBoolEvaler(t *testing.T) {
	t.Parallel()
	t.Run("Happy", func(t *testing.T) {
		t.Parallel()
		b := BoolEvaler{eval: newLiteralEval(types.True)}
		v, err := b.Eval(Env{})
		testutil.OK(t, err)
		testutil.Equals(t, v, true)
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		errWant := fmt.Errorf("error")
		b := BoolEvaler{eval: newErrorEval(errWant)}
		v, err := b.Eval(Env{})
		testutil.ErrorIs(t, err, errWant)
		testutil.Equals(t, v, false)
	})

	t.Run("NonBool", func(t *testing.T) {
		t.Parallel()
		b := BoolEvaler{eval: newLiteralEval(types.String("bad"))}
		v, err := b.Eval(Env{})
		testutil.ErrorIs(t, err, ErrType)
		testutil.Equals(t, v, false)
	})

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
			ast.True(),
		},
		{
			"forbid",
			ast.Forbid(),
			ast.True(),
		},
		{
			"eqs",

			ast.Permit().
				PrincipalEq(types.NewEntityUID("Account", "principal")).
				ActionEq(types.NewEntityUID("Action", "test")).
				ResourceEq(types.NewEntityUID("Resource", "database")),

			ast.Principal().Equal(ast.EntityUID("Account", "principal")).And(
				ast.Action().Equal(ast.EntityUID("Action", "test")).And(
					ast.Resource().Equal(ast.EntityUID("Resource", "database")),
				),
			),
		},

		{
			"conds",

			ast.Permit().
				When(ast.Long(123)).
				Unless(ast.Long(456)),

			ast.True().And(ast.Long(123).And(ast.Not(ast.Long(456)))),
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
			ast.Principal().Equal(ast.EntityUID("T", "42")),
		},
		{
			"in",
			ast.NewPrincipalNode(),
			ast.ScopeTypeIn{Entity: types.NewEntityUID("T", "42")},
			ast.Principal().In(ast.EntityUID("T", "42")),
		},
		{
			"inSet",
			ast.NewActionNode(),
			ast.ScopeTypeInSet{Entities: []types.EntityUID{types.NewEntityUID("T", "42")}},
			ast.Action().In(ast.Value(types.NewSet(types.NewEntityUID("T", "42")))),
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
			ast.Resource().IsIn("T", ast.EntityUID("T", "42")),
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
	testutil.Panic(t, func() {
		_ = scopeToNode(ast.NewPrincipalNode(), ast.ScopeNode{})
	})
}
