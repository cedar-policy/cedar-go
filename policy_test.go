package cedar_test

import (
	"bytes"
	"encoding/json"
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

	var policy cedar.Policy
	testutil.OK(t, policy.UnmarshalJSON(jsonEncodedPolicy))

	output, err := policy.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, string(prettifyJSON(output)), string(jsonEncodedPolicy))
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
		ActionEq(cedar.NewEntityUID("Action", "editPhoto")).
		When(ast.Resource().Access("owner").Equal(ast.Principal()))

	_ = cedar.NewPolicyFromAST(astExample)
}

func TestUnmarshalJSONPolicyErr(t *testing.T) {
	t.Parallel()
	var p cedar.Policy
	err := p.UnmarshalJSON([]byte("!@#$"))
	testutil.Error(t, err)
}

func TestUnmarshalCedarPolicyErr(t *testing.T) {
	t.Parallel()
	var p cedar.Policy
	err := p.UnmarshalCedar([]byte("!@#$"))
	testutil.Error(t, err)
}

func TestPositionJSON(t *testing.T) {
	t.Parallel()
	p := cedar.Position{Filename: "foo.cedar", Offset: 1, Line: 2, Column: 3}

	marshaled, err := json.MarshalIndent(p, "", "\t")
	testutil.OK(t, err)

	var want bytes.Buffer
	_ = json.Indent(&want, []byte(`{ "filename": "foo.cedar", "offset": 1, "line": 2, "column": 3 }`), "", "\t")

	testutil.Equals(t, string(marshaled), want.String())

	var unmarshaled cedar.Position
	testutil.OK(t, json.Unmarshal(want.Bytes(), &unmarshaled))

	testutil.Equals(t, unmarshaled, p)
}
