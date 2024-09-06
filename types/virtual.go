package types

type Lesser interface {
	Value
	Less(Value) bool
	LessEqual(Value) bool
}
