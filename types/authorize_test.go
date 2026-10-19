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
	t.Run("UnmarshalInvalid", func(t *testing.T) {
		t.Parallel()
		for _, in := range []string{`"Allow"`, `"DENY"`, `"permit"`, `""`, `true`, `null`} {
			d := Allow
			testutil.Error(t, json.Unmarshal([]byte(in), &d))
			testutil.Equals(t, d, Allow)
		}
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

// checkCaseSensitiveJSON checks that, even through encoding/json (v1), a
// miscased member of a T is ignored rather than matched and a repeated member
// is an error.
func checkCaseSensitiveJSON[T any](t *testing.T, miscased, repeated string) {
	t.Helper()
	var got, zero T
	testutil.OK(t, json.Unmarshal([]byte(miscased), &got))
	testutil.Equals(t, got, zero)
	testutil.Error(t, json.Unmarshal([]byte(repeated), &got))
}

func TestAuthorizeJSONCaseSensitive(t *testing.T) {
	t.Parallel()
	t.Run("Request", func(t *testing.T) {
		t.Parallel()
		checkCaseSensitiveJSON[Request](t,
			`{"Principal":{"type":"User","id":"alice"}}`,
			`{"principal":{"type":"User","id":"alice"},"principal":{"type":"User","id":"bob"}}`)
	})
	t.Run("Diagnostic", func(t *testing.T) {
		t.Parallel()
		checkCaseSensitiveJSON[Diagnostic](t,
			`{"Reasons":[{"policy":"p"}]}`,
			`{"reasons":[],"reasons":[]}`)
	})
	t.Run("DiagnosticReason", func(t *testing.T) {
		t.Parallel()
		checkCaseSensitiveJSON[DiagnosticReason](t,
			`{"Policy":"p"}`,
			`{"policy":"p","policy":"q"}`)
	})
	t.Run("DiagnosticError", func(t *testing.T) {
		t.Parallel()
		checkCaseSensitiveJSON[DiagnosticError](t,
			`{"Message":"m"}`,
			`{"message":"m","message":"n"}`)
	})
	t.Run("Position", func(t *testing.T) {
		t.Parallel()
		checkCaseSensitiveJSON[Position](t,
			`{"Line":2}`,
			`{"line":2,"line":3}`)
	})
}
