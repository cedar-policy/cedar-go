package json

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func (s *scopeJSON) FromNode(src ast.IsScopeNode) {
	switch t := src.(type) {
	case ast.ScopeTypeAll:
		s.Op = "All"
		return
	case ast.ScopeTypeEq:
		s.Op = "=="
		e := types.ImplicitlyMarshaledEntityUID(t.Entity)
		s.Entity = &e
		return
	case ast.ScopeTypeIn:
		s.Op = "in"
		e := types.ImplicitlyMarshaledEntityUID(t.Entity)
		s.Entity = &e
		return
	case ast.ScopeTypeInSet:
		s.Op = "in"
		es := make([]types.ImplicitlyMarshaledEntityUID, len(t.Entities))
		for i, e := range t.Entities {
			es[i] = types.ImplicitlyMarshaledEntityUID(e)
		}
		s.Entities = es
		return
	case ast.ScopeTypeIs:
		s.Op = "is"
		s.EntityType = string(t.Type)
		return
	case ast.ScopeTypeIsIn:
		s.Op = "is"
		s.EntityType = string(t.Type)
		s.In = &scopeInJSON{
			Entity: types.ImplicitlyMarshaledEntityUID(t.Entity),
		}
		return
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}

func unaryToJSON(dest **unaryJSON, src ast.UnaryNode) {
	res := &unaryJSON{}
	res.Arg.FromNode(src.Arg)
	*dest = res
}

func binaryToJSON(dest **binaryJSON, src ast.BinaryNode) {
	res := &binaryJSON{}
	res.Left.FromNode(src.Left)
	res.Right.FromNode(src.Right)
	*dest = res
}

func arrayToJSON(dest *arrayJSON, args []ast.IsNode) {
	res := arrayJSON{}
	for _, n := range args {
		var nn nodeJSON
		nn.FromNode(n)
		res = append(res, nn)
	}
	*dest = res
}

func extToJSON(dest *extensionJSON, name string, src types.Value) {
	res := arrayJSON{}
	str := src.String()
	val := valueJSON{v: types.String(str)}
	res = append(res, nodeJSON{
		Value: &val,
	})
	*dest = extensionJSON{
		name: res,
	}
}

func extCallToJSON(dest extensionJSON, src ast.NodeTypeExtensionCall) {
	jsonArgs := arrayJSON{}
	arrayToJSON(&jsonArgs, src.Args)
	dest[string(src.Name)] = jsonArgs
}

func strToJSON(dest **strJSON, src ast.StrOpNode) {
	res := &strJSON{}
	res.Left.FromNode(src.Arg)
	res.Attr = string(src.Value)
	*dest = res
}

func likeToJSON(dest **likeJSON, src ast.NodeTypeLike) {
	res := &likeJSON{}
	res.Left.FromNode(src.Arg)
	res.Pattern = src.Value
	*dest = res
}

func recordToJSON(dest *recordJSON, src ast.NodeTypeRecord) {
	res := recordJSON{}
	for _, kv := range src.Elements {
		var nn nodeJSON
		nn.FromNode(kv.Value)
		res[string(kv.Key)] = nn
	}
	*dest = res
}

func ifToJSON(dest **ifThenElseJSON, src ast.NodeTypeIfThenElse) {
	res := &ifThenElseJSON{}
	res.If.FromNode(src.If)
	res.Then.FromNode(src.Then)
	res.Else.FromNode(src.Else)
	*dest = res
}

func isToJSON(dest **isJSON, src ast.NodeTypeIs) {
	res := &isJSON{}
	res.Left.FromNode(src.Left)
	res.EntityType = string(src.EntityType)
	*dest = res
}

func isInToJSON(dest **isJSON, src ast.NodeTypeIsIn) {
	res := &isJSON{}
	res.Left.FromNode(src.Left)
	res.EntityType = string(src.EntityType)
	res.In = &nodeJSON{}
	res.In.FromNode(src.Entity)
	*dest = res
}

func (n *nodeJSON) FromNode(src ast.IsNode) {
	switch t := src.(type) {
	// Value
	// Value *json.RawMessage `json:"Value"` // could be any
	case ast.NodeValue:
		// Any other function: decimal, ip
		// Decimal arrayJSON `json:"decimal"`
		// IP      arrayJSON `json:"ip"`
		switch tt := t.Value.(type) {
		case types.Decimal:
			extToJSON(&n.ExtensionCall, "decimal", tt)
			return
		case types.IPAddr:
			extToJSON(&n.ExtensionCall, "ip", tt)
			return
		}
		val := valueJSON{v: t.Value}
		n.Value = &val
		return

	// Var
	// Var *string `json:"Var"`
	case ast.NodeTypeVariable:
		val := string(t.Name)
		n.Var = &val
		return

	// ! or neg operators
	// Not    *unaryJSON `json:"!"`
	// Negate *unaryJSON `json:"neg"`
	case ast.NodeTypeNot:
		unaryToJSON(&n.Not, t.UnaryNode)
		return
	case ast.NodeTypeNegate:
		unaryToJSON(&n.Negate, t.UnaryNode)
		return

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny, hasTag, getTag
	case ast.NodeTypeAdd:
		binaryToJSON(&n.Add, t.BinaryNode)
		return
	case ast.NodeTypeAnd:
		binaryToJSON(&n.And, t.BinaryNode)
		return
	case ast.NodeTypeContains:
		binaryToJSON(&n.Contains, t.BinaryNode)
		return
	case ast.NodeTypeContainsAll:
		binaryToJSON(&n.ContainsAll, t.BinaryNode)
		return
	case ast.NodeTypeContainsAny:
		binaryToJSON(&n.ContainsAny, t.BinaryNode)
		return
	case ast.NodeTypeIsEmpty:
		unaryToJSON(&n.IsEmpty, t.UnaryNode)
		return
	case ast.NodeTypeEquals:
		binaryToJSON(&n.Equals, t.BinaryNode)
		return
	case ast.NodeTypeGreaterThan:
		binaryToJSON(&n.GreaterThan, t.BinaryNode)
		return
	case ast.NodeTypeGreaterThanOrEqual:
		binaryToJSON(&n.GreaterThanOrEqual, t.BinaryNode)
		return
	case ast.NodeTypeIn:
		binaryToJSON(&n.In, t.BinaryNode)
		return
	case ast.NodeTypeLessThan:
		binaryToJSON(&n.LessThan, t.BinaryNode)
		return
	case ast.NodeTypeLessThanOrEqual:
		binaryToJSON(&n.LessThanOrEqual, t.BinaryNode)
		return
	case ast.NodeTypeMult:
		binaryToJSON(&n.Multiply, t.BinaryNode)
		return
	case ast.NodeTypeNotEquals:
		binaryToJSON(&n.NotEquals, t.BinaryNode)
		return
	case ast.NodeTypeOr:
		binaryToJSON(&n.Or, t.BinaryNode)
		return
	case ast.NodeTypeSub:
		binaryToJSON(&n.Subtract, t.BinaryNode)
		return
	case ast.NodeTypeGetTag:
		binaryToJSON(&n.GetTag, t.BinaryNode)
		return
	case ast.NodeTypeHasTag:
		binaryToJSON(&n.HasTag, t.BinaryNode)
		return

	// ., has
	// Access *strJSON `json:"."`
	// Has    *strJSON `json:"has"`
	case ast.NodeTypeAccess:
		strToJSON(&n.Access, t.StrOpNode)
		return
	case ast.NodeTypeHas:
		strToJSON(&n.Has, t.StrOpNode)
		return
	// is
	case ast.NodeTypeIs:
		isToJSON(&n.Is, t)
		return
	case ast.NodeTypeIsIn:
		isInToJSON(&n.Is, t)
		return

	// like
	// Like *strJSON `json:"like"`
	case ast.NodeTypeLike:
		likeToJSON(&n.Like, t)
		return

	// if-then-else
	// IfThenElse *ifThenElseJSON `json:"if-then-else"`
	case ast.NodeTypeIfThenElse:
		ifToJSON(&n.IfThenElse, t)
		return

	// Set
	// Set arrayJSON `json:"Set"`
	case ast.NodeTypeSet:
		arrayToJSON(&n.Set, t.Elements)
		return

	// Record
	// Record recordJSON `json:"Record"`
	case ast.NodeTypeRecord:
		recordToJSON(&n.Record, t)
		return

	// Any other method: ip, decimal, lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	// ExtensionMethod map[string]arrayJSON `json:"-"`
	case ast.NodeTypeExtensionCall:
		n.ExtensionCall = extensionJSON{}
		extCallToJSON(n.ExtensionCall, t)
		return
	default:
		panic(fmt.Sprintf("unknown node type %T", t))

	}

}

func (n *nodeJSON) MarshalJSON() ([]byte, error) {
	if len(n.ExtensionCall) > 0 {
		return json.Marshal(n.ExtensionCall)
	}

	type nodeJSONAlias nodeJSON
	return json.Marshal((*nodeJSONAlias)(n))
}

type Policy ast.Policy

func wrapPolicy(p *ast.Policy) *Policy {
	return (*Policy)(p)
}

func (p *Policy) unwrap() *ast.Policy {
	return (*ast.Policy)(p)
}

func (p *Policy) MarshalJSON() ([]byte, error) {
	var j policyJSON
	j.Effect = "forbid"
	if p.Effect {
		j.Effect = "permit"
	}
	if len(p.Annotations) > 0 {
		j.Annotations = map[string]string{}
	}
	for _, a := range p.Annotations {
		j.Annotations[string(a.Key)] = string(a.Value)
	}
	j.Principal.FromNode(p.Principal)
	j.Action.FromNode(p.Action)
	j.Resource.FromNode(p.Resource)
	for _, c := range p.Conditions {
		var cond conditionJSON
		cond.Kind = "when"
		if c.Condition == ast.ConditionUnless {
			cond.Kind = "unless"
		}
		cond.Body.FromNode(c.Body)
		j.Conditions = append(j.Conditions, cond)
	}
	return json.Marshal(j)
}
