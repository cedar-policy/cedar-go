package types

import (
	"encoding/json"
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
func (v String) TypeName() string                     { return "string" }

// String produces an unquoted string representation of the String, e.g. `hello`.
func (v String) String() string {
	return string(v)
}

// Cedar produces a valid Cedar language representation of the String, e.g. `"hello"`.
func (v String) Cedar() string {
	return strconv.Quote(string(v))
}
func (v String) deepClone() Value { return v }
