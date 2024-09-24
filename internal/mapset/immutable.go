package mapset

import (
	"encoding/json"
)

type ImmutableMapSet[T comparable] MapSet[T]

func Immutable[T comparable](args ...T) ImmutableMapSet[T] {
	return ImmutableMapSet[T](*FromSlice(args))
}

// Contains returns whether the item exists in the set
func (h ImmutableMapSet[T]) Contains(item T) bool {
	return MapSet[T](h).Contains(item)
}

// Intersection returns the items common to both h and o.
func (h ImmutableMapSet[T]) Intersection(o Container[T]) ImmutableMapSet[T] {
	return ImmutableMapSet[T](*MapSet[T](h).Intersection(o))
}

// Iterate the items in the set, calling callback for each item. If the callback returns false, iteration is halted.
// Iteration order is undefined.
func (h ImmutableMapSet[T]) Iterate(callback func(item T) bool) {
	MapSet[T](h).Iterate(callback)
}

func (h ImmutableMapSet[T]) Slice() []T {
	return MapSet[T](h).Slice()
}

// Len returns the size of the set
func (h ImmutableMapSet[T]) Len() int {
	return MapSet[T](h).Len()
}

// Equal returns whether the same items exist in both h and o
func (h ImmutableMapSet[T]) Equal(o ImmutableMapSet[T]) bool {
	om := MapSet[T](o)
	return MapSet[T](h).Equal(&om)
}

// MarshalJSON serializes an ImmutableMapSet as a JSON array. Order is non-deterministic.
func (h ImmutableMapSet[T]) MarshalJSON() ([]byte, error) {
	return MapSet[T](h).MarshalJSON()
}

// UnmarshalJSON deserializes an ImmutableMapSet from a JSON array.
func (h *ImmutableMapSet[T]) UnmarshalJSON(b []byte) error {
	var s MapSet[T]
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*h = ImmutableMapSet[T](s)
	return nil
}
