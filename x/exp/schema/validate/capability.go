package validate

import (
	"maps"

	"github.com/cedar-policy/cedar-go/types"
)

// capability represents a known-safe attribute access after a `has` guard.
type capability struct {
	varName types.String // variable or expression identity
	attr    types.String // attribute name
}

// capabilitySet tracks which attributes are safe to access.
type capabilitySet map[capability]bool

func newCapabilitySet() capabilitySet {
	return make(capabilitySet)
}

func (cs capabilitySet) clone() capabilitySet {
	return maps.Clone(cs)
}

func (cs capabilitySet) add(c capability) capabilitySet {
	out := cs.clone()
	out[c] = true
	return out
}

func (cs capabilitySet) has(c capability) bool {
	return cs[c]
}

// merge returns a new set containing all capabilities from both sets.
func (cs capabilitySet) merge(other capabilitySet) capabilitySet {
	out := cs.clone()
	maps.Copy(out, other)
	return out
}

// intersect returns capabilities present in both sets.
func (cs capabilitySet) intersect(other capabilitySet) capabilitySet {
	out := make(capabilitySet)
	for k := range cs {
		if other[k] {
			out[k] = true
		}
	}
	return out
}
