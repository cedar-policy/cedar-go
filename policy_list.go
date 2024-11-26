package cedar

import (
	"bytes"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/parser"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

// PolicyList represents a list of un-named Policy's. Cedar documents, unlike the PolicySet form, don't have a means of
// naming individual policies.
type PolicyList []*Policy

// NewPolicyListFromBytes will create a Policies from the given text document with the given file name used in Position
// data.  If there is an error parsing the document, it will be returned.
func NewPolicyListFromBytes(fileName string, document []byte) (PolicyList, error) {
	var policySlice PolicyList
	if err := policySlice.UnmarshalCedar(document); err != nil {
		return nil, err
	}
	for _, p := range policySlice {
		p.SetFilename(fileName)
	}
	return policySlice, nil
}

// UnmarshalCedar parses a concatenation of un-named Cedar policy statements. Names can be assigned to these policies
// when adding them to a PolicySet.
func (p *PolicyList) UnmarshalCedar(b []byte) error {
	var res parser.PolicySlice
	if err := res.UnmarshalCedar(b); err != nil {
		return fmt.Errorf("parser error: %w", err)
	}
	policySlice := make([]*Policy, 0, len(res))
	for _, p := range res {
		newPolicy := newPolicy((*internalast.Policy)(p))
		policySlice = append(policySlice, newPolicy)
	}
	*p = policySlice
	return nil
}

// MarshalCedar emits a concatenated Cedar representation of the policies.
func (p PolicyList) MarshalCedar() []byte {
	var buf bytes.Buffer
	for i, policy := range p {
		buf.Write(policy.MarshalCedar())

		if i < len(p)-1 {
			buf.WriteString("\n\n")
		}
	}
	return buf.Bytes()
}
