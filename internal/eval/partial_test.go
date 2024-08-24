package eval

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestPartial(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		ctx  *Context
		out  *ast.Policy
		keep bool
	}{
		{"smokeTest",
			ast.Permit(),
			&Context{},
			ast.Permit(),
			true,
		},
		{"principalEqual",
			ast.Permit().PrincipalEq(types.NewEntityUID("Account", "42")),
			&Context{
				Principal: types.NewEntityUID("Account", "42"),
			},
			ast.Permit(),
			true,
		},
		{"principalNotEqual",
			ast.Permit().PrincipalEq(types.NewEntityUID("Account", "42")),
			&Context{
				Principal: types.NewEntityUID("Account", "Other"),
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, keep := partialPolicy(tt.ctx, tt.in)
			if keep {
				testutil.Equals(t, out, tt.out)
				// gotP := (*parser.Policy)(out)
				// wantP := (*parser.Policy)(tt.out)
				// var gotB bytes.Buffer
				// gotP.MarshalCedar(&gotB)
				// var wantB bytes.Buffer
				// wantP.MarshalCedar(&wantB)
				// testutil.Equals(t, gotB.String(), wantB.String())
			}
			testutil.Equals(t, keep, tt.keep)

		})
	}

}
