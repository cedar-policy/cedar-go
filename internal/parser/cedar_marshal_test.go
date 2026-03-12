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

func TestMarshalExpr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		node ast.Node
		want string
	}{
		{"boolean", ast.Boolean(true), "true"},
		{"string", ast.String("hello"), `"hello"`},
		{"long", ast.Long(42), "42"},
		{"variable", ast.Principal(), "principal"},
		{"add", ast.Long(1).Add(ast.Long(2)), "1 + 2"},
		{"equals", ast.Principal().Equal(ast.Principal()), "principal == principal"},
		{"access", ast.Context().Access("foo"), "context.foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.MarshalExpr(tt.node.AsIsNode())
			testutil.Equals(t, got, tt.want)
		})
	}
}
