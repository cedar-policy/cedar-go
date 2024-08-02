package ast

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func (s *scopeJSON) FromNode(src isScopeNode) error {
	switch t := src.(type) {
	case scopeTypeAll:
		s.Op = "All"
		return nil
	case scopeTypeEq:
		s.Op = "=="
		e := t.Entity
		s.Entity = &e
		return nil
	case scopeTypeIn:
		s.Op = "in"
		e := t.Entity
		s.Entity = &e
		return nil
	case scopeTypeInSet:
		s.Op = "in"
		s.Entities = t.Entities
		return nil
	case scopeTypeIs:
		s.Op = "is"
		s.EntityType = string(t.Type)
		return nil
	case scopeTypeIsIn:
		s.Op = "is"
		s.EntityType = string(t.Type)
		s.In = &scopeInJSON{
			Entity: t.Entity,
		}
		return nil
	}
	return fmt.Errorf("unexpected scope node: %T", src)
}

func unaryToJSON(dest **unaryJSON, src unaryNode) error {
	n := unaryNode(src)
	res := &unaryJSON{}
	if err := res.Arg.FromNode(n.Arg); err != nil {
		return fmt.Errorf("error in arg: %w", err)
	}
	*dest = res
	return nil
}

func binaryToJSON(dest **binaryJSON, src binaryNode) error {
	n := binaryNode(src)
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

func arrayToJSON(dest *arrayJSON, args []node) error {
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

func extToJSON(dest *arrayJSON, src types.Value) error {
	res := arrayJSON{}
	str := src.String()               // TODO: is this the correct string?
	b, _ := json.Marshal(string(str)) // error impossible
	res = append(res, nodeJSON{
		Value: (*json.RawMessage)(&b),
	})
	*dest = res
	return nil
}

func extMethodToJSON(dest extMethodCallJSON, src nodeTypeExtMethodCall) error {
	objectNode := &nodeJSON{}
	err := objectNode.FromNode(src.Left)
	if err != nil {
		return err
	}
	jsonArgs := arrayJSON{*objectNode}
	for _, n := range src.Args {
		argNode := &nodeJSON{}
		err := argNode.FromNode(n)
		if err != nil {
			return err
		}
		jsonArgs = append(jsonArgs, *argNode)
	}
	dest[string(src.Method)] = jsonArgs
	return nil
}

func strToJSON(dest **strJSON, src strOpNode) error {
	res := &strJSON{}
	if err := res.Left.FromNode(src.Arg); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	res.Attr = string(src.Value)
	*dest = res
	return nil
}

func patternToJSON(dest **patternJSON, src strOpNode) error {
	res := &patternJSON{}
	if err := res.Left.FromNode(src.Arg); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	res.Pattern = string(src.Value)
	*dest = res
	return nil
}

func recordToJSON(dest *recordJSON, src nodeTypeRecord) error {
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

func ifToJSON(dest **ifThenElseJSON, src nodeTypeIf) error {
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

func isToJSON(dest **isJSON, src nodeTypeIs) error {
	res := &isJSON{}
	if err := res.Left.FromNode(src.Left); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	res.EntityType = string(src.EntityType)
	*dest = res
	return nil
}

func isInToJSON(dest **isJSON, src nodeTypeIsIn) error {
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

func (j *nodeJSON) FromNode(src node) error {
	switch t := src.(type) {
	// Value
	// Value *json.RawMessage `json:"Value"` // could be any
	case nodeValue:
		// Any other function: decimal, ip
		// Decimal arrayJSON `json:"decimal"`
		// IP      arrayJSON `json:"ip"`
		switch tt := t.Value.(type) {
		case types.Decimal:
			return extToJSON(&j.Decimal, tt)
		case types.IPAddr:
			return extToJSON(&j.IP, tt)
		}
		b, err := t.Value.ExplicitMarshalJSON()
		j.Value = (*json.RawMessage)(&b)
		return err

	// Var
	// Var *string `json:"Var"`
	case nodeTypeVariable:
		val := string(t.Name)
		j.Var = &val
		return nil

	// ! or neg operators
	// Not    *unaryJSON `json:"!"`
	// Negate *unaryJSON `json:"neg"`
	case nodeTypeNot:
		return unaryToJSON(&j.Not, t.unaryNode)
	case nodeTypeNegate:
		return unaryToJSON(&j.Negate, t.unaryNode)

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	case nodeTypeAdd:
		return binaryToJSON(&j.Plus, t.binaryNode)
	case nodeTypeAnd:
		return binaryToJSON(&j.And, t.binaryNode)
	case nodeTypeContains:
		return binaryToJSON(&j.Contains, t.binaryNode)
	case nodeTypeContainsAll:
		return binaryToJSON(&j.ContainsAll, t.binaryNode)
	case nodeTypeContainsAny:
		return binaryToJSON(&j.ContainsAny, t.binaryNode)
	case nodeTypeEquals:
		return binaryToJSON(&j.Equals, t.binaryNode)
	case nodeTypeGreaterThan:
		return binaryToJSON(&j.GreaterThan, t.binaryNode)
	case nodeTypeGreaterThanOrEqual:
		return binaryToJSON(&j.GreaterThanOrEqual, t.binaryNode)
	case nodeTypeIn:
		return binaryToJSON(&j.In, t.binaryNode)
	case nodeTypeLessThan:
		return binaryToJSON(&j.LessThan, t.binaryNode)
	case nodeTypeLessThanOrEqual:
		return binaryToJSON(&j.LessThanOrEqual, t.binaryNode)
	case nodeTypeMult:
		return binaryToJSON(&j.Times, t.binaryNode)
	case nodeTypeNotEquals:
		return binaryToJSON(&j.NotEquals, t.binaryNode)
	case nodeTypeOr:
		return binaryToJSON(&j.Or, t.binaryNode)
	case nodeTypeSub:
		return binaryToJSON(&j.Minus, t.binaryNode)

	// ., has
	// Access *strJSON `json:"."`
	// Has    *strJSON `json:"has"`
	case nodeTypeAccess:
		return strToJSON(&j.Access, t.strOpNode)
	case nodeTypeHas:
		return strToJSON(&j.Has, t.strOpNode)
	// is
	case nodeTypeIs:
		return isToJSON(&j.Is, t)
	case nodeTypeIsIn:
		return isInToJSON(&j.Is, t)

	// like
	// Like *strJSON `json:"like"`
	case nodeTypeLike:
		return patternToJSON(&j.Like, t.strOpNode)

	// if-then-else
	// IfThenElse *ifThenElseJSON `json:"if-then-else"`
	case nodeTypeIf:
		return ifToJSON(&j.IfThenElse, t)

	// Set
	// Set arrayJSON `json:"Set"`
	case nodeTypeSet:
		return arrayToJSON(&j.Set, t.Elements)

	// Record
	// Record recordJSON `json:"Record"`
	case nodeTypeRecord:
		return recordToJSON(&j.Record, t)

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	// ExtensionMethod map[string]arrayJSON `json:"-"`
	case nodeTypeExtMethodCall:
		j.ExtensionMethod = extMethodCallJSON{}
		return extMethodToJSON(j.ExtensionMethod, t)
	}
	// case nodeTypeRecordEntry:
	// case nodeTypeEntityType:
	// case nodeTypeAnnotation:
	// case nodeTypeWhen:
	// case nodeTypeUnless:
	return fmt.Errorf("unknown node type: %T", src)
}

func (j *nodeJSON) MarshalJSON() ([]byte, error) {
	if len(j.ExtensionMethod) > 0 {
		return json.Marshal(j.ExtensionMethod)
	}

	type nodeJSONAlias nodeJSON
	return json.Marshal((*nodeJSONAlias)(j))
}

func (p *Policy) MarshalJSON() ([]byte, error) {
	var j policyJSON
	j.Effect = "forbid"
	if p.effect {
		j.Effect = "permit"
	}
	if len(p.annotations) > 0 {
		j.Annotations = map[string]string{}
	}
	for _, a := range p.annotations {
		j.Annotations[string(a.Key)] = string(a.Value)
	}
	if err := j.Principal.FromNode(p.principal); err != nil {
		return nil, fmt.Errorf("error in principal: %w", err)
	}
	if err := j.Action.FromNode(p.action); err != nil {
		return nil, fmt.Errorf("error in action: %w", err)
	}
	if err := j.Resource.FromNode(p.resource); err != nil {
		return nil, fmt.Errorf("error in resource: %w", err)
	}
	for _, c := range p.conditions {
		var cond conditionJSON
		cond.Kind = "when"
		if c.Condition == conditionUnless {
			cond.Kind = "unless"
		}
		if err := cond.Body.FromNode(c.Body); err != nil {
			return nil, fmt.Errorf("error in condition: %w", err)
		}
		j.Conditions = append(j.Conditions, cond)
	}
	return json.Marshal(j)
}
