package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestDecimalInternal(t *testing.T) {
	t.Parallel()
	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		testutil.Equals(t, testutil.Must(NewDecimalFromInt(42)).hash(), testutil.Must(NewDecimalFromInt(42)).hash())
		testutil.Equals(t, testutil.Must(NewDecimalFromInt(-42)).hash(), testutil.Must(NewDecimalFromInt(-42)).hash())

		// This isn't necessarily true for all values of Decimal, but we want to ensure we aren't just returning the
		// same hash value for Decimal.hash() for every instance.
		testutil.FatalIf(
			t,
			testutil.Must(NewDecimal(42, 0)).hash() ==
				testutil.Must(NewDecimal(1337, 0)).hash(), "unexpected hash collision",
		)
	})
}
