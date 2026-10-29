package json

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/internal/scanner"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

type Mode int

const (
	ModeGo Mode = iota
	ModeGoLegacy
	ModeStrict
)

type isPrincipalResourceScopeNode interface {
	ast.IsPrincipalScopeNode
	ast.IsResourceScopeNode
}

func (s *scopeJSON) ToPrincipalResourceNode() (isPrincipalResourceScopeNode, error) {
	switch s.Op {
	case "All", "all":
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
		if !scanner.IsPath(s.EntityType) {
			return nil, fmt.Errorf("invalid entity_type: %q", s.EntityType)
		}
		if s.In == nil {
			return ast.Scope{}.Is(types.EntityType(s.EntityType)), nil
		}
		if s.In.Entity == nil {
			return nil, fmt.Errorf(`missing or null entity in "in"`)
		}
		return ast.Scope{}.IsIn(types.EntityType(s.EntityType), types.EntityUID(*s.In.Entity)), nil
	}
	return nil, fmt.Errorf("unknown op: %v", s.Op)
}

func (s *scopeJSON) ToActionNode(m Mode) (ast.IsActionScopeNode, error) {
	switch s.Op {
	case "All", "all":
		return ast.Scope{}.All(), nil
	case "==":
		if s.Entity == nil {
			return nil, fmt.Errorf("missing entity")
		}
		e := types.EntityUID(*s.Entity)
		if err := checkActionEntity(e); err != nil {
			return nil, err
		}
		return ast.Scope{}.Eq(e), nil
	case "in":
		if s.Entity != nil {
			e := types.EntityUID(*s.Entity)
			if err := checkActionEntity(e); err != nil {
				return nil, err
			}
			return ast.Scope{}.In(e), nil
		}
		if s.Entities == nil && m != ModeGoLegacy {
			return nil, fmt.Errorf(`missing "entity" or "entities"`)
		}
		es := make([]types.EntityUID, len(s.Entities))
		for i, e := range s.Entities {
			es[i] = types.EntityUID(e)
			if err := checkActionEntity(es[i]); err != nil {
				return nil, err
			}
		}
		return ast.Scope{}.InSet(es), nil
	}
	return nil, fmt.Errorf("unknown op: %v", s.Op)
}

func checkActionEntity(e types.EntityUID) error {
	if !scanner.IsActionType(string(e.Type)) {
		return fmt.Errorf("expected an entity uid with type `Action` but got `%s`", e.MarshalCedar())
	}
	return nil
}

func (j binaryJSON) ToNode(m Mode, f func(a, b ast.Node) ast.Node) (ast.Node, error) {
	left, err := j.Left.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}
	right, err := j.Right.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in right: %w", err)
	}
	return f(left, right), nil
}
func (j unaryJSON) ToNode(m Mode, f func(a ast.Node) ast.Node) (ast.Node, error) {
	arg, err := j.Arg.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in arg: %w", err)
	}
	return f(arg), nil
}
func (j strJSON) ToNode(m Mode, f func(a ast.Node, k types.String) ast.Node) (ast.Node, error) {
	left, err := j.Left.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}
	if j.Attr == nil {
		return ast.Node{}, fmt.Errorf("missing or null attr")
	}
	return f(left, types.String(*j.Attr)), nil
}
func (j likeJSON) ToNode(m Mode, f func(a ast.Node, k types.Pattern) ast.Node) (ast.Node, error) {
	left, err := j.Left.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}

	if j.Pattern == nil {
		return ast.Node{}, fmt.Errorf("missing or null pattern")
	}
	return f(left, *j.Pattern), nil
}
func (j isJSON) ToNode(m Mode) (ast.Node, error) {
	left, err := j.Left.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in left: %w", err)
	}
	if !scanner.IsPath(j.EntityType) {
		return ast.Node{}, fmt.Errorf("invalid entity_type: %q", j.EntityType)
	}
	if j.In != nil {
		right, err := j.In.ToNode(m)
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in entity: %w", err)
		}
		return left.IsIn(types.EntityType(j.EntityType), right), nil
	}
	return left.Is(types.EntityType(j.EntityType)), nil
}
func (j ifThenElseJSON) ToNode(m Mode) (ast.Node, error) {
	ifNode, err := j.If.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in if: %w", err)
	}
	thenNode, err := j.Then.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in then: %w", err)
	}
	elseNode, err := j.Else.ToNode(m)
	if err != nil {
		return ast.Node{}, fmt.Errorf("error in else: %w", err)
	}
	return ast.IfThenElse(ifNode, thenNode, elseNode), nil
}
func (j arrayJSON) ToNode(m Mode) (ast.Node, error) {
	var nodes []ast.Node
	for _, jj := range j {
		n, err := jj.ToNode(m)
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in set: %w", err)
		}
		nodes = append(nodes, n)
	}
	return ast.Set(nodes...), nil
}

func (j recordJSON) ToNode(m Mode) (ast.Node, error) {
	var nodes ast.Pairs
	for k, v := range j {
		n, err := v.ToNode(m)
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in record: %w", err)
		}
		nodes = append(nodes, ast.Pair{Key: types.String(k), Value: n})
	}
	return ast.Record(nodes), nil
}

func (e *extensionJSON) ToNode(m Mode) (ast.Node, error) {
	if _, ok := extensions.ExtMap[types.Path(e.Name)]; !ok {
		return ast.Node{}, fmt.Errorf("`%v` is not a known extension function or method", e.Name)
	}
	var argNodes []ast.Node
	for _, n := range e.Args {
		node, err := n.ToNode(m)
		if err != nil {
			return ast.Node{}, fmt.Errorf("error in extension arg: %w", err)
		}
		argNodes = append(argNodes, node)
	}
	return ast.NewExtensionCall(types.Path(e.Name), argNodes...), nil
}

func (e *valueJSON) ToNode(m Mode) (ast.Node, error) {
	if m != ModeStrict {
		var v types.Value
		if err := types.UnmarshalJSON(e.raw, &v); err == nil {
			return ast.Value(v), nil
		}
	}
	return valueToNode(e.raw)
}

func valueToNode(b jsontext.Value) (ast.Node, error) {
	switch b.Kind() {
	case '[':
		var elems []jsontext.Value
		if err := json.Unmarshal(b, &elems); err != nil {
			return ast.Node{}, err
		}
		nodes := make([]ast.Node, len(elems))
		for i, e := range elems {
			n, err := valueToNode(e)
			if err != nil {
				return ast.Node{}, err
			}
			nodes[i] = n
		}
		return ast.Set(nodes...), nil
	case '{':
		var members map[string]jsontext.Value
		if err := json.Unmarshal(b, &members); err != nil {
			return ast.Node{}, err
		}
		if len(members) == 1 {
			if _, ok := members["__entity"]; ok {
				var v types.Value
				if err := types.UnmarshalJSON(b, &v); err != nil {
					return ast.Node{}, err
				}
				if _, ok := v.(types.EntityUID); ok {
					return ast.Value(v), nil
				}
			}
			if extn, ok := members["__extn"]; ok {
				var call struct {
					Fn  *string        `json:"fn"`
					Arg jsontext.Value `json:"arg"`
				}
				if err := json.Unmarshal(extn, &call); err == nil && call.Fn != nil && call.Arg != nil {
					arg, err := valueToNode(call.Arg)
					if err != nil {
						return ast.Node{}, err
					}
					return ast.NewExtensionCall(types.Path(*call.Fn), arg), nil
				}
			}
		}
		var pairs ast.Pairs
		for _, k := range slices.Sorted(maps.Keys(members)) {
			n, err := valueToNode(members[k])
			if err != nil {
				return ast.Node{}, err
			}
			pairs = append(pairs, ast.Pair{Key: types.String(k), Value: n})
		}
		return ast.Record(pairs), nil
	default:
		var v types.Value
		if err := types.UnmarshalJSON(b, &v); err != nil {
			return ast.Node{}, err
		}
		return ast.Value(v), nil
	}
}

func (n nodeJSON) ToNode(m Mode) (ast.Node, error) {
	switch {
	// Value
	case n.Value != nil:
		return n.Value.ToNode(m)

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
		return n.Not.ToNode(m, ast.Not)
	case n.Negate != nil:
		return n.Negate.ToNode(m, ast.Negate)

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny, hasTag, getTag
	case n.Equals != nil:
		return n.Equals.ToNode(m, ast.Node.Equal)
	case n.NotEquals != nil:
		return n.NotEquals.ToNode(m, ast.Node.NotEqual)
	case n.In != nil:
		return n.In.ToNode(m, ast.Node.In)
	case n.LessThan != nil:
		return n.LessThan.ToNode(m, ast.Node.LessThan)
	case n.LessThanOrEqual != nil:
		return n.LessThanOrEqual.ToNode(m, ast.Node.LessThanOrEqual)
	case n.GreaterThan != nil:
		return n.GreaterThan.ToNode(m, ast.Node.GreaterThan)
	case n.GreaterThanOrEqual != nil:
		return n.GreaterThanOrEqual.ToNode(m, ast.Node.GreaterThanOrEqual)
	case n.And != nil:
		return n.And.ToNode(m, ast.Node.And)
	case n.Or != nil:
		return n.Or.ToNode(m, ast.Node.Or)
	case n.Add != nil:
		return n.Add.ToNode(m, ast.Node.Add)
	case n.Subtract != nil:
		return n.Subtract.ToNode(m, ast.Node.Subtract)
	case n.Multiply != nil:
		return n.Multiply.ToNode(m, ast.Node.Multiply)
	case n.Contains != nil:
		return n.Contains.ToNode(m, ast.Node.Contains)
	case n.ContainsAll != nil:
		return n.ContainsAll.ToNode(m, ast.Node.ContainsAll)
	case n.ContainsAny != nil:
		return n.ContainsAny.ToNode(m, ast.Node.ContainsAny)
	case n.IsEmpty != nil:
		return n.IsEmpty.ToNode(m, ast.Node.IsEmpty)
	case n.GetTag != nil:
		return n.GetTag.ToNode(m, ast.Node.GetTag)
	case n.HasTag != nil:
		return n.HasTag.ToNode(m, ast.Node.HasTag)

	// ., has
	case n.Access != nil:
		return n.Access.ToNode(m, ast.Node.Access)
	case n.Has != nil:
		return n.Has.ToNode(m, ast.Node.Has)

	// is
	case n.Is != nil:
		return n.Is.ToNode(m)

	// like
	case n.Like != nil:
		return n.Like.ToNode(m, ast.Node.Like)

	// if-then-else
	case n.IfThenElse != nil:
		return n.IfThenElse.ToNode(m)

	// Set
	case n.Set != nil:
		return (*n.Set).ToNode(m)

	// Record
	case n.Record != nil:
		return (*n.Record).ToNode(m)

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	case n.ExtensionCall != nil:
		return n.ExtensionCall.ToNode(m)

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

// unmarshalNonNull decodes the next value from dec into v, rejecting null.
func unmarshalNonNull[T any](dec *jsontext.Decoder, v *T) error {
	if dec.PeekKind() == jsontext.KindNull {
		return errors.New("must not be null")
	}
	return json.UnmarshalDecode(dec, v)
}

func (a *annotationsJSON) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return unmarshalNonNull(dec, (*map[string]string)(a))
}

func (c *conditionsJSON) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if err := unmarshalNonNull(dec, (*[]conditionJSON)(c)); err != nil {
		return err
	}
	if *c == nil {
		*c = conditionsJSON{}
	}
	return nil
}

func (e *scopeEntitiesJSON) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if err := unmarshalNonNull(dec, (*[]types.ImplicitlyMarshaledEntityUID)(e)); err != nil {
		return err
	}
	if *e == nil {
		*e = scopeEntitiesJSON{}
	}
	return nil
}

func lastMembers(b []byte) ([]byte, error) {
	dec := jsontext.NewDecoder(bytes.NewReader(b), jsontext.AllowDuplicateNames(true))
	v, err := lastMembersValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.ReadToken(); err != io.EOF {
		return nil, errors.New("unexpected data after the policy")
	}
	return v, nil
}

func lastMembersValue(dec *jsontext.Decoder) (jsontext.Value, error) {
	var buf bytes.Buffer
	enc := jsontext.NewEncoder(&buf)
	switch dec.PeekKind() {
	case '{':
		if _, err := dec.ReadToken(); err != nil {
			return nil, err
		}
		var names []string
		values := map[string]jsontext.Value{}
		for dec.PeekKind() != '}' {
			tok, err := dec.ReadToken()
			if err != nil {
				return nil, err
			}
			name := tok.String()
			v, err := lastMembersValue(dec)
			if err != nil {
				return nil, err
			}
			if _, ok := values[name]; !ok {
				names = append(names, name)
			}
			values[name] = v
		}
		if _, err := dec.ReadToken(); err != nil {
			return nil, err
		}
		if err := enc.WriteToken(jsontext.BeginObject); err != nil {
			return nil, err
		}
		for _, name := range names {
			if err := enc.WriteToken(jsontext.String(name)); err != nil {
				return nil, err
			}
			if err := enc.WriteValue(values[name]); err != nil {
				return nil, err
			}
		}
		if err := enc.WriteToken(jsontext.EndObject); err != nil {
			return nil, err
		}
	case '[':
		if _, err := dec.ReadToken(); err != nil {
			return nil, err
		}
		if err := enc.WriteToken(jsontext.BeginArray); err != nil {
			return nil, err
		}
		for dec.PeekKind() != ']' {
			v, err := lastMembersValue(dec)
			if err != nil {
				return nil, err
			}
			if err := enc.WriteValue(v); err != nil {
				return nil, err
			}
		}
		if _, err := dec.ReadToken(); err != nil {
			return nil, err
		}
		if err := enc.WriteToken(jsontext.EndArray); err != nil {
			return nil, err
		}
	default:
		v, err := dec.ReadValue()
		if err != nil {
			return nil, err
		}
		return v.Clone(), nil
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// UnmarshalJSON decodes a policy, rejecting unknown members everywhere. Rust
// Cedar also rejects them in the policy object, its conditions, expressions,
// and "is" scopes, but ignores them in its other scopes and in the "in" object
// of an "is" scope, e.g. {"op":"All","x":1}. That looks like a side effect of
// how serde applies deny_unknown_fields to enums rather than a deliberate
// choice, so cedar-go is stricter there.
func (p *Policy) UnmarshalJSON(b []byte) error {
	return p.UnmarshalJSONMode(b, ModeGo)
}

func (p *Policy) UnmarshalJSONMode(b []byte, m Mode) error {
	if jsontext.Value(b).Kind() == jsontext.KindNull {
		return errors.New("policy must not be null")
	}
	if m != ModeGoLegacy {
		var err error
		if b, err = lastMembers(b); err != nil {
			return fmt.Errorf("error unmarshalling json: %w", err)
		}
	}
	var j policyJSON
	if err := json.Unmarshal(b, &j, json.RejectUnknownMembers(true)); err != nil {
		return fmt.Errorf("error unmarshalling json: %w", err)
	}
	if j.Conditions == nil {
		return errors.New(`missing "conditions"`)
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
		if !scanner.IsAnyIdent(k) {
			return fmt.Errorf("invalid annotation key: %q", k)
		}
		p.unwrap().Annotate(types.AnnotationKey(k), types.String(v))
	}
	var err error
	p.Principal, err = j.Principal.ToPrincipalResourceNode()
	if err != nil {
		return fmt.Errorf("error in principal: %w", err)
	}
	p.Action, err = j.Action.ToActionNode(m)
	if err != nil {
		return fmt.Errorf("error in action: %w", err)
	}
	p.Resource, err = j.Resource.ToPrincipalResourceNode()
	if err != nil {
		return fmt.Errorf("error in resource: %w", err)
	}
	for _, c := range j.Conditions {
		n, err := c.Body.ToNode(m)
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
