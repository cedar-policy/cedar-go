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

func (lhs Node) Equal(rhs Node) Node {
	return wrapNode(lhs.Node.Equal(rhs.Node))
}

func (lhs Node) NotEqual(rhs Node) Node {
	return wrapNode(lhs.Node.NotEqual(rhs.Node))
}

func (lhs Node) LessThan(rhs Node) Node {
	return wrapNode(lhs.Node.LessThan(rhs.Node))
}

func (lhs Node) LessThanOrEqual(rhs Node) Node {
	return wrapNode(lhs.Node.LessThanOrEqual(rhs.Node))
}

func (lhs Node) GreaterThan(rhs Node) Node {
	return wrapNode(lhs.Node.GreaterThan(rhs.Node))
}

func (lhs Node) GreaterThanOrEqual(rhs Node) Node {
	return wrapNode(lhs.Node.GreaterThanOrEqual(rhs.Node))
}

func (lhs Node) DecimalLessThan(rhs Node) Node {
	return wrapNode(lhs.Node.DecimalLessThan(rhs.Node))
}

func (lhs Node) DecimalLessThanOrEqual(rhs Node) Node {
	return wrapNode(lhs.Node.DecimalLessThanOrEqual(rhs.Node))
}

func (lhs Node) DecimalGreaterThan(rhs Node) Node {
	return wrapNode(lhs.Node.DecimalGreaterThan(rhs.Node))
}

func (lhs Node) DecimalGreaterThanOrEqual(rhs Node) Node {
	return wrapNode(lhs.Node.DecimalGreaterThanOrEqual(rhs.Node))
}

func (lhs Node) Like(pattern types.Pattern) Node {
	return wrapNode(lhs.Node.Like(pattern))
}

//  _                _           _
// | |    ___   __ _(_) ___ __ _| |
// | |   / _ \ / _` | |/ __/ _` | |
// | |__| (_) | (_| | | (_| (_| | |
// |_____\___/ \__, |_|\___\__,_|_|
//             |___/

func (lhs Node) And(rhs Node) Node {
	return wrapNode(lhs.Node.And(rhs.Node))
}

func (lhs Node) Or(rhs Node) Node {
	return wrapNode(lhs.Node.Or(rhs.Node))
}

func Not(rhs Node) Node {
	return wrapNode(ast.Not(rhs.Node))
}

func IfThenElse(condition Node, thenNode Node, elseNode Node) Node {
	return wrapNode(ast.IfThenElse(condition.Node, thenNode.Node, elseNode.Node))
}

//     _         _ _   _                    _   _
//    / \   _ __(_) |_| |__  _ __ ___   ___| |_(_) ___
//   / _ \ | '__| | __| '_ \| '_ ` _ \ / _ \ __| |/ __|
//  / ___ \| |  | | |_| | | | | | | | |  __/ |_| | (__
// /_/   \_\_|  |_|\__|_| |_|_| |_| |_|\___|\__|_|\___|

func (lhs Node) Add(rhs Node) Node {
	return wrapNode(lhs.Node.Add(rhs.Node))
}

func (lhs Node) Subtract(rhs Node) Node {
	return wrapNode(lhs.Node.Subtract(rhs.Node))
}

func (lhs Node) Multiply(rhs Node) Node {
	return wrapNode(lhs.Node.Multiply(rhs.Node))
}

func Negate(rhs Node) Node {
	return wrapNode(ast.Negate(rhs.Node))
}

//  _   _ _                         _
// | | | (_) ___ _ __ __ _ _ __ ___| |__  _   _
// | |_| | |/ _ \ '__/ _` | '__/ __| '_ \| | | |
// |  _  | |  __/ | | (_| | | | (__| | | | |_| |
// |_| |_|_|\___|_|  \__,_|_|  \___|_| |_|\__, |
//                                        |___/

func (lhs Node) In(rhs Node) Node {
	return wrapNode(lhs.Node.In(rhs.Node))
}

func (lhs Node) Is(entityType types.EntityType) Node {
	return wrapNode(lhs.Node.Is(entityType))
}

func (lhs Node) IsIn(entityType types.EntityType, rhs Node) Node {
	return wrapNode(lhs.Node.IsIn(entityType, rhs.Node))
}

func (lhs Node) Contains(rhs Node) Node {
	return wrapNode(lhs.Node.Contains(rhs.Node))
}

func (lhs Node) ContainsAll(rhs Node) Node {
	return wrapNode(lhs.Node.ContainsAll(rhs.Node))
}

func (lhs Node) ContainsAny(rhs Node) Node {
	return wrapNode(lhs.Node.ContainsAny(rhs.Node))
}

func (lhs Node) Access(attr types.String) Node {
	return wrapNode(lhs.Node.Access(attr))
}

func (lhs Node) Has(attr types.String) Node {
	return wrapNode(lhs.Node.Has(attr))
}

//  ___ ____   _       _     _
// |_ _|  _ \ / \   __| | __| |_ __ ___  ___ ___
//  | || |_) / _ \ / _` |/ _` | '__/ _ \/ __/ __|
//  | ||  __/ ___ \ (_| | (_| | | |  __/\__ \__ \
// |___|_| /_/   \_\__,_|\__,_|_|  \___||___/___/

func (lhs Node) IsIpv4() Node {
	return wrapNode(lhs.Node.IsIpv4())
}

func (lhs Node) IsIpv6() Node {
	return wrapNode(lhs.Node.IsIpv6())
}

func (lhs Node) IsMulticast() Node {
	return wrapNode(lhs.Node.IsMulticast())
}

func (lhs Node) IsLoopback() Node {
	return wrapNode(lhs.Node.IsLoopback())
}

func (lhs Node) IsInRange(rhs Node) Node {
	return wrapNode(lhs.Node.IsInRange(rhs.Node))
}
