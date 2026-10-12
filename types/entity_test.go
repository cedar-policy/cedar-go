package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntityIsZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		uid  types.EntityUID
		want bool
	}{
		{"empty", types.EntityUID{}, true},
		{"empty-type", types.NewEntityUID("one", ""), false},
		{"empty-id", types.NewEntityUID("", "one"), false},
		{"not-empty", types.NewEntityUID("one", "two"), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.uid.IsZero(), tt.want)
		})
	}
}

func TestEntityMarshalJSON(t *testing.T) {
	t.Parallel()
	e := types.Entity{
		UID: types.NewEntityUID("FooType", "1"),
		Parents: types.NewEntityUIDSet(
			types.NewEntityUID("BazType", "1"),
			types.NewEntityUID("BarType", "2"),
			types.NewEntityUID("BarType", "1"),
			types.NewEntityUID("QuuxType", "30"),
			types.NewEntityUID("QuuxType", "3"),
		),
		Attributes: types.Record{},
		Tags:       types.Record{},
	}

	testutil.JSONMarshalsTo(t, e,
		`{
			"uid": {"type":"FooType","id":"1"},
			"parents": [
				{"type":"BarType","id":"1"},
				{"type":"BarType","id":"2"},
				{"type":"BazType","id":"1"},
				{"type":"QuuxType","id":"3"},
				{"type":"QuuxType","id":"30"}
			],
			"attrs":{},
			"tags":{}
		}`)
}

func TestEntityUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("exact keys", func(t *testing.T) {
		t.Parallel()
		var got types.Entity
		err := json.Unmarshal([]byte(`{
			"uid": {"type": "FooType", "id": "1"},
			"parents": [{"type": "BarType", "id": "2"}],
			"attrs": {"a": 1},
			"tags": {"t": 2}
		}`), &got)
		testutil.OK(t, err)
		testutil.Equals(t, got, types.Entity{
			UID:        types.NewEntityUID("FooType", "1"),
			Parents:    types.NewEntityUIDSet(types.NewEntityUID("BarType", "2")),
			Attributes: types.NewRecord(types.RecordMap{"a": types.Long(1)}),
			Tags:       types.NewRecord(types.RecordMap{"t": types.Long(2)}),
		})
	})

	t.Run("unknown keys ignored", func(t *testing.T) {
		t.Parallel()
		var got types.Entity
		err := json.Unmarshal([]byte(`{"uid": {"type": "FooType", "id": "1"}, "parents": [], "attrs": {}, "other": 1}`), &got)
		testutil.OK(t, err)
		testutil.Equals(t, got.UID, types.NewEntityUID("FooType", "1"))
	})

	// Keys are case-sensitive, as in the Rust SDK: a miscased key is an unknown key, so it is ignored.
	miscased := []struct {
		name  string
		input string
		want  types.Entity
	}{
		{"uid", `{"UID": {"type": "FooType", "id": "1"}, "parents": [], "attrs": {}}`,
			types.Entity{Parents: types.NewEntityUIDSet(), Attributes: types.NewRecord(nil)}},
		{"parents", `{"uid": {"type": "FooType", "id": "1"}, "Parents": [{"type": "BarType", "id": "2"}], "attrs": {}}`,
			types.Entity{UID: types.NewEntityUID("FooType", "1"), Attributes: types.NewRecord(nil)}},
		{"attrs", `{"uid": {"type": "FooType", "id": "1"}, "parents": [], "ATTRS": {"a": 1}}`,
			types.Entity{UID: types.NewEntityUID("FooType", "1"), Parents: types.NewEntityUIDSet()}},
		{"tags", `{"uid": {"type": "FooType", "id": "1"}, "parents": [], "attrs": {}, "Tags": {"t": 2}}`,
			types.Entity{UID: types.NewEntityUID("FooType", "1"), Parents: types.NewEntityUIDSet(), Attributes: types.NewRecord(nil)}},
	}
	for _, tt := range miscased {
		t.Run("miscased "+tt.name, func(t *testing.T) {
			t.Parallel()
			var got types.Entity
			err := json.Unmarshal([]byte(tt.input), &got)
			testutil.OK(t, err)
			testutil.Equals(t, got, tt.want)
		})
	}
}

func TestEntityTagMarshalJSON(t *testing.T) {
	t.Parallel()
	e := types.Entity{
		UID:        types.NewEntityUID("FooType", "1"),
		Parents:    types.NewEntityUIDSet(),
		Attributes: types.Record{},
		Tags: types.NewRecord(types.RecordMap{
			"key":      types.String("value"),
			"entity":   types.NewEntityUID("FootType", "1"),
			"datetime": types.NewDatetime(time.Unix(0, 0)),
		}),
	}

	testutil.JSONMarshalsTo(t, e,
		`{
			"uid": {"type":"FooType","id":"1"},
			"parents": [],
			"attrs":{},
            "tags": {
        		"datetime": {
        			"__extn": {
        				"fn": "datetime",
        				"arg": "1970-01-01T00:00:00.000Z"
        			}
        		},
        		"entity": {
        			"__entity": {
        				"type": "FootType",
        				"id": "1"
        			}
        		},
        		"key": "value"
            }
		}`)
}

func TestEntityEqual(t *testing.T) {
	t.Parallel()

	baseEntity := types.Entity{
		UID: types.NewEntityUID("FooType", "1"),
		Parents: types.NewEntityUIDSet(
			types.NewEntityUID("BarType", "1"),
			types.NewEntityUID("BazType", "2"),
		),
		Attributes: types.NewRecord(types.RecordMap{
			"attr1": types.String("value1"),
			"attr2": types.Long(42),
		}),
		Tags: types.NewRecord(types.RecordMap{
			"tag1": types.String("tagvalue1"),
		}),
	}

	tests := []struct {
		name   string
		entity types.Entity
		other  types.Entity
		want   bool
	}{
		{
			name:   "same content",
			entity: baseEntity,
			other:  baseEntity,
			want:   true,
		},
		{
			name:   "different UID",
			entity: baseEntity,
			other: types.Entity{
				UID:        types.NewEntityUID("FooType", "2"),
				Parents:    baseEntity.Parents,
				Attributes: baseEntity.Attributes,
				Tags:       baseEntity.Tags,
			},
			want: false,
		},
		{
			name:   "different parents",
			entity: baseEntity,
			other: types.Entity{
				UID: baseEntity.UID,
				Parents: types.NewEntityUIDSet(
					types.NewEntityUID("BarType", "1"),
					types.NewEntityUID("DifferentType", "2"),
				),
				Attributes: baseEntity.Attributes,
				Tags:       baseEntity.Tags,
			},
			want: false,
		},
		{
			name:   "different attributes",
			entity: baseEntity,
			other: types.Entity{
				UID:     baseEntity.UID,
				Parents: baseEntity.Parents,
				Attributes: types.NewRecord(types.RecordMap{
					"attr1": types.String("different_value"),
					"attr2": types.Long(42),
				}),
				Tags: baseEntity.Tags,
			},
			want: false,
		},
		{
			name:   "different tags",
			entity: baseEntity,
			other: types.Entity{
				UID:        baseEntity.UID,
				Parents:    baseEntity.Parents,
				Attributes: baseEntity.Attributes,
				Tags: types.NewRecord(types.RecordMap{
					"tag1": types.String("different_tagvalue"),
				}),
			},
			want: false,
		},
		{
			name:   "empty entities",
			entity: types.Entity{},
			other:  types.Entity{},
			want:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.entity.Equal(tt.other), tt.want)
		})
	}
}
