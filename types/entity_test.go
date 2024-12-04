package types_test

import (
	"testing"

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
