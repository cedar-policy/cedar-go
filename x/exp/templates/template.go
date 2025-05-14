package templates

import (
	"bytes"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/types"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

// Template represents a Cedar policy template that can be linked with slot values
// to create concrete policies. It's a wrapper around the internal parser.Policy type.
type Template parser.Policy

// MarshalCedar serializes the Template into its Cedar language representation.
// Returns the serialized template as a byte slice.
func (p *Template) MarshalCedar() []byte {
	cedarPolicy := (*parser.Policy)(p)

	var buf bytes.Buffer
	cedarPolicy.MarshalCedar(&buf)

	return buf.Bytes()
}

// SetFilename sets the filename of this template.
// This is useful for error reporting and debugging purposes.
func (p *Template) SetFilename(fileName string) {
	p.Position.Filename = fileName
}

// LinkedPolicy represents a template that has been linked with specific slot values.
// It's a wrapper around the internal parser.LinkedPolicy type.
type LinkedPolicy parser.LinkedPolicy

type LinkedPolicy2 struct {
	templateID PolicyID
	linkID     PolicyID
	slotEnv    map[types.SlotID]types.EntityUID
}

// LinkTemplate creates a LinkedPolicy by binding slot values to a template.
// Parameters:
//   - template: The policy template to link
//   - templateID: The identifier for the template
//   - linkID: The identifier for the resulting linked policy
//   - slotEnv: A map of slot IDs to entity UIDs that will be substituted into the template
//
// Returns a LinkedPolicy that can be rendered into a concrete Policy.
func (p *PolicySet) LinkTemplate(templateID PolicyID, linkID PolicyID, slotEnv map[types.SlotID]types.EntityUID) LinkedPolicy2 {
	link := LinkedPolicy2{templateID, linkID, slotEnv}
	p.links[linkID] = &link

	return link
}

// Render converts a LinkedPolicy into a concrete Policy by substituting all slot values.
// Returns the rendered Policy and any error that occurred during rendering.
// If rendering fails (e.g., due to missing slot values), an error is returned.
func (p LinkedPolicy) Render() (*Policy, error) {
	pl := parser.LinkedPolicy(p)

	policy, err := pl.Render()
	if err != nil {
		return nil, err
	}

	internalPolicy := internalast.Policy(policy)

	return newPolicy(&internalPolicy), nil
}

// MarshalJSON serializes the LinkedPolicy into its JSON representation.
// Returns the JSON representation as a byte slice and any error that occurred during marshaling.
func (p LinkedPolicy) MarshalJSON() ([]byte, error) {
	pl := parser.LinkedPolicy(p)

	return pl.MarshalJSON()
}

// AddLinkedPolicy renders a LinkedPolicy and adds the resulting concrete Policy to the PolicySet.
// The policy is added with the LinkID from the LinkedPolicy as its PolicyID.
// If rendering fails, no policy is added to the set.
//func (p *PolicySet) AddLinkedPolicy(lp LinkedPolicy) {
//	policy, err := lp.Render()
//	if err != nil {
//		return
//	}
//
//	p.Add(PolicyID(lp.LinkID), policy)
//}

// GetTemplate returns the Template with the given ID.
// If a template with the given ID does not exist, nil is returned.
func (p PolicySet) GetTemplate(templateID PolicyID) *Template {
	return p.templates[templateID]
}

// AddTemplate inserts or updates a template with the given ID.
// Returns true if a template with the given ID did not already exist in the set.
func (p *PolicySet) AddTemplate(templateID PolicyID, template *Template) bool {
	_, exists := p.templates[templateID]
	p.templates[templateID] = template
	return !exists
}

// RemoveTemplate removes a template from the PolicySet.
// Returns true if a template with the given ID already existed in the set.
func (p *PolicySet) RemoveTemplate(templateID PolicyID) bool {
	_, exists := p.templates[templateID]
	delete(p.templates, templateID)
	return exists
}
