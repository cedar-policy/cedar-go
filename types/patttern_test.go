package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestPatternFromBuilder(t *testing.T) {
	t.Run("saturate two wildcards", func(t *testing.T) {
		pattern1 := NewPattern(Wildcard(), Wildcard())
		pattern2 := NewPattern(Wildcard())
		testutil.Equals(t, pattern1, pattern2)
	})
	t.Run("saturate two literals", func(t *testing.T) {
		pattern1 := NewPattern(String("foo"), String("bar"))
		pattern2 := NewPattern(String("foobar"))
		testutil.Equals(t, pattern1, pattern2)
	})
}

func TestParsePattern(t *testing.T) {
	t.Parallel()
	a := String("a")
	tests := []struct {
		input   string
		wantOk  bool
		want    Pattern
		wantErr string
	}{
		{"", true, NewPattern(), ""},
		{"a", true, NewPattern(a), ""},
		{"*", true, NewPattern(Wildcard()), ""},
		{"*a", true, NewPattern(Wildcard(), a), ""},
		{"a*", true, NewPattern(a, Wildcard()), ""},
		{"**", true, NewPattern(Wildcard()), ""},
		{"**a", true, NewPattern(Wildcard(), a), ""},
		{"a**", true, NewPattern(a, Wildcard()), ""},
		{"*a*", true, NewPattern(Wildcard(), a, Wildcard()), ""},
		{"**a**", true, NewPattern(Wildcard(), a, Wildcard()), ""},
		{"abra*ca", true, NewPattern(String("abra"), Wildcard(), String("ca")), ""},
		{"abra**ca", true, NewPattern(String("abra"), Wildcard(), String("ca")), ""},
		{"*abra*ca", true, NewPattern(Wildcard(), String("abra"), Wildcard(), String("ca")), ""},
		{"abra*ca*", true, NewPattern(String("abra"), Wildcard(), String("ca"), Wildcard()), ""},
		{"*abra*ca*", true, NewPattern(Wildcard(), String("abra"), Wildcard(), String("ca"), Wildcard()), ""},
		{"*abra*ca*dabra", true, NewPattern(Wildcard(), String("abra"), Wildcard(), String("ca"), Wildcard(), String("dabra")), ""},
		{`*abra*c\**da\*bra`, true, NewPattern(Wildcard(), String("abra"), Wildcard(), String("c*"), Wildcard(), String("da*bra")), ""},
		{`\u`, false, Pattern{}, "bad unicode rune"},
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
			got := pat.Match(String(tt.target))
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
			NewPattern(Wildcard()),
			true,
		},
		{
			"like single literal",
			`[{"Literal":"foo"}]`,
			testutil.OK,
			NewPattern(String("foo")),
			true,
		},
		{
			"like wildcard then literal",
			`["Wildcard", {"Literal":"foo"}]`,
			testutil.OK,
			NewPattern(Wildcard(), String("foo")),
			true,
		},
		{
			"like literal then wildcard",
			`[{"Literal":"foo"}, "Wildcard"]`,
			testutil.OK,
			NewPattern(String("foo"), Wildcard()),
			true,
		},
		{
			"like literal with asterisk then wildcard",
			`[{"Literal":"f*oo"}, "Wildcard"]`,
			testutil.OK,
			NewPattern(String("f*oo"), Wildcard()),
			true,
		},
		{
			"like literal sandwich",
			`[{"Literal":"foo"}, "Wildcard", {"Literal":"bar"}]`,
			testutil.OK,
			NewPattern(String("foo"), Wildcard(), String("bar")),
			true,
		},
		{
			"like wildcard sandwich",
			`["Wildcard", {"Literal":"foo"}, "Wildcard"]`,
			testutil.OK,
			NewPattern(Wildcard(), String("foo"), Wildcard()),
			true,
		},
		{
			"double wildcard",
			`["Wildcard", "Wildcard", {"Literal":"foo"}]`,
			testutil.OK,
			NewPattern(Wildcard(), String("foo")),
			false,
		},
		{
			"double literal",
			`["Wildcard", {"Literal":"foo"}, {"Literal":"bar"}]`,
			testutil.OK,
			NewPattern(Wildcard(), String("foobar")),
			false,
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
