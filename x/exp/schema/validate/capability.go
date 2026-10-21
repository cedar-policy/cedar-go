package validate

import (
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// capability records that a value is known to exist on an expression, after
// a `has` or a `hasTag` guard.
type capability struct {
	on  ast.IsNode
	val ast.IsNode
}

func (c capability) equal(other capability) bool {
	return ast.Equal(ast.NewNode(c.on), ast.NewNode(other.on)) &&
		ast.Equal(ast.NewNode(c.val), ast.NewNode(other.val))
}

// capabilitySet tracks which attributes and tags are safe to access. It is
// immutable: the add methods, merge, and intersect return new sets.
type capabilitySet struct {
	attrs []capability
	tags  []capability
}

func newCapabilitySet() capabilitySet {
	return capabilitySet{}
}

func (cs capabilitySet) addAttribute(on ast.IsNode, attr types.String) capabilitySet {
	cs.attrs = addCapability(cs.attrs, capability{on: on, val: ast.NodeValue{Value: attr}})
	return cs
}

func (cs capabilitySet) hasAttribute(on ast.IsNode, attr types.String) bool {
	return slices.ContainsFunc(cs.attrs, capability{on: on, val: ast.NodeValue{Value: attr}}.equal)
}

func (cs capabilitySet) addTag(on, tag ast.IsNode) capabilitySet {
	cs.tags = addCapability(cs.tags, capability{on: on, val: tag})
	return cs
}

func (cs capabilitySet) hasTag(on, tag ast.IsNode) bool {
	return slices.ContainsFunc(cs.tags, capability{on: on, val: tag}.equal)
}

// merge returns a new set containing all capabilities from both sets.
func (cs capabilitySet) merge(other capabilitySet) capabilitySet {
	for _, c := range other.attrs {
		cs.attrs = addCapability(cs.attrs, c)
	}
	for _, c := range other.tags {
		cs.tags = addCapability(cs.tags, c)
	}
	return cs
}

// intersect returns capabilities present in both sets.
func (cs capabilitySet) intersect(other capabilitySet) capabilitySet {
	return capabilitySet{
		attrs: intersectCapabilities(cs.attrs, other.attrs),
		tags:  intersectCapabilities(cs.tags, other.tags),
	}
}

func addCapability(cs []capability, c capability) []capability {
	if slices.ContainsFunc(cs, c.equal) {
		return cs
	}
	// Use Concat() instead of append() to avoid mutating the original slice
	return slices.Concat(cs, []capability{c})
}

func intersectCapabilities(cs, other []capability) []capability {
	var out []capability
	for _, c := range cs {
		if slices.ContainsFunc(other, c.equal) {
			out = append(out, c)
		}
	}
	return out
}
