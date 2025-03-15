package cedar_test

import (
	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"testing"
)

func TestLinkTemplateToPolicy(t *testing.T) {
	linkTests := []struct {
		Name           string
		TemplateString string
		TemplateID     string
		LinkID         string
		Env            map[types.SlotID]types.EntityUID
		Want           string
	}{

		{
			"principal ScopeTypeEq",
			`@id("scope_eq_test")
permit (
   principal == ?principal,
   action,
   resource
);`,
			"scope_eq_test",
			"scope_eq_link",
			map[types.SlotID]types.EntityUID{"?principal": types.NewEntityUID("User", "bob")},
			`{"annotations":{"id":"scope_eq_test"},"effect":"permit","principal":{"op":"==","entity":{"type":"User","id":"bob"}},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},

		{
			"principal ScopeTypeIn",
			`@id("scope_in_test")
permit (
   principal in ?principal,
   action,
   resource
);`,
			"scope_in_test",
			"scope_in_link",
			map[types.SlotID]types.EntityUID{"?principal": types.NewEntityUID("User", "charlie")},
			`{"annotations":{"id":"scope_in_test"},"effect":"permit","principal":{"op":"in","entity":{"type":"User","id":"charlie"}},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"principal ScopeTypeIsIn",
			`@id("scope_isin_test")
permit (
   principal is User in ?principal,
   action,
   resource
);`,
			"scope_isin_test",
			"scope_isin_link",
			map[types.SlotID]types.EntityUID{"?principal": types.NewEntityUID("User", "dave")},
			`{"annotations":{"id":"scope_isin_test"},"effect":"permit","principal":{"op":"is","entity_type":"User","in":{"entity":{"type":"User","id":"dave"}}},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"resource ScopeTypeEq",
			`@id("resource_scope_eq_test")
permit (
   principal,
   action,
   resource == ?resource
);`,
			"resource_scope_eq_test",
			"scope_eq_link",
			map[types.SlotID]types.EntityUID{"?resource": types.NewEntityUID("Album", "trip")},
			`{"annotations":{"id":"resource_scope_eq_test"},"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"==","entity":{"type":"Album","id":"trip"}}}`,
		},
		{
			"resource ScopeTypeIn",
			`@id("resource_scope_in_test")
permit (
   principal,
   action,
   resource in ?resource
);`,
			"resource_scope_in_test",
			"scope_in_link",
			map[types.SlotID]types.EntityUID{"?resource": types.NewEntityUID("Album", "trip")},
			`{"annotations":{"id":"resource_scope_in_test"},"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"in","entity":{"type":"Album","id":"trip"}}}`,
		},
		{
			"resource ScopeTypeIsIn",
			`@id("resource_scope_isin_test")
permit (
   principal,
   action,
   resource is Album in ?resource
);`,
			"resource_scope_isin_test",
			"scope_isin_link",
			map[types.SlotID]types.EntityUID{"?resource": types.NewEntityUID("Album", "trip")},
			`{"annotations":{"id":"resource_scope_isin_test"},"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"is","entity_type":"Album","in":{"entity":{"type":"Album","id":"trip"}}}}`,
		},
	}

	for _, tt := range linkTests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var templateBody parser.Policy
			testutil.OK(t, templateBody.UnmarshalCedar([]byte(tt.TemplateString)))
			template := cedar.Template(templateBody)

			linkedPolicy := cedar.LinkTemplate(template, tt.TemplateID, tt.LinkID, tt.Env)

			testutil.Equals(t, linkedPolicy.LinkID, tt.LinkID)
			testutil.Equals(t, linkedPolicy.TemplateID, tt.TemplateID)

			policy, err := linkedPolicy.Render()
			testutil.OK(t, err)

			pj, err := policy.MarshalJSON()
			testutil.OK(t, err)

			testutil.Equals(t, string(pj), tt.Want)
		})
	}
}
