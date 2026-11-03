package scanner

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestIsReservedKeyword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"true", true},
		{"false", true},
		{"if", true},
		{"then", true},
		{"else", true},
		{"in", true},
		{"like", true},
		{"has", true},
		{"is", true},
		{"__cedar", true},
		{"", false},
		{"principal", false},
		{"If", false},
		{"__cedarx", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, IsReservedKeyword(tt.in), tt.want)
		})
	}
}
