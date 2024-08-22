package types

import (
	"bytes"
	"encoding/json"
	"slices"
	"strconv"

	"golang.org/x/exp/maps"
)

// A Record is a collection of attributes. Each attribute consists of a name and
// an associated value. Names are simple strings. Values can be of any type.
type Record map[String]Value

// Equals returns true if the records are Equal.
func (a Record) Equal(bi Value) bool {
	b, ok := bi.(Record)
	if !ok || len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok || !av.Equal(bv) {
			return false
		}
	}
	return true
}

func (v *Record) UnmarshalJSON(b []byte) error {
	var res map[string]explicitValue
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}
	*v = Record{}
	for kk, vv := range res {
		(*v)[String(kk)] = vv.Value
	}
	return nil
}

// MarshalJSON marshals the Record into JSON, the marshaller uses the explicit
// JSON form for all the values in the Record.
func (v Record) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('{')
	keys := maps.Keys(v)
	slices.Sort(keys)
	for i, kk := range keys {
		if i > 0 {
			w.WriteByte(',')
		}
		kb, _ := json.Marshal(kk) // json.Marshal cannot error on strings
		w.Write(kb)
		w.WriteByte(':')
		vv := v[kk]
		vb, err := vv.ExplicitMarshalJSON()
		if err != nil {
			return nil, err
		}
		w.Write(vb)
	}
	w.WriteByte('}')
	return w.Bytes(), nil
}

// ExplicitMarshalJSON marshals the Record into JSON, the marshaller uses the
// explicit JSON form for all the values in the Record.
func (v Record) ExplicitMarshalJSON() ([]byte, error) { return v.MarshalJSON() }

// String produces a string representation of the Record, e.g. `{"a":1,"b":2,"c":3}`.
func (r Record) String() string { return string(r.MarshalCedar()) }

// MarshalCedar produces a valid MarshalCedar language representation of the Record, e.g. `{"a":1,"b":2,"c":3}`.
func (r Record) MarshalCedar() []byte {
	var sb bytes.Buffer
	sb.WriteRune('{')
	first := true
	keys := maps.Keys(r)
	slices.Sort(keys)
	for _, k := range keys {
		v := r[k]
		if !first {
			sb.WriteString(", ")
		}
		first = false
		sb.WriteString(strconv.Quote(string(k)))
		sb.WriteString(":")
		sb.Write(v.MarshalCedar())
	}
	sb.WriteRune('}')
	return sb.Bytes()
}
func (v Record) deepClone() Value { return v.DeepClone() }

// DeepClone returns a deep clone of the Record.
func (v Record) DeepClone() Record {
	if v == nil {
		return v
	}
	res := make(Record, len(v))
	for k, vv := range v {
		res[k] = vv.deepClone()
	}
	return res
}
