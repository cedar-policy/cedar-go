package eval

import (
	"bytes"
	"fmt"

	publicast "github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/types"
)

type BatchRequest struct {
	Principals []types.EntityUID
	Actions    []types.EntityUID
	Resources  []types.EntityUID
	Context    types.Record
}

type BatchResult struct {
	Principal, Action, Resource types.EntityUID
	Decision                    bool
}

type batchPolicy struct {
	ast    *ast.Policy
	evaler Evaler
}

func PubBatch(policies []*publicast.Policy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	ps := make([]batchPolicy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		ps[i] = batchPolicy{
			ast:    p,
			evaler: Compile(p),
		}
	}
	batch(ps, entityMap, request, cb)
}

func Batch(policies []*ast.Policy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	ps := make([]batchPolicy, len(policies))
	for i, p := range policies {
		ps[i] = batchPolicy{
			ast:    p,
			evaler: Compile(p),
		}
	}
	batch(ps, entityMap, request, cb)
}

func batch(policies []batchPolicy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	pl, al, rl := len(request.Principals), len(request.Actions), len(request.Resources)
	if pl == 0 || al == 0 || rl == 0 {
		return
	}
	if pl == 1 && al == 1 && rl == 1 {
		ctx := &Context{
			Entities:  entityMap,
			Principal: request.Principals[0],
			Action:    request.Actions[0],
			Resource:  request.Resources[0],
		}
		ok := batchAuthz(policies, ctx)
		cb(BatchResult{
			Principal: request.Principals[0],
			Action:    request.Actions[0],
			Resource:  request.Resources[0],
			Decision:  ok,
		})
		return
	}

	// apply partial evaluation
	if pl == 1 || al == 1 || rl == 1 {
		ctx := &Context{Entities: entityMap}
		if pl == 1 {
			ctx.Principal = request.Principals[0]
		}
		if al == 1 {
			ctx.Action = request.Actions[0]
		}
		if rl == 1 {
			ctx.Resource = request.Resources[0]
		}
		var np []batchPolicy
		for _, p := range policies {
			p, keep := partialPolicy(ctx, p.ast)
			if !keep {
				continue
			}
			np = append(np, batchPolicy{
				ast:    p,
				evaler: Compile(p),
			})
		}
		// fmt.Println("partial", pl, al, rl)
		policies = np
	}

	if pl > 1 && (al == 1 || pl <= al) && (rl == 1 || pl <= rl) {
		batchPrincipal(policies, entityMap, request, cb)
	} else if al > 1 && (pl == 1 || al <= pl) && (rl == 1 || al <= rl) {
		batchAction(policies, entityMap, request, cb)
	} else {
		batchResource(policies, entityMap, request, cb)
	}
}

func batchPrincipal(policies []batchPolicy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	in := request.Principals
	for i := range in {
		request.Principals = in[i : i+1]
		batch(policies, entityMap, request, cb)
	}
}
func batchAction(policies []batchPolicy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	in := request.Actions
	for i := range in {
		request.Actions = in[i : i+1]
		batch(policies, entityMap, request, cb)
	}
}
func batchResource(policies []batchPolicy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	in := request.Resources
	for i := range in {
		request.Resources = in[i : i+1]
		batch(policies, entityMap, request, cb)
	}
}

func testPrintPolicy(p *ast.Policy) {
	pp := (*parser.Policy)(p)
	var got bytes.Buffer
	pp.MarshalCedar(&got)
	fmt.Println(got.String())
}

func batchAuthz(policies []batchPolicy, ctx *Context) bool {
	var decision bool
	for _, p := range policies {
		// testPrintPolicy(p.ast)
		v, err := p.evaler.Eval(ctx)
		if err != nil {
			continue
		}
		if v, ok := v.(types.Boolean); ok {
			if !v {
				continue
			}
			if p.ast.Effect == ast.EffectForbid {
				return false
			}
			decision = true
		}
	}
	return decision
}
