package cedar

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntityIsZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		uid  types.EntityUID
		want bool
	}{
		{"empty", types.EntityUID{}, true},
		{"empty-type", types.NewEntityUID("one", ""), false},
		{"empty-id", types.NewEntityUID("", "one"), false},
		{"not-empty", types.NewEntityUID("one", "two"), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.uid.IsZero(), tt.want)
		})
	}
}

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
		testutil.Equals(t, ps[0].Annotations, Annotations{"key": "value"})
	})
}
