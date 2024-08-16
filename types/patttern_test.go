package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestPatternFromBuilder(t *testing.T) {
	tests := []struct {
		name    string
		Pattern Pattern
		want    []PatternComponent
	}{
		{"empty", Pattern{}, Pattern{}},
		{"wildcard", (Pattern{}).Wildcard(), Pattern{{Wildcard: true}}},
		{"saturate two wildcards", (Pattern{}).Wildcard().Wildcard(), Pattern{{Wildcard: true}}},
		{"literal", (Pattern{}).Literal("foo"), Pattern{{Literal: "foo"}}},
		{"saturate two literals", (Pattern{}).Literal("foo").Literal("bar"), Pattern{{Literal: "foobar"}}},
		{"literal with asterisk", (Pattern{}).Literal("fo*o"), Pattern{{Literal: "fo*o"}}},
		{"wildcard sandwich", (Pattern{}).Literal("foo").Wildcard().Literal("bar"), Pattern{{Literal: "foo"}, {Wildcard: true, Literal: "bar"}}},
		{"literal sandwich", (Pattern{}).Wildcard().Literal("foo").Wildcard(), Pattern{{Wildcard: true, Literal: "foo"}, {Wildcard: true}}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.Pattern, tt.want)
		})
	}
}

func TestParsePattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantOk  bool
		want    []PatternComponent
		wantErr string
	}{
		{"", true, nil, ""},
		{"a", true, []PatternComponent{{false, "a"}}, ""},
		{"*", true, []PatternComponent{{true, ""}}, ""},
		{"*a", true, []PatternComponent{{true, "a"}}, ""},
		{"a*", true, []PatternComponent{{false, "a"}, {true, ""}}, ""},
		{"**", true, []PatternComponent{{true, ""}}, ""},
		{"**a", true, []PatternComponent{{true, "a"}}, ""},
		{"a**", true, []PatternComponent{{false, "a"}, {true, ""}}, ""},
		{"*a*", true, []PatternComponent{{true, "a"}, {true, ""}}, ""},
		{"**a**", true, []PatternComponent{{true, "a"}, {true, ""}}, ""},
		{"abra*ca", true, []PatternComponent{
			{false, "abra"}, {true, "ca"},
		}, ""},
		{"abra**ca", true, []PatternComponent{
			{false, "abra"}, {true, "ca"},
		}, ""},
		{"*abra*ca", true, []PatternComponent{
			{true, "abra"}, {true, "ca"},
		}, ""},
		{"abra*ca*", true, []PatternComponent{
			{false, "abra"}, {true, "ca"}, {true, ""},
		}, ""},
		{"*abra*ca*", true, []PatternComponent{
			{true, "abra"}, {true, "ca"}, {true, ""},
		}, ""},
		{"*abra*ca*dabra", true, []PatternComponent{
			{true, "abra"}, {true, "ca"}, {true, "dabra"},
		}, ""},
		{`*abra*c\**da\*ra`, true, []PatternComponent{
			{true, "abra"}, {true, "c*"}, {true, "da*ra"},
		}, ""},
		{`\u`, false, nil, "bad unicode rune"},
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

func TestMatch(t *testing.T) {
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
			got := pat.Match(tt.target)
			testutil.Equals(t, got, tt.want)
		})
	}
}
