package cedar

import (
	"bytes"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/types"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

type Template parser.Policy

func (p *Template) MarshalCedar() []byte {
	cedarPolicy := (*parser.Policy)(p)

	var buf bytes.Buffer
	cedarPolicy.MarshalCedar(&buf)

	return buf.Bytes()
}

// SetFilename sets the filename of this template.
func (p *Template) SetFilename(fileName string) {
	p.Position.Filename = fileName
}

type LinkedPolicy parser.LinkedPolicy

func LinkTemplate(template Template, templateID string, linkID string, slotEnv map[types.SlotID]types.EntityUID) LinkedPolicy {
	t := parser.Template(template)
	linkedPolicy := parser.NewLinkedPolicy(&t, templateID, linkID, slotEnv)

	return LinkedPolicy(linkedPolicy)
}

func (p LinkedPolicy) Render() (*Policy, error) {
	pl := parser.LinkedPolicy(p)

	policy, err := pl.Render()
	if err != nil {
		return nil, err
	}

	internalPolicy := internalast.Policy(policy)

	return newPolicy(&internalPolicy), nil
}

func (p LinkedPolicy) MarshalJSON() ([]byte, error) {
	pl := parser.LinkedPolicy(p)

	return pl.MarshalJSON()
}

func (p *PolicySet) AddLinkedPolicy(lp LinkedPolicy) {
	policy, err := lp.Render()
	if err != nil {
		return
	}

	p.Add(PolicyID(lp.LinkID), policy)
}

// GetTemplate returns the Template with the given ID. If a template with the given ID
// does not exist, nil is returned.
func (p PolicySet) GetTemplate(templateID PolicyID) *Template {
	return p.policies.Templates[templateID]
}

// AddTemplate inserts or updates a template with the given ID. Returns true if a template
// with the given ID did not already exist in the set.
func (p *PolicySet) AddTemplate(templateID PolicyID, template *Template) bool {
	_, exists := p.policies.Templates[templateID]
	p.policies.Templates[templateID] = template
	return !exists
}

// RemoveTemplate removes a template from the PolicySet. Returns true if a template with
// the given ID already existed in the set.
func (p *PolicySet) RemoveTemplate(templateID PolicyID) bool {
	_, exists := p.policies.Templates[templateID]
	delete(p.policies.Templates, templateID)
	return exists
}
