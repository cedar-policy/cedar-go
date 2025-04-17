package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
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
		return ast.Scope{}.Eq(types.EntityUID(*s.Entity)), nil
	case "in":
		if s.Entity == nil {
			return nil, fmt.Errorf("missing entity")
		}
		return ast.Scope{}.In(types.EntityUID(*s.Entity)), nil
	case "is":
		if s.In == nil {
			return ast.Scope{}.Is(types.EntityType(s.EntityType)), nil
		}
		return ast.Scope{}.IsIn(types.EntityType(s.EntityType), types.EntityUID(s.In.Entity)), nil
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
		return ast.Scope{}.Eq(types.EntityUID(*s.Entity)), nil
	case "in":
		if s.Entity != nil {
			return ast.Scope{}.In(types.EntityUID(*s.Entity)), nil
		}
		es := make([]types.EntityUID, len(s.Entities))
		for i, e := range s.Entities {
			es[i] = types.EntityUID(e)
		}
		return ast.Scope{}.InSet(es), nil
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
	ifNode, err := j.If.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in if: %w", err)
	}
	thenNode, err := j.Then.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in then: %w", err)
	}
	elseNode, err := j.Else.ToNode()
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in else: %w", err)
	}
	return ast.IfThenElse(ifNode, thenNode, elseNode), nil
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

func (n nodeJSON) ToNode() (ast.Node, error) {
	switch {
	// Value
	case n.Value != nil:
		return ast.Value(n.Value.v), nil

	// Var
	case n.Var != nil:
		switch *n.Var {
		case consts.Principal:
			return ast.Principal(), nil
		case consts.Action:
			return ast.Action(), nil
		case consts.Resource:
			return ast.Resource(), nil
		case consts.Context:
			return ast.Context(), nil
		}
		return ast.Node{}, fmt.Errorf("unknown variable: %v", n.Var)

	// Slot
	// Unknown

	// ! or neg operators
	case n.Not != nil:
		return n.Not.ToNode(ast.Not)
	case n.Negate != nil:
		return n.Negate.ToNode(ast.Negate)

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny, hasTag, getTag
	case n.Equals != nil:
		return n.Equals.ToNode(ast.Node.Equal)
	case n.NotEquals != nil:
		return n.NotEquals.ToNode(ast.Node.NotEqual)
	case n.In != nil:
		return n.In.ToNode(ast.Node.In)
	case n.LessThan != nil:
		return n.LessThan.ToNode(ast.Node.LessThan)
	case n.LessThanOrEqual != nil:
		return n.LessThanOrEqual.ToNode(ast.Node.LessThanOrEqual)
	case n.GreaterThan != nil:
		return n.GreaterThan.ToNode(ast.Node.GreaterThan)
	case n.GreaterThanOrEqual != nil:
		return n.GreaterThanOrEqual.ToNode(ast.Node.GreaterThanOrEqual)
	case n.And != nil:
		return n.And.ToNode(ast.Node.And)
	case n.Or != nil:
		return n.Or.ToNode(ast.Node.Or)
	case n.Add != nil:
		return n.Add.ToNode(ast.Node.Add)
	case n.Subtract != nil:
		return n.Subtract.ToNode(ast.Node.Subtract)
	case n.Multiply != nil:
		return n.Multiply.ToNode(ast.Node.Multiply)
	case n.Contains != nil:
		return n.Contains.ToNode(ast.Node.Contains)
	case n.ContainsAll != nil:
		return n.ContainsAll.ToNode(ast.Node.ContainsAll)
	case n.ContainsAny != nil:
		return n.ContainsAny.ToNode(ast.Node.ContainsAny)
	case n.IsEmpty != nil:
		return n.IsEmpty.ToNode(ast.Node.IsEmpty)
	case n.GetTag != nil:
		return n.GetTag.ToNode(ast.Node.GetTag)
	case n.HasTag != nil:
		return n.HasTag.ToNode(ast.Node.HasTag)

	// ., has
	case n.Access != nil:
		return n.Access.ToNode(ast.Node.Access)
	case n.Has != nil:
		return n.Has.ToNode(ast.Node.Has)

	// is
	case n.Is != nil:
		return n.Is.ToNode()

	// like
	case n.Like != nil:
		return n.Like.ToNode(ast.Node.Like)

	// if-then-else
	case n.IfThenElse != nil:
		return n.IfThenElse.ToNode()

	// Set
	case n.Set != nil:
		return n.Set.ToNode()

	// Record
	case n.Record != nil:
		return n.Record.ToNode()

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	default:
		return n.ExtensionCall.ToNode()
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
