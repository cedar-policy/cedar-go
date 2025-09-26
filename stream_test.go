package cedar_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

func newPolicy() *cedar.Policy {
	policyJSON := []byte(`{
    "effect": "permit",
    "principal": {
        "op": "==",
        "entity": { "type": "User", "id": "bob" }
    },
    "action": {
        "op": "==",
        "entity": { "type": "Action", "id": "view" }
    },
    "resource": {
        "op": "in",
        "entity": { "type": "Folder", "id": "abc" }
    }
}`)

	var policy cedar.Policy

	_ = policy.UnmarshalJSON(policyJSON)

	return &policy
}

func TestEncoder(t *testing.T) {
	policy0 := newPolicy()
	policy1 := newPolicy()

	var buf bytes.Buffer

	encoder := cedar.NewEncoder(&buf)

	err := encoder.Encode(policy0)
	testutil.OK(t, err)

	err = encoder.Encode(policy1)
	testutil.OK(t, err)

	const expected = `permit (
    principal == User::"bob",
    action == Action::"view",
    resource in Folder::"abc"
);
permit (
    principal == User::"bob",
    action == Action::"view",
    resource in Folder::"abc"
);
`

	testutil.Equals(t, expected, buf.String())
}

func TestEncoderError(t *testing.T) {
	policy := newPolicy()

	_, w := io.Pipe()
	_ = w.Close()

	encoder := cedar.NewEncoder(w)

	err := encoder.Encode(policy)
	testutil.Error(t, err)
}

func TestDecoder(t *testing.T) {
	policyStr := `permit (
		principal,
		action,
		resource
	);
	forbid (
		principal,
		action,
		resource
	);

`

	decoder := cedar.NewDecoder(strings.NewReader(policyStr))

	var actualPolicy0 cedar.Policy
	testutil.OK(t, decoder.Decode(&actualPolicy0))

	astPolicy0 := internalast.Permit()
	astPolicy0.Position = internalast.Position{Offset: 0, Line: 1, Column: 1}
	expectedPolicy0 := cedar.NewPolicyFromAST((*ast.Policy)(astPolicy0))
	testutil.Equals(t, &actualPolicy0, expectedPolicy0)

	var actualPolicy1 cedar.Policy
	testutil.OK(t, decoder.Decode(&actualPolicy1))

	astPolicy1 := internalast.Forbid()
	astPolicy1.Position = internalast.Position{Offset: 48, Line: 6, Column: 2}
	expectedPolicy1 := cedar.NewPolicyFromAST((*ast.Policy)(astPolicy1))
	testutil.Equals(t, &actualPolicy1, expectedPolicy1)

	testutil.ErrorIs(t, decoder.Decode(nil), io.EOF)
}
