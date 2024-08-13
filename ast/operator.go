package ast

import (
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

//   ____                                 _
//  / ___|___  _ __ ___  _ __   __ _ _ __(_)___  ___  _ __
// | |   / _ \| '_ ` _ \| '_ \ / _` | '__| / __|/ _ \| '_ \
// | |__| (_) | | | | | | |_) | (_| | |  | \__ \ (_) | | | |
//  \____\___/|_| |_| |_| .__/ \__,_|_|  |_|___/\___/|_| |_|
//                      |_|

func (lhs Node) Equals(rhs Node) Node {
	return newNode(lhs.Node.Equals(rhs.Node))
}

func (lhs Node) NotEquals(rhs Node) Node {
	return newNode(lhs.Node.NotEquals(rhs.Node))
}

func (lhs Node) LessThan(rhs Node) Node {
	return newNode(lhs.Node.LessThan(rhs.Node))
}

func (lhs Node) LessThanOrEqual(rhs Node) Node {
	return newNode(lhs.Node.LessThanOrEqual(rhs.Node))
}

func (lhs Node) GreaterThan(rhs Node) Node {
	return newNode(lhs.Node.GreaterThan(rhs.Node))
}

func (lhs Node) GreaterThanOrEqual(rhs Node) Node {
	return newNode(lhs.Node.GreaterThanOrEqual(rhs.Node))
}

func (lhs Node) LessThanExt(rhs Node) Node {
	return newNode(lhs.Node.LessThanExt(rhs.Node))
}

func (lhs Node) LessThanOrEqualExt(rhs Node) Node {
	return newNode(lhs.Node.LessThanOrEqualExt(rhs.Node))
}

func (lhs Node) GreaterThanExt(rhs Node) Node {
	return newNode(lhs.Node.GreaterThanExt(rhs.Node))
}

func (lhs Node) GreaterThanOrEqualExt(rhs Node) Node {
	return newNode(lhs.Node.GreaterThanOrEqualExt(rhs.Node))
}

func (lhs Node) Like(pattern types.Pattern) Node {
	return newNode(lhs.Node.Like(pattern))
}

//  _                _           _
// | |    ___   __ _(_) ___ __ _| |
// | |   / _ \ / _` | |/ __/ _` | |
// | |__| (_) | (_| | | (_| (_| | |
// |_____\___/ \__, |_|\___\__,_|_|
//             |___/

func (lhs Node) And(rhs Node) Node {
	return newNode(lhs.Node.And(rhs.Node))
}

func (lhs Node) Or(rhs Node) Node {
	return newNode(lhs.Node.Or(rhs.Node))
}

func Not(rhs Node) Node {
	return newNode(ast.Not(rhs.Node))
}

func If(condition Node, ifTrue Node, ifFalse Node) Node {
	return newNode(ast.If(condition.Node, ifTrue.Node, ifFalse.Node))
}

//     _         _ _   _                    _   _
//    / \   _ __(_) |_| |__  _ __ ___   ___| |_(_) ___
//   / _ \ | '__| | __| '_ \| '_ ` _ \ / _ \ __| |/ __|
//  / ___ \| |  | | |_| | | | | | | | |  __/ |_| | (__
// /_/   \_\_|  |_|\__|_| |_|_| |_| |_|\___|\__|_|\___|

func (lhs Node) Plus(rhs Node) Node {
	return newNode(lhs.Node.Plus(rhs.Node))
}

func (lhs Node) Minus(rhs Node) Node {
	return newNode(lhs.Node.Minus(rhs.Node))
}

func (lhs Node) Times(rhs Node) Node {
	return newNode(lhs.Node.Times(rhs.Node))
}

func Negate(rhs Node) Node {
	return newNode(ast.Negate(rhs.Node))
}

//  _   _ _                         _
// | | | (_) ___ _ __ __ _ _ __ ___| |__  _   _
// | |_| | |/ _ \ '__/ _` | '__/ __| '_ \| | | |
// |  _  | |  __/ | | (_| | | | (__| | | | |_| |
// |_| |_|_|\___|_|  \__,_|_|  \___|_| |_|\__, |
//                                        |___/

func (lhs Node) In(rhs Node) Node {
	return newNode(lhs.Node.In(rhs.Node))
}

func (lhs Node) Is(entityType types.Path) Node {
	return newNode(lhs.Node.Is(entityType))
}

func (lhs Node) IsIn(entityType types.Path, rhs Node) Node {
	return newNode(lhs.Node.IsIn(entityType, rhs.Node))
}

func (lhs Node) Contains(rhs Node) Node {
	return newNode(lhs.Node.Contains(rhs.Node))
}

func (lhs Node) ContainsAll(rhs Node) Node {
	return newNode(lhs.Node.ContainsAll(rhs.Node))
}

func (lhs Node) ContainsAny(rhs Node) Node {
	return newNode(lhs.Node.ContainsAny(rhs.Node))
}

func (lhs Node) Access(attr string) Node {
	return newNode(lhs.Node.Access(attr))
}

func (lhs Node) Has(attr string) Node {
	return newNode(lhs.Node.Has(attr))
}

//  ___ ____   _       _     _
// |_ _|  _ \ / \   __| | __| |_ __ ___  ___ ___
//  | || |_) / _ \ / _` |/ _` | '__/ _ \/ __/ __|
//  | ||  __/ ___ \ (_| | (_| | | |  __/\__ \__ \
// |___|_| /_/   \_\__,_|\__,_|_|  \___||___/___/

func (lhs Node) IsIpv4() Node {
	return newNode(lhs.Node.IsIpv4())
}

func (lhs Node) IsIpv6() Node {
	return newNode(lhs.Node.IsIpv6())
}

func (lhs Node) IsMulticast() Node {
	return newNode(lhs.Node.IsMulticast())
}

func (lhs Node) IsLoopback() Node {
	return newNode(lhs.Node.IsLoopback())
}

func (lhs Node) IsInRange(rhs Node) Node {
	return newNode(lhs.Node.IsInRange(rhs.Node))
}
