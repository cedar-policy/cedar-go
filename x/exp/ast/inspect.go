package ast

// Inspect traverses an AST in depth-first order. For each node, the
// provided function fn is called. If fn returns true, Inspect will
// recursively inspect the node's children. Returning false skips the
// children of that node.
func Inspect(n Node, fn func(IsNode) bool) {
	inspectNode(n.v, fn)
}

func inspectNode(n IsNode, fn func(IsNode) bool) {
	if n == nil {
		return
	}
	if !fn(n) {
		return
	}
	n.inspect(fn)
}
