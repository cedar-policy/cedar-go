package mapset

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func immutableMapSetMustNotContain[T comparable](t *testing.T, s ImmutableMapSet[T], item T) {
	testutil.FatalIf(t, s.Contains(item), "set %v unexpectedly contained item %v", s, 1)
}

func TestImmutableMapSet(t *testing.T) {
	t.Run("empty set contains nothing", func(t *testing.T) {
		s := Immutable[int]()
		testutil.Equals(t, s.Len(), 0)
		immutableMapSetMustNotContain(t, s, 1)
	})

	t.Run("one element", func(t *testing.T) {
		s := Immutable[int](1)
		testutil.Equals(t, s.Contains(1), true)
	})

	t.Run("two elements", func(t *testing.T) {
		s := Immutable[int](1, 2)
		testutil.Equals(t, s.Contains(1), true)
		testutil.Equals(t, s.Contains(2), true)
		testutil.Equals(t, s.Contains(3), false)
	})

	t.Run("deduplicate elements", func(t *testing.T) {
		s := Immutable[int](1, 1)
		testutil.Equals(t, s.Contains(1), true)
		testutil.Equals(t, s.Len(), 1)
	})

	t.Run("slice", func(t *testing.T) {
		s := Immutable[int]()
		testutil.Equals(t, s.Slice(), []int(nil))

		inSlice := []int{1, 2, 3}
		s = Immutable[int](inSlice...)

		outSlice := s.Slice()
		slices.Sort(outSlice)
		testutil.Equals(t, inSlice, outSlice)
	})

	t.Run("equal", func(t *testing.T) {
		s1 := Immutable(1, 2, 3)
		testutil.Equals(t, s1.Equal(s1), true)

		s2 := Immutable(1, 2, 3)
		testutil.Equals(t, s1.Equal(s2), true)

		s3 := Immutable(1, 2, 3, 4)
		testutil.Equals(t, s1.Equal(s3), false)
	})

	t.Run("equality of MapSet and ImmutableMapSet", func(t *testing.T) {
		s1 := FromItems(1, 2, 3, 4)
		s2 := Immutable(1, 2, 3, 4)
		s3 := Immutable(1, 2, 3)
		testutil.Equals(t, s1.Equal(s2), true)
		testutil.Equals(t, s2.Equal(s1), true)
		testutil.Equals(t, s1.Equal(s3), false)
		testutil.Equals(t, s3.Equal(s1), false)
	})

	t.Run("iterate", func(t *testing.T) {
		s1 := Immutable(1, 2, 3)

		s2 := Make[int]()
		s1.Iterate(func(item int) bool {
			s2.Add(item)
			return true
		})

		testutil.Equals(t, s1.Equal(s2), true)
	})

	t.Run("iterate break early", func(t *testing.T) {
		s1 := Immutable(1, 2, 3)

		var items []int
		s1.Iterate(func(item int) bool {
			if len(items) == 2 {
				return false
			}
			items = append(items, item)
			return true
		})

		// Because iteration order is non-deterministic, all we can say is that the right number of items ended up in
		// the set and that the items were in the original set.
		testutil.Equals(t, len(items), 2)
		testutil.Equals(t, s1.Contains(items[0]), true)
		testutil.Equals(t, s1.Contains(items[1]), true)
	})

	t.Run("all", func(t *testing.T) {
		s1 := Immutable(1, 2, 3)

		s2 := Make[int]()
		for item := range s1.All() {
			s2.Add(item)
		}

		testutil.Equals(t, s1.Equal(s2), true)
	})

	t.Run("all break early", func(t *testing.T) {
		s1 := Immutable(1, 2, 3)

		var items []int
		for item := range s1.All() {
			if len(items) == 2 {
				break
			}
			items = append(items, item)
		}

		// Because iteration order is non-deterministic, all we can say is that the right number of items ended up in
		// the set and that the items were in the original set.
		testutil.Equals(t, len(items), 2)
		testutil.Equals(t, s1.Contains(items[0]), true)
		testutil.Equals(t, s1.Contains(items[1]), true)
	})

	t.Run("intersection with overlap", func(t *testing.T) {
		s1 := Immutable(1, 2, 3)
		s2 := Immutable(2, 3, 4)

		testutil.Equals(t, s1.Intersects(s2), true)
	})

	t.Run("intersection disjoint", func(t *testing.T) {
		s1 := Immutable(1, 2)
		s2 := Immutable(3, 4)

		testutil.Equals(t, s1.Intersects(s2), false)
	})

	t.Run("encode nil set", func(t *testing.T) {
		s := ImmutableMapSet[int]{}

		out, err := json.Marshal(s)

		testutil.OK(t, err)
		testutil.Equals(t, string(out), "[]")
	})

	t.Run("encode json", func(t *testing.T) {
		s := Immutable(1, 2, 3)

		out, err := json.Marshal(s)

		testutil.OK(t, err)
		testutil.Equals(t, string(out), "[1,2,3]")
	})

	t.Run("decode json", func(t *testing.T) {
		var s1 ImmutableMapSet[int]
		err := s1.UnmarshalJSON([]byte("[2,3,1,2]"))
		testutil.OK(t, err)
		testutil.Equals(t, s1, Immutable(1, 2, 3))
	})

	t.Run("decode json empty", func(t *testing.T) {
		var s1 ImmutableMapSet[int]
		err := s1.UnmarshalJSON([]byte("[]"))
		testutil.OK(t, err)
		testutil.Equals(t, s1.Len(), 0)
	})

	t.Run("decode mixed types in array", func(t *testing.T) {
		var s1 ImmutableMapSet[int]
		err := s1.UnmarshalJSON([]byte(`[2,3,1,"2"]`))
		testutil.Error(t, err)
		testutil.Equals(t, err.Error(), "json: cannot unmarshal string into Go value of type int")
		testutil.Equals(t, s1.Len(), 0)
	})

	t.Run("decode wrong type", func(t *testing.T) {
		var s1 ImmutableMapSet[int]
		err := s1.UnmarshalJSON([]byte(`"1,2,3"`))
		testutil.Error(t, err)
		testutil.Equals(t, err.Error(), "json: cannot unmarshal string into Go value of type []int")
		testutil.Equals(t, s1.Len(), 0)
	})
}
