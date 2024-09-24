package types_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/mapset"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntities(t *testing.T) {
	t.Parallel()
	t.Run("Clone", func(t *testing.T) {
		t.Parallel()
		e := types.Entities{
			types.EntityUID{Type: "A", ID: "A"}: {},
			types.EntityUID{Type: "A", ID: "B"}: {},
			types.EntityUID{Type: "B", ID: "A"}: {},
			types.EntityUID{Type: "B", ID: "B"}: {},
		}
		clone := e.Clone()
		testutil.Equals(t, clone, e)
	})

}

func assertJSONEquals(t *testing.T, e any, want string) {
	b, err := json.MarshalIndent(e, "", "\t")
	testutil.OK(t, err)

	var wantBuf bytes.Buffer
	err = json.Indent(&wantBuf, []byte(want), "", "\t")
	testutil.OK(t, err)
	testutil.Equals(t, string(b), wantBuf.String())
}

func TestEntitiesJSON(t *testing.T) {
	t.Parallel()
	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		e := types.Entities{}
		ent := &types.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    mapset.MapSet[types.EntityUID]{},
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		ent2 := &types.Entity{
			UID:        types.NewEntityUID("Type", "id2"),
			Parents:    *mapset.FromSlice([]types.EntityUID{ent.UID}),
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		e[ent.UID] = ent
		e[ent2.UID] = ent2
		assertJSONEquals(
			t,
			e,
			`[
				{"uid": {"type": "Type", "id": "id"}, "parents": [], "attrs": {"key": 42}},
				{"uid": {"type": "Type" ,"id" :"id2"}, "parents": [{"type":"Type","id":"id"}], "attrs": {"key": 42}}
			]`)
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()
		b := []byte(`[{"uid":{"type":"Type","id":"id"},"parents":[],"attrs":{"key":42}}]`)
		var e types.Entities
		err := json.Unmarshal(b, &e)
		testutil.OK(t, err)
		want := types.Entities{}
		ent := &types.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    *types.NewEntityUIDSet(),
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		want[ent.UID] = ent
		testutil.Equals(t, e, want)
	})

	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var e types.Entities
		err := e.UnmarshalJSON([]byte(`!@#$`))
		testutil.Error(t, err)
	})
}
