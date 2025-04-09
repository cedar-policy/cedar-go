package json

import (
	"encoding/json"
	"net/netip"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    *ast.Policy
		errFunc func(testutil.TB, error)
	}{
		/*
		   @key("value")
		   permit (
		       principal == User::"12UA45",
		       action == Action::"view",
		       resource in Folder::"abc"
		   ) when {
		       context.tls_version == "1.3"
		   };
		*/
		{"exampleFromDocs", `{
	"annotations": {
		"key": "value"
	},
    "effect": "permit",
    "principal": {
        "op": "==",
        "entity": {
			"type": "User",
			"id": "12UA45"
		}
    },
    "action": {
        "op": "==",
        "entity": {
			"type": "Action",
			"id": "view"
		}
    },
    "resource": {
        "op": "in",
        "entity": {
			"type": "Folder",
			"id": "abc"
		}
    },
    "conditions": [
        {
            "kind": "when",
            "body": {
                "==": {
                    "left": {
                        ".": {
                            "left": {
                                "Var": "context"
                            },
                            "attr": "tls_version"
                        }
                    },
                    "right": {
                        "Value": "1.3"
                    }
                }
            }
        }
    ]
}`,
			ast.Permit().
				Annotate("key", "value").
				PrincipalEq(types.NewEntityUID("User", "12UA45")).
				ActionEq(types.NewEntityUID("Action", "view")).
				ResourceIn(types.NewEntityUID("Folder", "abc")).
				When(
					ast.Context().Access("tls_version").Equal(ast.String("1.3")),
				),
			testutil.OK,
		},

		{
			"permit",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`,
			ast.Permit(),
			testutil.OK,
		},
		{
			"forbid",
			`{"effect":"forbid","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`,
			ast.Forbid(),
			testutil.OK,
		},
		{
			"annotations",
			`{"annotations":{"key":"value"},"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`,
			ast.Permit().Annotate("key", "value"),
			testutil.OK,
		},
		{
			"principalEq",
			`{"effect":"permit","principal":{"op":"==","entity":{"type":"T","id":"42"}},"action":{"op":"All"},"resource":{"op":"All"}}`,
			ast.Permit().PrincipalEq(types.NewEntityUID("T", "42")),
			testutil.OK,
		},
		{
			"principalIn",
			`{"effect":"permit","principal":{"op":"in","entity":{"type":"T","id":"42"}},"action":{"op":"All"},"resource":{"op":"All"}}`,
			ast.Permit().PrincipalIn(types.NewEntityUID("T", "42")),
			testutil.OK,
		},
		{
			"principalIs",
			`{"effect":"permit","principal":{"op":"is","entity_type":"T"},"action":{"op":"All"},"resource":{"op":"All"}}`,
			ast.Permit().PrincipalIs(types.EntityType("T")),
			testutil.OK,
		},
		{
			"principalIsIn",
			`{"effect":"permit","principal":{"op":"is","entity_type":"T","in":{"entity":{"type":"P","id":"42"}}},"action":{"op":"All"},"resource":{"op":"All"}}`,
			ast.Permit().PrincipalIsIn(types.EntityType("T"), types.NewEntityUID("P", "42")),
			testutil.OK,
		},
		{
			"actionEq",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"==","entity":{"type":"T","id":"42"}},"resource":{"op":"All"}}`,
			ast.Permit().ActionEq(types.NewEntityUID("T", "42")),
			testutil.OK,
		},
		{
			"actionIn",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"in","entity":{"type":"T","id":"42"}},"resource":{"op":"All"}}`,
			ast.Permit().ActionIn(types.NewEntityUID("T", "42")),
			testutil.OK,
		},
		{
			"actionInSet",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"in","entities":[{"type":"T","id":"42"},{"type":"T","id":"43"}]},"resource":{"op":"All"}}`,
			ast.Permit().ActionInSet(types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43")),
			testutil.OK,
		},
		{
			"resourceEq",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"==","entity":{"type":"T","id":"42"}}}`,
			ast.Permit().ResourceEq(types.NewEntityUID("T", "42")),
			testutil.OK,
		},
		{
			"resourceIn",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"in","entity":{"type":"T","id":"42"}}}`,
			ast.Permit().ResourceIn(types.NewEntityUID("T", "42")),
			testutil.OK,
		},
		{
			"resourceIs",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"is","entity_type":"T"}}`,
			ast.Permit().ResourceIs(types.EntityType("T")),
			testutil.OK,
		},
		{
			"resourceIsIn",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"is","entity_type":"T","in":{"entity":{"type":"P","id":"42"}}}}`,
			ast.Permit().ResourceIsIn(types.EntityType("T"), types.NewEntityUID("P", "42")),
			testutil.OK,
		},
		{
			"when",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Value":true}}]}`,
			ast.Permit().When(ast.True()),
			testutil.OK,
		},
		{
			"unless",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"unless","body":{"Value":false}}]}`,
			ast.Permit().Unless(ast.False()),
			testutil.OK,
		},
		{
			"long",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Value":42}}]}`,
			ast.Permit().When(ast.Long(42)),
			testutil.OK,
		},
		{
			"string",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Value":"bananas"}}]}`,
			ast.Permit().When(ast.String("bananas")),
			testutil.OK,
		},
		{
			"entity",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Value":{"__entity":{"type":"T","id":"42"}}}}]}`,
			ast.Permit().When(ast.EntityUID("T", "42")),
			testutil.OK,
		},
		{
			"set",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Set":[{"Value":42},{"Value":"bananas"}]}}]}`,
			ast.Permit().When(ast.Set(ast.Long(42), ast.String("bananas"))),
			testutil.OK,
		},
		{
			"record",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Record":{"key":{"Value":42}}}}]}`,
			ast.Permit().When(ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(42)}})),
			testutil.OK,
		},
		{
			"principalVar",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Var":"principal"}}]}`,
			ast.Permit().When(ast.Principal()),
			testutil.OK,
		},
		{
			"actionVar",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Var":"action"}}]}`,
			ast.Permit().When(ast.Action()),
			testutil.OK,
		},
		{
			"resourceVar",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Var":"resource"}}]}`,
			ast.Permit().When(ast.Resource()),
			testutil.OK,
		},
		{
			"contextVar",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Var":"context"}}]}`,
			ast.Permit().When(ast.Context()),
			testutil.OK,
		},
		{
			"not",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"!":{"arg":{"Value":true}}}}]}`,
			ast.Permit().When(ast.Not(ast.True())),
			testutil.OK,
		},
		{
			"negate",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"neg":{"arg":{"Value":42}}}}]}`,
			ast.Permit().When(ast.Negate(ast.Long(42))),
			testutil.OK,
		},
		{
			"equals",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"==":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).Equal(ast.Long(24))),
			testutil.OK,
		},
		{
			"notEquals",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"!=":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).NotEqual(ast.Long(24))),
			testutil.OK,
		},
		{
			"in",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"in":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).In(ast.Long(24))),
			testutil.OK,
		},
		{
			"lessThan",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"<":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).LessThan(ast.Long(24))),
			testutil.OK,
		},
		{
			"lessThanEquals",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"<=":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).LessThanOrEqual(ast.Long(24))),
			testutil.OK,
		},
		{
			"greaterThan",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{">":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).GreaterThan(ast.Long(24))),
			testutil.OK,
		},
		{
			"greaterThanEquals",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{">=":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).GreaterThanOrEqual(ast.Long(24))),
			testutil.OK,
		},
		{
			"and",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"&&":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).And(ast.Long(24))),
			testutil.OK,
		},
		{
			"or",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"||":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).Or(ast.Long(24))),
			testutil.OK,
		},
		{
			"plus",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"+":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).Add(ast.Long(24))),
			testutil.OK,
		},
		{
			"minus",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"-":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).Subtract(ast.Long(24))),
			testutil.OK,
		},
		{
			"times",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"*":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).Multiply(ast.Long(24))),
			testutil.OK,
		},
		{
			"contains",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"contains":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).Contains(ast.Long(24))),
			testutil.OK,
		},
		{
			"containsAll",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"containsAll":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).ContainsAll(ast.Long(24))),
			testutil.OK,
		},
		{
			"containsAny",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"containsAny":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).ContainsAny(ast.Long(24))),
			testutil.OK,
		},
		{
			"isEmpty",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"isEmpty":{"arg":{"Value":42}}}}]}`,
			ast.Permit().When(ast.Long(42).IsEmpty()),
			testutil.OK,
		},
		{
			"access",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{".":{"left":{"Var":"context"},"attr":"key"}}}]}`,
			ast.Permit().When(ast.Context().Access("key")),
			testutil.OK,
		},
		{
			"has",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"has":{"left":{"Var":"context"},"attr":"key"}}}]}`,
			ast.Permit().When(ast.Context().Has("key")),
			testutil.OK,
		},
		{
			"getTag",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"getTag":{"left": {"Var": "principal"},"right": {"Value": "key"}}}}]}`,
			ast.Permit().When(ast.Principal().GetTag(ast.String("key"))),
			testutil.OK,
		},
		{
			"hasTag",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"hasTag":{"left": {"Var": "principal"},"right": {"Value": "key"}}}}]}`,
			ast.Permit().When(ast.Principal().HasTag(ast.String("key"))),
			testutil.OK,
		},
		{
			"is",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"is":{"left":{"Var":"resource"},"entity_type":"T"}}}]}`,
			ast.Permit().When(ast.Resource().Is("T")),
			testutil.OK,
		},
		{
			"isIn",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"is":{"left":{"Var":"resource"},"entity_type":"T","in":{"Value":{"__entity":{"type":"P","id":"42"}}}}}}]}`,
			ast.Permit().When(ast.Resource().IsIn("T", ast.EntityUID("P", "42"))),
			testutil.OK,
		},
		// N.B. Most pattern parsing tests can be found in types/pattern_test.go
		{
			"like single wildcard",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"like":{"left":{"Value":"text"},"pattern":["Wildcard"]}}}]}`,
			ast.Permit().When(ast.String("text").Like(types.NewPattern(types.Wildcard{}))),
			testutil.OK,
		},
		{
			"like single literal",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"like":{"left":{"Value":"text"},"pattern":[{"Literal":"foo"}]}}}]}`,
			ast.Permit().When(ast.String("text").Like(types.NewPattern(types.String("foo")))),
			testutil.OK,
		},
		{
			"like wildcard then literal",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"like":{"left":{"Value":"text"},"pattern":["Wildcard", {"Literal":"foo"}]}}}]}`,
			ast.Permit().When(ast.String("text").Like(types.NewPattern(types.Wildcard{}, types.String("foo")))),
			testutil.OK,
		},
		{
			"like literal then wildcard",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"like":{"left":{"Value":"text"},"pattern":[{"Literal":"foo"}, "Wildcard"]}}}]}`,
			ast.Permit().When(ast.String("text").Like(types.NewPattern(types.String("foo"), types.Wildcard{}))),
			testutil.OK,
		},
		{
			"ifThenElse",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"if-then-else":{"if":{"Value":true},"then":{"Value":42},"else":{"Value":24}}}}]}`,
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.Long(42), ast.Long(24))),
			testutil.OK,
		},
		{
			"decimal",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"decimal":[{"Value":"42.24"}]}}]}`,
			ast.Permit().When(ast.ExtensionCall("decimal", ast.String("42.24"))),
			testutil.OK,
		},
		{
			"ip",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"ip":[{"Value":"10.0.0.42/8"}]}}]}`,
			ast.Permit().When(ast.ExtensionCall("ip", ast.String("10.0.0.42/8"))),
			testutil.OK,
		},
		{
			"isInRange",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"isInRange":[
				{"ip":[{"Value":"10.0.0.43"}]},
				{"ip":[{"Value":"10.0.0.42/8"}]}
			]}}]}`,
			ast.Permit().When(ast.ExtensionCall("ip", ast.String("10.0.0.43")).IsInRange(ast.ExtensionCall("ip", ast.String("10.0.0.42/8")))),
			testutil.OK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p Policy
			err := json.Unmarshal([]byte(tt.input), &p)
			tt.errFunc(t, err)
			if err != nil {
				return
			}
			testutil.Equals(t, p.unwrap(), tt.want)
			b, err := json.Marshal(&p)
			testutil.OK(t, err)
			normInput := testNormalizeJSON(t, tt.input)
			normOutput := testNormalizeJSON(t, string(b))
			testutil.Equals(t, normOutput, normInput)
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	// most cases are covered in the TestUnmarshalJSON round-trip.
	// this covers some cases that aren't 1:1 round-tripppable, such as hard-coded IP/Decimal values.

	t.Parallel()
	tests := []struct {
		name    string
		input   *ast.Policy
		want    string
		errFunc func(testutil.TB, error)
	}{
		{
			"decimal",
			ast.Permit().When(ast.Value(mustParseDecimal("42.24"))),
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"decimal":[{"Value":"42.24"}]}}]}`,
			testutil.OK,
		},
		{
			"ip",
			ast.Permit().When(ast.IPAddr(mustParseIPAddr("10.0.0.42/8"))),
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"ip":[{"Value":"10.0.0.42/8"}]}}]}`,
			testutil.OK,
		},
		{
			"isInRange",
			ast.Permit().When(ast.IPAddr(mustParseIPAddr("10.0.0.43")).IsInRange(ast.IPAddr(mustParseIPAddr("10.0.0.42/8")))),
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"isInRange":[
				{"ip":[{"Value":"10.0.0.43"}]},
				{"ip":[{"Value":"10.0.0.42/8"}]}
			]}}]}`,
			testutil.OK,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pp := wrapPolicy(tt.input)
			b, err := json.Marshal(pp)
			tt.errFunc(t, err)
			if err != nil {
				return
			}
			normGot := testNormalizeJSON(t, string(b))
			normWant := testNormalizeJSON(t, tt.want)
			testutil.Equals(t, normGot, normWant)
		})
	}
}

func testNormalizeJSON(t testutil.TB, in string) string {
	var x any
	err := json.Unmarshal([]byte(in), &x)
	testutil.OK(t, err)
	out, err := json.MarshalIndent(x, "", "    ")
	testutil.OK(t, err)
	return string(out)
}

func mustParseDecimal(v string) types.Decimal {
	res, _ := types.ParseDecimal(v)
	return res
}
func mustParseIPAddr(v string) netip.Prefix {
	res, _ := types.ParseIPAddr(v)
	return netip.Prefix(res)
}

func TestMarshalPanics(t *testing.T) {
	t.Parallel()
	t.Run("nilScope", func(t *testing.T) {
		t.Parallel()
		testutil.Panic(t, func() {
			s := scopeJSON{}
			var v ast.IsScopeNode
			s.FromNode(v)
		})
	})
	t.Run("nilNode", func(t *testing.T) {
		t.Parallel()
		testutil.Panic(t, func() {
			s := nodeJSON{}
			var v ast.IsNode
			s.FromNode(v)
		})
	})
}

func TestUnmarshalErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{
			"effect",
			`{"effect":"unknown","principal":{"op":"=="},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"principalScopeEqMissingEntity",
			`{"effect":"permit","principal":{"op":"=="},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"principalScopeInMissingEntity",
			`{"effect":"permit","principal":{"op":"in"},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"actionScopeEqMissingEntity",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"action":{"op":"=="}}`,
		},
		{
			"scopeUnknownOp",
			`{"effect":"permit","principal":{"op":"???"},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"actionUnknownOp",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"???"},"resource":{"op":"All"}}`,
		},
		{
			"resourceUnknownOp",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"???"}}`,
		},
		{
			"conditionUnknown",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"unknown","body":{"Value":24}}]}`,
		},
		{
			"binaryLeft",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"&&":{"left":null,"right":{"Value":24}}}}]}`,
		},
		{
			"binaryRight",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"&&":{"left":{"Value":24},"right":null}}}]}`,
		},
		{
			"unaryArg",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"!":{"arg":null}}}]}`,
		},
		{
			"accessLeft",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{".":{"left":null,"attr":"key"}}}]}`,
		},
		{
			"patternLeft",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"like":{"left":null,"pattern":["Wildcard"]}}}]}`,
		},
		{
			"patternWildcard",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"like":{"left":{"Value":"text"},"pattern":["invalid"]}}}]}`,
		},
		{
			"isLeft",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"is":{"left":null,"entity_type":"T"}}}]}`,
		},
		{
			"isIn",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"is":{"left":{"Var":"resource"},"entity_type":"T","in":{"Value":null}}}}]}`,
		},
		{
			"ifErrThenElse",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"if-then-else":{"if":{"Value":null},"then":{"Value":42},"else":{"Value":24}}}}]}`,
		},
		{
			"ifThenErrElse",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"if-then-else":{"if":{"Value":true},"then":{"Value":null},"else":{"Value":24}}}}]}`,
		},
		{
			"ifThenElseErr",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"if-then-else":{"if":{"Value":true},"then":{"Value":42},"else":{"Value":null}}}}]}`,
		},
		{
			"setErr",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Set":[{"Value":null},{"Value":"bananas"}]}}]}`,
		},
		{
			"recordErr",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Record":{"key":{"Value":null}}}}]}`,
		},
		{
			"extensionTooMany",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"ip":[{"Value":"10.0.0.42/8"}],"pi":[{"Value":"3.14"}]}}]}`,
		},
		{
			"extensionArg",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"ip":[{"Value":null}]}}]}`,
		},
		{
			"var",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Var":"unknown"}}]}`,
		},
		{
			"otherJSONerror",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":42}]}`,
		},
		{
			"unknown-extension-function",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"not_an_extension_function":[]}}]}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p Policy
			err := json.Unmarshal([]byte(tt.input), &p)
			testutil.Error(t, err)
		})
	}
}

func TestMarshalExtensions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		out  string
	}{
		{
			"decimalType",
			ast.Permit().When(ast.Value(testutil.Must(types.NewDecimalFromInt(42)))),
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[{"kind":"when","body":{"decimal":[{"Value":"42.0"}]}}]}`,
		},
		{
			"decimalExtension",
			ast.Permit().When(ast.ExtensionCall("decimal", ast.String("42.0"))),
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[{"kind":"when","body":{"decimal":[{"Value":"42.0"}]}}]}`,
		},
		{
			"ipType",
			ast.Permit().When(ast.Value(types.IPAddr(netip.MustParsePrefix("127.0.0.1/16")))),
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[{"kind":"when","body":{"ip":[{"Value":"127.0.0.1/16"}]}}]}`,
		},
		{
			"ipExtension",
			ast.Permit().When(ast.ExtensionCall("ip", ast.String("127.0.0.1/16"))),
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[{"kind":"when","body":{"ip":[{"Value":"127.0.0.1/16"}]}}]}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := (*Policy)(tt.in)
			out, err := p.MarshalJSON()
			testutil.OK(t, err)
			testutil.Equals(t, string(out), tt.out)

		})
	}
}
