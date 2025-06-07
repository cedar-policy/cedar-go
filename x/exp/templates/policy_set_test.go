package templates_test

import (
	"fmt"
	"maps"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/templates"
)

func TestPolicyMap(t *testing.T) {
	t.Parallel()
	t.Run("All", func(t *testing.T) {
		t.Parallel()
		pm := templates.PolicyMap{
			"foo": templates.NewPolicyFromAST(ast.Permit()),
			"bar": templates.NewPolicyFromAST(ast.Permit()),
		}

		got := maps.Collect(pm.All())
		testutil.Equals(t, got, pm)
	})
}

func TestNewPolicySetFromFile(t *testing.T) {
	t.Parallel()
	t.Run("err-in-tokenize", func(t *testing.T) {
		t.Parallel()
		_, err := templates.NewPolicySetFromBytes("policy.cedar", []byte(`"`))
		testutil.Error(t, err)
	})
	t.Run("err-in-parse", func(t *testing.T) {
		t.Parallel()
		_, err := templates.NewPolicySetFromBytes("policy.cedar", []byte(`err`))
		testutil.Error(t, err)
	})
	t.Run("annotations", func(t *testing.T) {
		t.Parallel()
		ps, err := templates.NewPolicySetFromBytes("policy.cedar", []byte(`@key("value") permit (principal, action, resource);`))
		testutil.OK(t, err)
		testutil.Equals(t, ps.Get("policy0").Annotations(), cedar.Annotations{"key": "value"})
	})
}

func TestUpsertPolicy(t *testing.T) {
	t.Parallel()
	t.Run("insert", func(t *testing.T) {
		t.Parallel()

		policy0 := templates.NewPolicyFromAST(ast.Forbid())

		var policy1 templates.Policy
		testutil.OK(t, policy1.UnmarshalJSON(
			[]byte(`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`),
		))

		ps := templates.NewPolicySet()
		added := ps.Add("policy0", policy0)
		testutil.Equals(t, added, true)
		added = ps.Add("policy1", &policy1)
		testutil.Equals(t, added, true)

		testutil.Equals(t, ps.Get("policy0"), policy0)
		testutil.Equals(t, ps.Get("policy1"), &policy1)
		testutil.Equals(t, ps.Get("policy2"), nil)
	})
	t.Run("upsert", func(t *testing.T) {
		t.Parallel()

		ps := templates.NewPolicySet()

		p1 := templates.NewPolicyFromAST(ast.Forbid())
		ps.Add("a wavering policy", p1)

		p2 := templates.NewPolicyFromAST(ast.Permit())
		added := ps.Add("a wavering policy", p2)
		testutil.Equals(t, added, false)

		testutil.Equals(t, ps.Get("a wavering policy"), p2)
	})
}

func TestDeletePolicy(t *testing.T) {
	t.Parallel()
	t.Run("delete non-existent", func(t *testing.T) {
		t.Parallel()

		ps := templates.NewPolicySet()

		existed := ps.Remove("not a policy")
		testutil.Equals(t, existed, false)
	})
	t.Run("delete existing", func(t *testing.T) {
		t.Parallel()

		ps := templates.NewPolicySet()

		p1 := templates.NewPolicyFromAST(ast.Forbid())
		ps.Add("a policy", p1)
		existed := ps.Remove("a policy")
		testutil.Equals(t, existed, true)

		testutil.Equals(t, ps.Get("a policy"), nil)
	})
}

func TestNewPolicySetFromSlice(t *testing.T) {
	t.Parallel()

	policiesStr := `permit (
    principal,
    action == Action::"editPhoto",
    resource
)
when { resource.owner == principal };

forbid (
    principal in Groups::"bannedUsers",
    action,
    resource
);`

	policies, err := templates.NewPolicyListFromBytes("", []byte(policiesStr))
	testutil.OK(t, err)

	ps := templates.NewPolicySet()
	for i, p := range policies.StaticPolicies {
		p.SetFilename("example.cedar")
		ps.Add(cedar.PolicyID(fmt.Sprintf("policy%d", i)), p)
	}

	testutil.Equals(t, ps.Get("policy0").Effect(), cedar.Permit)
	testutil.Equals(t, ps.Get("policy1").Effect(), cedar.Forbid)

	testutil.Equals(t, string(ps.MarshalCedar()), policiesStr)

}

func TestPolicySetMap(t *testing.T) {
	t.Parallel()
	ps, err := templates.NewPolicySetFromBytes("", []byte(`permit (principal, action, resource);`))
	testutil.OK(t, err)
	m := maps.Collect(ps.All())
	testutil.Equals(t, len(m), 1)
}

func TestPolicySetJSON(t *testing.T) {
	t.Parallel()
	t.Run("UnmarshalError", func(t *testing.T) {
		t.Parallel()
		var ps templates.PolicySet
		err := ps.UnmarshalJSON([]byte(`!@#$`))
		testutil.Error(t, err)
	})

	t.Run("UnmarshalOK", func(t *testing.T) {
		t.Parallel()
		var ps templates.PolicySet
		err := ps.UnmarshalJSON([]byte(`{"staticPolicies":{"policy0":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}},"templates":{"template0":{"effect":"permit","principal":{"op":"==","slot":"?principal"},"action":{"op":"All"},"resource":{"op":"All"}}},"templateLinks":[{"templateId":"template0","newId":"linked0","values":{"?principal":{"type":"User","id":"alice"}}}]}`))
		testutil.OK(t, err)
		testutil.Equals(t, len(maps.Collect(ps.All())), 2)
		testutil.Equals(t, ps.GetTemplate("template0") != nil, true)
		testutil.Equals(t, ps.GetLinkedPolicy("linked0") != nil, true)
	})

	t.Run("MarshalOK", func(t *testing.T) {
		t.Parallel()
		ps, err := templates.NewPolicySetFromBytes("", []byte(`permit (principal, action, resource);

permit (principal == ?principal, action, resource);`))
		testutil.OK(t, err)

		err = ps.LinkTemplate("template0", "linked0", map[types.SlotID]types.EntityUID{
			"?principal": types.NewEntityUID("User", "alice"),
		})
		testutil.OK(t, err)

		out, err := ps.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(out), `{"staticPolicies":{"policy0":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]}},"templates":{"template0":{"effect":"permit","principal":{"op":"==","slot":"?principal"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]}},"templateLinks":[{"templateId":"template0","newId":"linked0","values":{"?principal":{"type":"User","id":"alice"}}}]}`)
	})
}

func TestAll(t *testing.T) {
	t.Parallel()
	t.Run("all", func(t *testing.T) {
		t.Parallel()

		policies := map[cedar.PolicyID]*templates.Policy{
			"policy0": templates.NewPolicyFromAST(ast.Forbid()),
			"policy1": templates.NewPolicyFromAST(ast.Forbid()),
			"policy2": templates.NewPolicyFromAST(ast.Forbid()),
		}

		ps := templates.NewPolicySet()
		for k, v := range policies {
			ps.Add(k, v)
		}

		got := map[cedar.PolicyID]*templates.Policy{}
		for k, v := range ps.All() {
			got[k] = v
		}

		testutil.Equals(t, policies, got)
	})

	t.Run("break early", func(t *testing.T) {
		t.Parallel()

		policies := map[cedar.PolicyID]*templates.Policy{
			"policy0": templates.NewPolicyFromAST(ast.Forbid()),
			"policy1": templates.NewPolicyFromAST(ast.Forbid()),
			"policy2": templates.NewPolicyFromAST(ast.Forbid()),
		}

		ps := templates.NewPolicySet()
		for k, v := range policies {
			ps.Add(k, v)
		}

		got := map[cedar.PolicyID]*templates.Policy{}
		for k, v := range ps.All() {
			got[k] = v
			if len(got) == 2 {
				break
			}
		}

		testutil.Equals(t, len(got), 2)
		for k, v := range got {
			testutil.Equals(t, policies[k], v)
		}
	})
}

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
