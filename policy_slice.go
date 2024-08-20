package cedar

import (
	"bytes"
	"fmt"

	internalast "github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/parser"
)

// PolicySlice represents a set of un-named Policy's. Cedar documents, unlike the JSON format, don't have a means of
// naming individual policies.
type PolicySlice []*Policy

// NewPolicySliceFromBytes will create a PolicySet from the given text document with the given file name used in Position
// data.  If there is an error parsing the document, it will be returned.
func NewPolicySliceFromBytes(fileName string, document []byte) (PolicySlice, error) {
	var policySlice PolicySlice
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
func (p *PolicySlice) UnmarshalCedar(b []byte) error {
	var res parser.PolicySlice
	if err := res.UnmarshalCedar(b); err != nil {
		return fmt.Errorf("parser error: %w", err)
	}
	policySlice := make([]*Policy, 0, len(res))
	for _, p := range res {
		newPolicy := newPolicy((*internalast.Policy)(p))
		policySlice = append(policySlice, &newPolicy)
	}
	*p = policySlice
	return nil
}

// MarshalCedar emits a concatenated Cedar representation of a PolicySlice
func (p PolicySlice) MarshalCedar() []byte {
	var buf bytes.Buffer
	for i, policy := range p {
		buf.Write(policy.MarshalCedar())

		if i < len(p)-1 {
			buf.WriteString("\n\n")
		}
	}
	return buf.Bytes()
}
