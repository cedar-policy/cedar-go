package types

import (
	"bytes"
	"encoding/json"
	"iter"
	"maps"
	"slices"
)

// A Set is an immutable collection of elements that can be of the same or different types.
type Set struct {
	s       map[uint64]Value
	hashVal uint64
}

// NewSet returns an immutable Set given a variadic set of Values. Duplicates are removed and order is not preserved.
func NewSet(v ...Value) Set {
	var set map[uint64]Value
	if v != nil {
		set = make(map[uint64]Value, len(v))
	}
	for _, vv := range v {
		hash := vv.hash()

		// Insert the value into the map. Deal with collisions via open addressing by simply incrementing the hash
		// value. This method is safe so long as Set is immutable because nothing can be removed from the map.
		for {
			existing, ok := set[hash]
			if !ok {
				set[hash] = vv
				break
			} else if vv.Equal(existing) {
				// found duplicate in slice
				break
			}
			hash++
		}
	}

	// Special case hashVal for empty set to 0 so that the return value of Value.hash() of Set{} and NewSet([]Value{})
	// are the same
	var hashVal uint64
	for v := range maps.Values(set) {
		hashVal += v.hash()
	}

	return Set{s: set, hashVal: hashVal}
}

// Len returns the number of unique Values in the Set
func (s Set) Len() int {
	return len(s.s)
}

// SetIterator defines the type of the iteration callback function
type SetIterator func(Value) bool

// Iterate calls iter for each item in the Set. Returning false from the iter function causes iteration to cease.
// Iteration order is non-deterministic.
//
// Deprecated: use All() instead.
func (s Set) Iterate(iter SetIterator) {
	for _, v := range s.s {
		if !iter(v) {
			break
		}
	}
}

// All returns an iterator over elements in the set. Iteration order is non-deterministic.
func (s Set) All() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		for _, item := range s.s {
			if !yield(item) {
				return
			}
		}
	}
}

// Contains returns true if the Value v is present in the Set
func (s Set) Contains(v Value) bool {
	hash := v.hash()

	for {
		existing, ok := s.s[hash]
		if !ok {
			return false
		} else if v.Equal(existing) {
			return true
		}
		hash++
	}
}

// Slice returns a slice of the Values in the Set which is safe to mutate. The order of the values is non-deterministic.
func (s Set) Slice() []Value {
	if s.s == nil {
		return nil
	}
	return slices.Collect(maps.Values(s.s))
}

// Equal returns true if the sets are Equal.
func (s Set) Equal(bi Value) bool {
	bs, ok := bi.(Set)
	if !ok {
		return false
	}

	if len(s.s) != len(bs.s) || s.hashVal != bs.hashVal {
		return false
	}

	for _, v := range s.s {
		if !bs.Contains(v) {
			return false
		}
	}
	return true
}

func (v *explicitValue) UnmarshalJSON(b []byte) error {
	return UnmarshalJSON(b, &v.Value)
}

// UnmarshalJSON parses a JSON-encoded Cedar set literal into a Set
func (s *Set) UnmarshalJSON(b []byte) error {
	var res []explicitValue
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	vals := make([]Value, len(res))
	for i, vv := range res {
		vals[i] = vv.Value
	}

	*s = NewSet(vals...)
	return nil
}

// MarshalJSON marshals the Set into JSON.
// Set elements are rendered in hash order, which may differ from the original order.
func (s Set) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('[')
	orderedKeys := slices.Collect(maps.Keys(s.s))
	slices.Sort(orderedKeys)
	for i, k := range orderedKeys {
		if i != 0 {
			w.WriteByte(',')
		}
		b, err := json.Marshal(s.s[k])
		if err != nil {
			return nil, err
		}
		w.Write(b)
	}
	w.WriteByte(']')
	return w.Bytes(), nil
}

// String produces a string representation of the Set, e.g. `[1,2,3]`.
func (s Set) String() string { return string(s.MarshalCedar()) }

// MarshalCedar produces a valid MarshalCedar language representation of the Set, e.g. `[1,2,3]`.
// Set elements are rendered in hash order, which may differ from the original order.
func (s Set) MarshalCedar() []byte {
	var sb bytes.Buffer
	sb.WriteRune('[')
	orderedKeys := slices.Collect(maps.Keys(s.s))
	slices.Sort(orderedKeys)
	for i, k := range orderedKeys {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.Write(s.s[k].MarshalCedar())
	}
	sb.WriteRune(']')
	return sb.Bytes()
}

func (s Set) hash() uint64 {
	return s.hashVal
}
