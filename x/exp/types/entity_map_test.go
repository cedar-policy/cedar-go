package exptypes

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

func TestUnmarshalEntityMap(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name: "User",
				Shape: resolved.RecordType{
					"name": {Type: resolved.StringType{}, Optional: false},
				},
			},
		},
		Actions: map[types.EntityUID]resolved.Action{
			types.NewEntityUID("Action", "view"): {},
		},
	}

	data := []byte(`[
		{
			"uid": {"type": "User", "id": "alice"},
			"attrs": {"name": "Alice"},
			"parents": []
		},
		{
			"uid": {"type": "Action", "id": "view"},
			"attrs": {},
			"parents": []
		}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)

	m := types.EntityMap(em)
	userEntity, ok := m.Get(types.NewEntityUID("User", "alice"))
	testutil.FatalIf(t, !ok, "expected User::\"alice\" in entity map")
	name, ok := userEntity.Attributes.Get("name")
	testutil.FatalIf(t, !ok, "expected name attribute")
	testutil.Equals(t, name, types.Value(types.String("Alice")))
}

func TestUnmarshalEntityMapWithParents(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User":  {Name: "User", Shape: resolved.RecordType{}},
			"Group": {Name: "Group", Shape: resolved.RecordType{}},
		},
		Actions: map[types.EntityUID]resolved.Action{},
	}

	data := []byte(`[
		{
			"uid": {"type": "User", "id": "alice"},
			"attrs": {},
			"parents": [{"type": "Group", "id": "admins"}]
		},
		{
			"uid": {"type": "Group", "id": "admins"},
			"attrs": {},
			"parents": []
		}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)

	m := types.EntityMap(em)
	userEntity, ok := m.Get(types.NewEntityUID("User", "alice"))
	testutil.FatalIf(t, !ok, "expected User::\"alice\" in entity map")
	testutil.FatalIf(t, !userEntity.Parents.Contains(types.NewEntityUID("Group", "admins")),
		"expected Group::\"admins\" in parents")
}

func TestUnmarshalEntityMapUnknownType(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{},
		Actions:  map[types.EntityUID]resolved.Action{},
	}

	data := []byte(`[{"uid":{"type":"Unknown","id":"x"},"attrs":{},"parents":[]}]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.Error(t, err)
}

func TestUnmarshalEntityMapWithEntityUIDAttr(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name: "User",
				Shape: resolved.RecordType{
					"manager": {Type: resolved.EntityType("User"), Optional: true},
				},
			},
		},
		Actions: map[types.EntityUID]resolved.Action{},
	}

	data := []byte(`[
		{
			"uid": {"type": "User", "id": "alice"},
			"attrs": {"manager": {"type": "User", "id": "bob"}},
			"parents": []
		}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)

	m := types.EntityMap(em)
	userEntity, _ := m.Get(types.NewEntityUID("User", "alice"))
	mgr, ok := userEntity.Attributes.Get("manager")
	testutil.FatalIf(t, !ok, "expected manager attribute")
	testutil.Equals(t, mgr, types.Value(types.NewEntityUID("User", "bob")))
}

func TestUnmarshalEntityMapDuplicateIdentical(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {Name: "User", Shape: resolved.RecordType{}},
		},
		Actions: map[types.EntityUID]resolved.Action{},
	}

	data := []byte(`[
		{"uid":{"type":"User","id":"a"},"attrs":{},"parents":[]},
		{"uid":{"type":"User","id":"a"},"attrs":{},"parents":[]}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)
}

func TestUnmarshalEntityMapDuplicateDifferent(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name: "User",
				Shape: resolved.RecordType{
					"name": {Type: resolved.StringType{}, Optional: true},
				},
			},
		},
		Actions: map[types.EntityUID]resolved.Action{},
	}

	data := []byte(`[
		{"uid":{"type":"User","id":"a"},"attrs":{"name":"Alice"},"parents":[]},
		{"uid":{"type":"User","id":"a"},"attrs":{"name":"Bob"},"parents":[]}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.Error(t, err)
}

func TestUnmarshalEntityMapWithTags(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name:  "User",
				Shape: resolved.RecordType{},
				Tags:  resolved.StringType{},
			},
		},
		Actions: map[types.EntityUID]resolved.Action{},
	}

	data := []byte(`[
		{"uid":{"type":"User","id":"a"},"attrs":{},"parents":[],"tags":{"role":"admin"}}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)

	m := types.EntityMap(em)
	e, _ := m.Get(types.NewEntityUID("User", "a"))
	role, ok := e.Tags.Get("role")
	testutil.FatalIf(t, !ok, "expected role tag")
	testutil.Equals(t, role, types.Value(types.String("admin")))
}

func TestUnmarshalEntityMapTagsNotAllowed(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {Name: "User", Shape: resolved.RecordType{}, Tags: nil},
		},
		Actions: map[types.EntityUID]resolved.Action{},
	}

	data := []byte(`[
		{"uid":{"type":"User","id":"a"},"attrs":{},"parents":[],"tags":{"bad":"tag"}}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.Error(t, err)
}

func TestUnmarshalEntityMapMissingTags(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {Name: "User", Shape: resolved.RecordType{}},
		},
		Actions: map[types.EntityUID]resolved.Action{},
	}

	// tags field absent from JSON
	data := []byte(`[{"uid":{"type":"User","id":"a"},"attrs":{},"parents":[]}]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)
}
