package cedar_test

import (
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestNewPolicySetFromFile(t *testing.T) {
	t.Parallel()
	t.Run("err-in-tokenize", func(t *testing.T) {
		t.Parallel()
		_, err := cedar.NewPolicySetFromBytes("policy.cedar", []byte(`"`))
		testutil.Error(t, err)
	})
	t.Run("err-in-parse", func(t *testing.T) {
		t.Parallel()
		_, err := cedar.NewPolicySetFromBytes("policy.cedar", []byte(`err`))
		testutil.Error(t, err)
	})
	t.Run("annotations", func(t *testing.T) {
		t.Parallel()
		ps, err := cedar.NewPolicySetFromBytes("policy.cedar", []byte(`@key("value") permit (principal, action, resource);`))
		testutil.OK(t, err)
		testutil.Equals(t, ps.Get("policy0").Annotations(), cedar.Annotations{"key": "value"})
	})
}

func TestUpsertPolicy(t *testing.T) {
	t.Parallel()
	t.Run("insert", func(t *testing.T) {
		t.Parallel()

		policy0 := cedar.NewPolicyFromAST(ast.Forbid())

		var policy1 cedar.Policy
		testutil.OK(t, policy1.UnmarshalJSON(
			[]byte(`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`),
		))

		ps := cedar.NewPolicySet()
		ps.Set("policy0", policy0)
		ps.Set("policy1", policy1)

		testutil.Equals(t, ps.Get("policy0"), policy0)
		testutil.Equals(t, ps.Get("policy1"), policy1)
		testutil.Equals(t, ps.Get("policy2"), cedar.Policy{})
	})
	t.Run("upsert", func(t *testing.T) {
		t.Parallel()

		ps := cedar.NewPolicySet()

		p1 := cedar.NewPolicyFromAST(ast.Forbid())
		ps.Set("a wavering policy", p1)

		p2 := cedar.NewPolicyFromAST(ast.Permit())
		ps.Set("a wavering policy", p2)

		testutil.Equals(t, ps.Get("a wavering policy"), p2)
	})
}

func TestDeletePolicy(t *testing.T) {
	t.Parallel()
	t.Run("delete non-existent", func(t *testing.T) {
		t.Parallel()

		ps := cedar.NewPolicySet()

		// Just verify that this doesn't crash
		ps.Delete("not a policy")
	})
	t.Run("delete existing", func(t *testing.T) {
		t.Parallel()

		ps := cedar.NewPolicySet()

		p1 := cedar.NewPolicyFromAST(ast.Forbid())
		ps.Set("a policy", p1)
		ps.Delete("a policy")

		testutil.Equals(t, ps.Get("a policy"), cedar.Policy{})
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

	policies, err := cedar.NewPolicyListFromBytes("", []byte(policiesStr))
	testutil.OK(t, err)

	ps := cedar.NewPolicySet()
	for i, p := range policies {
		p.SetFilename("example.cedar")
		ps.Set(cedar.PolicyID(fmt.Sprintf("policy%d", i)), p)
	}

	testutil.Equals(t, ps.Get("policy0").Effect(), cedar.Permit)
	testutil.Equals(t, ps.Get("policy1").Effect(), cedar.Forbid)

	testutil.Equals(t, string(ps.MarshalCedar()), policiesStr)

}

func TestPolicyMap(t *testing.T) {
	t.Parallel()
	ps, err := cedar.NewPolicySetFromBytes("", []byte(`permit (principal, action, resource);`))
	testutil.OK(t, err)
	m := ps.Map()
	testutil.Equals(t, len(m), 1)
}
