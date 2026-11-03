package scanner

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestIsPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"User", true},
		{"A::B::C", true},
		{"_a::b_1", true},
		{"", false},
		{"A::", false},
		{"::A", false},
		{"A::::B", false},
		{"A :: B", false},
		{"A::if", false},
		{"__cedar::A", false},
		{`A::"x"`, false},
		{"Nobody, action, resource ); permit ( principal", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, IsPath(tt.in), tt.want)
		})
	}
}

func TestIsActionType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"Action", true},
		{"NS::Action", true},
		{"A::B::Action", true},
		{"action", false},
		{"Actions", false},
		{"Action::Foo", false},
		{"table", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, IsActionType(tt.in), tt.want)
		})
	}
}
