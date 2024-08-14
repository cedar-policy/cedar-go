// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/parser"
)

type PolicyID parser.PolicyID

// A PolicySet is a slice of policies.
type PolicySet struct {
	policies map[PolicyID]*Policy
}

// NewPolicySet will create a PolicySet from the given text document with the
// given file name used in Position data.  If there is an error parsing the
// document, it will be returned.
func NewPolicySet(fileName string, document []byte) (PolicySet, error) {
	var res parser.PolicySet
	if err := res.UnmarshalCedar(document); err != nil {
		return PolicySet{}, fmt.Errorf("parser error: %w", err)
	}
	policyMap := make(map[PolicyID]*Policy, len(res))
	for name, p := range res {
		policyMap[PolicyID(name)] = &Policy{
			Position: Position{
				Filename: fileName,
				Offset:   p.Position.Offset,
				Line:     p.Position.Line,
				Column:   p.Position.Column,
			},
			Annotations: newAnnotationsFromSlice(p.Policy.Annotations),
			Effect:      Effect(p.Policy.Effect),
			eval:        eval.Compile(p.Policy.Policy),
			ast:         &p.Policy.Policy,
		}
	}
	return PolicySet{policies: policyMap}, nil
}

// GetPolicy returns a pointer to the Policy with the given ID. If a policy with the given ID does not exist, nil is
// returned.
func (p PolicySet) GetPolicy(policyID PolicyID) *Policy {
	return p.policies[policyID]
}
