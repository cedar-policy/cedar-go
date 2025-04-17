/*
Package ast exposes the internal AST used within cedar-go.  This AST is
subject to change.  The AST is most useful for analyzing existing policies
created by the Cedar / JSON parser or created using the external AST.  The
external AST is a type definition of the internal AST, so you can cast from the
external to internal types.

Example:

	import (
		"github.com/cedar-policy/cedar-go/ast"
		internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
	)

	func main() {
		policy := ast.Permit()
		internal := (*internalast.Policy)(policy)
		_ = internal
	}
*/
package ast
