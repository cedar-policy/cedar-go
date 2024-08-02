package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/testutil"
)

func TestPatternFromStringLiteral(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantOk  bool
		want    []PatternComponent
		wantErr string
	}{
		{`""`, true, nil, ""},
		{`"a"`, true, []PatternComponent{{false, "a"}}, ""},
		{`"*"`, true, []PatternComponent{{true, ""}}, ""},
		{`"*a"`, true, []PatternComponent{{true, "a"}}, ""},
		{`"a*"`, true, []PatternComponent{{false, "a"}, {true, ""}}, ""},
		{`"**"`, true, []PatternComponent{{true, ""}}, ""},
		{`"**a"`, true, []PatternComponent{{true, "a"}}, ""},
		{`"a**"`, true, []PatternComponent{{false, "a"}, {true, ""}}, ""},
		{`"*a*"`, true, []PatternComponent{{true, "a"}, {true, ""}}, ""},
		{`"**a**"`, true, []PatternComponent{{true, "a"}, {true, ""}}, ""},
		{`"abra*ca"`, true, []PatternComponent{
			{false, "abra"}, {true, "ca"},
		}, ""},
		{`"abra**ca"`, true, []PatternComponent{
			{false, "abra"}, {true, "ca"},
		}, ""},
		{`"*abra*ca"`, true, []PatternComponent{
			{true, "abra"}, {true, "ca"},
		}, ""},
		{`"abra*ca*"`, true, []PatternComponent{
			{false, "abra"}, {true, "ca"}, {true, ""},
		}, ""},
		{`"*abra*ca*"`, true, []PatternComponent{
			{true, "abra"}, {true, "ca"}, {true, ""},
		}, ""},
		{`"*abra*ca*dabra"`, true, []PatternComponent{
			{true, "abra"}, {true, "ca"}, {true, "dabra"},
		}, ""},
		{`"*abra*c\**da\*ra"`, true, []PatternComponent{
			{true, "abra"}, {true, "c*"}, {true, "da*ra"},
		}, ""},
		{`"\u"`, false, nil, "bad unicode rune"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := NewPattern(tt.input)
			if err != nil {
				testutil.Equals(t, tt.wantOk, false)
				testutil.Equals(t, err.Error(), tt.wantErr)
			} else {
				testutil.Equals(t, tt.wantOk, true)
				testutil.Equals(t, got.Comps, tt.want)
				testutil.Equals(t, got.String(), tt.input)
			}
		})
	}
}
