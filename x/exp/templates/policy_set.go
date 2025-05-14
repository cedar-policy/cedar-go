// Package templates provides an implementation of the Cedar language authorizer.
package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"iter"
	"maps"
	"slices"

	internaljson "github.com/cedar-policy/cedar-go/internal/json"
)

// PolicySet is a set of named policies against which a request can be authorized.
type PolicySet struct {
	// policies are stored internally so we can handle performance, concurrency bookkeeping however we want
	policies cedar.PolicyMap

	templates map[cedar.PolicyID]*Template
	links     map[cedar.PolicyID]*LinkedPolicy
}

// NewPolicySet creates a new, empty PolicySet
func NewPolicySet() *PolicySet {
	return &PolicySet{
		policies:  cedar.PolicyMap{},
		templates: make(map[cedar.PolicyID]*Template),
		links:     make(map[cedar.PolicyID]*LinkedPolicy),
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
	policyMap := make(cedar.PolicyMap, len(policySlice.StaticPolicies))
	for i, p := range policySlice.StaticPolicies {
		policyID := cedar.PolicyID(fmt.Sprintf("policy%d", i))
		policyMap[policyID] = p
	}

	templateMap := make(map[cedar.PolicyID]*Template, len(policySlice.Templates))
	for i, p := range policySlice.Templates {
		policyID := cedar.PolicyID(fmt.Sprintf("template%d", i))
		templateMap[policyID] = p
	}

	return &PolicySet{policies: policyMap, templates: templateMap, links: make(map[cedar.PolicyID]*LinkedPolicy)}, nil
}

// Get returns the Policy with the given ID. If a policy with the given ID
// does not exist, nil is returned.
func (p *PolicySet) Get(policyID cedar.PolicyID) *cedar.Policy {
	return p.policies[policyID]
}

// Add inserts or updates a policy with the given ID. Returns true if a policy
// with the given ID did not already exist in the set.
func (p *PolicySet) Add(policyID cedar.PolicyID, policy *cedar.Policy) bool {
	_, exists := p.policies[policyID]
	p.policies[policyID] = policy
	return !exists
}

// Remove removes a policy from the PolicySet. Returns true if a policy with
// the given ID already existed in the set.
func (p *PolicySet) Remove(policyID cedar.PolicyID) bool {
	_, exists := p.policies[policyID]
	delete(p.policies, policyID)
	return exists
}

// Map returns a new PolicyMap instance of the policies in the PolicySet.
//
// Deprecated: use the iterator returned by All() like so: maps.Collect(ps.All())
func (p *PolicySet) Map() cedar.PolicyMap {
	return maps.Clone(p.policies)
}

// MarshalCedar emits a concatenated Cedar representation of a PolicySet. The policy names are stripped, but policies
// are emitted in lexicographical order by ID.
func (p *PolicySet) MarshalCedar() []byte {
	ids := make([]cedar.PolicyID, 0, len(p.policies))
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
		jsonPolicySet.StaticPolicies[string(k)] = (*internaljson.Policy)(v.AST())
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
		policies: make(cedar.PolicyMap, len(jsonPolicySet.StaticPolicies)),
	}
	for k, v := range jsonPolicySet.StaticPolicies {
		p.policies[cedar.PolicyID(k)] = cedar.NewPolicyFromAST((*ast.Policy)(v))
	}
	return nil
}

// All returns an iterator over the (PolicyID, *Policy) tuples in the PolicySet
func (p *PolicySet) All() iter.Seq2[cedar.PolicyID, *cedar.Policy] {
	return func(yield func(cedar.PolicyID, *cedar.Policy) bool) {
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

func (p *PolicySet) render(link LinkedPolicy) (*cedar.Policy, error) {
	template := p.GetTemplate(link.templateID)
	if template == nil {
		return nil, fmt.Errorf("no such template %q", link.templateID)
	}

	pTemplate := parser.Template(*template)

	policy, err := parser.RenderLinkedPolicy(&pTemplate, link.slotEnv)
	if err != nil {
		return nil, err
	}

	astPolicy := ast.Policy(policy)

	return cedar.NewPolicyFromAST(&astPolicy), nil
}
