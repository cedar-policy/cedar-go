// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/parser"
)

// A PolicySet is a slice of policies.
type PolicySet []Policy

// NewPolicySet will create a PolicySet from the given text document with the
// given file name used in Position data.  If there is an error parsing the
// document, it will be returned.
func NewPolicySet(fileName string, document []byte) (PolicySet, error) {
	var res parser.PolicySet
	if err := res.UnmarshalCedar(document); err != nil {
		return nil, fmt.Errorf("parser error: %w", err)
	}
	var policies PolicySet
	for _, p := range res {
		policies = append(policies, Policy{
			Position: Position{
				Filename: fileName,
				Offset:   p.Position.Offset,
				Line:     p.Position.Line,
				Column:   p.Position.Column,
			},
			Annotations: newAnnotationsFromSlice(p.Policy.Annotations),
			Effect:      Effect(p.Policy.Effect),
			eval:        eval.Compile(p.Policy.Policy),
			ast:         p.Policy.Policy,
		})
	}
	return policies, nil
}
