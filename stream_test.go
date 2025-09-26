package cedar_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
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
	w.Close()

	encoder := cedar.NewEncoder(w)

	err := encoder.Encode(policy)
	testutil.Error(t, err)
}
