package templates_test

import (
	"bytes"
	"encoding/json"
	"github.com/cedar-policy/cedar-go/x/exp/templates"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func prettifyJSON(in []byte) []byte {
	var buf bytes.Buffer
	_ = json.Indent(&buf, in, "", "    ")
	return buf.Bytes()
}

func TestPolicyJSON(t *testing.T) {
	t.Parallel()

	// Taken from https://docs.cedarpolicy.com/policies/json-format.html
	jsonEncodedPolicy := prettifyJSON([]byte(`
		{
			"effect": "permit",
			"principal": {
				"op": "==",
				"entity": { "type": "User", "id": "12UA45" }
			},
			"action": {
				"op": "==",
				"entity": { "type": "Action", "id": "view" }
			},
			"resource": {
				"op": "in",
				"entity": { "type": "Folder", "id": "abc" }
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
	))

	var policy templates.Policy
	testutil.OK(t, policy.UnmarshalJSON(jsonEncodedPolicy))

	output, err := policy.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, string(prettifyJSON(output)), string(jsonEncodedPolicy))
}

func TestTemplateJSON(t *testing.T) {
	t.Parallel()

	// Taken from https://docs.cedarpolicy.com/policies/json-format.html
	jsonEncodedTemplate := prettifyJSON([]byte(`
		{
            "effect": "forbid",
            "principal": {
                "op": "==",
                "entity": { "type": "User", "id": "12UA45" }
            },
            "action": {
                "op": "==",
                "entity": { "type": "Action", "id": "view" }
            },
            "resource": {
                "op": "in",
                "slot": "?resource"
            },
            "conditions": []
        }`,
	))

	var policy templates.Template
	testutil.OK(t, policy.UnmarshalJSON(jsonEncodedTemplate))

	output, err := policy.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, string(prettifyJSON(output)), string(jsonEncodedTemplate))
}

func TestPolicyCedar(t *testing.T) {
	t.Parallel()

	// Taken from https://docs.cedarpolicy.com/policies/syntax-policy.html
	policyStr := `permit (
    principal,
    action == Action::"editPhoto",
    resource
)
when { resource.owner == principal };`

	var policy templates.Policy
	testutil.OK(t, policy.UnmarshalCedar([]byte(policyStr)))

	testutil.Equals(t, string(policy.MarshalCedar()), policyStr)
}

func TestTemplateCedar(t *testing.T) {
	t.Parallel()

	policyStr := `permit (
    principal == ?principal,
    action,
    resource == ?resource
)
when { resource.owner == principal };`

	var policy templates.Template
	testutil.OK(t, policy.UnmarshalCedar([]byte(policyStr)))

	testutil.Equals(t, string(policy.MarshalCedar()), policyStr)
}

func TestPolicyAST(t *testing.T) {
	t.Parallel()

	astExample := ast.Permit().
		ActionEq(cedar.NewEntityUID("Action", "editPhoto")).
		When(ast.Resource().Access("owner").Equal(ast.Principal()))

	_ = templates.NewPolicyFromAST(astExample)
}

func TestUnmarshalJSONPolicyErr(t *testing.T) {
	t.Parallel()
	var p templates.Policy
	err := p.UnmarshalJSON([]byte("!@#$"))
	testutil.Error(t, err)
}

func TestUnmarshalCedarPolicyErr(t *testing.T) {
	t.Parallel()
	var p templates.Policy
	err := p.UnmarshalCedar([]byte("!@#$"))
	testutil.Error(t, err)
}
