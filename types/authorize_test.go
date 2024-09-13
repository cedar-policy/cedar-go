package types

import (
	"encoding/json"
	"testing"

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

func TestError(t *testing.T) {
	t.Parallel()
	e := DiagnosticError{PolicyID: "policy42", Message: "bad error"}
	testutil.Equals(t, e.String(), "while evaluating policy `policy42`: bad error")
}
