package eval

import (
	"errors"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/internal/mapset"
	"github.com/cedar-policy/cedar-go/types"
)

var errOverflow = fmt.Errorf("integer overflow")
var errUnknownExtensionFunction = fmt.Errorf("function does not exist")
var errArity = fmt.Errorf("wrong number of arguments provided to extension function")
var errAttributeAccess = fmt.Errorf("does not have the attribute")
var errTagAccess = fmt.Errorf("does not have the tag")
var errEntityNotExist = fmt.Errorf("does not exist")
var errUnspecifiedEntity = fmt.Errorf("unspecified entity")

func zeroValue() types.Value {
	return nil
}

type Env struct {
	Entities                    types.EntityGetter
	Principal, Action, Resource types.Value
	Context                     types.Value
}

type Evaler interface {
	Eval(Env) (types.Value, error)
}

func evalBool(n Evaler, env Env) (types.Boolean, error) {
	v, err := n.Eval(env)
	if err != nil {
		return false, err
	}
	b, err := ValueToBool(v)
	if err != nil {
		return false, err
	}
	return b, nil
}

func evalLong(n Evaler, env Env) (types.Long, error) {
	v, err := n.Eval(env)
	if err != nil {
		return 0, err
	}
	l, err := ValueToLong(v)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func evalComparableValue(n Evaler, env Env) (ComparableValue, error) {
	v, err := n.Eval(env)
	if err != nil {
		return nil, err
	}
	l, ok := v.(ComparableValue)
	if !ok {
		return nil, fmt.Errorf("%w: expected comparable value, got %v", ErrType, TypeName(v))
	}
	return l, nil
}

func evalString(n Evaler, env Env) (types.String, error) {
	v, err := n.Eval(env)
	if err != nil {
		return "", err
	}
	s, err := ValueToString(v)
	if err != nil {
		return "", err
	}
	return s, nil
}

func evalSet(n Evaler, env Env) (types.Set, error) {
	v, err := n.Eval(env)
	if err != nil {
		return types.Set{}, err
	}
	s, err := ValueToSet(v)
	if err != nil {
		return types.Set{}, err
	}
	return s, nil
}

func evalEntity(n Evaler, env Env) (types.EntityUID, error) {
	v, err := n.Eval(env)
	if err != nil {
		return types.EntityUID{}, err
	}
	e, err := ValueToEntity(v)
	if err != nil {
		return types.EntityUID{}, err
	}
	return e, nil
}

func evalDatetime(n Evaler, env Env) (types.Datetime, error) {
	v, err := n.Eval(env)
	if err != nil {
		return types.Datetime{}, err
	}
	d, err := ValueToDatetime(v)
	if err != nil {
		return types.Datetime{}, err
	}
	return d, nil
}

func evalDecimal(n Evaler, env Env) (types.Decimal, error) {
	v, err := n.Eval(env)
	if err != nil {
		return types.Decimal{}, err
	}
	d, err := ValueToDecimal(v)
	if err != nil {
		return types.Decimal{}, err
	}
	return d, nil
}

func evalDuration(n Evaler, env Env) (types.Duration, error) {
	v, err := n.Eval(env)
	if err != nil {
		return types.Duration{}, err
	}
	d, err := ValueToDuration(v)
	if err != nil {
		return types.Duration{}, err
	}
	return d, nil
}

func evalIP(n Evaler, env Env) (types.IPAddr, error) {
	v, err := n.Eval(env)

	if err != nil {
		return types.IPAddr{}, err
	}
	i, err := ValueToIP(v)
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

func (n *errorEval) Eval(Env) (types.Value, error) {
	return zeroValue(), n.err
}

// literalEval
type literalEval struct {
	value types.Value
}

func newLiteralEval(value types.Value) *literalEval {
	return &literalEval{value: value}
}

func (n *literalEval) Eval(Env) (types.Value, error) {
	return n.value, nil
}

// orEval
type orEval struct {
	lhs Evaler
	rhs Evaler
}

func newOrEval(lhs Evaler, rhs Evaler) Evaler {
	return &orEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *orEval) Eval(env Env) (types.Value, error) {
	v, err := n.lhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	b, err := ValueToBool(v)
	if err != nil {
		return zeroValue(), err
	}
	if b {
		return v, nil
	}
	v, err = n.rhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	_, err = ValueToBool(v)
	if err != nil {
		return zeroValue(), err
	}
	return v, nil
}

// andEval
type andEval struct {
	lhs Evaler
	rhs Evaler
}

func newAndEval(lhs Evaler, rhs Evaler) Evaler {
	return &andEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *andEval) Eval(env Env) (types.Value, error) {
	v, err := n.lhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	b, err := ValueToBool(v)
	if err != nil {
		return zeroValue(), err
	}
	if !b {
		return v, nil
	}
	v, err = n.rhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	_, err = ValueToBool(v)
	if err != nil {
		return zeroValue(), err
	}
	return v, nil
}

// notEval
type notEval struct {
	inner Evaler
}

func newNotEval(inner Evaler) Evaler {
	return &notEval{
		inner: inner,
	}
}

func (n *notEval) Eval(env Env) (types.Value, error) {
	v, err := n.inner.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	b, err := ValueToBool(v)
	if err != nil {
		return zeroValue(), err
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

func newAddEval(lhs Evaler, rhs Evaler) Evaler {
	return &addEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *addEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalLong(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	res, ok := checkedAddI64(lhs, rhs)
	if !ok {
		return zeroValue(), fmt.Errorf("%w while attempting to add `%d` with `%d`", errOverflow, lhs, rhs)
	}
	return res, nil
}

// subtractEval
type subtractEval struct {
	lhs Evaler
	rhs Evaler
}

func newSubtractEval(lhs Evaler, rhs Evaler) Evaler {
	return &subtractEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *subtractEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalLong(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	res, ok := checkedSubI64(lhs, rhs)
	if !ok {
		return zeroValue(), fmt.Errorf("%w while attempting to subtract `%d` from `%d`", errOverflow, rhs, lhs)
	}
	return res, nil
}

// multiplyEval
type multiplyEval struct {
	lhs Evaler
	rhs Evaler
}

func newMultiplyEval(lhs Evaler, rhs Evaler) Evaler {
	return &multiplyEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *multiplyEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalLong(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	res, ok := checkedMulI64(lhs, rhs)
	if !ok {
		return zeroValue(), fmt.Errorf("%w while attempting to multiply `%d` by `%d`", errOverflow, lhs, rhs)
	}
	return res, nil
}

// negateEval
type negateEval struct {
	inner Evaler
}

func newNegateEval(inner Evaler) Evaler {
	return &negateEval{
		inner: inner,
	}
}

func (n *negateEval) Eval(env Env) (types.Value, error) {
	inner, err := evalLong(n.inner, env)
	if err != nil {
		return zeroValue(), err
	}
	res, ok := checkedNegI64(inner)
	if !ok {
		return zeroValue(), fmt.Errorf("%w while attempting to negate `%d`", errOverflow, inner)
	}
	return res, nil
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

func (n *decimalLessThanEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Compare(rhs) == -1), nil
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

func (n *decimalLessThanOrEqualEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Compare(rhs) <= 0), nil
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

func (n *decimalGreaterThanEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Compare(rhs) == 1), nil
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

func (n *decimalGreaterThanOrEqualEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Compare(rhs) >= 0), nil
}

// ifThenElseEval
type ifThenElseEval struct {
	ifNode   Evaler
	thenNode Evaler
	elseNode Evaler
}

func newIfThenElseEval(ifNode, thenNode, elseNode Evaler) *ifThenElseEval {
	return &ifThenElseEval{
		ifNode:   ifNode,
		thenNode: thenNode,
		elseNode: elseNode,
	}
}

func (n *ifThenElseEval) Eval(env Env) (types.Value, error) {
	cond, err := evalBool(n.ifNode, env)
	if err != nil {
		return zeroValue(), err
	}
	if cond {
		return n.thenNode.Eval(env)
	}
	return n.elseNode.Eval(env)
}

// notEqualNode
type equalEval struct {
	lhs, rhs Evaler
}

func newEqualEval(lhs, rhs Evaler) Evaler {
	return &equalEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *equalEval) Eval(env Env) (types.Value, error) {
	lv, err := n.lhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	rv, err := n.rhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lv.Equal(rv)), nil
}

// notEqualEval
type notEqualEval struct {
	lhs, rhs Evaler
}

func newNotEqualEval(lhs, rhs Evaler) Evaler {
	return &notEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *notEqualEval) Eval(env Env) (types.Value, error) {
	lv, err := n.lhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	rv, err := n.rhs.Eval(env)
	if err != nil {
		return zeroValue(), err
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

func (n *setLiteralEval) Eval(env Env) (types.Value, error) {
	vals := make([]types.Value, len(n.elements))
	for i, e := range n.elements {
		v, err := e.Eval(env)
		if err != nil {
			return zeroValue(), err
		}
		vals[i] = v
	}
	return types.NewSet(vals...), nil
}

// containsEval
type containsEval struct {
	lhs, rhs Evaler
}

func newContainsEval(lhs, rhs Evaler) Evaler {
	return &containsEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalSet(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := n.rhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Contains(rhs)), nil
}

// containsAllEval
type containsAllEval struct {
	lhs, rhs Evaler
}

func newContainsAllEval(lhs, rhs Evaler) Evaler {
	return &containsAllEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsAllEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalSet(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalSet(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	for e := range rhs.All() {
		if !lhs.Contains(e) {
			return types.Boolean(false), nil
		}
	}
	return types.Boolean(true), nil
}

// containsAnyEval
type containsAnyEval struct {
	lhs, rhs Evaler
}

func newContainsAnyEval(lhs, rhs Evaler) Evaler {
	return &containsAnyEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsAnyEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalSet(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalSet(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	for e := range rhs.All() {
		if lhs.Contains(e) {
			return types.Boolean(true), nil
		}
	}
	return types.Boolean(false), nil
}

// isEmptyEval
type isEmptyEval struct {
	lhs Evaler
}

func newIsEmptyEval(lhs Evaler) Evaler {
	return &isEmptyEval{lhs: lhs}
}

func (n *isEmptyEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalSet(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Len() == 0), nil
}

// recordLiteralEval
type recordLiteralEval struct {
	elements map[types.String]Evaler
}

func newRecordLiteralEval(elements map[types.String]Evaler) *recordLiteralEval {
	return &recordLiteralEval{elements: elements}
}

func (n *recordLiteralEval) Eval(env Env) (types.Value, error) {
	vals := types.RecordMap{}
	for k, en := range n.elements {
		v, err := en.Eval(env)
		if err != nil {
			return zeroValue(), err
		}
		vals[k] = v
	}
	return types.NewRecord(vals), nil
}

// attributeAccessEval
type attributeAccessEval struct {
	object    Evaler
	attribute types.String
}

func newAttributeAccessEval(record Evaler, attribute types.String) *attributeAccessEval {
	return &attributeAccessEval{object: record, attribute: attribute}
}

func (n *attributeAccessEval) Eval(env Env) (types.Value, error) {
	v, err := n.object.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	switch vv := v.(type) {
	case types.EntityUID:
		var unspecified types.EntityUID
		if vv == unspecified {
			return zeroValue(), fmt.Errorf("cannot access attribute `%s` of %w", n.attribute, errUnspecifiedEntity)
		}
		rec, ok := env.Entities.Get(vv)
		if !ok {
			return zeroValue(), fmt.Errorf("entity `%v` %w", vv.String(), errEntityNotExist)
		}
		val, ok := rec.Attributes.Get(n.attribute)
		if !ok {
			return zeroValue(), fmt.Errorf("`%s` %w `%s`", vv.String(), errAttributeAccess, n.attribute)
		}
		return val, nil
	case types.Record:
		val, ok := vv.Get(n.attribute)
		if !ok {
			return zeroValue(), fmt.Errorf("record %w `%s`", errAttributeAccess, n.attribute)
		}
		return val, nil
	default:
		return zeroValue(), fmt.Errorf("%w: expected one of [record, (entity of type `any_entity_type`)], got %v", ErrType, TypeName(v))
	}
}

// hasEval
type hasEval struct {
	object    Evaler
	attribute types.String
}

func newHasEval(record Evaler, attribute types.String) *hasEval {
	return &hasEval{object: record, attribute: attribute}
}

func (n *hasEval) Eval(env Env) (types.Value, error) {
	v, err := n.object.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	var record types.Record
	switch vv := v.(type) {
	case types.EntityUID:
		if rec, ok := env.Entities.Get(vv); ok {
			record = rec.Attributes
		}
	case types.Record:
		record = vv
	default:
		return zeroValue(), fmt.Errorf("%w: expected one of [record, (entity of type `any_entity_type`)], got %v", ErrType, TypeName(v))
	}
	_, ok := record.Get(n.attribute)
	return types.Boolean(ok), nil
}

// getTagEval
type getTagEval struct {
	lhs, rhs Evaler
}

func newGetTagEval(object, tag Evaler) *getTagEval {
	return &getTagEval{lhs: object, rhs: tag}
}

func (n *getTagEval) Eval(env Env) (types.Value, error) {
	eid, err := evalEntity(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}

	var zero types.EntityUID
	if eid == zero {
		return zeroValue(), fmt.Errorf("cannot access tag `%s` of %w", n.rhs, errUnspecifiedEntity)
	}

	t, err := evalString(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}

	e, ok := env.Entities.Get(eid)
	if !ok {
		return zeroValue(), fmt.Errorf("entity `%v` %w", eid.String(), errEntityNotExist)
	}

	val, ok := e.Tags.Get(t)
	if !ok {
		return zeroValue(), fmt.Errorf("`%s` %w `%s`", eid.String(), errTagAccess, t)
	}

	return val, nil
}

// hasTagEval
type hasTagEval struct {
	lhs, rhs Evaler
}

func newHasTagEval(object, tag Evaler) *hasTagEval {
	return &hasTagEval{lhs: object, rhs: tag}
}

func (n *hasTagEval) Eval(env Env) (types.Value, error) {
	eid, err := evalEntity(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}

	t, err := evalString(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}

	e, ok := env.Entities.Get(eid)
	if !ok {
		return types.False, nil
	}

	_, ok = e.Tags.Get(t)
	return types.Boolean(ok), nil
}

// likeEval
type likeEval struct {
	lhs     Evaler
	pattern types.Pattern
}

func newLikeEval(lhs Evaler, pattern types.Pattern) *likeEval {
	return &likeEval{lhs: lhs, pattern: pattern}
}

func (l *likeEval) Eval(env Env) (types.Value, error) {
	v, err := evalString(l.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(l.pattern.Match(v)), nil
}

// variableEval
type variableEval struct {
	variableName types.String
}

func newVariableEval(variableName types.String) *variableEval {
	return &variableEval{variableName: variableName}
}

func (n *variableEval) Eval(env Env) (types.Value, error) {
	switch n.variableName {
	case consts.Principal:
		return env.Principal, nil
	case consts.Action:
		return env.Action, nil
	case consts.Resource:
		return env.Resource, nil
	default: // context
		return env.Context, nil
	}
}

// inEval
type inEval struct {
	lhs, rhs Evaler
}

func newInEval(lhs, rhs Evaler) Evaler {
	return &inEval{lhs: lhs, rhs: rhs}
}

func entityInOne(env Env, entity types.EntityUID, parent types.EntityUID) bool {
	if entity == parent {
		return true
	}
	var known mapset.MapSet[types.EntityUID]
	var todo []types.EntityUID
	var candidate = entity
	for {
		if fe, ok := env.Entities.Get(candidate); ok {
			if fe.Parents.Contains(parent) {
				return true
			}
			for k := range fe.Parents.All() {
				p, ok := env.Entities.Get(k)
				if !ok || p.Parents.Len() == 0 || k == entity || known.Contains(k) {
					continue
				}
				todo = append(todo, k)
				known.Add(k)
			}
		}
		if len(todo) == 0 {
			return false
		}
		candidate, todo = todo[len(todo)-1], todo[:len(todo)-1]
	}
}

func entityInSet(env Env, entity types.EntityUID, parents mapset.Container[types.EntityUID]) bool {
	if parents.Contains(entity) {
		return true
	}
	var known mapset.MapSet[types.EntityUID]
	var todo []types.EntityUID
	var candidate = entity
	for {
		if fe, ok := env.Entities.Get(candidate); ok {
			if fe.Parents.Intersects(parents) {
				return true
			}
			for k := range fe.Parents.All() {
				p, ok := env.Entities.Get(k)
				if !ok || p.Parents.Len() == 0 || k == entity || known.Contains(k) {
					continue
				}
				todo = append(todo, k)
				known.Add(k)
			}
		}
		if len(todo) == 0 {
			return false
		}
		candidate, todo = todo[len(todo)-1], todo[:len(todo)-1]
	}
}

func (n *inEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalEntity(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}

	rhs, err := n.rhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}

	return doInEval(env, lhs, rhs)
}

func doInEval(env Env, lhs types.EntityUID, rhs types.Value) (types.Value, error) {
	switch rhsv := rhs.(type) {
	case types.EntityUID:
		return types.Boolean(entityInOne(env, lhs, rhsv)), nil
	case types.Set:
		query := mapset.Make[types.EntityUID](rhsv.Len())
		for rhv := range rhsv.All() {
			e, err := ValueToEntity(rhv)
			if err != nil {
				return zeroValue(), err
			}
			query.Add(e)
		}
		return types.Boolean(entityInSet(env, lhs, query)), nil
	}
	return zeroValue(), fmt.Errorf(
		"%w: expected one of [set, (entity of type `any_entity_type`)], got %v", ErrType, TypeName(rhs))
}

// isEval
type isEval struct {
	lhs Evaler
	rhs types.EntityType
}

func newIsEval(lhs Evaler, rhs types.EntityType) *isEval {
	return &isEval{lhs: lhs, rhs: rhs}
}

func (n *isEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalEntity(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Type == n.rhs), nil
}

// isInEval
type isInEval struct {
	lhs Evaler
	is  types.EntityType
	rhs Evaler
}

func newIsInEval(lhs Evaler, is types.EntityType, rhs Evaler) Evaler {
	return &isInEval{lhs: lhs, is: is, rhs: rhs}
}

func (n *isInEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalEntity(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	if lhs.Type != n.is {
		return types.False, nil
	}
	rhs, err := n.rhs.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	return doInEval(env, lhs, rhs)
}

// decimalLiteralEval
type decimalLiteralEval struct {
	literal Evaler
}

func newDecimalLiteralEval(literal Evaler) *decimalLiteralEval {
	return &decimalLiteralEval{literal: literal}
}

func (n *decimalLiteralEval) Eval(env Env) (types.Value, error) {
	literal, err := evalString(n.literal, env)
	if err != nil {
		return zeroValue(), err
	}

	d, err := types.ParseDecimal(string(literal))
	if err != nil {
		return zeroValue(), err
	}

	return d, nil
}

// datetimeLiteralEval
type datetimeLiteralEval struct {
	literal Evaler
}

func newDatetimeLiteralEval(literal Evaler) *datetimeLiteralEval {
	return &datetimeLiteralEval{literal: literal}
}

func (n *datetimeLiteralEval) Eval(env Env) (types.Value, error) {
	literal, err := evalString(n.literal, env)
	if err != nil {
		return zeroValue(), err
	}

	d, err := types.ParseDatetime(string(literal))
	if err != nil {
		return zeroValue(), err
	}

	return d, nil
}

type durationLiteralEval struct {
	literal Evaler
}

func newDurationLiteralEval(literal Evaler) *durationLiteralEval {
	return &durationLiteralEval{literal: literal}
}

func (n *durationLiteralEval) Eval(env Env) (types.Value, error) {
	literal, err := evalString(n.literal, env)
	if err != nil {
		return zeroValue(), err
	}

	d, err := types.ParseDuration(string(literal))
	if err != nil {
		return zeroValue(), err
	}

	return d, nil
}

type ipLiteralEval struct {
	literal Evaler
}

func newIPLiteralEval(literal Evaler) *ipLiteralEval {
	return &ipLiteralEval{literal: literal}
}

func (n *ipLiteralEval) Eval(env Env) (types.Value, error) {
	literal, err := evalString(n.literal, env)
	if err != nil {
		return zeroValue(), err
	}

	i, err := types.ParseIPAddr(string(literal))
	if err != nil {
		return zeroValue(), err
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

func (n *ipTestEval) Eval(env Env) (types.Value, error) {
	i, err := evalIP(n.object, env)
	if err != nil {
		return zeroValue(), err
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

func (n *ipIsInRangeEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalIP(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalIP(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(rhs.Contains(lhs)), nil
}

// extensionEval

func newExtensionEval(name types.Path, args []Evaler) Evaler {
	// error is not part of the cedar spec, so leaving it out of the list of extensions that can be parsed, etc
	if name == partialErrorName && len(args) == 1 {
		return newPartialErrorEval(args[0])
	}

	if i, ok := extensions.ExtMap[name]; ok {
		if i.Args != len(args) {
			return newErrorEval(fmt.Errorf("%w: %s takes %d parameter(s), but %d provided", errArity, name, i.Args, len(args)))
		}
		switch name {
		case "datetime":
			return newDatetimeLiteralEval(args[0])
		case "decimal":
			return newDecimalLiteralEval(args[0])
		case "duration":
			return newDurationLiteralEval(args[0])
		case "ip":
			return newIPLiteralEval(args[0])

		case "lessThan":
			return newDecimalLessThanEval(args[0], args[1])
		case "lessThanOrEqual":
			return newDecimalLessThanOrEqualEval(args[0], args[1])
		case "greaterThan":
			return newDecimalGreaterThanEval(args[0], args[1])
		case "greaterThanOrEqual":
			return newDecimalGreaterThanOrEqualEval(args[0], args[1])

		case "isIpv4":
			return newIPTestEval(args[0], ipTestIPv4)
		case "isIpv6":
			return newIPTestEval(args[0], ipTestIPv6)
		case "isLoopback":
			return newIPTestEval(args[0], ipTestLoopback)
		case "isMulticast":
			return newIPTestEval(args[0], ipTestMulticast)
		case "isInRange":
			return newIPIsInRangeEval(args[0], args[1])

		case "toDate":
			return newToDateEval(args[0])
		case "toTime":
			return newToTimeEval(args[0])
		case "toMilliseconds":
			return newToMillisecondsEval(args[0])
		case "toSeconds":
			return newToSecondsEval(args[0])
		case "toMinutes":
			return newToMinutesEval(args[0])
		case "toHours":
			return newToHoursEval(args[0])
		case "toDays":
			return newToDaysEval(args[0])

		case "offset":
			return newOffsetEval(args[0], args[1])
		case "durationSince":
			return newDurationSinceEval(args[0], args[1])
		}
	}
	return newErrorEval(fmt.Errorf("%w: %s", errUnknownExtensionFunction, name))
}

// comparableValueLessThanEval struct
type comparableValueLessThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newComparableValueLessThanEval(lhs Evaler, rhs Evaler) Evaler {
	return &comparableValueLessThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *comparableValueLessThanEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalComparableValue(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalComparableValue(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	ok, err := lhs.LessThan(rhs)
	if err != nil {
		return types.False, fmt.Errorf("%w: %w", ErrType, err)
	}
	return types.Boolean(ok), nil
}

var _ Evaler = &comparableValueLessThanEval{}

// comparableValueGreaterThanEval struct
type comparableValueGreaterThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newComparableValueGreaterThanEval(lhs Evaler, rhs Evaler) Evaler {
	return &comparableValueGreaterThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *comparableValueGreaterThanEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalComparableValue(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalComparableValue(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	ok, err := lhs.LessThanOrEqual(rhs)
	if err != nil {
		return types.False, fmt.Errorf("%w: %w", ErrType, err)
	}
	return types.Boolean(!ok), nil
}

// comparableValueLessEqualThanEval struct
type comparableValueLessThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newComparableValueLessThanOrEqualEval(lhs Evaler, rhs Evaler) Evaler {
	return &comparableValueLessThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *comparableValueLessThanOrEqualEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalComparableValue(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalComparableValue(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	ok, err := lhs.LessThanOrEqual(rhs)
	if err != nil {
		return types.False, fmt.Errorf("%w: %w", ErrType, err)
	}
	return types.Boolean(ok), nil
}

// comparableValueGreaterEqualThanEval struct
type comparableValueGreaterThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newComparableValueGreaterThanOrEqualEval(lhs Evaler, rhs Evaler) Evaler {
	return &comparableValueGreaterThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *comparableValueGreaterThanOrEqualEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalComparableValue(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalComparableValue(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	ok, err := lhs.LessThan(rhs)
	if err != nil {
		return types.False, errors.Join(ErrType, err)
	}
	return types.Boolean(!ok), nil
}

type toDateEval struct {
	lhs Evaler
}

func newToDateEval(lhs Evaler) *toDateEval {
	return &toDateEval{lhs: lhs}
}

func (n *toDateEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDatetime(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.NewDatetimeFromMillis(lhs.Milliseconds() - (lhs.Milliseconds() % consts.MillisPerDay)), nil
}

type toTimeEval struct {
	lhs Evaler
}

func newToTimeEval(lhs Evaler) *toTimeEval {
	return &toTimeEval{lhs: lhs}
}

func (n *toTimeEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDatetime(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.NewDurationFromMillis(lhs.Milliseconds() % consts.MillisPerDay), nil
}

type toMillisecondsEval struct {
	lhs Evaler
}

func newToMillisecondsEval(lhs Evaler) *toMillisecondsEval {
	return &toMillisecondsEval{lhs: lhs}
}

func (n *toMillisecondsEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDuration(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Long(lhs.ToMilliseconds()), nil
}

type toSecondsEval struct {
	lhs Evaler
}

func newToSecondsEval(lhs Evaler) *toSecondsEval {
	return &toSecondsEval{lhs: lhs}
}

func (n *toSecondsEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDuration(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Long(lhs.ToMilliseconds() / consts.MillisPerSecond), nil
}

type toMinutesEval struct {
	lhs Evaler
}

func newToMinutesEval(lhs Evaler) *toMinutesEval {
	return &toMinutesEval{lhs: lhs}
}

func (n *toMinutesEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDuration(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Long(lhs.ToMilliseconds() / consts.MillisPerMinute), nil
}

type toHoursEval struct {
	lhs Evaler
}

func newToHoursEval(lhs Evaler) *toHoursEval {
	return &toHoursEval{lhs: lhs}
}

func (n *toHoursEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDuration(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Long(lhs.ToMilliseconds() / consts.MillisPerHour), nil
}

type toDaysEval struct {
	lhs Evaler
}

func newToDaysEval(lhs Evaler) *toDaysEval {
	return &toDaysEval{lhs: lhs}
}

func (n *toDaysEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDuration(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Long(lhs.ToMilliseconds() / consts.MillisPerDay), nil
}

type offsetEval struct {
	lhs Evaler
	rhs Evaler
}

func newOffsetEval(lhs Evaler, rhs Evaler) *offsetEval {
	return &offsetEval{lhs: lhs, rhs: rhs}
}

func (n *offsetEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDatetime(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDuration(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.NewDatetimeFromMillis(lhs.Milliseconds() + rhs.ToMilliseconds()), nil
}

type durationSinceEval struct {
	lhs Evaler
	rhs Evaler
}

func newDurationSinceEval(lhs Evaler, rhs Evaler) *durationSinceEval {
	return &durationSinceEval{lhs: lhs, rhs: rhs}
}

func (n *durationSinceEval) Eval(env Env) (types.Value, error) {
	lhs, err := evalDatetime(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDatetime(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.NewDurationFromMillis(lhs.Milliseconds() - rhs.Milliseconds()), nil
}
