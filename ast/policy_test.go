package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestPolicy_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("roundtrip", func(t *testing.T) {
		p := ast.Permit().PrincipalEq(types.NewEntityUID("Foo::Bar", "Baz"))
		expected := `{
		"effect": "permit",
		"principal": {
			"op": "==",
			"entity": {
				"type": "Foo::Bar",
				"id": "Baz"
			}
		},
		"action": {
			"op": "All"
		},
		"resource": {
			"op": "All"
		}
	}`
		testutil.JSONMarshalsTo(t, p, expected)

		var unmarshaled ast.Policy
		err := unmarshaled.UnmarshalJSON([]byte(expected))
		testutil.OK(t, err)
		testutil.Equals(t, &unmarshaled, p)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		var unmarshaled ast.Policy
		err := unmarshaled.UnmarshalJSON([]byte("[]"))
		testutil.Error(t, err)
	})
}

func TestPolicy_MarshalCedar(t *testing.T) {
	t.Parallel()

	t.Run("roundtrip", func(t *testing.T) {
		p := ast.Permit().PrincipalEq(types.NewEntityUID("Foo::Bar", "Baz"))
		expected := `permit (
    principal == Foo::Bar::"Baz",
    action,
    resource
);`

		testutil.Equals(t, string(p.MarshalCedar()), expected)

		var unmarshaled ast.Policy
		err := unmarshaled.UnmarshalCedar([]byte(expected))

		p.Position = internalast.Position{Offset: 0, Line: 1, Column: 1}
		testutil.OK(t, err)
		testutil.Equals(t, &unmarshaled, p)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		var unmarshaled ast.Policy
		err := unmarshaled.UnmarshalCedar([]byte("this isn't Cedar"))
		testutil.Error(t, err)
	})
}
