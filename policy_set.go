// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"fmt"

	internalast "github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/parser"
)

type PolicyID string

// A set of named policies against which a request can be authorized.
type PolicySet struct {
	policies map[PolicyID]*Policy
}

// Create a new, empty PolicySet
func NewPolicySet() PolicySet {
	return PolicySet{policies: map[PolicyID]*Policy{}}
}

// NewPolicySetFromFile will create a PolicySet from the given text document with the/ given file name used in Position
// data.  If there is an error parsing the document, it will be returned.
//
// NewPolicySetFromFile assigns default PolicyIDs to the policies contained in fileName.
func NewPolicySetFromFile(fileName string, document []byte) (PolicySet, error) {
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
			Annotations: newAnnotationsFromSlice(p.Annotations),
			Effect:      Effect(p.Effect),
			eval:        eval.Compile((*internalast.Policy)(p)),
			ast:         (*internalast.Policy)(p),
		}
	}
	return PolicySet{policies: policyMap}, nil
}

// GetPolicy returns a pointer to the Policy with the given ID. If a policy with the given ID does not exist, nil is
// returned.
func (p PolicySet) GetPolicy(policyID PolicyID) *Policy {
	return p.policies[policyID]
}

// UpsertPolicy inserts or updates a policy with the given ID.
func (p *PolicySet) UpsertPolicy(policyID PolicyID, policy *Policy) {
	p.policies[policyID] = policy
}

// DeletePolicy removes a policy from the PolicySet. Deleting a non-existent policy is a no-op.
func (p *PolicySet) DeletePolicy(policyID PolicyID) {
	delete(p.policies, policyID)
}
