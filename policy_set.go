// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"bytes"
	"fmt"
	"slices"
)

type PolicyID string

// PolicySet is a set of named policies against which a request can be authorized.
type PolicySet struct {
	policies map[PolicyID]*Policy
}

// NewPolicySet creates a new, empty PolicySet
func NewPolicySet() PolicySet {
	return PolicySet{policies: map[PolicyID]*Policy{}}
}

// NewPolicySetFromBytes will create a PolicySet from the given text document with the/ given file name used in Position
// data.  If there is an error parsing the document, it will be returned.
//
// NewPolicySetFromBytes assigns default PolicyIDs to the policies contained in fileName in the format "policy<n>" where
// <n> is incremented for each new policy found in the file.
func NewPolicySetFromBytes(fileName string, document []byte) (PolicySet, error) {
	policySlice, err := NewPolicySliceFromBytes(fileName, document)
	if err != nil {
		return PolicySet{}, err
	}
	policyMap := make(map[PolicyID]*Policy, len(policySlice))
	for i, p := range policySlice {
		policyID := PolicyID(fmt.Sprintf("policy%d", i))
		policyMap[policyID] = p
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

// UpsertPolicySet inserts or updates all the policies from src into this PolicySet. Policies in this PolicySet with
// identical IDs in src are clobbered by the policies from src.
func (p *PolicySet) UpsertPolicySet(src PolicySet) {
	for id, policy := range src.policies {
		p.policies[id] = policy
	}
}

// MarshalCedar emits a concatenated Cedar representation of a PolicySet. The policy names are stripped, but policies
// are emitted in lexicographical order by ID.
func (p PolicySet) MarshalCedar() []byte {
	ids := make([]PolicyID, 0, len(p.policies))
	for k := range p.policies {
		ids = append(ids, k)
	}
	slices.Sort(ids)

	var buf bytes.Buffer
	i := 0
	for _, id := range ids {
		policy := p.policies[id]
		buf.Write(policy.MarshalCedar())

		if i < len(p.policies)-1 {
			buf.WriteString("\n\n")
		}
		i++
	}
	return buf.Bytes()
}
