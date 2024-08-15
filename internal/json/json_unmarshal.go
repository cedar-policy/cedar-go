package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/types"
)

func (s *scopeJSON) ToNode(variable ast.Scope) (ast.IsScopeNode, error) {
	// TODO: should we be careful to be more strict about what is allowed here?
	switch s.Op {
	case "All":
		return variable.All(), nil
	case "==":
		if s.Entity == nil {
			return nil, fmt.Errorf("missing entity")
		}
		return variable.Eq(*s.Entity), nil
	case "in":
		if s.Entity != nil {
			return variable.In(*s.Entity), nil
		}
		return variable.InSet(s.Entities), nil
	case "is":
		if s.In == nil {
			return variable.Is(types.Path(s.EntityType)), nil
		}
		return variable.IsIn(types.Path(s.EntityType), s.In.Entity), nil
	}
	return nil, fmt.Errorf("unknown op: %v", s.Op)
}

func (j binaryJSON) ToNode(f func(a, b ast.Node) ast.Node) (ast.Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}
	right, err := j.Right.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in right: %w", err)
	}
	return f(left, right), nil
}
func (j unaryJSON) ToNode(f func(a ast.Node) ast.Node) (ast.Node, error) {
	arg, err := j.Arg.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in arg: %w", err)
	}
	return f(arg), nil
}
func (j strJSON) ToNode(f func(a ast.Node, k string) ast.Node) (ast.Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}
	return f(left, j.Attr), nil
}
func (j patternJSON) ToNode(f func(a ast.Node, k types.Pattern) ast.Node) (ast.Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}
	pattern := &types.Pattern{}
	for _, compJSON := range j.Pattern {
		if compJSON.Wildcard {
			pattern = pattern.AddWildcard()
		} else {
			pattern = pattern.AddLiteral(compJSON.Literal.Literal)
		}
	}

	return f(left, *pattern), nil
}
func (j isJSON) ToNode() (ast.Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}
	if j.In != nil {
		right, err := j.In.ToNode()
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in entity: %w", err)
		}
		return left.IsIn(types.Path(j.EntityType), right), nil
	}
	return left.Is(types.Path(j.EntityType)), nil
}
func (j ifThenElseJSON) ToNode() (ast.Node, error) {
	if_, err := j.If.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in if: %w", err)
	}
	then, err := j.Then.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in then: %w", err)
	}
	else_, err := j.Else.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in else: %w", err)
	}
	return ast.If(if_, then, else_), nil
}
func (j arrayJSON) ToNode() (ast.Node, error) {
	var nodes []ast.Node
	for _, jj := range j {
		n, err := jj.ToNode()
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in set: %w", err)
		}
		nodes = append(nodes, n)
	}
	return ast.SetNodes(nodes...), nil
}

func (j recordJSON) ToNode() (ast.Node, error) {
	nodes := map[types.String]ast.Node{}
	for k, v := range j {
		n, err := v.ToNode()
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in record: %w", err)
		}
		nodes[types.String(k)] = n
	}
	return ast.RecordNodes(nodes), nil
}

func (e extensionJSON) ToNode() (ast.Node, error) {
	if len(e) != 1 {
		return ast.Node{}, fmt.Errorf("unexpected number of extensions in node: %v", len(e))
	}
	var k string
	var v arrayJSON
	for k, v = range e {
		_, _ = k, v
	}
	_, ok := extensions.ExtMap[types.String(k)]
	if !ok {
		return ast.Node{}, fmt.Errorf("`%v` is not a known extension function or method", k)
	}
	var argNodes []ast.Node
	for _, n := range v {
		node, err := n.ToNode()
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in extension arg: %w", err)
		}
		argNodes = append(argNodes, node)
	}
	return ast.NewExtensionCall(types.String(k), argNodes...), nil
}

func (j nodeJSON) ToNode() (ast.Node, error) {
	switch {
	// Value
	case j.Value != nil:
		return ast.NewValueNode(j.Value.v), nil

	// Var
	case j.Var != nil:
		switch *j.Var {
		case consts.Principal:
			return ast.Principal(), nil
		case consts.Action:
			return ast.Action(), nil
		case consts.Resource:
			return ast.Resource(), nil
		case consts.Context:
			return ast.Context(), nil
		}
		return ast.Node{}, fmt.Errorf("unknown variable: %v", j.Var)

	// Slot
	// Unknown

	// ! or neg operators
	case j.Not != nil:
		return j.Not.ToNode(ast.Not)
	case j.Negate != nil:
		return j.Negate.ToNode(ast.Negate)

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	case j.Equals != nil:
		return j.Equals.ToNode(ast.Node.Equals)
	case j.NotEquals != nil:
		return j.NotEquals.ToNode(ast.Node.NotEquals)
	case j.In != nil:
		return j.In.ToNode(ast.Node.In)
	case j.LessThan != nil:
		return j.LessThan.ToNode(ast.Node.LessThan)
	case j.LessThanOrEqual != nil:
		return j.LessThanOrEqual.ToNode(ast.Node.LessThanOrEqual)
	case j.GreaterThan != nil:
		return j.GreaterThan.ToNode(ast.Node.GreaterThan)
	case j.GreaterThanOrEqual != nil:
		return j.GreaterThanOrEqual.ToNode(ast.Node.GreaterThanOrEqual)
	case j.And != nil:
		return j.And.ToNode(ast.Node.And)
	case j.Or != nil:
		return j.Or.ToNode(ast.Node.Or)
	case j.Plus != nil:
		return j.Plus.ToNode(ast.Node.Plus)
	case j.Minus != nil:
		return j.Minus.ToNode(ast.Node.Minus)
	case j.Times != nil:
		return j.Times.ToNode(ast.Node.Times)
	case j.Contains != nil:
		return j.Contains.ToNode(ast.Node.Contains)
	case j.ContainsAll != nil:
		return j.ContainsAll.ToNode(ast.Node.ContainsAll)
	case j.ContainsAny != nil:
		return j.ContainsAny.ToNode(ast.Node.ContainsAny)

	// ., has
	case j.Access != nil:
		return j.Access.ToNode(ast.Node.Access)
	case j.Has != nil:
		return j.Has.ToNode(ast.Node.Has)

	// is
	case j.Is != nil:
		return j.Is.ToNode()

	// like
	case j.Like != nil:
		return j.Like.ToNode(ast.Node.Like)

	// if-then-else
	case j.IfThenElse != nil:
		return j.IfThenElse.ToNode()

	// Set
	case j.Set != nil:
		return j.Set.ToNode()

	// Record
	case j.Record != nil:
		return j.Record.ToNode()

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	default:
		return j.ExtensionCall.ToNode()
	}
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
	return json.Unmarshal(b, &n.ExtensionCall)
}

func (p *patternComponentJSON) UnmarshalJSON(b []byte) error {
	var wildcard string
	err := json.Unmarshal(b, &wildcard)
	if err == nil {
		if wildcard != "Wildcard" {
			return fmt.Errorf("unknown pattern component: \"%v\"", wildcard)
		}
		p.Wildcard = true
		return nil
	}

	return json.Unmarshal(b, &p.Literal)
}

func (p *Policy) UnmarshalJSON(b []byte) error {
	var j policyJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return fmt.Errorf("error unmarshalling json: %w", err)
	}
	switch j.Effect {
	case "permit":
		*(p.unwrap()) = *ast.Permit()
	case "forbid":
		*(p.unwrap()) = *ast.Forbid()
	default:
		return fmt.Errorf("unknown effect: %v", j.Effect)
	}
	for k, v := range j.Annotations {
		p.unwrap().Annotate(types.String(k), types.String(v))
	}
	var err error
	p.Principal, err = j.Principal.ToNode(ast.Scope(ast.NewPrincipalNode()))
	if err != nil {
		return fmt.Errorf("error in principal: %w", err)
	}
	p.Action, err = j.Action.ToNode(ast.Scope(ast.NewActionNode()))
	if err != nil {
		return fmt.Errorf("error in action: %w", err)
	}
	p.Resource, err = j.Resource.ToNode(ast.Scope(ast.NewResourceNode()))
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
			p.unwrap().When(n)
		case "unless":
			p.unwrap().Unless(n)
		default:
			return fmt.Errorf("unknown condition kind: %v", c.Kind)
		}
	}

	return nil
}
