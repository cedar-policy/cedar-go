package ast

import (
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

//   ____                                 _
//  / ___|___  _ __ ___  _ __   __ _ _ __(_)___  ___  _ __
// | |   / _ \| '_ ` _ \| '_ \ / _` | '__| / __|/ _ \| '_ \
// | |__| (_) | | | | | | |_) | (_| | |  | \__ \ (_) | | | |
//  \____\___/|_| |_| |_| .__/ \__,_|_|  |_|___/\___/|_| |_|
//                      |_|

// Equal builds an AST node representing the = operator
func (lhs Node) Equal(rhs Node) Node {
	return wrapNode(lhs.Node.Equal(rhs.Node))
}

// NotEqual builds an AST node representing the != operator
func (lhs Node) NotEqual(rhs Node) Node {
	return wrapNode(lhs.Node.NotEqual(rhs.Node))
}

// LessThan builds an AST node representing the < operator
func (lhs Node) LessThan(rhs Node) Node {
	return wrapNode(lhs.Node.LessThan(rhs.Node))
}

// LessThanOrEqual builds an AST node representing the <= operator
func (lhs Node) LessThanOrEqual(rhs Node) Node {
	return wrapNode(lhs.Node.LessThanOrEqual(rhs.Node))
}

// GreaterThan builds an AST node representing the > operator
func (lhs Node) GreaterThan(rhs Node) Node {
	return wrapNode(lhs.Node.GreaterThan(rhs.Node))
}

// GreaterThanOrEqual builds an AST node representing the >= operator
func (lhs Node) GreaterThanOrEqual(rhs Node) Node {
	return wrapNode(lhs.Node.GreaterThanOrEqual(rhs.Node))
}

// DecimalLessThan builds an AST node representing the .lessThan() operator
func (lhs Node) DecimalLessThan(rhs Node) Node {
	return wrapNode(lhs.Node.DecimalLessThan(rhs.Node))
}

// DecimalLessThanOrEqual builds an AST node representing the .lessThanOrEqual() operator
func (lhs Node) DecimalLessThanOrEqual(rhs Node) Node {
	return wrapNode(lhs.Node.DecimalLessThanOrEqual(rhs.Node))
}

// DecimalGreaterThan builds an AST node representing the .greaterThan() operator
func (lhs Node) DecimalGreaterThan(rhs Node) Node {
	return wrapNode(lhs.Node.DecimalGreaterThan(rhs.Node))
}

// DecimalGreaterThanOrEqual builds an AST node representing the .greaterThanOrEqual() operator
func (lhs Node) DecimalGreaterThanOrEqual(rhs Node) Node {
	return wrapNode(lhs.Node.DecimalGreaterThanOrEqual(rhs.Node))
}

// Like builds an AST node representing the like operator
func (lhs Node) Like(pattern types.Pattern) Node {
	return wrapNode(lhs.Node.Like(pattern))
}

//  _                _           _
// | |    ___   __ _(_) ___ __ _| |
// | |   / _ \ / _` | |/ __/ _` | |
// | |__| (_) | (_| | | (_| (_| | |
// |_____\___/ \__, |_|\___\__,_|_|
//             |___/

// And builds an AST node representing the && operator
func (lhs Node) And(rhs Node) Node {
	return wrapNode(lhs.Node.And(rhs.Node))
}

// Or builds an AST node representing the || operator
func (lhs Node) Or(rhs Node) Node {
	return wrapNode(lhs.Node.Or(rhs.Node))
}

// Not builds an AST node representing the ! operator
func Not(rhs Node) Node {
	return wrapNode(ast.Not(rhs.Node))
}

// IfThenElse builds an AST node representing the if (CONDITIONAL) operator
func IfThenElse(condition Node, thenNode Node, elseNode Node) Node {
	return wrapNode(ast.IfThenElse(condition.Node, thenNode.Node, elseNode.Node))
}

//     _         _ _   _                    _   _
//    / \   _ __(_) |_| |__  _ __ ___   ___| |_(_) ___
//   / _ \ | '__| | __| '_ \| '_ ` _ \ / _ \ __| |/ __|
//  / ___ \| |  | | |_| | | | | | | | |  __/ |_| | (__
// /_/   \_\_|  |_|\__|_| |_|_| |_| |_|\___|\__|_|\___|

// Add builds an AST node representing the + operator
func (lhs Node) Add(rhs Node) Node {
	return wrapNode(lhs.Node.Add(rhs.Node))
}

// Subtract builds an AST node representing the - operator
func (lhs Node) Subtract(rhs Node) Node {
	return wrapNode(lhs.Node.Subtract(rhs.Node))
}

// Multiply builds an AST node representing the * operator
func (lhs Node) Multiply(rhs Node) Node {
	return wrapNode(lhs.Node.Multiply(rhs.Node))
}

// Negate builds an AST node representing the ! operator
func Negate(rhs Node) Node {
	return wrapNode(ast.Negate(rhs.Node))
}

//  _   _ _                         _
// | | | (_) ___ _ __ __ _ _ __ ___| |__  _   _
// | |_| | |/ _ \ '__/ _` | '__/ __| '_ \| | | |
// |  _  | |  __/ | | (_| | | | (__| | | | |_| |
// |_| |_|_|\___|_|  \__,_|_|  \___|_| |_|\__, |
//                                        |___/

// In builds an AST node representing the in operator
func (lhs Node) In(rhs Node) Node {
	return wrapNode(lhs.Node.In(rhs.Node))
}

// Is builds an AST node representing the is operator
func (lhs Node) Is(entityType types.EntityType) Node {
	return wrapNode(lhs.Node.Is(entityType))
}

// IsIn builds an AST node representing the "is in" operator
func (lhs Node) IsIn(entityType types.EntityType, rhs Node) Node {
	return wrapNode(lhs.Node.IsIn(entityType, rhs.Node))
}

// Contains builds an AST node representing the .contains() operator
func (lhs Node) Contains(rhs Node) Node {
	return wrapNode(lhs.Node.Contains(rhs.Node))
}

// ContainsAll builds an AST node representing the .containsAll() operator
func (lhs Node) ContainsAll(rhs Node) Node {
	return wrapNode(lhs.Node.ContainsAll(rhs.Node))
}

// ContainsAny builds an AST node representing the .containsAny() operator
func (lhs Node) ContainsAny(rhs Node) Node {
	return wrapNode(lhs.Node.ContainsAny(rhs.Node))
}

// IsEmpty builds an AST node representing the .isEmpty() operator
func (lhs Node) IsEmpty() Node { return wrapNode(lhs.Node.IsEmpty()) }

// Access builds an AST node representing the . and [] operators to access entity attributes
func (lhs Node) Access(attr types.String) Node {
	return wrapNode(lhs.Node.Access(attr))
}

// Has builds an AST node representing the has operator
func (lhs Node) Has(attr types.String) Node {
	return wrapNode(lhs.Node.Has(attr))
}

// GetTag builds an AST node representing the .getTag() operator
func (lhs Node) GetTag(rhs Node) Node {
	return wrapNode(lhs.Node.GetTag(rhs.Node))
}

// HasTag builds an AST node representing the .hasTag() operator
func (lhs Node) HasTag(rhs Node) Node {
	return wrapNode(lhs.Node.HasTag(rhs.Node))
}

//  ___ ____   _       _     _
// |_ _|  _ \ / \   __| | __| |_ __ ___  ___ ___
//  | || |_) / _ \ / _` |/ _` | '__/ _ \/ __/ __|
//  | ||  __/ ___ \ (_| | (_| | | |  __/\__ \__ \
// |___|_| /_/   \_\__,_|\__,_|_|  \___||___/___/

// IsIpv4 builds an AST node representing the .isIpv4() operator
func (lhs Node) IsIpv4() Node {
	return wrapNode(lhs.Node.IsIpv4())
}

// IsIpv6 builds an AST node representing the .isIpv6() operator
func (lhs Node) IsIpv6() Node {
	return wrapNode(lhs.Node.IsIpv6())
}

// IsMulticast builds an AST node representing the .isMulticast() operator
func (lhs Node) IsMulticast() Node {
	return wrapNode(lhs.Node.IsMulticast())
}

// IsLoopback builds an AST node representing the .isLoopback() operator
func (lhs Node) IsLoopback() Node {
	return wrapNode(lhs.Node.IsLoopback())
}

// IsInRange builds an AST node representing the .isInRange() operator
func (lhs Node) IsInRange(rhs Node) Node {
	return wrapNode(lhs.Node.IsInRange(rhs.Node))
}

//  ____        _       _   _
// |  _ \  __ _| |_ ___| |_(_)_ __ ___   ___
// | | | |/ _` | __/ _ \ __| | '_ ` _ \ / _ \
// | |_| | (_| | ||  __/ |_| | | | | | |  __/
// |____/ \__,_|\__\___|\__|_|_| |_| |_|\___|

// Offset builds an AST node representing the .offset() operator
func (lhs Node) Offset(rhs Node) Node { return wrapNode(lhs.Node.Offset(rhs.Node)) }

// DurationSince builds an AST node representing the .durationSince() operator
func (lhs Node) DurationSince(rhs Node) Node { return wrapNode(lhs.Node.DurationSince(rhs.Node)) }

// ToDate builds an AST node representing the .toDate() operator
func (lhs Node) ToDate() Node { return wrapNode(lhs.Node.ToDate()) }

// ToTime builds an AST node representing the .toTime() operator
func (lhs Node) ToTime() Node { return wrapNode(lhs.Node.ToTime()) }

//  ____                  _   _
// |  _ \ _   _ _ __ __ _| |_(_) ___  _ __
// | | | | | | | '__/ _` | __| |/ _ \| '_ \
// | |_| | |_| | | | (_| | |_| | (_) | | | |
// |____/ \__,_|_|  \__,_|\__|_|\___/|_| |_|

// ToDays builds an AST node representing the .toDays() operator
func (lhs Node) ToDays() Node { return wrapNode(lhs.Node.ToDays()) }

// ToHours builds an AST node representing the .toHours() operator
func (lhs Node) ToHours() Node { return wrapNode(lhs.Node.ToHours()) }

// ToMinutes builds an AST node representing the .toMinutes() operator
func (lhs Node) ToMinutes() Node { return wrapNode(lhs.Node.ToMinutes()) }

// ToSeconds builds an AST node representing the .toSeconds() operator
func (lhs Node) ToSeconds() Node { return wrapNode(lhs.Node.ToSeconds()) }

// ToMilliseconds builds an AST node representing the .toMilliseconds() operator
func (lhs Node) ToMilliseconds() Node { return wrapNode(lhs.Node.ToMilliseconds()) }
