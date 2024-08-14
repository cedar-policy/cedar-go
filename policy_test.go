package cedar

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
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

	var policy Policy
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

	var policy Policy
	testutil.OK(t, policy.UnmarshalCedar([]byte(policyStr)))

	var buf bytes.Buffer
	policy.MarshalCedar(&buf)

	testutil.Equals(t, buf.String(), policyStr)
}
