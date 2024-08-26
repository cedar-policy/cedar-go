package eval

import (
	publicast "github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/ast"
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

type batcher struct {
	policies []*ast.Policy
	compiled bool
	forbids  []Evaler
	permits  []Evaler
	ctx      *Context
}

func PubBatch(policies []*publicast.Policy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	var res batcher
	res.policies = make([]*ast.Policy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		res.policies[i] = p
	}
	res.ctx = PrepContext(&Context{})
	batch(&res, entityMap, request, cb)
}

func Batch(policies []*ast.Policy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	var res batcher
	res.policies = make([]*ast.Policy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		res.policies[i] = p
	}
	res.ctx = PrepContext(&Context{})
	batch(&res, entityMap, request, cb)
}

func batch(b *batcher, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
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
			Context:   request.Context,
			inCache:   b.ctx.inCache,
		}
		ok := batchAuthz(b, ctx)
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
		ctx := &Context{Entities: entityMap, inCache: b.ctx.inCache}
		if pl == 1 {
			ctx.Principal = request.Principals[0]
		}
		if al == 1 {
			ctx.Action = request.Actions[0]
		}
		if rl == 1 {
			ctx.Resource = request.Resources[0]
		}
		var np []*ast.Policy
		for _, p := range b.policies {
			p, keep := partialPolicy(ctx, p)
			if !keep {
				continue
			}
			np = append(np, p)
		}
		b = &batcher{
			ctx:      b.ctx,
			policies: np,
		}
	}

	if pl > 1 && (al == 1 || pl <= al) && (rl == 1 || pl <= rl) {
		batchPrincipal(b, entityMap, request, cb)
	} else if al > 1 && (pl == 1 || al <= pl) && (rl == 1 || al <= rl) {
		batchAction(b, entityMap, request, cb)
	} else {
		batchResource(b, entityMap, request, cb)
	}
}

func batchPrincipal(b *batcher, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	in := request.Principals
	for i := range in {
		request.Principals = in[i : i+1]
		batch(b, entityMap, request, cb)
	}
}
func batchAction(b *batcher, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	in := request.Actions
	for i := range in {
		request.Actions = in[i : i+1]
		batch(b, entityMap, request, cb)
	}
}
func batchResource(b *batcher, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) {
	in := request.Resources
	for i := range in {
		request.Resources = in[i : i+1]
		batch(b, entityMap, request, cb)
	}
}

// func testPrintPolicy(p *ast.Policy) {
// 	pp := (*parser.Policy)(p)
// 	var got bytes.Buffer
// 	pp.MarshalCedar(&got)
// 	fmt.Println(got.String())
// }

func batchAuthz(b *batcher, ctx *Context) bool {
	batchCompile(b)

	for _, p := range b.forbids {
		v, err := p.Eval(ctx)
		if err != nil {
			continue
		}
		if v, ok := v.(types.Boolean); ok && bool(v) {
			return false
		}
	}
	for _, p := range b.permits {
		v, err := p.Eval(ctx)
		if err != nil {
			continue
		}
		if v, ok := v.(types.Boolean); ok && bool(v) {
			return true
		}
	}
	return false
}

func batchCompile(b *batcher) {
	if b.compiled {
		return
	}
	for _, p := range b.policies {
		if p.Effect == ast.EffectPermit {
			b.permits = append(b.permits, Compile(p))
		} else {
			b.forbids = append(b.forbids, Compile(p))
		}
	}
	b.compiled = true
}
