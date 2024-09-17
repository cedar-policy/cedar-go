package types

import (
	"encoding/json"
	"hash/fnv"
	"strconv"
)

// A String is a sequence of characters consisting of letters, numbers, or symbols.
type String string

func (a String) Equal(bi Value) bool {
	b, ok := bi.(String)
	return ok && a == b
}

// ExplicitMarshalJSON marshals the String into JSON.
func (v String) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(v) }

// String produces an unquoted string representation of the String, e.g. `hello`.
func (v String) String() string {
	return string(v)
}

// MarshalCedar produces a valid MarshalCedar language representation of the String, e.g. `"hello"`.
func (v String) MarshalCedar() []byte {
	return []byte(strconv.Quote(string(v)))
}

func (v String) hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(v))
	return h.Sum64()
}
