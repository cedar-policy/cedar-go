package cedar

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/parser"
)

var errOverflow = fmt.Errorf("integer overflow")
var errUnknownMethod = fmt.Errorf("unknown method")
var errUnknownExtensionFunction = fmt.Errorf("function does not exist")
var errArity = fmt.Errorf("wrong number of arguments provided to extension function")
var errAttributeAccess = fmt.Errorf("does not have the attribute")
var errEntityNotExist = fmt.Errorf("does not exist")
var errUnspecifiedEntity = fmt.Errorf("unspecified entity")

type evalContext struct {
	Entities                    Entities
	Principal, Action, Resource types.Value
	Context                     types.Value
}

type evaler interface {
	Eval(*evalContext) (types.Value, error)
}

func evalBool(n evaler, ctx *evalContext) (types.Boolean, error) {
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

func evalLong(n evaler, ctx *evalContext) (types.Long, error) {
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

func evalString(n evaler, ctx *evalContext) (types.String, error) {
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

func evalSet(n evaler, ctx *evalContext) (types.Set, error) {
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

func evalEntity(n evaler, ctx *evalContext) (types.EntityUID, error) {
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

func evalPath(n evaler, ctx *evalContext) (types.Path, error) {
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

func evalDecimal(n evaler, ctx *evalContext) (types.Decimal, error) {
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

func evalIP(n evaler, ctx *evalContext) (types.IPAddr, error) {
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

func (n *errorEval) Eval(_ *evalContext) (types.Value, error) {
	return types.ZeroValue(), n.err
}

// literalEval
type literalEval struct {
	value types.Value
}

func newLiteralEval(value types.Value) *literalEval {
	return &literalEval{value: value}
}

func (n *literalEval) Eval(_ *evalContext) (types.Value, error) {
	return n.value, nil
}

// orEval
type orEval struct {
	lhs evaler
	rhs evaler
}

func newOrNode(lhs evaler, rhs evaler) *orEval {
	return &orEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *orEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newAndEval(lhs evaler, rhs evaler) *andEval {
	return &andEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *andEval) Eval(ctx *evalContext) (types.Value, error) {
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
	inner evaler
}

func newNotEval(inner evaler) *notEval {
	return &notEval{
		inner: inner,
	}
}

func (n *notEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newAddEval(lhs evaler, rhs evaler) *addEval {
	return &addEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *addEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newSubtractEval(lhs evaler, rhs evaler) *subtractEval {
	return &subtractEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *subtractEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newMultiplyEval(lhs evaler, rhs evaler) *multiplyEval {
	return &multiplyEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *multiplyEval) Eval(ctx *evalContext) (types.Value, error) {
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
	inner evaler
}

func newNegateEval(inner evaler) *negateEval {
	return &negateEval{
		inner: inner,
	}
}

func (n *negateEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newLongLessThanEval(lhs evaler, rhs evaler) *longLessThanEval {
	return &longLessThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longLessThanEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newLongLessThanOrEqualEval(lhs evaler, rhs evaler) *longLessThanOrEqualEval {
	return &longLessThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longLessThanOrEqualEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newLongGreaterThanEval(lhs evaler, rhs evaler) *longGreaterThanEval {
	return &longGreaterThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longGreaterThanEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newLongGreaterThanOrEqualEval(lhs evaler, rhs evaler) *longGreaterThanOrEqualEval {
	return &longGreaterThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longGreaterThanOrEqualEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newDecimalLessThanEval(lhs evaler, rhs evaler) *decimalLessThanEval {
	return &decimalLessThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *decimalLessThanEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newDecimalLessThanOrEqualEval(lhs evaler, rhs evaler) *decimalLessThanOrEqualEval {
	return &decimalLessThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *decimalLessThanOrEqualEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newDecimalGreaterThanEval(lhs evaler, rhs evaler) *decimalGreaterThanEval {
	return &decimalGreaterThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *decimalGreaterThanEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs evaler
	rhs evaler
}

func newDecimalGreaterThanOrEqualEval(lhs evaler, rhs evaler) *decimalGreaterThanOrEqualEval {
	return &decimalGreaterThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *decimalGreaterThanOrEqualEval) Eval(ctx *evalContext) (types.Value, error) {
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
	if_   evaler
	then  evaler
	else_ evaler
}

func newIfThenElseEval(if_, then, else_ evaler) *ifThenElseEval {
	return &ifThenElseEval{
		if_:   if_,
		then:  then,
		else_: else_,
	}
}

func (n *ifThenElseEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs, rhs evaler
}

func newEqualEval(lhs, rhs evaler) *equalEval {
	return &equalEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *equalEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs, rhs evaler
}

func newNotEqualEval(lhs, rhs evaler) *notEqualEval {
	return &notEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *notEqualEval) Eval(ctx *evalContext) (types.Value, error) {
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
	elements []evaler
}

func newSetLiteralEval(elements []evaler) *setLiteralEval {
	return &setLiteralEval{elements: elements}
}

func (n *setLiteralEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs, rhs evaler
}

func newContainsEval(lhs, rhs evaler) *containsEval {
	return &containsEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs, rhs evaler
}

func newContainsAllEval(lhs, rhs evaler) *containsAllEval {
	return &containsAllEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsAllEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs, rhs evaler
}

func newContainsAnyEval(lhs, rhs evaler) *containsAnyEval {
	return &containsAnyEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsAnyEval) Eval(ctx *evalContext) (types.Value, error) {
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
	elements map[string]evaler
}

func newRecordLiteralEval(elements map[string]evaler) *recordLiteralEval {
	return &recordLiteralEval{elements: elements}
}

func (n *recordLiteralEval) Eval(ctx *evalContext) (types.Value, error) {
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
	object    evaler
	attribute string
}

func newAttributeAccessEval(record evaler, attribute string) *attributeAccessEval {
	return &attributeAccessEval{object: record, attribute: attribute}
}

func (n *attributeAccessEval) Eval(ctx *evalContext) (types.Value, error) {
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
	object    evaler
	attribute string
}

func newHasEval(record evaler, attribute string) *hasEval {
	return &hasEval{object: record, attribute: attribute}
}

func (n *hasEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs     evaler
	pattern parser.Pattern
}

func newLikeEval(lhs evaler, pattern parser.Pattern) *likeEval {
	return &likeEval{lhs: lhs, pattern: pattern}
}

func (l *likeEval) Eval(ctx *evalContext) (types.Value, error) {
	v, err := evalString(l.lhs, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(match(l.pattern, string(v))), nil
}

type variableName func(ctx *evalContext) types.Value

func variableNamePrincipal(ctx *evalContext) types.Value { return ctx.Principal }
func variableNameAction(ctx *evalContext) types.Value    { return ctx.Action }
func variableNameResource(ctx *evalContext) types.Value  { return ctx.Resource }
func variableNameContext(ctx *evalContext) types.Value   { return ctx.Context }

// variableEval
type variableEval struct {
	variableName variableName
}

func newVariableEval(variableName variableName) *variableEval {
	return &variableEval{variableName: variableName}
}

func (n *variableEval) Eval(ctx *evalContext) (types.Value, error) {
	return n.variableName(ctx), nil
}

// inEval
type inEval struct {
	lhs, rhs evaler
}

func newInEval(lhs, rhs evaler) *inEval {
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

func (n *inEval) Eval(ctx *evalContext) (types.Value, error) {
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
	lhs, rhs evaler
}

func newIsEval(lhs, rhs evaler) *isEval {
	return &isEval{lhs: lhs, rhs: rhs}
}

func (n *isEval) Eval(ctx *evalContext) (types.Value, error) {
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
	literal evaler
}

func newDecimalLiteralEval(literal evaler) *decimalLiteralEval {
	return &decimalLiteralEval{literal: literal}
}

func (n *decimalLiteralEval) Eval(ctx *evalContext) (types.Value, error) {
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
	literal evaler
}

func newIPLiteralEval(literal evaler) *ipLiteralEval {
	return &ipLiteralEval{literal: literal}
}

func (n *ipLiteralEval) Eval(ctx *evalContext) (types.Value, error) {
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
	object evaler
	test   ipTestType
}

func newIPTestEval(object evaler, test ipTestType) *ipTestEval {
	return &ipTestEval{object: object, test: test}
}

func (n *ipTestEval) Eval(ctx *evalContext) (types.Value, error) {
	i, err := evalIP(n.object, ctx)
	if err != nil {
		return types.ZeroValue(), err
	}
	return types.Boolean(n.test(i)), nil
}

// ipIsInRangeEval

type ipIsInRangeEval struct {
	lhs, rhs evaler
}

func newIPIsInRangeEval(lhs, rhs evaler) *ipIsInRangeEval {
	return &ipIsInRangeEval{lhs: lhs, rhs: rhs}
}

func (n *ipIsInRangeEval) Eval(ctx *evalContext) (types.Value, error) {
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
