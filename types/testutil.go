package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

// TODO: this file should not be public, it should be moved into the eval code

func AssertValue(t *testing.T, got, want Value) {
	t.Helper()
	testutil.FatalIf(
		t,
		!((got == ZeroValue() && want == ZeroValue()) ||
			(got != ZeroValue() && want != ZeroValue() && got.Equal(want))),
		"got %v want %v", got, want)
}

func AssertBoolValue(t *testing.T, got Value, want bool) {
	t.Helper()
	testutil.Equals[Value](t, got, Boolean(want))
}

func AssertLongValue(t *testing.T, got Value, want int64) {
	t.Helper()
	testutil.Equals[Value](t, got, Long(want))
}

func AssertZeroValue(t *testing.T, got Value) {
	t.Helper()
	testutil.Equals(t, got, ZeroValue())
}

func AssertValueString(t *testing.T, v Value, want string) {
	t.Helper()
	testutil.Equals(t, v.String(), want)
}
