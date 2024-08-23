package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

func bakePolicy(p *ast.Policy) *ast.Policy {
	if len(p.Conditions) == 0 {
		return p
	}
	p2 := *p
	p2.Conditions = make([]ast.ConditionType, len(p.Conditions))
	for i, c := range p.Conditions {
		p2.Conditions[i] = ast.ConditionType{Condition: c.Condition, Body: bake(c.Body)}
	}
	return &p2
}

// fold takes in an ast.Node and finds all the Sets, Records, IP/Decimal extensions that can be pre-baked as values.
// It does not attempt to do any further calculation.  The returned node should serialize to JSON and Cedar
// exactly the same as the input, with possible changes in ordering of Records.
func bake(n ast.IsNode) ast.IsNode {
	switch v := n.(type) {
	case ast.NodeTypeAccess:
		return ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{Arg: bake(v.Arg), Value: v.Value}}
	case ast.NodeTypeHas:
		return ast.NodeTypeHas{StrOpNode: ast.StrOpNode{Arg: bake(v.Arg), Value: v.Value}}
	case ast.NodeTypeLike:
		return ast.NodeTypeLike{Arg: bake(v.Arg), Value: v.Value}
	case ast.NodeTypeIfThenElse:
		return ast.NodeTypeIfThenElse{If: bake(v.If), Then: bake(v.Then), Else: bake(v.Else)}
	case ast.NodeTypeIs:
		return ast.NodeTypeIs{Left: bake(v.Left), EntityType: v.EntityType}
	case ast.NodeTypeIsIn:
		return ast.NodeTypeIsIn{NodeTypeIs: ast.NodeTypeIs{Left: v.Left, EntityType: v.EntityType}, Entity: bake(v.Entity)}
	case ast.NodeTypeExtensionCall:
		switch {
		case v.Name == "ip" && len(v.Args) == 1:
			arg := bake(v.Args[0])
			if v, ok := arg.(ast.NodeValue); ok {
				if vv, ok := v.Value.(types.String); ok {
					ip, err := types.ParseIPAddr(string(vv))
					if err == nil {
						return ast.NodeValue{Value: ip}
					}
				}
			}
		case v.Name == "decimal" && len(v.Args) == 1:
			arg := bake(v.Args[0])
			if v, ok := arg.(ast.NodeValue); ok {
				if vv, ok := v.Value.(types.String); ok {
					dec, err := types.ParseDecimal(string(vv))
					if err == nil {
						return ast.NodeValue{Value: dec}
					}
				}
			}
		}

		args := make([]ast.IsNode, len(v.Args))
		for i, arg := range v.Args {
			args[i] = bake(arg)
		}
		return ast.NodeTypeExtensionCall{Name: v.Name, Args: args}
	case ast.NodeValue:
		// switch t := v.Value.(type) {
		// case types.Record:
		// 	return ast.NodeValue{Value: maps.Clone(t)}
		// case types.Set:
		// 	return ast.NodeValue{Value: slices.Clone(t)}
		// default:
		// 	return ast.NodeValue{Value: t}
		// }
		// return ast.NodeValue{Value: v.Value}
		return n
	case ast.NodeTypeRecord:
		elements := make([]ast.RecordElementNode, len(v.Elements))
		record := make(types.Record, len(v.Elements))
		ok := true
		for i, pair := range v.Elements {
			elements[i] = ast.RecordElementNode{Key: pair.Key, Value: bake(pair.Value)}
			if !ok {
				continue
			}
			if v, vok := elements[i].Value.(ast.NodeValue); vok {
				record[pair.Key] = v.Value
				continue
			}
			ok = false
		}
		if ok {
			return ast.NodeValue{Value: record}
		}
		return ast.NodeTypeRecord{Elements: elements}
	case ast.NodeTypeSet:
		elements := make([]ast.IsNode, len(v.Elements))
		set := make(types.Set, len(v.Elements))
		ok := true
		for i, item := range v.Elements {
			elements[i] = bake(item)
			if !ok {
				continue
			}
			if v, vok := elements[i].(ast.NodeValue); vok {
				set[i] = v.Value
				continue
			}
			ok = false
		}
		if ok {
			return ast.NodeValue{Value: set}
		}
		return ast.NodeTypeSet{Elements: elements}
	case ast.NodeTypeNegate:
		return ast.NodeTypeNegate{UnaryNode: ast.UnaryNode{Arg: bake(v.Arg)}}
	case ast.NodeTypeNot:
		return ast.NodeTypeNot{UnaryNode: ast.UnaryNode{Arg: bake(v.Arg)}}
	case ast.NodeTypeVariable:
		// return ast.NodeTypeVariable{Name: v.Name}
		return n
	case ast.NodeTypeIn:
		return ast.NodeTypeIn{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeAnd:
		return ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeOr:
		return ast.NodeTypeOr{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeEquals:
		return ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeNotEquals:
		return ast.NodeTypeNotEquals{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeGreaterThan:
		return ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeGreaterThanOrEqual:
		return ast.NodeTypeGreaterThanOrEqual{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeLessThan:
		return ast.NodeTypeLessThan{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeLessThanOrEqual:
		return ast.NodeTypeLessThanOrEqual{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeSub:
		return ast.NodeTypeSub{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeAdd:
		return ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeMult:
		return ast.NodeTypeMult{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeContains:
		return ast.NodeTypeContains{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeContainsAll:
		return ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	case ast.NodeTypeContainsAny:
		return ast.NodeTypeContainsAny{BinaryNode: ast.BinaryNode{Left: bake(v.Left), Right: bake(v.Right)}}
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}
