package ast_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    *ast.Policy
		errFunc func(testing.TB, error)
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
					ast.Context().Access("tls_version").Equals(ast.String("1.3")),
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
			ast.Permit().PrincipalIs(types.String("T")),
			testutil.OK,
		},
		{
			"principalIsIn",
			`{"effect":"permit","principal":{"op":"is","entity_type":"T","in":{"entity":{"type":"P","id":"42"}}},"action":{"op":"All"},"resource":{"op":"All"}}`,
			ast.Permit().PrincipalIsIn(types.String("T"), types.NewEntityUID("P", "42")),
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
			ast.Permit().ResourceIs(types.String("T")),
			testutil.OK,
		},
		{
			"resourceIsIn",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"is","entity_type":"T","in":{"entity":{"type":"P","id":"42"}}}}`,
			ast.Permit().ResourceIsIn(types.String("T"), types.NewEntityUID("P", "42")),
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
			ast.Permit().When(ast.Entity(types.NewEntityUID("T", "42"))),
			testutil.OK,
		},
		{
			"set",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Set":[{"Value":42},{"Value":"bananas"}]}}]}`,
			ast.Permit().When(ast.Set(types.Set{types.Long(42), types.String("bananas")})),
			testutil.OK,
		},
		{
			"record",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"Record":{"key":{"Value":42}}}}]}`,
			ast.Permit().When(ast.Record(types.Record{"key": types.Long(42)})),
			testutil.OK,
		},
		{
			"decimal",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"decimal":[{"Value":"42.24"}]}}]}`,
			ast.Permit().When(ast.Decimal(mustParseDecimal("42.24"))),
			testutil.OK,
		},
		{
			"ip",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"ip":[{"Value":"10.0.0.42/8"}]}}]}`,
			ast.Permit().When(ast.IPAddr(mustParseIPAddr("10.0.0.42/8"))),
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
			ast.Permit().When(ast.Long(42).Equals(ast.Long(24))),
			testutil.OK,
		},
		{
			"notEquals",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"!=":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).NotEquals(ast.Long(24))),
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
			ast.Permit().When(ast.Long(42).Plus(ast.Long(24))),
			testutil.OK,
		},
		{
			"minus",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"-":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).Minus(ast.Long(24))),
			testutil.OK,
		},
		{
			"times",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"*":{"left":{"Value":42},"right":{"Value":24}}}}]}`,
			ast.Permit().When(ast.Long(42).Times(ast.Long(24))),
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
			ast.Permit().When(ast.Resource().IsIn("T", ast.Entity(types.NewEntityUID("P", "42")))),
			testutil.OK,
		},
		{
			"like",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"like":{"left":{"Value":"text"},"pattern":"*"}}}]}`,
			ast.Permit().When(ast.String("text").Like("*")),
			testutil.OK,
		},
		{
			"ifThenElse",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"if-then-else":{"if":{"Value":true},"then":{"Value":42},"else":{"Value":24}}}}]}`,
			ast.Permit().When(ast.If(ast.True(), ast.Long(42), ast.Long(24))),
			testutil.OK,
		},
		{
			"isInRange",
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},
			"conditions":[{"kind":"when","body":{"isInRange":[
				{"ip":[{"Value":"10.0.0.43"}]},
				{"ip":[{"Value":"10.0.0.42/8"}]}
			]}}]}`,
			ast.Permit().When(ast.IPAddr(mustParseIPAddr("10.0.0.43")).IsInRange(ast.IPAddr(mustParseIPAddr("10.0.0.42/8")))),
			testutil.OK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p ast.Policy
			err := json.Unmarshal([]byte(tt.input), &p)
			tt.errFunc(t, err)
			if err != nil {
				return
			}
			testutil.Equals(t, p, *tt.want)
			b, err := json.Marshal(&p)
			testutil.OK(t, err)
			normInput := testNormalizeJSON(t, tt.input)
			normOutput := testNormalizeJSON(t, string(b))
			testutil.Equals(t, normOutput, normInput)
		})
	}
}

func testNormalizeJSON(t testing.TB, in string) string {
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
func mustParseIPAddr(v string) types.IPAddr {
	res, _ := types.ParseIPAddr(v)
	return res
}
