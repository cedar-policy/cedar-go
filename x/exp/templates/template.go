package templates

import (
	"bytes"
	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/types"
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
//type LinkedPolicy parser.LinkedPolicy

type LinkedPolicy struct {
	templateID cedar.PolicyID
	linkID     cedar.PolicyID
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
func (p *PolicySet) LinkTemplate(templateID cedar.PolicyID, linkID cedar.PolicyID, slotEnv map[types.SlotID]types.EntityUID) LinkedPolicy {
	link := LinkedPolicy{templateID, linkID, slotEnv}
	p.links[linkID] = &link

	return link
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
	delete(p.templates, templateID)
	return exists
}
