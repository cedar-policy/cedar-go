package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/testutil"
)

func TestPatternFromCedar(t *testing.T) {
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
			got, err := PatternFromCedar(tt.input)
			if err != nil {
				testutil.Equals(t, tt.wantOk, false)
				testutil.Equals(t, err.Error(), tt.wantErr)
			} else {
				testutil.Equals(t, tt.wantOk, true)
				testutil.Equals(t, got.Comps, tt.want)
			}
		})
	}
}

func TestPatternFromBuilder(t *testing.T) {
	tests := []struct {
		name    string
		Pattern *Pattern
		want    []PatternComponent
	}{
		{"empty", &Pattern{}, nil},
		{"wildcard", (&Pattern{}).AddWildcard(), []PatternComponent{{Star: true}}},
		{"saturate two wildcards", (&Pattern{}).AddWildcard().AddWildcard(), []PatternComponent{{Star: true}}},
		{"literal", (&Pattern{}).AddLiteral("foo"), []PatternComponent{{Chunk: "foo"}}},
		{"saturate two literals", (&Pattern{}).AddLiteral("foo").AddLiteral("bar"), []PatternComponent{{Chunk: "foobar"}}},
		{"literal with asterisk", (&Pattern{}).AddLiteral("fo*o"), []PatternComponent{{Chunk: "fo*o"}}},
		{"wildcard sandwich", (&Pattern{}).AddLiteral("foo").AddWildcard().AddLiteral("bar"), []PatternComponent{{Chunk: "foo"}, {Star: true, Chunk: "bar"}}},
		{"literal sandwich", (&Pattern{}).AddWildcard().AddLiteral("foo").AddWildcard(), []PatternComponent{{Star: true, Chunk: "foo"}, {Star: true}}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.Pattern.Comps, tt.want)
		})
	}
}
