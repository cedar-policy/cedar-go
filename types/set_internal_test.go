package types

import (
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

type colliderValue struct {
	Value   Value
	HashVal uint64
}

func (c colliderValue) String() string       { return "" }
func (c colliderValue) MarshalCedar() []byte { return nil }
func (c colliderValue) Equal(v Value) bool   { return v.Equal(c.Value) }
func (c colliderValue) hash() uint64         { return c.HashVal }

func TestSet(t *testing.T) {
	t.Parallel()

	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		t.Run("order independent", func(t *testing.T) {
			t.Parallel()
			s1 := NewSet([]Value{Long(42), Long(1337)})
			s2 := NewSet([]Value{Long(1337), Long(42)})
			testutil.Equals(t, s1.hash(), s2.hash())
		})

		t.Run("order independent with collisions", func(t *testing.T) {
			t.Parallel()

			v1 := colliderValue{Value: String("foo"), HashVal: 1337}
			v2 := colliderValue{Value: String("bar"), HashVal: 1337}
			v3 := colliderValue{Value: String("baz"), HashVal: 1337}

			permutations := []Set{
				NewSet([]Value{v1, v2, v3}),
				NewSet([]Value{v1, v3, v2}),
				NewSet([]Value{v2, v1, v3}),
				NewSet([]Value{v2, v3, v1}),
				NewSet([]Value{v3, v1, v2}),
				NewSet([]Value{v3, v2, v1}),
			}
			expected := permutations[0].hash()
			for _, p := range permutations {
				testutil.Equals(t, p.hash(), expected)
			}
		})

		t.Run("order independent with interleaving collisions", func(t *testing.T) {
			t.Parallel()

			v1 := colliderValue{Value: String("foo"), HashVal: 1337}
			v2 := colliderValue{Value: String("bar"), HashVal: 1338}
			v3 := colliderValue{Value: String("baz"), HashVal: 1337}

			permutations := []Set{
				NewSet([]Value{v1, v2, v3}),
				NewSet([]Value{v1, v3, v2}),
				NewSet([]Value{v2, v1, v3}),
				NewSet([]Value{v2, v3, v1}),
				NewSet([]Value{v3, v1, v2}),
				NewSet([]Value{v3, v2, v1}),
			}
			expected := permutations[0].hash()
			for _, p := range permutations {
				testutil.Equals(t, p.hash(), expected)
			}
		})

		t.Run("duplicates unimportant", func(t *testing.T) {
			t.Parallel()
			s1 := NewSet([]Value{Long(42), Long(1337)})
			s2 := NewSet([]Value{Long(42), Long(1337), Long(1337)})
			testutil.Equals(t, s1.hash(), s2.hash())
		})

		t.Run("empty set", func(t *testing.T) {
			t.Parallel()
			m1 := Set{}
			m2 := NewSet([]Value{})
			m3 := NewSet(nil)
			testutil.Equals(t, m1.hash(), m2.hash())
			testutil.Equals(t, m2.hash(), m3.hash())
		})

		// These tests don't necessarily hold for all values of Set, but we want to ensure we are considering
		// different aspects of the Set, which these particular tests demonstrate.

		t.Run("extra element", func(t *testing.T) {
			t.Parallel()
			s1 := NewSet([]Value{Long(42), Long(1337)})
			s2 := NewSet([]Value{Long(42), Long(1337), Long(1)})
			testutil.FatalIf(t, s1.hash() == s2.hash(), "unexpected hash collision")
		})

		t.Run("disjoint", func(t *testing.T) {
			t.Parallel()
			s1 := NewSet([]Value{Long(42), Long(1337)})
			s2 := NewSet([]Value{Long(0), String("hi")})
			testutil.FatalIf(t, s1.hash() == s2.hash(), "unexpected hash collision")
		})
	})

	t.Run("collisions", func(t *testing.T) {
		t.Parallel()

		v1 := colliderValue{Value: String("foo"), HashVal: 1337}
		v2 := colliderValue{Value: String("bar"), HashVal: 1337}
		v3 := colliderValue{Value: String("baz"), HashVal: 1338}
		v4 := colliderValue{Value: String("baz"), HashVal: 1337}

		set := NewSet([]Value{v1, v2, v3, v4})

		testutil.Equals(t, set.Len(), 3)

		var vals []Value
		set.Iterate(func(v Value) bool {
			vals = append(vals, v)
			return true
		})

		testutil.Equals(t, slices.ContainsFunc(vals, func(v Value) bool { return v.Equal(v1) }), true)
		testutil.Equals(t, slices.ContainsFunc(vals, func(v Value) bool { return v.Equal(v2) }), true)
		testutil.Equals(t, slices.ContainsFunc(vals, func(v Value) bool { return v.Equal(v3) }), true)
		testutil.Equals(t, slices.ContainsFunc(vals, func(v Value) bool { return v.Equal(v4) }), true)
	})
}
