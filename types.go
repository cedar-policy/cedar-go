package cedar

// The cedar package re-exports the types package's API, so that most callers
// only need to import cedar.

//go:generate go run ./internal/cmd/reexport -from github.com/cedar-policy/cedar-go/types -out types_gen.go -exclude Ident,ImplicitlyMarshaledEntityUID,RecordIterator,SetIterator,UnmarshalJSON
