package templates

import (
	"bytes"
	"fmt"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/parser"
)

// PolicyList represents a list of un-named Policy's. Cedar documents, unlike the PolicySet form, don't have a means of
// naming individual policies.
type PolicyList struct {
	StaticPolicies []*Policy   // StaticPolicies holds the list of static (non-template) policies.
	Templates      []*Template // Templates holds the list of policy templates.
}

// NewPolicyListFromBytes creates a PolicyList from the given Cedar policy document bytes and assigns the provided file name
// to each policy and template for position tracking. Returns an error if parsing fails.
func NewPolicyListFromBytes(fileName string, document []byte) (PolicyList, error) {
	var policySlice PolicyList
	if err := policySlice.UnmarshalCedar(document); err != nil {
		return PolicyList{}, err
	}
	for _, p := range policySlice.StaticPolicies {
		p.SetFilename(fileName)
	}

	for _, p := range policySlice.Templates {
		p.SetFilename(fileName)
	}

	return policySlice, nil
}

// UnmarshalCedar parses a concatenation of un-named Cedar policy statements from the provided byte slice and populates
// the PolicyList with static policies and templates. Returns an error if parsing fails.
func (p *PolicyList) UnmarshalCedar(b []byte) error {
	var res parser.PolicySlice
	if err := res.UnmarshalCedar(b); err != nil {
		return fmt.Errorf("parser error: %w", err)
	}

	staticPolicies := make([]*Policy, 0, len(res.StaticPolicies))
	for _, p := range res.StaticPolicies {
		newPolicy := NewPolicyFromAST((*ast.Policy)(p))
		staticPolicies = append(staticPolicies, newPolicy)
	}

	templates := make([]*Template, 0, len(res.Templates))
	for _, p := range res.Templates {
		t := Template(*p)
		templates = append(templates, &t)
	}

	p.StaticPolicies = staticPolicies
	p.Templates = templates

	return nil
}

// MarshalCedar emits a concatenated Cedar representation of the policies and templates in the PolicyList as a byte slice.
func (p PolicyList) MarshalCedar() []byte {
	var buf bytes.Buffer
	for i, policy := range p.StaticPolicies {
		buf.Write(policy.MarshalCedar())

		if i < len(p.StaticPolicies)-1 {
			buf.WriteString("\n\n")
		}
	}

	if len(p.Templates) > 0 {
		buf.WriteString("\n\n")
	}

	for i, template := range p.Templates {
		buf.Write(template.MarshalCedar())

		if i < len(p.Templates)-1 {
			buf.WriteString("\n\n")
		}
	}

	return buf.Bytes()
}
