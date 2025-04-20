package parser

import (
	"encoding/json"
	"fmt"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/types"
)

type Template ast.Policy

func (p *Template) ClonePolicy() *Policy {
	clone := (*ast.Policy)(p).Clone()
	parserPolicy := Policy(clone)

	return &parserPolicy
}

type LinkedPolicy struct {
	TemplateID string
	LinkID     string
	Template   *Template

	slotEnv map[types.SlotID]types.EntityUID
}

// NewLinkedPolicy creates a new instance of LinkedPolicy.
func NewLinkedPolicy(template *Template, templateID string, linkID string, slotEnv map[types.SlotID]types.EntityUID) LinkedPolicy {
	return LinkedPolicy{
		Template:   template,
		TemplateID: templateID,
		LinkID:     linkID,
		slotEnv:    slotEnv,
	}
}

func (p LinkedPolicy) Render() (Policy, error) {
	body := p.Template.ClonePolicy().unwrap()

	if len(body.Slots()) != len(p.slotEnv) {
		return Policy{}, fmt.Errorf("slot env length %d does not match template slot length %d", len(p.slotEnv), len(body.Slots()))
	}

	for _, slot := range body.Slots() {
		switch slot {
		case types.PrincipalSlot:
			body.Principal = linkScope(body.Principal, p.slotEnv)
		case types.ResourceSlot:
			body.Resource = linkScope(body.Resource, p.slotEnv)
		default:
			return Policy{}, fmt.Errorf("unknown variable %s", slot)
		}
	}

	return Policy(*body), nil
}

func linkScope[T ast.IsScopeNode](scope T, slotEnv map[types.SlotID]types.EntityUID) T {
	var linkedScope any = scope

	switch t := any(scope).(type) {
	case ast.ScopeTypeEq:
		t.Entity = resolveSlot(t.Entity, slotEnv)

		linkedScope = t
	case ast.ScopeTypeIn:
		t.Entity = resolveSlot(t.Entity, slotEnv)

		linkedScope = t
	case ast.ScopeTypeIsIn:
		t.Entity = resolveSlot(t.Entity, slotEnv)

		linkedScope = t
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}

	return linkedScope.(T)
}

func resolveSlot(ef types.EntityReference, slotEnv map[types.SlotID]types.EntityUID) types.EntityReference {
	switch e := ef.(type) {
	case types.EntityUID:
		return e
	case types.VariableSlot:
		return slotEnv[e.ID]
	default:
		panic(fmt.Sprintf("unknown entity reference type %T", e))
	}
}

// MarshalJSON marshals a LinkedPolicy to JSON following cedar-cli format.
func (p LinkedPolicy) MarshalJSON() ([]byte, error) {
	lp := struct {
		TemplateID string            `json:"template_id"`
		LinkID     string            `json:"link_id"`
		Args       map[string]string `json:"args"`
	}{
		TemplateID: p.TemplateID,
		LinkID:     p.LinkID,
	}

	lp.Args = make(map[string]string, len(p.slotEnv))
	for k, v := range p.slotEnv {
		lp.Args[string(k)] = v.String()
	}

	return json.Marshal(lp)
}
