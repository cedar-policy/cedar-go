package types

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cedar-policy/cedar-go/internal/testutil"
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

	emptyRequest := Request{}
	principalRequest := Request{
		Principal: EntityUID{
			Type: "a",
			ID:   "b",
		},
	}

	contextRequest := Request{
		Context: NewRecord(RecordMap{
			"a": Long(42),
		}),
	}

	testcases := []struct {
		lhs      Request
		rhs      Request
		expected bool
	}{
		{
			emptyRequest,
			emptyRequest,
			true,
		},
		{
			principalRequest,
			contextRequest,
			false,
		},
		{
			contextRequest,
			contextRequest,
			true,
		},
	}

	for _, tt := range testcases {
		testutil.Equals(t, tt.lhs.Equal(tt.rhs), tt.expected)
		testutil.Equals(t, tt.rhs.Equal(tt.lhs), tt.expected)
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
