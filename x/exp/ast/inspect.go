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
	switch n := n.(type) {
	case NodeTypeIfThenElse:
		inspectNode(n.If, fn)
		inspectNode(n.Then, fn)
		inspectNode(n.Else, fn)
	case NodeTypeOr:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeAnd:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeLessThan:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeLessThanOrEqual:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeGreaterThan:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeGreaterThanOrEqual:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeNotEquals:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeEquals:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeIn:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeHasTag:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeGetTag:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeContains:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeContainsAll:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeContainsAny:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeAdd:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeSub:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeMult:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case NodeTypeHas:
		inspectNode(n.Arg, fn)
	case NodeTypeAccess:
		inspectNode(n.Arg, fn)
	case NodeTypeLike:
		inspectNode(n.Arg, fn)
	case NodeTypeIs:
		inspectNode(n.Left, fn)
	case NodeTypeIsIn:
		inspectNode(n.Left, fn)
		inspectNode(n.Entity, fn)
	case NodeTypeNegate:
		inspectNode(n.Arg, fn)
	case NodeTypeNot:
		inspectNode(n.Arg, fn)
	case NodeTypeIsEmpty:
		inspectNode(n.Arg, fn)
	case NodeTypeExtensionCall:
		for _, a := range n.Args {
			inspectNode(a, fn)
		}
	case NodeTypeRecord:
		for _, e := range n.Elements {
			inspectNode(e.Value, fn)
		}
	case NodeTypeSet:
		for _, e := range n.Elements {
			inspectNode(e, fn)
		}
	case StrOpNode:
		inspectNode(n.Arg, fn)
	case BinaryNode:
		inspectNode(n.Left, fn)
		inspectNode(n.Right, fn)
	case UnaryNode:
		inspectNode(n.Arg, fn)
	}
}
