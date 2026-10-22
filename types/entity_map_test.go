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

func TestEntitiesUnmarshalJSON(t *testing.T) {
	t.Parallel()

	user := types.NewEntityUID("User", "alice")
	group := types.NewEntityUID("Group", "admins")
	tests := []struct {
		name string
		in   string
		want types.EntityMap
	}{
		{"full", `[{"uid":{"type":"User","id":"alice"},"parents":[{"type":"Group","id":"admins"}],"attrs":{"a":1},"tags":{"t":2}}]`,
			types.EntityMap{user: {UID: user, Parents: types.NewEntityUIDSet(group), Attributes: types.NewRecord(types.RecordMap{"a": types.Long(1)}), Tags: types.NewRecord(types.RecordMap{"t": types.Long(2)})}}},
		{"explicit UIDs", `[{"uid":{"__entity":{"type":"User","id":"alice"}},"parents":[{"__entity":{"type":"Group","id":"admins"}}],"attrs":{}}]`,
			types.EntityMap{user: {UID: user, Parents: types.NewEntityUIDSet(group)}}},
		{"no tags", `[{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{}}]`,
			types.EntityMap{user: {UID: user, Parents: types.NewEntityUIDSet()}}},
		{"unknown member", `[{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{},"other":[1,{"x":2}]}]`,
			types.EntityMap{user: {UID: user, Parents: types.NewEntityUIDSet()}}},
		{"empty", `[]`, types.EntityMap{}},
		// cedar-go v1.2.0 through v1.8.0 encoded an empty EntityMap as null.
		{"null", `null`, types.EntityMap{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got types.EntityMap
			testutil.OK(t, json.Unmarshal([]byte(tt.in), &got))
			testutil.Equals(t, got, tt.want)
		})
	}

	errTests := []struct {
		name string
		in   string
	}{
		{"missing uid", `[{"parents":[],"attrs":{}}]`},
		{"missing parents", `[{"uid":{"type":"User","id":"alice"},"attrs":{}}]`},
		{"missing attrs", `[{"uid":{"type":"User","id":"alice"},"parents":[]}]`},
		{"miscased members", `[{"UID":{"type":"User","id":"alice"},"Parents":[],"Attrs":{}}]`},
		{"null uid", `[{"uid":null,"parents":[],"attrs":{}}]`},
		{"null parents", `[{"uid":{"type":"User","id":"alice"},"parents":null,"attrs":{}}]`},
		{"null attrs", `[{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":null}]`},
		{"null tags", `[{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{},"tags":null}]`},
		{"null entity", `[null]`},
		{"non-object entity", `[1]`},
		{"duplicate member", `[{"uid":{"type":"User","id":"alice"},"uid":{"type":"User","id":"bob"},"parents":[],"attrs":{}}]`},
		{"duplicate entity", `[{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{"a":1}},{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{"a":2}}]`},
		{"invalid tags", `[{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{},"tags":[]}]`},
		{"invalid attrs", `[{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":[]}]`},
		{"invalid unknown member", `[{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{},"other":[1,}]`},
		{"invalid member name", `[{"uid":{"type":"User","id":"alice"},1:2}]`},
		{"truncated", `[{"uid":{"type":"User","id":"alice"}`},
		{"not an array", `{}`},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got types.EntityMap
			testutil.Error(t, json.Unmarshal([]byte(tt.in), &got))
		})
	}

	t.Run("Entity", func(t *testing.T) {
		t.Parallel()
		var got types.Entity
		testutil.OK(t, json.Unmarshal([]byte(`{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{}}`), &got))
		testutil.Equals(t, got, types.Entity{UID: user, Parents: types.NewEntityUIDSet()})
	})

	// encoding/json (v1) allows duplicate member names, but decoding an Entity
	// with it must not.
	t.Run("EntityDuplicateMember", func(t *testing.T) {
		t.Parallel()
		var got types.Entity
		err := json.Unmarshal([]byte(`{"uid":{"type":"User","id":"alice"},"uid":{"type":"User","id":"bob"},"attrs":{}}`), &got)
		testutil.Error(t, err)
	})
}

func TestEntitiesMarshalJSONEmpty(t *testing.T) {
	t.Parallel()
	testutil.JSONMarshalsTo(t, types.EntityMap{}, `[]`)
	var nilMap types.EntityMap
	testutil.JSONMarshalsTo(t, nilMap, `[]`)
}
