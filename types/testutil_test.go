package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func assertValueString(t *testing.T, v types.Value, want string) {
	t.Helper()
	testutil.Equals(t, v.String(), want)
}
