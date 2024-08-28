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
		{"conditionOmitTrue",
			ast.Permit().When(ast.True()),
			&Context{},
			ast.Permit(),
			true,
		},
		{"conditionDropFalse",
			ast.Permit().When(ast.False()),
			&Context{},
			nil,
			false,
		},
		{"conditionDropError",
			ast.Permit().When(ast.Long(42).GreaterThan(ast.String("bananas"))),
			&Context{},
			nil,
			false,
		},
		{"conditionDropTypeError",
			ast.Permit().When(ast.Long(42)),
			&Context{},
			nil,
			false,
		},
		{"conditionKeepUnfolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			&Context{},
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			true,
		},
		{"conditionOmitTrueFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			&Context{
				Context: types.Long(43),
			},
			ast.Permit(),
			true,
		},
		{"conditionDropFalseFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			&Context{
				Context: types.Long(41),
			},
			nil,
			false,
		},
		{"conditionDropErrorFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			&Context{
				Context: types.String("bananas"),
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, keep := partialPolicy(PrepContext(tt.ctx), tt.in)
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

func TestPartialIfThenElse(t *testing.T) {
	errorN := ast.Long(42).GreaterThan(ast.String("bananas"))
	trueN := ast.True()
	falseN := ast.False()
	valueN := ast.String("test")
	keepN := ast.Context()
	_, _, _, _, _ = errorN, trueN, falseN, valueN, keepN
	valueA := ast.String("a")
	valueB := ast.String("b")

	tests := []struct {
		name string
		in   ast.Node
		out  any
		ok   bool
	}{
		{"ifTrueAB", ast.IfThenElse(trueN, valueA, valueB), valueA, true},
		{"ifFalseAB", ast.IfThenElse(falseN, valueA, valueB), valueB, true},
		{"ifValueAB", ast.IfThenElse(valueN, valueA, valueB), nil, false},
		{"ifKeepAB", ast.IfThenElse(keepN, valueA, valueB), ast.IfThenElse(keepN, valueA, valueB), true},
		{"ifErrorAB", ast.IfThenElse(errorN, valueA, valueB), nil, false},

		{"ifTrueErrorB", ast.IfThenElse(trueN, errorN, valueB), nil, false},
		{"ifFalseAError", ast.IfThenElse(falseN, valueA, errorN), nil, false},
		{"ifTrueAError", ast.IfThenElse(trueN, valueA, errorN), valueA, true},
		{"ifFalseErrorB", ast.IfThenElse(falseN, errorN, valueB), valueB, true},

		{"ifKeepKeepKeep", ast.IfThenElse(keepN, keepN, keepN), ast.IfThenElse(keepN, keepN, keepN), true},
		{"ifKeepErrorError", ast.IfThenElse(keepN, errorN, errorN), ast.IfThenElse(keepN, ast.ExtensionCall("error"), ast.ExtensionCall("error")), true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeIfThenElse)
			testutil.Equals(t, ok, true)
			out, ok := partialIfThenElse(&Context{}, n)
			testutil.Equals(t, ok, tt.ok)
			if !tt.ok && out == nil && tt.out == nil {
				return
			}
			nd, ok := tt.out.(ast.Node)
			testutil.Equals(t, ok, true)
			testutil.Equals(t, out, nd.AsIsNode())
		})
	}
}

func TestPartialAnd(t *testing.T) {
	errorN := ast.Long(42).GreaterThan(ast.String("bananas"))
	trueN := ast.True()
	falseN := ast.False()
	valueN := ast.String("test")
	keepN := ast.Context()
	_, _, _, _, _ = errorN, trueN, falseN, valueN, keepN

	tests := []struct {
		name string
		in   ast.Node
		out  any
		ok   bool
	}{

		{"andTrueTrue", trueN.And(trueN), ast.True(), true},
		{"andTrueFalse", trueN.And(falseN), ast.False(), true},
		{"andTrueValue", trueN.And(valueN), nil, false},
		{"andTrueKeep", trueN.And(keepN), trueN.And(keepN), true},
		{"andTrueError", trueN.And(errorN), nil, false},

		{"andFalseTrue", falseN.And(trueN), ast.False(), true},
		{"andFalseFalse", falseN.And(falseN), ast.False(), true},
		{"andFalseValue", falseN.And(valueN), ast.False(), true},
		{"andFalseKeep", falseN.And(keepN), ast.False(), true},
		{"andFalseError", falseN.And(errorN), ast.False(), true},

		{"andValueTrue", valueN.And(trueN), nil, false},
		{"andValueFalse", valueN.And(falseN), nil, false},
		{"andValueValue", valueN.And(valueN), nil, false},
		{"andValueKeep", valueN.And(keepN), nil, false},
		{"andValueError", valueN.And(errorN), nil, false},

		{"andKeepTrue", keepN.And(trueN), keepN.And(trueN), true},
		{"andKeepFalse", keepN.And(falseN), keepN.And(falseN), true},
		{"andKeepValue", keepN.And(valueN), keepN.And(valueN), true},
		{"andKeepKeep", keepN.And(keepN), keepN.And(keepN), true},
		{"andKeepError", keepN.And(errorN), keepN.And(ast.ExtensionCall("error")), true},

		{"andErrorTrue", errorN.And(trueN), nil, false},
		{"andErrorFalse", errorN.And(falseN), nil, false},
		{"andErrorValue", errorN.And(valueN), nil, false},
		{"andErrorKeep", errorN.And(keepN), nil, false},
		{"andErrorError", errorN.And(errorN), nil, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeAnd)
			testutil.Equals(t, ok, true)
			out, ok := partialAnd(&Context{}, n)
			testutil.Equals(t, ok, tt.ok)
			if !tt.ok && out == nil && tt.out == nil {
				return
			}
			nd, ok := tt.out.(ast.Node)
			testutil.Equals(t, ok, true)
			testutil.Equals(t, out, nd.AsIsNode())
		})
	}
}

func TestPartialOr(t *testing.T) {
	errorN := ast.Long(42).GreaterThan(ast.String("bananas"))
	trueN := ast.True()
	falseN := ast.False()
	valueN := ast.String("test")
	keepN := ast.Context()
	_, _, _, _, _ = errorN, trueN, falseN, valueN, keepN

	tests := []struct {
		name string
		in   ast.Node
		out  any
		ok   bool
	}{

		{"orTrueTrue", trueN.Or(trueN), ast.True(), true},
		{"orTrueFalse", trueN.Or(falseN), ast.True(), true},
		{"orTrueValue", trueN.Or(valueN), ast.True(), true},
		{"orTrueKeep", trueN.Or(keepN), ast.True(), true},
		{"orTrueError", trueN.Or(errorN), ast.True(), true},

		{"orFalseTrue", falseN.Or(trueN), ast.True(), true},
		{"orFalseFalse", falseN.Or(falseN), ast.False(), true},
		{"orFalseValue", falseN.Or(valueN), nil, false},
		{"orFalseKeep", falseN.Or(keepN), falseN.Or(keepN), true},
		{"orFalseError", falseN.Or(errorN), nil, false},

		{"orValueTrue", valueN.Or(trueN), nil, false},
		{"orValueFalse", valueN.Or(falseN), nil, false},
		{"orValueValue", valueN.Or(valueN), nil, false},
		{"orValueKeep", valueN.Or(keepN), nil, false},
		{"orValueError", valueN.Or(errorN), nil, false},

		{"orKeepTrue", keepN.Or(trueN), keepN.Or(trueN), true},
		{"orKeepFalse", keepN.Or(falseN), keepN.Or(falseN), true},
		{"orKeepValue", keepN.Or(valueN), keepN.Or(valueN), true},
		{"orKeepKeep", keepN.Or(keepN), keepN.Or(keepN), true},
		{"orKeepError", keepN.Or(errorN), keepN.Or(ast.ExtensionCall("error")), true},

		{"orErrorTrue", errorN.Or(trueN), nil, false},
		{"orErrorFalse", errorN.Or(falseN), nil, false},
		{"orErrorValue", errorN.Or(valueN), nil, false},
		{"orErrorKeep", errorN.Or(keepN), nil, false},
		{"orErrorError", errorN.Or(errorN), nil, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeOr)
			testutil.Equals(t, ok, true)
			out, ok := partialOr(&Context{}, n)
			testutil.Equals(t, ok, tt.ok)
			if !tt.ok && out == nil && tt.out == nil {
				return
			}
			nd, ok := tt.out.(ast.Node)
			testutil.Equals(t, ok, true)
			testutil.Equals(t, out, nd.AsIsNode())
		})
	}
}
