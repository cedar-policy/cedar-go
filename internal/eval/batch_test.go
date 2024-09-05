package eval

import (
	"context"
	"maps"
	"reflect"
	"testing"

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
		request  BatchRequest
		results  []BatchDiagnosticResult
	}{
		{"smokeTest",
			ast.Permit(),
			types.Entities{},
			BatchRequest{
				Principal: p1,
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context:   types.Record{},
				Variables: Variables{
					"action":   []types.Value{a1, a2},
					"resource": []types.Value{r1, r2, r3},
				},
			},
			[]BatchDiagnosticResult{
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}, Decision: true, Values: Values{"action": a1, "resource": r1}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}, Decision: true, Values: Values{"action": a1, "resource": r2}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r3, Context: types.Record{}, Decision: true, Values: Values{"action": a1, "resource": r3}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r1, Context: types.Record{}, Decision: true, Values: Values{"action": a2, "resource": r1}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r2, Context: types.Record{}, Decision: true, Values: Values{"action": a2, "resource": r2}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r3, Context: types.Record{}, Decision: true, Values: Values{"action": a2, "resource": r3}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
			},
		},

		{"someOk",
			ast.Permit().PrincipalEq(p1).ActionEq(a2).ResourceEq(r3),
			types.Entities{},
			BatchRequest{
				Principal: p1,
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context:   types.Record{},
				Variables: Variables{
					"action":   []types.Value{a1, a2},
					"resource": []types.Value{r1, r2, r3},
				},
			},
			[]BatchDiagnosticResult{
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}, Decision: false, Values: Values{"action": a1, "resource": r1}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}, Decision: false, Values: Values{"action": a1, "resource": r2}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r3, Context: types.Record{}, Decision: false, Values: Values{"action": a1, "resource": r3}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r1, Context: types.Record{}, Decision: false, Values: Values{"action": a2, "resource": r1}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r2, Context: types.Record{}, Decision: false, Values: Values{"action": a2, "resource": r2}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r3, Context: types.Record{}, Decision: true, Values: Values{"action": a2, "resource": r3}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
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
			BatchRequest{
				Principal: Variable("principal"),
				Action:    a1,
				Resource:  Variable("resource"),
				Context:   types.Record{},
				Variables: Variables{
					"principal": []types.Value{p1, p2},
					"resource":  []types.Value{r1, r2},
				},
			},
			[]BatchDiagnosticResult{
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}, Decision: true, Values: Values{"principal": p1, "resource": r1}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}, Decision: true, Values: Values{"principal": p1, "resource": r2}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p2, Action: a1, Resource: r1, Context: types.Record{}, Decision: false, Values: Values{"principal": p2, "resource": r1}}},
				{BatchResult: BatchResult{Principal: p2, Action: a1, Resource: r2, Context: types.Record{}, Decision: false, Values: Values{"principal": p2, "resource": r2}}},
			},
		},

		{"contextAccess",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			types.Entities{},
			BatchRequest{
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
			[]BatchDiagnosticResult{
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(42)}, Decision: true, Values: Values{"action": a1, "resource": r1}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r2, Context: types.Record{"key": types.Long(42)}, Decision: true, Values: Values{"action": a1, "resource": r2}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r3, Context: types.Record{"key": types.Long(42)}, Decision: true, Values: Values{"action": a1, "resource": r3}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r1, Context: types.Record{"key": types.Long(42)}, Decision: true, Values: Values{"action": a2, "resource": r1}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r2, Context: types.Record{"key": types.Long(42)}, Decision: true, Values: Values{"action": a2, "resource": r2}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a2, Resource: r3, Context: types.Record{"key": types.Long(42)}, Decision: true, Values: Values{"action": a2, "resource": r3}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
			},
		},

		{"variableContext",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			types.Entities{},
			BatchRequest{
				Principal: p1,
				Action:    a1,
				Resource:  r1,
				Context:   Variable("context"),
				Variables: Variables{
					"context": []types.Value{types.Record{"key": types.Long(41)}, types.Record{"key": types.Long(42)}, types.Record{"key": types.Long(43)}},
				},
			},
			[]BatchDiagnosticResult{
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(41)}, Decision: false, Values: Values{"context": types.Record{"key": types.Long(41)}}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(42)}, Decision: true, Values: Values{"context": types.Record{"key": types.Long(42)}}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(43)}, Decision: false, Values: Values{"context": types.Record{"key": types.Long(43)}}}},
			},
		},

		{"variableContextAccess",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			types.Entities{},
			BatchRequest{
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
			[]BatchDiagnosticResult{
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(41)}, Decision: false, Values: Values{"key": types.Long(41)}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(42)}, Decision: true, Values: Values{"key": types.Long(42)}}, Diagnostic: Diagnostic{Reasons: []string{"0"}}},
				{BatchResult: BatchResult{Principal: p1, Action: a1, Resource: r1, Context: types.Record{"key": types.Long(43)}, Decision: false, Values: Values{"key": types.Long(43)}}},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Run("basicCallback", func(t *testing.T) {
				var res []BatchResult
				err := Batch(context.Background(), []*ast.Policy{tt.policy}, tt.entities, tt.request, WithCallback(func(br BatchResult) {
					br.Context = maps.Clone(br.Context)
					br.Values = maps.Clone(br.Values)
					res = append(res, br)
				}))
				testutil.OK(t, err)
				testutil.Equals(t, len(res), len(tt.results))
				for _, a := range tt.results {
					var found bool
					for _, b := range res {
						found = found || reflect.DeepEqual(a.BatchResult, b)
					}
					testutil.Equals(t, found, true)
				}
			})
			t.Run("diagnosticCallback", func(t *testing.T) {
				var res []BatchDiagnosticResult
				err := Batch(context.Background(), []*ast.Policy{tt.policy}, tt.entities, tt.request, WithDiagnosticCallback(func(br BatchDiagnosticResult) {
					br.Context = maps.Clone(br.Context)
					br.Values = maps.Clone(br.Values)
					res = append(res, br)
				}))
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
		})
	}
}
