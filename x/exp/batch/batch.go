package batch

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"

	publicast "github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/eval"
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

const unknownEntityType = "__cedar::unknown"

func unknownEntity() types.EntityUID {
	return types.NewEntityUID(unknownEntityType, "")
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

type batchOptions struct {
	authz func(be *batchEvaler, entityMap types.Entities, req batchRequestState) error
}

type BatchOption func(*batchOptions)

func WithCallback(cb func(BatchResult)) BatchOption {
	return func(bo *batchOptions) {
		bo.authz = func(be *batchEvaler, entityMap types.Entities, req batchRequestState) error {
			return basicAuthzWithCallback(be, entityMap, req, cb)
		}
	}
}

type BatchDiagnosticResult struct {
	// TODO: reshape to use main package diagnostic info, use policyID, etc
	// TODO: consider if errors are worth capturing
	BatchResult
	Diagnostic Diagnostic
}

type Diagnostic struct {
	Reasons []string
}

func WithDiagnosticCallback(cb func(BatchDiagnosticResult)) BatchOption {
	return func(bo *batchOptions) {
		bo.authz = func(be *batchEvaler, entityMap types.Entities, req batchRequestState) error {
			return diagnosticAuthzWithCallback(be, entityMap, req, cb)
		}

	}
}

func PubBatch(ctx context.Context, policies []*publicast.Policy, entityMap types.Entities, request BatchRequest, opts ...BatchOption) error {
	pol2 := make([]*ast.Policy, len(policies))
	for i, pub := range policies {
		p := (*ast.Policy)(pub)
		pol2[i] = p
	}
	return Batch(ctx, pol2, entityMap, request, opts...)
}

// Batch will run a batch of authorization evaluations.
// It will error in case of early termination.
// It will error in case any of PARC are an incorrect type at eval type.
// The result passed to the callback must be used immediately and not modified.
func Batch(ctx context.Context, policies []*ast.Policy, entityMap types.Entities, request BatchRequest, opts ...BatchOption) error {
	var be batchEvaler
	for _, o := range opts {
		o(&be.options)
	}
	be.policies = make([]idPolicy, len(policies))
	for i, p := range policies {
		be.policies[i] = idPolicy{PolicyID: strconv.Itoa(i), Policy: p}
	}
	switch {
	case request.Principal == nil:
		return fmt.Errorf("batch missing principal")
	case request.Action == nil:
		return fmt.Errorf("batch missing action")
	case request.Resource == nil:
		return fmt.Errorf("batch missing resource")
	case request.Context == nil:
		return fmt.Errorf("batch missing context")
	}
	be.env = eval.NewEnv()
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
	return doBatch(ctx, &be, entityMap, state)
}

func doBatch(ctx context.Context, be *batchEvaler, entityMap types.Entities, request batchRequestState) error {
	// check for context cancellation only if there is more work to be done
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(request.Variables) == 0 {
		return be.options.authz(be, entityMap, request)
	}

	// else, partial eval what we have so far
	var np []idPolicy
	for _, p := range be.policies {
		part, keep, _ := eval.PartialPolicy(eval.InitEnvWithCacheFrom(&eval.Env{
			Entities:  entityMap,
			Principal: request.Principal,
			Action:    request.Action,
			Resource:  request.Resource,
			Context:   request.Context,
		}, be.env), p.Policy)
		if !keep {
			continue
		}
		np = append(np, idPolicy{PolicyID: p.PolicyID, Policy: part})
	}
	be = &batchEvaler{
		env:      be.env,
		policies: np,
		options:  be.options,
	}

	// if no more partial evaluation, fill in ignores with defaults
	if len(request.Variables) == 1 {
		if eval.IsIgnore(request.Principal) {
			request.Principal = unknownEntity()
		}
		if eval.IsIgnore(request.Action) {
			request.Action = unknownEntity()
		}
		if eval.IsIgnore(request.Resource) {
			request.Resource = unknownEntity()
		}
		if eval.IsIgnore(request.Context) {
			request.Context = types.Record{}
		}
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
		if err := doBatch(ctx, be, entityMap, child); err != nil {
			return err
		}
	}
	delete(request.Values, u.Key)
	return nil
}

type idEvaler struct {
	PolicyID string
	Evaler   eval.BoolEvaler
}

type idPolicy struct {
	PolicyID string
	Policy   *ast.Policy
}

type batchEvaler struct {
	policies []idPolicy
	compiled bool
	forbids  []idEvaler
	permits  []idEvaler
	env      *eval.Env
	options  batchOptions
}

func buildResultEnv(be *batchEvaler, entityMap types.Entities, request batchRequestState) (BatchResult, *eval.Env, error) {
	var res BatchResult
	var err error
	if res.Principal, err = eval.ValueToEntity(request.Principal); err != nil {
		return BatchResult{}, nil, err
	}
	if res.Action, err = eval.ValueToEntity(request.Action); err != nil {
		return BatchResult{}, nil, err
	}
	if res.Resource, err = eval.ValueToEntity(request.Resource); err != nil {
		return BatchResult{}, nil, err
	}
	if res.Context, err = eval.ValueToRecord(request.Context); err != nil {
		return BatchResult{}, nil, err
	}
	res.Values = request.Values
	env := eval.InitEnvWithCacheFrom(&eval.Env{
		Entities:  entityMap,
		Principal: request.Principal,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
	}, be.env)
	return res, env, nil
}

func basicAuthzWithCallback(be *batchEvaler, entityMap types.Entities, request batchRequestState, cb func(BatchResult)) error {
	res, env, err := buildResultEnv(be, entityMap, request)
	if err != nil {
		return err
	}
	res.Decision = batchAuthz(be, env)
	cb(res)
	return nil
}

func diagnosticAuthzWithCallback(be *batchEvaler, entityMap types.Entities, request batchRequestState, cb func(BatchDiagnosticResult)) error {
	res, env, err := buildResultEnv(be, entityMap, request)
	if err != nil {
		return err
	}
	diagRes := BatchDiagnosticResult{
		BatchResult: res,
	}
	diagRes.Decision, diagRes.Diagnostic = diagnosticAuthz(be, env)
	cb(diagRes)
	return nil
}

func diagnosticAuthz(b *batchEvaler, env *eval.Env) (bool, Diagnostic) {
	batchCompile(b)
	var d Diagnostic

	for _, p := range b.forbids {
		v, err := p.Evaler.Eval(env)
		if err != nil {
			continue
		}
		if v {
			d.Reasons = append(d.Reasons, p.PolicyID)
		}
	}
	if len(d.Reasons) > 0 {
		return false, d
	}
	for _, p := range b.permits {
		v, err := p.Evaler.Eval(env)
		if err != nil {
			continue
		}
		if v {
			d.Reasons = append(d.Reasons, p.PolicyID)
		}
	}
	if len(d.Reasons) > 0 {
		return true, d
	}
	return false, d
}

// func testPrintPolicy(p *ast.Policy) {
// 	pp := (*parser.Policy)(p)
// 	var got bytes.Buffer
// 	pp.MarshalCedar(&got)
// 	fmt.Println(got.String())
// }

func batchAuthz(b *batchEvaler, env *eval.Env) bool {
	batchCompile(b)

	for _, p := range b.forbids {
		v, err := p.Evaler.Eval(env)
		if err != nil {
			continue
		}
		if v {
			return false
		}
	}
	for _, p := range b.permits {
		v, err := p.Evaler.Eval(env)
		if err != nil {
			continue
		}
		if v {
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
		idEval := idEvaler{PolicyID: p.PolicyID, Evaler: eval.Compile(p.Policy)}
		if p.Policy.Effect == ast.EffectPermit {
			b.permits = append(b.permits, idEval)
		} else {
			b.forbids = append(b.forbids, idEval)
		}
	}
	b.compiled = true
}

// cloneSub will return a new value if any of its children have changed
// and signal the change via the boolean
func cloneSub(r types.Value, k types.String, v types.Value) (types.Value, bool) {
	switch t := r.(type) {
	case types.EntityUID:
		if key, ok := eval.ToVariable(t); ok && key == k {
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
