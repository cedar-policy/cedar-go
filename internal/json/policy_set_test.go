package json

import (
	"testing"

	"github.com/go-json-experiment/json"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestPolicySetJSON(t *testing.T) {
	t.Parallel()

	t.Run("MarshalSorted", func(t *testing.T) {
		t.Parallel()
		staticPolicies := PolicySet{}
		for _, id := range []string{"c", "a", "b"} {
			staticPolicies[id] = Policy(*ast.Permit())
		}
		ps := PolicySetJSON{StaticPolicies: &staticPolicies}
		out, err := json.Marshal(&ps)
		testutil.OK(t, err)
		testutil.Equals(t, string(out), `{"staticPolicies":{"a":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]},"b":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]},"c":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]}},"templates":{},"templateLinks":[]}`)
	})

	t.Run("UnmarshalLenient", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name  string
			input string
		}{
			// cedar-go previously omitted "templates" and "templateLinks".
			{"missingTemplates", `{"staticPolicies":{"p":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]}}}`},
			{"emptyTemplates", `{"staticPolicies":{"p":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]}},"templates":{},"templateLinks":[]}`},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var ps PolicySetJSON
				testutil.OK(t, json.Unmarshal([]byte(tt.input), &ps))
				testutil.Equals(t, *ps.StaticPolicies, PolicySet{"p": Policy(*ast.Permit())})
			})
		}
	})

	t.Run("UnmarshalErrors", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name  string
			input string
		}{
			{"invalidJSON", `{`},
			{"missingStaticPolicies", `{"templates":{},"templateLinks":[]}`},
			{"nullStaticPolicies", `{"staticPolicies":null}`},
			{"nullPolicy", `{"staticPolicies":{"p":null}}`},
			{"duplicatePolicyID", `{"staticPolicies":{"p":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]},"p":{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]}}}`},
			{"unknownMember", `{"staticPolicies":{},"x":1}`},
			{"nullTemplates", `{"staticPolicies":{},"templates":null}`},
			{"nullTemplateLinks", `{"staticPolicies":{},"templateLinks":null}`},
			{"templates", `{"staticPolicies":{},"templates":{"t":{"effect":"permit","principal":{"op":"==","slot":"?principal"},"action":{"op":"All"},"resource":{"op":"All"},"conditions":[]}},"templateLinks":[]}`},
			{"templateLinks", `{"staticPolicies":{},"templates":{},"templateLinks":[{"templateId":"t","newId":"l","values":{"?principal":{"type":"U","id":"a"}}}]}`},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var ps PolicySetJSON
				testutil.Error(t, json.Unmarshal([]byte(tt.input), &ps))
			})
		}
	})
}
