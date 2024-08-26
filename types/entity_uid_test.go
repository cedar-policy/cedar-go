package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntity(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		twoElems := types.EntityUID{"type", "id"}
		twoElems2 := types.EntityUID{"type", "id"}
		differentValues := types.EntityUID{"asdf", "vfds"}
		testutil.FatalIf(t, !twoElems.Equal(twoElems), "%v not Equal to %v", twoElems, twoElems)
		testutil.FatalIf(t, !twoElems.Equal(twoElems2), "%v not Equal to %v", twoElems, twoElems2)
		testutil.FatalIf(t, twoElems.Equal(differentValues), "%v Equal to %v", twoElems, differentValues)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.EntityUID{Type: "type", ID: "id"}.String(), `type::"id"`)
		testutil.Equals(t, types.EntityUID{Type: "namespace::type", ID: "id"}.String(), `namespace::type::"id"`)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, string(types.EntityUID{"type", "id"}.MarshalCedar()), `type::"id"`)
	})

}
