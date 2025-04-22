package types_test

import (
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestSet(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		empty := types.Set{}
		empty2 := types.Set{}
		empty3 := types.NewSet()
		oneTrue := types.NewSet(types.Boolean(true))
		oneTrue2 := types.NewSet(types.Boolean(true))
		oneFalse := types.NewSet(types.Boolean(false))
		nestedOnce := types.NewSet(empty, oneTrue, oneFalse)
		nestedOnce2 := types.NewSet(empty, oneTrue, oneFalse)
		nestedTwice := types.NewSet(empty, oneTrue, oneFalse, nestedOnce)
		nestedTwice2 := types.NewSet(empty, oneTrue, oneFalse, nestedOnce)
		oneTwoThree := types.NewSet(
			types.Long(1), types.Long(2), types.Long(3),
		)
		threeTwoTwoOne := types.NewSet(
			types.Long(3), types.Long(2), types.Long(2), types.Long(1),
		)

		testutil.FatalIf(t, !empty.Equal(empty), "%v not Equal to %v", empty, empty)
		testutil.FatalIf(t, !empty.Equal(empty2), "%v not Equal to %v", empty, empty2)
		testutil.FatalIf(t, !empty.Equal(empty3), "%v not Equal to %v", empty, empty3)
		testutil.FatalIf(t, !oneTrue.Equal(oneTrue), "%v not Equal to %v", oneTrue, oneTrue)
		testutil.FatalIf(t, !oneTrue.Equal(oneTrue2), "%v not Equal to %v", oneTrue, oneTrue2)
		testutil.FatalIf(t, !nestedOnce.Equal(nestedOnce), "%v not Equal to %v", nestedOnce, nestedOnce)
		testutil.FatalIf(t, !nestedOnce.Equal(nestedOnce2), "%v not Equal to %v", nestedOnce, nestedOnce2)
		testutil.FatalIf(t, !nestedTwice.Equal(nestedTwice), "%v not Equal to %v", nestedTwice, nestedTwice)
		testutil.FatalIf(t, !nestedTwice.Equal(nestedTwice2), "%v not Equal to %v", nestedTwice, nestedTwice2)
		testutil.FatalIf(t, !oneTwoThree.Equal(threeTwoTwoOne), "%v not Equal to %v", oneTwoThree, threeTwoTwoOne)

		testutil.FatalIf(t, empty.Equal(oneFalse), "%v Equal to %v", empty, oneFalse)
		testutil.FatalIf(t, oneTrue.Equal(oneFalse), "%v Equal to %v", oneTrue, oneFalse)
		testutil.FatalIf(t, nestedOnce.Equal(nestedTwice), "%v Equal to %v", nestedOnce, nestedTwice)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.Set{}.String(), "[]")
		testutil.Equals(
			t,
			types.NewSet(types.Boolean(true), types.Long(1)).String(),
			"[true, 1]")
	})

	t.Run("Len", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.Set{}.Len(), 0)
		testutil.Equals(t, types.Set{}.Len(), 0)
		testutil.Equals(t, types.NewSet(types.Long(1)).Len(), 1)
		testutil.Equals(t, types.NewSet(types.Long(1), types.Long(2)).Len(), 2)
	})

	t.Run("IterateEntire", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			values []types.Value
		}{
			{name: "empty set", values: nil},
			{name: "one item", values: []types.Value{types.Long(42)}},
			{name: "two items", values: []types.Value{types.Long(42), types.Long(1337)}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				set := types.NewSet(tt.values...)

				var got []types.Value
				set.Iterate(func(v types.Value) bool {
					got = append(got, v)
					return true
				})

				testutil.Equals(t, len(got), len(tt.values))
				for _, v := range got {
					testutil.FatalIf(
						t,
						!slices.ContainsFunc(tt.values, func(vv types.Value) bool { return vv.Equal(v) }),
						"value %v not in input slice %v", v, tt.values,
					)
				}
			})
		}
	})

	t.Run("IteratePartial", func(t *testing.T) {
		t.Parallel()

		set := types.NewSet(types.Long(42), types.Long(1337))

		// It would be nice to know which element or elements were returned when iteration ends early, but iteration
		// order for Sets is non-deterministic
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
				var got []types.Value
				var i int
				set.Iterate(func(v types.Value) bool {
					if i == tt.breakOn {
						return false
					}
					i++
					got = append(got, v)
					return true
				})

				testutil.Equals(t, len(got), tt.breakOn)
				for _, g := range got {
					testutil.FatalIf(t, !set.Contains(g), "got value %v which is not in set %v", g, set)
				}
			})
		}
	})

	t.Run("AllEntire", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			values []types.Value
		}{
			{name: "empty set", values: nil},
			{name: "one item", values: []types.Value{types.Long(42)}},
			{name: "two items", values: []types.Value{types.Long(42), types.Long(1337)}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				set := types.NewSet(tt.values...)

				var got []types.Value
				for v := range set.All() {
					got = append(got, v)
				}

				testutil.Equals(t, len(got), len(tt.values))
				for _, v := range got {
					testutil.FatalIf(
						t,
						!slices.ContainsFunc(tt.values, func(vv types.Value) bool { return vv.Equal(v) }),
						"value %v not in input slice %v", v, tt.values,
					)
				}
			})
		}
	})

	t.Run("IteratePartial", func(t *testing.T) {
		t.Parallel()

		set := types.NewSet(types.Long(42), types.Long(1337))

		// It would be nice to know which element or elements were returned when iteration ends early, but iteration
		// order for Sets is non-deterministic
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
				var got []types.Value
				var i int
				for v := range set.All() {
					if i == tt.breakOn {
						break
					}
					i++
					got = append(got, v)
				}

				testutil.Equals(t, len(got), tt.breakOn)
				for _, g := range got {
					testutil.FatalIf(t, !set.Contains(g), "got value %v which is not in set %v", g, set)
				}
			})
		}
	})

	t.Run("Slice", func(t *testing.T) {
		t.Parallel()

		s := types.Set{}.Slice()
		testutil.Equals(t, s, nil)

		s = types.NewSet().Slice()
		testutil.Equals(t, s, nil)

		s = types.NewSet(types.True).Slice()
		testutil.Equals(t, s, []types.Value{types.True})

		s = types.NewSet(types.True, types.False).Slice()
		testutil.Equals(t, len(s), 2)
		testutil.FatalIf(t, !slices.ContainsFunc(s, func(v types.Value) bool { return v.Equal(types.True) }), "")
		testutil.FatalIf(t, !slices.ContainsFunc(s, func(v types.Value) bool { return v.Equal(types.False) }), "")

		s = types.NewSet(types.True, types.False, types.True).Slice()
		testutil.Equals(t, len(s), 2)
		testutil.FatalIf(t, !slices.ContainsFunc(s, func(v types.Value) bool { return v.Equal(types.True) }), "")
		testutil.FatalIf(t, !slices.ContainsFunc(s, func(v types.Value) bool { return v.Equal(types.False) }), "")

		// Show that mutating the returned slice doesn't affect Set's internal state
		r := types.NewSet(types.True, types.False)
		s = r.Slice()
		_ = append(s, types.Long(0))
		testutil.Equals(t, r, types.NewSet(types.True, types.False))
	})

	// This test is intended to show the NewSet makes a copy of the Values in the input slice
	t.Run("immutable", func(t *testing.T) {
		t.Parallel()

		slice := []types.Value{types.Long(42)}
		p := &slice[0]

		set := types.NewSet(slice...)

		*p = types.Long(1337)

		testutil.Equals(t, set.Len(), 1)

		var got types.Long
		set.Iterate(func(v types.Value) bool {
			var ok bool
			got, ok = v.(types.Long)
			testutil.FatalIf(t, !ok, "incorrect type for set element")
			return true
		})

		testutil.Equals(t, got, types.Long(42))
	})

	t.Run("no duplicates", func(t *testing.T) {
		t.Parallel()

		set := types.NewSet(types.Long(42), types.Long(42))

		testutil.Equals(t, set.Len(), 1)

		var got types.Long
		set.Iterate(func(v types.Value) bool {
			var ok bool
			got, ok = v.(types.Long)
			testutil.FatalIf(t, !ok, "incorrect type for set element")
			return true
		})

		testutil.Equals(t, got, types.Long(42))
	})

	t.Run("Contains", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name  string
			set   types.Set
			value types.Value
			want  bool
		}{
			{"trueLong", types.NewSet(types.Long(42)), types.Long(42), true},
			{"falseLong", types.NewSet(types.Long(42)), types.Long(1234), false},
			{"trueDecimal",
				types.NewSet(testutil.Must(types.NewDecimalFromInt(42))),
				testutil.Must(types.NewDecimalFromInt(42)),
				true,
			},
			{"falseDecimal",
				types.NewSet(testutil.Must(types.NewDecimalFromInt(42))),
				testutil.Must(types.NewDecimalFromInt(1234)),
				false,
			},
			{"trueDuration",
				types.NewSet(types.NewDurationFromMillis(42)),
				types.NewDurationFromMillis(42),
				true,
			},
			{"falseDuration",
				types.NewSet(types.NewDurationFromMillis(42)),
				types.NewDurationFromMillis(1234),
				false,
			},
			{"trueDatetime",
				types.NewSet(types.NewDatetimeFromMillis(42)),
				types.NewDatetimeFromMillis(42),
				true,
			},
			{"falseDatetime",
				types.NewSet(types.NewDatetimeFromMillis(42)),
				types.NewDatetimeFromMillis(1234),
				false,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				got := tt.set.Contains(tt.value)
				testutil.Equals(t, got, tt.want)
			})
		}
	})

	t.Run("Equals", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name  string
			set   types.Set
			value types.Value
			want  bool
		}{
			{"true", types.NewSet(types.Long(42)), types.NewSet(types.Long(42)), true},
			{"falseSet", types.NewSet(types.Long(42)), types.NewSet(types.Long(1234)), false},
			{"falseOtherType", types.NewSet(types.Long(42)), types.Long(24), false},
			{"falseSameHash", types.NewSet(types.Long(0)), types.NewSet(testutil.Must(types.NewDecimalFromInt(0))), false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				got := tt.set.Equal(tt.value)
				testutil.Equals(t, got, tt.want)
			})
		}
	})
}
