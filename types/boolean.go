package types

// A Boolean is a value that is either true or false.
type Boolean bool

const (
	True  = Boolean(true)
	False = Boolean(false)
)

func (b Boolean) Equal(bi Value) bool {
	bo, ok := bi.(Boolean)
	return ok && b == bo
}

// String produces a string representation of the Boolean, e.g. `true`.
func (b Boolean) String() string { return string(b.MarshalCedar()) }

// MarshalCedar produces a valid MarshalCedar language representation of the Boolean, e.g. `true`.
func (b Boolean) MarshalCedar() []byte {
	if b {
		return []byte("true")
	}
	return []byte("false")
}

func (b Boolean) hash() uint64 {
	if b {
		return 1
	}
	return 0
}
