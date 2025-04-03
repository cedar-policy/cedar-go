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

func (p PolicySet) StoreLinkedPolicy(lp LinkedPolicy) {
	policy, err := lp.Render()
	if err != nil {
		return
	}

	p.Add(PolicyID(lp.LinkID), policy)
}
