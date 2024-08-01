package ast

import "github.com/cedar-policy/cedar-go/types"

//   ____                                 _
//  / ___|___  _ __ ___  _ __   __ _ _ __(_)___  ___  _ __
// | |   / _ \| '_ ` _ \| '_ \ / _` | '__| / __|/ _ \| '_ \
// | |__| (_) | | | | | | |_) | (_| | |  | \__ \ (_) | | | |
//  \____\___/|_| |_| |_| .__/ \__,_|_|  |_|___/\___/|_| |_|
//                      |_|

func (lhs Node) Equals(rhs Node) Node {
	return newBinaryNode(nodeTypeEquals, lhs, rhs)
}

func (lhs Node) NotEquals(rhs Node) Node {
	return newBinaryNode(nodeTypeNotEquals, lhs, rhs)
}

func (lhs Node) LessThan(rhs Node) Node {
	return newBinaryNode(nodeTypeLess, lhs, rhs)
}

func (lhs Node) LessThanOrEqual(rhs Node) Node {
	return newBinaryNode(nodeTypeLessEqual, lhs, rhs)
}

func (lhs Node) GreaterThan(rhs Node) Node {
	return newBinaryNode(nodeTypeGreater, lhs, rhs)
}

func (lhs Node) GreaterThanOrEqual(rhs Node) Node {
	return newBinaryNode(nodeTypeGreaterEqual, lhs, rhs)
}

func (lhs Node) LessThanExt(rhs Node) Node {
	return newExtMethodCallNode(lhs, "lessThan", rhs)
}

func (lhs Node) LessThanOrEqualExt(rhs Node) Node {
	return newExtMethodCallNode(lhs, "lessThanOrEqual", rhs)
}

func (lhs Node) GreaterThanExt(rhs Node) Node {
	return newExtMethodCallNode(lhs, "greaterThan", rhs)
}

func (lhs Node) GreaterThanOrEqualExt(rhs Node) Node {
	return newExtMethodCallNode(lhs, "greaterThanOrEqual", rhs)
}

func (lhs Node) Like(patt string) Node {
	return newBinaryNode(nodeTypeLike, lhs, String(types.String(patt)))
}

//  _                _           _
// | |    ___   __ _(_) ___ __ _| |
// | |   / _ \ / _` | |/ __/ _` | |
// | |__| (_) | (_| | | (_| (_| | |
// |_____\___/ \__, |_|\___\__,_|_|
//             |___/

func (lhs Node) And(rhs Node) Node {
	return newBinaryNode(nodeTypeAnd, lhs, rhs)
}

func (lhs Node) Or(rhs Node) Node {
	return newBinaryNode(nodeTypeOr, lhs, rhs)
}

func Not(rhs Node) Node {
	return newUnaryNode(nodeTypeNot, rhs)
}

func Negate(rhs Node) Node {
	return newUnaryNode(nodeTypeNegate, rhs)
}

func If(condition Node, ifTrue Node, ifFalse Node) Node {
	return newTrinaryNode(nodeTypeIf, condition, ifTrue, ifFalse)
}

//     _         _ _   _                    _   _
//    / \   _ __(_) |_| |__  _ __ ___   ___| |_(_) ___
//   / _ \ | '__| | __| '_ \| '_ ` _ \ / _ \ __| |/ __|
//  / ___ \| |  | | |_| | | | | | | | |  __/ |_| | (__
// /_/   \_\_|  |_|\__|_| |_|_| |_| |_|\___|\__|_|\___|

func (lhs Node) Plus(rhs Node) Node {
	return newBinaryNode(nodeTypeAdd, lhs, rhs)
}

func (lhs Node) Minus(rhs Node) Node {
	return newBinaryNode(nodeTypeSub, lhs, rhs)
}

func (lhs Node) Times(rhs Node) Node {
	return newBinaryNode(nodeTypeMult, lhs, rhs)
}

//  _   _ _                         _
// | | | (_) ___ _ __ __ _ _ __ ___| |__  _   _
// | |_| | |/ _ \ '__/ _` | '__/ __| '_ \| | | |
// |  _  | |  __/ | | (_| | | | (__| | | | |_| |
// |_| |_|_|\___|_|  \__,_|_|  \___|_| |_|\__, |
//                                        |___/

func (lhs Node) In(rhs Node) Node {
	return newBinaryNode(nodeTypeIn, lhs, rhs)
}

func (lhs Node) Is(entityType types.String) Node {
	return newBinaryNode(nodeTypeIs, lhs, String(entityType))
}

func (lhs Node) IsIn(entityType types.String, rhs Node) Node {
	return newTrinaryNode(nodeTypeIsIn, lhs, String(entityType), rhs)
}

func (lhs Node) Contains(rhs Node) Node {
	return newBinaryNode(nodeTypeContains, lhs, rhs)
}

func (lhs Node) ContainsAll(rhs Node) Node {
	return newBinaryNode(nodeTypeContainsAll, lhs, rhs)
}

func (lhs Node) ContainsAny(rhs Node) Node {
	return newBinaryNode(nodeTypeContainsAny, lhs, rhs)
}

// Access is a convenience function that wraps a simple string
// in an ast.String() and passes it along to AccessNode.
func (lhs Node) Access(attr string) Node {
	return lhs.AccessNode(String(types.String(attr)))
}

// AccessNode is a version of the access operator which allows
// more complex access of attributes, such as might be expressed
// by this Cedar text:
//
//	resource[context.resourceAttribute] == "foo"
//
// In Golang, this could be expressed as:
//
//	ast.Resource().AccessNode(
//	    ast.Context().Access("resourceAttribute")
//	).Equals(ast.String("foo"))
func (lhs Node) AccessNode(rhs Node) Node {
	return newBinaryNode(nodeTypeAccess, lhs, rhs)
}

func (lhs Node) Has(attr string) Node {
	return newBinaryNode(nodeTypeHas, lhs, String(types.String(attr)))
}

//  ___ ____   _       _     _
// |_ _|  _ \ / \   __| | __| |_ __ ___  ___ ___
//  | || |_) / _ \ / _` |/ _` | '__/ _ \/ __/ __|
//  | ||  __/ ___ \ (_| | (_| | | |  __/\__ \__ \
// |___|_| /_/   \_\__,_|\__,_|_|  \___||___/___/

func (lhs Node) IsIpv4() Node {
	return newExtMethodCallNode(lhs, "isIpv4")
}

func (lhs Node) IsIpv6() Node {
	return newExtMethodCallNode(lhs, "isIpv6")
}

func (lhs Node) IsMulticast() Node {
	return newExtMethodCallNode(lhs, "isMulticast")
}

func (lhs Node) IsLoopback() Node {
	return newExtMethodCallNode(lhs, "isLoopback")
}

func (lhs Node) IsInRange(rhs Node) Node {
	return newExtMethodCallNode(lhs, "isInRange", rhs)
}
