package mapset

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func hashSetMustNotContain[T comparable](t *testing.T, s *MapSet[T], item T) {
	testutil.FatalIf(t, s.Contains(item), "set %v unexpectedly contained item %v", s, 1)
}

func TestMapSet(t *testing.T) {
	t.Run("empty set contains nothing", func(t *testing.T) {
		s := Make[int]()
		testutil.Equals(t, s.Len(), 0)
		hashSetMustNotContain(t, s, 1)

		s = Make[int](10)
		testutil.Equals(t, s.Len(), 0)
		hashSetMustNotContain(t, s, 1)
	})

	t.Run("add => contains", func(t *testing.T) {
		s := Make[int]()
		s.Add(1)
		testutil.Equals(t, s.Contains(1), true)
	})

	t.Run("add twice", func(t *testing.T) {
		s := Make[int]()
		testutil.Equals(t, s.Add(1), true)
		testutil.Equals(t, s.Add(1), false)
	})

	t.Run("remove nonexistent", func(t *testing.T) {
		s := Make[int]()
		testutil.Equals(t, s.Remove(1), false)
	})

	t.Run("remove existing", func(t *testing.T) {
		s := Make[int]()
		s.Add(1)
		testutil.Equals(t, s.Remove(1), true)
	})

	t.Run("remove => !contains", func(t *testing.T) {
		s := Make[int]()
		s.Add(1)
		s.Remove(1)
		testutil.FatalIf(t, s.Contains(1), "set unexpectedly contained item")
	})

	t.Run("new from slice", func(t *testing.T) {
		s := FromItems(1, 2, 2, 3)
		testutil.Equals(t, s.Len(), 3)
		testutil.Equals(t, s.Contains(1), true)
		testutil.Equals(t, s.Contains(2), true)
		testutil.Equals(t, s.Contains(3), true)
	})

	t.Run("slice", func(t *testing.T) {
		s := Make[int]()
		testutil.Equals(t, s.Slice(), []int(nil))

		s = Make[int](10)
		testutil.Equals(t, s.Slice(), []int(nil))

		inSlice := []int{1, 2, 3}
		s = FromItems(inSlice...)
		outSlice := s.Slice()
		slices.Sort(outSlice)
		testutil.Equals(t, inSlice, outSlice)
	})

	t.Run("equal", func(t *testing.T) {
		s1 := FromItems(1, 2, 3)
		testutil.Equals(t, s1.Equal(s1), true)

		s2 := FromItems(1, 2, 3)
		testutil.Equals(t, s1.Equal(s2), true)

		s2.Add(4)
		testutil.Equals(t, s1.Equal(s2), false)

		s2.Remove(3)
		testutil.Equals(t, s1.Equal(s2), false)

		s1.Add(4)
		s1.Remove(3)
		testutil.Equals(t, s1.Equal(s2), true)
		testutil.Equals(t, s2.Equal(s1), true)
	})

	t.Run("iterate", func(t *testing.T) {
		s1 := FromItems(1, 2, 3)

		s2 := Make[int]()
		s1.Iterate(func(item int) bool {
			s2.Add(item)
			return true
		})

		testutil.Equals(t, s1.Equal(s2), true)
	})

	t.Run("iterate break early", func(t *testing.T) {
		s1 := FromItems(1, 2, 3)

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
		s1 := FromItems(1, 2, 3)

		s2 := Make[int]()
		for item := range s1.All() {
			s2.Add(item)
		}

		testutil.Equals(t, s1.Equal(s2), true)
	})

	t.Run("all break early", func(t *testing.T) {
		s1 := FromItems(1, 2, 3)

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
		s1 := FromItems(1, 2, 3)
		s2 := FromItems(2, 3, 4)

		testutil.Equals(t, s1.Intersects(s2), true)
	})

	t.Run("intersection disjoint", func(t *testing.T) {
		s1 := FromItems(1, 2)
		s2 := FromItems(3, 4)

		testutil.Equals(t, s1.Intersects(s2), false)
	})

	t.Run("encode nil set", func(t *testing.T) {
		s := Make[int]()

		out, err := json.Marshal(s)

		testutil.OK(t, err)
		testutil.Equals(t, string(out), "[]")
	})

	t.Run("marshal error", func(t *testing.T) {
		s := FromItems(complex(0, 0))
		_, err := json.Marshal(s)
		testutil.Error(t, err)
	})

	t.Run("encode json one int", func(t *testing.T) {
		s := FromItems(1)

		out, err := json.Marshal(s)

		testutil.OK(t, err)
		testutil.Equals(t, string(out), "[1]")
	})

	t.Run("encode json multiple int", func(t *testing.T) {
		s := FromItems(3, 2, 1)

		out, err := json.Marshal(s)

		testutil.OK(t, err)
		testutil.Equals(t, string(out), "[1,2,3]")
	})

	t.Run("encode json multiple string", func(t *testing.T) {
		s := FromItems("1", "2", "3")

		out, err := json.Marshal(s)

		testutil.OK(t, err)
		testutil.Equals(t, string(out), `["1","2","3"]`)
	})

	t.Run("decode json", func(t *testing.T) {
		var s1 MapSet[int]
		err := s1.UnmarshalJSON([]byte("[2,3,1,2]"))
		testutil.OK(t, err)
		testutil.Equals(t, &s1, FromItems(1, 2, 3))
	})

	t.Run("decode json empty", func(t *testing.T) {
		var s1 MapSet[int]
		err := s1.UnmarshalJSON([]byte("[]"))
		testutil.OK(t, err)
		testutil.Equals(t, s1.Len(), 0)
	})

	t.Run("decode mixed types in array", func(t *testing.T) {
		var s1 MapSet[int]
		err := s1.UnmarshalJSON([]byte(`[2,3,1,"2"]`))
		testutil.Error(t, err)
		testutil.Equals(t, err.Error(), "json: cannot unmarshal string into Go value of type int")
		testutil.Equals(t, s1.Len(), 0)
	})

	t.Run("decode wrong type", func(t *testing.T) {
		var s1 MapSet[int]
		err := s1.UnmarshalJSON([]byte(`"1,2,3"`))
		testutil.Error(t, err)
		testutil.Equals(t, err.Error(), "json: cannot unmarshal string into Go value of type []int")
		testutil.Equals(t, s1.Len(), 0)
	})

	t.Run("panic if too many args", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("code did not panic as expected")
			}
		}()

		Make[int](0, 1)
	})

	// The zero value MapSet is usable, but care must be taken to ensure that it is not mutated when passed by value
	// because those mutations may or may not be reflected in the caller's version of the MapSet.
	t.Run("zero value", func(t *testing.T) {
		s := MapSet[int]{}
		hashSetMustNotContain(t, &s, 0)
		testutil.Equals(t, s.Slice(), nil)

		addByValue := func(m MapSet[int], val int) {
			m.Add(val)
		}

		// Calling addByValue when s is still the zero value results in no mutation
		addByValue(s, 1)
		testutil.Equals(t, s.Len(), 0)

		// However, calling addByValue after the internal map in s has been initialized results in mutation
		s.Add(0)
		testutil.Equals(t, s.Len(), 1)
		addByValue(s, 1)
		testutil.Equals(t, s.Len(), 2)
	})

}
