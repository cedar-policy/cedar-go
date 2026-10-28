package scanner

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestIsAnyIdent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"a", true},
		{"_", true},
		{"A_1", true},
		{"if", true},
		{"__cedar", true},
		{"", false},
		{"1a", false},
		{"a b", false},
		{"a-b", false},
		{"a::b", false},
		{"é", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, IsAnyIdent(tt.in), tt.want)
		})
	}
}

func TestIsIdent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"a", true},
		{"ifx", true},
		{"if", false},
		{"true", false},
		{"__cedar", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, IsIdent(tt.in), tt.want)
		})
	}
}
