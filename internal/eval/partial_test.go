package eval

import (
	"errors"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestPartialScopeEval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		env    Env
		ent    types.Value
		in     ast.IsScopeNode
		evaled bool
		result bool
	}{
		{
			"notEntity",
			Env{},
			types.String("test"),
			ast.ScopeTypeAll{},
			false, false,
		},

		{
			"isVariable",
			Env{},
			Variable("principal"),
			ast.ScopeTypeEq{},
			false, false,
		},

		{
			"isIgnore",
			Env{},
			Ignore(),
			ast.ScopeTypeEq{},
			true, true,
		},

		{
			"scopeTypeAll",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeAll{},
			true, true,
		},

		{
			"scopeTypeEq",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeEq{Entity: types.NewEntityUID("T", "1")},
			true, true,
		},
		{
			"scopeTypeEq/fail",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeEq{Entity: types.NewEntityUID("FAIL", "1")},
			true, false,
		},

		{
			"scopeTypeIn",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeIn{Entity: types.NewEntityUID("T", "1")},
			true, true,
		},
		{
			"scopeTypeIn/fail",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeIn{Entity: types.NewEntityUID("T", "FAIL")},
			true, false,
		},

		{
			"scopeTypeInSet",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeInSet{Entities: []types.EntityUID{types.NewEntityUID("T", "1")}},
			true, true,
		},

		{
			"scopeTypeInSet/fail",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeInSet{Entities: []types.EntityUID{types.NewEntityUID("T", "FAIL")}},
			true, false,
		},

		{
			"scopeTypeIs",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeIs{Type: "T"},
			true, true,
		},

		{
			"scopeTypeIs/fail",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeIs{Type: "FAIL"},
			true, false,
		},

		{
			"scopeTypeIsIn",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeIsIn{Type: "T", Entity: types.NewEntityUID("T", "1")},
			true, true,
		},
		{
			"scopeTypeIsIn/failIs",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeIsIn{Type: "FAIL", Entity: types.NewEntityUID("T", "1")},
			true, false,
		},
		{
			"scopeTypeIsIn/failIn",
			Env{},
			types.NewEntityUID("T", "1"),
			ast.ScopeTypeIsIn{Type: "T", Entity: types.NewEntityUID("T", "FAIL")},
			true, false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.env.Entities = types.EntityMap{}
			evaled, result := partialScopeEval(tt.env, tt.ent, tt.in)
			testutil.Equals(t, evaled, tt.evaled)
			testutil.Equals(t, result, tt.result)
		})
	}

}

func TestPartialScopeEvalPanic(t *testing.T) {
	t.Parallel()
	testutil.Panic(t, func() {
		partialScopeEval(Env{}, types.NewEntityUID("T", "1"), nil)
	})
}

func TestPartialPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		env  Env
		out  *ast.Policy
		keep bool
		out2 ast.Node
	}{
		{"smokeTest",
			ast.Permit(),
			Env{},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"principalEqual",
			ast.Permit().PrincipalEq(types.NewEntityUID("Account", "42")),
			Env{
				Principal: types.NewEntityUID("Account", "42"),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"principalNotEqual",
			ast.Permit().PrincipalEq(types.NewEntityUID("Account", "42")),
			Env{
				Principal: types.NewEntityUID("Account", "Other"),
			},
			nil,
			false,
			ast.False(),
		},
		{"actionEqual",
			ast.Permit().ActionEq(types.NewEntityUID("Action", "42")),
			Env{
				Action: types.NewEntityUID("Action", "42"),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"actionNotEqual",
			ast.Permit().ActionEq(types.NewEntityUID("Action", "42")),
			Env{
				Action: types.NewEntityUID("Action", "Other"),
			},
			nil,
			false,
			ast.False(),
		},
		{"resourceEqual",
			ast.Permit().ResourceEq(types.NewEntityUID("Resource", "42")),
			Env{
				Resource: types.NewEntityUID("Resource", "42"),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"resourceNotEqual",
			ast.Permit().ResourceEq(types.NewEntityUID("Resource", "42")),
			Env{
				Resource: types.NewEntityUID("Resource", "Other"),
			},
			nil,
			false,
			ast.False(),
		},
		{"conditionOmitTrue",
			ast.Permit().When(ast.True()),
			Env{},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"conditionDropFalse",
			ast.Permit().When(ast.False()),
			Env{},
			nil,
			false,
			ast.False(),
		},
		{"conditionDropError",
			ast.Permit().When(ast.Long(42).GreaterThan(ast.String("bananas"))),
			Env{},
			ast.Permit().When(ast.NewNode(extError(errors.New("type error: expected comparable value, got string")))),
			true,
			ast.True().And(ast.NewNode(extError(errors.New("type error: expected comparable value, got string")))),
		},
		{"conditionDropTypeError",
			ast.Permit().When(ast.Long(42)),
			Env{},
			ast.Permit().When(ast.NewNode(extError(errors.New("type error: condition expected bool")))),
			true,
			ast.True().And(ast.NewNode(extError(errors.New("type error: condition expected bool")))),
		},
		{"conditionKeepUnfolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			Env{Context: Variable("context")},
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			true,
			ast.True().And(ast.Context().GreaterThan(ast.Long(42))),
		},
		{"conditionOmitTrueFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			Env{
				Context: types.Long(43),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"conditionDropFalseFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			Env{
				Context: types.Long(41),
			},
			nil,
			false,
			ast.False(),
		},
		{"conditionDropErrorFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			Env{
				Context: types.String("bananas"),
			},
			ast.Permit().When(ast.NewNode(extError(errors.New("type error: expected comparable value, got string")))),
			true,
			ast.True().And(ast.NewNode(extError(errors.New("type error: expected comparable value, got string")))),
		},
		{"contextVariableAccess",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"key": Variable("var"),
				}),
			},
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			true,
			ast.True().And(ast.Context().Access("key").Equal(ast.Long(42))),
		},

		{"ignorePermitContext",
			ast.Permit().When(ast.Context().Equal(ast.Long(42))),
			Env{
				Context: Ignore(),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"ignoreForbidContext",
			ast.Forbid().When(ast.Context().Equal(ast.Long(42))),
			Env{
				Context: Ignore(),
			},
			nil,
			false,
			ast.False(),
		},
		{"ignorePermitScope",
			ast.Permit().PrincipalEq(types.NewEntityUID("T", "42")),
			Env{
				Principal: Ignore(),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"ignoreForbidScope",
			ast.Forbid().PrincipalEq(types.NewEntityUID("T", "42")),
			Env{
				Principal: Ignore(),
			},
			ast.Forbid(),
			true,
			ast.True(),
		},
		{"ignoreAnd",
			ast.Permit().When(ast.Context().Access("variable").And(ast.Context().Access("ignore").Equal(ast.Long(42)))),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				}),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"ignoreOr",
			ast.Permit().When(ast.Context().Access("variable").Or(ast.Context().Access("ignore").Equal(ast.Long(42)))),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				}),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"ignoreIfThen",
			ast.Permit().When(ast.IfThenElse(ast.Context().Access("variable"), ast.Context().Access("ignore").Equal(ast.Long(42)), ast.True())),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				}),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"ignoreIfElse",
			ast.Permit().When(ast.IfThenElse(ast.Context().Access("variable"), ast.True(), ast.Context().Access("ignore").Equal(ast.Long(42)))),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				}),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"ignoreHas",
			ast.Permit().When(ast.Context().Has("ignore")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				}),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"ignoreHasNot",
			ast.Permit().When(ast.Not(ast.Context().Has("ignore"))),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				}),
			},
			ast.Permit(),
			true,
			ast.True(),
		},
		{"errorShortCircuit",
			ast.Permit().When(ast.True()).When(ast.String("test").LessThan(ast.Long(42))).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.NewNode(extError(errors.New("type error: expected comparable value, got string")))),
			true,
			ast.True().And(ast.NewNode(extError(errors.New("type error: expected comparable value, got string")))),
		},
		{"errorShortCircuitKept",
			ast.Permit().When(ast.Context().Access("variable")).When(ast.String("test").LessThan(ast.Long(42))).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.Context().Access("variable")).When(ast.NewNode(extError(errors.New("type error: expected comparable value, got string")))),
			true,
			ast.True().And(ast.Context().Access("variable").And(ast.NewNode(extError(errors.New("type error: expected comparable value, got string"))))),
		},
		{"errorConditionShortCircuit",
			ast.Permit().When(ast.True()).When(ast.String("test")).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.NewNode(extError(errors.New("type error: condition expected bool")))),
			true,
			ast.True().And(ast.NewNode(extError(errors.New("type error: condition expected bool")))),
		},
		{"errorConditionShortCircuitKept",
			ast.Permit().When(ast.Context().Access("variable")).When(ast.String("test")).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.Context().Access("variable")).When(ast.NewNode(extError(errors.New("type error: condition expected bool")))),
			true,
			ast.True().And(ast.Context().Access("variable").And(ast.NewNode(extError(errors.New("type error: condition expected bool"))))),
		},
		{"errorConditionShortCircuitKeptDeeper",
			ast.Permit().When(ast.Context().Access("variable")).When(ast.String("test")).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.Context().Access("variable")).When(ast.NewNode(extError(errors.New("type error: condition expected bool")))),
			true,
			ast.True().And(ast.Context().Access("variable").And(ast.NewNode(extError(errors.New("type error: condition expected bool"))))),
		},
		{"keepDeepVariables",
			ast.Permit().When(ast.True().Equal(ast.False().Equal(ast.Context()))),
			Env{
				Context: Variable("context"),
			},
			ast.Permit().When(ast.True().Equal(ast.False().Equal(ast.Context()))),
			true,
			ast.True().And(ast.True().Equal(ast.False().Equal(ast.Context()))),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, keep := PartialPolicy(tt.env, tt.in)
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
			out2, keep2 := PartialPolicyToNode(tt.env, tt.in)
			if keep2 {
				testutil.Equals(t, out2, tt.out2)
			} else {
				testutil.Equals(t, out2, ast.False())
			}
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
		name    string
		in      ast.Node
		out     any
		errTest func(testutil.TB, error)
	}{
		{"ifTrueAB", ast.IfThenElse(trueN, valueA, valueB), valueA, testutil.OK},
		{"ifFalseAB", ast.IfThenElse(falseN, valueA, valueB), valueB, testutil.OK},
		{"ifValueAB", ast.IfThenElse(valueN, valueA, valueB), nil, testutil.Error},
		{"ifKeepAB", ast.IfThenElse(keepN, valueA, valueB), ast.IfThenElse(keepN, valueA, valueB), testutil.OK},
		{"ifErrorAB", ast.IfThenElse(errorN, valueA, valueB), nil, testutil.Error},

		{"ifTrueErrorB", ast.IfThenElse(trueN, errorN, valueB), nil, testutil.Error},
		{"ifFalseAError", ast.IfThenElse(falseN, valueA, errorN), nil, testutil.Error},
		{"ifTrueAError", ast.IfThenElse(trueN, valueA, errorN), valueA, testutil.OK},
		{"ifFalseErrorB", ast.IfThenElse(falseN, errorN, valueB), valueB, testutil.OK},

		{"ifKeepKeepKeep", ast.IfThenElse(keepN, keepN, keepN), ast.IfThenElse(keepN, keepN, keepN), testutil.OK},
		{"ifKeepErrorError", ast.IfThenElse(keepN, errorN, errorN), ast.IfThenElse(keepN, ast.ExtensionCall(partialErrorName, ast.String("type error: expected comparable value, got string")), ast.ExtensionCall(partialErrorName, ast.String("type error: expected comparable value, got string"))), testutil.OK},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeIfThenElse)
			testutil.Equals(t, ok, true)
			out, err := partialIfThenElse(Env{
				Context: Variable("context"),
			}, n)
			tt.errTest(t, err)
			if err != nil {
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
		name    string
		in      ast.Node
		out     any
		errTest func(testutil.TB, error)
	}{

		{"andTrueTrue", trueN.And(trueN), ast.True(), testutil.OK},
		{"andTrueFalse", trueN.And(falseN), ast.False(), testutil.OK},
		{"andTrueValue", trueN.And(valueN), nil, testutil.Error},
		{"andTrueKeep", trueN.And(keepN), trueN.And(keepN), testutil.OK},
		{"andTrueError", trueN.And(errorN), nil, testutil.Error},

		{"andFalseTrue", falseN.And(trueN), ast.False(), testutil.OK},
		{"andFalseFalse", falseN.And(falseN), ast.False(), testutil.OK},
		{"andFalseValue", falseN.And(valueN), ast.False(), testutil.OK},
		{"andFalseKeep", falseN.And(keepN), ast.False(), testutil.OK},
		{"andFalseError", falseN.And(errorN), ast.False(), testutil.OK},

		{"andValueTrue", valueN.And(trueN), nil, testutil.Error},
		{"andValueFalse", valueN.And(falseN), nil, testutil.Error},
		{"andValueValue", valueN.And(valueN), nil, testutil.Error},
		{"andValueKeep", valueN.And(keepN), nil, testutil.Error},
		{"andValueError", valueN.And(errorN), nil, testutil.Error},

		{"andKeepTrue", keepN.And(trueN), keepN.And(trueN), testutil.OK},
		{"andKeepFalse", keepN.And(falseN), keepN.And(falseN), testutil.OK},
		{"andKeepValue", keepN.And(valueN), keepN.And(valueN), testutil.OK},
		{"andKeepKeep", keepN.And(keepN), keepN.And(keepN), testutil.OK},
		{"andKeepError", keepN.And(errorN), keepN.And(ast.ExtensionCall(partialErrorName, ast.String("type error: expected comparable value, got string"))), testutil.OK},

		{"andErrorTrue", errorN.And(trueN), nil, testutil.Error},
		{"andErrorFalse", errorN.And(falseN), nil, testutil.Error},
		{"andErrorValue", errorN.And(valueN), nil, testutil.Error},
		{"andErrorKeep", errorN.And(keepN), nil, testutil.Error},
		{"andErrorError", errorN.And(errorN), nil, testutil.Error},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeAnd)
			testutil.Equals(t, ok, true)
			out, err := partialAnd(Env{
				Context: Variable("context"),
			}, n)
			tt.errTest(t, err)
			if err != nil {
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
		name    string
		in      ast.Node
		out     any
		errTest func(testutil.TB, error)
	}{

		{"orTrueTrue", trueN.Or(trueN), ast.True(), testutil.OK},
		{"orTrueFalse", trueN.Or(falseN), ast.True(), testutil.OK},
		{"orTrueValue", trueN.Or(valueN), ast.True(), testutil.OK},
		{"orTrueKeep", trueN.Or(keepN), ast.True(), testutil.OK},
		{"orTrueError", trueN.Or(errorN), ast.True(), testutil.OK},

		{"orFalseTrue", falseN.Or(trueN), ast.True(), testutil.OK},
		{"orFalseFalse", falseN.Or(falseN), ast.False(), testutil.OK},
		{"orFalseValue", falseN.Or(valueN), nil, testutil.Error},
		{"orFalseKeep", falseN.Or(keepN), falseN.Or(keepN), testutil.OK},
		{"orFalseError", falseN.Or(errorN), nil, testutil.Error},

		{"orValueTrue", valueN.Or(trueN), nil, testutil.Error},
		{"orValueFalse", valueN.Or(falseN), nil, testutil.Error},
		{"orValueValue", valueN.Or(valueN), nil, testutil.Error},
		{"orValueKeep", valueN.Or(keepN), nil, testutil.Error},
		{"orValueError", valueN.Or(errorN), nil, testutil.Error},

		{"orKeepTrue", keepN.Or(trueN), keepN.Or(trueN), testutil.OK},
		{"orKeepFalse", keepN.Or(falseN), keepN.Or(falseN), testutil.OK},
		{"orKeepValue", keepN.Or(valueN), keepN.Or(valueN), testutil.OK},
		{"orKeepKeep", keepN.Or(keepN), keepN.Or(keepN), testutil.OK},
		{"orKeepError", keepN.Or(errorN), keepN.Or(ast.ExtensionCall(partialErrorName, ast.String("type error: expected comparable value, got string"))), testutil.OK},

		{"orErrorTrue", errorN.Or(trueN), nil, testutil.Error},
		{"orErrorFalse", errorN.Or(falseN), nil, testutil.Error},
		{"orErrorValue", errorN.Or(valueN), nil, testutil.Error},
		{"orErrorKeep", errorN.Or(keepN), nil, testutil.Error},
		{"orErrorError", errorN.Or(errorN), nil, testutil.Error},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeOr)
			testutil.Equals(t, ok, true)
			out, err := partialOr(Env{
				Context: Variable("context"),
			}, n)
			tt.errTest(t, err)
			if err != nil {
				return
			}
			nd, ok := tt.out.(ast.Node)
			testutil.Equals(t, ok, true)
			testutil.Equals(t, out, nd.AsIsNode())
		})
	}
}

func errNode(msg string) ast.Node {
	return ast.NewNode(extError(errors.New(msg)))
}

func errorIs(want error) func(testutil.TB, error) {
	return func(t testutil.TB, got error) {
		testutil.ErrorIs(t, got, want)
	}
}

func TestPartialBasic(t *testing.T) {
	t.Parallel()
	nul := ast.NewNode(nil)
	tests := []struct {
		name string
		in   ast.Node
		out  ast.Node
		err  func(testutil.TB, error)
	}{
		{
			"variablePrincipalKeep",
			ast.Principal(),
			ast.Principal(),
			errorIs(errVariable),
		},
		{
			"variableActionKeep",
			ast.Action(),
			ast.Action(),
			errorIs(errVariable),
		},
		{
			"variableResourceKeep",
			ast.Resource(),
			ast.Resource(),
			errorIs(errVariable),
		},
		{
			"variableContextKeep",
			ast.Context(),
			ast.Context(),
			errorIs(errVariable),
		},
		{
			"valueTrueKeep",
			ast.True(),
			ast.True(),
			testutil.OK,
		},
		{
			"valueFalseKeep",
			ast.False(),
			ast.False(),
			testutil.OK,
		},
		{
			"valueStringKeep",
			ast.String("cedar"),
			ast.String("cedar"),
			testutil.OK,
		},
		{
			"valueLongKeep",
			ast.Long(42),
			ast.Long(42),
			testutil.OK,
		},
		{
			"valueSetNodesKeep",
			ast.Set(ast.Long(42), ast.Long(43), ast.Context()),
			ast.Set(ast.Long(42), ast.Long(43), ast.Context()),
			testutil.OK,
		},
		{
			"valueSetNodesFold",
			ast.Set(ast.Long(42), ast.Long(43)),
			ast.Value(types.NewSet(types.Long(42), types.Long(43))),
			testutil.OK,
		},
		{
			"valueSetNodesError",
			ast.Set(errNode("err")),
			nul,
			testutil.Error,
		},
		{
			"valueRecordElementsKeep",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Context()}}),
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Context()}}),
			testutil.OK,
		},
		{
			"valueRecordElementsFold",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(42)}}),
			ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)})),
			testutil.OK,
		},
		{
			"valueRecordElementsError",
			ast.Record(ast.Pairs{{Key: "key", Value: errNode("err")}}),
			nul,
			testutil.Error,
		},
		{
			"valueEntityUIDKeep",
			ast.EntityUID("T", "42"),
			ast.EntityUID("T", "42"),
			testutil.OK,
		},
		{
			"valueIPAddrKeep",
			ast.IPAddr(netip.MustParsePrefix("127.0.0.1/16")),
			ast.IPAddr(netip.MustParsePrefix("127.0.0.1/16")),
			testutil.OK,
		},
		{
			"opEqualsKeep",
			ast.Long(42).Equal(ast.Context()),
			ast.Long(42).Equal(ast.Context()),
			testutil.OK,
		},
		{
			"opEqualsFold",
			ast.Long(42).Equal(ast.Long(43)),
			ast.False(),
			testutil.OK,
		},
		{
			"opEqualsError",
			ast.Long(42).Equal(errNode("err")),
			nul,
			testutil.Error,
		},
		{
			"opNotEqualsKeep",
			ast.Long(42).NotEqual(ast.Context()),
			ast.Long(42).NotEqual(ast.Context()),
			testutil.OK,
		},
		{
			"opNotEqualsFold",
			ast.Long(42).NotEqual(ast.Long(43)),
			ast.True(),
			testutil.OK,
		},
		{
			"opNotEqualsError",
			ast.Long(42).NotEqual(errNode("err")),
			nul,
			testutil.Error,
		},
		{
			"opLessThanKeep",
			ast.Long(42).LessThan(ast.Context()),
			ast.Long(42).LessThan(ast.Context()),
			testutil.OK,
		},
		{
			"opLessThanFold",
			ast.Long(42).LessThan(ast.Long(43)),
			ast.True(),
			testutil.OK,
		},
		{
			"opLessThanComparableKeep",
			ast.Datetime(time.UnixMilli(42)).LessThan(ast.Context()),
			ast.Datetime(time.UnixMilli(42)).LessThan(ast.Context()),
			testutil.OK,
		},
		{
			"opLessThanComparableFold",
			ast.Datetime(time.UnixMilli(42)).LessThan(ast.Datetime(time.UnixMilli(43))),
			ast.True(),
			testutil.OK,
		},
		{
			"opLessThanError",
			ast.Long(42).LessThan(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opLessThanOrEqualKeep",
			ast.Long(42).LessThanOrEqual(ast.Context()),
			ast.Long(42).LessThanOrEqual(ast.Context()),
			testutil.OK,
		},
		{
			"opLessThanOrEqualFold",
			ast.Long(42).LessThanOrEqual(ast.Long(43)),
			ast.True(),
			testutil.OK,
		},
		{
			"opLessThanOrEqualComparableKeep",
			ast.Datetime(time.UnixMilli(42)).LessThanOrEqual(ast.Context()),
			ast.Datetime(time.UnixMilli(42)).LessThanOrEqual(ast.Context()),
			testutil.OK,
		},
		{
			"opLessThanOrEqualComparableFold",
			ast.Datetime(time.UnixMilli(42)).LessThanOrEqual(ast.Datetime(time.UnixMilli(43))),
			ast.True(),
			testutil.OK,
		},
		{
			"opLessThanOrEqualError",
			ast.Long(42).LessThanOrEqual(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opGreaterThanKeep",
			ast.Long(42).GreaterThan(ast.Context()),
			ast.Long(42).GreaterThan(ast.Context()),
			testutil.OK,
		},
		{
			"opGreaterThanFold",
			ast.Long(42).GreaterThan(ast.Long(43)),
			ast.False(),
			testutil.OK,
		},
		{
			"opGreaterThanComparableKeep",
			ast.Datetime(time.UnixMilli(42)).GreaterThan(ast.Context()),
			ast.Datetime(time.UnixMilli(42)).GreaterThan(ast.Context()),
			testutil.OK,
		},
		{
			"opGreaterThanComparableFold",
			ast.Datetime(time.UnixMilli(42)).GreaterThan(ast.Datetime(time.UnixMilli(43))),
			ast.False(),
			testutil.OK,
		},
		{
			"opGreaterThanError",
			ast.Long(42).GreaterThan(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opGreaterThanOrEqualKeep",
			ast.Long(42).GreaterThanOrEqual(ast.Context()),
			ast.Long(42).GreaterThanOrEqual(ast.Context()),
			testutil.OK,
		},
		{
			"opGreaterThanOrEqualFold",
			ast.Long(42).GreaterThanOrEqual(ast.Long(43)),
			ast.False(),
			testutil.OK,
		},
		{
			"opGreaterThanOrEqualComparableKeep",
			ast.Datetime(time.UnixMilli(42)).GreaterThanOrEqual(ast.Context()),
			ast.Datetime(time.UnixMilli(42)).GreaterThanOrEqual(ast.Context()),
			testutil.OK,
		},
		{
			"opGreaterThanOrEqualComparableFold",
			ast.Datetime(time.UnixMilli(42)).GreaterThanOrEqual(ast.Datetime(time.UnixMilli(43))),
			ast.False(),
			testutil.OK,
		},
		{
			"opGreaterThanOrEqualError",
			ast.Long(42).GreaterThanOrEqual(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opLessThanExtKeep",
			ast.Value(testutil.Must(types.NewDecimalFromInt(42))).DecimalLessThan(ast.Context()),
			ast.Value(testutil.Must(types.NewDecimalFromInt(42))).DecimalLessThan(ast.Context()),
			testutil.OK,
		},
		{
			"opLessThanExtFold",
			ast.Value(testutil.Must(types.NewDecimalFromInt(42))).DecimalLessThan(ast.Value(testutil.Must(types.NewDecimalFromInt(43)))),
			ast.True(),
			testutil.OK,
		},
		{
			"opLessThanExtError",
			ast.Value(testutil.Must(types.NewDecimalFromInt(42))).DecimalLessThan(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opLikeKeep",
			ast.Context().Like(types.NewPattern(types.Wildcard{})),
			ast.Context().Like(types.NewPattern(types.Wildcard{})),
			testutil.OK,
		},
		{
			"opLikeFold",
			ast.String("test").Like(types.NewPattern(types.Wildcard{})),
			ast.True(),
			testutil.OK,
		},
		{
			"opLikeError",
			ast.Long(42).Like(types.NewPattern(types.Wildcard{})),
			nul,
			testutil.Error,
		},
		{
			"opAndKeep",
			ast.True().And(ast.Context()),
			ast.True().And(ast.Context()),
			testutil.OK,
		},
		{
			"opAndFold",
			ast.True().And(ast.True()),
			ast.True(),
			testutil.OK,
		},
		{
			"opAndError",
			ast.True().And(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opOrKeep",
			ast.False().Or(ast.Context()),
			ast.False().Or(ast.Context()),
			testutil.OK,
		},
		{
			"opOrFold",
			ast.False().Or(ast.True()),
			ast.True(),
			testutil.OK,
		},
		{
			"opOrError",
			ast.False().Or(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opNotKeep",
			ast.Not(ast.Context()),
			ast.Not(ast.Context()),
			testutil.OK,
		},
		{
			"opNotFold",
			ast.Not(ast.True()),
			ast.False(),
			testutil.OK,
		},
		{
			"opNotError",
			ast.Not(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opIfKeep",
			ast.IfThenElse(ast.Context(), ast.Long(42), ast.Long(43)),
			ast.IfThenElse(ast.Context(), ast.Long(42), ast.Long(43)),
			testutil.OK,
		},
		{
			"opIfFold",
			ast.IfThenElse(ast.True(), ast.Long(42), ast.Long(43)),
			ast.Long(42),
			testutil.OK,
		},
		{
			"opIfError",
			ast.IfThenElse(ast.String("test"), ast.Long(42), ast.Long(43)),
			nul,
			testutil.Error,
		},
		{
			"opPlusKeep",
			ast.Long(42).Add(ast.Context()),
			ast.Long(42).Add(ast.Context()),
			testutil.OK,
		},
		{
			"opPlusFold",
			ast.Long(42).Add(ast.Long(43)),
			ast.Long(85),
			testutil.OK,
		},
		{
			"opPlusError",
			ast.Long(42).Add(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opMinusKeep",
			ast.Long(42).Subtract(ast.Context()),
			ast.Long(42).Subtract(ast.Context()),
			testutil.OK,
		},
		{
			"opMinusFold",
			ast.Long(42).Subtract(ast.Long(43)),
			ast.Long(-1),
			testutil.OK,
		},
		{
			"opMinusError",
			ast.Long(42).Subtract(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opTimesKeep",
			ast.Long(42).Multiply(ast.Context()),
			ast.Long(42).Multiply(ast.Context()),
			testutil.OK,
		},
		{
			"opTimesFold",
			ast.Long(42).Multiply(ast.Long(43)),
			ast.Long(1806),
			testutil.OK,
		},
		{
			"opTimesError",
			ast.Long(42).Multiply(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opNegateKeep",
			ast.Negate(ast.Context()),
			ast.Negate(ast.Context()),
			testutil.OK,
		},
		{
			"opNegateFold",
			ast.Negate(ast.Long(42)),
			ast.Long(-42),
			testutil.OK,
		},
		{
			"opNegateError",
			ast.Negate(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opInKeep",
			ast.EntityUID("T", "1").In(ast.Context()),
			ast.EntityUID("T", "1").In(ast.Context()),
			testutil.OK,
		},
		{
			"opInFold",
			ast.EntityUID("T", "1").In(ast.EntityUID("T", "1")),
			ast.True(),
			testutil.OK,
		},
		{
			"opInError",
			ast.EntityUID("T", "1").In(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opIsKeep",
			ast.Principal().Is(types.EntityType("T")),
			ast.Principal().Is(types.EntityType("T")),
			testutil.OK,
		},
		{
			"opIsFold",
			ast.EntityUID("T", "1").Is(types.EntityType("T")),
			ast.True(),
			testutil.OK,
		},
		{
			"opIsError",
			ast.String("test").Is(types.EntityType("T")),
			nul,
			testutil.Error,
		},
		{
			"opIsInKeep",
			ast.Principal().IsIn(types.EntityType("T"), ast.Context()),
			ast.Principal().IsIn(types.EntityType("T"), ast.Context()),
			testutil.OK,
		},
		{
			"opIsInFold",
			ast.EntityUID("T", "1").IsIn(types.EntityType("T"), ast.EntityUID("T", "1")),
			ast.True(),
			testutil.OK,
		},
		{
			"opIsInError",
			ast.EntityUID("T", "1").IsIn(types.EntityType("T"), ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opContainsKeep",
			ast.Set(ast.Long(42)).Contains(ast.Context()),
			ast.Value(types.NewSet(types.Long(42))).Contains(ast.Context()),
			testutil.OK,
		},
		{
			"opContainsFold",
			ast.Set(ast.Long(42)).Contains(ast.Long(43)),
			ast.False(),
			testutil.OK,
		},
		{
			"opContainsError",
			ast.String("test").Contains(ast.Long(43)),
			nul,
			testutil.Error,
		},
		{
			"opContainsAllKeep",
			ast.Set(ast.Long(42)).ContainsAll(ast.Context()),
			ast.Value(types.NewSet(types.Long(42))).ContainsAll(ast.Context()),
			testutil.OK,
		},
		{
			"opContainsAllFold",
			ast.Set(ast.Long(42)).ContainsAll(ast.Set(ast.Long(43))),
			ast.False(),
			testutil.OK,
		},
		{
			"opContainsAllError",
			ast.Set(ast.Long(42)).ContainsAll(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opContainsAnyKeep",
			ast.Set(ast.Long(42)).ContainsAny(ast.Context()),
			ast.Value(types.NewSet(types.Long(42))).ContainsAny(ast.Context()),
			testutil.OK,
		},
		{
			"opContainsAnyFold",
			ast.Set(ast.Long(42)).ContainsAny(ast.Set(ast.Long(43))),
			ast.False(),
			testutil.OK,
		},
		{
			"opContainsAnyError",
			ast.Set(ast.Long(42)).ContainsAny(ast.String("test")),
			nul,
			testutil.Error,
		},
		{
			"opIsEmptyKeep",
			ast.Set(ast.Context()).IsEmpty(),
			ast.Set(ast.Context()).IsEmpty(),
			testutil.OK,
		},
		{
			"opIsEmptyFold",
			ast.Set(ast.Long(42)).IsEmpty(),
			ast.False(),
			testutil.OK,
		},
		{
			"opIsEmptyError",
			ast.String("test").IsEmpty(),
			nul,
			testutil.Error,
		},
		{
			"opAccessKeep",
			ast.Context().Access("key"),
			ast.Context().Access("key"),
			testutil.OK,
		},
		{
			"opAccessFold",
			ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)})).Access("key"),
			ast.Long(42),
			testutil.OK,
		},
		{
			"opAccessError",
			ast.String("test").Access("key"),
			nul,
			testutil.Error,
		},
		{
			"opHasKeep",
			ast.Context().Has("key"),
			ast.Context().Has("key"),
			testutil.OK,
		},
		{
			"opHasFold",
			ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)})).Has("key"),
			ast.True(),
			testutil.OK,
		},
		{
			"opHasError",
			ast.String("test").Has("key"),
			nul,
			testutil.Error,
		},
		{
			"opGetTagKeep",
			ast.Principal().GetTag(ast.String("key")),
			ast.Principal().GetTag(ast.String("key")),
			testutil.OK,
		},
		{
			"opGetTagError",
			ast.String("test").GetTag(ast.String("key")),
			nul,
			testutil.Error,
		},
		{
			"opHasTagKeep",
			ast.Principal().HasTag(ast.String("key")),
			ast.Principal().HasTag(ast.String("key")),
			testutil.OK,
		},
		{
			"opHasTagError",
			ast.String("test").HasTag(ast.String("key")),
			nul,
			testutil.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := partial((Env{
				Principal: Variable("principal"),
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context:   Variable("context"),
			}), tt.in.AsIsNode())
			tt.err(t, err)
			testutil.Equals(t, out, tt.out.AsIsNode())
		})
	}
}

func TestPartialWithContext(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		context  types.Value
		entities types.EntityGetter
		in       ast.Node
		out      ast.Node
		err      func(testutil.TB, error)
	}{
		{
			"opGetTagFold",
			types.NewRecord(types.RecordMap{
				"entity": types.NewEntityUID("T", "1"),
			}),
			types.EntityMap{types.NewEntityUID("T", "1"): types.Entity{
				Tags: types.NewRecord(types.RecordMap{
					"a": types.Long(42),
				}),
			}},
			ast.Context().Access("entity").GetTag(ast.String("a")),
			ast.Long(42),
			testutil.OK,
		},
		{
			"opHasTagFold",
			types.NewRecord(types.RecordMap{
				"entity": types.NewEntityUID("T", "1"),
			}),
			types.EntityMap{types.NewEntityUID("T", "1"): types.Entity{
				Tags: types.NewRecord(types.RecordMap{
					"a": types.Long(42),
				}),
			}},
			ast.Context().Access("entity").HasTag(ast.String("a")),
			ast.True(),
			testutil.OK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := partial((Env{
				Entities:  tt.entities,
				Principal: Variable("principal"),
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context:   tt.context,
			}), tt.in.AsIsNode())
			tt.err(t, err)
			testutil.Equals(t, out, tt.out.AsIsNode())
		})
	}
}

func TestPartialPanic(t *testing.T) {
	t.Parallel()
	testutil.Panic(t, func() {
		_, _ = partial(Env{}, nil)
	})
}

func TestPartialErrorEval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		env  Env
		in   Evaler
		out  types.Value
		err  func(testutil.TB, error)
	}{
		{"happy",
			Env{},
			newPartialErrorEval(newLiteralEval(types.String("err"))),
			nil, testutil.Error,
		},

		{"err",
			Env{},
			newPartialErrorEval(newLiteralEval(types.Long(42))),
			nil, testutil.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := tt.in.Eval(tt.env)
			testutil.Equals(t, out, tt.out)
			tt.err(t, err)
		})
	}
}

func TestPartialHasEval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		env  Env
		in   Evaler
		out  types.Value
		err  func(testutil.TB, error)
	}{
		{"happy",
			Env{},
			newPartialHasEval(newLiteralEval(types.NewRecord(types.RecordMap{"key": types.Long(42)})), "key"),
			types.True, testutil.OK,
		},
		{"badArg",
			Env{},
			newPartialHasEval(newErrorEval(fmt.Errorf("err")), "key"),
			nil, testutil.Error,
		},
		{"badType",
			Env{},
			newPartialHasEval(newLiteralEval(types.String("test")), "key"),
			nil, testutil.Error,
		},
		{"entity",
			Env{Entities: types.EntityMap{
				types.NewEntityUID("T", "1"): types.Entity{
					UID:        types.NewEntityUID("T", "1"),
					Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
				},
			}},
			newPartialHasEval(newLiteralEval(types.NewEntityUID("T", "1")), "key"),
			types.True, testutil.OK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := tt.in.Eval(tt.env)
			testutil.Equals(t, out, tt.out)
			tt.err(t, err)
		})
	}
}

func TestIsTrue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   ast.IsNode
		out  bool
	}{
		{"happy", ast.True().AsIsNode(), true},
		{"false", ast.False().AsIsNode(), false},
		{"notBoolean", ast.String("test").AsIsNode(), false},
		{"notValue", ast.Context().AsIsNode(), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := isTrue(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestIsFalse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   ast.IsNode
		out  bool
	}{
		{"happy", ast.False().AsIsNode(), true},
		{"true", ast.True().AsIsNode(), false},
		{"notBoolean", ast.String("test").AsIsNode(), false},
		{"notValue", ast.Context().AsIsNode(), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := isFalse(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestIsNonBoolValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   ast.IsNode
		out  bool
	}{
		{"happy", ast.String("test").AsIsNode(), true},
		{"true", ast.True().AsIsNode(), false},
		{"notValue", ast.Context().AsIsNode(), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := isNonBoolValue(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestIsVariable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   types.Value
		out  bool
	}{
		{"happy", Variable("test"), true},
		{"sad", types.String("test"), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := IsVariable(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestIsIgnore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   types.Value
		out  bool
	}{
		{"happy", Ignore(), true},
		{"sad", types.String("test"), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := IsIgnore(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestToVariable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   types.EntityUID
		key  types.String
		out  bool
	}{
		{"happy", types.NewEntityUID(variableEntityType, "test"), "test", true},
		{"sad", types.NewEntityUID("X", "1"), "", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key, out := ToVariable(tt.in)
			testutil.Equals(t, key, tt.key)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestIsPartialError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   ast.IsNode
		out  bool
	}{
		{"partialerror", PartialError(errors.New("value must be boolean")), true},
		{"othernode", ast.True().AsIsNode(), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := IsPartialError(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}
