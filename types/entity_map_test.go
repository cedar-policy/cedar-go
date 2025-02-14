package types_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntities(t *testing.T) {
	t.Parallel()
	t.Run("Clone", func(t *testing.T) {
		t.Parallel()
		e := types.EntityMap{
			types.EntityUID{Type: "A", ID: "A"}: {},
			types.EntityUID{Type: "A", ID: "B"}: {},
			types.EntityUID{Type: "B", ID: "A"}: {},
			types.EntityUID{Type: "B", ID: "B"}: {},
		}
		clone := e.Clone()
		testutil.Equals(t, clone, e)
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		ent := types.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		e := types.EntityMap{
			ent.UID: ent,
		}
		got, ok := e.Get(ent.UID)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, got, ent)
		_, ok = e.Get(types.NewEntityUID("Type", "id2"))
		testutil.Equals(t, ok, false)
	})
}

func TestEntitiesJSON(t *testing.T) {
	t.Parallel()
	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		e := types.EntityMap{}
		ent := types.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    types.EntityUIDSet{},
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		ent2 := types.Entity{
			UID:        types.NewEntityUID("Type", "id2"),
			Parents:    types.NewEntityUIDSet(ent.UID),
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		e[ent.UID] = ent
		e[ent2.UID] = ent2
		testutil.JSONMarshalsTo(
			t,
			e,
			`[
				{"uid": {"type": "Type", "id": "id"}, "parents": [], "attrs": {"key": 42}, "tags": {}},
				{"uid": {"type": "Type" ,"id" :"id2"}, "parents": [{"type":"Type","id":"id"}], "attrs": {"key": 42}, "tags":{}}
			]`)
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()
		b := []byte(`[{"uid":{"type":"Type","id":"id"},"parents":[],"attrs":{"key":42}}]`)
		var e types.EntityMap
		err := json.Unmarshal(b, &e)
		testutil.OK(t, err)
		want := types.EntityMap{}
		ent := types.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    types.NewEntityUIDSet(),
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		want[ent.UID] = ent
		testutil.Equals(t, e, want)
	})

	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var e types.EntityMap
		err := e.UnmarshalJSON([]byte(`!@#$`))
		testutil.Error(t, err)
	})
}
