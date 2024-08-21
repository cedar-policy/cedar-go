// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
)

// PolicyID is a string identifier for the policy within the PolicySet
type PolicyID string

// PolicyMap is a map of policy IDs to policy
type PolicyMap map[PolicyID]Policy

// PolicySet is a set of named policies against which a request can be authorized.
type PolicySet struct {
	// policies are stored internally so we can handle performance, concurrency bookkeeping however we want
	policies PolicyMap
}

// NewPolicySet creates a new, empty PolicySet
func NewPolicySet() PolicySet {
	return PolicySet{policies: PolicyMap{}}
}

// NewPolicySetFromBytes will create a PolicySet from the given text document with the given file name used in Position
// data.  If there is an error parsing the document, it will be returned.
//
// NewPolicySetFromBytes assigns default PolicyIDs to the policies contained in fileName in the format "policy<n>" where
// <n> is incremented for each new policy found in the file.
func NewPolicySetFromBytes(fileName string, document []byte) (PolicySet, error) {
	policySlice, err := NewPolicyListFromBytes(fileName, document)
	if err != nil {
		return PolicySet{}, err
	}
	policyMap := make(PolicyMap, len(policySlice))
	for i, p := range policySlice {
		policyID := PolicyID(fmt.Sprintf("policy%d", i))
		policyMap[policyID] = p
	}
	return PolicySet{policies: policyMap}, nil
}

// Get returns the Policy with the given ID. If a policy with the given ID does not exist, an empty policy is returned.
func (p PolicySet) Get(policyID PolicyID) Policy {
	return p.policies[policyID]
}

// Set inserts or updates a policy with the given ID.
func (p *PolicySet) Set(policyID PolicyID, policy Policy) {
	p.policies[policyID] = policy
}

// Delete removes a policy from the PolicySet. Deleting a non-existent policy is a no-op.
func (p *PolicySet) Delete(policyID PolicyID) {
	delete(p.policies, policyID)
}

// Map returns a new PolicyMap instance of the policies in the PolicySet.
func (p *PolicySet) Map() PolicyMap {
	return maps.Clone(p.policies)
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
