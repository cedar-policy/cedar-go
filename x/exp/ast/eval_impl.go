package ast

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

var errOverflow = fmt.Errorf("integer overflow")
var errUnknownExtensionFunction = fmt.Errorf("function does not exist")
var errArity = fmt.Errorf("wrong number of arguments provided to extension function")
var errAttributeAccess = fmt.Errorf("does not have the attribute")
var errEntityNotExist = fmt.Errorf("does not exist")
var errUnspecifiedEntity = fmt.Errorf("unspecified entity")

// TODO: make private again
type EvalContext struct {
	Entities                    Entities
	Principal, Action, Resource types.Value
	Context                     types.Value
}

type Evaler interface {
	Eval(*EvalContext) (types.Value, error)
}

func evalBool(n Evaler, ctx *EvalContext) (types.Boolean, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return false, err
	}
	b, err := types.ValueToBool(v)
	if err != nil {
		return false, err
	}
	return b, nil
}

func evalLong(n Evaler, ctx *EvalContext) (types.Long, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return 0, err
	}
	l, err := types.ValueToLong(v)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func evalString(n Evaler, ctx *EvalContext) (types.String, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return "", err
	}
	s, err := types.ValueToString(v)
	if err != nil {
		return "", err
	}
	return s, nil
}

func evalSet(n Evaler, ctx *EvalContext) (types.Set, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return nil, err
	}
	s, err := types.ValueToSet(v)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func evalEntity(n Evaler, ctx *EvalContext) (types.EntityUID, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return types.EntityUID{}, err
	}
	e, err := types.ValueToEntity(v)
	if err != nil {
		return types.EntityUID{}, err
	}
	return e, nil
}

func evalPath(n Evaler, ctx *EvalContext) (types.Path, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return "", err
	}
	e, err := types.ValueToPath(v)
	if err != nil {
		return "", err
	}
	return e, nil
}

func evalDecimal(n Evaler, ctx *EvalContext) (types.Decimal, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return types.Decimal(0), err
	}
	d, err := types.ValueToDecimal(v)
	if err != nil {
		return types.Decimal(0), err
	}
	return d, nil
}

func evalIP(n Evaler, ctx *EvalContext) (types.IPAddr, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return types.IPAddr{}, err
	}
	i, err := types.ValueToIP(v)
	if err != nil {
		return types.IPAddr{}, err
	}
	return i, nil
}

// errorEval
type errorEval struct {
	err error
}

func newErrorEval(err error) *errorEval {
	return &errorEval{
		err: err,
	}
}

func (n *errorEval) Eval(_ *EvalContext) (types.Value, error) {
	return types.ZeroValue(), n.err
}

// literalEval
type literalEval struct {
	value types.Value
}

func newLiteralEval(value types.Value) *literalEval {
	return &literalEval{value: value}
}

func (n *literalEval) Eval(_ *EvalContext) (types.Value, error) {
	return n.value, nil
}

// orEval
type orEval struct {
	lhs Evaler
	rhs Evaler
}

func newOrNode(lhs Evaler, rhs Evaler) *orEval {
	return &orEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *orEval) Eval(ctx *EvalContext) (types.Value, error) {
	v, err := n.lhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	b, err := types.ValueToBool(v)
	if err != nil {
		return types.ZeroValue(), err
	}
	if b {
		return v, nil
	}
	v, err = n.rhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	_, err = types.ValueToBool(v)
	if err != nil {
		return types.ZeroValue(), err
	}
	return v, nil
}

// andEval
type andEval struct {
	lhs Evaler
	rhs Evaler
}

func newAndEval(lhs Evaler, rhs Evaler) *andEval {
	return &andEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *andEval) Eval(ctx *EvalContext) (types.Value, error) {
	v, err := n.lhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	b, err := types.ValueToBool(v)
	if err != nil {
		return types.ZeroValue(), err
	}
	if !b {
		return v, nil
	}
	v, err = n.rhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	_, err = types.ValueToBool(v)
	if err != nil {
		return types.ZeroValue(), err
	}
	return v, nil
}

// notEval
type notEval struct {
	inner Evaler
}

func newNotEval(inner Evaler) *notEval {
	return &notEval{
		inner: inner,
	}
}

func (n *notEval) Eval(ctx *EvalContext) (types.Value, error) {
	v, err := n.inner.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	b, err := types.ValueToBool(v)
	if err != nil {
		return types.ZeroValue(), err
	}
	return !b, nil
}

// Overflow
// The Go spec specifies that overflow results in defined and deterministic
// behavior (https://go.dev/ref/spec#Integer_overflow), so we can go ahead and
// do the operations and then check for overflow ex post facto.

func checkedAddI64(lhs, rhs types.Long) (types.Long, bool) {
	result := lhs + rhs
	if (result > lhs) != (rhs > 0) {
		return result, false
	}
	return result, true
}

func checkedSubI64(lhs, rhs types.Long) (types.Long, bool) {
	result := lhs - rhs
	if (result > lhs) != (rhs < 0) {
		return result, false
	}
	return result, true
}

func checkedMulI64(lhs, rhs types.Long) (types.Long, bool) {
	if lhs == 0 || rhs == 0 {
		return 0, true
	}
	result := lhs * rhs
	if (result < 0) != ((lhs < 0) != (rhs < 0)) {
		// If the result doesn't have the correct sign, then we overflowed.
		return result, false
	}
	if result/lhs != rhs {
		// If division doesn't yield the original value, then we overflowed.
		return result, false
	}
	return result, true
}

func checkedNegI64(a types.Long) (types.Long, bool) {
	if a == -9_223_372_036_854_775_808 {
		return 0, false
	}
	return -a, true
}

// addEval
type addEval struct {
	lhs Evaler
	rhs Evaler
}

func newAddEval(lhs Evaler, rhs Evaler) *addEval {
	return &addEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *addEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	res, ok := checkedAddI64(lhs, rhs)
	if !ok {
		return types.ZeroValue(), fmt.Errorf("%w while attempting to add `%d` with `%d`", errOverflow, lhs, rhs)
	}
	return res, nil
}

// subtractEval
type subtractEval struct {
	lhs Evaler
	rhs Evaler
}

func newSubtractEval(lhs Evaler, rhs Evaler) *subtractEval {
	return &subtractEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *subtractEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	res, ok := checkedSubI64(lhs, rhs)
	if !ok {
		return types.ZeroValue(), fmt.Errorf("%w while attempting to subtract `%d` from `%d`", errOverflow, rhs, lhs)
	}
	return res, nil
}

// multiplyEval
type multiplyEval struct {
	lhs Evaler
	rhs Evaler
}

func newMultiplyEval(lhs Evaler, rhs Evaler) *multiplyEval {
	return &multiplyEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *multiplyEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	res, ok := checkedMulI64(lhs, rhs)
	if !ok {
		return types.ZeroValue(), fmt.Errorf("%w while attempting to multiply `%d` by `%d`", errOverflow, lhs, rhs)
	}
	return res, nil
}

// negateEval
type negateEval struct {
	inner Evaler
}

func newNegateEval(inner Evaler) *negateEval {
	return &negateEval{
		inner: inner,
	}
}

func (n *negateEval) Eval(ctx *EvalContext) (types.Value, error) {
	inner, err := evalLong(n.inner, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	res, ok := checkedNegI64(inner)
	if !ok {
		return types.ZeroValue(), fmt.Errorf("%w while attempting to negate `%d`", errOverflow, inner)
	}
	return res, nil
}

// longLessThanEval
type longLessThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newLongLessThanEval(lhs Evaler, rhs Evaler) *longLessThanEval {
	return &longLessThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longLessThanEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs < rhs), nil
}

// longLessThanOrEqualEval
type longLessThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newLongLessThanOrEqualEval(lhs Evaler, rhs Evaler) *longLessThanOrEqualEval {
	return &longLessThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longLessThanOrEqualEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs <= rhs), nil
}

// longGreaterThanEval
type longGreaterThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newLongGreaterThanEval(lhs Evaler, rhs Evaler) *longGreaterThanEval {
	return &longGreaterThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longGreaterThanEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs > rhs), nil
}

// longGreaterThanOrEqualEval
type longGreaterThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newLongGreaterThanOrEqualEval(lhs Evaler, rhs Evaler) *longGreaterThanOrEqualEval {
	return &longGreaterThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longGreaterThanOrEqualEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs >= rhs), nil
}

// decimalLessThanEval
type decimalLessThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newDecimalLessThanEval(lhs Evaler, rhs Evaler) *decimalLessThanEval {
	return &decimalLessThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *decimalLessThanEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs < rhs), nil
}

// decimalLessThanOrEqualEval
type decimalLessThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newDecimalLessThanOrEqualEval(lhs Evaler, rhs Evaler) *decimalLessThanOrEqualEval {
	return &decimalLessThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *decimalLessThanOrEqualEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs <= rhs), nil
}

// decimalGreaterThanEval
type decimalGreaterThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newDecimalGreaterThanEval(lhs Evaler, rhs Evaler) *decimalGreaterThanEval {
	return &decimalGreaterThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *decimalGreaterThanEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs > rhs), nil
}

// decimalGreaterThanOrEqualEval
type decimalGreaterThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newDecimalGreaterThanOrEqualEval(lhs Evaler, rhs Evaler) *decimalGreaterThanOrEqualEval {
	return &decimalGreaterThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *decimalGreaterThanOrEqualEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs >= rhs), nil
}

// ifThenElseEval
type ifThenElseEval struct {
	if_   Evaler
	then  Evaler
	else_ Evaler
}

func newIfThenElseEval(if_, then, else_ Evaler) *ifThenElseEval {
	return &ifThenElseEval{
		if_:   if_,
		then:  then,
		else_: else_,
	}
}

func (n *ifThenElseEval) Eval(ctx *EvalContext) (types.Value, error) {
	cond, err := evalBool(n.if_, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	if cond {
		return n.then.Eval(ctx)
	}
	return n.else_.Eval(ctx)
}

// notEqualNode
type equalEval struct {
	lhs, rhs Evaler
}

func newEqualEval(lhs, rhs Evaler) *equalEval {
	return &equalEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *equalEval) Eval(ctx *EvalContext) (types.Value, error) {
	lv, err := n.lhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rv, err := n.rhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lv.Equal(rv)), nil
}

// notEqualEval
type notEqualEval struct {
	lhs, rhs Evaler
}

func newNotEqualEval(lhs, rhs Evaler) *notEqualEval {
	return &notEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *notEqualEval) Eval(ctx *EvalContext) (types.Value, error) {
	lv, err := n.lhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rv, err := n.rhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(!lv.Equal(rv)), nil
}

// setLiteralEval
type setLiteralEval struct {
	elements []Evaler
}

func newSetLiteralEval(elements []Evaler) *setLiteralEval {
	return &setLiteralEval{elements: elements}
}

func (n *setLiteralEval) Eval(ctx *EvalContext) (types.Value, error) {
	var vals types.Set
	for _, e := range n.elements {
		v, err := e.Eval(ctx)
		if err != nil {
			return types.ZeroValue(), err
		}
		vals = append(vals, v)
	}
	return vals, nil
}

// containsEval
type containsEval struct {
	lhs, rhs Evaler
}

func newContainsEval(lhs, rhs Evaler) *containsEval {
	return &containsEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalSet(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := n.rhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(lhs.Contains(rhs)), nil
}

// containsAllEval
type containsAllEval struct {
	lhs, rhs Evaler
}

func newContainsAllEval(lhs, rhs Evaler) *containsAllEval {
	return &containsAllEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsAllEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalSet(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalSet(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	result := true
	for _, e := range rhs {
		if !lhs.Contains(e) {
			result = false
			break
		}
	}
	return types.Boolean(result), nil
}

// containsAnyEval
type containsAnyEval struct {
	lhs, rhs Evaler
}

func newContainsAnyEval(lhs, rhs Evaler) *containsAnyEval {
	return &containsAnyEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsAnyEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalSet(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalSet(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	result := false
	for _, e := range rhs {
		if lhs.Contains(e) {
			result = true
			break
		}
	}
	return types.Boolean(result), nil
}

// recordLiteralEval
type recordLiteralEval struct {
	elements map[string]Evaler
}

func newRecordLiteralEval(elements map[string]Evaler) *recordLiteralEval {
	return &recordLiteralEval{elements: elements}
}

func (n *recordLiteralEval) Eval(ctx *EvalContext) (types.Value, error) {
	vals := types.Record{}
	for k, en := range n.elements {
		v, err := en.Eval(ctx)
		if err != nil {
			return types.ZeroValue(), err
		}
		vals[k] = v
	}
	return vals, nil
}

// attributeAccessEval
type attributeAccessEval struct {
	object    Evaler
	attribute string
}

func newAttributeAccessEval(record Evaler, attribute string) *attributeAccessEval {
	return &attributeAccessEval{object: record, attribute: attribute}
}

func (n *attributeAccessEval) Eval(ctx *EvalContext) (types.Value, error) {
	v, err := n.object.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	var record types.Record
	key := "record"
	switch vv := v.(type) {
	case types.EntityUID:
		key = "`" + vv.String() + "`"
		var unspecified types.EntityUID
		if vv == unspecified {
			return types.ZeroValue(), fmt.Errorf("cannot access attribute `%s` of %w", n.attribute, errUnspecifiedEntity)
		}
		rec, ok := ctx.Entities[vv]
		if !ok {
			return types.ZeroValue(), fmt.Errorf("entity `%v` %w", vv.String(), errEntityNotExist)
		} else {
			record = rec.Attributes
		}
	case types.Record:
		record = vv
	default:
		return types.ZeroValue(), fmt.Errorf("%w: expected one of [record, (entity of type `any_entity_type`)], got %v", types.ErrType, v.TypeName())
	}
	val, ok := record[n.attribute]
	if !ok {
		return types.ZeroValue(), fmt.Errorf("%s %w `%s`", key, errAttributeAccess, n.attribute)
	}
	return val, nil
}

// hasEval
type hasEval struct {
	object    Evaler
	attribute string
}

func newHasEval(record Evaler, attribute string) *hasEval {
	return &hasEval{object: record, attribute: attribute}
}

func (n *hasEval) Eval(ctx *EvalContext) (types.Value, error) {
	v, err := n.object.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	var record types.Record
	switch vv := v.(type) {
	case types.EntityUID:
		rec, ok := ctx.Entities[vv]
		if !ok {
			record = types.Record{}
		} else {
			record = rec.Attributes
		}
	case types.Record:
		record = vv
	default:
		return types.ZeroValue(), fmt.Errorf("%w: expected one of [record, (entity of type `any_entity_type`)], got %v", types.ErrType, v.TypeName())
	}
	_, ok := record[n.attribute]
	return types.Boolean(ok), nil
}

// likeEval
type likeEval struct {
	lhs     Evaler
	pattern Pattern
}

func newLikeEval(lhs Evaler, pattern Pattern) *likeEval {
	return &likeEval{lhs: lhs, pattern: pattern}
}

func (l *likeEval) Eval(ctx *EvalContext) (types.Value, error) {
	v, err := evalString(l.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(match(l.pattern, string(v))), nil
}

type variableName func(ctx *EvalContext) types.Value

func variableNamePrincipal(ctx *EvalContext) types.Value { return ctx.Principal }
func variableNameAction(ctx *EvalContext) types.Value    { return ctx.Action }
func variableNameResource(ctx *EvalContext) types.Value  { return ctx.Resource }
func variableNameContext(ctx *EvalContext) types.Value   { return ctx.Context }

// variableEval
type variableEval struct {
	variableName variableName
}

func newVariableEval(variableName variableName) *variableEval {
	return &variableEval{variableName: variableName}
}

func (n *variableEval) Eval(ctx *EvalContext) (types.Value, error) {
	return n.variableName(ctx), nil
}

// inEval
type inEval struct {
	lhs, rhs Evaler
}

func newInEval(lhs, rhs Evaler) *inEval {
	return &inEval{lhs: lhs, rhs: rhs}
}

func entityIn(entity types.EntityUID, query map[types.EntityUID]struct{}, entities Entities) bool {
	checked := map[types.EntityUID]struct{}{}
	toCheck := []types.EntityUID{entity}
	for len(toCheck) > 0 {
		var candidate types.EntityUID
		candidate, toCheck = toCheck[len(toCheck)-1], toCheck[:len(toCheck)-1]
		if _, ok := checked[candidate]; ok {
			continue
		}
		if _, ok := query[candidate]; ok {
			return true
		}
		toCheck = append(toCheck, entities[candidate].Parents...)
		checked[candidate] = struct{}{}
	}
	return false
}

func (n *inEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalEntity(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}

	rhs, err := n.rhs.Eval(ctx)
	if err != nil {
		return types.ZeroValue(), err
	}

	query := map[types.EntityUID]struct{}{}
	switch rhsv := rhs.(type) {
	case types.EntityUID:
		query[rhsv] = struct{}{}
	case types.Set:
		for _, rhv := range rhsv {
			e, err := types.ValueToEntity(rhv)
			if err != nil {
				return types.ZeroValue(), err
			}
			query[e] = struct{}{}
		}
	default:
		return types.ZeroValue(), fmt.Errorf(
			"%w: expected one of [set, (entity of type `any_entity_type`)], got %v", types.ErrType, rhs.TypeName())
	}
	return types.Boolean(entityIn(lhs, query, ctx.Entities)), nil
}

// isEval
type isEval struct {
	lhs, rhs Evaler
}

func newIsEval(lhs, rhs Evaler) *isEval {
	return &isEval{lhs: lhs, rhs: rhs}
}

func (n *isEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalEntity(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}

	rhs, err := evalPath(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}

	return types.Boolean(types.Path(lhs.Type) == rhs), nil
}

// decimalLiteralEval
type decimalLiteralEval struct {
	literal Evaler
}

func newDecimalLiteralEval(literal Evaler) *decimalLiteralEval {
	return &decimalLiteralEval{literal: literal}
}

func (n *decimalLiteralEval) Eval(ctx *EvalContext) (types.Value, error) {
	literal, err := evalString(n.literal, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}

	d, err := types.ParseDecimal(string(literal))
	if err != nil {
		return types.ZeroValue(), err
	}

	return d, nil
}

type ipLiteralEval struct {
	literal Evaler
}

func newIPLiteralEval(literal Evaler) *ipLiteralEval {
	return &ipLiteralEval{literal: literal}
}

func (n *ipLiteralEval) Eval(ctx *EvalContext) (types.Value, error) {
	literal, err := evalString(n.literal, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}

	i, err := types.ParseIPAddr(string(literal))
	if err != nil {
		return types.ZeroValue(), err
	}

	return i, nil
}

type ipTestType func(v types.IPAddr) bool

func ipTestIPv4(v types.IPAddr) bool      { return v.IsIPv4() }
func ipTestIPv6(v types.IPAddr) bool      { return v.IsIPv6() }
func ipTestLoopback(v types.IPAddr) bool  { return v.IsLoopback() }
func ipTestMulticast(v types.IPAddr) bool { return v.IsMulticast() }

// ipTestEval
type ipTestEval struct {
	object Evaler
	test   ipTestType
}

func newIPTestEval(object Evaler, test ipTestType) *ipTestEval {
	return &ipTestEval{object: object, test: test}
}

func (n *ipTestEval) Eval(ctx *EvalContext) (types.Value, error) {
	i, err := evalIP(n.object, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(n.test(i)), nil
}

// ipIsInRangeEval

type ipIsInRangeEval struct {
	lhs, rhs Evaler
}

func newIPIsInRangeEval(lhs, rhs Evaler) *ipIsInRangeEval {
	return &ipIsInRangeEval{lhs: lhs, rhs: rhs}
}

func (n *ipIsInRangeEval) Eval(ctx *EvalContext) (types.Value, error) {
	lhs, err := evalIP(n.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	rhs, err := evalIP(n.rhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(rhs.Contains(lhs)), nil
}