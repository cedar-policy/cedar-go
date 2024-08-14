package cedar

import (
	"testing"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestNewPolicySet(t *testing.T) {
	t.Parallel()
	t.Run("err-in-tokenize", func(t *testing.T) {
		t.Parallel()
		_, err := NewPolicySet("policy.cedar", []byte(`"`))
		testutil.Error(t, err)
	})
	t.Run("err-in-parse", func(t *testing.T) {
		t.Parallel()
		_, err := NewPolicySet("policy.cedar", []byte(`err`))
		testutil.Error(t, err)
	})
	t.Run("annotations", func(t *testing.T) {
		t.Parallel()
		ps, err := NewPolicySet("policy.cedar", []byte(`@key("value") permit (principal, action, resource);`))
		testutil.OK(t, err)
		testutil.Equals(t, ps.GetPolicy("policy0").Annotations, Annotations{"key": "value"})
	})
}

func TestNewPolicySetFromPolicies(t *testing.T) {
	t.Parallel()
	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		var policies []*Policy
		ps := NewPolicySetFromPolicies(policies)

		testutil.Equals(t, ps.GetPolicy("policy0"), nil)
	})
	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()

		policy0 := NewPolicyFromAST(ast.Forbid())

		var policy1 Policy
		testutil.OK(t, policy1.UnmarshalJSON(
			[]byte(`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`),
		))

		ps := NewPolicySetFromPolicies([]*Policy{policy0, &policy1})

		testutil.Equals(t, ps.GetPolicy("policy0"), policy0)
		testutil.Equals(t, ps.GetPolicy("policy1"), &policy1)
		testutil.Equals(t, ps.GetPolicy("policy2"), nil)
	})
}
