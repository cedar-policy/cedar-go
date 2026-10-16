package json

import (
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

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
		if s.EntityType == "" {
			return nil, fmt.Errorf("missing entity_type")
		}
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
	if j.EntityType == "" {
		return ast.Node{}, fmt.Errorf("missing entity_type")
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
		if v == nil {
			return ast.Node{}, fmt.Errorf("error in record: null value for %q", k)
		}
		n, err := v.ToNode()
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in record: %w", err)
		}
		nodes = append(nodes, ast.Pair{Key: types.String(k), Value: n})
	}
	return ast.Record(nodes), nil
}

func (e *extensionJSON) ToNode() (ast.Node, error) {
	_, ok := extensions.ExtMap[types.Path(e.Name)]
	if !ok {
		return ast.Node{}, fmt.Errorf("`%v` is not a known extension function or method", e.Name)
	}
	var argNodes []ast.Node
	for _, n := range e.Args {
		node, err := n.ToNode()
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in extension arg: %w", err)
		}
		argNodes = append(argNodes, node)
	}
	return ast.NewExtensionCall(types.Path(e.Name), argNodes...), nil
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
		return (*n.Set).ToNode()

	// Record
	case n.Record != nil:
		return (*n.Record).ToNode()

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	case n.ExtensionCall != nil:
		return n.ExtensionCall.ToNode()

	default:
		return ast.Node{}, errors.New("missing expression")
	}
}

// UnmarshalJSONFrom decodes a JSONExpr object, which must have exactly one
// member whose name selects the kind of expression. Any name that isn't a
// known operator is treated as an extension function or method call where the
// "key is treated as the name of an extension function or method. The value
// must be a JSON array of values, each of which is itself a JsonExpr object."
func (n *nodeJSON) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var newNode nodeJSON
	if tok, err := dec.ReadToken(); err != nil {
		return err
	} else if tok.Kind() != jsontext.KindBeginObject {
		return fmt.Errorf("expression must be an object, got %v", tok.Kind())
	}
	if dec.PeekKind() == jsontext.KindEndObject {
		return errors.New("expression object must have exactly one member, got none")
	}
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	name := tok.String()
	if dec.PeekKind() == jsontext.KindNull {
		return fmt.Errorf("expression member %q must not be null", name)
	}
	// This is unfortunate, but somewhat cleanly enforces the single member
	// constraint without unmarshaling multiple times. We might consider a
	// sumtype interface-based approach to implementing nodeJSON in the future,
	// which might be tidier than this approach of duplicating JSON name tags.
	fieldMapping := map[string]any{
		"Value":        &newNode.Value,
		"Var":          &newNode.Var,
		"!":            &newNode.Not,
		"neg":          &newNode.Negate,
		"isEmpty":      &newNode.IsEmpty,
		"==":           &newNode.Equals,
		"!=":           &newNode.NotEquals,
		"in":           &newNode.In,
		"<":            &newNode.LessThan,
		"<=":           &newNode.LessThanOrEqual,
		">":            &newNode.GreaterThan,
		">=":           &newNode.GreaterThanOrEqual,
		"&&":           &newNode.And,
		"||":           &newNode.Or,
		"+":            &newNode.Add,
		"-":            &newNode.Subtract,
		"*":            &newNode.Multiply,
		"contains":     &newNode.Contains,
		"containsAll":  &newNode.ContainsAll,
		"containsAny":  &newNode.ContainsAny,
		"getTag":       &newNode.GetTag,
		"hasTag":       &newNode.HasTag,
		".":            &newNode.Access,
		"has":          &newNode.Has,
		"is":           &newNode.Is,
		"like":         &newNode.Like,
		"if-then-else": &newNode.IfThenElse,
		"Set":          &newNode.Set,
		"Record":       &newNode.Record,
	}
	if field, ok := fieldMapping[name]; ok {
		err = json.UnmarshalDecode(dec, field, json.RejectUnknownMembers(true))
	} else {
		// If not a known field, the spec tells us to treat this as extension
		// function or method call.
		var args arrayJSON
		err = json.UnmarshalDecode(dec, &args, json.RejectUnknownMembers(true))
		newNode.ExtensionCall = &extensionJSON{Name: name, Args: args}
	}
	if err != nil {
		return err
	}
	if tok, err := dec.ReadToken(); err != nil {
		return err
	} else if tok.Kind() != jsontext.KindEndObject {
		return errors.New("expression object must have exactly one member")
	}
	*n = newNode
	return nil
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
