package cedar_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func prettifyJson(in []byte) []byte {
	var buf bytes.Buffer
	_ = json.Indent(&buf, in, "", "    ")
	return buf.Bytes()
}

func TestPolicyJSON(t *testing.T) {
	t.Parallel()

	// Taken from https://docs.cedarpolicy.com/policies/json-format.html
	jsonEncodedPolicy := prettifyJson([]byte(`
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

	var policy cedar.Policy
	testutil.OK(t, policy.UnmarshalJSON(jsonEncodedPolicy))

	output, err := policy.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, string(prettifyJson(output)), string(jsonEncodedPolicy))
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

	var policy cedar.Policy
	testutil.OK(t, policy.UnmarshalCedar([]byte(policyStr)))

	testutil.Equals(t, string(policy.MarshalCedar()), policyStr)
}

func TestPolicyAST(t *testing.T) {
	t.Parallel()

	astExample := ast.Permit().
		ActionEq(types.NewEntityUID("Action", "editPhoto")).
		When(ast.Resource().Access("owner").Equals(ast.Principal()))

	_ = cedar.NewPolicyFromAST(astExample)
}
