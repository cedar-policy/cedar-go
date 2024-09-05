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
	Principal types.Value
	Action    types.Value
	Resource  types.Value
	Context   types.Value
	Variables Variables
}

type batchRequestState struct {
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

type BatchResult struct {
	Principal types.EntityUID
	Action    types.EntityUID
	Resource  types.EntityUID
	Context   types.Record
	Decision  bool
	Values    Values
}
type Values map[types.String]types.Value

func PubBatch(ctx context.Context, policies []*publicast.Policy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) error {
	pol2 := make([]*ast.Policy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		pol2[i] = p
	}
	return Batch(ctx, pol2, entityMap, request, cb)
}

// Batch will run a batch of authorization evaluations.
// It will error in case of early termination.
// It will error in case any of PARC are an incorrect type at eval type.
// The result passed to the callback must be used immediately and not modified.
func Batch(ctx context.Context, policies []*ast.Policy, entityMap types.Entities, request BatchRequest, cb func(BatchResult)) error {
	var be batchEvaler
	be.policies = policies
	be.evalCtx = PrepContext(&Context{})
	state := batchRequestState{
		Principal: request.Principal,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
		Values:    Values{},
	}
	for k, v := range request.Variables {
		state.Variables = append(state.Variables, variableItem{Key: k, Values: v})
	}
	slices.SortFunc(state.Variables, func(a, b variableItem) int {
		return len(a.Values) - len(b.Values)
	})
	return doBatch(ctx, &be, entityMap, state, cb)
}

func doBatch(ctx context.Context, be *batchEvaler, entityMap types.Entities, request batchRequestState, cb func(BatchResult)) error {
	// check for context cancellation only if there is more work to be done
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(request.Variables) == 0 {
		var res BatchResult
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
			inCache:   be.evalCtx.inCache,
		}
		res.Decision = batchAuthz(be, ctx)
		res.Values = request.Values
		cb(res)
		return nil
	}

	// else, partial eval what we have so far
	var np []*ast.Policy
	for _, p := range be.policies {
		p, keep, _ := partialPolicy(&Context{
			Entities:  entityMap,
			inCache:   be.evalCtx.inCache,
			Principal: request.Principal,
			Action:    request.Action,
			Resource:  request.Resource,
			Context:   request.Context,
		}, p)
		if !keep {
			continue
		}
		np = append(np, p)
	}
	be = &batchEvaler{
		evalCtx:  be.evalCtx,
		policies: np,
	}

	// then loop the current unknowns
	u := request.Variables[0]
	_, chPrincipal := cloneSub(request.Principal, u.Key, nil)
	_, chAction := cloneSub(request.Action, u.Key, nil)
	_, chResource := cloneSub(request.Resource, u.Key, nil)
	_, chContext := cloneSub(request.Context, u.Key, nil)
	uks := request.Variables[1:]
	for _, v := range u.Values {
		child := batchRequestState{
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
		if err := doBatch(ctx, be, entityMap, child, cb); err != nil {
			return err
		}
	}
	delete(request.Values, u.Key)
	return nil
}

type batchEvaler struct {
	policies []*ast.Policy
	compiled bool
	forbids  []Evaler
	permits  []Evaler
	evalCtx  *Context
}

// func testPrintPolicy(p *ast.Policy) {
// 	pp := (*parser.Policy)(p)
// 	var got bytes.Buffer
// 	pp.MarshalCedar(&got)
// 	fmt.Println(got.String())
// }

func batchAuthz(b *batchEvaler, evalCtx *Context) bool {
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

func batchCompile(b *batchEvaler) {
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
