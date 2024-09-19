<<<<<<<< HEAD:internal/hashset.go
package internal
========
package sets
>>>>>>>> 0f61d5f (fixup):internal/sets/mapset.go

import (
	"encoding/json"
	"fmt"

	"golang.org/x/exp/maps"
)

// Similar to the concept of a [legal peppercorn](https://en.wikipedia.org/wiki/Peppercorn_(law)), this instance of
// nothingness is required in order to transact with Go's map[T]struct{} idiom.
var peppercorn = struct{}{}

// MapSet is a struct that adds some convenience to the otherwise cumbersome map[T]struct{} idiom used in Go to
// implement sets of comparable types.
type MapSet[T comparable] struct {
	m map[T]struct{}
}

// NewMapSet returns a MapSet ready for use. Optionally, a desired size for the MapSet can be passed as an argument,
// as in the argument to make() for a map type.
func NewMapSet[T comparable](args ...int) MapSet[T] {
	if len(args) > 1 {
		panic(fmt.Sprintf("too many arguments passed to NewMapSet(). got: %v, expected 0 or 1", len(args)))
	}

	var size int
	if len(args) == 1 {
		size = args[0]
	}

	var m map[T]struct{}
	if size > 0 {
		m = make(map[T]struct{}, size)
	}

	return MapSet[T]{m: m}
}

// NewMapSetFromSlice creates a MapSet of size len(items) and calls AddSlice(items) on it.
func NewMapSetFromSlice[T comparable](items []T) MapSet[T] {
	h := NewMapSet[T](len(items))
	h.AddSlice(items)
	return h
}

// Add an item to the set. Returns true if the item did not exist in the set.
func (h *MapSet[T]) Add(item T) bool {
	if h.m == nil {
		h.m = map[T]struct{}{}
	}

	_, exists := h.m[item]
	h.m[item] = peppercorn
	return !exists
}

// AddSlice adds a slice of items to the set, returning true if any new items were added to the set.
func (h *MapSet[T]) AddSlice(items []T) bool {
	modified := false
	for _, i := range items {
		modified = h.Add(i) || modified
	}
	return modified
}

// Remove an item from the Set. Returns true if the item existed in the set.
func (h *MapSet[T]) Remove(item T) bool {
	_, exists := h.m[item]
	delete(h.m, item)
	return exists
}

// RemoveSlice removes a slice of items from the set, returning true if any items existed in the set.
func (h *MapSet[T]) RemoveSlice(items []T) bool {
	modified := false
	for _, i := range items {
		modified = h.Remove(i) || modified
	}
	return modified
}

// Contains returns whether the item exists in the set
func (h MapSet[T]) Contains(item T) bool {
	_, exists := h.m[item]
	return exists
}

// Intersection returns the items common to both h and o.
func (h MapSet[T]) Intersection(o MapSet[T]) MapSet[T] {
	intersection := NewMapSet[T]()
	for item := range h.m {
		if o.Contains(item) {
			intersection.Add(item)
		}
	}
	return intersection
}

// Iterate the items in the set, calling callback for each item. If the callback returns false, iteration is halted.
// Iteration order is undefined.
func (h MapSet[T]) Iterate(callback func(item T) bool) {
	for item := range h.m {
		if !callback(item) {
			break
		}
	}
}

func (h MapSet[T]) Slice() []T {
	if h.m == nil {
		return nil
	}
	return maps.Keys(h.m)
}

// Len returns the size of the MapSet
func (h MapSet[T]) Len() int {
	return len(h.m)
}

// Equal returns whether the same items exist in both h and o
func (h MapSet[T]) Equal(o MapSet[T]) bool {
	if len(h.m) != len(o.m) {
		return false
	}

	for item := range h.m {
		if !o.Contains(item) {
			return false
		}
	}
	return true
}

// MarshalJSON serializes a MapSet as a JSON array. Order is non-deterministic.
func (h MapSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Slice())
}

// UnmarshalJSON deserializes a MapSet from a JSON array.
func (h *MapSet[T]) UnmarshalJSON(b []byte) error {
	var s []T
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*h = NewMapSetFromSlice(s)
	return nil
}
