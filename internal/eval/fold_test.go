package eval

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestFold(t *testing.T) {
	tests := []struct {
		name string
		in   ast.Node
		out  ast.Node
	}{
		{"record-bake",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.True()}}),
			ast.Value(types.Record{"key": types.True}),
		},
		{"set-bake",
			ast.Set(ast.True()),
			ast.Value(types.Set{types.True}),
		},
		{"record-fold",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(6).Multiply(ast.Long(7))}}),
			ast.Value(types.Record{"key": types.Long(42)}),
		},
		{"set-fold",
			ast.Set(ast.Long(6).Multiply(ast.Long(7))),
			ast.Value(types.Set{types.Long(42)}),
		},
		{"record-blocked",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(6).Multiply(ast.Context())}}),
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(6).Multiply(ast.Context())}}),
		},
		{"set-blocked",
			ast.Set(ast.Long(6).Multiply(ast.Context())),
			ast.Set(ast.Long(6).Multiply(ast.Context())),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			out := fold(tt.in.AsIsNode())
			testutil.Equals(t, out, tt.out.AsIsNode())
		})
	}
}
