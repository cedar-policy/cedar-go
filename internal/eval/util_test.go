package eval

import (
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestUtil(t *testing.T) {
	t.Parallel()
	t.Run("Boolean", func(t *testing.T) {
		t.Parallel()
		t.Run("roundTrip", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToBool(types.Boolean(true))
			testutil.OK(t, err)
			testutil.Equals(t, v, true)
		})

		t.Run("toBoolOnNonBool", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToBool(types.Long(0))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v, false)
		})
	})

	t.Run("Long", func(t *testing.T) {
		t.Parallel()
		t.Run("roundTrip", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToLong(types.Long(42))
			testutil.OK(t, err)
			testutil.Equals(t, v, 42)
		})

		t.Run("toLongOnNonLong", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToLong(types.Boolean(true))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v, 0)
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		t.Run("roundTrip", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToString(types.String("hello"))
			testutil.OK(t, err)
			testutil.Equals(t, v, "hello")
		})

		t.Run("toStringOnNonString", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToString(types.Boolean(true))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v, "")
		})
	})

	t.Run("Set", func(t *testing.T) {
		t.Parallel()
		t.Run("roundTrip", func(t *testing.T) {
			t.Parallel()
			v := types.NewSet(types.Boolean(true), types.Long(1))
			slice, err := ValueToSet(v)
			testutil.OK(t, err)
			v2 := slice
			testutil.FatalIf(t, !v.Equal(v2), "got %v want %v", v, v2)
		})

		t.Run("ToSetOnNonSet", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToSet(types.Boolean(true))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v.Len(), 0)
		})
	})

	t.Run("Record", func(t *testing.T) {
		t.Parallel()
		t.Run("roundTrip", func(t *testing.T) {
			t.Parallel()
			v := types.NewRecord(types.RecordMap{
				"foo": types.Boolean(true),
				"bar": types.Long(1),
			})
			record, err := ValueToRecord(v)
			testutil.OK(t, err)
			v2 := record
			testutil.FatalIf(t, !v.Equal(v2), "got %v want %v", v, v2)
		})

		t.Run("toRecordOnNonRecord", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToRecord(types.String("hello"))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v, types.Record{})
		})
	})

	t.Run("Entity", func(t *testing.T) {
		t.Parallel()
		t.Run("roundTrip", func(t *testing.T) {
			t.Parallel()
			want := types.EntityUID{Type: "User", ID: "bananas"}
			v, err := ValueToEntity(want)
			testutil.OK(t, err)
			testutil.Equals(t, v, want)
		})
		t.Run("ToEntityOnNonEntity", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToEntity(types.String("hello"))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v, types.EntityUID{})
		})

	})

	t.Run("Datetime", func(t *testing.T) {
		t.Parallel()
		t.Run("roundTrip", func(t *testing.T) {
			t.Parallel()
			dv, err := types.ParseDatetime("2024-07-15T09:00:00Z")
			testutil.OK(t, err)
			v, err := ValueToDatetime(dv)
			testutil.OK(t, err)
			testutil.FatalIf(t, !v.Equal(dv), "got %v want %v", v, dv)
		})

		t.Run("toDatetimeOnNonDatetime", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToDatetime(types.Boolean(true))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v, types.Datetime{})
		})

	})

	t.Run("Decimal", func(t *testing.T) {
		t.Parallel()
		t.Run("roundTrip", func(t *testing.T) {
			t.Parallel()
			dv, err := types.ParseDecimal("1.20")
			testutil.OK(t, err)
			v, err := ValueToDecimal(dv)
			testutil.OK(t, err)
			testutil.FatalIf(t, !v.Equal(dv), "got %v want %v", v, dv)
		})

		t.Run("toDecimalOnNonDecimal", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToDecimal(types.Boolean(true))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v, types.Decimal{})
		})

	})

	t.Run("IP", func(t *testing.T) {
		t.Parallel()

		t.Run("toIPOnNonIP", func(t *testing.T) {
			t.Parallel()
			v, err := ValueToIP(types.Boolean(true))
			testutil.ErrorIs(t, err, ErrType)
			testutil.Equals(t, v, types.IPAddr{})
		})
	})

}

func TestTypeName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   types.Value
		out  string
	}{

		{"boolean", types.Boolean(true), "bool"},
		{"datetime", types.NewDatetime(time.UnixMilli(42)), "datetime"},
		{"decimal", testutil.Must(types.NewDecimalFromInt(42)), "decimal"},
		{"entityUID", types.NewEntityUID("T", "42"), "(entity of type `T`)"},
		{"ip", types.IPAddr{}, "IP"},
		{"long", types.Long(42), "long"},
		{"record", types.Record{}, "record"},
		{"set", types.Set{}, "set"},
		{"string", types.String("test"), "string"},
		{"nil", nil, "unknown type"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := TypeName(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}
