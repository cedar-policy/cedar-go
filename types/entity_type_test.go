package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntityType(t *testing.T) {
	t.Parallel()
	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		a := types.EntityType("X")
		b := types.EntityType("X")
		c := types.EntityType("Y")
		testutil.Equals(t, a.Equal(b), true)
		testutil.Equals(t, b.Equal(a), true)
		testutil.Equals(t, a.Equal(c), false)
		testutil.Equals(t, c.Equal(a), false)
	})

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		a := types.EntityType("X")
		testutil.Equals(t, a.String(), "X")
	})
	t.Run("Cedar", func(t *testing.T) {
		t.Parallel()
		a := types.EntityType("X")
		testutil.Equals(t, a.MarshalCedar(), []byte("X"))
	})
	t.Run("ExplicitMarshalJSON", func(t *testing.T) {
		t.Parallel()
		a := types.EntityType("X")
		v, err := a.ExplicitMarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(v), `"X"`)
	})
	t.Run("pathFromSlice", func(t *testing.T) {
		t.Parallel()
		a := types.EntityTypeFromSlice([]string{"X", "Y"})
		testutil.Equals(t, a, types.EntityType("X::Y"))
	})

}
