package types

import (
	"bytes"
	"encoding/json"
	"strings"
)

// A Set is a collection of elements that can be of the same or different types.
type Set []Value

func (s Set) Contains(v Value) bool {
	for _, e := range s {
		if e.Equal(v) {
			return true
		}
	}
	return false
}

// Equal returns true if the sets are Equal.
func (as Set) Equal(bi Value) bool {
	bs, ok := bi.(Set)
	if !ok {
		return false
	}
	for _, a := range as {
		if !bs.Contains(a) {
			return false
		}
	}
	for _, b := range bs {
		if !as.Contains(b) {
			return false
		}
	}
	return true
}

func (v *explicitValue) UnmarshalJSON(b []byte) error {
	return UnmarshalJSON(b, &v.Value)
}

func (v *Set) UnmarshalJSON(b []byte) error {
	var res []explicitValue
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}
	for _, vv := range res {
		*v = append(*v, vv.Value)
	}
	return nil
}

// MarshalJSON marshals the Set into JSON, the marshaller uses the explicit JSON
// form for all the values in the Set.
func (v Set) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('[')
	for i, vv := range v {
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
func (v Set) String() string { return v.Cedar() }

// Cedar produces a valid Cedar language representation of the Set, e.g. `[1,2,3]`.
func (v Set) Cedar() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, elem := range v {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(elem.Cedar())
	}
	sb.WriteRune(']')
	return sb.String()
}
func (v Set) deepClone() Value { return v.DeepClone() }

// DeepClone returns a deep clone of the Set.
func (v Set) DeepClone() Set {
	if v == nil {
		return v
	}
	res := make(Set, len(v))
	for i, vv := range v {
		res[i] = vv.deepClone()
	}
	return res
}
