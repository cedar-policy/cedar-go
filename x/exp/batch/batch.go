package batch

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/types"
)

func Ignore() types.Value                    { return eval.Ignore() }
func Variable(name types.String) types.Value { return eval.Variable(name) }

type Request struct {
	Principal types.Value
	Action    types.Value
	Resource  types.Value
	Context   types.Value
	Variables Variables
}

type idEvaler struct {
	PolicyID types.PolicyID
	Evaler   eval.BoolEvaler
	Effect   types.Effect
	Position types.Position
}

type idPolicy struct {
	PolicyID types.PolicyID
	Policy   *ast.Policy
}

type batchEvaler struct {
	Variables []variableItem
	Values    Values

	policies []idPolicy
	compiled bool
	// policySet *cedar.PolicySet
	evalers  []*idEvaler
	env      *eval.Env
	callback Callback
}

type variableItem struct {
	Key    types.String
	Values []types.Value
}

type Variables map[types.String][]types.Value

const unknownEntityType = "__cedar::unknown"

func unknownEntity(v types.String) types.EntityUID {
	return types.NewEntityUID(unknownEntityType, v)
}

type Values map[types.String]types.Value

type Result struct {
	Request    types.Request
	Values     Values
	Decision   types.Decision
	Diagnostic types.Diagnostic
}

type Callback func(Result)

// Authorize will run a batch of authorization evaluations.
// It will error in case of early termination.
// It will error in case any of PARC are an incorrect type at eval type.
// The result passed to the callback must be used / cloned immediately and not modified.
func Authorize(ctx context.Context, ps *cedar.PolicySet, entityMap types.Entities, request Request, cb Callback) error {
	var be batchEvaler
	pm := ps.Map()
	be.policies = make([]idPolicy, len(pm))
	i := 0
	for k, p := range pm {
		be.policies[i] = idPolicy{PolicyID: k, Policy: (*ast.Policy)(p.AST())}
		i++
	}
	be.callback = cb
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
	be.env = eval.InitEnv(&eval.Env{
		Entities:  entityMap,
		Principal: request.Principal,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
	})
	be.Values = Values{}
	for k, v := range request.Variables {
		be.Variables = append(be.Variables, variableItem{Key: k, Values: v})
	}
	slices.SortFunc(be.Variables, func(a, b variableItem) int {
		return len(a.Values) - len(b.Values)
	})
	return doBatch(ctx, &be)
}

func doBatch(ctx context.Context, be *batchEvaler) error {
	// check for context cancellation only if there is more work to be done
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(be.Variables) == 0 {
		return diagnosticAuthzWithCallback(be)
	}

	prevState := *be

	// else, partial eval what we have so far
	var np []idPolicy
	for _, p := range be.policies {
		part, keep := eval.PartialPolicy(be.env, p.Policy)
		if !keep {
			continue
		}
		np = append(np, idPolicy{PolicyID: p.PolicyID, Policy: part})
	}
	be.compiled = false
	be.policies = np
	be.evalers = nil

	// if no more partial evaluation, fill in ignores with defaults
	if len(be.Variables) == 1 {
		if eval.IsIgnore(be.env.Principal) {
			be.env.Principal = unknownEntity(consts.Principal)
		}
		if eval.IsIgnore(be.env.Action) {
			be.env.Action = unknownEntity(consts.Action)
		}
		if eval.IsIgnore(be.env.Resource) {
			be.env.Resource = unknownEntity(consts.Resource)
		}
		if eval.IsIgnore(be.env.Context) {
			be.env.Context = types.Record{}
		}
	}
	loopEnv := *be.env

	// then loop the current unknowns
	u := be.Variables[0]
	_, chPrincipal := cloneSub(be.env.Principal, u.Key, nil)
	_, chAction := cloneSub(be.env.Action, u.Key, nil)
	_, chResource := cloneSub(be.env.Resource, u.Key, nil)
	_, chContext := cloneSub(be.env.Context, u.Key, nil)
	be.Variables = be.Variables[1:]
	be.Values = maps.Clone(be.Values)

	for _, v := range u.Values {
		*be.env = loopEnv
		be.Values[u.Key] = v
		if chPrincipal {
			be.env.Principal, _ = cloneSub(loopEnv.Principal, u.Key, v)
		}
		if chAction {
			be.env.Action, _ = cloneSub(loopEnv.Action, u.Key, v)
		}
		if chResource {
			be.env.Resource, _ = cloneSub(loopEnv.Resource, u.Key, v)
		}
		if chContext {
			be.env.Context, _ = cloneSub(loopEnv.Context, u.Key, v)
		}
		if err := doBatch(ctx, be); err != nil {
			return err
		}
	}
	*be = prevState
	return nil
}

func diagnosticAuthzWithCallback(be *batchEvaler) error {
	var res Result
	var err error
	if res.Request.Principal, err = eval.ValueToEntity(be.env.Principal); err != nil {
		return err
	}
	if res.Request.Action, err = eval.ValueToEntity(be.env.Action); err != nil {
		return err
	}
	if res.Request.Resource, err = eval.ValueToEntity(be.env.Resource); err != nil {
		return err
	}
	if res.Request.Context, err = eval.ValueToRecord(be.env.Context); err != nil {
		return err
	}
	res.Values = be.Values
	batchCompile(be)
	// TODO: is there a way to share a cache across requests when using cedar.PolicySet?
	// res.Decision, res.Diagnostic = be.policySet.IsAuthorized(entityMap, res.Request)
	res.Decision, res.Diagnostic = isAuthorized(be.evalers, be.env)
	be.callback(res)
	return nil
}

func isAuthorized(ps []*idEvaler, env *eval.Env) (types.Decision, types.Diagnostic) {
	var diag types.Diagnostic
	var forbids []types.DiagnosticReason
	var permits []types.DiagnosticReason
	// Don't try to short circuit this.
	// - Even though single forbid means forbid
	// - All policy should be run to collect errors
	// - For permit, all permits must be run to collect annotations
	// - For forbid, forbids must be run to collect annotations
	for _, po := range ps {
		result, err := po.Evaler.Eval(env)
		if err != nil {
			diag.Errors = append(diag.Errors, types.DiagnosticError{PolicyID: po.PolicyID, Position: po.Position, Message: err.Error()})
			continue
		}
		if !result {
			continue
		}
		if po.Effect == types.Forbid {
			forbids = append(forbids, types.DiagnosticReason{PolicyID: po.PolicyID, Position: po.Position})
		} else {
			permits = append(permits, types.DiagnosticReason{PolicyID: po.PolicyID, Position: po.Position})
		}
	}
	if len(forbids) > 0 {
		diag.Reasons = forbids
		return types.Deny, diag
	}
	if len(permits) > 0 {
		diag.Reasons = permits
		return types.Allow, diag
	}
	return types.Deny, diag
}

// func testPrintPolicy(p *ast.Policy) {
// 	pp := (*parser.Policy)(p)
// 	var got bytes.Buffer
// 	pp.MarshalCedar(&got)
// 	fmt.Println(got.String())
// }

func batchCompile(be *batchEvaler) {
	if be.compiled {
		return
	}
	// b.policySet = cedar.NewPolicySet() // TODO: pre-set size?
	// for _, p := range b.policies {
	// 	b.policySet.Store(p.PolicyID, cedar.NewPolicyFromAST((*publicast.Policy)(p.Policy)))
	// }
	be.evalers = make([]*idEvaler, len(be.policies))
	for i, p := range be.policies {
		be.evalers[i] = &idEvaler{PolicyID: p.PolicyID, Evaler: eval.Compile(p.Policy), Effect: types.Effect(p.Policy.Effect), Position: types.Position(p.Policy.Position)}
	}
	be.compiled = true
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
