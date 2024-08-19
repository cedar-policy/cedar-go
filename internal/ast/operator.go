package ast

import "github.com/cedar-policy/cedar-go/types"

//   ____                                 _
//  / ___|___  _ __ ___  _ __   __ _ _ __(_)___  ___  _ __
// | |   / _ \| '_ ` _ \| '_ \ / _` | '__| / __|/ _ \| '_ \
// | |__| (_) | | | | | | |_) | (_| | |  | \__ \ (_) | | | |
//  \____\___/|_| |_| |_| .__/ \__,_|_|  |_|___/\___/|_| |_|
//                      |_|

func (lhs Node) Equals(rhs Node) Node {
	return NewNode(NodeTypeEquals{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) NotEquals(rhs Node) Node {
	return NewNode(NodeTypeNotEquals{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) LessThan(rhs Node) Node {
	return NewNode(NodeTypeLessThan{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) LessThanOrEqual(rhs Node) Node {
	return NewNode(NodeTypeLessThanOrEqual{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) GreaterThan(rhs Node) Node {
	return NewNode(NodeTypeGreaterThan{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) GreaterThanOrEqual(rhs Node) Node {
	return NewNode(NodeTypeGreaterThanOrEqual{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) DecimalLessThan(rhs Node) Node {
	return NewMethodCall(lhs, "lessThan", rhs)
}

func (lhs Node) DecimalLessThanOrEqual(rhs Node) Node {
	return NewMethodCall(lhs, "lessThanOrEqual", rhs)
}

func (lhs Node) DecimalGreaterThan(rhs Node) Node {
	return NewMethodCall(lhs, "greaterThan", rhs)
}

func (lhs Node) DecimalGreaterThanOrEqual(rhs Node) Node {
	return NewMethodCall(lhs, "greaterThanOrEqual", rhs)
}

func (lhs Node) Like(pattern types.Pattern) Node {
	return NewNode(NodeTypeLike{Arg: lhs.v, Value: pattern})
}

//  _                _           _
// | |    ___   __ _(_) ___ __ _| |
// | |   / _ \ / _` | |/ __/ _` | |
// | |__| (_) | (_| | | (_| (_| | |
// |_____\___/ \__, |_|\___\__,_|_|
//             |___/

func (lhs Node) And(rhs Node) Node {
	return NewNode(NodeTypeAnd{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Or(rhs Node) Node {
	return NewNode(NodeTypeOr{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func Not(rhs Node) Node {
	return NewNode(NodeTypeNot{UnaryNode: UnaryNode{Arg: rhs.v}})
}

func If(condition Node, ifTrue Node, ifFalse Node) Node {
	return NewNode(NodeTypeIf{If: condition.v, Then: ifTrue.v, Else: ifFalse.v})
}

//     _         _ _   _                    _   _
//    / \   _ __(_) |_| |__  _ __ ___   ___| |_(_) ___
//   / _ \ | '__| | __| '_ \| '_ ` _ \ / _ \ __| |/ __|
//  / ___ \| |  | | |_| | | | | | | | |  __/ |_| | (__
// /_/   \_\_|  |_|\__|_| |_|_| |_| |_|\___|\__|_|\___|

func (lhs Node) Add(rhs Node) Node {
	return NewNode(NodeTypeAdd{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Subtract(rhs Node) Node {
	return NewNode(NodeTypeSub{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Multiply(rhs Node) Node {
	return NewNode(NodeTypeMult{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func Negate(rhs Node) Node {
	return NewNode(NodeTypeNegate{UnaryNode: UnaryNode{Arg: rhs.v}})
}

//  _   _ _                         _
// | | | (_) ___ _ __ __ _ _ __ ___| |__  _   _
// | |_| | |/ _ \ '__/ _` | '__/ __| '_ \| | | |
// |  _  | |  __/ | | (_| | | | (__| | | | |_| |
// |_| |_|_|\___|_|  \__,_|_|  \___|_| |_|\__, |
//                                        |___/

func (lhs Node) In(rhs Node) Node {
	return NewNode(NodeTypeIn{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Is(entityType types.EntityType) Node {
	return NewNode(NodeTypeIs{Left: lhs.v, EntityType: entityType})
}

func (lhs Node) IsIn(entityType types.EntityType, rhs Node) Node {
	return NewNode(NodeTypeIsIn{NodeTypeIs: NodeTypeIs{Left: lhs.v, EntityType: entityType}, Entity: rhs.v})
}

func (lhs Node) Contains(rhs Node) Node {
	return NewNode(NodeTypeContains{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) ContainsAll(rhs Node) Node {
	return NewNode(NodeTypeContainsAll{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) ContainsAny(rhs Node) Node {
	return NewNode(NodeTypeContainsAny{BinaryNode: BinaryNode{Left: lhs.v, Right: rhs.v}})
}

func (lhs Node) Access(attr string) Node {
	return NewNode(NodeTypeAccess{StrOpNode: StrOpNode{Arg: lhs.v, Value: types.String(attr)}})
}

func (lhs Node) Has(attr string) Node {
	return NewNode(NodeTypeHas{StrOpNode: StrOpNode{Arg: lhs.v, Value: types.String(attr)}})
}

//  ___ ____   _       _     _
// |_ _|  _ \ / \   __| | __| |_ __ ___  ___ ___
//  | || |_) / _ \ / _` |/ _` | '__/ _ \/ __/ __|
//  | ||  __/ ___ \ (_| | (_| | | |  __/\__ \__ \
// |___|_| /_/   \_\__,_|\__,_|_|  \___||___/___/

func (lhs Node) IsIpv4() Node {
	return NewMethodCall(lhs, "isIpv4")
}

func (lhs Node) IsIpv6() Node {
	return NewMethodCall(lhs, "isIpv6")
}

func (lhs Node) IsMulticast() Node {
	return NewMethodCall(lhs, "isMulticast")
}

func (lhs Node) IsLoopback() Node {
	return NewMethodCall(lhs, "isLoopback")
}

func (lhs Node) IsInRange(rhs Node) Node {
	return NewMethodCall(lhs, "isInRange", rhs)
}
