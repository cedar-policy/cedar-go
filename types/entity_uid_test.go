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

	t.Run("Marshal EntityUID round trip", func(t *testing.T) {
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
		marshalFuncs := []struct {
			name      string
			marshal   func(types.EntityUID) ([]byte, error)
			unmarshal func([]byte, *types.EntityUID) error
		}{
			{
				name:    "MarshalBinary",
				marshal: func(uid types.EntityUID) ([]byte, error) { return uid.MarshalBinary() },
				unmarshal: func(bin []byte, uid *types.EntityUID) error {
					return uid.UnmarshalBinary(bin)
				},
			},
			{
				name:    "MarshalCedar",
				marshal: func(uid types.EntityUID) ([]byte, error) { return uid.MarshalCedar(), nil },
				unmarshal: func(bin []byte, uid *types.EntityUID) error {
					return (uid).UnmarshalCedar(bin)
				},
			},
		}

		for _, marshalFunc := range marshalFuncs {
			t.Run(marshalFunc.name, func(t *testing.T) {
				t.Parallel()
				for _, testCase := range testCases {
					t.Run(testCase.bin, func(t *testing.T) {
						t.Parallel()
						uid := types.NewEntityUID(cedar.EntityType(testCase.typ), cedar.String(testCase.id))
						gotBin, err := marshalFunc.marshal(uid)
						testutil.OK(t, err)

						wantBin := []byte(testCase.bin)
						testutil.Equals(t, gotBin, wantBin)

						want := types.NewEntityUID(cedar.EntityType(testCase.typ), cedar.String(testCase.id))
						got := types.EntityUID{}
						err = marshalFunc.unmarshal(gotBin, &got)
						testutil.OK(t, err)
						testutil.Equals(t, got.String(), want.String())
						testutil.FatalIf(t, !uid.Equal(got), "expected %v to not equal %v", got, want)
					})
				}
			})
		}
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
			{"partial", `Type::"`},
			{"unescaped unicode", `Type::"\u00ab"`},
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
