package types_test

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/mapset"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestRecord(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		empty := types.Record{}
		empty2 := types.NewRecord(nil)
		empty3 := types.NewRecord(types.RecordMap{})
		twoElems := types.NewRecord(types.RecordMap{
			"foo": types.Boolean(true),
			"bar": types.String("blah"),
		})
		twoElems2 := types.NewRecord(types.RecordMap{
			"foo": types.Boolean(true),
			"bar": types.String("blah"),
		})
		differentValues := types.NewRecord(types.RecordMap{
			"foo": types.Boolean(false),
			"bar": types.String("blaz"),
		})
		differentKeys := types.NewRecord(types.RecordMap{
			"foo": types.Boolean(false),
			"bar": types.Long(1),
		})
		nested := types.NewRecord(types.RecordMap{
			"one":  types.Long(1),
			"two":  types.Long(2),
			"nest": twoElems,
		})
		nested2 := types.NewRecord(types.RecordMap{
			"one":  types.Long(1),
			"two":  types.Long(2),
			"nest": twoElems,
		})
		sameHash1 := types.NewRecord(types.RecordMap{
			"key": types.Long(0),
		})
		sameHash2 := types.NewRecord(types.RecordMap{
			"key": testutil.Must(types.NewDecimalFromInt(0)),
		})

		testutil.FatalIf(t, !empty.Equal(empty), "%v not Equal to %v", empty, empty)
		testutil.FatalIf(t, !empty.Equal(empty2), "%v not Equal to %v", empty, empty2)
		testutil.FatalIf(t, !empty.Equal(empty3), "%v not Equal to %v", empty, empty3)

		testutil.FatalIf(t, !twoElems.Equal(twoElems), "%v not Equal to %v", twoElems, twoElems)
		testutil.FatalIf(t, !twoElems.Equal(twoElems2), "%v not Equal to %v", twoElems, twoElems2)

		testutil.FatalIf(t, !nested.Equal(nested), "%v not Equal to %v", nested, nested)
		testutil.FatalIf(t, !nested.Equal(nested2), "%v not Equal to %v", nested, nested2)

		testutil.FatalIf(t, nested.Equal(twoElems), "%v Equal to %v", nested, twoElems)
		testutil.FatalIf(t, twoElems.Equal(differentValues), "%v Equal to %v", twoElems, differentValues)
		testutil.FatalIf(t, twoElems.Equal(differentKeys), "%v Equal to %v", twoElems, differentKeys)

		testutil.FatalIf(t, sameHash1.Equal(sameHash2), "%v Equal to %v", sameHash1, sameHash2)

	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.Record{}.String(), "{}")
		testutil.Equals(t, types.NewRecord(nil).String(), "{}")
		testutil.Equals(t, types.NewRecord(types.RecordMap{}).String(), "{}")
		testutil.Equals(
			t,
			types.NewRecord(types.RecordMap{"foo": types.Boolean(true)}).String(),
			`{"foo":true}`)
		testutil.Equals(
			t,
			types.NewRecord(types.RecordMap{
				"foo": types.Boolean(true),
				"bar": types.String("blah"),
			}).String(),
			`{"bar":"blah", "foo":true}`)
		testutil.Equals(
			t,
			types.NewRecord(types.RecordMap{
				"\x0b": types.Boolean(true),
			}).String(),
			`{"\u{b}":true}`)
	})

	t.Run("Len", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.Record{}.Len(), 0)
		testutil.Equals(t, types.NewRecord(nil).Len(), 0)
		testutil.Equals(t, types.NewRecord(types.RecordMap{}).Len(), 0)
		testutil.Equals(t, types.NewRecord(types.RecordMap{"foo": types.Long(1)}).Len(), 1)
		testutil.Equals(t, types.NewRecord(types.RecordMap{"foo": types.Long(1), "bar": types.Long(2)}).Len(), 2)
	})

	t.Run("IterateEntire", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			values types.RecordMap
		}{
			{name: "empty map", values: types.RecordMap{}},
			{name: "one item", values: types.RecordMap{"foo": types.Long(42)}},
			{name: "two items", values: types.RecordMap{"foo": types.Long(42), "bar": types.Long(1337)}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				record := types.NewRecord(tt.values)

				got := types.RecordMap{}
				record.Iterate(func(k types.String, v types.Value) bool {
					got[k] = v
					return true
				})
				testutil.Equals(t, got, tt.values)
			})
		}
	})

	t.Run("IteratePartial", func(t *testing.T) {
		t.Parallel()

		record := types.NewRecord(types.RecordMap{"foo": types.Long(42), "bar": types.Long(1337)})

		// It would be nice to know which element or elements were returned when iteration ends early, but iteration
		// order for Records is non-deterministic
		tests := []struct {
			name    string
			breakOn int
		}{
			{name: "empty record", breakOn: 0},
			{name: "one item", breakOn: 1},
			{name: "two items", breakOn: 2},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				got := types.RecordMap{}
				var i int
				record.Iterate(func(k types.String, v types.Value) bool {
					if i == tt.breakOn {
						return false
					}
					i++
					got[k] = v
					return true
				})

				testutil.Equals(t, len(got), tt.breakOn)
			})
		}
	})

	t.Run("IteratorsEntire", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			rm   types.RecordMap
		}{
			{name: "empty map", rm: types.RecordMap{}},
			{name: "one item", rm: types.RecordMap{"foo": types.Long(42)}},
			{name: "two items", rm: types.RecordMap{"foo": types.Long(42), "bar": types.Long(1337)}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				t.Run("All", func(t *testing.T) {
					t.Parallel()
					record := types.NewRecord(tt.rm)

					got := types.RecordMap{}
					for k, v := range record.All() {
						got[k] = v
					}
					testutil.Equals(t, got, tt.rm)
				})

				t.Run("Keys", func(t *testing.T) {
					t.Parallel()
					record := types.NewRecord(tt.rm)

					got := mapset.Make[types.String]()
					for k := range record.Keys() {
						got.Add(k)
					}
					testutil.Equals(t, got.Equal(mapset.Immutable(slices.Collect(maps.Keys(tt.rm))...)), true)

				})

				t.Run("Values", func(t *testing.T) {
					t.Parallel()
					record := types.NewRecord(tt.rm)

					got := mapset.Make[types.Value]()
					for v := range record.Values() {
						got.Add(v)
					}
					testutil.Equals(t, got.Equal(mapset.Immutable(slices.Collect(maps.Values(tt.rm))...)), true)
				})
			})
		}
	})

	t.Run("IteratorsPartial", func(t *testing.T) {
		t.Parallel()

		rm := types.RecordMap{"foo": types.Long(42), "bar": types.Long(1337)}
		record := types.NewRecord(rm)

		// It would be nice to know which element or elements were returned when iteration ends early, but iteration
		// order for Records is non-deterministic
		tests := []struct {
			name    string
			breakOn int
		}{
			{name: "empty record", breakOn: 0},
			{name: "one item", breakOn: 1},
			{name: "two items", breakOn: 2},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				t.Run("All", func(t *testing.T) {
					t.Parallel()

					got := types.RecordMap{}
					var i int
					for k, v := range record.All() {
						if i == tt.breakOn {
							break
						}
						i++
						got[k] = v
					}
					testutil.Equals(t, len(got), tt.breakOn)
				})

				t.Run("Keys", func(t *testing.T) {
					t.Parallel()

					got := mapset.Make[types.String]()
					var i int
					for k := range record.Keys() {
						if i == tt.breakOn {
							break
						}
						i++
						got.Add(k)
					}
					testutil.Equals(t, got.Len(), tt.breakOn)
					for k := range got.All() {
						_, ok := record.Get(k)
						testutil.Equals(t, ok, true)
					}
				})

				t.Run("Values", func(t *testing.T) {
					t.Parallel()

					got := mapset.Make[types.Value]()
					var i int
					for k := range record.Values() {
						if i == tt.breakOn {
							break
						}
						i++
						got.Add(k)
					}
					testutil.Equals(t, got.Len(), tt.breakOn)
					for v := range got.All() {
						testutil.Equals(t, slices.Contains(slices.Collect(maps.Values(rm)), v), true)
					}
				})
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		v, ok := types.Record{}.Get("foo")
		testutil.Equals(t, ok, false)
		testutil.Equals(t, v, nil)

		v, ok = types.NewRecord(types.RecordMap{}).Get("foo")
		testutil.Equals(t, ok, false)
		testutil.Equals(t, v, nil)

		r := types.NewRecord(types.RecordMap{"foo": types.Long(42), "bar": types.Long(1337)})

		v, ok = r.Get("foo")
		testutil.Equals(t, ok, true)
		testutil.Equals(t, v, types.Value(types.Long(42)))

		v, ok = r.Get("bar")
		testutil.Equals(t, ok, true)
		testutil.Equals(t, v, types.Value(types.Long(1337)))

		v, ok = r.Get("Bar")
		testutil.Equals(t, ok, false)
		testutil.Equals(t, v, nil)
	})

	t.Run("Map", func(t *testing.T) {
		t.Parallel()

		m := types.Record{}.Map()
		testutil.Equals(t, m, nil)

		m = types.NewRecord(nil).Map()
		testutil.Equals(t, m, nil)

		m = types.NewRecord(types.RecordMap{}).Map()
		testutil.Equals(t, m, types.RecordMap{})

		m = types.NewRecord(types.RecordMap{"foo": types.True}).Map()
		testutil.Equals(t, m, types.RecordMap{"foo": types.True})

		m = types.NewRecord(types.RecordMap{"foo": types.True, "bar": types.False}).Map()
		testutil.Equals(t, m, types.RecordMap{"foo": types.True, "bar": types.False})

		// Show that mutating the returned map doesn't affect Record's internal map
		r := types.NewRecord(types.RecordMap{"foo": types.True, "bar": types.False})
		m = r.Map()
		delete(m, "foo")
		m["bar"] = types.True
		testutil.Equals(t, r, types.NewRecord(types.RecordMap{"foo": types.True, "bar": types.False}))
	})

	t.Run("UnmarshalJSON_empty", func(t *testing.T) {
		t.Parallel()
		var r types.Record
		err := json.Unmarshal([]byte(`{}`), &r)
		testutil.OK(t, err)
		testutil.Equals(t, r, types.Record{})
	})

	// This test is intended to show that NewMap() makes a copy of the values from the input map
	t.Run("immutable", func(t *testing.T) {
		t.Parallel()

		m := types.RecordMap{"foo": types.Long(42), "bar": types.Long(1337)}
		r := types.NewRecord(m)

		delete(m, "foo")
		m["bar"] = types.Long(42)

		testutil.Equals(t, r.Len(), 2)

		got := types.RecordMap{}
		r.Iterate(func(k types.String, v types.Value) bool {
			got[k] = v
			return true
		})

		testutil.Equals(t, got, types.RecordMap{"foo": types.Long(42), "bar": types.Long(1337)})
	})

	// Add some rather paranoid testing to ensure that future Go versions don't
	// change the behavior of json.Deterministic(true) with respect to map
	// encoding, which ensures that cedar-go maintains parity with the Rust
	// implementation's encoding.
	t.Run("MarshalJSON", func(t *testing.T) {
		t.Run("sorted", func(t *testing.T) {
			t.Parallel()
			r := types.NewRecord(types.RecordMap{
				"b": types.Long(1), "a": types.Long(2), "B": types.Long(3), "A": types.Long(4),
				"aa": types.Long(5), "a\x00": types.Long(6), "": types.Long(7), "_": types.Long(8),
				// Characters that JSON escapes
				"\t": types.Long(9), "\x1f": types.Long(10), "\"": types.Long(11), "\\": types.Long(12),
				// The last printable ASCII character, and DEL, which JSON
				// doesn't escape
				"~": types.Long(13), "\x7f": types.Long(14),
				// é decomposed (e + U+0301 COMBINING ACUTE ACCENT) and
				// precomposed, which aren't normalized
				"e\u0301": types.Long(15), "\u00e9": types.Long(16),
				// 2-byte UTF-8 on either side of the Latin-1 boundary
				"\u00ff": types.Long(17), "\u0100": types.Long(18),
				// 3-byte UTF-8: LINE SEPARATOR (which JavaScript doesn't allow
				// in string literals), a CJK character, the first private-use
				// character, and the replacement character
				"\u2028": types.Long(19), "\u4e2d": types.Long(20), "\ue000": types.Long(21), "\ufffd": types.Long(22),
				// 4-byte UTF-8 (surrogate pairs in UTF-16), including the
				// highest code point. Bytewise, these sort after everything
				// above; in UTF-16 code unit order (e.g. RFC 8785) they'd sort
				// before U+E000.
				"\U0001f600": types.Long(23), "\U0010ffff": types.Long(24),
			})
			got, err := r.MarshalJSON()
			testutil.OK(t, err)
			want := `{"":7,"\t":9,"\u001f":10,"\"":11,"A":4,"B":3,"\\":12,"_":8,"a":2,"a\u0000":6,"aa":5,"b":1,` +
				"\"e\u0301\":15,\"~\":13,\"\x7f\":14,\"\u00e9\":16,\"\u00ff\":17,\"\u0100\":18," +
				"\"\u2028\":19,\"\u4e2d\":20,\"\ue000\":21,\"\ufffd\":22,\"\U0001f600\":23,\"\U0010ffff\":24}"
			testutil.Equals(t, string(got), want)
		})

		t.Run("nested sorted", func(t *testing.T) {
			t.Parallel()
			r := types.NewRecord(types.RecordMap{
				"y": types.NewRecord(types.RecordMap{"a": types.Long(0), "d": types.Long(3), "b": types.Long(1), "c": types.Long(2)}),
				"x": types.NewSet(types.NewRecord(types.RecordMap{"f": types.Long(5), "e": types.Long(4), "z": types.Long(6)})),
			})
			got, err := r.MarshalJSON()
			testutil.OK(t, err)
			testutil.Equals(t, string(got), `{"x":[{"e":4,"f":5,"z":6}],"y":{"a":0,"b":1,"c":2,"d":3}}`)
		})
	})
}
