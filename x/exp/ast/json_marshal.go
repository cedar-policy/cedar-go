package ast

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func (s *scopeJSON) FromNode(src Node) error {
	switch src.nodeType {
	case nodeTypeAll:
		s.Op = "All"
		return nil
	case nodeTypeEquals:
		n := scopeEqNode(src)
		s.Op = "=="
		e := n.Entity()
		s.Entity = &e
		return nil
	case nodeTypeIn:
		n := scopeInNode(src)
		s.Op = "in"
		if n.IsSet() {
			s.Entities = n.Set()
		} else {
			e := n.Entity()
			s.Entity = &e
		}
		return nil
	case nodeTypeIs:
		n := scopeIsNode(src)
		s.Op = "is"
		s.EntityType = string(n.EntityType())
		return nil
	case nodeTypeIsIn: // is in
		n := scopeIsInNode(src)
		s.Op = "is"
		s.EntityType = string(n.EntityType())
		s.In = &inJSON{
			Entity: n.Entity(),
		}
		return nil
	}
	return fmt.Errorf("unexpected scope node: %v", src.nodeType)
}

func unaryToJSON(dest **unaryJSON, src Node) error {
	n := unaryNode(src)
	res := &unaryJSON{}
	if err := res.Arg.FromNode(n.Arg()); err != nil {
		return fmt.Errorf("error in arg: %w", err)
	}
	*dest = res
	return nil
}

func binaryToJSON(dest **binaryJSON, src Node) error {
	n := binaryNode(src)
	res := &binaryJSON{}
	if err := res.Left.FromNode(n.Left()); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	if err := res.Right.FromNode(n.Right()); err != nil {
		return fmt.Errorf("error in right: %w", err)
	}
	*dest = res
	return nil
}

func arrayToJSON(dest *arrayJSON, src Node) error {
	res := arrayJSON{}
	for _, n := range src.args {
		var nn nodeJSON
		if err := nn.FromNode(n); err != nil {
			return fmt.Errorf("error in array: %w", err)
		}
		res = append(res, nn)
	}
	*dest = res
	return nil
}

func extToJSON(dest *arrayJSON, src Node) error {
	res := arrayJSON{}
	if src.value == nil {
		return fmt.Errorf("missing value")
	}
	str := src.value.String()         // TODO: is this the correct string?
	b, _ := json.Marshal(string(str)) // error impossible
	res = append(res, nodeJSON{
		Value: (*json.RawMessage)(&b),
	})
	*dest = res
	return nil
}

func strToJSON(dest **strJSON, src Node) error {
	n := binaryNode(src)
	res := &strJSON{}
	if err := res.Left.FromNode(n.Left()); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	str, ok := n.Right().value.(types.String)
	if !ok {
		return fmt.Errorf("right not string")
	}
	res.Attr = string(str)
	*dest = res
	return nil
}

func patternToJSON(dest **patternJSON, src Node) error {
	n := binaryNode(src)
	res := &patternJSON{}
	if err := res.Left.FromNode(n.Left()); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	str, ok := n.Right().value.(types.String)
	if !ok {
		return fmt.Errorf("right not string")
	}
	res.Pattern = string(str)
	*dest = res
	return nil
}

func recordToJSON(dest *recordJSON, src Node) error {
	res := recordJSON{}
	for _, kv := range src.args {
		n := binaryNode(kv)
		var nn nodeJSON
		if err := nn.FromNode(n.Right()); err != nil {
			return err
		}
		str, ok := n.Left().value.(types.String)
		if !ok {
			return fmt.Errorf("left not string")
		}
		res[string(str)] = nn
	}
	*dest = res
	return nil
}

func ifToJSON(dest **ifThenElseJSON, src Node) error {
	n := trinaryNode(src)
	res := &ifThenElseJSON{}
	if err := res.If.FromNode(n.A()); err != nil {
		return fmt.Errorf("error in if: %w", err)
	}
	if err := res.Then.FromNode(n.B()); err != nil {
		return fmt.Errorf("error in then: %w", err)
	}
	if err := res.Else.FromNode(n.C()); err != nil {
		return fmt.Errorf("error in else: %w", err)
	}
	*dest = res
	return nil
}

func isToJSON(dest **isJSON, src Node) error {
	n := binaryNode(src)
	res := &isJSON{}
	if err := res.Left.FromNode(n.Left()); err != nil {
		return fmt.Errorf("error in left: %w", err)
	}
	str, ok := n.Right().value.(types.String)
	if !ok {
		return fmt.Errorf("right not a string")
	}
	res.EntityType = string(str)
	if len(src.args) == 3 {
		ent, ok := src.args[2].value.(types.EntityUID)
		if !ok {
			return fmt.Errorf("in not an entity")
		}
		res.In = &inJSON{
			Entity: ent,
		}
	}
	*dest = res
	return nil
}

func (j *nodeJSON) FromNode(src Node) error {
	switch src.nodeType {
	// Value
	// Value *json.RawMessage `json:"Value"` // could be any
	case nodeTypeBoolean, nodeTypeLong, nodeTypeString, nodeTypeEntity:
		b, err := src.value.ExplicitMarshalJSON()
		j.Value = (*json.RawMessage)(&b)
		return err

	// Var
	// Var *string `json:"Var"`
	case nodeTypeVariable:
		n := variableNode(src)
		val := string(n.String())
		j.Var = &val
		return nil

	// ! or neg operators
	// Not    *unaryJSON `json:"!"`
	// Negate *unaryJSON `json:"neg"`
	case nodeTypeNot:
		return unaryToJSON(&j.Not, src)
	case nodeTypeNegate:
		return unaryToJSON(&j.Negate, src)

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	case nodeTypeAdd:
		return binaryToJSON(&j.Plus, src)
	case nodeTypeAnd:
		return binaryToJSON(&j.And, src)
	case nodeTypeContains:
		return binaryToJSON(&j.Contains, src)
	case nodeTypeContainsAll:
		return binaryToJSON(&j.ContainsAll, src)
	case nodeTypeContainsAny:
		return binaryToJSON(&j.ContainsAny, src)
	case nodeTypeEquals:
		return binaryToJSON(&j.Equals, src)
	case nodeTypeGreater:
		return binaryToJSON(&j.GreaterThan, src)
	case nodeTypeGreaterEqual:
		return binaryToJSON(&j.GreaterThanOrEqual, src)
	case nodeTypeIn:
		return binaryToJSON(&j.In, src)
	case nodeTypeLess:
		return binaryToJSON(&j.LessThan, src)
	case nodeTypeLessEqual:
		return binaryToJSON(&j.LessThanOrEqual, src)
	case nodeTypeMult:
		return binaryToJSON(&j.Times, src)
	case nodeTypeNotEquals:
		return binaryToJSON(&j.NotEquals, src)
	case nodeTypeOr:
		return binaryToJSON(&j.Or, src)
	case nodeTypeSub:
		return binaryToJSON(&j.Minus, src)

	// ., has
	// Access *strJSON `json:"."`
	// Has    *strJSON `json:"has"`
	case nodeTypeAccess:
		return strToJSON(&j.Access, src)
	case nodeTypeHas:
		return strToJSON(&j.Has, src)
	// is
	case nodeTypeIs, nodeTypeIsIn:
		return isToJSON(&j.Is, src)

	// like
	// Like *strJSON `json:"like"`
	case nodeTypeLike:
		return patternToJSON(&j.Like, src)

	// if-then-else
	// IfThenElse *ifThenElseJSON `json:"if-then-else"`
	case nodeTypeIf:
		return ifToJSON(&j.IfThenElse, src)

	// Set
	// Set arrayJSON `json:"Set"`
	case nodeTypeSet:
		return arrayToJSON(&j.Set, src)

	// Record
	// Record recordJSON `json:"Record"`
	case nodeTypeRecord:
		return recordToJSON(&j.Record, src)

	// Any other function: decimal, ip
	// Decimal arrayJSON `json:"decimal"`
	// IP      arrayJSON `json:"ip"`
	case nodeTypeDecimal:
		return extToJSON(&j.Decimal, src)

	case nodeTypeIpAddr:
		return extToJSON(&j.IP, src)

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	// LessThanExt           arrayJSON `json:"lessThan"`
	// LessThanOrEqualExt    arrayJSON `json:"lessThanOrEqual"`
	// GreaterThanExt        arrayJSON `json:"greaterThan"`
	// GreaterThanOrEqualExt arrayJSON `json:"greaterThanOrEqual"`
	// IsIpv4Ext             arrayJSON `json:"isIpv4"`
	// IsIpv6Ext             arrayJSON `json:"isIpv6"`
	// IsLoopbackExt         arrayJSON `json:"isLoopback"`
	// IsMulticastExt        arrayJSON `json:"isMulticast"`
	// IsInRangeExt          arrayJSON `json:"isInRange"`
	case nodeTypeLessExt:
		return arrayToJSON(&j.LessThanExt, src)
	case nodeTypeLessEqualExt:
		return arrayToJSON(&j.LessThanOrEqualExt, src)
	case nodeTypeGreaterExt:
		return arrayToJSON(&j.GreaterThanExt, src)
	case nodeTypeGreaterEqualExt:
		return arrayToJSON(&j.GreaterThanOrEqualExt, src)
	case nodeTypeIsInRange:
		return arrayToJSON(&j.IsInRangeExt, src)
	case nodeTypeIsIpv4:
		return arrayToJSON(&j.IsIpv4Ext, src)
	case nodeTypeIsIpv6:
		return arrayToJSON(&j.IsIpv6Ext, src)
	case nodeTypeIsLoopback:
		return arrayToJSON(&j.IsLoopbackExt, src)
	case nodeTypeIsMulticast:
		return arrayToJSON(&j.IsMulticastExt, src)
	}
	// case nodeTypeRecordEntry:
	// case nodeTypeEntityType:
	// case nodeTypeAnnotation:
	// case nodeTypeWhen:
	// case nodeTypeUnless:
	return fmt.Errorf("unknown node type: %v", src.nodeType)
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
		n := annotationNode(a)
		j.Annotations[string(n.Key())] = string(n.Value())
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
		n := unaryNode(c)
		var cond conditionJSON
		cond.Kind = "when"
		if c.nodeType == nodeTypeUnless {
			cond.Kind = "unless"
		}
		if err := cond.Body.FromNode(n.Arg()); err != nil {
			return nil, fmt.Errorf("error in condition: %w", err)
		}
		j.Conditions = append(j.Conditions, cond)
	}
	return json.Marshal(j)
}
