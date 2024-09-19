package types

import (
	"bytes"
	"encoding/json"
	"slices"

	"golang.org/x/exp/maps"
)

// A Set is an immutable collection of elements that can be of the same or different types.
type Set struct {
	s       map[uint64]Value
	hashVal uint64
}

// NewSet returns an immutable Set given a Go slice of Values. Duplicates are removed and order is not preserved.
func NewSet(v []Value) Set {
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
	for _, v := range maps.Values(set) {
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
func (s Set) Iterate(iter SetIterator) {
	for _, v := range s.s {
		if !iter(v) {
			break
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
	return maps.Values(s.s)
}

// Equal returns true if the sets are Equal.
func (as Set) Equal(bi Value) bool {
	bs, ok := bi.(Set)
	if !ok {
		return false
	}

	if len(as.s) != len(bs.s) || as.hashVal != bs.hashVal {
		return false
	}

	for _, v := range as.s {
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
func (v *Set) UnmarshalJSON(b []byte) error {
	var res []explicitValue
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	vals := make([]Value, len(res))
	for i, vv := range res {
		vals[i] = vv.Value
	}

	*v = NewSet(vals)
	return nil
}

// MarshalJSON marshals the Set into JSON, the marshaller uses the explicit JSON
// form for all the values in the Set and always orders elements by their hash
// hash order, which may differ from the original order.
func (v Set) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('[')
	orderedKeys := maps.Keys(v.s)
	slices.Sort(orderedKeys)
	for i, k := range orderedKeys {
		if i != 0 {
			w.WriteByte(',')
		}
		b, err := v.s[k].ExplicitMarshalJSON()
		if err != nil {
			return nil, err
		}
		w.Write(b)
	}
	w.WriteByte(']')
	return w.Bytes(), nil
}

// ExplicitMarshalJSON marshals the Set into JSON, the marshaller uses the
// explicit JSON form for all the values in the Set.
func (v Set) ExplicitMarshalJSON() ([]byte, error) { return v.MarshalJSON() }

// String produces a string representation of the Set, e.g. `[1,2,3]`.
func (v Set) String() string { return string(v.MarshalCedar()) }

// MarshalCedar produces a valid MarshalCedar language representation of the Set, e.g. `[1,2,3]`.
// Set elements are rendered in hash order, which may differ from the original order.
func (v Set) MarshalCedar() []byte {
	var sb bytes.Buffer
	sb.WriteRune('[')
	orderedKeys := maps.Keys(v.s)
	slices.Sort(orderedKeys)
	for i, k := range orderedKeys {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.Write(v.s[k].MarshalCedar())
	}
	sb.WriteRune(']')
	return sb.Bytes()
}

func (v Set) hash() uint64 {
	return v.hashVal
}
