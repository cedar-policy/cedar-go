package ast

import "github.com/cedar-policy/cedar-go/types"

//   ____                                 _
//  / ___|___  _ __ ___  _ __   __ _ _ __(_)___  ___  _ __
// | |   / _ \| '_ ` _ \| '_ \ / _` | '__| / __|/ _ \| '_ \
// | |__| (_) | | | | | | |_) | (_| | |  | \__ \ (_) | | | |
//  \____\___/|_| |_| |_| .__/ \__,_|_|  |_|___/\___/|_| |_|
//                      |_|

func (lhs Node) Equals(rhs Node) Node {
	return newOpNode(nodeTypeEquals, lhs, rhs)
}

func (lhs Node) NotEquals(rhs Node) Node {
	return newOpNode(nodeTypeNotEquals, lhs, rhs)
}

func (lhs Node) LessThan(rhs Node) Node {
	return newOpNode(nodeTypeLess, lhs, rhs)
}

func (lhs Node) LessThanOrEqual(rhs Node) Node {
	return newOpNode(nodeTypeLessEqual, lhs, rhs)
}

func (lhs Node) GreaterThan(rhs Node) Node {
	return newOpNode(nodeTypeGreater, lhs, rhs)
}

func (lhs Node) GreaterThanOrEqual(rhs Node) Node {
	return newOpNode(nodeTypeGreaterEqual, lhs, rhs)
}

//  _                _           _
// | |    ___   __ _(_) ___ __ _| |
// | |   / _ \ / _` | |/ __/ _` | |
// | |__| (_) | (_| | | (_| (_| | |
// |_____\___/ \__, |_|\___\__,_|_|
//             |___/

func (lhs Node) And(rhs Node) Node {
	return newOpNode(nodeTypeAnd, lhs, rhs)
}

func (lhs Node) Or(rhs Node) Node {
	return newOpNode(nodeTypeOr, lhs, rhs)
}

func Not(rhs Node) Node {
	return newOpNode(nodeTypeNot, rhs)
}

func Negate(rhs Node) Node {
	return newOpNode(nodeTypeNegate, rhs)
}

func If(condition Node, ifTrue Node, ifFalse Node) Node {
	return newOpNode(nodeTypeIf, condition, ifTrue, ifFalse)
}

//     _         _ _   _                    _   _
//    / \   _ __(_) |_| |__  _ __ ___   ___| |_(_) ___
//   / _ \ | '__| | __| '_ \| '_ ` _ \ / _ \ __| |/ __|
//  / ___ \| |  | | |_| | | | | | | | |  __/ |_| | (__
// /_/   \_\_|  |_|\__|_| |_|_| |_| |_|\___|\__|_|\___|

func (lhs Node) Plus(rhs Node) Node {
	return newOpNode(nodeTypeAdd, lhs, rhs)
}

func (lhs Node) Minus(rhs Node) Node {
	return newOpNode(nodeTypeSub, lhs, rhs)
}

func (lhs Node) Times(rhs Node) Node {
	return newOpNode(nodeTypeMult, lhs, rhs)
}

//  _   _ _                         _
// | | | (_) ___ _ __ __ _ _ __ ___| |__  _   _
// | |_| | |/ _ \ '__/ _` | '__/ __| '_ \| | | |
// |  _  | |  __/ | | (_| | | | (__| | | | |_| |
// |_| |_|_|\___|_|  \__,_|_|  \___|_| |_|\__, |
//                                        |___/

func (lhs Node) In(rhs Node) Node {
	return newOpNode(nodeTypeIn, lhs, rhs)
}

func (lhs Node) Has(rhs Node) Node {
	return newOpNode(nodeTypeHas, lhs, rhs)
}

func (lhs Node) Is(rhs Node) Node {
	return newOpNode(nodeTypeIs, lhs, rhs)
}

func (lhs Node) Contains(rhs Node) Node {
	return newOpNode(nodeTypeContains, lhs, rhs)
}

func (lhs Node) ContainsAll(rhs Node) Node {
	return newOpNode(nodeTypeContainsAll, lhs, rhs)
}

func (lhs Node) ContainsAny(rhs Node) Node {
	return newOpNode(nodeTypeContainsAny, lhs, rhs)
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
	return newOpNode(nodeTypeAccess, lhs, rhs)
}

//  ___ ____   _       _     _
// |_ _|  _ \ / \   __| | __| |_ __ ___  ___ ___
//  | || |_) / _ \ / _` |/ _` | '__/ _ \/ __/ __|
//  | ||  __/ ___ \ (_| | (_| | | |  __/\__ \__ \
// |___|_| /_/   \_\__,_|\__,_|_|  \___||___/___/

func (lhs Node) IsIpv4() Node {
	return newOpNode(nodeTypeIsIpv4, lhs)
}

func (lhs Node) IsIpv6() Node {
	return newOpNode(nodeTypeIsIpv6, lhs)
}

func (lhs Node) IsMulticast() Node {
	return newOpNode(nodeTypeIsMulticast, lhs)
}

func (lhs Node) IsLoopback() Node {
	return newOpNode(nodeTypeIsLoopback, lhs)
}

func (lhs Node) IsInRange(rhs Node) Node {
	return newOpNode(nodeTypeIsInRange, lhs, rhs)
}

func newOpNode(op nodeType, args ...Node) Node {
	return Node{nodeType: op, args: args}
}
