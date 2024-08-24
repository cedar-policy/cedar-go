package eval

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

func partialPolicy(ctx *Context, p *ast.Policy) (policy *ast.Policy, keep bool) {
	p2 := *p
	if p2.Principal, keep = partialPrincipalScope(ctx, ctx.Principal, p2.Principal); !keep {
		return nil, false
	}
	if p2.Action, keep = partialActionScope(ctx, ctx.Action, p2.Action); !keep {
		return nil, false
	}
	if p2.Resource, keep = partialResourceScope(ctx, ctx.Resource, p2.Resource); !keep {
		return nil, false
	}
	if p.Conditions != nil { // preserve nility for test purposes
		p2.Conditions = make([]ast.ConditionType, len(p.Conditions))
		for i, c := range p.Conditions {
			p2.Conditions[i] = ast.ConditionType{Condition: c.Condition, Body: partial(c.Body)}
		}
	}
	p2.Annotations = slices.Clone(p.Annotations)
	return &p2, true
}

func partial(v ast.IsNode) ast.IsNode {
	return fold(v)
}

func partialPrincipalScope(ctx *Context, ent types.Value, scope ast.IsPrincipalScopeNode) (ast.IsPrincipalScopeNode, bool) {
	evaled, result := partialScopeEval(ctx, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialActionScope(ctx *Context, ent types.Value, scope ast.IsActionScopeNode) (ast.IsActionScopeNode, bool) {
	evaled, result := partialScopeEval(ctx, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialResourceScope(ctx *Context, ent types.Value, scope ast.IsResourceScopeNode) (ast.IsResourceScopeNode, bool) {
	evaled, result := partialScopeEval(ctx, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialScopeEval(ctx *Context, ent types.Value, in ast.IsScopeNode) (evaled bool, result bool) {
	e, ok := ent.(types.EntityUID)
	if !ok {
		return false, false
	}
	switch t := in.(type) {
	case ast.ScopeTypeAll:
		return true, true
	case ast.ScopeTypeEq:
		return true, e == t.Entity
	case ast.ScopeTypeIn:
		return true, entityInOne(ctx, e, t.Entity)
	case ast.ScopeTypeInSet:
		set := make(map[types.EntityUID]struct{}, len(t.Entities))
		for _, e := range t.Entities {
			set[e] = struct{}{}
		}
		return true, entityInSet(ctx, e, set)
	case ast.ScopeTypeIs:
		return true, e.Type == t.Type
	case ast.ScopeTypeIsIn:
		return true, e.Type == t.Type && entityInOne(ctx, e, t.Entity)
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}
