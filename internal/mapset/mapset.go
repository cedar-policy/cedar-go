package mapset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"slices"
)

// Similar to the concept of a [legal peppercorn](https://en.wikipedia.org/wiki/Peppercorn_(law)), this instance of
// nothingness is required in order to transact with Go's map[T]struct{} idiom.
var peppercorn = struct{}{}

// MapSet is a struct that adds some convenience to the otherwise cumbersome map[T]struct{} idiom used in Go to
// implement mapset of comparable types.
//
// Note: the zero value of MapSet[T] (i.e. MapSet[T]{}) is fully usable and avoids unnecessary allocations in the case
// where nothing gets added to the MapSet. However, take care in using it, especially when passing it by value to other
// functions. If passed by value, mutating operations (e.g. Add(), Remove()) in the called function will persist in the
// calling function's version if the MapSet[T] has been changed from the zero value prior to the call.
// See the "zero value" test for an example.
type MapSet[T comparable] struct {
	m map[T]struct{}
}

// Make returns a MapSet ready for use. Optionally, a desired size for the MapSet can be passed as an argument,
// as in the argument to make() for a map type.
func Make[T comparable](args ...int) *MapSet[T] {
	if len(args) > 1 {
		panic(fmt.Sprintf("too many arguments passed to Make(). got: %v, expected 0 or 1", len(args)))
	}

	var size int
	if len(args) == 1 {
		size = args[0]
	}

	return &MapSet[T]{m: make(map[T]struct{}, size)}
}

// FromItems creates a MapSet of size len(items) and calls Add for each of the items to it.
func FromItems[T comparable](items ...T) *MapSet[T] {
	h := Make[T](len(items))
	for _, i := range items {
		h.Add(i)
	}
	return h
}

// Add an item to the set. Returns true if the item did not exist in the set.
func (h *MapSet[T]) Add(item T) bool {
	if h.m == nil {
		h.m = map[T]struct{}{}
	}

	if _, exists := h.m[item]; exists {
		return false
	}
	h.m[item] = peppercorn
	return true
}

// Remove an item from the Set. Returns true if the item existed in the set.
func (h *MapSet[T]) Remove(item T) bool {
	_, exists := h.m[item]
	delete(h.m, item)
	return exists
}

// Contains returns whether the item exists in the set
func (h MapSet[T]) Contains(item T) bool {
	_, exists := h.m[item]
	return exists
}

type Container[T comparable] interface {
	Contains(T) bool
	Len() int
}

// Intersects returns whether any items in this set exist in o
func (h MapSet[T]) Intersects(o Container[T]) bool {
	for item := range h.m {
		if o.Contains(item) {
			return true
		}
	}
	return false
}

// Iterate the items in the set, calling callback for each item. If the callback returns false, iteration is halted.
// Iteration order is undefined.
//
// Deprecated: Use All() instead.
func (h MapSet[T]) Iterate(callback func(item T) bool) {
	for item := range h.m {
		if !callback(item) {
			break
		}
	}
}

// All returns an iterator over elements in the set. Iteration order is undefined.
func (h MapSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range h.m {
			if !yield(item) {
				return
			}
		}
	}
}

func (h MapSet[T]) Slice() []T {
	if h.m == nil {
		return nil
	}
	return slices.Collect(maps.Keys(h.m))
}

// Len returns the size of the MapSet
func (h MapSet[T]) Len() int {
	return len(h.m)
}

// Equal returns whether the same items exist in both h and o
func (h MapSet[T]) Equal(o Container[T]) bool {
	if len(h.m) != o.Len() {
		return false
	}

	for item := range h.m {
		if !o.Contains(item) {
			return false
		}
	}
	return true
}

// MarshalJSON serializes a MapSet as a JSON array. Elements are ordered lexicographically by their marshaled value.
func (h MapSet[T]) MarshalJSON() ([]byte, error) {
	if h.m == nil {
		return []byte("[]"), nil
	}

	elems := h.Slice()
	marshaledElems := make([][]byte, 0, len(elems))
	for _, elem := range elems {
		b, err := json.Marshal(elem)
		if err != nil {
			return nil, err
		}
		marshaledElems = append(marshaledElems, b)
	}
	slices.SortFunc(marshaledElems, func(a, b []byte) int { return slices.Compare(a, b) })
	return slices.Concat([]byte{'['}, bytes.Join(marshaledElems, []byte{','}), []byte{']'}), nil
}

// UnmarshalJSON deserializes a MapSet from a JSON array.
func (h *MapSet[T]) UnmarshalJSON(b []byte) error {
	var s []T
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*h = *FromItems(s...)
	return nil
}
