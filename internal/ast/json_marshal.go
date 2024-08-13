package ast

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func (s *scopeJSON) FromNode(src IsScopeNode) error {
	switch t := src.(type) {
	case ScopeTypeAll:
		s.Op = "All"
		return nil
	case ScopeTypeEq:
		s.Op = "=="
		e := t.Entity
		s.Entity = &e
		return nil
	case ScopeTypeIn:
		s.Op = "in"
		e := t.Entity
		s.Entity = &e
		return nil
	case ScopeTypeInSet:
		s.Op = "in"
		s.Entities = t.Entities
		return nil
	case ScopeTypeIs:
		s.Op = "is"
		s.EntityType = string(t.Type)
		return nil
	case ScopeTypeIsIn:
		s.Op = "is"
		s.EntityType = string(t.Type)
		s.In = &scopeInJSON{
			Entity: t.Entity,
		}
		return nil
	}
	return fmt.Errorf("unexpected scope node: %T", src)
}

func unaryToJSON(dest **unaryJSON, src UnaryNode) error {
	n := UnaryNode(src)
	res := &unaryJSON{}
	if err := res.Arg.FromNode(n.Arg); err != nil {
		return fmt.Errorf("error in arg: %w", err)
	}
	*dest = res
	return nil
}

func binaryToJSON(dest **binaryJSON, src BinaryNode) error {
	n := BinaryNode(src)
	res := &binaryJSON{}
	if err := res.Left.FromNode(n.Left); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	if err := res.Right.FromNode(n.Right); err != nil {
		return fmt.Errorf("error in right: %w", err)
	}
	*dest = res
	return nil
}

func arrayToJSON(dest *arrayJSON, args []IsNode) error {
	res := arrayJSON{}
	for _, n := range args {
		var nn nodeJSON
		if err := nn.FromNode(n); err != nil {
			return fmt.Errorf("error in array: %w", err)
		}
		res = append(res, nn)
	}
	*dest = res
	return nil
}

func extToJSON(dest *extensionCallJSON, name string, src types.Value) error {
	res := arrayJSON{}
	str := src.String()               // TODO: is this the correct string?
	b, _ := json.Marshal(string(str)) // error impossible
	res = append(res, nodeJSON{
		Value: (*json.RawMessage)(&b),
	})
	*dest = extensionCallJSON{
		name: res,
	}
	return nil
}

func extCallToJSON(dest extensionCallJSON, src NodeTypeExtensionCall) error {
	jsonArgs := arrayJSON{}
	for _, n := range src.Args {
		argNode := &nodeJSON{}
		err := argNode.FromNode(n)
		if err != nil {
			return err
		}
		jsonArgs = append(jsonArgs, *argNode)
	}
	dest[string(src.Name)] = jsonArgs
	return nil
}

func strToJSON(dest **strJSON, src StrOpNode) error {
	res := &strJSON{}
	if err := res.Left.FromNode(src.Arg); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	res.Attr = string(src.Value)
	*dest = res
	return nil
}

func patternToJSON(dest **patternJSON, src NodeTypeLike) error {
	res := &patternJSON{}
	if err := res.Left.FromNode(src.Arg); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	for _, comp := range src.Value.Components {
		if comp.Wildcard {
			res.Pattern = append(res.Pattern, patternComponentJSON{Wildcard: true})
		}
		if comp.Literal != "" {
			res.Pattern = append(res.Pattern, patternComponentJSON{Literal: patternComponentLiteralJSON{Literal: comp.Literal}})
		}
	}
	*dest = res
	return nil
}

func recordToJSON(dest *recordJSON, src NodeTypeRecord) error {
	res := recordJSON{}
	for _, kv := range src.Elements {
		var nn nodeJSON
		if err := nn.FromNode(kv.Value); err != nil {
			return err
		}
		res[string(kv.Key)] = nn
	}
	*dest = res
	return nil
}

func ifToJSON(dest **ifThenElseJSON, src NodeTypeIf) error {
	res := &ifThenElseJSON{}
	if err := res.If.FromNode(src.If); err != nil {
		return fmt.Errorf("error in if: %w", err)
	}
	if err := res.Then.FromNode(src.Then); err != nil {
		return fmt.Errorf("error in then: %w", err)
	}
	if err := res.Else.FromNode(src.Else); err != nil {
		return fmt.Errorf("error in else: %w", err)
	}
	*dest = res
	return nil
}

func isToJSON(dest **isJSON, src NodeTypeIs) error {
	res := &isJSON{}
	if err := res.Left.FromNode(src.Left); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	res.EntityType = string(src.EntityType)
	*dest = res
	return nil
}

func isInToJSON(dest **isJSON, src NodeTypeIsIn) error {
	res := &isJSON{}
	if err := res.Left.FromNode(src.Left); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	res.EntityType = string(src.EntityType)
	res.In = &nodeJSON{}
	if err := res.In.FromNode(src.Entity); err != nil {
		return fmt.Errorf("error in entity: %w", err)
	}
	*dest = res
	return nil
}

func (j *nodeJSON) FromNode(src IsNode) error {
	switch t := src.(type) {
	// Value
	// Value *json.RawMessage `json:"Value"` // could be any
	case NodeValue:
		// Any other function: decimal, ip
		// Decimal arrayJSON `json:"decimal"`
		// IP      arrayJSON `json:"ip"`
		switch tt := t.Value.(type) {
		case types.Decimal:
			return extToJSON(&j.ExtensionCall, "decimal", tt)
		case types.IPAddr:
			return extToJSON(&j.ExtensionCall, "ip", tt)
		}
		b, err := t.Value.ExplicitMarshalJSON()
		j.Value = (*json.RawMessage)(&b)
		return err

	// Var
	// Var *string `json:"Var"`
	case NodeTypeVariable:
		val := string(t.Name)
		j.Var = &val
		return nil

	// ! or neg operators
	// Not    *unaryJSON `json:"!"`
	// Negate *unaryJSON `json:"neg"`
	case NodeTypeNot:
		return unaryToJSON(&j.Not, t.UnaryNode)
	case NodeTypeNegate:
		return unaryToJSON(&j.Negate, t.UnaryNode)

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	case NodeTypeAdd:
		return binaryToJSON(&j.Plus, t.BinaryNode)
	case NodeTypeAnd:
		return binaryToJSON(&j.And, t.BinaryNode)
	case NodeTypeContains:
		return binaryToJSON(&j.Contains, t.BinaryNode)
	case NodeTypeContainsAll:
		return binaryToJSON(&j.ContainsAll, t.BinaryNode)
	case NodeTypeContainsAny:
		return binaryToJSON(&j.ContainsAny, t.BinaryNode)
	case NodeTypeEquals:
		return binaryToJSON(&j.Equals, t.BinaryNode)
	case NodeTypeGreaterThan:
		return binaryToJSON(&j.GreaterThan, t.BinaryNode)
	case NodeTypeGreaterThanOrEqual:
		return binaryToJSON(&j.GreaterThanOrEqual, t.BinaryNode)
	case NodeTypeIn:
		return binaryToJSON(&j.In, t.BinaryNode)
	case NodeTypeLessThan:
		return binaryToJSON(&j.LessThan, t.BinaryNode)
	case NodeTypeLessThanOrEqual:
		return binaryToJSON(&j.LessThanOrEqual, t.BinaryNode)
	case NodeTypeMult:
		return binaryToJSON(&j.Times, t.BinaryNode)
	case NodeTypeNotEquals:
		return binaryToJSON(&j.NotEquals, t.BinaryNode)
	case NodeTypeOr:
		return binaryToJSON(&j.Or, t.BinaryNode)
	case NodeTypeSub:
		return binaryToJSON(&j.Minus, t.BinaryNode)

	// ., has
	// Access *strJSON `json:"."`
	// Has    *strJSON `json:"has"`
	case NodeTypeAccess:
		return strToJSON(&j.Access, t.StrOpNode)
	case NodeTypeHas:
		return strToJSON(&j.Has, t.StrOpNode)
	// is
	case NodeTypeIs:
		return isToJSON(&j.Is, t)
	case NodeTypeIsIn:
		return isInToJSON(&j.Is, t)

	// like
	// Like *strJSON `json:"like"`
	case NodeTypeLike:
		return patternToJSON(&j.Like, t)

	// if-then-else
	// IfThenElse *ifThenElseJSON `json:"if-then-else"`
	case NodeTypeIf:
		return ifToJSON(&j.IfThenElse, t)

	// Set
	// Set arrayJSON `json:"Set"`
	case NodeTypeSet:
		return arrayToJSON(&j.Set, t.Elements)

	// Record
	// Record recordJSON `json:"Record"`
	case NodeTypeRecord:
		return recordToJSON(&j.Record, t)

	// Any other method: ip, decimal, lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	// ExtensionMethod map[string]arrayJSON `json:"-"`
	case NodeTypeExtensionCall:
		j.ExtensionCall = extensionCallJSON{}
		return extCallToJSON(j.ExtensionCall, t)
	}
	// case nodeTypeRecordEntry:
	// case nodeTypeEntityType:
	// case nodeTypeAnnotation:
	// case nodeTypeWhen:
	// case nodeTypeUnless:
	return fmt.Errorf("unknown node type: %T", src)
}

func (j *nodeJSON) MarshalJSON() ([]byte, error) {
	if len(j.ExtensionCall) > 0 {
		return json.Marshal(j.ExtensionCall)
	}

	type nodeJSONAlias nodeJSON
	return json.Marshal((*nodeJSONAlias)(j))
}

func (p *patternComponentJSON) MarshalJSON() ([]byte, error) {
	if p.Wildcard {
		return json.Marshal("Wildcard")
	}
	return json.Marshal(p.Literal)
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
	if err := j.Principal.FromNode(p.Principal); err != nil {
		return nil, fmt.Errorf("error in principal: %w", err)
	}
	if err := j.Action.FromNode(p.Action); err != nil {
		return nil, fmt.Errorf("error in action: %w", err)
	}
	if err := j.Resource.FromNode(p.Resource); err != nil {
		return nil, fmt.Errorf("error in resource: %w", err)
	}
	for _, c := range p.Conditions {
		var cond conditionJSON
		cond.Kind = "when"
		if c.Condition == ConditionUnless {
			cond.Kind = "unless"
		}
		if err := cond.Body.FromNode(c.Body); err != nil {
			return nil, fmt.Errorf("error in condition: %w", err)
		}
		j.Conditions = append(j.Conditions, cond)
	}
	return json.Marshal(j)
}
