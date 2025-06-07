package templates_test

import (
	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/templates"
	"testing"
)

func TestPolicySetTemplateManagement(t *testing.T) {
	t.Run("template round-trip", func(t *testing.T) {
		policySet := templates.NewPolicySet()

		var templateBody parser.Policy
		templateString := `@id("test_template")
permit (
  principal == ?principal,
  action,
  resource
);`
		testutil.OK(t, templateBody.UnmarshalCedar([]byte(templateString)))
		template := templates.Template(templateBody)

		templateID := cedar.PolicyID("test_template_id")
		added := policySet.AddTemplate(templateID, &template)
		testutil.Equals(t, added, true)

		retrievedTemplate := policySet.GetTemplate(templateID)
		testutil.Equals(t, retrievedTemplate != nil, true)

		originalBytes := template.MarshalCedar()
		retrievedBytes := retrievedTemplate.MarshalCedar()
		testutil.Equals(t, string(retrievedBytes), string(originalBytes))

		removed := policySet.RemoveTemplate(templateID)
		testutil.Equals(t, removed, true)

		retrievedTemplateAfterRemoval := policySet.GetTemplate(templateID)
		testutil.Equals(t, retrievedTemplateAfterRemoval, (*templates.Template)(nil))
	})

	t.Run("remove non-existent template", func(t *testing.T) {
		policySet := templates.NewPolicySet()
		templateID := cedar.PolicyID("non_existent_template")
		removed := policySet.RemoveTemplate(templateID)
		testutil.Equals(t, removed, false)
	})

	t.Run("add template with existing ID", func(t *testing.T) {
		policySet := templates.NewPolicySet()
		templateID := cedar.PolicyID("duplicate_template_id")

		var templateBody parser.Policy
		templateString := `@id("test_template")
permit (
  principal,
  action,
  resource
);`
		testutil.OK(t, templateBody.UnmarshalCedar([]byte(templateString)))
		template := templates.Template(templateBody)

		// First add should succeed
		isNew := policySet.AddTemplate(templateID, &template)
		testutil.Equals(t, isNew, true)

		// Second add with same ID should return false
		isNew = policySet.AddTemplate(templateID, &template)
		testutil.Equals(t, isNew, false)
	})

	t.Run("cannot use link id already used by static policy", func(t *testing.T) {
		templateString := `permit (
  principal == ?principal,
  action,
  resource
);

permit (
  principal,
  action,
  resource
);`
		templateID := cedar.PolicyID("template0")
		policyID := cedar.PolicyID("policy0")

		policySet, err := templates.NewPolicySetFromBytes("test.cedar", []byte(templateString))
		testutil.OK(t, err)

		// Link a policy to the template
		//linkID := cedar.PolicyID("linked_policy_id")
		env := map[types.SlotID]types.EntityUID{
			"?principal": types.NewEntityUID("User", "alice"),
		}
		err = policySet.LinkTemplate(templateID, policyID, env)
		testutil.Error(t, err)
	})

	t.Run("removing template removes linked policies", func(t *testing.T) {
		templateString := `permit (
  principal == ?principal,
  action,
  resource
);`
		templateID := cedar.PolicyID("template0")

		policySet, err := templates.NewPolicySetFromBytes("test.cedar", []byte(templateString))
		testutil.OK(t, err)

		// Link a policy to the template
		linkID := cedar.PolicyID("linked_policy_id")
		env := map[types.SlotID]types.EntityUID{
			"?principal": types.NewEntityUID("User", "alice"),
		}
		err = policySet.LinkTemplate(templateID, linkID, env)
		testutil.OK(t, err)

		// Ensure the linked policy exists
		linkedPolicy := policySet.GetLinkedPolicy(linkID)
		testutil.Equals(t, linkedPolicy != nil, true)

		// Remove the template
		removed := policySet.RemoveTemplate(templateID)
		testutil.Equals(t, removed, true)

		// The linked policy should also be removed
		linkedPolicyAfterRemoval := policySet.GetLinkedPolicy(linkID)
		testutil.Equals(t, linkedPolicyAfterRemoval == nil, true)
	})

	t.Run("remove method can also remove linked policy", func(t *testing.T) {
		templateString := `permit (
  principal == ?principal,
  action,
  resource
);`
		templateID := cedar.PolicyID("template0")

		policySet, err := templates.NewPolicySetFromBytes("test.cedar", []byte(templateString))
		testutil.OK(t, err)

		// Link a policy to the template
		linkID := cedar.PolicyID("linked_policy_id")
		env := map[types.SlotID]types.EntityUID{
			"?principal": types.NewEntityUID("User", "alice"),
		}
		err = policySet.LinkTemplate(templateID, linkID, env)
		testutil.OK(t, err)

		// Ensure the linked policy exists
		linkedPolicy := policySet.GetLinkedPolicy(linkID)
		testutil.Equals(t, linkedPolicy != nil, true)

		// Remove the template
		removed := policySet.Remove(linkID)
		testutil.Equals(t, removed, true)

		// The linked policy should also be removed
		linkedPolicyAfterRemoval := policySet.GetLinkedPolicy(linkID)
		testutil.Equals(t, linkedPolicyAfterRemoval == nil, true)
	})
}

func TestLinkTemplateToPolicy(t *testing.T) {
	linkTests := []struct {
		Name           string
		TemplateString string
		LinkID         cedar.PolicyID
		Env            map[types.SlotID]types.EntityUID
		Want           string
	}{
		{
			"principal ScopeTypeEq",
			`permit (
  principal == ?principal,
  action,
  resource
);`,
			"scope_eq_link",
			map[types.SlotID]types.EntityUID{"?principal": types.NewEntityUID("User", "bob")},
			`{"effect":"permit","principal":{"op":"==","entity":{"type":"User","id":"bob"}},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"principal ScopeTypeIn",
			`permit (
  principal in ?principal,
  action,
  resource
);`,
			"scope_in_link",
			map[types.SlotID]types.EntityUID{"?principal": types.NewEntityUID("User", "charlie")},
			`{"effect":"permit","principal":{"op":"in","entity":{"type":"User","id":"charlie"}},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"principal ScopeTypeIsIn",
			`permit (
  principal is User in ?principal,
  action,
  resource
);`,
			"scope_isin_link",
			map[types.SlotID]types.EntityUID{"?principal": types.NewEntityUID("User", "dave")},
			`{"effect":"permit","principal":{"op":"is","entity_type":"User","in":{"entity":{"type":"User","id":"dave"}}},"action":{"op":"All"},"resource":{"op":"All"}}`,
		},
		{
			"resource ScopeTypeEq",
			`permit (
  principal,
  action,
  resource == ?resource
);`,
			"scope_eq_link",
			map[types.SlotID]types.EntityUID{"?resource": types.NewEntityUID("Album", "trip")},
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"==","entity":{"type":"Album","id":"trip"}}}`,
		},
		{
			"resource ScopeTypeIn",
			`permit (
  principal,
  action,
  resource in ?resource
);`,
			"scope_in_link",
			map[types.SlotID]types.EntityUID{"?resource": types.NewEntityUID("Album", "trip")},
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"in","entity":{"type":"Album","id":"trip"}}}`,
		},
		{
			"resource ScopeTypeIsIn",
			`permit (
  principal,
  action,
  resource is Album in ?resource
);`,
			"scope_isin_link",
			map[types.SlotID]types.EntityUID{"?resource": types.NewEntityUID("Album", "trip")},
			`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"is","entity_type":"Album","in":{"entity":{"type":"Album","id":"trip"}}}}`,
		},
	}

	for _, tt := range linkTests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			policySet, err := templates.NewPolicySetFromBytes("test.cedar", []byte(tt.TemplateString))
			testutil.OK(t, err)

			templateID := cedar.PolicyID("template0")

			err = policySet.LinkTemplate(templateID, tt.LinkID, tt.Env)
			testutil.OK(t, err)

			linkedPolicy := policySet.GetLinkedPolicy(tt.LinkID)

			testutil.Equals(t, linkedPolicy.LinkID(), tt.LinkID)
			testutil.Equals(t, linkedPolicy.TemplateID(), templateID)

			for policyID, policy := range policySet.All() {
				if policyID == tt.LinkID {
					pj, err := policy.MarshalJSON()
					testutil.OK(t, err)

					testutil.Equals(t, string(pj), tt.Want)

					break
				}
			}
		})
	}
}
