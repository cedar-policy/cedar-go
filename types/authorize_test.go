package types

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestJSONDecision(t *testing.T) {
	t.Parallel()
	t.Run("MarshalAllow", func(t *testing.T) {
		t.Parallel()
		d := Allow
		b, err := d.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), `"allow"`)
	})
	t.Run("MarshalDeny", func(t *testing.T) {
		t.Parallel()
		d := Deny
		b, err := d.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), `"deny"`)
	})
	t.Run("UnmarshalAllow", func(t *testing.T) {
		t.Parallel()
		var d Decision
		err := json.Unmarshal([]byte(`"allow"`), &d)
		testutil.OK(t, err)
		testutil.Equals(t, d, Allow)
	})
	t.Run("UnmarshalDeny", func(t *testing.T) {
		t.Parallel()
		var d Decision
		err := json.Unmarshal([]byte(`"deny"`), &d)
		testutil.OK(t, err)
		testutil.Equals(t, d, Deny)
	})
}

func TestRequestEqual(t *testing.T) {
	t.Parallel()

	testcases := []Request{
		{},
		{
			Principal: EntityUID{
				Type: "a",
				ID:   "b",
			},
		},
		{
			Action: EntityUID{
				Type: "a",
				ID:   "b",
			},
		},
		{
			Resource: EntityUID{
				Type: "a",
				ID:   "b",
			},
		},
		{
			Context: NewRecord(RecordMap{
				"a": Long(42),
			}),
		},
	}

	for i, r := range testcases {
		for j, s := range testcases[:i] {
			expected := i == j
			actual := r.Equal(s)
			if actual != s.Equal(r) {
				t.Errorf("Request.Equal not symmetric for examples #%d and #%d, actual=%v, expected=%v", i, j, actual, expected)
			}
			if actual != expected {
				t.Errorf("Request.Equal returned unexpected result for #%d and #%d, actual=%v, expected=%v", i, j, actual, expected)
			}
		}
	}
}

func TestRequestEqualSupportsCmpDiff(t *testing.T) {
	t.Parallel()

	// This test is not very interesting, we just want to make
	// sure that it compiles.
	if diff := cmp.Diff(Record{}, Record{}); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

func TestError(t *testing.T) {
	t.Parallel()
	e := DiagnosticError{PolicyID: "policy42", Message: "bad error"}
	testutil.Equals(t, e.String(), "while evaluating policy `policy42`: bad error")
}
