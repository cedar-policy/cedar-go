package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestPatternFromBuilder(t *testing.T) {
	t.Run("saturate two wildcards", func(t *testing.T) {
		pattern1 := PatternBuilder{}.AddWildcard().AddWildcard().Build()
		testutil.Equals(t, pattern1, WildcardPattern)
	})
	t.Run("saturate two literals", func(t *testing.T) {
		pattern1 := PatternBuilder{}.AddLiteral("foo").AddLiteral("bar").Build()
		pattern2 := LiteralPattern("foobar")
		testutil.Equals(t, pattern1, pattern2)
	})
}

func TestParsePattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantOk  bool
		want    Pattern
		wantErr string
	}{
		{"", true, LiteralPattern(""), ""},
		{"a", true, LiteralPattern("a"), ""},
		{"*", true, WildcardPattern, ""},
		{"*a", true, PatternBuilder{}.AddWildcard().AddLiteral("a").Build(), ""},
		{"a*", true, PatternBuilder{}.AddLiteral("a").AddWildcard().Build(), ""},
		{"**", true, WildcardPattern, ""},
		{"**a", true, PatternBuilder{}.AddWildcard().AddLiteral("a").Build(), ""},
		{"a**", true, PatternBuilder{}.AddLiteral("a").AddWildcard().Build(), ""},
		{"*a*", true, PatternBuilder{}.AddWildcard().AddLiteral("a").AddWildcard().Build(), ""},
		{"**a**", true, PatternBuilder{}.AddWildcard().AddLiteral("a").AddWildcard().Build(), ""},
		{"abra*ca", true, PatternBuilder{}.AddLiteral("abra").AddWildcard().AddLiteral("ca").Build(), ""},
		{"abra**ca", true, PatternBuilder{}.AddLiteral("abra").AddWildcard().AddLiteral("ca").Build(), ""},
		{"*abra*ca", true, PatternBuilder{}.AddWildcard().AddLiteral("abra").AddWildcard().AddLiteral("ca").Build(), ""},
		{"abra*ca*", true, PatternBuilder{}.AddLiteral("abra").AddWildcard().AddLiteral("ca").AddWildcard().Build(), ""},
		{"*abra*ca*", true, PatternBuilder{}.AddWildcard().AddLiteral("abra").AddWildcard().AddLiteral("ca").AddWildcard().Build(), ""},
		{"*abra*ca*dabra", true, PatternBuilder{}.AddWildcard().AddLiteral("abra").AddWildcard().AddLiteral("ca").AddWildcard().AddLiteral("dabra").Build(), ""},
		{`*abra*c\**da\*bra`, true, PatternBuilder{}.AddWildcard().AddLiteral("abra").AddWildcard().AddLiteral("c*").AddWildcard().AddLiteral("da*bra").Build(), ""},
		{`\u`, false, Pattern{}, "bad unicode rune"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			var got Pattern
			if err := got.UnmarshalCedar([]byte(tt.input)); err != nil {
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
			var pat Pattern
			err := pat.UnmarshalCedar([]byte(tt.pattern[1 : len(tt.pattern)-1]))
			testutil.OK(t, err)
			got := pat.Match(tt.target)
			testutil.Equals(t, got, tt.want)
		})
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		pattern         string
		errFunc         func(testing.TB, error)
		target          Pattern
		shouldRoundTrip bool
	}{
		{
			"like single wildcard",
			`["Wildcard"]`,
			testutil.OK,
			WildcardPattern,
			true,
		},
		{
			"like single literal",
			`[{"Literal":"foo"}]`,
			testutil.OK,
			LiteralPattern("foo"),
			true,
		},
		{
			"like wildcard then literal",
			`["Wildcard", {"Literal":"foo"}]`,
			testutil.OK,
			PatternBuilder{}.AddWildcard().AddLiteral("foo").Build(),
			true,
		},
		{
			"like literal then wildcard",
			`[{"Literal":"foo"}, "Wildcard"]`,
			testutil.OK,
			PatternBuilder{}.AddLiteral("foo").AddWildcard().Build(),
			true,
		},
		{
			"like literal with asterisk then wildcard",
			`[{"Literal":"f*oo"}, "Wildcard"]`,
			testutil.OK,
			PatternBuilder{}.AddLiteral("f*oo").AddWildcard().Build(),
			true,
		},
		{
			"like literal sandwich",
			`[{"Literal":"foo"}, "Wildcard", {"Literal":"bar"}]`,
			testutil.OK,
			PatternBuilder{}.AddLiteral("foo").AddWildcard().AddLiteral("bar").Build(),
			true,
		},
		{
			"like wildcard sandwich",
			`["Wildcard", {"Literal":"foo"}, "Wildcard"]`,
			testutil.OK,
			PatternBuilder{}.AddWildcard().AddLiteral("foo").AddWildcard().Build(),
			true,
		},
		{
			"double wildcard",
			`["Wildcard", "Wildcard", {"Literal":"foo"}]`,
			testutil.OK,
			PatternBuilder{}.AddWildcard().AddLiteral("foo").Build(),
			false,
		},
		{
			"double literal",
			`["Wildcard", {"Literal":"foo"}, {"Literal":"bar"}]`,
			testutil.OK,
			PatternBuilder{}.AddWildcard().AddLiteral("foobar").Build(),
			false,
		},
		{
			"literal with asterisk",
			`["Wildcard", {"Literal":"foo*"}, "Wildcard"]`,
			testutil.OK,
			PatternBuilder{}.AddWildcard().AddLiteral("foo*").AddWildcard().Build(),
			true,
		},
		{
			"not list",
			`"Wildcard"`,
			testutil.Error,
			Pattern{},
			false,
		},
		{
			"lower case wildcard",
			`["wildcard"]`,
			testutil.Error,
			Pattern{},
			false,
		},
		{
			"other string",
			`["cardwild"]`,
			testutil.Error,
			Pattern{},
			false,
		},
		{
			"lowercase literal",
			`[{"literal": "foo"}]`,
			testutil.Error,
			Pattern{},
			false,
		},
		{
			"missing literal",
			`[{"figurative": "haha"}]`,
			testutil.Error,
			Pattern{},
			false,
		},
		{
			"two keys",
			`[{"Literal": "foo", "Figurative": "haha"}]`,
			testutil.Error,
			Pattern{},
			false,
		},
		{
			"nonstring literal",
			`[{"Literal": 2}]`,
			testutil.Error,
			Pattern{},
			false,
		},
		{
			"empty pattern",
			`[]`,
			testutil.Error,
			Pattern{},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var pat Pattern
			err := pat.UnmarshalJSON([]byte(tt.pattern))
			tt.errFunc(t, err)
			if err != nil {
				return
			}

			marshaled, err := pat.MarshalJSON()
			testutil.OK(t, err)

			if tt.shouldRoundTrip {
				testutil.Equals(t, string(marshaled), tt.pattern)
			}
		})
	}
}
