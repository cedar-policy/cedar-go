// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
    "bytes"
    "encoding/json"
    "fmt"
    "maps"
    "slices"

    internalast "github.com/cedar-policy/cedar-go/internal/ast"
    internaljson "github.com/cedar-policy/cedar-go/internal/json"
    "github.com/cedar-policy/cedar-go/types"
)

type PolicyID = types.PolicyID

// PolicyMap is a map of policy IDs to policy
type PolicyMap struct {
    StaticPolicies map[PolicyID]*Policy
    Templates      map[PolicyID]*Template
}

func makePolicyMap2() PolicyMap {
    return PolicyMap{
        StaticPolicies: make(map[PolicyID]*Policy),
        Templates:      make(map[PolicyID]*Template),
    }
}

// PolicySet is a set of named policies against which a request can be authorized.
type PolicySet struct {
    // policies are stored internally so we can handle performance, concurrency bookkeeping however we want
    policies PolicyMap
}

// NewPolicySet creates a new, empty PolicySet
func NewPolicySet() *PolicySet {
    return &PolicySet{policies: makePolicyMap2()}
}

// NewPolicySetFromBytes will create a PolicySet from the given text document with the given file name used in Position
// data.  If there is an error parsing the document, it will be returned.
//
// NewPolicySetFromBytes assigns default PolicyIDs to the policies contained in fileName in the format "policy<n>" where
// <n> is incremented for each new policy found in the file.
func NewPolicySetFromBytes(fileName string, document []byte) (*PolicySet, error) {
    policySlice, err := NewPolicyListFromBytes(fileName, document)
    if err != nil {
        return nil, err
    }

    pm := PolicyMap{
        StaticPolicies: make(map[PolicyID]*Policy, len(policySlice.StaticPolicies)),
        Templates:      make(map[PolicyID]*Template, len(policySlice.Templates)),
    }

    for i, p := range policySlice.StaticPolicies {
        policyID := PolicyID(fmt.Sprintf("policy%d", i))
        pm.StaticPolicies[policyID] = p
    }

    for i, t := range policySlice.Templates {
        policyID := PolicyID(fmt.Sprintf("template%d", i))
        pm.Templates[policyID] = t
    }

    return &PolicySet{policies: pm}, nil
}

// Get returns the Policy with the given ID. If a policy with the given ID does not exist, nil is returned.
func (p PolicySet) Get(policyID PolicyID) *Policy {
    return p.policies.StaticPolicies[policyID]
}

// Store inserts or updates a policy with the given ID.
func (p *PolicySet) Store(policyID PolicyID, policy *Policy) {
    p.policies.StaticPolicies[policyID] = policy
}

// Delete removes a policy from the PolicySet. Deleting a non-existent policy is a no-op.
func (p *PolicySet) Delete(policyID PolicyID) {
    delete(p.policies.StaticPolicies, policyID)
}

// Map returns a new PolicyMap instance of the policies in the PolicySet.
func (p *PolicySet) Map() PolicyMap {
    return PolicyMap{
        StaticPolicies: maps.Clone(p.policies.StaticPolicies),
        Templates:      maps.Clone(p.policies.Templates),
    }
}

func (p PolicySet) Len() int {
    return len(p.policies.StaticPolicies) + len(p.policies.Templates)
}

// MarshalCedar emits a concatenated Cedar representation of a PolicySet. The policy names are stripped, but policies
// are emitted in lexicographical order by ID.
// todo: add support for Templates
func (p *PolicySet) MarshalCedar() []byte {
    ids := make([]PolicyID, 0, len(p.policies.StaticPolicies))
    for k := range p.policies.StaticPolicies {
        ids = append(ids, k)
    }
    slices.Sort(ids)

    var buf bytes.Buffer
    i := 0
    for _, id := range ids {
        policy := p.policies.StaticPolicies[id]
        buf.Write(policy.MarshalCedar())

        if i < len(p.policies.StaticPolicies)-1 {
            buf.WriteString("\n\n")
        }
        i++
    }
    return buf.Bytes()
}

// MarshalJSON encodes a PolicySet in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
// todo: add support for Templates
func (p *PolicySet) MarshalJSON() ([]byte, error) {
    jsonPolicySet := internaljson.PolicySetJSON{
        StaticPolicies: make(internaljson.PolicySet, len(p.policies.StaticPolicies)),
    }
    for k, v := range p.policies.StaticPolicies {
        jsonPolicySet.StaticPolicies[string(k)] = (*internaljson.Policy)(v.ast)
    }
    return json.Marshal(jsonPolicySet)
}

// UnmarshalJSON parses and compiles a PolicySet in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
func (p *PolicySet) UnmarshalJSON(b []byte) error {
    var jsonPolicySet internaljson.PolicySetJSON
    if err := json.Unmarshal(b, &jsonPolicySet); err != nil {
        return err
    }
    *p = PolicySet{
        policies: makePolicyMap2(),
    }

    for k, v := range jsonPolicySet.StaticPolicies {
        p.policies.StaticPolicies[PolicyID(k)] = newPolicy((*internalast.Policy)(v))
    }

    return nil
}
