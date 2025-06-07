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
		LinkErr                     bool
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
		{
			Name:       "simple-forbid",
			Policy:     `forbid(principal == ?principal,action,resource);`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cuzco},
			Entities:   cedar.EntityMap{},
			Principal:  cuzco,
			Action:     dropTable,
			Resource:   cedar.NewEntityUID("table", "whatever"),
			Context:    cedar.Record{},
			Want:       cedar.Deny,
			DiagErr:    0,
		},
		{
			Name:       "permit-resource-equals",
			Policy:     `permit(principal,action,resource == ?resource);`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?resource": cedar.NewEntityUID("table", "whatever")},
			Entities:   cedar.EntityMap{},
			Principal:  cuzco,
			Action:     dropTable,
			Resource:   cedar.NewEntityUID("table", "whatever"),
			Context:    cedar.Record{},
			Want:       cedar.Allow,
			DiagErr:    0,
		},
		{
			Name:       "permit-when-in-hierarchy",
			Policy:     `permit(principal in ?principal,action,resource);`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cedar.NewEntityUID("team", "osiris")},
			Entities: cedar.EntityMap{
				cuzco: cedar.Entity{
					UID:     cuzco,
					Parents: cedar.NewEntityUIDSet(cedar.NewEntityUID("team", "osiris")),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      cedar.Allow,
			DiagErr:   0,
		},
		{
			Name:       "permit-when-condition",
			Policy:     `permit(principal == ?principal,action,resource) when { context.x == 42 };`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cuzco},
			Entities:   cedar.EntityMap{},
			Principal:  cuzco,
			Action:     dropTable,
			Resource:   cedar.NewEntityUID("table", "whatever"),
			Context:    cedar.NewRecord(cedar.RecordMap{"x": cedar.Long(42)}),
			Want:       cedar.Allow,
			DiagErr:    0,
		},
		{
			Name:       "permit-when-condition-fails",
			Policy:     `permit(principal == ?principal,action,resource) when { context.x == 42 };`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cuzco},
			Entities:   cedar.EntityMap{},
			Principal:  cuzco,
			Action:     dropTable,
			Resource:   cedar.NewEntityUID("table", "whatever"),
			Context:    cedar.NewRecord(cedar.RecordMap{"x": cedar.Long(43)}),
			Want:       cedar.Deny,
			DiagErr:    0,
		},
		{
			Name:       "permit-requires-entities",
			Policy:     `permit(principal == ?principal,action,resource) when { principal.x == 42 };`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cuzco},
			Entities: cedar.EntityMap{
				cuzco: cedar.Entity{
					UID:        cuzco,
					Attributes: cedar.NewRecord(cedar.RecordMap{"x": cedar.Long(42)}),
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      cedar.Allow,
			DiagErr:   0,
		},
		{
			Name:       "multiple-slots-without-action",
			Policy:     `permit(principal == ?principal,action,resource == ?resource);`,
			TemplateID: "template0",
			LinkEnv: map[types.SlotID]types.EntityUID{
				"?principal": cuzco,
				"?resource":  cedar.NewEntityUID("table", "whatever"),
			},
			Entities:  cedar.EntityMap{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      cedar.Allow,
			DiagErr:   0,
		},
		{
			Name:       "incorrect-env-size",
			Policy:     `permit(principal == ?principal,action,resource == ?resource);`,
			TemplateID: "template0",
			LinkEnv: map[types.SlotID]types.EntityUID{
				"?principal": cuzco,
				// Missing ?resource slot
			},
			Entities:  cedar.EntityMap{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      cedar.Deny,
			LinkErr:   true,
		},
		{
			Name:       "missing-template-slot",
			Policy:     `permit(principal == ?principal,action,resource == ?resource);`,
			TemplateID: "template0",
			LinkEnv: map[types.SlotID]types.EntityUID{
				"?resource": cedar.NewEntityUID("table", "whatever"),
			},
			Entities:  cedar.EntityMap{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  cedar.NewEntityUID("table", "whatever"),
			Context:   cedar.Record{},
			Want:      cedar.Deny,
			LinkErr:   true,
		},
		{
			Name:       "error-in-policy",
			Policy:     `permit(principal == ?principal,action,resource) when { resource in "foo" };`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cuzco},
			Entities:   cedar.EntityMap{},
			Principal:  cuzco,
			Action:     dropTable,
			Resource:   cedar.NewEntityUID("table", "whatever"),
			Context:    cedar.Record{},
			Want:       cedar.Deny,
			DiagErr:    1,
		},
		{
			Name:       "permit-unless",
			Policy:     `permit(principal == ?principal,action,resource) unless { context.x > 100 };`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cuzco},
			Entities:   cedar.EntityMap{},
			Principal:  cuzco,
			Action:     dropTable,
			Resource:   cedar.NewEntityUID("table", "whatever"),
			Context:    cedar.NewRecord(cedar.RecordMap{"x": cedar.Long(50)}),
			Want:       cedar.Allow,
			DiagErr:    0,
		},
		{
			Name:       "variable-used-in-wrong-place",
			Policy:     `permit(principal is coder,action,resource) when { principal == ?principal };`,
			TemplateID: "template0",
			LinkEnv:    map[types.SlotID]types.EntityUID{"?principal": cuzco},
			Entities:   cedar.EntityMap{},
			Principal:  cuzco,
			Action:     dropTable,
			Resource:   cedar.NewEntityUID("table", "whatever"),
			Context:    cedar.Record{},
			Want:       cedar.Deny,
			DiagErr:    0,
			ParseErr:   true,
			LinkErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			ps, err := templates.NewPolicySetFromBytes("policy.cedar", []byte(tt.Policy))
			testutil.Equals(t, err != nil, tt.ParseErr)

			err = ps.LinkTemplate(tt.TemplateID, "link0", tt.LinkEnv)
			testutil.Equals(t, err != nil, tt.LinkErr)

			ok, diag := templates.Authorize(ps, tt.Entities, cedar.Request{
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
