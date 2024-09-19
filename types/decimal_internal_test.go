package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestDecimal(t *testing.T) {
	t.Parallel()
	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		testutil.Equals(t, UnsafeDecimal(42).hash(), UnsafeDecimal(42).hash())
		testutil.Equals(t, UnsafeDecimal(-42).hash(), UnsafeDecimal(-42).hash())

		// This isn't necessarily true for all values of Decimal, but we want to ensure we aren't just returning the
		// same hash value for Decimal.hash() for every instance.
		testutil.FatalIf(t, UnsafeDecimal(42).hash() == UnsafeDecimal(1337).hash(), "unexpected hash collision")
	})
}
