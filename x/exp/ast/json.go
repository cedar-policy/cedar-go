package ast

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

type policyJSON struct {
	Effect      string            `json:"effect"`
	Annotations map[string]string `json:"annotations"`
	Principal   scopeJSON         `json:"principal"`
	Action      scopeJSON         `json:"action"`
	Resource    scopeJSON         `json:"resource"`
	Conditions  []conditionJSON   `json:"conditions"`
}

type scopeJSON struct {
	Op         string            `json:"op"`
	Entity     types.EntityUID   `json:"entity"`
	Entities   []types.EntityUID `json:"entities"`
	EntityType string            `json:"entity_type"`
	In         *struct {
		Entity types.EntityUID `json:"entity"`
	} `json:"in"`
}

func (s *scopeJSON) ToNode(n Node) (Node, error) {
	switch s.Op {
	case "All":
		return True(), nil
	case "==":
		return n.Equals(Entity(s.Entity)), nil
	case "in":
		var zero types.EntityUID
		if s.Entity != zero {
			return n.In(Entity(s.Entity)), nil // TODO: review shape, maybe .In vs .InNode
		}
		var set types.Set
		for _, e := range s.Entities {
			set = append(set, e)
		}
		return n.In(Set(set)), nil // TODO: maybe there is an In and an InSet Node?
	case "is":
		isNode := n.Is(String(types.String(s.EntityType))) // TODO: hmmm, I'm not sure can this be Stronger-typed?
		if s.In == nil {
			return isNode, nil
		}
		return isNode.And(n.In(Entity(s.In.Entity))), nil
	}
	return Node{}, fmt.Errorf("unknown op: %v", s.Op)
}

type conditionJSON struct {
	Kind string   `json:"kind"`
	Body nodeJSON `json:"body"`
}

type binaryJSON struct {
	Left  nodeJSON `json:"left"`
	Right nodeJSON `json:"right"`
}

type accessJSON struct {
	Left nodeJSON `json:"left"`
	Attr string   `json:"attr"`
}

type nodeJSON struct {
	Equals *binaryJSON `json:"=="`
	Access *accessJSON `json:"."`
	Var    *string     `json:"Var"`
	Value  *string     `json:"Value"` // could be any
}

func (j nodeJSON) ToNode() (Node, error) {
	switch {
	case j.Equals != nil:
		left, err := j.Equals.Left.ToNode()
		if err != nil {
			return Node{}, fmt.Errorf("error in left of equals: %w", err)
		}
		right, err := j.Equals.Right.ToNode()
		if err != nil {
			return Node{}, fmt.Errorf("error in right of equals: %w", err)
		}
		return left.Equals(right), nil
	case j.Access != nil:
		left, err := j.Access.Left.ToNode()
		if err != nil {
			return Node{}, fmt.Errorf("error in left of access: %w", err)
		}
		return left.Access(j.Access.Attr), nil
	case j.Var != nil:
		switch *j.Var {
		case "principal":
			return Principal(), nil
		case "action":
			return Action(), nil
		case "resource":
			return Resource(), nil
		case "context":
			return Context(), nil
		}
		return Node{}, fmt.Errorf("unknown var: %v", j.Var)
	case j.Value != nil:
		return String(types.String(*j.Value)), nil
	}

	return Node{}, fmt.Errorf("unknown node")
}

func (p *Policy) UnmarshalJSON(b []byte) error {
	var j policyJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	switch j.Effect {
	case "permit":
		*p = *Permit()
	case "forbid":
		*p = *Forbid()
	default:
		return fmt.Errorf("unknown effect: %v", j.Effect)
	}
	for k, v := range j.Annotations {
		p.Annotate(types.String(k), types.String(v))
	}
	var err error
	p.principal, err = j.Principal.ToNode(Principal())
	if err != nil {
		return fmt.Errorf("error in principal: %w", err)
	}
	p.action, err = j.Action.ToNode(Action())
	if err != nil {
		return fmt.Errorf("error in action: %w", err)
	}
	p.resource, err = j.Resource.ToNode(Resource())
	if err != nil {
		return fmt.Errorf("error in resource: %w", err)
	}
	for _, c := range j.Conditions {
		n, err := c.Body.ToNode()
		if err != nil {
			return fmt.Errorf("error in conditions: %w", err)
		}
		switch c.Kind {
		case "when":
			p.When(n)
		case "unless":
			p.Unless(n)
		default:
			return fmt.Errorf("unknown condition kind: %v", c.Kind)
		}
	}

	return nil
}
