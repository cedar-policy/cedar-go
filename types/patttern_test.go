package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestPattern(t *testing.T) {
	t.Parallel()
	t.Run("saturate two wildcards", func(t *testing.T) {
		t.Parallel()
		pattern1 := NewPattern(Wildcard{}, Wildcard{})
		pattern2 := NewPattern(Wildcard{})
		testutil.Equals(t, pattern1, pattern2)
	})
	t.Run("saturate two literals", func(t *testing.T) {
		t.Parallel()
		pattern1 := NewPattern(String("foo"), String("bar"))
		pattern2 := NewPattern(String("foobar"))
		testutil.Equals(t, pattern1, pattern2)
	})
	t.Run("panicOnNil", func(t *testing.T) {
		t.Parallel()
		testutil.Panic(t, func() {
			NewPattern(nil)
		})
	})
	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, string(NewPattern(String("*foo"), Wildcard{}).MarshalCedar()), `"\*foo*"`)
	})
}

func TestPatternMatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern Pattern
		target  string
		want    bool
	}{
		{NewPattern(), "", true},
		{NewPattern(), "hello", false},
		{NewPattern(Wildcard{}), "hello", true},
		{NewPattern(String("e")), "hello", false},
		{NewPattern(Wildcard{}, String("e")), "hello", false},
		{NewPattern(Wildcard{}, String("e"), Wildcard{}), "hello", true},
		{NewPattern(String("hello")), "hello", true},
		{NewPattern(String("hello"), Wildcard{}), "hello", true},
		{NewPattern(Wildcard{}, String("h"), Wildcard{}, String("llo"), Wildcard{}), "hello", true},
		{NewPattern(String("h"), Wildcard{}, String("e"), Wildcard{}, String("o")), "hello", true},
		{NewPattern(String("h"), Wildcard{}, String("e"), Wildcard{}, Wildcard{}, String("o")), "hello", true},
		{NewPattern(String("h"), Wildcard{}, String("z"), Wildcard{}, String("o")), "hello", false},

		{NewPattern(String("**"), Wildcard{}, String("**")), "**foo**", true},
		{NewPattern(String("**"), Wildcard{}, String("**")), "**bar**", true},
		{NewPattern(String("**"), Wildcard{}, String("**")), "*bar*", false},

		// with native strings
		{NewPattern(Wildcard{}, "ell", Wildcard{}), "hello", true},
		{NewPattern("he", Wildcard{}), "hello", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.pattern.MarshalCedar())+":"+tt.target, func(t *testing.T) {
			t.Parallel()
			got := tt.pattern.Match(String(tt.target))
			testutil.Equals(t, got, tt.want)
		})
	}
}

func TestPatternJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		pattern         string
		errFunc         func(testutil.TB, error)
		target          Pattern
		shouldRoundTrip bool
	}{
		{
			"like single wildcard",
			`["Wildcard"]`,
			testutil.OK,
			NewPattern(Wildcard{}),
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
			NewPattern(Wildcard{}, String("foo")),
			true,
		},
		{
			"like literal then wildcard",
			`[{"Literal":"foo"}, "Wildcard"]`,
			testutil.OK,
			NewPattern(String("foo"), Wildcard{}),
			true,
		},
		{
			"like literal with asterisk then wildcard",
			`[{"Literal":"f*oo"}, "Wildcard"]`,
			testutil.OK,
			NewPattern(String("f*oo"), Wildcard{}),
			true,
		},
		{
			"like literal sandwich",
			`[{"Literal":"foo"}, "Wildcard", {"Literal":"bar"}]`,
			testutil.OK,
			NewPattern(String("foo"), Wildcard{}, String("bar")),
			true,
		},
		{
			"like wildcard sandwich",
			`["Wildcard", {"Literal":"foo"}, "Wildcard"]`,
			testutil.OK,
			NewPattern(Wildcard{}, String("foo"), Wildcard{}),
			true,
		},
		{
			"double wildcard",
			`["Wildcard", "Wildcard", {"Literal":"foo"}]`,
			testutil.OK,
			NewPattern(Wildcard{}, String("foo")),
			false,
		},
		{
			"double literal",
			`["Wildcard", {"Literal":"foo"}, {"Literal":"bar"}]`,
			testutil.OK,
			NewPattern(Wildcard{}, String("foobar")),
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
			"other type",
			`[null]`,
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
