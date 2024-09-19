package types_test

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
