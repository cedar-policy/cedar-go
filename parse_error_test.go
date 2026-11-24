package cedar_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestParseError(t *testing.T) {
	t.Parallel()

	const syntaxErr = "permit (principal, action, resource)\nwhen { context has a.b && };"
	const scanErr = "permit (principal, action, resource)\nwhen { \"unterminated };"

	tests := []struct {
		name     string
		parse    func(string) error
		src      string
		position cedar.Position
		message  string
		text     string
	}{
		{
			"policy syntax error",
			func(s string) error { var p cedar.Policy; return p.UnmarshalCedar([]byte(s)) },
			syntaxErr,
			cedar.Position{Offset: 64, Line: 2, Column: 28},
			"invalid primary",
			`parse error at <input>:2:28 ";": invalid primary`,
		},
		{
			"policy scan error",
			func(s string) error { var p cedar.Policy; return p.UnmarshalCedar([]byte(s)) },
			scanErr,
			cedar.Position{Offset: 44, Line: 2, Column: 8},
			"literal not terminated",
			"<input>:2:8: literal not terminated",
		},
		{
			"policy list",
			func(s string) error { var l cedar.PolicyList; return l.UnmarshalCedar([]byte(s)) },
			syntaxErr,
			cedar.Position{Offset: 64, Line: 2, Column: 28},
			"invalid primary",
			`parser error: parse error at <input>:2:28 ";": invalid primary`,
		},
		{
			"policy list with file name",
			func(s string) error { _, err := cedar.NewPolicyListFromBytes("a.cedar", []byte(s)); return err },
			syntaxErr,
			cedar.Position{Filename: "a.cedar", Offset: 64, Line: 2, Column: 28},
			"invalid primary",
			`parser error: parse error at <input>:2:28 ";": invalid primary`,
		},
		{
			"policy set with file name",
			func(s string) error { _, err := cedar.NewPolicySetFromBytes("b.cedar", []byte(s)); return err },
			scanErr,
			cedar.Position{Filename: "b.cedar", Offset: 44, Line: 2, Column: 8},
			"literal not terminated",
			"parser error: <input>:2:8: literal not terminated",
		},
		{
			"decoder",
			func(s string) error { var p cedar.Policy; return cedar.NewDecoder(strings.NewReader(s)).Decode(&p) },
			syntaxErr,
			cedar.Position{Offset: 64, Line: 2, Column: 28},
			"invalid primary",
			`parse error at <input>:2:28 ";": invalid primary`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.parse(tt.src)
			var pe *cedar.ParseError
			testutil.Equals(t, errors.As(err, &pe), true)
			testutil.Equals(t, pe.Position, tt.position)
			testutil.Equals(t, pe.Message, tt.message)
			testutil.Equals(t, err.Error(), tt.text)
			testutil.Equals(t, errors.Unwrap(pe) != nil, true)
		})
	}
}

func TestParseErrorNotWrappedForOtherErrors(t *testing.T) {
	t.Parallel()

	var p cedar.Policy
	dec := cedar.NewDecoder(strings.NewReader(""))
	err := dec.Decode(&p)
	testutil.Equals(t, err.Error(), "EOF")
	var pe *cedar.ParseError
	testutil.Equals(t, errors.As(err, &pe), false)
}
