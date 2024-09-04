package eval

import (
	"context"

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
	evalCtx  *Context
}

// PubBatch will run a batch of authorization evaluations.  It will only error in case of early termination.
func PubBatch(ctx context.Context, policies []*publicast.Policy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) error {
	var res batcher
	res.policies = make([]*ast.Policy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		res.policies[i] = p
	}
	res.evalCtx = PrepContext(&Context{})
	return batch(ctx, &res, entityMap, request, cb)
}

// Batch will run a batch of authorization evaluations.  It will only error in case of early termination.
func Batch(ctx context.Context, policies []*ast.Policy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) error {
	var res batcher
	res.policies = make([]*ast.Policy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		res.policies[i] = p
	}
	res.evalCtx = PrepContext(&Context{})
	return batch(ctx, &res, entityMap, request, cb)
}

func batch(ctx context.Context, b *batcher, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) error {
	pl, al, rl := len(request.Principals), len(request.Actions), len(request.Resources)
	if pl == 0 || al == 0 || rl == 0 {
		return nil
	}

	// check for context cancellation only if there is more work to be done
	if err := ctx.Err(); err != nil {
		return err
	}

	if pl == 1 && al == 1 && rl == 1 {
		ctx := &Context{
			Entities:  entityMap,
			Principal: request.Principals[0],
			Action:    request.Actions[0],
			Resource:  request.Resources[0],
			Context:   request.Context,
			inCache:   b.evalCtx.inCache,
		}
		ok := batchAuthz(b, ctx)
		cb(BatchResult{
			Principal: request.Principals[0],
			Action:    request.Actions[0],
			Resource:  request.Resources[0],
			Decision:  ok,
		})
		return nil
	}

	// apply partial evaluation
	if pl == 1 || al == 1 || rl == 1 {
		evalCtx := &Context{Entities: entityMap, inCache: b.evalCtx.inCache}
		if pl == 1 {
			evalCtx.Principal = request.Principals[0]
		}
		if al == 1 {
			evalCtx.Action = request.Actions[0]
		}
		if rl == 1 {
			evalCtx.Resource = request.Resources[0]
		}
		var np []*ast.Policy
		for _, p := range b.policies {
			p, _, keep := partialPolicy(evalCtx, p)
			if !keep {
				continue
			}
			np = append(np, p)
		}
		b = &batcher{
			evalCtx:  b.evalCtx,
			policies: np,
		}
	}

	if pl > 1 && (al == 1 || pl <= al) && (rl == 1 || pl <= rl) {
		return batchPrincipal(ctx, b, entityMap, request, cb)
	} else if al > 1 && (pl == 1 || al <= pl) && (rl == 1 || al <= rl) {
		return batchAction(ctx, b, entityMap, request, cb)
	} else {
		return batchResource(ctx, b, entityMap, request, cb)
	}
}

func batchPrincipal(ctx context.Context, b *batcher, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) error {
	in := request.Principals
	for i := range in {
		request.Principals = in[i : i+1]
		if err := batch(ctx, b, entityMap, request, cb); err != nil {
			return err
		}
	}
	return nil
}
func batchAction(ctx context.Context, b *batcher, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) error {
	in := request.Actions
	for i := range in {
		request.Actions = in[i : i+1]
		if err := batch(ctx, b, entityMap, request, cb); err != nil {
			return err
		}
	}
	return nil
}
func batchResource(ctx context.Context, b *batcher, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) error {
	in := request.Resources
	for i := range in {
		request.Resources = in[i : i+1]
		if err := batch(ctx, b, entityMap, request, cb); err != nil {
			return err
		}
	}
	return nil
}

// func testPrintPolicy(p *ast.Policy) {
// 	pp := (*parser.Policy)(p)
// 	var got bytes.Buffer
// 	pp.MarshalCedar(&got)
// 	fmt.Println(got.String())
// }

func batchAuthz(b *batcher, evalCtx *Context) bool {
	batchCompile(b)

	for _, p := range b.forbids {
		v, err := p.Eval(evalCtx)
		if err != nil {
			continue
		}
		if v, ok := v.(types.Boolean); ok && bool(v) {
			return false
		}
	}
	for _, p := range b.permits {
		v, err := p.Eval(evalCtx)
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
