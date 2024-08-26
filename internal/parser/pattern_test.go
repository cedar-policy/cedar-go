package parser

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestParsePattern(t *testing.T) {
	t.Parallel()
	a := types.String("a")
	tests := []struct {
		input   string
		wantOk  bool
		want    types.Pattern
		wantErr string
	}{
		{"", true, types.NewPattern(), ""},
		{"a", true, types.NewPattern(a), ""},
		{"*", true, types.NewPattern(types.Wildcard{}), ""},
		{"*a", true, types.NewPattern(types.Wildcard{}, a), ""},
		{"a*", true, types.NewPattern(a, types.Wildcard{}), ""},
		{"**", true, types.NewPattern(types.Wildcard{}), ""},
		{"**a", true, types.NewPattern(types.Wildcard{}, a), ""},
		{"a**", true, types.NewPattern(a, types.Wildcard{}), ""},
		{"*a*", true, types.NewPattern(types.Wildcard{}, a, types.Wildcard{}), ""},
		{"**a**", true, types.NewPattern(types.Wildcard{}, a, types.Wildcard{}), ""},
		{"abra*ca", true, types.NewPattern(types.String("abra"), types.Wildcard{}, types.String("ca")), ""},
		{"abra**ca", true, types.NewPattern(types.String("abra"), types.Wildcard{}, types.String("ca")), ""},
		{"*abra*ca", true, types.NewPattern(types.Wildcard{}, types.String("abra"), types.Wildcard{}, types.String("ca")), ""},
		{"abra*ca*", true, types.NewPattern(types.String("abra"), types.Wildcard{}, types.String("ca"), types.Wildcard{}), ""},
		{"*abra*ca*", true, types.NewPattern(types.Wildcard{}, types.String("abra"), types.Wildcard{}, types.String("ca"), types.Wildcard{}), ""},
		{"*abra*ca*dabra", true, types.NewPattern(types.Wildcard{}, types.String("abra"), types.Wildcard{}, types.String("ca"), types.Wildcard{}, types.String("dabra")), ""},
		{`*abra*c\**da\*bra`, true, types.NewPattern(types.Wildcard{}, types.String("abra"), types.Wildcard{}, types.String("c*"), types.Wildcard{}, types.String("da*bra")), ""},
		{`\u`, false, types.Pattern{}, "bad unicode rune"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParsePattern(tt.input)
			if err != nil {
				testutil.Equals(t, tt.wantOk, false)
				testutil.Equals(t, err.Error(), tt.wantErr)
			} else {
				testutil.Equals(t, tt.wantOk, true)
				testutil.Equals(t, got, tt.want)
			}
		})
	}
}

func TestPatternParseAndMatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern string
		target  string
		want    bool
	}{
		{`""`, "", true},
		{`""`, "hello", false},
		{`"*"`, "hello", true},
		{`"e"`, "hello", false},
		{`"*e"`, "hello", false},
		{`"*e*"`, "hello", true},
		{`"hello"`, "hello", true},
		{`"hello*"`, "hello", true},
		{`"*h*llo*"`, "hello", true},
		{`"h*e*o"`, "hello", true},
		{`"h*e**o"`, "hello", true},
		{`"h*z*o"`, "hello", false},

		{`"\u{210d}*"`, "ℍello", true},
		{`"\u{210d}*"`, "Hello", false},

		{`"\*\**\*\*"`, "**foo**", true},
		{`"\*\**\*\*"`, "**bar**", true},
		{`"\*\**\*\*"`, "*bar*", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.pattern+":"+tt.target, func(t *testing.T) {
			t.Parallel()
			pat, err := ParsePattern(tt.pattern[1 : len(tt.pattern)-1])
			testutil.OK(t, err)
			got := pat.Match(types.String(tt.target))
			testutil.Equals(t, got, tt.want)
		})
	}
}
