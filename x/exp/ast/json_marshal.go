package ast

import (
	"encoding/json"
	"fmt"
)

func (s *scopeJSON) FromNode(src Node) error {
	switch src.nodeType {
	case nodeTypeBoolean:
		s.Op = "All"
		return nil
	case nodeTypeEquals:
		n := scopeEqNode(src)
		s.Op = "=="
		e := n.Entity()
		s.Entity = &e
		return nil
	case nodeTypeIn:
		n := scopeInNode(src)
		s.Op = "in"
		if n.IsSet() {
			s.Entities = n.Set()
		} else {
			e := n.Entity()
			s.Entity = &e
		}
		return nil
	case nodeTypeIs:
		n := scopeIsNode(src)
		s.Op = "is"
		s.EntityType = string(n.EntityType())
		return nil
	case nodeTypeIsIn: // is in
		n := scopeIsInNode(src)
		s.Op = "is"
		s.EntityType = string(n.EntityType())
		s.In = &inJSON{
			Entity: n.Entity(),
		}
		return nil
	}
	return fmt.Errorf("unexpected scope node: %v", src.nodeType)
}
func (j nodeJSON) FromNode(src Node) error {
	// TODO: all this
	return nil
}
func (p *Policy) MarshalJSON() ([]byte, error) {
	var j policyJSON
	j.Effect = "forbid"
	if p.effect {
		j.Effect = "permit"
	}
	if len(p.annotations) > 0 {
		j.Annotations = map[string]string{}
	}
	for _, a := range p.annotations {
		n := annotationNode(a)
		j.Annotations[string(n.Key())] = string(n.Value())
	}
	if err := j.Principal.FromNode(p.principal); err != nil {
		return nil, fmt.Errorf("error in principal: %w", err)
	}
	if err := j.Action.FromNode(p.action); err != nil {
		return nil, fmt.Errorf("error in action: %w", err)
	}
	if err := j.Resource.FromNode(p.resource); err != nil {
		return nil, fmt.Errorf("error in resource: %w", err)
	}
	for _, c := range p.conditions {
		var cond conditionJSON
		if err := cond.Body.FromNode(c); err != nil {
			return nil, fmt.Errorf("error in condition: %w", err)
		}
		j.Conditions = append(j.Conditions, cond)
	}
	return json.Marshal(j)
}
