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

func (j binaryJSON) ToNode(f func(a, b Node) Node) (Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in left: %w", err)
	}
	right, err := j.Right.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in right: %w", err)
	}
	return f(left, right), nil
}

type unaryJSON struct {
	Arg nodeJSON `json:"arg"`
}

func (j unaryJSON) ToNode(f func(a Node) Node) (Node, error) {
	arg, err := j.Arg.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in arg: %w", err)
	}
	return f(arg), nil
}

type attrJSON struct {
	Left nodeJSON `json:"left"`
	Attr string   `json:"attr"`
}

func (j attrJSON) ToNode(f func(a Node, k string) Node) (Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in left: %w", err)
	}
	return f(left, j.Attr), nil
}

type nodeJSON struct {

	// Value
	Value *string `json:"Value"` // could be any

	// Var
	Var *string `json:"Var"`

	// Slot
	// Unknown

	// ! or neg operators
	Not    *unaryJSON `json:"!"`
	Negate *unaryJSON `json:"neg"`

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	Equals             *binaryJSON `json:"=="`
	NotEquals          *binaryJSON `json:"!="`
	In                 *binaryJSON `json:"in"`
	LessThan           *binaryJSON `json:"<"`
	LessThanOrEqual    *binaryJSON `json:"<="`
	GreaterThan        *binaryJSON `json:">"`
	GreaterThanOrEqual *binaryJSON `json:">="`
	And                *binaryJSON `json:"&&"`
	Or                 *binaryJSON `json:"||"`
	Plus               *binaryJSON `json:"+"`
	Minus              *binaryJSON `json:"-"`
	Times              *binaryJSON `json:"*"`
	Contains           *binaryJSON `json:"contains"`
	ContainsAll        *binaryJSON `json:"containsAll"`
	ContainsAny        *binaryJSON `json:"containsAny"`

	// ., has
	Access *attrJSON `json:"."`
	Has    *attrJSON `json:"has"`

	// like
	// if-then-else
	// Set
	// Record
	// Any other key

}

func (j nodeJSON) ToNode() (Node, error) {
	switch {
	// Value
	case j.Value != nil:
		return String(types.String(*j.Value)), nil

	// Var
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

	// Slot
	// Unknown

	// ! or neg operators
	case j.Not != nil:
		return j.Not.ToNode(Not)
	case j.Negate != nil:
		return j.Negate.ToNode(Negate)

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	case j.Equals != nil:
		return j.Equals.ToNode(Node.Equals)
	case j.NotEquals != nil:
		return j.NotEquals.ToNode(Node.NotEquals)
	case j.In != nil:
		return j.In.ToNode(Node.In)
	case j.LessThan != nil:
		return j.LessThan.ToNode(Node.LessThan)
	case j.LessThanOrEqual != nil:
		return j.LessThanOrEqual.ToNode(Node.LessThanOrEqual)
	case j.GreaterThan != nil:
		return j.GreaterThan.ToNode(Node.GreaterThan)
	case j.GreaterThanOrEqual != nil:
		return j.GreaterThanOrEqual.ToNode(Node.GreaterThanOrEqual)
	case j.And != nil:
		return j.And.ToNode(Node.And)
	case j.Or != nil:
		return j.Or.ToNode(Node.Or)
	case j.Plus != nil:
		return j.Plus.ToNode(Node.Plus)
	case j.Minus != nil:
		return j.Minus.ToNode(Node.Minus)
	case j.Times != nil:
		return j.Times.ToNode(Node.Times)
	case j.Contains != nil:
		return j.Contains.ToNode(Node.Contains)
	case j.ContainsAll != nil:
		return j.ContainsAll.ToNode(Node.ContainsAll)
	case j.ContainsAny != nil:
		return j.ContainsAny.ToNode(Node.ContainsAny)

	// ., has
	case j.Access != nil:
		return j.Access.ToNode(Node.Access)
	case j.Has != nil:
		return j.Has.ToNode(Node.Has)
	}

	// like
	// if-then-else
	// Set
	// Record
	// Any other key

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
