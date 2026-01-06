package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntity(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		twoElems := types.EntityUID{"type", "id"}
		twoElems2 := types.EntityUID{"type", "id"}
		differentValues := types.EntityUID{"asdf", "vfds"}
		testutil.FatalIf(t, !twoElems.Equal(twoElems), "%v not Equal to %v", twoElems, twoElems)
		testutil.FatalIf(t, !twoElems.Equal(twoElems2), "%v not Equal to %v", twoElems, twoElems2)
		testutil.FatalIf(t, twoElems.Equal(differentValues), "%v Equal to %v", twoElems, differentValues)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.EntityUID{Type: "type", ID: "id"}.String(), `type::"id"`)
		testutil.Equals(t, types.EntityUID{Type: "namespace::type", ID: "id"}.String(), `namespace::type::"id"`)
	})

	t.Run("MarshalCedar round trip", func(t *testing.T) {
		t.Parallel()
		gotBin := types.NewEntityUID("type", "id").MarshalCedar()
		wantBin := []byte(`type::"id"`)
		testutil.Equals(t, gotBin, wantBin)

		want := types.NewEntityUID("type", "id")
		got := types.EntityUID{}
		err := got.UnmarshalCedar(gotBin)
		testutil.OK(t, err)
		testutil.Equals(t, got, want)
	})

	t.Run("UnmarshalCedar invalid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			input string
		}{
			{"unquoted string", `Type::id`},
			{"missing double colon", `Type"id"`},
			{"missing a quote at beginning", `Type::id"`},
			{"missing a quote at end", `Type::"id`},
			{"empty input", ``},
			{"just quoted string", `"id"`},
			{"no type", `::"id"`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var got types.EntityUID
				err := got.UnmarshalCedar([]byte(tt.input))
				testutil.FatalIf(t, err == nil, "expected error for input %q, got nil", tt.input)
			})
		}
	})

	t.Run("UnmarshalCedar with ::\" in ID", func(t *testing.T) {
		t.Parallel()
		// ID ending in :: should be parsed correctly
		want := types.NewEntityUID("X::Y", "asdf::")
		gotBin := want.MarshalCedar()
		testutil.Equals(t, gotBin, []byte(`X::Y::"asdf::"`))

		got := types.EntityUID{}
		err := got.UnmarshalCedar(gotBin)
		testutil.OK(t, err)
		testutil.Equals(t, got, want)
	})

	t.Run("MarshalBinary round trip", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			typ, id, bin string
		}{
			{"namespace::type", "id", `namespace::type::"id"`},
			{"namespace::type", "", `namespace::type::""`},
			{"X::Y", "abc::", `X::Y::"abc::"`},
			{"Search::Algorithm", "A*", `Search::Algorithm::"A*"`},
			{"Super", "*", `Super::"*"`},
		}
		for _, testCase := range testCases {
			t.Run(testCase.typ+"::"+testCase.id, func(t *testing.T) {
				t.Parallel()
				uid := types.NewEntityUID(cedar.EntityType(testCase.typ), cedar.String(testCase.id))
				gotBin := uid.MarshalBinary()

				wantBin := []byte(testCase.bin)
				testutil.Equals(t, gotBin, wantBin)

				want := types.NewEntityUID(cedar.EntityType(testCase.typ), cedar.String(testCase.id))
				got := types.EntityUID{}
				err := got.UnmarshalBinary(gotBin)
				testutil.OK(t, err)
				testutil.Equals(t, got.String(), want.String())
				testutil.FatalIf(t, !uid.Equal(got), "expected %v to not equal %v", got, want)
			})
		}
	})

}

func TestEntityUIDSet(t *testing.T) {
	t.Parallel()

	t.Run("new empty set", func(t *testing.T) {
		emptySets := []types.EntityUIDSet{
			{},
			types.NewEntityUIDSet(),
		}

		for _, es := range emptySets {
			testutil.Equals(t, es.Len(), 0)
			testutil.Equals(t, emptySets[0].Equal(es), true)
			testutil.Equals(t, es.Equal(emptySets[0]), true)
		}
	})

	t.Run("new set", func(t *testing.T) {
		a := types.NewEntityUID("typeA", "1")
		b := types.NewEntityUID("typeB", "2")
		o := types.NewEntityUID("typeO", "2")
		s := types.NewEntityUIDSet(a, b, o)

		testutil.Equals(t, s.Len(), 3)
		testutil.Equals(t, s.Contains(a), true)
		testutil.Equals(t, s.Contains(b), true)
		testutil.Equals(t, s.Contains(o), true)
	})
}
