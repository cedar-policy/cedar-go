package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestLong(t *testing.T) {
	t.Parallel()

	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		testutil.Equals(t, Long(42).hash(), Long(42).hash())
		testutil.Equals(t, Long(-42).hash(), Long(-42).hash())

		// This isn't necessarily true for all values of Long, but we want to ensure we aren't just returning the
		// same hash value for Long.hash() for every instance.
		testutil.FatalIf(t, Long(42).hash() == Long(1337).hash(), "unexpected hash collision")
	})
}
