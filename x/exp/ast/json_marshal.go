package ast

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func (s *scopeJSON) FromNode(src Node) error {
	switch src.nodeType {
	case nodeTypeNone:
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
		n := unaryNode(src)
		j.Not = &unaryJSON{}
		j.Not.Arg.FromNode(n.Arg())
		return nil
	case nodeTypeNegate:
		n := unaryNode(src)
		j.Negate = &unaryJSON{}
		j.Negate.Arg.FromNode(n.Arg())
		return nil

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	case nodeTypeAdd:
		n := binaryNode(src)
		j.Plus = &binaryJSON{}
		j.Plus.Left.FromNode(n.Left())
		j.Plus.Right.FromNode(n.Right())
		return nil
	case nodeTypeAnd:
		n := binaryNode(src)
		j.And = &binaryJSON{}
		j.And.Left.FromNode(n.Left())
		j.And.Right.FromNode(n.Right())
		return nil
	case nodeTypeContains:
		n := binaryNode(src)
		j.Contains = &binaryJSON{}
		j.Contains.Left.FromNode(n.Left())
		j.Contains.Right.FromNode(n.Right())
		return nil
	case nodeTypeContainsAll:
		n := binaryNode(src)
		j.ContainsAll = &binaryJSON{}
		j.ContainsAll.Left.FromNode(n.Left())
		j.ContainsAll.Right.FromNode(n.Right())
		return nil
	case nodeTypeContainsAny:
		n := binaryNode(src)
		j.ContainsAny = &binaryJSON{}
		j.ContainsAny.Left.FromNode(n.Left())
		j.ContainsAny.Right.FromNode(n.Right())
		return nil
	case nodeTypeEquals:
		n := binaryNode(src)
		j.Equals = &binaryJSON{}
		j.Equals.Left.FromNode(n.Left())
		j.Equals.Right.FromNode(n.Right())
		return nil
	case nodeTypeGreater:
		n := binaryNode(src)
		j.GreaterThan = &binaryJSON{}
		j.GreaterThan.Left.FromNode(n.Left())
		j.GreaterThan.Right.FromNode(n.Right())
		return nil
	case nodeTypeGreaterEqual:
		n := binaryNode(src)
		j.GreaterThanOrEqual = &binaryJSON{}
		j.GreaterThanOrEqual.Left.FromNode(n.Left())
		j.GreaterThanOrEqual.Right.FromNode(n.Right())
		return nil
	case nodeTypeIn:
		n := binaryNode(src)
		j.In = &binaryJSON{}
		j.In.Left.FromNode(n.Left())
		j.In.Right.FromNode(n.Right())
		return nil
	case nodeTypeLess:
		n := binaryNode(src)
		j.LessThan = &binaryJSON{}
		j.LessThan.Left.FromNode(n.Left())
		j.LessThan.Right.FromNode(n.Right())
		return nil
	case nodeTypeLessEqual:
		n := binaryNode(src)
		j.LessThanOrEqual = &binaryJSON{}
		j.LessThanOrEqual.Left.FromNode(n.Left())
		j.LessThanOrEqual.Right.FromNode(n.Right())
		return nil
	case nodeTypeMult:
		n := binaryNode(src)
		j.Times = &binaryJSON{}
		j.Times.Left.FromNode(n.Left())
		j.Times.Right.FromNode(n.Right())
		return nil
	case nodeTypeNotEquals:
		n := binaryNode(src)
		j.NotEquals = &binaryJSON{}
		j.NotEquals.Left.FromNode(n.Left())
		j.NotEquals.Right.FromNode(n.Right())
		return nil
	case nodeTypeOr:
		n := binaryNode(src)
		j.Or = &binaryJSON{}
		j.Or.Left.FromNode(n.Left())
		j.Or.Right.FromNode(n.Right())
		return nil
	case nodeTypeSub:
		n := binaryNode(src)
		j.Minus = &binaryJSON{}
		j.Minus.Left.FromNode(n.Left()) // TODO: in all these cases, check for an error, handle it ...
		j.Minus.Right.FromNode(n.Right())
		return nil

	// ., has
	// Access *strJSON `json:"."`
	// Has    *strJSON `json:"has"`
	case nodeTypeAccess:
		n := binaryNode(src)
		j.Access = &strJSON{}
		j.Access.Left.FromNode(n.Left())
		j.Access.Attr = n.Right().value.String() // TODO: make this nicer
		return nil
	case nodeTypeHas:
		n := binaryNode(src)
		j.Has = &strJSON{}
		j.Has.Left.FromNode(n.Left())
		j.Has.Attr = n.Right().value.String() // TODO: make this nicer
		return nil
	// is
	case nodeTypeIs:
		n := binaryNode(src)
		j.Is = &isJSON{
			EntityType: string(n.Right().value.(types.String)), // TODO: make this nicer
		}
		j.Is.Left.FromNode(n.Left())
		return nil
	case nodeTypeIsIn:
		n := trinaryNode(src)
		j.Is = &isJSON{
			EntityType: string(n.B().value.(types.String)), // TODO: make this nicer
			In:         &inJSON{},
		}
		j.Is.Left.FromNode(n.A())
		j.Is.In.Entity = n.C().value.(types.EntityUID)
		return nil

	// like
	// Like *strJSON `json:"like"`
	case nodeTypeLike:
		n := binaryNode(src)
		j.Like = &strJSON{}
		j.Like.Left.FromNode(n.Left())
		j.Like.Attr = n.Right().value.String() // TODO: make this nicer

	// if-then-else
	// IfThenElse *ifThenElseJSON `json:"if-then-else"`
	case nodeTypeIf:
		n := trinaryNode(src)
		j.IfThenElse = &ifThenElseJSON{}
		j.IfThenElse.If.FromNode(n.A())
		j.IfThenElse.Then.FromNode(n.B())
		j.IfThenElse.Else.FromNode(n.C())

	// Set
	// Set arrayJSON `json:"Set"`
	case nodeTypeSet:
		j.Set = arrayJSON{}
		for _, v := range src.args {
			var nn nodeJSON
			if err := nn.FromNode(v); err != nil {
				return err
			}
			j.Set = append(j.Set, nn)
		}
		return nil

	// Record
	// Record recordJSON `json:"Record"`
	case nodeTypeRecord:
		j.Record = recordJSON{}
		for _, kv := range src.args {
			n := binaryNode(kv)
			// TODO: make this nicer
			var nn nodeJSON
			if err := nn.FromNode(n.Right()); err != nil {
				return err
			}
			j.Record[n.Left().value.String()] = nn
		}
		return nil

	// Any other function: decimal, ip
	// Decimal arrayJSON `json:"decimal"`
	// IP      arrayJSON `json:"ip"`
	case nodeTypeDecimal:
		j.Decimal = arrayJSON{}
		str := src.value.String() // TODO: make this nicer
		b := []byte(str)
		j.Decimal = append(j.Decimal, nodeJSON{
			Value: (*json.RawMessage)(&b),
		},
		)
		return nil

	case nodeTypeIpAddr:
		j.IP = arrayJSON{}
		str := src.value.String() // TODO: make this nicer
		b := []byte(str)
		j.IP = append(j.IP, nodeJSON{
			Value: (*json.RawMessage)(&b),
		},
		)
		return nil

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
		n := binaryNode(src)
		j.LessThanExt = arrayJSON{}
		var left, right nodeJSON
		if err := left.FromNode(n.Left()); err != nil {
			return err
		}
		if err := right.FromNode(n.Right()); err != nil {
			return err
		}
		j.LessThanExt = append(j.LessThanExt, left, right)
		return nil
	case nodeTypeLessEqualExt:
		n := binaryNode(src)
		j.LessThanOrEqualExt = arrayJSON{}
		var left, right nodeJSON
		if err := left.FromNode(n.Left()); err != nil {
			return err
		}
		if err := right.FromNode(n.Right()); err != nil {
			return err
		}
		j.LessThanOrEqualExt = append(j.LessThanOrEqualExt, left, right)
		return nil
	case nodeTypeGreaterExt:
		n := binaryNode(src)
		j.GreaterThanExt = arrayJSON{}
		var left, right nodeJSON
		if err := left.FromNode(n.Left()); err != nil {
			return err
		}
		if err := right.FromNode(n.Right()); err != nil {
			return err
		}
		j.GreaterThanExt = append(j.GreaterThanExt, left, right)
		return nil
	case nodeTypeGreaterEqualExt:
		n := binaryNode(src)
		j.GreaterThanOrEqualExt = arrayJSON{}
		var left, right nodeJSON
		if err := left.FromNode(n.Left()); err != nil {
			return err
		}
		if err := right.FromNode(n.Right()); err != nil {
			return err
		}
		j.GreaterThanOrEqualExt = append(j.GreaterThanOrEqualExt, left, right)
		return nil
	case nodeTypeIsInRange:
		n := binaryNode(src)
		j.IsInRangeExt = arrayJSON{}
		var left, right nodeJSON
		if err := left.FromNode(n.Left()); err != nil {
			return err
		}
		if err := right.FromNode(n.Right()); err != nil {
			return err
		}
		j.IsInRangeExt = append(j.IsInRangeExt, left, right)
		return nil

	case nodeTypeIsIpv4:
		n := unaryNode(src)
		j.IsIpv4Ext = arrayJSON{}
		var arg nodeJSON
		if err := arg.FromNode(n.Arg()); err != nil {
			return err
		}
		j.IsIpv4Ext = append(j.IsIpv4Ext, arg)
		return nil
	case nodeTypeIsIpv6:
		n := unaryNode(src)
		j.IsIpv6Ext = arrayJSON{}
		var arg nodeJSON
		if err := arg.FromNode(n.Arg()); err != nil {
			return err
		}
		j.IsIpv6Ext = append(j.IsIpv6Ext, arg)
		return nil
	case nodeTypeIsLoopback:
		n := unaryNode(src)
		j.IsLoopbackExt = arrayJSON{}
		var arg nodeJSON
		if err := arg.FromNode(n.Arg()); err != nil {
			return err
		}
		j.IsLoopbackExt = append(j.IsLoopbackExt, arg)
		return nil
	case nodeTypeIsMulticast:
		n := unaryNode(src)
		j.IsMulticastExt = arrayJSON{}
		var arg nodeJSON
		if err := arg.FromNode(n.Arg()); err != nil {
			return err
		}
		j.IsMulticastExt = append(j.IsMulticastExt, arg)
		return nil

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
