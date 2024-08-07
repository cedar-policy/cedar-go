package ast

import "github.com/cedar-policy/cedar-go/types"

//   ____                                 _
//  / ___|___  _ __ ___  _ __   __ _ _ __(_)___  ___  _ __
// | |   / _ \| '_ ` _ \| '_ \ / _` | '__| / __|/ _ \| '_ \
// | |__| (_) | | | | | | |_) | (_| | |  | \__ \ (_) | | | |
//  \____\___/|_| |_| |_| .__/ \__,_|_|  |_|___/\___/|_| |_|
//                      |_|

func (lhs Node) Equals(rhs Node) Node {
	return newNode(nodeTypeEquals{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) NotEquals(rhs Node) Node {
	return newNode(nodeTypeNotEquals{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) LessThan(rhs Node) Node {
	return newNode(nodeTypeLessThan{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) LessThanOrEqual(rhs Node) Node {
	return newNode(nodeTypeLessThanOrEqual{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) GreaterThan(rhs Node) Node {
	return newNode(nodeTypeGreaterThan{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) GreaterThanOrEqual(rhs Node) Node {
	return newNode(nodeTypeGreaterThanOrEqual{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) LessThanExt(rhs Node) Node {
	return newMethodCall(lhs, "lessThan", rhs)
}

func (lhs Node) LessThanOrEqualExt(rhs Node) Node {
	return newMethodCall(lhs, "lessThanOrEqual", rhs)
}

func (lhs Node) GreaterThanExt(rhs Node) Node {
	return newMethodCall(lhs, "greaterThan", rhs)
}

func (lhs Node) GreaterThanOrEqualExt(rhs Node) Node {
	return newMethodCall(lhs, "greaterThanOrEqual", rhs)
}

func (lhs Node) Like(pattern Pattern) Node {
	return newNode(nodeTypeLike{Arg: lhs.v, Value: pattern})
}

//  _                _           _
// | |    ___   __ _(_) ___ __ _| |
// | |   / _ \ / _` | |/ __/ _` | |
// | |__| (_) | (_| | | (_| (_| | |
// |_____\___/ \__, |_|\___\__,_|_|
//             |___/

func (lhs Node) And(rhs Node) Node {
	return newNode(nodeTypeAnd{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Or(rhs Node) Node {
	return newNode(nodeTypeOr{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func Not(rhs Node) Node {
	return newNode(nodeTypeNot{unaryNode: unaryNode{Arg: rhs.v}})
}

func If(condition Node, ifTrue Node, ifFalse Node) Node {
	return newNode(nodeTypeIf{If: condition.v, Then: ifTrue.v, Else: ifFalse.v})
}

//     _         _ _   _                    _   _
//    / \   _ __(_) |_| |__  _ __ ___   ___| |_(_) ___
//   / _ \ | '__| | __| '_ \| '_ ` _ \ / _ \ __| |/ __|
//  / ___ \| |  | | |_| | | | | | | | |  __/ |_| | (__
// /_/   \_\_|  |_|\__|_| |_|_| |_| |_|\___|\__|_|\___|

func (lhs Node) Plus(rhs Node) Node {
	return newNode(nodeTypeAdd{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Minus(rhs Node) Node {
	return newNode(nodeTypeSub{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Times(rhs Node) Node {
	return newNode(nodeTypeMult{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func Negate(rhs Node) Node {
	return newNode(nodeTypeNegate{unaryNode: unaryNode{Arg: rhs.v}})
}

//  _   _ _                         _
// | | | (_) ___ _ __ __ _ _ __ ___| |__  _   _
// | |_| | |/ _ \ '__/ _` | '__/ __| '_ \| | | |
// |  _  | |  __/ | | (_| | | | (__| | | | |_| |
// |_| |_|_|\___|_|  \__,_|_|  \___|_| |_|\__, |
//                                        |___/

func (lhs Node) In(rhs Node) Node {
	return newNode(nodeTypeIn{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Is(entityType types.Path) Node {
	return newNode(nodeTypeIs{Left: lhs.v, EntityType: entityType})
}

func (lhs Node) IsIn(entityType types.Path, rhs Node) Node {
	return newNode(nodeTypeIsIn{nodeTypeIs: nodeTypeIs{Left: lhs.v, EntityType: entityType}, Entity: rhs.v})
}

func (lhs Node) Contains(rhs Node) Node {
	return newNode(nodeTypeContains{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) ContainsAll(rhs Node) Node {
	return newNode(nodeTypeContainsAll{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) ContainsAny(rhs Node) Node {
	return newNode(nodeTypeContainsAny{binaryNode: binaryNode{Left: lhs.v, Right: rhs.v}})
}

// Access is a convenience function that wraps a simple string
// in an ast.String() and passes it along to AccessNode.
func (lhs Node) Access(attr string) Node {
	return newNode(nodeTypeAccess{strOpNode: strOpNode{Arg: lhs.v, Value: types.String(attr)}})
}

func (lhs Node) Has(attr string) Node {
	return newNode(nodeTypeHas{strOpNode: strOpNode{Arg: lhs.v, Value: types.String(attr)}})
}

//  ___ ____   _       _     _
// |_ _|  _ \ / \   __| | __| |_ __ ___  ___ ___
//  | || |_) / _ \ / _` |/ _` | '__/ _ \/ __/ __|
//  | ||  __/ ___ \ (_| | (_| | | |  __/\__ \__ \
// |___|_| /_/   \_\__,_|\__,_|_|  \___||___/___/

func (lhs Node) IsIpv4() Node {
	return newMethodCall(lhs, "isIpv4")
}

func (lhs Node) IsIpv6() Node {
	return newMethodCall(lhs, "isIpv6")
}

func (lhs Node) IsMulticast() Node {
	return newMethodCall(lhs, "isMulticast")
}

func (lhs Node) IsLoopback() Node {
	return newMethodCall(lhs, "isLoopback")
}

func (lhs Node) IsInRange(rhs Node) Node {
	return newMethodCall(lhs, "isInRange", rhs)
}
