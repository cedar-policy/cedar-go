// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/parser"
)

type PolicyID string

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
	for i, p := range res {
		policyID := PolicyID(fmt.Sprintf("policy%d", i))
		policyMap[policyID] = &Policy{
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

// NewPolicySetFromPolicies will create a PolicySet from a slice of existing Policys. This constructor can be used to
// support the creation of a PolicySet from JSON-encoded policies or policies created via the ast package, like so:
//
//	policy0 := NewPolicyFromAST(ast.Forbid())
//
//	var policy1 Policy
//	_ = policy1.UnmarshalJSON(
//		[]byte(`{"effect":"permit","principal":{"op":"All"},"action":{"op":"All"},"resource":{"op":"All"}}`),
//	))
//
//	ps := NewPolicySetFromPolicies([]*Policy{policy0, &policy1})
func NewPolicySetFromPolicies(policies []*Policy) PolicySet {
	policyMap := make(map[PolicyID]*Policy, len(policies))
	for i, p := range policies {
		policyID := PolicyID(fmt.Sprintf("policy%d", i))
		policyMap[policyID] = p
	}
	return PolicySet{policies: policyMap}
}

// GetPolicy returns a pointer to the Policy with the given ID. If a policy with the given ID does not exist, nil is
// returned.
func (p PolicySet) GetPolicy(policyID PolicyID) *Policy {
	return p.policies[policyID]
}
