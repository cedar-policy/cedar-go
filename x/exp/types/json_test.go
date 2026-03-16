package exptypes

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// --- coerceValue unit tests ---

func TestCoerceValueRecordToEntityUID(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"type": types.String("User"),
		"id":   types.String("alice"),
	})
	got := coerceValue(rec, resolved.EntityType("User"))
	testutil.Equals(t, got, types.Value(types.NewEntityUID("User", "alice")))
}

func TestCoerceValueRecordToEntityUIDExtraKeys(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"type":  types.String("User"),
		"id":    types.String("alice"),
		"extra": types.String("nope"),
	})
	got := coerceValue(rec, resolved.EntityType("User"))
	testutil.Equals(t, got, types.Value(rec))
}

func TestCoerceValueRecordToEntityUIDNonStringID(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"type": types.String("User"),
		"id":   types.Long(42),
	})
	got := coerceValue(rec, resolved.EntityType("User"))
	testutil.Equals(t, got, types.Value(rec))
}

func TestCoerceValueRecordToEntityUIDNonStringType(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"type": types.Long(1),
		"id":   types.String("alice"),
	})
	got := coerceValue(rec, resolved.EntityType("User"))
	testutil.Equals(t, got, types.Value(rec))
}

func TestCoerceValueRecordToEntityUIDMissingID(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"type": types.String("User"),
		"name": types.String("alice"),
	})
	got := coerceValue(rec, resolved.EntityType("User"))
	testutil.Equals(t, got, types.Value(rec))
}

func TestCoerceValueRecordToEntityUIDMissingType(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"id":   types.String("alice"),
		"name": types.String("User"),
	})
	got := coerceValue(rec, resolved.EntityType("User"))
	testutil.Equals(t, got, types.Value(rec))
}

func TestCoerceValueExplicitEntityUIDUnchanged(t *testing.T) {
	t.Parallel()
	uid := types.NewEntityUID("User", "alice")
	got := coerceValue(uid, resolved.EntityType("User"))
	testutil.Equals(t, got, types.Value(uid))
}

func TestCoerceValueStringToIPAddr(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("10.0.0.1"), resolved.ExtensionType("ipaddr"))
	expected, _ := types.ParseIPAddr("10.0.0.1")
	testutil.Equals(t, got, types.Value(expected))
}

func TestCoerceValueStringToDecimal(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("1.23"), resolved.ExtensionType("decimal"))
	expected, _ := types.ParseDecimal("1.23")
	testutil.Equals(t, got, types.Value(expected))
}

func TestCoerceValueStringToDatetime(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("2024-01-01"), resolved.ExtensionType("datetime"))
	expected, _ := types.ParseDatetime("2024-01-01")
	testutil.Equals(t, got, types.Value(expected))
}

func TestCoerceValueStringToDuration(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("1h30m"), resolved.ExtensionType("duration"))
	expected, _ := types.ParseDuration("1h30m")
	testutil.Equals(t, got, types.Value(expected))
}

func TestCoerceValueStringToExtensionFailure(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("not-an-ip"), resolved.ExtensionType("ipaddr"))
	testutil.Equals(t, got, types.Value(types.String("not-an-ip")))
}

func TestCoerceValueStringToUnknownExtension(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("foo"), resolved.ExtensionType("unknown"))
	testutil.Equals(t, got, types.Value(types.String("foo")))
}

func TestCoerceValueExplicitExtensionUnchanged(t *testing.T) {
	t.Parallel()
	ip, _ := types.ParseIPAddr("10.0.0.1")
	got := coerceValue(ip, resolved.ExtensionType("ipaddr"))
	testutil.Equals(t, got, types.Value(ip))
}

func TestCoerceValueSetRecurse(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"type": types.String("User"),
		"id":   types.String("alice"),
	})
	set := types.NewSet(rec)
	got := coerceValue(set, resolved.SetType{Element: resolved.EntityType("User")})
	expected := types.NewSet(types.NewEntityUID("User", "alice"))
	testutil.Equals(t, got, types.Value(expected))
}

func TestCoerceValueSetNoChange(t *testing.T) {
	t.Parallel()
	set := types.NewSet(types.String("a"), types.String("b"))
	got := coerceValue(set, resolved.SetType{Element: resolved.StringType{}})
	testutil.Equals(t, got, types.Value(set))
}

func TestCoerceValueSetNonSet(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("not-a-set"), resolved.SetType{Element: resolved.StringType{}})
	testutil.Equals(t, got, types.Value(types.String("not-a-set")))
}

func TestCoerceValueRecordRecurse(t *testing.T) {
	t.Parallel()
	inner := types.NewRecord(types.RecordMap{
		"type": types.String("User"),
		"id":   types.String("alice"),
	})
	rec := types.NewRecord(types.RecordMap{
		"manager": inner,
		"name":    types.String("Bob"),
	})
	schema := resolved.RecordType{
		"manager": {Type: resolved.EntityType("User"), Optional: false},
		"name":    {Type: resolved.StringType{}, Optional: false},
	}
	got := coerceValue(rec, schema)
	expected := types.NewRecord(types.RecordMap{
		"manager": types.NewEntityUID("User", "alice"),
		"name":    types.String("Bob"),
	})
	testutil.Equals(t, got, types.Value(expected))
}

func TestCoerceValueRecordUnknownAttrsPassThrough(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"name":    types.String("Alice"),
		"unknown": types.String("extra"),
	})
	schema := resolved.RecordType{
		"name": {Type: resolved.StringType{}, Optional: false},
	}
	got := coerceValue(rec, schema)
	testutil.Equals(t, got, types.Value(rec))
}

func TestCoerceValueRecordNonRecord(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("not-a-record"), resolved.RecordType{})
	testutil.Equals(t, got, types.Value(types.String("not-a-record")))
}

func TestCoerceValueRecordEmpty(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(nil)
	got := coerceValue(rec, resolved.RecordType{
		"name": {Type: resolved.StringType{}, Optional: true},
	})
	testutil.Equals(t, got, types.Value(rec))
}

func TestCoerceValueRecordMissingSchemaAttr(t *testing.T) {
	t.Parallel()
	rec := types.NewRecord(types.RecordMap{
		"name": types.String("Alice"),
	})
	schema := resolved.RecordType{
		"name":  {Type: resolved.StringType{}, Optional: false},
		"age":   {Type: resolved.LongType{}, Optional: true},
	}
	got := coerceValue(rec, schema)
	testutil.Equals(t, got, types.Value(rec))
}

func TestCoerceValuePrimitiveUnchanged(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("hello"), resolved.StringType{})
	testutil.Equals(t, got, types.Value(types.String("hello")))

	got = coerceValue(types.Long(42), resolved.LongType{})
	testutil.Equals(t, got, types.Value(types.Long(42)))

	got = coerceValue(types.Boolean(true), resolved.BoolType{})
	testutil.Equals(t, got, types.Value(types.Boolean(true)))
}

func TestCoerceValueMismatchPassThrough(t *testing.T) {
	t.Parallel()
	got := coerceValue(types.String("hello"), resolved.LongType{})
	testutil.Equals(t, got, types.Value(types.String("hello")))
}

// --- EntityMap.UnmarshalJSONWithSchema tests ---

func testSchema() *resolved.Schema {
	return &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name: "User",
				Shape: resolved.RecordType{
					"name":    {Type: resolved.StringType{}, Optional: false},
					"manager": {Type: resolved.EntityType("User"), Optional: true},
				},
			},
		},
		Actions: map[types.EntityUID]resolved.Action{
			types.NewEntityUID("Action", "view"): {
				Entity: types.Entity{
					UID:     types.NewEntityUID("Action", "view"),
					Parents: types.NewEntityUIDSet(),
				},
			},
		},
		Enums: map[types.EntityType]resolved.Enum{},
	}
}

func TestEntityMapUnmarshalJSONWithSchema(t *testing.T) {
	t.Parallel()
	data := []byte(`[
		{
			"uid": {"type": "User", "id": "alice"},
			"attrs": {"name": "Alice", "manager": {"type": "User", "id": "bob"}},
			"parents": []
		},
		{
			"uid": {"type": "User", "id": "bob"},
			"attrs": {"name": "Bob"},
			"parents": []
		}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, testSchema())
	testutil.OK(t, err)

	m := types.EntityMap(em)
	alice, ok := m.Get(types.NewEntityUID("User", "alice"))
	testutil.FatalIf(t, !ok, "expected User::\"alice\"")
	mgr, ok := alice.Attributes.Get("manager")
	testutil.FatalIf(t, !ok, "expected manager attribute")
	testutil.Equals(t, mgr, types.Value(types.NewEntityUID("User", "bob")))
}

func TestEntityMapUnmarshalJSONWithSchemaAction(t *testing.T) {
	t.Parallel()
	data := []byte(`[
		{"uid": {"type": "Action", "id": "view"}, "attrs": {}, "parents": []}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, testSchema())
	testutil.OK(t, err)
}

func TestEntityMapUnmarshalJSONWithSchemaBadJSON(t *testing.T) {
	t.Parallel()
	var em EntityMap
	err := em.UnmarshalJSONWithSchema([]byte(`not json`), testSchema())
	testutil.Error(t, err)
}

func TestEntityMapUnmarshalJSONWithSchemaValidationError(t *testing.T) {
	t.Parallel()
	// Missing required attribute "name"
	data := []byte(`[
		{"uid": {"type": "User", "id": "alice"}, "attrs": {}, "parents": []}
	]`)

	var em EntityMap
	err := em.UnmarshalJSONWithSchema(data, testSchema())
	testutil.Error(t, err)
}

// --- Entity.UnmarshalJSONWithSchema tests ---

func TestEntityUnmarshalJSONWithSchema(t *testing.T) {
	t.Parallel()
	data := []byte(`{
		"uid": {"type": "User", "id": "alice"},
		"attrs": {"name": "Alice", "manager": {"type": "User", "id": "bob"}},
		"parents": []
	}`)

	var e Entity
	err := e.UnmarshalJSONWithSchema(data, testSchema())
	testutil.OK(t, err)

	mgr, ok := types.Entity(e).Attributes.Get("manager")
	testutil.FatalIf(t, !ok, "expected manager attribute")
	testutil.Equals(t, mgr, types.Value(types.NewEntityUID("User", "bob")))
}

func TestEntityUnmarshalJSONWithSchemaBadJSON(t *testing.T) {
	t.Parallel()
	var e Entity
	err := e.UnmarshalJSONWithSchema([]byte(`{bad`), testSchema())
	testutil.Error(t, err)
}

func TestEntityUnmarshalJSONWithSchemaValidationError(t *testing.T) {
	t.Parallel()
	// Unknown entity type
	data := []byte(`{
		"uid": {"type": "Unknown", "id": "x"},
		"attrs": {},
		"parents": []
	}`)

	var e Entity
	err := e.UnmarshalJSONWithSchema(data, testSchema())
	testutil.Error(t, err)
}

// --- coerceTagValues tests ---

func TestCoerceTagValues(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name:  "User",
				Shape: resolved.RecordType{
					"name": {Type: resolved.StringType{}, Optional: false},
				},
				Tags:  resolved.ExtensionType("ipaddr"),
			},
		},
		Actions: map[types.EntityUID]resolved.Action{},
		Enums:   map[types.EntityType]resolved.Enum{},
	}

	data := []byte(`{
		"uid": {"type": "User", "id": "alice"},
		"attrs": {"name": "Alice"},
		"parents": [],
		"tags": {"home": "10.0.0.1"}
	}`)

	var e Entity
	err := e.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)

	home, ok := types.Entity(e).Tags.Get("home")
	testutil.FatalIf(t, !ok, "expected home tag")
	expected, _ := types.ParseIPAddr("10.0.0.1")
	testutil.Equals(t, home, types.Value(expected))
}

func TestCoerceTagValuesEmptyTags(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name:  "User",
				Shape: resolved.RecordType{
					"name": {Type: resolved.StringType{}, Optional: false},
				},
				Tags: resolved.ExtensionType("ipaddr"),
			},
		},
		Actions: map[types.EntityUID]resolved.Action{},
		Enums:   map[types.EntityType]resolved.Enum{},
	}

	data := []byte(`{
		"uid": {"type": "User", "id": "alice"},
		"attrs": {"name": "Alice"},
		"parents": []
	}`)

	var e Entity
	err := e.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)
	testutil.Equals(t, types.Entity(e).Tags.Len(), 0)
}

func TestCoerceTagValuesNoChange(t *testing.T) {
	t.Parallel()
	schema := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name:  "User",
				Shape: resolved.RecordType{
					"name": {Type: resolved.StringType{}, Optional: false},
				},
				Tags:  resolved.StringType{},
			},
		},
		Actions: map[types.EntityUID]resolved.Action{},
		Enums:   map[types.EntityType]resolved.Enum{},
	}

	data := []byte(`{
		"uid": {"type": "User", "id": "alice"},
		"attrs": {"name": "Alice"},
		"parents": [],
		"tags": {"role": "admin"}
	}`)

	var e Entity
	err := e.UnmarshalJSONWithSchema(data, schema)
	testutil.OK(t, err)

	role, ok := types.Entity(e).Tags.Get("role")
	testutil.FatalIf(t, !ok, "expected role tag")
	testutil.Equals(t, role, types.Value(types.String("admin")))
}
