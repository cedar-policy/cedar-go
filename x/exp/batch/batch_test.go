package batch

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go"
	publicast "github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/mapset"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
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
		entities types.EntityGetter
		request  Request
		results  []Result
	}{
		{"smokeTest",
			ast.Permit(),
			types.EntityMap{},
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

		{"nilEntityMap",
			ast.Permit(),
			nil,
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
			types.EntityMap{},
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
			types.EntityMap{
				p1: {
					UID: p1,
					Attributes: types.NewRecord(types.RecordMap{
						"tags": types.NewRecord(types.RecordMap{"a": types.String("a")}),
					}),
				},
				p2: {
					UID: p2,
					Attributes: types.NewRecord(types.RecordMap{
						"tags": types.NewRecord(types.RecordMap{"b": types.String("b")}),
					}),
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
			types.EntityMap{},
			Request{
				Principal: p1,
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context: types.NewRecord(types.RecordMap{
					"key": types.Long(42),
				}),
				Variables: Variables{
					"action":   []types.Value{a1, a2},
					"resource": []types.Value{r1, r2, r3},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Decision: true, Values: Values{"action": a1, "resource": r1}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Decision: true, Values: Values{"action": a1, "resource": r2}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r3, Context: types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Decision: true, Values: Values{"action": a1, "resource": r3}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r1, Context: types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Decision: true, Values: Values{"action": a2, "resource": r1}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r2, Context: types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Decision: true, Values: Values{"action": a2, "resource": r2}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a2, Resource: r3, Context: types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Decision: true, Values: Values{"action": a2, "resource": r3}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
			},
		},

		{"variableContext",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			types.EntityMap{},
			Request{
				Principal: p1,
				Action:    a1,
				Resource:  r1,
				Context:   Variable("context"),
				Variables: Variables{
					"context": []types.Value{
						types.NewRecord(types.RecordMap{"key": types.Long(41)}),
						types.NewRecord(types.RecordMap{"key": types.Long(42)}),
						types.NewRecord(types.RecordMap{"key": types.Long(43)}),
					},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.NewRecord(types.RecordMap{"key": types.Long(41)})}, Decision: false, Values: Values{"context": types.NewRecord(types.RecordMap{"key": types.Long(41)})}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Decision: true, Values: Values{"context": types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.NewRecord(types.RecordMap{"key": types.Long(43)})}, Decision: false, Values: Values{"context": types.NewRecord(types.RecordMap{"key": types.Long(43)})}},
			},
		},

		{"variableContextAccess",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			types.EntityMap{},
			Request{
				Principal: p1,
				Action:    a1,
				Resource:  r1,
				Context: types.NewRecord(types.RecordMap{
					"key": Variable("key"),
				}),
				Variables: Variables{
					"key": []types.Value{types.Long(41), types.Long(42), types.Long(43)},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.NewRecord(types.RecordMap{"key": types.Long(41)})}, Decision: false, Values: Values{"key": types.Long(41)}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.NewRecord(types.RecordMap{"key": types.Long(42)})}, Decision: true, Values: Values{"key": types.Long(42)}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.NewRecord(types.RecordMap{"key": types.Long(43)})}, Decision: false, Values: Values{"key": types.Long(43)}},
			},
		},

		{"ignoreContext",
			ast.Permit().
				When(ast.Context().Access("key").Equal(ast.Long(42))).
				When(ast.Principal().Equal(ast.Value(p1))).
				When(ast.Action().Equal(ast.Value(a1))).
				When(ast.Resource().Equal(ast.Value(r2))),

			types.EntityMap{},
			Request{
				Principal: p1,
				Action:    a1,
				Resource:  Variable("resource"),
				Context:   Ignore(),
				Variables: Variables{
					"resource": []types.Value{r1, r2},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}}, Decision: false, Values: Values{"resource": r1}},
				{Request: types.Request{Principal: p1, Action: a1, Resource: r2, Context: types.Record{}}, Decision: true, Values: Values{"resource": r2}, Diagnostic: types.Diagnostic{Reasons: []types.DiagnosticReason{{PolicyID: "0"}}}},
			},
		},

		{"errors",
			ast.Permit().
				When(ast.String("test").LessThan(ast.Long(42))),
			types.EntityMap{},
			Request{
				Principal: Variable("principal"),
				Action:    Variable("action"),
				Resource:  Variable("resource"),
				Context:   types.Record{},
				Variables: Variables{
					"principal": []types.Value{p1},
					"action":    []types.Value{a1},
					"resource":  []types.Value{r1},
				},
			},
			[]Result{
				{Request: types.Request{Principal: p1, Action: a1, Resource: r1, Context: types.Record{}}, Decision: false,
					Values: Values{"principal": p1, "action": a1, "resource": r1},
					Diagnostic: types.Diagnostic{
						Errors: []types.DiagnosticError{
							{PolicyID: "0", Message: "type error: expected comparable value, got string"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			var res []Result
			ps := cedar.NewPolicySet()
			ps.Add("0", cedar.NewPolicyFromAST((*publicast.Policy)(tt.policy)))

			err := Authorize(context.Background(), ps, tt.entities, tt.request, func(br Result) error {
				// Need to clone this because it could be mutated in successive authorizations
				br.Values = maps.Clone(br.Values)
				res = append(res, br)
				return nil
			})
			testutil.OK(t, err)
			testutil.Equals(t, len(res), len(tt.results))
			for _, a := range tt.results {
				var found bool
				for _, b := range res {
					found = found || reflect.DeepEqual(a, b)
				}
				testutil.FatalIf(t, !found, "missing result: %+v, from: %+v", a, res)
			}
		})
	}
}

func TestBatchErrors(t *testing.T) {
	t.Parallel()
	t.Run("unboundVariables", func(t *testing.T) {
		t.Parallel()
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: Variable("bananas"),
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errUnboundVariable)
	})
	t.Run("unusedVariables", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Variables: Variables{
				"bananas": []types.Value{types.String("test")},
			},
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errUnusedVariable)
	})

	t.Run("nothingTodoNotError", func(t *testing.T) {
		var total int
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: Variable("bananas"),
			Variables: Variables{
				"bananas": nil,
			},
		}, func(_ Result) error { total++; return nil },
		)
		testutil.OK(t, err)
		testutil.Equals(t, total, 0)

	})

	t.Run("missingPrincipal", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: nil,
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  types.NewEntityUID("Resource", "resource"),
			Context:   types.Record{},
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errMissingPart)
	})
	t.Run("missingAction", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    nil,
			Resource:  types.NewEntityUID("Resource", "resource"),
			Context:   types.Record{},
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errMissingPart)
	})
	t.Run("missingPrincipal", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  nil,
			Context:   types.Record{},
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errMissingPart)
	})
	t.Run("missingPrincipal", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  types.NewEntityUID("Resource", "resource"),
			Context:   nil,
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errMissingPart)
	})

	t.Run("contextCancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := Authorize(ctx, cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  types.NewEntityUID("Resource", "resource"),
			Context:   types.Record{},
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, context.Canceled)
	})

	t.Run("firstContextCancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var total int
		err := Authorize(ctx, cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  Variable("resource"),
			Context:   types.Record{},
			Variables: Variables{
				"resource": []types.Value{
					types.NewEntityUID("Resource", "1"),
					types.NewEntityUID("Resource", "2"),
					types.NewEntityUID("Resource", "3"),
				},
			},
		}, func(_ Result) error {
			total++
			cancel()
			return nil
		},
		)
		testutil.ErrorIs(t, err, context.Canceled)
		testutil.Equals(t, total, 1)
	})

	t.Run("lastContextCancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var total int
		err := Authorize(ctx, cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  Variable("resource"),
			Context:   types.Record{},
			Variables: Variables{
				"resource": []types.Value{
					types.NewEntityUID("Resource", "1"),
					types.NewEntityUID("Resource", "2"),
					types.NewEntityUID("Resource", "3"),
				},
			},
		}, func(_ Result) error {
			total++
			if total == 3 {
				cancel()
			}
			return nil
		},
		)
		testutil.ErrorIs(t, err, context.Canceled)
		testutil.Equals(t, total, 3)
	})

	t.Run("callbackErrored", func(t *testing.T) {
		var total int
		errWant := fmt.Errorf("errWant")
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  Variable("resource"),
			Context:   types.Record{},
			Variables: Variables{
				"resource": []types.Value{
					types.NewEntityUID("Resource", "1"),
					types.NewEntityUID("Resource", "2"),
					types.NewEntityUID("Resource", "3"),
				},
			},
		}, func(_ Result) error {
			total++
			if total == 2 {
				return errWant
			}
			return nil
		},
		)
		testutil.ErrorIs(t, err, errWant)
		testutil.Equals(t, total, 2)
	})

	t.Run("contextAndCallbackErrored", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		errWant := fmt.Errorf("errWant")
		err := Authorize(ctx, cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  Variable("resource"),
			Context:   types.Record{},
			Variables: Variables{
				"resource": []types.Value{
					types.NewEntityUID("Resource", "1"),
					types.NewEntityUID("Resource", "2"),
					types.NewEntityUID("Resource", "3"),
				},
			},
		}, func(_ Result) error {
			cancel()
			return errWant
		},
		)
		testutil.ErrorIs(t, err, context.Canceled)
		testutil.ErrorIs(t, err, errWant)
	})

	t.Run("invalidPrincipal", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.String("invalid"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  types.NewEntityUID("Resource", "resource"),
			Context:   types.Record{},
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errInvalidPart)
	})
	t.Run("invalidAction", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.String("invalid"),
			Resource:  types.NewEntityUID("Resource", "resource"),
			Context:   types.Record{},
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errInvalidPart)
	})
	t.Run("invalidPrincipal", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  types.String("invalid"),
			Context:   types.Record{},
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errInvalidPart)
	})
	t.Run("invalidPrincipal", func(t *testing.T) {
		err := Authorize(context.Background(), cedar.NewPolicySet(), types.EntityMap{}, Request{
			Principal: types.NewEntityUID("Principal", "principal"),
			Action:    types.NewEntityUID("Action", "action"),
			Resource:  types.NewEntityUID("Resource", "resource"),
			Context:   types.String("invalid"),
		}, func(_ Result) error { return nil },
		)
		testutil.ErrorIs(t, err, errInvalidPart)
	})

}

func TestIgnoreReasons(t *testing.T) {
	t.Parallel()

	doc := `
	@id("bob0")
	permit (
		principal == Principal::"bob",
		action == Action::"access",
		resource == Resource::"prod"
	)
		when { context has device && context.device == "good" }
	;

	@id("bob1")
	permit (
		principal == Principal::"bob",
		action == Action::"access",
		resource == Resource::"prod"
	)
		when { context has onCall && context.onCall == true }
	;

	@id("bob2")
	forbid (
		principal == Principal::"bob",
		action == Action::"access",
		resource == Resource::"prod"
	)
		when { !(context has device) || context.device == "bad" }
	;

	@id("bob3")
	forbid (
		principal == Principal::"bob",
		action == Action::"access",
		resource == Resource::"prod"
	)
		when { !(context has location) || context.location == "unknown" }
	;

	@id("bob4")
	permit (
		principal == Principal::"bob",
		action == Action::"write",
		resource == Resource::"mitm"
	);

	@id("bob5-condition")
	permit (
		principal,
		action == Action::"write",
		resource == Resource::"mitm"
	)
		when { principal == Principal::"bob" }
	;

	@id("alice0")
	permit (
		principal == Principal::"alice",
		action == Action::"access",
		resource == Resource::"prod"
	)
		when { context has device && context.device == "good" }
	;

	@id("alice1")
	permit (
		principal == Principal::"alice",
		action == Action::"drop",
		resource == Resource::"prod"
	)
		when { context has device && context.device == "good" }
	;

	@id("eve0")
	permit (
		principal == Principal::"eve",
		action == Action::"drop",
		resource == Resource::"mitm"
	)
		when { context has device && context.device == "good" }
	;

	@id("spy0")
	permit (
		principal in Roles::"spy",
		action == Action::"drop",
		resource == Resource::"prod"
	);

	`

	ps := cedar.NewPolicySet()
	pp, err := cedar.NewPolicyListFromBytes("policy.cedar", []byte(doc))
	testutil.OK(t, err)
	for _, p := range pp {
		pid := types.PolicyID(p.Annotations()["id"])
		testutil.FatalIf(t, ps.Get(pid) != nil, "policy already exists: %v", pid)
		if !ps.Add(pid, p) {
			panic("duplicate policy IDs")
		}
	}

	tests := []struct {
		Name     string
		Request  Request
		Total    int
		Decision types.Decision
		Reasons  []types.PolicyID
	}{
		{"when-could-bob-access-prod-ignoring-context",
			Request{
				Principal: types.NewEntityUID("Principal", "bob"),
				Action:    types.NewEntityUID("Action", "access"),
				Resource:  types.NewEntityUID("Resource", "prod"),
				Context:   Ignore(),
			},
			1,
			types.Allow,
			[]types.PolicyID{"bob0", "bob1"},
		},
		{"bob-is-forbidden",
			Request{
				Principal: types.NewEntityUID("Principal", "bob"),
				Action:    types.NewEntityUID("Action", "access"),
				Resource:  types.NewEntityUID("Resource", "prod"),
				Context: types.NewRecord(types.RecordMap{
					"location": types.String("unknown"),
					"device":   types.String("bad"),
				}),
			},
			1,
			types.Deny,
			[]types.PolicyID{"bob2", "bob3"},
		},
		{"can-anyone-access-prod-ignoring-context",
			Request{
				Principal: Ignore(),
				Action:    types.NewEntityUID("Action", "access"),
				Resource:  types.NewEntityUID("Resource", "prod"),
				Context:   Ignore(),
			},
			1,
			types.Allow,
			[]types.PolicyID{"bob0", "bob1", "alice0"},
		},
		{"can-anyone-drop-prod-ignoring-context",
			Request{
				Principal: Ignore(),
				Action:    types.NewEntityUID("Action", "drop"),
				Resource:  types.NewEntityUID("Resource", "prod"),
				Context:   Ignore(),
			},
			1,
			types.Allow,
			[]types.PolicyID{"alice1", "spy0"},
		},
		{"what-permit-policies-relate-to-drops",
			Request{
				Principal: Ignore(),
				Action:    types.NewEntityUID("Action", "drop"),
				Resource:  Ignore(),
				Context:   Ignore(),
			},
			1,
			types.Allow,
			[]types.PolicyID{"alice1", "eve0", "spy0"},
		},
		{"what-permit-policies-relate-to-bob",
			Request{
				Principal: types.NewEntityUID("Principal", "bob"),
				Action:    Ignore(),
				Resource:  Ignore(),
				Context:   Ignore(),
			},
			1,
			types.Allow,
			[]types.PolicyID{"bob0", "bob1", "bob4", "bob5-condition"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var reasons []types.PolicyID
			var total int
			err := Authorize(context.Background(), ps, types.EntityMap{}, tt.Request, func(r Result) error {
				total++
				testutil.Equals(t, r.Decision, tt.Decision)
				for _, v := range r.Diagnostic.Reasons {
					if !slices.Contains(reasons, v.PolicyID) {
						reasons = append(reasons, v.PolicyID)
					}
				}
				return nil
			})
			testutil.OK(t, err)
			testutil.Equals(t, total, tt.Total)
			slices.Sort(reasons)
			slices.Sort(tt.Reasons)
			testutil.Equals(t, reasons, tt.Reasons)
		})
	}
}

func TestCloneSub(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		in    types.Value
		key   types.String
		value types.Value
		out   types.Value
		match bool
	}{
		{
			"variable",
			Variable("bananas"), "bananas", types.String("hello"),
			types.String("hello"), true,
		},
		{
			"record",
			types.NewRecord(types.RecordMap{"key": Variable("bananas")}), "bananas", types.String("hello"),
			types.NewRecord(types.RecordMap{"key": types.String("hello")}), true,
		},
		{
			"set",
			types.NewSet(Variable("bananas")), "bananas", types.String("hello"),
			types.NewSet(types.String("hello")), true,
		},
		{
			"recordNoChange",
			types.NewRecord(types.RecordMap{"key": Variable("asdf")}), "bananas", types.String("hello"),
			types.NewRecord(types.RecordMap{"key": Variable("asdf")}), false,
		},
		{
			"setNoChange",
			types.NewSet(Variable("asdf")), "bananas", types.String("hello"),
			types.NewSet(Variable("asdf")), false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, match := cloneSub(tt.in, tt.key, tt.value)
			testutil.Equals(t, out, tt.out)
			testutil.Equals(t, match, tt.match)
		})
	}
}

func TestFindVariables(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   types.Value
		out  []types.String
	}{
		{"record", types.NewRecord(types.RecordMap{"key": Variable("bananas")}), []types.String{"bananas"}},
		{"set", types.NewSet(Variable("bananas")), []types.String{"bananas"}},
		{"dupes", types.NewSet(Variable("bananas"), Variable("bananas")), []types.String{"bananas"}},
		{"none", types.String("test"), nil},
		{"multi", types.NewSet(Variable("bananas"), Variable("test")), []types.String{"bananas", "test"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := mapset.Make[types.String]()
			findVariables(out, tt.in)
			testutil.Equals(t, out, mapset.FromItems(tt.out...))
		})
	}

}
