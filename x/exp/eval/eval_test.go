package eval

import (
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestToEval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   ast.Node
		out  types.Value
		err  func(testutil.TB, error)
	}{
		{
			"access",
			ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)})).Access("key"),
			types.Long(42),
			testutil.OK,
		},
		{
			"has",
			ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)})).Has("key"),
			types.True,
			testutil.OK,
		},
		{
			"getTag",
			ast.EntityUID("T", "ID").GetTag(ast.String("key")),
			types.Long(42),
			testutil.OK,
		},
		{
			"hasTag",
			ast.EntityUID("T", "ID").HasTag(ast.String("key")),
			types.True,
			testutil.OK,
		},
		{
			"like",
			ast.String("test").Like(types.Pattern{}),
			types.False,
			testutil.OK,
		},
		{
			"if",
			ast.IfThenElse(ast.True(), ast.Long(42), ast.Long(43)),
			types.Long(42),
			testutil.OK,
		},
		{
			"is",
			ast.EntityUID("T", "42").Is("T"),
			types.True,
			testutil.OK,
		},
		{
			"isIn",
			ast.EntityUID("T", "42").IsIn("T", ast.EntityUID("T", "42")),
			types.True,
			testutil.OK,
		},
		{
			"value",
			ast.Long(42),
			types.Long(42),
			testutil.OK,
		},
		{
			"record",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(42)}}),
			types.NewRecord(types.RecordMap{"key": types.Long(42)}),
			testutil.OK,
		},
		{
			"set",
			ast.Set(ast.Long(42)),
			types.NewSet(types.Long(42)),
			testutil.OK,
		},
		{
			"negate",
			ast.Negate(ast.Long(42)),
			types.Long(-42),
			testutil.OK,
		},
		{
			"not",
			ast.Not(ast.True()),
			types.False,
			testutil.OK,
		},
		{
			"principal",
			ast.Principal(),
			types.NewEntityUID("Actor", "principal"),
			testutil.OK,
		},
		{
			"action",
			ast.Action(),
			types.NewEntityUID("Action", "test"),
			testutil.OK,
		},
		{
			"resource",
			ast.Resource(),
			types.NewEntityUID("Resource", "database"),
			testutil.OK,
		},
		{
			"context",
			ast.Context(),
			types.Record{},
			testutil.OK,
		},
		{
			"in",
			ast.EntityUID("T", "42").In(ast.EntityUID("T", "43")),
			types.False,
			testutil.OK,
		},
		{
			"and",
			ast.True().And(ast.False()),
			types.False,
			testutil.OK,
		},
		{
			"or",
			ast.True().Or(ast.False()),
			types.True,
			testutil.OK,
		},
		{
			"equals",
			ast.Long(42).Equal(ast.Long(43)),
			types.False,
			testutil.OK,
		},
		{
			"notEquals",
			ast.Long(42).NotEqual(ast.Long(43)),
			types.True,
			testutil.OK,
		},
		{
			"greaterThan",
			ast.Long(42).GreaterThan(ast.Long(43)),
			types.False,
			testutil.OK,
		},
		{
			"greaterThanOrEqual",
			ast.Long(42).GreaterThanOrEqual(ast.Long(43)),
			types.False,
			testutil.OK,
		},
		{
			"lessThan",
			ast.Long(42).LessThan(ast.Long(43)),
			types.True,
			testutil.OK,
		},
		{
			"lessThanOrEqual",
			ast.Long(42).LessThanOrEqual(ast.Long(43)),
			types.True,
			testutil.OK,
		},
		{
			"sub",
			ast.Long(42).Subtract(ast.Long(2)),
			types.Long(40),
			testutil.OK,
		},
		{
			"add",
			ast.Long(40).Add(ast.Long(2)),
			types.Long(42),
			testutil.OK,
		},
		{
			"mult",
			ast.Long(6).Multiply(ast.Long(7)),
			types.Long(42),
			testutil.OK,
		},
		{
			"contains",
			ast.Value(types.NewSet(types.Long(42))).Contains(ast.Long(42)),
			types.True,
			testutil.OK,
		},
		{
			"containsAll",
			ast.Value(types.NewSet(types.Long(42), types.Long(43), types.Long(44))).ContainsAll(ast.Value(types.NewSet(types.Long(42), types.Long(43)))),
			types.True,
			testutil.OK,
		},
		{
			"containsAny",
			ast.Value(types.NewSet(types.Long(42), types.Long(43), types.Long(44))).ContainsAny(ast.Value(types.NewSet(types.Long(1), types.Long(42)))),
			types.True,
			testutil.OK,
		},
		{
			"isEmpty",
			ast.Value(types.NewSet(types.Long(42), types.Long(43), types.Long(44))).IsEmpty(),
			types.False,
			testutil.OK,
		},
		{
			"ip",
			ast.ExtensionCall("ip", ast.String("127.0.0.42/16")),
			types.IPAddr(netip.MustParsePrefix("127.0.0.42/16")),
			testutil.OK,
		},
		{
			"decimal",
			ast.ExtensionCall("decimal", ast.String("42.42")),
			testutil.Must(types.NewDecimal(4242, -2)),
			testutil.OK,
		},
		{
			"datetime",
			ast.ExtensionCall("datetime", ast.String("1970-01-01T00:00:00.001Z")),
			types.NewDatetime(time.UnixMilli(1)),
			testutil.OK,
		},
		{
			"duration",
			ast.ExtensionCall("duration", ast.String("1ms")),
			types.NewDuration(1 * time.Millisecond),
			testutil.OK,
		},
		{
			"toDate",
			ast.ExtensionCall("toDate", ast.Value(types.NewDatetime(time.UnixMilli(1)))),
			types.NewDatetime(time.UnixMilli(0)),
			testutil.OK,
		},
		{
			"toTime",
			ast.ExtensionCall("toTime", ast.Value(types.NewDatetime(time.UnixMilli(1)))),
			types.NewDuration(1 * time.Millisecond),
			testutil.OK,
		},
		{
			"toDays",
			ast.ExtensionCall("toDays", ast.Value(types.NewDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"toHours",
			ast.ExtensionCall("toHours", ast.Value(types.NewDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"toMinutes",
			ast.ExtensionCall("toMinutes", ast.Value(types.NewDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"toSeconds",
			ast.ExtensionCall("toSeconds", ast.Value(types.NewDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"toMilliseconds",
			ast.ExtensionCall("toMilliseconds", ast.Value(types.NewDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"offset",
			ast.ExtensionCall("offset", ast.Value(types.NewDatetime(time.UnixMilli(0))), ast.Value(types.NewDuration(1*time.Millisecond))),
			types.NewDatetime(time.UnixMilli(1)),
			testutil.OK,
		},
		{
			"durationSince",
			ast.ExtensionCall("durationSince", ast.Value(types.NewDatetime(time.UnixMilli(1))), ast.Value(types.NewDatetime(time.UnixMilli(1)))),
			types.NewDuration(time.Duration(0)),
			testutil.OK,
		},

		{
			"lessThan",
			ast.ExtensionCall("lessThan", ast.Value(testutil.Must(types.NewDecimal(42, 0))), ast.Value(testutil.Must(types.NewDecimalFromInt(43)))),
			types.True,
			testutil.OK,
		},
		{
			"lessThanOrEqual",
			ast.ExtensionCall("lessThanOrEqual", ast.Value(testutil.Must(types.NewDecimal(42, 0))), ast.Value(testutil.Must(types.NewDecimalFromInt(43)))),
			types.True,
			testutil.OK,
		},
		{
			"greaterThan",
			ast.ExtensionCall("greaterThan", ast.Value(testutil.Must(types.NewDecimal(42, 0))), ast.Value(testutil.Must(types.NewDecimalFromInt(43)))),
			types.False,
			testutil.OK,
		},
		{
			"greaterThanOrEqual",
			ast.ExtensionCall("greaterThanOrEqual", ast.Value(testutil.Must(types.NewDecimal(42, 0))), ast.Value(testutil.Must(types.NewDecimalFromInt(43)))),
			types.False,
			testutil.OK,
		},
		{
			"isIpv4",
			ast.ExtensionCall("isIpv4", ast.IPAddr(netip.MustParsePrefix("127.0.0.42/16"))),
			types.True,
			testutil.OK,
		},
		{
			"isIpv6",
			ast.ExtensionCall("isIpv6", ast.IPAddr(netip.MustParsePrefix("::1/16"))),
			types.True,
			testutil.OK,
		},
		{
			"isLoopback",
			ast.ExtensionCall("isLoopback", ast.IPAddr(netip.MustParsePrefix("127.0.0.1/32"))),
			types.True,
			testutil.OK,
		},
		{
			"isMulticast",
			ast.ExtensionCall("isMulticast", ast.IPAddr(netip.MustParsePrefix("239.255.255.255/32"))),
			types.True,
			testutil.OK,
		},
		{
			"isInRange",
			ast.ExtensionCall("isInRange", ast.IPAddr(netip.MustParsePrefix("127.0.0.42/32")), ast.IPAddr(netip.MustParsePrefix("127.0.0.0/16"))),
			types.True,
			testutil.OK,
		},
		{
			"extUnknown",
			ast.ExtensionCall("unknown", ast.String("hello")),
			nil,
			testutil.Error,
		},
		{
			"extArgs",
			ast.ExtensionCall("ip", ast.String("1"), ast.String("2")),
			nil,
			testutil.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := Eval(tt.in.AsIsNode(), Env{
				Principal: types.NewEntityUID("Actor", "principal"),
				Action:    types.NewEntityUID("Action", "test"),
				Resource:  types.NewEntityUID("Resource", "database"),
				Context:   types.Record{},
				Entities: types.EntityMap{
					types.NewEntityUID("T", "ID"): types.Entity{
						Tags: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
					},
				},
			})
			tt.err(t, err)
			testutil.Equals(t, out, tt.out)
		})
	}

}

func TestToEvalPanic(t *testing.T) {
	t.Parallel()
	testutil.Panic(t, func() {
		_, _ = Eval(ast.Node{}.AsIsNode(), Env{})
	})
}

func TestToEvalVariablePanic(t *testing.T) {
	t.Parallel()
	testutil.Panic(t, func() {
		_, _ = Eval(ast.NodeTypeVariable{Name: "bananas"}, Env{})
	})
}
func TestPartialPolicyToNode(t *testing.T) {
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
			ast.Permit().When(ast.NewNode(PartialError(errors.New("type error: expected comparable value, got string")))),
			true,
			ast.True().And(ast.NewNode(PartialError(errors.New("type error: expected comparable value, got string")))),
		},
		{"conditionDropTypeError",
			ast.Permit().When(ast.Long(42)),
			Env{},
			ast.Permit().When(ast.NewNode(PartialError(errors.New("type error: condition expected bool")))),
			true,
			ast.True().And(ast.NewNode(PartialError(errors.New("type error: condition expected bool")))),
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
			ast.Permit().When(ast.NewNode(PartialError(errors.New("type error: expected comparable value, got string")))),
			true,
			ast.True().And(ast.NewNode(PartialError(errors.New("type error: expected comparable value, got string")))),
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
			ast.Permit().When(ast.NewNode(PartialError(errors.New("type error: expected comparable value, got string")))),
			true,
			ast.True().And(ast.NewNode(PartialError(errors.New("type error: expected comparable value, got string")))),
		},
		{"errorShortCircuitKept",
			ast.Permit().When(ast.Context().Access("variable")).When(ast.String("test").LessThan(ast.Long(42))).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.Context().Access("variable")).When(ast.NewNode(PartialError(errors.New("type error: expected comparable value, got string")))),
			true,
			ast.True().And(ast.Context().Access("variable").And(ast.NewNode(PartialError(errors.New("type error: expected comparable value, got string"))))),
		},
		{"errorConditionShortCircuit",
			ast.Permit().When(ast.True()).When(ast.String("test")).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.NewNode(PartialError(errors.New("type error: condition expected bool")))),
			true,
			ast.True().And(ast.NewNode(PartialError(errors.New("type error: condition expected bool")))),
		},
		{"errorConditionShortCircuitKept",
			ast.Permit().When(ast.Context().Access("variable")).When(ast.String("test")).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.Context().Access("variable")).When(ast.NewNode(PartialError(errors.New("type error: condition expected bool")))),
			true,
			ast.True().And(ast.Context().Access("variable").And(ast.NewNode(PartialError(errors.New("type error: condition expected bool"))))),
		},
		{"errorConditionShortCircuitKeptDeeper",
			ast.Permit().When(ast.Context().Access("variable")).When(ast.String("test")).When(ast.Context().Access("variable")),
			Env{
				Context: types.NewRecord(types.RecordMap{
					"variable": Variable("variable"),
				}),
			},
			ast.Permit().When(ast.Context().Access("variable")).When(ast.NewNode(PartialError(errors.New("type error: condition expected bool")))),
			true,
			ast.True().And(ast.Context().Access("variable").And(ast.NewNode(PartialError(errors.New("type error: condition expected bool"))))),
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
			out, keep := eval.PartialPolicy(tt.env, tt.in)
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
