package templates_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/templates"
)

func TestIsAuthorizedFromLinkedPolicies(t *testing.T) {
	t.Parallel()
	cuzco := cedar.NewEntityUID("coder", "cuzco")
	dropTable := cedar.NewEntityUID("table", "drop")
	tests := []struct {
		Name                        string
		Policy                      string
		LinkEnv                     map[types.SlotID]types.EntityUID
		TemplateID                  cedar.PolicyID
		Entities                    types.EntityGetter
		Principal, Action, Resource cedar.EntityUID
		Context                     cedar.Record
		Want                        cedar.Decision
		DiagErr                     int
		ParseErr                    bool
	}{
		{
			Name:       "simple-permit",
			Policy:     `permit(principal == ?principal,action,resource);`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cuzco},
			Entities:   cedar.EntityMap{},
			Principal:  cuzco,
			Action:     dropTable,
			Resource:   cedar.NewEntityUID("table", "whatever"),
			Context:    cedar.Record{},
			Want:       cedar.Allow,
			DiagErr:    0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			ps, err := templates.NewPolicySetFromBytes("policy.cedar", []byte(tt.Policy))
			testutil.Equals(t, err != nil, tt.ParseErr)

			ps.LinkTemplate(tt.TemplateID, "link0", tt.LinkEnv)

			ok, diag := cedar.Authorize(ps, tt.Entities, cedar.Request{
				Principal: tt.Principal,
				Action:    tt.Action,
				Resource:  tt.Resource,
				Context:   tt.Context,
			})
			testutil.Equals(t, len(diag.Errors), tt.DiagErr)
			testutil.Equals(t, ok, tt.Want)
		})
	}
}
