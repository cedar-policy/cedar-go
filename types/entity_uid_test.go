package types_test

import (
	"encoding/json"
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
		// Bell char must be escaped as \u{7}, not Go's \a
		testutil.Equals(t, types.EntityUID{Type: "T", ID: "\a"}.String(), `T::"\u{7}"`)
		// Backspace must be escaped as \u{8}, not Go's \b
		testutil.Equals(t, types.EntityUID{Type: "T", ID: "\b"}.String(), `T::"\u{8}"`)
		// Form feed must be escaped as \u{c}, not Go's \f
		testutil.Equals(t, types.EntityUID{Type: "T", ID: "\f"}.String(), `T::"\u{c}"`)
		// Vertical tab must be escaped as \u{b}, not Go's \v
		testutil.Equals(t, types.EntityUID{Type: "T", ID: "\v"}.String(), `T::"\u{b}"`)
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

	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name  string
			input string
		}{
			{"explicit", `{ "__entity": { "type": "Type", "id": "id" } }`},
			{"implicit", `{ "type": "Type", "id": "id" }`},
			{"explicit - extra keys in inner", `{ "__entity": { "type": "Type", "id": "id", "floob": "blah" } }`},
			{"implicit - extra keys", `{ "type": "Type", "id": "id", "floob": "blah" }`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var got types.EntityUID
				err := got.UnmarshalJSON([]byte(tt.input))
				testutil.OK(t, err)
				testutil.Equals(t, got, types.NewEntityUID("Type", "id"))
			})
		}
	})

	t.Run("UnmarshalJSON invalid", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name  string
			input string
		}{
			{"non-object", `"floob"`},
			{"explicit - wrong type", `{ "__entity": "wrong type"`},
			{"implicit - type wrong", `{ "type": 123, "id": "123" }`},
			{"implicit - id wrong type", `{ "type": "Type", "id": 123 }`},
			{"implicit - only type", `{ "type": "Type" }`},
			{"implicit - only id", `{ "id": 123 }`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var got types.EntityUID
				err := got.UnmarshalJSON([]byte(tt.input))
				testutil.Error(t, err)
			})
		}
	})
}

func TestImplicitlyMarshaledEntityUID(t *testing.T) {
	t.Parallel()

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name string
			in   types.EntityUID
			want string
		}{
			{"simple", types.NewEntityUID("Type", "id"), `{"type":"Type","id":"id"}`},
			{"namespaced type", types.NewEntityUID("NS::Type", "id"), `{"type":"NS::Type","id":"id"}`},
			{"empty", types.EntityUID{}, `{"type":"","id":""}`},
			{"escaped id", types.NewEntityUID("Type", "a\"b\\c"), `{"type":"Type","id":"a\"b\\c"}`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				testutil.JSONMarshalsTo(t, types.ImplicitlyMarshaledEntityUID(tt.in), tt.want)
			})
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name  string
			input string
		}{
			{"implicit", `{ "type": "Type", "id": "id" }`},
			{"explicit", `{ "__entity": { "type": "Type", "id": "id" } }`},
			{"implicit - extra keys", `{ "type": "Type", "id": "id", "floob": "blah" }`},
			{"explicit - extra keys in inner", `{ "__entity": { "type": "Type", "id": "id", "floob": "blah" } }`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var got types.ImplicitlyMarshaledEntityUID
				testutil.OK(t, json.Unmarshal([]byte(tt.input), &got))
				testutil.Equals(t, got, types.ImplicitlyMarshaledEntityUID(types.NewEntityUID("Type", "id")))
			})
		}
	})

	t.Run("UnmarshalJSON invalid", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name  string
			input string
		}{
			{"non-object", `"floob"`},
			{"empty object", `{}`},
			{"explicit - wrong type", `{ "__entity": "wrong type" }`},
			{"implicit - type wrong type", `{ "type": 123, "id": "123" }`},
			{"implicit - id wrong type", `{ "type": "Type", "id": 123 }`},
			{"implicit - only type", `{ "type": "Type" }`},
			{"implicit - only id", `{ "id": "id" }`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var got types.ImplicitlyMarshaledEntityUID
				testutil.Error(t, json.Unmarshal([]byte(tt.input), &got))
			})
		}
	})

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()
		want := types.ImplicitlyMarshaledEntityUID(types.NewEntityUID("NS::Type", "a\"b\\c"))
		b, err := json.Marshal(want)
		testutil.OK(t, err)
		var got types.ImplicitlyMarshaledEntityUID
		testutil.OK(t, json.Unmarshal(b, &got))
		testutil.Equals(t, got, want)
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

	// Checks that callers still using encoding/json (v1) can encode and decode
	// an EntityUIDSet, which is an alias of an internal type.
	t.Run("json roundtrip", func(t *testing.T) {
		t.Parallel()
		want := types.NewEntityUIDSet(types.NewEntityUID("User", "bob"), types.NewEntityUID("User", "alice"))
		b, err := json.Marshal(want)
		testutil.OK(t, err)
		testutil.Equals(t, string(b), `[{"__entity":{"type":"User","id":"alice"}},{"__entity":{"type":"User","id":"bob"}}]`)

		var got types.EntityUIDSet
		testutil.OK(t, json.Unmarshal(b, &got))
		testutil.Equals(t, got.Equal(want), true)

		// Ensure v1's lenient options don't reach EntityUIDSet's own decoding.
		testutil.Error(t, json.Unmarshal([]byte(`[{"type":"User","id":"a","id":"b"}]`), &got))
	})
}
