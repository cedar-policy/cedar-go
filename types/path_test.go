package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntityType(t *testing.T) {
	t.Parallel()
	t.Run("pathFromSlice", func(t *testing.T) {
		t.Parallel()
		a := types.PathFromSlice([]string{"X", "Y"})
		testutil.Equals(t, a, types.Path("X::Y"))
	})

}
