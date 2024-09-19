package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestEntityUID(t *testing.T) {
	t.Parallel()
	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		testutil.Equals(t, NewEntityUID("type", "id").hash(), NewEntityUID("type", "id").hash())
		testutil.Equals(t, NewEntityUID("type2", "id2").hash(), NewEntityUID("type2", "id2").hash())

		// This isn't necessarily true for all EntityUIDs, but we want to make sure we're not just returning the same
		// hash value for all EntityUIDs
		testutil.FatalIf(t, NewEntityUID("type", "id").hash() == NewEntityUID("type2", "id2").hash(), "unexpected hash collision")
	})
}
