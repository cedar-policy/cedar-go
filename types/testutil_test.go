package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

// TODO: this file should not be public, it should be moved into the eval code

func AssertValueString(t *testing.T, v types.Value, want string) {
	t.Helper()
	testutil.Equals(t, v.String(), want)
}
