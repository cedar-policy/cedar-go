// Package templates provides an implementation of the Cedar language authorizer.
package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/types"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
	"iter"
	"maps"
	"slices"

	internaljson "github.com/cedar-policy/cedar-go/internal/json"
)

type PolicyMap map[cedar.PolicyID]*Policy

// All returns an iterator over the policy IDs and policies in the PolicyMap.
func (p PolicyMap) All() iter.Seq2[cedar.PolicyID, *Policy] {
	return maps.All(p)
}

// PolicySet is a set of named policies against which a request can be authorized.
type PolicySet struct {
	// policies are stored internally so we can handle performance, concurrency bookkeeping however we want
	staticPolicies PolicyMap
	linkedPolicies map[cedar.PolicyID]*LinkedPolicy

	templates map[cedar.PolicyID]*Template
}

// NewPolicySet creates a new, empty PolicySet
func NewPolicySet() *PolicySet {
	return &PolicySet{
		staticPolicies: PolicyMap{},
		templates:      make(map[cedar.PolicyID]*Template),
		linkedPolicies: make(map[cedar.PolicyID]*LinkedPolicy),
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
		policyID := cedar.PolicyID(fmt.Sprintf("policy%d", i))
		policyMap[policyID] = p
	}

	templateMap := make(map[cedar.PolicyID]*Template, len(policySlice.Templates))
	for i, p := range policySlice.Templates {
		policyID := cedar.PolicyID(fmt.Sprintf("template%d", i))
		templateMap[policyID] = p
	}

	return &PolicySet{staticPolicies: policyMap, templates: templateMap, linkedPolicies: make(map[cedar.PolicyID]*LinkedPolicy)}, nil
}

// Get returns the Policy with the given ID. If a policy with the given ID
// does not exist, nil is returned.
func (p *PolicySet) Get(policyID cedar.PolicyID) *Policy {
	return p.staticPolicies[policyID]
}

// Add inserts or updates a policy with the given ID. Returns true if a policy
// with the given ID did not already exist in the set.
func (p *PolicySet) Add(policyID cedar.PolicyID, policy *Policy) bool {
	_, exists := p.staticPolicies[policyID]
	p.staticPolicies[policyID] = policy
	return !exists
}

// Remove removes a policy from the PolicySet. Returns true if a policy with
// the given ID already existed in the set.
func (p *PolicySet) Remove(policyID cedar.PolicyID) bool {
	_, staticExists := p.staticPolicies[policyID]
	delete(p.staticPolicies, policyID)

	_, linkExists := p.linkedPolicies[policyID]
	delete(p.linkedPolicies, policyID)

	return staticExists || linkExists
}

// Map returns a new PolicyMap instance of the policies in the PolicySet.
//
// Deprecated: use the iterator returned by All() like so: maps.Collect(ps.All())
func (p *PolicySet) Map() PolicyMap {
	return maps.Clone(p.staticPolicies)
}

// MarshalCedar emits a concatenated Cedar representation of a PolicySet. The policy names are stripped, but policies
// are emitted in lexicographical order by ID.
func (p *PolicySet) MarshalCedar() []byte {
	ids := make([]cedar.PolicyID, 0, len(p.staticPolicies))
	for k := range p.staticPolicies {
		ids = append(ids, k)
	}
	slices.Sort(ids)

	var buf bytes.Buffer
	i := 0
	for _, id := range ids {
		policy := p.staticPolicies[id]
		buf.Write(policy.MarshalCedar())

		if i < len(p.staticPolicies)-1 {
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
		StaticPolicies: make(internaljson.PolicySet, len(p.staticPolicies)),
		Templates:      make(internaljson.TemplateSet, len(p.templates)),
		TemplateLinks:  make([]internaljson.LinkedPolicy, 0, len(p.linkedPolicies)),
	}
	for k, v := range p.staticPolicies {
		jsonPolicySet.StaticPolicies[string(k)] = (*internaljson.Policy)(v.AST())
	}
	for k, v := range p.templates {
		jsonPolicySet.Templates[string(k)] = (*internaljson.Policy)(v.AST())
	}
	for _, v := range p.linkedPolicies {
		lp := internaljson.LinkedPolicy{
			TemplateID: string(v.templateID),
			LinkID:     string(v.linkID),
			Values:     make(map[string]types.ImplicitlyMarshaledEntityUID, len(v.slotEnv)),
		}

		for slotID, entityUID := range v.slotEnv {
			lp.Values[string(slotID)] = types.ImplicitlyMarshaledEntityUID(entityUID)
		}

		jsonPolicySet.TemplateLinks = append(jsonPolicySet.TemplateLinks, lp)
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
		staticPolicies: make(PolicyMap, len(jsonPolicySet.StaticPolicies)),
		templates:      make(map[cedar.PolicyID]*Template, len(jsonPolicySet.Templates)),
		linkedPolicies: make(map[cedar.PolicyID]*LinkedPolicy),
	}
	for k, v := range jsonPolicySet.StaticPolicies {
		p.staticPolicies[cedar.PolicyID(k)] = newPolicy((*internalast.Policy)(v))
	}
	for k, v := range jsonPolicySet.Templates {
		p.templates[cedar.PolicyID(k)] = newTemplate((*internalast.Policy)(v))
	}
	for _, v := range jsonPolicySet.TemplateLinks {
		lp := &LinkedPolicy{
			templateID: cedar.PolicyID(v.TemplateID),
			linkID:     cedar.PolicyID(v.LinkID),
			slotEnv:    make(map[types.SlotID]types.EntityUID, len(v.Values)),
		}

		for slotID, entityUID := range v.Values {
			slotIDTyped := types.SlotID(slotID)
			entityUIDTyped := types.EntityUID(entityUID)

			lp.slotEnv[slotIDTyped] = entityUIDTyped
		}

		p.linkedPolicies[cedar.PolicyID(v.LinkID)] = lp
	}

	return nil
}

// All returns an iterator over the (PolicyID, *Policy) tuples in the PolicySet
func (p *PolicySet) All() iter.Seq2[cedar.PolicyID, *Policy] {
	return func(yield func(cedar.PolicyID, *Policy) bool) {
		for k, v := range p.staticPolicies {
			if !yield(k, v) {
				break
			}
		}

		for k, v := range p.linkedPolicies {
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

func (p *PolicySet) render(link LinkedPolicy) (*Policy, error) {
	template := p.GetTemplate(link.templateID)
	if template == nil {
		return nil, fmt.Errorf("no such template %q", link.templateID)
	}

	pTemplate := parser.Template(*template)

	policy, err := parser.RenderLinkedPolicy(&pTemplate, link.slotEnv)
	if err != nil {
		return nil, err
	}

	astPolicy := internalast.Policy(policy)

	return newPolicy(&astPolicy), nil
}

// LinkedPolicy represents a template that has been linked with specific slot values.
// It's a wrapper around the internal parser.LinkedPolicy type.
//type LinkedPolicy parser.LinkedPolicy

type LinkedPolicy struct {
	templateID cedar.PolicyID
	linkID     cedar.PolicyID
	slotEnv    map[types.SlotID]types.EntityUID
}

// TemplateID returns the PolicyID of the template associated with this LinkedPolicy.
func (l *LinkedPolicy) TemplateID() cedar.PolicyID {
	return l.templateID
}

// LinkID returns the PolicyID of this LinkedPolicy.
func (l *LinkedPolicy) LinkID() cedar.PolicyID {
	return l.linkID
}

// LinkTemplate creates a LinkedPolicy by binding slot values to a template.
// Parameters:
//   - template: The policy template to link
//   - templateID: The identifier for the template
//   - linkID: The identifier for the resulting linked policy
//   - slotEnv: A map of slot IDs to entity UIDs that will be substituted into the template
//
// Returns a LinkedPolicy that can be rendered into a concrete Policy.
func (p *PolicySet) LinkTemplate(templateID cedar.PolicyID, linkID cedar.PolicyID, slotEnv map[types.SlotID]types.EntityUID) error {
	_, exists := p.staticPolicies[linkID]
	if exists {
		return fmt.Errorf("link ID %s already exists in the policy set", linkID)
	}

	template := p.GetTemplate(templateID)
	if template == nil {
		return fmt.Errorf("template %s not found", templateID)
	}

	if len(slotEnv) < len(template.Slots()) {
		return fmt.Errorf("template %s requires %d variables, slot env has %d", templateID, len(template.Slots()), len(slotEnv))
	}

	for _, slotID := range template.Slots() {
		if _, ok := slotEnv[slotID]; !ok {
			return fmt.Errorf("template %s requires variable %s, missing from slot env", templateID, slotID)
		}
	}

	link := LinkedPolicy{templateID, linkID, slotEnv}
	p.linkedPolicies[linkID] = &link

	return nil
}

// GetLinkedPolicy returns the LinkedPolicy associated with the given link ID.
// If the linked policy does not exist, it returns nil.
func (p *PolicySet) GetLinkedPolicy(linkID cedar.PolicyID) *LinkedPolicy {
	return p.linkedPolicies[linkID]
}

// GetTemplate returns the Template with the given ID.
// If a template with the given ID does not exist, nil is returned.
func (p PolicySet) GetTemplate(templateID cedar.PolicyID) *Template {
	return p.templates[templateID]
}

// AddTemplate inserts or updates a template with the given ID.
// Returns true if a template with the given ID did not already exist in the set.
func (p *PolicySet) AddTemplate(templateID cedar.PolicyID, template *Template) bool {
	_, exists := p.templates[templateID]
	p.templates[templateID] = template
	return !exists
}

// RemoveTemplate removes a template from the PolicySet.
// Returns true if a template with the given ID already existed in the set.
func (p *PolicySet) RemoveTemplate(templateID cedar.PolicyID) bool {
	_, exists := p.templates[templateID]
	if exists {
		// Remove all linked policies that reference this template
		for linkID, link := range p.linkedPolicies {
			if link.templateID == templateID {
				delete(p.linkedPolicies, linkID)
			}
		}
	}

	delete(p.templates, templateID)
	return exists
}
