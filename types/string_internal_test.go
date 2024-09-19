package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestString(t *testing.T) {
	t.Parallel()

	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		testutil.Equals(t, String("foo").hash(), String("foo").hash())
		testutil.Equals(t, String("bar").hash(), String("bar").hash())

		// This isn't necessarily true for all values of String, but we want to ensure we aren't just returning the
		// same hash value for String.hash() for every instance.
		testutil.FatalIf(t, String("foo").hash() == String("bar").hash(), "unexpected hash collision")
	})
}
