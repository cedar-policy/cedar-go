package parser

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestScopeToNode(t *testing.T) {
	t.Parallel()
	t.Run("all", func(t *testing.T) {
		t.Parallel()
		x := scopeToNode(ast.NodeTypeVariable{Name: "principal"}, ast.ScopeTypeAll{})
		testutil.Equals(t, x, ast.True())
	})
}

func TestAstNodeToMarshalNode(t *testing.T) {
	t.Parallel()
	t.Run("panic", func(t *testing.T) {
		t.Parallel()
		testutil.Panic(t, func() {
			astNodeToMarshalNode(nil)
		})
	})
}
