package eval

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestBake(t *testing.T) {
	tests := []struct {
		name string
		in   ast.Node
		out  ast.Node
	}{
		{"record-baked",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.True()}}),
			ast.Value(types.Record{"key": types.True}),
		},
		{"set-baked",
			ast.Set(ast.True()),
			ast.Value(types.Set{types.True}),
		},
		{"record-raw",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(6).Multiply(ast.Long(7))}}),
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(6).Multiply(ast.Long(7))}}),
		},
		{"set-raw",
			ast.Set(ast.Long(6).Multiply(ast.Long(7))),
			ast.Set(ast.Long(6).Multiply(ast.Long(7))),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			out := bake(tt.in.AsIsNode())
			testutil.Equals(t, out, tt.out.AsIsNode())
		})
	}
}
