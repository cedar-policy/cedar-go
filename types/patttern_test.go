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
	}{{
		"like single wildcard",
		`["Wildcard"]`,
		testutil.OK,
		Pattern{}.Wildcard(),
		true,
	},
		{
			"like single literal",
			`[{"Literal":"foo"}]`,
			testutil.OK,
			Pattern{}.Literal("foo"),
			true,
		},
		{
			"like wildcard then literal",
			`["Wildcard", {"Literal":"foo"}]`,
			testutil.OK,
			Pattern{}.Wildcard().Literal("foo"),
			true,
		},
		{
			"like literal then wildcard",
			`[{"Literal":"foo"}, "Wildcard"]`,
			testutil.OK,
			Pattern{}.Literal("foo").Wildcard(),
			true,
		},
		{
			"like literal with asterisk then wildcard",
			`[{"Literal":"f*oo"}, "Wildcard"]`,
			testutil.OK,
			Pattern{}.Literal("f*oo").Wildcard(),
			true,
		},
		{
			"like literal sandwich",
			`[{"Literal":"foo"}, "Wildcard", {"Literal":"bar"}]`,
			testutil.OK,
			Pattern{}.Literal("foo").Wildcard().Literal("bar"),
			true,
		},
		{
			"like wildcard sandwich",
			`["Wildcard", {"Literal":"foo"}, "Wildcard"]`,
			testutil.OK,
			Pattern{}.Wildcard().Literal("foo").Wildcard(),
			true,
		},
		{
			"double wildcard",
			`["Wildcard", "Wildcard", {"Literal":"foo"}]`,
			testutil.OK,
			Pattern{}.Wildcard().Literal("foo"),
			false,
		},
		{
			"double literal",
			`["Wildcard", {"Literal":"foo"}, {"Literal":"bar"}]`,
			testutil.OK,
			Pattern{}.Wildcard().Literal("foobar"),
			false,
		},
		{
			"literal with asterisk",
			`["Wildcard", {"Literal":"foo*"}, "Wildcard"]`,
			testutil.OK,
			Pattern{}.Wildcard().Literal("foo*").Wildcard(),
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
