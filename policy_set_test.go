package cedar

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestNewPolicySetFromFile(t *testing.T) {
	t.Parallel()
	t.Run("err-in-tokenize", func(t *testing.T) {
		t.Parallel()
		_, err := NewPolicySetFromFile("policy.cedar", []byte(`"`))
		testutil.Error(t, err)
	})
	t.Run("err-in-parse", func(t *testing.T) {
		t.Parallel()
		_, err := NewPolicySetFromFile("policy.cedar", []byte(`err`))
		testutil.Error(t, err)
	})
	t.Run("annotations", func(t *testing.T) {
		t.Parallel()
		ps, err := NewPolicySetFromFile("policy.cedar", []byte(`@key("value") permit (principal, action, resource);`))
		testutil.OK(t, err)
		testutil.Equals(t, ps.GetPolicy("policy0").Annotations(), Annotations{"key": "value"})
	})
}

func TestUpsertPolicy(t *testing.T) {
	t.Parallel()
	t.Run("insert", func(t *testing.T) {
		t.Parallel()

		policy0 := NewPolicyFromAST(ast.Forbid())

		var policy1 Policy
		testutil.OK(t, policy1.UnmarshalJSON(
			[]byte(`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`),
		))

		ps := NewPolicySet()
		ps.UpsertPolicy("policy0", policy0)
		ps.UpsertPolicy("policy1", &policy1)

		testutil.Equals(t, ps.GetPolicy("policy0"), policy0)
		testutil.Equals(t, ps.GetPolicy("policy1"), &policy1)
		testutil.Equals(t, ps.GetPolicy("policy2"), nil)
	})
	t.Run("upsert", func(t *testing.T) {
		t.Parallel()

		ps := NewPolicySet()

		p1 := NewPolicyFromAST(ast.Forbid())
		ps.UpsertPolicy("a wavering policy", p1)

		p2 := NewPolicyFromAST(ast.Permit())
		ps.UpsertPolicy("a wavering policy", p2)

		testutil.Equals(t, ps.GetPolicy("a wavering policy"), p2)
	})
}

func TestDeletePolicy(t *testing.T) {
	t.Parallel()
	t.Run("delete non-existent", func(t *testing.T) {
		t.Parallel()

		ps := NewPolicySet()

		// Just verify that this doesn't crash
		ps.DeletePolicy("not a policy")
	})
	t.Run("delete existing", func(t *testing.T) {
		t.Parallel()

		ps := NewPolicySet()

		p1 := NewPolicyFromAST(ast.Forbid())
		ps.UpsertPolicy("a policy", p1)
		ps.DeletePolicy("a policy")

		testutil.Equals(t, ps.GetPolicy("a policy"), nil)
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

	var policies PolicySlice
	testutil.OK(t, policies.UnmarshalCedar([]byte(policiesStr)))

	ps := NewPolicySet()
	for i, p := range policies {
		p.SetSourceFile("example.cedar")
		ps.UpsertPolicy(PolicyID(fmt.Sprintf("policy%d", i)), p)
	}

	testutil.Equals(t, ps.GetPolicy("policy0").Effect(), Permit)
	testutil.Equals(t, ps.GetPolicy("policy1").Effect(), Forbid)

	var buf bytes.Buffer
	ps.MarshalCedar(&buf)

	testutil.Equals(t, buf.String(), policiesStr)
}
