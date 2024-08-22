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

type isPrincipalResourceScopeNode interface {
	ast.IsPrincipalScopeNode
	ast.IsResourceScopeNode
}

func (s *scopeJSON) ToPrincipalResourceNode() (isPrincipalResourceScopeNode, error) {
	switch s.Op {
	case "All":
		return ast.Scope{}.All(), nil
	case "==":
		if s.Entity == nil {
			return nil, fmt.Errorf("missing entity")
		}
		return ast.Scope{}.Eq(*s.Entity), nil
	case "in":
		if s.Entity == nil {
			return nil, fmt.Errorf("missing entity")
		}
		return ast.Scope{}.In(*s.Entity), nil
	case "is":
		if s.In == nil {
			return ast.Scope{}.Is(types.EntityType(s.EntityType)), nil
		}
		return ast.Scope{}.IsIn(types.EntityType(s.EntityType), s.In.Entity), nil
	}
	return nil, fmt.Errorf("unknown op: %v", s.Op)
}

func (s *scopeJSON) ToActionNode() (ast.IsActionScopeNode, error) {
	switch s.Op {
	case "All":
		return ast.Scope{}.All(), nil
	case "==":
		if s.Entity == nil {
			return nil, fmt.Errorf("missing entity")
		}
		return ast.Scope{}.Eq(*s.Entity), nil
	case "in":
		if s.Entity != nil {
			return ast.Scope{}.In(*s.Entity), nil
		}
		return ast.Scope{}.InSet(s.Entities), nil
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
func (j strJSON) ToNode(f func(a ast.Node, k types.String) ast.Node) (ast.Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}
	return f(left, types.String(j.Attr)), nil
}
func (j likeJSON) ToNode(f func(a ast.Node, k types.Pattern) ast.Node) (ast.Node, error) {
	left, err := j.Left.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}

	return f(left, j.Pattern), nil
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
		return left.IsIn(types.EntityType(j.EntityType), right), nil
	}
	return left.Is(types.EntityType(j.EntityType)), nil
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
	return ast.IfThenElse(if_, then, else_), nil
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
	return ast.Set(nodes...), nil
}

func (j recordJSON) ToNode() (ast.Node, error) {
	var nodes ast.Pairs
	for k, v := range j {
		n, err := v.ToNode()
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in record: %w", err)
		}
		nodes = append(nodes, ast.Pair{Key: types.String(k), Value: n})
	}
	return ast.Record(nodes), nil
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
	_, ok := extensions.ExtMap[types.Path(k)]
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
	return ast.NewExtensionCall(types.Path(k), argNodes...), nil
}

func (j nodeJSON) ToNode() (ast.Node, error) {
	switch {
	// Value
	case j.Value != nil:
		return ast.Value(j.Value.v), nil

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
		return j.Equals.ToNode(ast.Node.Equal)
	case j.NotEquals != nil:
		return j.NotEquals.ToNode(ast.Node.NotEqual)
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
	case j.Add != nil:
		return j.Add.ToNode(ast.Node.Add)
	case j.Subtract != nil:
		return j.Subtract.ToNode(ast.Node.Subtract)
	case j.Multiply != nil:
		return j.Multiply.ToNode(ast.Node.Multiply)
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
		p.unwrap().Annotate(types.Ident(k), types.String(v))
	}
	var err error
	p.Principal, err = j.Principal.ToPrincipalResourceNode()
	if err != nil {
		return fmt.Errorf("error in principal: %w", err)
	}
	p.Action, err = j.Action.ToActionNode()
	if err != nil {
		return fmt.Errorf("error in action: %w", err)
	}
	p.Resource, err = j.Resource.ToPrincipalResourceNode()
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
