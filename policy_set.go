// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"bytes"
	"fmt"
	"slices"
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
// NewPolicySetFromFile assigns default PolicyIDs to the policies contained in fileName in the format "policy<n>" where
// <n> is incremented for each new policy found in the file.
func NewPolicySetFromFile(fileName string, document []byte) (PolicySet, error) {
	var policySlice PolicySlice
	if err := policySlice.UnmarshalCedar(document); err != nil {
		return PolicySet{}, err
	}

	policyMap := make(map[PolicyID]*Policy, len(policySlice))
	for i, p := range policySlice {
		policyID := PolicyID(fmt.Sprintf("policy%d", i))
		p.SetSourceFile(fileName)
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

// MarshalCedar emits a concatenated Cedar representation of a PolicySet. The policy names are stripped, but policies
// are emitted in lexicographical order by ID.
func (p PolicySet) MarshalCedar(buf *bytes.Buffer) {
	ids := make([]PolicyID, 0, len(p.policies))
	for k := range p.policies {
		ids = append(ids, k)
	}
	slices.Sort(ids)

	i := 0
	for _, id := range ids {
		policy := p.policies[id]
		policy.MarshalCedar(buf)

		if i < len(p.policies)-1 {
			buf.WriteString("\n\n")
		}
		i++
	}
}
