package exptypes

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

func TestUnmarshalValueString(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`"hello"`), resolved.StringType{})
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.String("hello")))
}

func TestUnmarshalValueLong(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`42`), resolved.LongType{})
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.Long(42)))
}

func TestUnmarshalValueLongNegative(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`-100`), resolved.LongType{})
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.Long(-100)))
}

func TestUnmarshalValueBool(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`true`), resolved.BoolType{})
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.Boolean(true)))
}

func TestUnmarshalValueBoolFalse(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`false`), resolved.BoolType{})
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.Boolean(false)))
}

func TestUnmarshalValueEntityUIDImplicit(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`{"type":"User","id":"alice"}`), resolved.EntityType("User"))
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewEntityUID("User", "alice")))
}

func TestUnmarshalValueEntityUIDExplicit(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON(
		[]byte(`{"__entity":{"type":"User","id":"alice"}}`),
		resolved.EntityType("User"),
	)
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewEntityUID("User", "alice")))
}

func TestUnmarshalValueIPAddr(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`{"__extn":{"fn":"ip","arg":"10.0.0.1"}}`), resolved.ExtensionType("ipaddr"))
	testutil.OK(t, err)
	expected, _ := types.ParseIPAddr("10.0.0.1")
	testutil.Equals(t, v, types.Value(expected))
}

func TestUnmarshalValueIPAddrBareString(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`"10.0.0.1"`), resolved.ExtensionType("ipaddr"))
	testutil.OK(t, err)
	expected, _ := types.ParseIPAddr("10.0.0.1")
	testutil.Equals(t, v, types.Value(expected))
}

func TestUnmarshalValueDecimal(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`{"__extn":{"fn":"decimal","arg":"1.23"}}`), resolved.ExtensionType("decimal"))
	testutil.OK(t, err)
	expected, _ := types.ParseDecimal("1.23")
	testutil.Equals(t, v, types.Value(expected))
}

func TestUnmarshalValueDecimalBareString(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`"1.23"`), resolved.ExtensionType("decimal"))
	testutil.OK(t, err)
	expected, _ := types.ParseDecimal("1.23")
	testutil.Equals(t, v, types.Value(expected))
}

func TestUnmarshalValueDatetime(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`{"__extn":{"fn":"datetime","arg":"2024-01-01"}}`), resolved.ExtensionType("datetime"))
	testutil.OK(t, err)
	expected, _ := types.ParseDatetime("2024-01-01")
	testutil.Equals(t, v, types.Value(expected))
}

func TestUnmarshalValueDuration(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`{"__extn":{"fn":"duration","arg":"1h30m"}}`), resolved.ExtensionType("duration"))
	testutil.OK(t, err)
	expected, _ := types.ParseDuration("1h30m")
	testutil.Equals(t, v, types.Value(expected))
}

func TestUnmarshalValueSet(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON(
		[]byte(`["a", "b"]`),
		resolved.SetType{Element: resolved.StringType{}},
	)
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewSet(types.String("a"), types.String("b"))))
}

func TestUnmarshalValueSetOfEntities(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON(
		[]byte(`[{"type":"User","id":"alice"}]`),
		resolved.SetType{Element: resolved.EntityType("User")},
	)
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewSet(types.NewEntityUID("User", "alice"))))
}

func TestUnmarshalValueEmptySet(t *testing.T) {
	t.Parallel()
	v, err := unmarshalValueJSON([]byte(`[]`), resolved.SetType{Element: resolved.StringType{}})
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewSet()))
}

func TestUnmarshalValueRecord(t *testing.T) {
	t.Parallel()
	schema := resolved.RecordType{
		"name": {Type: resolved.StringType{}, Optional: false},
		"age":  {Type: resolved.LongType{}, Optional: true},
	}
	v, err := unmarshalValueJSON([]byte(`{"name":"alice","age":30}`), schema)
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewRecord(types.RecordMap{
		"name": types.String("alice"),
		"age":  types.Long(30),
	})))
}

func TestUnmarshalValueRecordOptionalMissing(t *testing.T) {
	t.Parallel()
	schema := resolved.RecordType{
		"name": {Type: resolved.StringType{}, Optional: false},
		"age":  {Type: resolved.LongType{}, Optional: true},
	}
	v, err := unmarshalValueJSON([]byte(`{"name":"alice"}`), schema)
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewRecord(types.RecordMap{
		"name": types.String("alice"),
	})))
}

// The key test: {"type":"X","id":"Y"} parsed as Record, NOT EntityUID
func TestUnmarshalValueRecordWithTypeAndID(t *testing.T) {
	t.Parallel()
	schema := resolved.RecordType{
		"type": {Type: resolved.StringType{}, Optional: false},
		"id":   {Type: resolved.StringType{}, Optional: false},
	}
	v, err := unmarshalValueJSON([]byte(`{"type":"User","id":"alice"}`), schema)
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewRecord(types.RecordMap{
		"type": types.String("User"),
		"id":   types.String("alice"),
	})))
}

func TestUnmarshalValueRecordNested(t *testing.T) {
	t.Parallel()
	schema := resolved.RecordType{
		"inner": {Type: resolved.RecordType{
			"value": {Type: resolved.LongType{}, Optional: false},
		}, Optional: false},
	}
	v, err := unmarshalValueJSON([]byte(`{"inner":{"value":42}}`), schema)
	testutil.OK(t, err)
	testutil.Equals(t, v, types.Value(types.NewRecord(types.RecordMap{
		"inner": types.NewRecord(types.RecordMap{
			"value": types.Long(42),
		}),
	})))
}

func TestUnmarshalValueRecordMissingRequired(t *testing.T) {
	t.Parallel()
	schema := resolved.RecordType{
		"name": {Type: resolved.StringType{}, Optional: false},
	}
	_, err := unmarshalValueJSON([]byte(`{}`), schema)
	testutil.Error(t, err)
}

func TestUnmarshalValueRecordUnknownAttribute(t *testing.T) {
	t.Parallel()
	schema := resolved.RecordType{
		"name": {Type: resolved.StringType{}, Optional: false},
	}
	_, err := unmarshalValueJSON([]byte(`{"name":"alice","extra":"bad"}`), schema)
	testutil.Error(t, err)
}

func TestUnmarshalValueTypeMismatchStringGotNumber(t *testing.T) {
	t.Parallel()
	_, err := unmarshalValueJSON([]byte(`42`), resolved.StringType{})
	testutil.Error(t, err)
}

func TestUnmarshalValueTypeMismatchLongGotString(t *testing.T) {
	t.Parallel()
	_, err := unmarshalValueJSON([]byte(`"hello"`), resolved.LongType{})
	testutil.Error(t, err)
}

func TestUnmarshalValueTypeMismatchBoolGotObject(t *testing.T) {
	t.Parallel()
	_, err := unmarshalValueJSON([]byte(`{"a":1}`), resolved.BoolType{})
	testutil.Error(t, err)
}
