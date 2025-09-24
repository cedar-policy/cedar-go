package parser_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestEncoder(t *testing.T) {
	var buf bytes.Buffer

	encoder := parser.NewEncoder(&buf)

	policy := ast.Permit().
		PrincipalEq(johnny).
		ActionEq(sow).
		ResourceEq(apple).
		When(ast.Boolean(true)).
		Unless(ast.Boolean(false))

	err := encoder.Encode((*parser.Policy)(policy))
	testutil.OK(t, err)

	const expected = `permit (
    principal == User::"johnny",
    action == Action::"sow",
    resource == Crop::"apple"
)
when { true }
unless { false };
`

	testutil.Equals(t, buf.String(), expected)
}

func TestEncoderError(t *testing.T) {
	_, w := io.Pipe()
	_ = w.Close()

	encoder := parser.NewEncoder(w)

	policy := ast.Permit().
		PrincipalEq(johnny).
		ActionEq(sow).
		ResourceEq(apple).
		When(ast.Boolean(true)).
		Unless(ast.Boolean(false))

	err := encoder.Encode((*parser.Policy)(policy))
	testutil.Error(t, err)
}
