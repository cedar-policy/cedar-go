package eval

import (
	"context"
	"maps"
	"slices"

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
			p, keep, _ := partialPolicy(evalCtx, p)
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

type AdvancedBatchRequest struct {
	Principal types.Value
	Action    types.Value
	Resource  types.Value
	Context   types.Value
	Variables Variables
}

type sortedAdvancedBatchRequest struct {
	Principal types.Value
	Action    types.Value
	Resource  types.Value
	Context   types.Value
	Variables []variableItem
	Values    Values
}

type variableItem struct {
	Key    types.String
	Values []types.Value
}

type Variables map[types.String][]types.Value

const variableEntityType = "__cedar::variable"

func Variable(v types.String) types.Value {
	return types.NewEntityUID(variableEntityType, v)
}

type AdvancedBatchResult struct {
	Principal types.EntityUID
	Action    types.EntityUID
	Resource  types.EntityUID
	Context   types.Record
	Decision  bool
	Values    Values
}
type Values map[types.String]types.Value

func PubAdvancedBatch(ctx context.Context, policies []*publicast.Policy, entityMap types.Entities, request AdvancedBatchRequest, cb func(AdvancedBatchResult)) error {
	var res batcher
	res.policies = make([]*ast.Policy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		res.policies[i] = p
	}
	res.evalCtx = PrepContext(&Context{})
	sReq := sortedAdvancedBatchRequest{
		Principal: request.Principal,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
		Values:    Values{},
	}
	for k, v := range request.Variables {
		sReq.Variables = append(sReq.Variables, variableItem{Key: k, Values: v})
	}
	slices.SortStableFunc(sReq.Variables, func(a, b variableItem) int {
		return len(a.Values) - len(b.Values)
	})
	return advancedBatch(ctx, &res, entityMap, sReq, cb)
}

// AdvancedBatch will run a batch of authorization evaluations.
// It will error in case of early termination.
// It will error in case any of PARC are an incorrect type at eval type.
// The result passed to the callback must be used immediately and not modified.
func AdvancedBatch(ctx context.Context, policies []*ast.Policy, entityMap types.Entities, request AdvancedBatchRequest, cb func(AdvancedBatchResult)) error {
	var res batcher
	res.policies = make([]*ast.Policy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		res.policies[i] = p
	}
	res.evalCtx = PrepContext(&Context{})
	sReq := sortedAdvancedBatchRequest{
		Principal: request.Principal,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
		Values:    Values{},
	}
	for k, v := range request.Variables {
		sReq.Variables = append(sReq.Variables, variableItem{Key: k, Values: v})
	}
	slices.SortStableFunc(sReq.Variables, func(a, b variableItem) int {
		return len(a.Values) - len(b.Values)
	})
	return advancedBatch(ctx, &res, entityMap, sReq, cb)
}

// cloneSub will return a new value if any of its children have changed
// and signal the change via the boolean
func cloneSub(r types.Value, k types.String, v types.Value) (types.Value, bool) {
	switch t := r.(type) {
	case types.EntityUID:
		if t.Type == variableEntityType && t.ID == k {
			return v, true
		}
	case types.Record:
		cloned := false
		for kk, vv := range t {
			if vv, delta := cloneSub(vv, k, v); delta {
				if !cloned {
					t = maps.Clone(t) // intentional shallow clone
					cloned = true
				}
				t[kk] = vv
			}
		}
		return t, cloned
	case types.Set:
		cloned := false
		for kk, vv := range t {
			if vv, delta := cloneSub(vv, k, v); delta {
				if !cloned {
					t = slices.Clone(t) // intentional shallow clone
					cloned = true
				}
				t[kk] = vv
			}
		}
		return t, cloned
	}
	return r, false
}

func advancedBatch(ctx context.Context, b *batcher, entityMap types.Entities, request sortedAdvancedBatchRequest, cb func(AdvancedBatchResult)) error {
	// check for context cancellation only if there is more work to be done
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(request.Variables) == 0 {
		var res AdvancedBatchResult
		var err error
		if res.Principal, err = ValueToEntity(request.Principal); err != nil {
			return err
		}
		if res.Action, err = ValueToEntity(request.Action); err != nil {
			return err
		}
		if res.Resource, err = ValueToEntity(request.Resource); err != nil {
			return err
		}
		if res.Context, err = ValueToRecord(request.Context); err != nil {
			return err
		}
		ctx := &Context{
			Entities:  entityMap,
			Principal: request.Principal,
			Action:    request.Action,
			Resource:  request.Resource,
			Context:   request.Context,
			inCache:   b.evalCtx.inCache,
		}
		res.Decision = batchAuthz(b, ctx)
		res.Values = request.Values
		cb(res)
		return nil
	}

	// else, partial eval what we have so far
	evalCtx := &Context{Entities: entityMap, inCache: b.evalCtx.inCache}
	evalCtx.Principal = request.Principal
	evalCtx.Action = request.Action
	evalCtx.Resource = request.Resource
	evalCtx.Context = request.Context
	var np []*ast.Policy
	for _, p := range b.policies {
		p, keep, _ := partialPolicy(evalCtx, p)
		if !keep {
			continue
		}
		np = append(np, p)
	}
	b = &batcher{
		evalCtx:  b.evalCtx,
		policies: np,
	}

	// then loop the current unknowns
	u := request.Variables[0]
	_, chPrincipal := cloneSub(request.Principal, u.Key, types.True)
	_, chAction := cloneSub(request.Action, u.Key, types.True)
	_, chResource := cloneSub(request.Resource, u.Key, types.True)
	_, chContext := cloneSub(request.Context, u.Key, types.True)
	uks := request.Variables[1:]
	for _, v := range u.Values {
		child := sortedAdvancedBatchRequest{
			Principal: request.Principal,
			Action:    request.Action,
			Resource:  request.Resource,
			Context:   request.Context,
			Variables: uks,
			Values:    request.Values,
		}
		request.Values[u.Key] = v
		if chPrincipal {
			child.Principal, _ = cloneSub(request.Principal, u.Key, v)
		}
		if chAction {
			child.Action, _ = cloneSub(request.Action, u.Key, v)
		}
		if chResource {
			child.Resource, _ = cloneSub(request.Resource, u.Key, v)
		}
		if chContext {
			child.Context, _ = cloneSub(request.Context, u.Key, v)
		}
		if err := advancedBatch(ctx, b, entityMap, child, cb); err != nil {
			return err
		}
	}
	delete(request.Values, u.Key)
	return nil
}
