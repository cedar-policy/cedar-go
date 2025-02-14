// Package types contains primitive, plain-old-data types including:
//   - [Cedar data types], which implement the [Value] interface
//   - [Entity] and friends, including a JSON marshaler for interacting with [JSON-encoded entities]
//   - The [Pattern] struct, used for both programmatic and textual/JSON AST construction
//   - Authorization types used by both the cedar package and the experimental batch package, in order to avoid a
//     circular dependency
//
// Types contained herein which are generally useful to the public are re-exported via the cedar package; it should be
// unlikely that users need to import this package directly.
//
// [Cedar data types]: https://docs.cedarpolicy.com/policies/syntax-datatypes.html
// [JSON-encoded entities]: https://docs.cedarpolicy.com/auth/entities-syntax.html
package types
