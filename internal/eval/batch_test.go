package eval

import (
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
		results  []BatchResult
	}{
		{"smokeTest",
			ast.Permit(),
			types.Entities{},
			BatchRequest{
				Principals: []types.EntityUID{p1},
				Actions:    []types.EntityUID{a1, a2},
				Resources:  []types.EntityUID{r1, r2, r3},
			},
			[]BatchResult{
				{Principal: p1, Action: a1, Resource: r1, Decision: true},
				{Principal: p1, Action: a1, Resource: r2, Decision: true},
				{Principal: p1, Action: a1, Resource: r3, Decision: true},
				{Principal: p1, Action: a2, Resource: r1, Decision: true},
				{Principal: p1, Action: a2, Resource: r2, Decision: true},
				{Principal: p1, Action: a2, Resource: r3, Decision: true},
			},
		},

		{"someOk",
			ast.Permit().PrincipalEq(p1).ActionEq(a2).ResourceEq(r3),
			types.Entities{},
			BatchRequest{
				Principals: []types.EntityUID{p1},
				Actions:    []types.EntityUID{a1, a2},
				Resources:  []types.EntityUID{r1, r2, r3},
			},
			[]BatchResult{
				{Principal: p1, Action: a1, Resource: r1, Decision: false},
				{Principal: p1, Action: a1, Resource: r2, Decision: false},
				{Principal: p1, Action: a1, Resource: r3, Decision: false},
				{Principal: p1, Action: a2, Resource: r1, Decision: false},
				{Principal: p1, Action: a2, Resource: r2, Decision: false},
				{Principal: p1, Action: a2, Resource: r3, Decision: true},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var res []BatchResult
			Batch([]*ast.Policy{tt.policy}, tt.entities, tt.request, func(br BatchResult) { res = append(res, br) })
			testutil.Equals(t, res, tt.results)
		})
	}

}
