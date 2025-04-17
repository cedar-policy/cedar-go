package types

import (
	"hash/fnv"
	"strconv"
)

// A String is a sequence of characters consisting of letters, numbers, or symbols.
type String string

// Equal returns true if two Strings are equal
func (s String) Equal(bi Value) bool {
	b, ok := bi.(String)
	return ok && s == b
}

// String produces an unquoted string representation of the String, e.g. `hello`.
func (s String) String() string {
	return string(s)
}

// MarshalCedar produces a valid MarshalCedar language representation of the String, e.g. `"hello"`.
func (s String) MarshalCedar() []byte {
	return []byte(strconv.Quote(string(s)))
}

func (s String) hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
