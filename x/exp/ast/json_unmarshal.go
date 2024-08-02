package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
)

func (s *scopeJSON) ToNode(variable scope) (Node, error) {
	switch s.Op {
	case "All":
		return variable.All(), nil
	case "==":
		if s.Entity == nil {
			return Node{}, fmt.Errorf("missing entity")
		}
		return variable.Eq(*s.Entity), nil
	case "in":
		if s.Entity != nil {
			return variable.In(*s.Entity), nil
		}
		return variable.InSet(s.Entities), nil
	case "is":
		if s.In == nil {
			return variable.Is(types.String(s.EntityType)), nil
		}
		return variable.IsIn(types.String(s.EntityType), s.In.Entity), nil
	}
	return Node{}, fmt.Errorf("unknown op: %v", s.Op)
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
func (j unaryJSON) ToNode(f func(a Node) Node) (Node, error) {
	arg, err := j.Arg.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in arg: %w", err)
	}
	return f(arg), nil
}
func (j strJSON) ToNode(f func(a Node, k string) Node) (Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in left: %w", err)
	}
	return f(left, j.Attr), nil
}
func (j patternJSON) ToNode(f func(a Node, k string) Node) (Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in left: %w", err)
	}
	return f(left, j.Pattern), nil
}
func (j isJSON) ToNode() (Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in left: %w", err)
	}
	if j.In != nil {
		return left.IsIn(types.String(j.EntityType), Entity(j.In.Entity)), nil
	}
	return left.Is(types.String(j.EntityType)), nil
}
func (j ifThenElseJSON) ToNode() (Node, error) {
	if_, err := j.If.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in if: %w", err)
	}
	then, err := j.Then.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in then: %w", err)
	}
	else_, err := j.Else.ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in else: %w", err)
	}
	return If(if_, then, else_), nil
}
func (j arrayJSON) ToNode() (Node, error) {
	var nodes []Node
	for _, jj := range j {
		n, err := jj.ToNode()
		if err != nil {
			return Node{}, fmt.Errorf("error in set: %w", err)
		}
		nodes = append(nodes, n)
	}
	return SetNodes(nodes...), nil
}

func (j arrayJSON) ToDecimalNode() (Node, error) {
	if len(j) != 1 {
		return Node{}, fmt.Errorf("unexpected number of arguments for extension: %v", len(j))
	}
	arg, err := j[0].ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in extension: %w", err)
	}
	s, ok := arg.value.(types.String)
	if !ok {
		return Node{}, fmt.Errorf("unexpected type for decimal")
	}
	v, err := types.ParseDecimal(string(s))
	if err != nil {
		return Node{}, fmt.Errorf("error parsing decimal: %w", err)
	}
	return Decimal(v), nil
}

func (j arrayJSON) ToIPAddrNode() (Node, error) {
	if len(j) != 1 {
		return Node{}, fmt.Errorf("unexpected number of arguments for extension: %v", len(j))
	}
	arg, err := j[0].ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in extension: %w", err)
	}
	s, ok := arg.value.(types.String)
	if !ok {
		return Node{}, fmt.Errorf("unexpected type for ipaddr")
	}
	v, err := types.ParseIPAddr(string(s))
	if err != nil {
		return Node{}, fmt.Errorf("error parsing ipaddr: %w", err)
	}
	return IPAddr(v), nil
}

func (j recordJSON) ToNode() (Node, error) {
	nodes := map[types.String]Node{}
	for k, v := range j {
		n, err := v.ToNode()
		if err != nil {
			return Node{}, fmt.Errorf("error in record: %w", err)
		}
		nodes[types.String(k)] = n
	}
	return RecordNodes(nodes), nil
}

func (e extMethodCallJSON) ToNode() (Node, error) {
	if len(e) != 1 {
		return Node{}, fmt.Errorf("unexpected number of extension methods in node: %v", len(e))
	}
	for k, v := range e {
		if len(v) == 0 {
			return Node{}, fmt.Errorf("extension method '%v' must have at least one argument", k)
		}
		var argNodes []Node
		for _, n := range v {
			node, err := n.ToNode()
			if err != nil {
				return Node{}, fmt.Errorf("error in extension method argument: %w", err)
			}
			argNodes = append(argNodes, node)
		}
		return newExtMethodCallNode(argNodes[0], k, argNodes[1:]...), nil
	}
	panic("unreachable code")
}

func (j nodeJSON) ToNode() (Node, error) {
	switch {
	// Value
	case j.Value != nil:
		var v types.Value
		if err := types.UnmarshalJSON(*j.Value, &v); err != nil {
			return Node{}, fmt.Errorf("error unmarshalling value: %w", err)
		}
		return valueToNode(v), nil

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

	// is
	case j.Is != nil:
		return j.Is.ToNode()

	// like
	case j.Like != nil:
		return j.Like.ToNode(Node.Like)

	// if-then-else
	case j.IfThenElse != nil:
		return j.IfThenElse.ToNode()

	// Set
	case j.Set != nil:
		return j.Set.ToNode()

	// Record
	case j.Record != nil:
		return j.Record.ToNode()

	// Any other function: decimal, ip
	case j.Decimal != nil:
		return j.Decimal.ToDecimalNode()
	case j.IP != nil:
		return j.IP.ToIPAddrNode()

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	case j.ExtensionMethod != nil:
		return j.ExtensionMethod.ToNode()
	}

	return Node{}, fmt.Errorf("unknown node")
}

func (n *nodeJSON) UnmarshalJSON(b []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()

	type nodeJSONAlias nodeJSON
	if err := decoder.Decode((*nodeJSONAlias)(n)); err == nil {
		return nil
	} else if !strings.HasPrefix(err.Error(), "json: unknown field") {
		return err
	}

	// If an unknown field was parsed, the spec tells us to treat it as an extension method:
	// > Any other key
	// > This key is treated as the name of an extension function or method. The value must
	// > be a JSON array of values, each of which is itself an JsonExpr object. Note that for
	// > method calls, the method receiver is the first argument.
	return json.Unmarshal(b, &n.ExtensionMethod)
}

func (p *Policy) UnmarshalJSON(b []byte) error {
	var j policyJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return fmt.Errorf("error unmarshalling json: %w", err)
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
	p.principal, err = j.Principal.ToNode(scope(Principal()))
	if err != nil {
		return fmt.Errorf("error in principal: %w", err)
	}
	p.action, err = j.Action.ToNode(scope(Action()))
	if err != nil {
		return fmt.Errorf("error in action: %w", err)
	}
	p.resource, err = j.Resource.ToNode(scope(Resource()))
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
