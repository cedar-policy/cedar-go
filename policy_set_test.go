package cedar

import (
	"testing"

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
