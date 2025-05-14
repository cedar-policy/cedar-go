// Package templates provides an implementation of the Cedar language authorizer.
package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"iter"
	"maps"
	"slices"

	internaljson "github.com/cedar-policy/cedar-go/internal/json"
	"github.com/cedar-policy/cedar-go/types"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

//revive:disable-next-line:exported
type PolicyID = types.PolicyID

// PolicyMap is a map of policy IDs to policy
type PolicyMap map[PolicyID]*Policy

// All returns an iterator over the policy IDs and policies in the PolicyMap.
func (p PolicyMap) All() iter.Seq2[PolicyID, *Policy] {
	return maps.All(p)
}

// PolicySet is a set of named policies against which a request can be authorized.
type PolicySet struct {
	// policies are stored internally so we can handle performance, concurrency bookkeeping however we want
	policies PolicyMap

	templates map[PolicyID]*Template
	links     map[PolicyID]*LinkedPolicy2
}

// NewPolicySet creates a new, empty PolicySet
func NewPolicySet() *PolicySet {
	return &PolicySet{
		policies:  PolicyMap{},
		templates: make(map[PolicyID]*Template),
		links:     make(map[PolicyID]*LinkedPolicy2),
	}
}

// NewPolicySetFromBytes will create a PolicySet from the given text document with the given file name used in Position
// data.  If there is an error parsing the document, it will be returned.
//
// NewPolicySetFromBytes assigns default PolicyIDs to the policies contained in fileName in the format "policy<n>" where
// <n> is incremented for each new policy found in the file.
func NewPolicySetFromBytes(fileName string, document []byte) (*PolicySet, error) {
	policySlice, err := NewPolicyListFromBytes(fileName, document)
	if err != nil {
		return &PolicySet{}, err
	}
	policyMap := make(PolicyMap, len(policySlice.StaticPolicies))
	for i, p := range policySlice.StaticPolicies {
		policyID := PolicyID(fmt.Sprintf("policy%d", i))
		policyMap[policyID] = p
	}

	templateMap := make(map[PolicyID]*Template, len(policySlice.Templates))
	for i, p := range policySlice.Templates {
		policyID := PolicyID(fmt.Sprintf("template%d", i))
		templateMap[policyID] = p
	}

	return &PolicySet{policies: policyMap, templates: templateMap, links: make(map[PolicyID]*LinkedPolicy2)}, nil
}

// Get returns the Policy with the given ID. If a policy with the given ID
// does not exist, nil is returned.
func (p *PolicySet) Get(policyID PolicyID) *Policy {
	return p.policies[policyID]
}

// Add inserts or updates a policy with the given ID. Returns true if a policy
// with the given ID did not already exist in the set.
func (p *PolicySet) Add(policyID PolicyID, policy *Policy) bool {
	_, exists := p.policies[policyID]
	p.policies[policyID] = policy
	return !exists
}

// Remove removes a policy from the PolicySet. Returns true if a policy with
// the given ID already existed in the set.
func (p *PolicySet) Remove(policyID PolicyID) bool {
	_, exists := p.policies[policyID]
	delete(p.policies, policyID)
	return exists
}

// Map returns a new PolicyMap instance of the policies in the PolicySet.
//
// Deprecated: use the iterator returned by All() like so: maps.Collect(ps.All())
func (p *PolicySet) Map() PolicyMap {
	return maps.Clone(p.policies)
}

// MarshalCedar emits a concatenated Cedar representation of a PolicySet. The policy names are stripped, but policies
// are emitted in lexicographical order by ID.
func (p *PolicySet) MarshalCedar() []byte {
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

// MarshalJSON encodes a PolicySet in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
func (p *PolicySet) MarshalJSON() ([]byte, error) {
	jsonPolicySet := internaljson.PolicySetJSON{
		StaticPolicies: make(internaljson.PolicySet, len(p.policies)),
	}
	for k, v := range p.policies {
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
		policies: make(PolicyMap, len(jsonPolicySet.StaticPolicies)),
	}
	for k, v := range jsonPolicySet.StaticPolicies {
		p.policies[PolicyID(k)] = newPolicy((*internalast.Policy)(v))
	}
	return nil
}

// IsAuthorized uses the combination of the PolicySet and Entities to determine
// if the given Request to determine Decision and Diagnostic.
//
// Deprecated: Use the Authorize() function instead
func (p *PolicySet) IsAuthorized(entities types.EntityGetter, req cedar.Request) (cedar.Decision, cedar.Diagnostic) {
	return Authorize(p, entities, req)
}

// All returns an iterator over the (PolicyID, *Policy) tuples in the PolicySet
func (p *PolicySet) All() iter.Seq2[PolicyID, *Policy] {
	return func(yield func(PolicyID, *Policy) bool) {
		for k, v := range p.policies {
			if !yield(k, v) {
				break
			}
		}

		for k, v := range p.links {
			// Render links on read to make template changes propagate
			policy, err := p.render(*v)
			if err != nil { //todo: think how to propagate this error
				continue
			}

			if !yield(k, policy) {
				break
			}
		}
	}
}

func (p *PolicySet) render(link LinkedPolicy2) (*Policy, error) {
	template := p.GetTemplate(link.templateID)
	if template == nil {
		return nil, fmt.Errorf("no such template %q", link.templateID)
	}

	pTemplate := parser.Template(*template)

	policy, err := parser.RenderLinkedPolicy(&pTemplate, link.slotEnv)
	if err != nil {
		return nil, err
	}

	internalPolicy := internalast.Policy(policy)

	return newPolicy(&internalPolicy), nil
}
