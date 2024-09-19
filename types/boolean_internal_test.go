package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestBoolean(t *testing.T) {
	t.Parallel()
	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		testutil.Equals(t, Boolean(true).hash(), Boolean(true).hash())
		testutil.Equals(t, Boolean(false).hash(), Boolean(false).hash())
		testutil.FatalIf(t, Boolean(true).hash() == Boolean(false).hash(), "unexpected hash collision")
	})
}
