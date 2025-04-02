package cedar

import (
	"bytes"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/parser"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

// PolicyList represents a list of un-named Policy's. Cedar documents, unlike the PolicySet form, don't have a means of
// naming individual policies.
type PolicyList struct {
	StaticPolicies []*Policy
	Templates      []*Template
}

// NewPolicyListFromBytes will create a Policies from the given text document with the given file name used in Position
// data.  If there is an error parsing the document, it will be returned.
func NewPolicyListFromBytes(fileName string, document []byte) (PolicyList, error) {
	var policySlice PolicyList
	if err := policySlice.UnmarshalCedar(document); err != nil {
		return PolicyList{}, err
	}
	for _, p := range policySlice.StaticPolicies {
		p.SetFilename(fileName)
	}
	//todo: set template filename
	return policySlice, nil
}

// UnmarshalCedar parses a concatenation of un-named Cedar policy statements. Names can be assigned to these policies
// when adding them to a PolicySet.
func (p *PolicyList) UnmarshalCedar(b []byte) error {
	var res parser.PolicySlice
	if err := res.UnmarshalCedar(b); err != nil {
		return fmt.Errorf("parser error: %w", err)
	}

	staticPolicies := make([]*Policy, 0, len(res.StaticPolicies))
	for _, p := range res.StaticPolicies {
		newPolicy := newPolicy((*internalast.Policy)(p))
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

// MarshalCedar emits a concatenated Cedar representation of the policies.
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
