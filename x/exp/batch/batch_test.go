package batch

import (
	"context"
	"maps"
	"reflect"
	"testing"

	"github.com/cedar-policy/cedar-go"
	publicast "github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestBatch(t *testing.T) {
	t.Parallel()
	p1, p2, p3 := types.NewEntityUID("P", "1"), types.NewEntityUID("P", "2"), types.NewEntityUID("P", "3")
	a1, a2, a3 := types.NewEntityUID("A", "1"), types.NewEntityUID("A", "2"), types.NewEntityUID("A", "3")
	r1, r2, r3 := types.NewEntityUID("R", "1"), types.NewEntityUID("R", "2"), types.NewEntityUID("R", "3")
	_, _, _, _, _, _, _, _, _ = p1, p2, p3, a1, a2, a3, r1, r2, r3
	tests := []struct {
		name     string
		policy   *ast.Policy
		entities types.Entities
		request  Request
		results  []Result
	}{
		{"smokeTest",
			ast.Permit(),
			types.Entities{},
			Request{
				Principal: p1,
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context:   types.Record{},
				Variables: Variables{
					"action":   []types.Value{a1, a2},
					"resource": []types.Value{r1, r2, r3},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}}, Decision: true, Values: Values{"action": a1, "resource": r1}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}}, Decision: true, Values: Values{"action": a1, "resource": r2}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r3, Context: types.Record{}}, Decision: true, Values: Values{"action": a1, "resource": r3}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r1, Context: types.Record{}}, Decision: true, Values: Values{"action": a2, "resource": r1}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r2, Context: types.Record{}}, Decision: true, Values: Values{"action": a2, "resource": r2}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r3, Context: types.Record{}}, Decision: true, Values: Values{"action": a2, "resource": r3}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
			},
		},

		{"someOk",
			ast.Permit().PrincipalEq(p1).ActionEq(a2).ResourceEq(r3),
			types.Entities{},
			Request{
				Principal: p1,
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context:   types.Record{},
				Variables: Variables{
					"action":   []types.Value{a1, a2},
					"resource": []types.Value{r1, r2, r3},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}}, Decision: false, Values: Values{"action": a1, "resource": r1}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}}, Decision: false, Values: Values{"action": a1, "resource": r2}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r3, Context: types.Record{}}, Decision: false, Values: Values{"action": a1, "resource": r3}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r1, Context: types.Record{}}, Decision: false, Values: Values{"action": a2, "resource": r1}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r2, Context: types.Record{}}, Decision: false, Values: Values{"action": a2, "resource": r2}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r3, Context: types.Record{}}, Decision: true, Values: Values{"action": a2, "resource": r3}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
			},
		},

		{"attributeAccess",
			ast.Permit().When(ast.Principal().Access("tags").Has("a").And(ast.Principal().Access("tags").Access("a").Equal(ast.String("a")))),
			types.Entities{
				p1: {
					UID: p1,
					Attributes: types.Record{
						"tags": types.Record{"a": types.String("a")},
					},
				},
				p2: {
					UID: p2,
					Attributes: types.Record{
						"tags": types.Record{"b": types.String("b")},
					},
				},
			},
			Request{
				Principal: Variable("principal"),
				Action:    a1,
				Resource:  Variable("resource"),
				Context:   types.Record{},
				Variables: Variables{
					"principal": []types.Value{p1, p2},
					"resource":  []types.Value{r1, r2},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}}, Decision: true, Values: Values{"principal": p1, "resource": r1}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}}, Decision: true, Values: Values{"principal": p1, "resource": r2}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p2, Action: a1, Resource: r1, Context: types.Record{}}, Decision: false, Values: Values{"principal": p2, "resource": r1}},
				{Request: types.Request{Principal: p2, Action: a1, Resource: r2, Context: types.Record{}}, Decision: false, Values: Values{"principal": p2, "resource": r2}},
			},
		},

		{"contextAccess",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			types.Entities{},
			Request{
				Principal: p1,
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context: types.Record{
					"key": types.Long(42),
				},
				Variables: Variables{
					"action":   []types.Value{a1, a2},
					"resource": []types.Value{r1, r2, r3},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(42)}}, Decision: true, Values: Values{"action": a1, "resource": r1}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.Record{"key": types.Long(42)}}, Decision: true, Values: Values{"action": a1, "resource": r2}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r3, Context: types.Record{"key": types.Long(42)}}, Decision: true, Values: Values{"action": a1, "resource": r3}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r1, Context: types.Record{"key": types.Long(42)}}, Decision: true, Values: Values{"action": a2, "resource": r1}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r2, Context: types.Record{"key": types.Long(42)}}, Decision: true, Values: Values{"action": a2, "resource": r2}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r3, Context: types.Record{"key": types.Long(42)}}, Decision: true, Values: Values{"action": a2, "resource": r3}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
			},
		},

		{"variableContext",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			types.Entities{},
			Request{
				Principal: p1,
				Action:    a1,
				Resource:  r1,
				Context:   Variable("context"),
				Variables: Variables{
					"context": []types.Value{types.Record{"key": types.Long(41)}, types.Record{"key": types.Long(42)}, types.Record{"key": types.Long(43)}},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(41)}}, Decision: false, Values: Values{"context": types.Record{"key": types.Long(41)}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(42)}}, Decision: true, Values: Values{"context": types.Record{"key": types.Long(42)}}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(43)}}, Decision: false, Values: Values{"context": types.Record{"key": types.Long(43)}}},
			},
		},

		{"variableContextAccess",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			types.Entities{},
			Request{
				Principal: p1,
				Action:    a1,
				Resource:  r1,
				Context: types.Record{
					"key": Variable("key"),
				},
				Variables: Variables{
					"key": []types.Value{types.Long(41), types.Long(42), types.Long(43)},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(41)}}, Decision: false, Values: Values{"key": types.Long(41)}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(42)}}, Decision: true, Values: Values{"key": types.Long(42)}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(43)}}, Decision: false, Values: Values{"key": types.Long(43)}},
			},
		},

		{"ignoreContext",
			ast.Permit().
				When(ast.Context().Access("key").Equal(ast.Long(42))).
				When(ast.Principal().Equal(ast.Value(p1))).
				When(ast.Action().Equal(ast.Value(a1))).
				When(ast.Resource().Equal(ast.Value(r2))),

			types.Entities{},
			Request{
				Principal: p1,
				Action:    a1,
				Resource:  Variable("resource"),
				Context:   Ignore(),
				Variables: Variables{
					"resource": []types.Value{r1, r2},
					"key":      []types.Value{types.Long(41), types.Long(42), types.Long(43)},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}}, Decision: false, Values: Values{"resource": r1, "key": types.Long(41)}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}}, Decision: false, Values: Values{"resource": r1, "key": types.Long(42)}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}}, Decision: false, Values: Values{"resource": r1, "key": types.Long(43)}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}}, Decision: true, Values: Values{"resource": r2, "key": types.Long(41)}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}}, Decision: true, Values: Values{"resource": r2, "key": types.Long(42)}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}}, Decision: true, Values: Values{"resource": r2, "key": types.Long(43)}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			var res []Result
			ps := cedar.NewPolicySet()
			ps.Store("0", cedar.NewPolicyFromAST((*publicast.Policy)(tt.policy)))

			err := Authorize(context.Background(), ps, tt.entities, tt.request, func(br Result) {
				br.Request.Context = maps.Clone(br.Request.Context)
				br.Values = maps.Clone(br.Values)
				res = append(res, br)
			})
			testutil.OK(t, err)
			testutil.Equals(t, len(res), len(tt.results))
			for _, a := range tt.results {
				var found bool
				for _, b := range res {
					found = found || reflect.DeepEqual(a, b)
				}
				testutil.Equals(t, found, true)
			}
		})
	}
}
