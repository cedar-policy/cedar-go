package exptypes

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestCoerceEntityUID(t *testing.T) {
	t.Parallel()
	entity := func(typ string) types.Record {
		return types.NewRecord(types.RecordMap{"type": types.String(typ), "id": types.String("id")})
	}
	explicit := func(typ string) types.Record {
		return types.NewRecord(types.RecordMap{"__entity": entity(typ), "extra": types.True})
	}
	tests := []struct {
		name string
		in   types.Value
		want types.Value
	}{
		{"implicit", entity("A::B"), types.NewEntityUID("A::B", "id")},
		{"explicit", explicit("A::B"), types.NewEntityUID("A::B", "id")},
		{"implicitInvalidType", entity("A B"), entity("A B")},
		{"explicitInvalidType", explicit("A::if"), explicit("A::if")},
		{"notRecord", types.String("x"), types.String("x")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, coerceEntityUID(tt.in), tt.want)
		})
	}
}
