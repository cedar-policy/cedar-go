package types

import (
	"bytes"
	"encoding/json"
	"slices"
)

// A Set is an immutable collection of elements that can be of the same or different types.
type Set struct {
	v []Value
}

// NewSet takes a slice of Values and stores a clone of the values internally.
func NewSet(v []Value) Set {
	newSlice := make([]Value, 0, len(v))
	for _, vv := range v {
		if slices.ContainsFunc(newSlice, func(vvv Value) bool { return vv.Equal(vvv) }) {
			continue
		}
		newSlice = append(newSlice, vv.deepClone())
	}
	return Set{v: newSlice}
}

// Len returns the number of unique Values in the Set
func (s Set) Len() int {
	return len(s.v)
}

// SetIterator defines the type of the iteration callback function
type SetIterator func(Value) bool

// Iterate calls iter for each item in the Set. Returning false from the iter function causes iteration to cease.
// Iteration order is non-deterministic.
func (s Set) Iterate(iter SetIterator) {
	for _, v := range s.v {
		if !iter(v) {
			break
		}
	}
}

// Contains returns true if the Value v is present in the Set
func (s Set) Contains(v Value) bool {
	for _, e := range s.v {
		if e.Equal(v) {
			return true
		}
	}
	return false
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
	for _, a := range as.v {
		if !bs.Contains(a) {
			return false
		}
	}
	for _, b := range bs.v {
		if !as.Contains(b) {
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
// form for all the values in the Set.
func (v Set) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('[')
	for i, vv := range v.v {
		if i > 0 {
			w.WriteByte(',')
		}
		b, err := vv.ExplicitMarshalJSON()
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
func (v Set) MarshalCedar() []byte {
	var sb bytes.Buffer
	sb.WriteRune('[')
	for i, elem := range v.v {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.Write(elem.MarshalCedar())
	}
	sb.WriteRune(']')
	return sb.Bytes()
}
func (v Set) deepClone() Value { return v.DeepClone() }

// DeepClone returns a deep clone of the Set.
func (v Set) DeepClone() Set {
	if v.v == nil {
		return Set{nil}
	}
	vals := make([]Value, len(v.v))
	for i, vv := range v.v {
		vals[i] = vv.deepClone()
	}
	return NewSet(vals)
}
