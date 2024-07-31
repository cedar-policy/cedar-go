package ast

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func (s *scopeJSON) ToNode(variable Node) (Node, error) {
	switch s.Op {
	case "All":
		return True(), nil
	case "==":
		if s.Entity == nil {
			return Node{}, fmt.Errorf("missing entity")
		}
		return variable.Equals(Entity(*s.Entity)), nil
	case "in":
		if s.Entity != nil {
			return variable.In(Entity(*s.Entity)), nil // TODO: review shape, maybe .In vs .InNode
		}
		var set types.Set
		for _, e := range s.Entities {
			set = append(set, e)
		}
		return variable.In(Set(set)), nil // TODO: maybe there is an In and an InSet Node?
	case "is":
		if s.In == nil {
			return variable.Is(types.String(s.EntityType)), nil // TODO: hmmm, I'm not sure can this be Stronger-typed?
		}
		return variable.IsIn(types.String(s.EntityType), Entity(s.In.Entity)), nil
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
	return SetNodes(nodes), nil
}

func (j arrayJSON) ToExt1(f func(Node) Node) (Node, error) {
	if len(j) != 1 {
		return Node{}, fmt.Errorf("unexpected number of arguments for extension: %v", len(j))
	}
	arg, err := j[0].ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in extension: %w", err)
	}
	return f(arg), nil
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

func (j arrayJSON) ToExt2(f func(Node, Node) Node) (Node, error) {
	if len(j) != 2 {
		return Node{}, fmt.Errorf("unexpected number of arguments for extension: %v", len(j))
	}
	left, err := j[0].ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in argument 0: %w", err)
	}
	right, err := j[1].ToNode()
	if err != nil {
		return Node{}, fmt.Errorf("error in argument 1: %w", err)
	}
	return f(left, right), nil
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

var ( // TODO: de-dupe from types?
	errJSONDecode          = fmt.Errorf("error decoding json")
	errJSONLongOutOfRange  = fmt.Errorf("long out of range")
	errJSONUnsupportedType = fmt.Errorf("unsupported type")
)

func parseRawMessage(j *json.RawMessage) (Node, error) {
	// TODO: de-dupe from types?  though it's not 100% compat, because of extensions :(
	// TODO: make this faster if it matters
	{
		var res types.EntityUID
		ptr := &res
		if err := ptr.UnmarshalJSON(*j); err == nil {
			return Entity(res), nil
		}
	}

	var res interface{}
	dec := json.NewDecoder(bytes.NewBuffer(*j))
	dec.UseNumber()
	if err := dec.Decode(&res); err != nil {
		return Node{}, fmt.Errorf("%w: %w", errJSONDecode, err)
	}
	switch vv := res.(type) {
	case string:
		return String(types.String(vv)), nil
	case bool:
		return Boolean(types.Boolean(vv)), nil
	case json.Number:
		l, err := vv.Int64()
		if err != nil {
			return Node{}, fmt.Errorf("%w: %w", errJSONLongOutOfRange, err)
		}
		return Long(types.Long(l)), nil
	}
	return Node{}, errJSONUnsupportedType

}

func (j nodeJSON) ToNode() (Node, error) {
	switch {
	// Value
	case j.Value != nil:
		return parseRawMessage(j.Value)

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
	case j.LessThanExt != nil:
		return j.LessThanExt.ToExt2(Node.LessThanExt)
	case j.LessThanOrEqualExt != nil:
		return j.LessThanOrEqualExt.ToExt2(Node.LessThanOrEqualExt)
	case j.GreaterThanExt != nil:
		return j.GreaterThanExt.ToExt2(Node.GreaterThanExt)
	case j.GreaterThanOrEqualExt != nil:
		return j.GreaterThanOrEqualExt.ToExt2(Node.GreaterThanOrEqualExt)
	case j.IsIpv4Ext != nil:
		return j.IsIpv4Ext.ToExt1(Node.IsIpv4)
	case j.IsIpv6Ext != nil:
		return j.IsIpv6Ext.ToExt1(Node.IsIpv6)
	case j.IsLoopbackExt != nil:
		return j.IsLoopbackExt.ToExt1(Node.IsLoopback)
	case j.IsMulticastExt != nil:
		return j.IsMulticastExt.ToExt1(Node.IsMulticast)
	case j.IsInRangeExt != nil:
		return j.IsInRangeExt.ToExt2(Node.IsInRange)
	}

	return Node{}, fmt.Errorf("unknown node")
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
