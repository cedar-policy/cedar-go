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
		ps.Upsert("policy0", policy0)
		ps.Upsert("policy1", policy1)

		testutil.Equals(t, ps.Get("policy0"), policy0)
		testutil.Equals(t, ps.Get("policy1"), policy1)
		testutil.Equals(t, ps.Get("policy2"), cedar.Policy{})
	})
	t.Run("upsert", func(t *testing.T) {
		t.Parallel()

		ps := cedar.NewPolicySet()

		p1 := cedar.NewPolicyFromAST(ast.Forbid())
		ps.Upsert("a wavering policy", p1)

		p2 := cedar.NewPolicyFromAST(ast.Permit())
		ps.Upsert("a wavering policy", p2)

		testutil.Equals(t, ps.Get("a wavering policy"), p2)
	})
}

// func TestUpsertPolicySet(t *testing.T) {
// 	t.Parallel()
// 	t.Run("empty dst", func(t *testing.T) {
// 		t.Parallel()

// 		policy0 := cedar.NewPolicyFromAST(ast.Forbid())

// 		var policy1 cedar.Policy
// 		testutil.OK(t, policy1.UnmarshalJSON(
// 			[]byte(`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`),
// 		))

// 		ps1 := cedar.NewPolicySet()
// 		ps1.Upsert("policy0", policy0)
// 		ps1.Upsert("policy1", policy1)

// 		ps2 := cedar.NewPolicySet()
// 		ps2.UpsertPolicySet(ps1)

// 		testutil.Equals(t, ps2.Get("policy0"), policy0)
// 		testutil.Equals(t, ps2.Get("policy1"), &policy1)
// 		testutil.Equals(t, ps2.Get("policy2"), nil)
// 	})
// 	t.Run("upsert", func(t *testing.T) {
// 		t.Parallel()

// 		policyA := cedar.NewPolicyFromAST(ast.Forbid())

// 		var policyB cedar.Policy
// 		testutil.OK(t, policyB.UnmarshalJSON(
// 			[]byte(`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`),
// 		))

// 		policyC := cedar.NewPolicyFromAST(ast.Permit())

// 		// ps1 maps 0 -> A and 1 -> B
// 		ps1 := cedar.NewPolicySet()
// 		ps1.Upsert("policy0", policyA)
// 		ps1.Upsert("policy1", &policyB)

// 		// ps1 maps 0 -> b and 2 -> C
// 		ps2 := cedar.NewPolicySet()
// 		ps2.Upsert("policy0", &policyB)
// 		ps2.Upsert("policy2", policyC)

// 		// Upsert should clobber ps2's policy0, insert policy1, and leave policy2 untouched
// 		ps2.UpsertPolicySet(ps1)

// 		testutil.Equals(t, ps2.Get("policy0"), policyA)
// 		testutil.Equals(t, ps2.Get("policy1"), &policyB)
// 		testutil.Equals(t, ps2.Get("policy2"), policyC)
// 	})
// }

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
		ps.Upsert("a policy", p1)
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

	policies, err := cedar.NewPoliciesFromBytes("", []byte(policiesStr))
	testutil.OK(t, err)

	ps := cedar.NewPolicySet()
	for i, p := range policies {
		p.SetFilename("example.cedar")
		ps.Upsert(cedar.PolicyID(fmt.Sprintf("policy%d", i)), p)
	}

	testutil.Equals(t, ps.Get("policy0").Effect(), cedar.Permit)
	testutil.Equals(t, ps.Get("policy1").Effect(), cedar.Forbid)

	testutil.Equals(t, string(ps.MarshalCedar()), policiesStr)
}
