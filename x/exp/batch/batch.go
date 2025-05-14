// Package batch allows for performant batch evaluations of Cedar policy given a set of principals, actions, resources,
// and/or context as variables. The batch evaluation takes advantage of a form of [partial evaluation] to whittle the
// policy set down to just those policies which refer to the set of unknown variables. This allows for queries over a
// policy set, such as "to which resources can user A connect when the request comes from outside the United States?"
// which can run much faster than a brute force trawl through every possible authorization request.
//
// [partial evaluation]: https://en.wikipedia.org/wiki/Partial_evaluation
package batch

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/mapset"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// Ignore returns a value that should be ignored during batch evaluation.
func Ignore() types.Value { return eval.Ignore() }

// Variable returns a named variable that is populated during batch evaluation.
func Variable(name types.String) types.Value { return eval.Variable(name) }

// Request defines the PARC and map of Variables to batch evaluate.
type Request struct {
	Principal types.Value
	Action    types.Value
	Resource  types.Value
	Context   types.Value
	Variables Variables
}

// Variables is a map of String to slice of Value.
type Variables map[types.String][]types.Value

// Values is a map of String to Value.  This structure is part of the result and
// reveals the current variable substitutions.
type Values map[types.String]types.Value

// Result is the result of a single batched authorization.  It includes a
// specific Request, the Values that were substituted, and the resulting
// Decision and Diagnostics.
type Result struct {
	Request    types.Request
	Values     Values
	Decision   types.Decision
	Diagnostic types.Diagnostic
}

// Callback is a function that is called for each single batch authorization with
// a Result.
type Callback func(Result) error

type idEvaler struct {
	Policy *ast.Policy
	Evaler eval.BoolEvaler
}

type batchEvaler struct {
	Variables []variableItem
	Values    Values

	policies map[types.PolicyID]*ast.Policy
	compiled bool
	evalers  map[types.PolicyID]*idEvaler
	env      eval.Env
	callback Callback
}

type variableItem struct {
	Key    types.String
	Values []types.Value
}

const unknownEntityType = "__cedar::unknown"

func unknownEntity(v types.String) types.EntityUID {
	return types.NewEntityUID(unknownEntityType, v)
}

var errUnboundVariable = fmt.Errorf("unbound variable")
var errUnusedVariable = fmt.Errorf("unused variable")
var errMissingPart = fmt.Errorf("missing part")
var errInvalidPart = fmt.Errorf("invalid part")

// Authorize will run a batch of authorization evaluations.
//
// All the request parts (PARC) must be specified, but you can
// specify [Variable] or [Ignore].  Variables can be enumerated
// using the Variables.
//
// Using [Ignore] you can ask questions like "When ignoring context could this request be allowed?"
//
//  1. When a Permit Policy Condition refers to an ignored value, the Condition is dropped from the Policy.
//  2. When a Forbid Policy Condition refers to an ignored value, the Policy is dropped.
//  3. When a Scope clause refers to an ignored value, that scope clause is set to match any.
//
// Errors may be returned for a variety of reasons:
//
//   - It will error in case of a context.Context error (e.g. cancellation).
//   - It will error in case any of PARC are an incorrect type at authorization.
//   - It will error in case there are unbound variables.
//   - It will error in case there are unused variables.
//   - It will error in case of a callback error.
//
// The result passed to the callback must be used / cloned immediately and not modified.
func Authorize(ctx context.Context, policies cedar.PolicyIterator, entities types.EntityGetter, request Request, cb Callback) error {
	be := &batchEvaler{}
	var found mapset.MapSet[types.String]
	findVariables(&found, request.Principal)
	findVariables(&found, request.Action)
	findVariables(&found, request.Resource)
	findVariables(&found, request.Context)
	for key := range found.All() {
		if _, ok := request.Variables[key]; !ok {
			return fmt.Errorf("%w: %v", errUnboundVariable, key)
		}
	}
	for k := range request.Variables {
		if !found.Contains(k) {
			return fmt.Errorf("%w: %v", errUnusedVariable, k)
		}
	}
	for _, vs := range request.Variables {
		if len(vs) == 0 {
			return nil
		}
	}
	be.policies = map[types.PolicyID]*ast.Policy{}
	for k, p := range policies.All() {
		be.policies[k] = (*ast.Policy)(p.AST())
	}
	be.callback = cb
	switch {
	case request.Principal == nil:
		return fmt.Errorf("%w: principal", errMissingPart)
	case request.Action == nil:
		return fmt.Errorf("%w: action", errMissingPart)
	case request.Resource == nil:
		return fmt.Errorf("%w: resource", errMissingPart)
	case request.Context == nil:
		return fmt.Errorf("%w: context", errMissingPart)
	}
	if entities == nil {
		var zero types.EntityMap
		entities = zero
	}
	be.env = eval.Env{
		Entities:  entities,
		Principal: request.Principal,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
	}
	be.Values = Values{}
	for k, v := range request.Variables {
		be.Variables = append(be.Variables, variableItem{Key: k, Values: v})
	}
	slices.SortFunc(be.Variables, func(a, b variableItem) int {
		return len(a.Values) - len(b.Values)
	})

	// resolve ignores if no variables exist
	if len(be.Variables) == 0 {
		doPartial(be)
		fixIgnores(be)
	}

	return errors.Join(doBatch(ctx, be), ctx.Err())
}

func doPartial(be *batchEvaler) {
	np := map[types.PolicyID]*ast.Policy{}
	for k, p := range be.policies {
		part, keep := eval.PartialPolicy(be.env, p)
		if !keep {
			continue
		}
		np[k] = part
	}
	be.compiled = false
	be.policies = np
	be.evalers = nil
}

// fixIgnores replaces the Ignore PAR (which may not be EntityUID's in the future) with
// EntityUID's so that the conversion to Result is successful.  An ignored context is
// replaced with a nil Record for the same reason.
func fixIgnores(be *batchEvaler) {
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
		var nilRecord types.Record
		be.env.Context = nilRecord
	}
}

func doBatch(ctx context.Context, be *batchEvaler) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// if no variables, authorize
	if len(be.Variables) == 0 {
		return diagnosticAuthzWithCallback(be)
	}

	// save previous state
	prevState := *be

	// else, partial eval what we have so far
	doPartial(be)

	// if no more partial evaluation, fill in ignores with defaults
	if len(be.Variables) == 1 {
		fixIgnores(be)
	}

	// then loop the current variable
	loopEnv := be.env
	u := be.Variables[0]
	dummyVal := types.True
	_, chPrincipal := cloneSub(be.env.Principal, u.Key, dummyVal)
	_, chAction := cloneSub(be.env.Action, u.Key, dummyVal)
	_, chResource := cloneSub(be.env.Resource, u.Key, dummyVal)
	_, chContext := cloneSub(be.env.Context, u.Key, dummyVal)
	be.Variables = be.Variables[1:]
	be.Values = maps.Clone(be.Values)
	for _, v := range u.Values {
		be.env = loopEnv
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

	// restore previous state
	*be = prevState
	return nil
}

func diagnosticAuthzWithCallback(be *batchEvaler) error {
	var res Result
	var err error
	if res.Request.Principal, err = eval.ValueToEntity(be.env.Principal); err != nil {
		return fmt.Errorf("%w: %w", errInvalidPart, err)
	}
	if res.Request.Action, err = eval.ValueToEntity(be.env.Action); err != nil {
		return fmt.Errorf("%w: %w", errInvalidPart, err)
	}
	if res.Request.Resource, err = eval.ValueToEntity(be.env.Resource); err != nil {
		return fmt.Errorf("%w: %w", errInvalidPart, err)
	}
	if res.Request.Context, err = eval.ValueToRecord(be.env.Context); err != nil {
		return fmt.Errorf("%w: %w", errInvalidPart, err)
	}
	res.Values = be.Values
	batchCompile(be)
	res.Decision, res.Diagnostic = isAuthorized(be.evalers, be.env)
	return be.callback(res)
}

func isAuthorized(ps map[types.PolicyID]*idEvaler, env eval.Env) (types.Decision, types.Diagnostic) {
	var diag types.Diagnostic
	var forbids []types.DiagnosticReason
	var permits []types.DiagnosticReason
	// Don't try to short circuit this.
	// - Even though single forbid means forbid
	// - All policy should be run to collect errors
	// - For permit, all permits must be run to collect annotations
	// - For forbid, forbids must be run to collect annotations
	for pid, po := range ps {
		result, err := po.Evaler.Eval(env)
		if err != nil {
			diag.Errors = append(diag.Errors, types.DiagnosticError{PolicyID: pid, Position: types.Position(po.Policy.Position), Message: err.Error()})
			continue
		}
		if !result {
			continue
		}
		if po.Policy.Effect == ast.EffectPermit {
			permits = append(permits, types.DiagnosticReason{PolicyID: pid, Position: types.Position(po.Policy.Position)})
		} else {
			forbids = append(forbids, types.DiagnosticReason{PolicyID: pid, Position: types.Position(po.Policy.Position)})
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

func batchCompile(be *batchEvaler) {
	if be.compiled {
		return
	}
	be.evalers = make(map[types.PolicyID]*idEvaler, len(be.policies))
	for k, p := range be.policies {
		be.evalers[k] = &idEvaler{Policy: p, Evaler: eval.Compile(p)}
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
		var newMap types.RecordMap
		for kk, vv := range t.All() {
			if vv, delta := cloneSub(vv, k, v); delta && newMap == nil {
				if newMap == nil {
					newMap = t.Map()
				}
				newMap[kk] = vv
			}
		}

		if newMap == nil {
			return t, false
		}
		return types.NewRecord(newMap), true
	case types.Set:
		hasDeltas := false

		// Look for deltas. Unfortunately, due to the indeterminate nature of the set iteration order,
		// we can't pull the same trick as we do for Records above
		for vv := range t.All() {
			if _, delta := cloneSub(vv, k, v); delta {
				hasDeltas = true
				break
			}
		}

		// If no deltas, just return the input Value
		if !hasDeltas {
			return t, false
		}

		// If there were deltas, build a new Set
		newSlice := make([]types.Value, 0, t.Len())
		for vv := range t.All() {
			vv, _ = cloneSub(vv, k, v)
			newSlice = append(newSlice, vv)
		}

		return types.NewSet(newSlice...), true
	}
	return r, false
}

func findVariables(found *mapset.MapSet[types.String], r types.Value) {
	switch t := r.(type) {
	case types.EntityUID:
		if key, ok := eval.ToVariable(t); ok {
			found.Add(key)
		}
	case types.Record:
		for vv := range t.Values() {
			findVariables(found, vv)
		}
	case types.Set:
		for vv := range t.All() {
			findVariables(found, vv)
		}
	}
}
