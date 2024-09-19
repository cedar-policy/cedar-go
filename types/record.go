package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"hash/fnv"
	"slices"
	"strconv"

	"golang.org/x/exp/maps"
)

type RecordMap = map[String]Value

// A Record is an immutable collection of attributes. Each attribute consists of a name and
// an associated value. Names are simple strings. Values can be of any type.
type Record struct {
	m       RecordMap
	hashVal uint64
}

// NewRecord returns an immutable Record given a Go map of Strings to Values
func NewRecord(m RecordMap) Record {
	// Special case hashVal for empty map to 0 so that the return value of Value.hash() of Record{} and
	// NewRecord(RecordMap{}) are the same
	var hashVal uint64
	if len(m) > 0 {
		orderedKeys := maps.Keys(m)
		slices.Sort(orderedKeys)

		h := fnv.New64()
		for _, k := range orderedKeys {
			_, _ = h.Write([]byte(k))
			_ = binary.Write(h, binary.LittleEndian, m[k].hash())
		}
		hashVal = h.Sum64()
	}

	if m != nil {
		m = maps.Clone(m)
	}

	return Record{m: m, hashVal: hashVal}
}

func (r Record) Len() int {
	return len(r.m)
}

// RecordIterator is called for each item in the Record when passed to Iterate. Returning false from this function
// causes iteration to cease.
type RecordIterator func(String, Value) bool

// Iterate calls iter for each key/value pair in the record. Iteration order is non-deterministic.
func (r Record) Iterate(iter RecordIterator) {
	for k, v := range r.m {
		if !iter(k, v) {
			break
		}
	}
}

// Get returns (v, true) where v is the Value associated with key s, if Record contains key s. Get returns (nil, false)
// if Record does not contain key s.
func (r Record) Get(s String) (Value, bool) {
	v, ok := r.m[s]
	return v, ok
}

// Map returns a clone of the Record's internal RecordMap which is safe to mutate.
func (r Record) Map() RecordMap {
	if r.m == nil {
		return nil
	}
	return maps.Clone(r.m)
}

// Equals returns true if the records are Equal.
func (a Record) Equal(bi Value) bool {
	b, ok := bi.(Record)
	if !ok || len(a.m) != len(b.m) || a.hashVal != b.hashVal {
		return false
	}
	for k, av := range a.m {
		bv, ok := b.m[k]
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
	m := make(RecordMap, len(res))
	for kk, vv := range res {
		m[String(kk)] = vv.Value
	}
	*v = NewRecord(m)
	return nil
}

// MarshalJSON marshals the Record into JSON, the marshaller uses the explicit
// JSON form for all the values in the Record.
func (v Record) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('{')
	keys := maps.Keys(v.m)
	slices.Sort(keys)
	for i, kk := range keys {
		if i > 0 {
			w.WriteByte(',')
		}
		kb, _ := json.Marshal(kk) // json.Marshal cannot error on strings
		w.Write(kb)
		w.WriteByte(':')
		vv := v.m[kk]
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
	keys := maps.Keys(r.m)
	slices.Sort(keys)
	for _, k := range keys {
		v := r.m[k]
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

func (v Record) hash() uint64 {
	return v.hashVal
}
