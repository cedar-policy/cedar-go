package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/types"
)

var errOverflow = fmt.Errorf("integer overflow")
var errUnknownExtensionFunction = fmt.Errorf("function does not exist")
var errArity = fmt.Errorf("wrong number of arguments provided to extension function")
var errAttributeAccess = fmt.Errorf("does not have the attribute")
var errEntityNotExist = fmt.Errorf("does not exist")
var errUnspecifiedEntity = fmt.Errorf("unspecified entity")

func zeroValue() types.Value {
	return nil
}

type Env struct {
	Entities                    types.Entities
	Principal, Action, Resource types.Value
	Env                         types.Value

	inCache map[inKey]bool
}

type inKey struct {
	a, b types.EntityUID
}

func NewEnv() *Env {
	return InitEnv(&Env{})
}

func InitEnv(in *Env) *Env {
	// add caches if applicable
	in.inCache = map[inKey]bool{}
	return in
}

func InitEnvWithCacheFrom(in *Env, parent *Env) *Env {
	in.inCache = parent.inCache
	return in
}

type Evaler interface {
	Eval(*Env) (types.Value, error)
}

func evalBool(n Evaler, env *Env) (types.Boolean, error) {
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

func evalLong(n Evaler, env *Env) (types.Long, error) {
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

func evalLesser(n Evaler, env *Env) (types.Lesser, error) {
	v, err := n.Eval(env)
	if err != nil {
		return nil, err
	}
	l, ok := v.(types.Lesser)
	if !ok {
		return nil, fmt.Errorf("%w: expected comparable value, got %v", ErrType, TypeName(v))
	}
	return l, nil
}

func evalString(n Evaler, env *Env) (types.String, error) {
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

func evalSet(n Evaler, env *Env) (types.Set, error) {
	v, err := n.Eval(env)
	if err != nil {
		return nil, err
	}
	s, err := ValueToSet(v)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func evalEntity(n Evaler, env *Env) (types.EntityUID, error) {
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

func evalDatetime(n Evaler, env *Env) (types.Datetime, error) {
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

func evalDecimal(n Evaler, env *Env) (types.Decimal, error) {
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

func evalIP(n Evaler, env *Env) (types.IPAddr, error) {
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

func (n *errorEval) Eval(_ *Env) (types.Value, error) {
	return zeroValue(), n.err
}

// literalEval
type literalEval struct {
	value types.Value
}

func newLiteralEval(value types.Value) *literalEval {
	return &literalEval{value: value}
}

func (n *literalEval) Eval(_ *Env) (types.Value, error) {
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

func (n *orEval) Eval(env *Env) (types.Value, error) {
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

func (n *andEval) Eval(env *Env) (types.Value, error) {
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

func (n *notEval) Eval(env *Env) (types.Value, error) {
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

func (n *addEval) Eval(env *Env) (types.Value, error) {
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

func (n *subtractEval) Eval(env *Env) (types.Value, error) {
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

func (n *multiplyEval) Eval(env *Env) (types.Value, error) {
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

func (n *negateEval) Eval(env *Env) (types.Value, error) {
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

// longLessThanEval
type longLessThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newLongLessThanEval(lhs Evaler, rhs Evaler) Evaler {
	return &longLessThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longLessThanEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalLong(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs < rhs), nil
}

// longLessThanOrEqualEval
type longLessThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newLongLessThanOrEqualEval(lhs Evaler, rhs Evaler) Evaler {
	return &longLessThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longLessThanOrEqualEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalLong(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs <= rhs), nil
}

// longGreaterThanEval
type longGreaterThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newLongGreaterThanEval(lhs Evaler, rhs Evaler) Evaler {
	return &longGreaterThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longGreaterThanEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalLong(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs > rhs), nil
}

// longGreaterThanOrEqualEval
type longGreaterThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newLongGreaterThanOrEqualEval(lhs Evaler, rhs Evaler) Evaler {
	return &longGreaterThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longGreaterThanOrEqualEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalLong(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, env)
	if err != nil {
		return zeroValue(), err
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

func (n *decimalLessThanEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Value < rhs.Value), nil
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

func (n *decimalLessThanOrEqualEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Value <= rhs.Value), nil
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

func (n *decimalGreaterThanEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Value > rhs.Value), nil
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

func (n *decimalGreaterThanOrEqualEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalDecimal(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Value >= rhs.Value), nil
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

func (n *ifThenElseEval) Eval(env *Env) (types.Value, error) {
	cond, err := evalBool(n.if_, env)
	if err != nil {
		return zeroValue(), err
	}
	if cond {
		return n.then.Eval(env)
	}
	return n.else_.Eval(env)
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

func (n *equalEval) Eval(env *Env) (types.Value, error) {
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

func (n *notEqualEval) Eval(env *Env) (types.Value, error) {
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

func (n *setLiteralEval) Eval(env *Env) (types.Value, error) {
	var vals types.Set
	for _, e := range n.elements {
		v, err := e.Eval(env)
		if err != nil {
			return zeroValue(), err
		}
		vals = append(vals, v)
	}
	return vals, nil
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

func (n *containsEval) Eval(env *Env) (types.Value, error) {
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

func (n *containsAllEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalSet(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalSet(n.rhs, env)
	if err != nil {
		return zeroValue(), err
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

func newContainsAnyEval(lhs, rhs Evaler) Evaler {
	return &containsAnyEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *containsAnyEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalSet(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalSet(n.rhs, env)
	if err != nil {
		return zeroValue(), err
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
	elements map[types.String]Evaler
}

func newRecordLiteralEval(elements map[types.String]Evaler) *recordLiteralEval {
	return &recordLiteralEval{elements: elements}
}

func (n *recordLiteralEval) Eval(env *Env) (types.Value, error) {
	vals := types.Record{}
	for k, en := range n.elements {
		v, err := en.Eval(env)
		if err != nil {
			return zeroValue(), err
		}
		vals[k] = v
	}
	return vals, nil
}

// attributeAccessEval
type attributeAccessEval struct {
	object    Evaler
	attribute types.String
}

func newAttributeAccessEval(record Evaler, attribute types.String) *attributeAccessEval {
	return &attributeAccessEval{object: record, attribute: attribute}
}

func (n *attributeAccessEval) Eval(env *Env) (types.Value, error) {
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
		rec, ok := env.Entities[vv]
		if !ok {
			return zeroValue(), fmt.Errorf("entity `%v` %w", vv.String(), errEntityNotExist)
		}
		val, ok := rec.Attributes[n.attribute]
		if !ok {
			return zeroValue(), fmt.Errorf("`%s` %w `%s`", vv.String(), errAttributeAccess, n.attribute)
		}
		return val, nil
	case types.Record:
		val, ok := vv[n.attribute]
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

func (n *hasEval) Eval(env *Env) (types.Value, error) {
	v, err := n.object.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	var record types.Record
	switch vv := v.(type) {
	case types.EntityUID:
		if rec, ok := env.Entities[vv]; ok {
			record = rec.Attributes
		}
	case types.Record:
		record = vv
	default:
		return zeroValue(), fmt.Errorf("%w: expected one of [record, (entity of type `any_entity_type`)], got %v", ErrType, TypeName(v))
	}
	_, ok := record[n.attribute]
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

func (l *likeEval) Eval(env *Env) (types.Value, error) {
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

func (n *variableEval) Eval(env *Env) (types.Value, error) {
	switch n.variableName {
	case consts.Principal:
		return env.Principal, nil
	case consts.Action:
		return env.Action, nil
	case consts.Resource:
		return env.Resource, nil
	default: // context
		return env.Env, nil
	}
}

// inEval
type inEval struct {
	lhs, rhs Evaler
}

func newInEval(lhs, rhs Evaler) Evaler {
	return &inEval{lhs: lhs, rhs: rhs}
}

func hasKnown(known map[types.EntityUID]struct{}, k types.EntityUID) bool {
	_, ok := known[k]
	return ok
}

func entityInOne(env *Env, entity types.EntityUID, parent types.EntityUID) bool {
	key := inKey{a: entity, b: parent}
	if cached, ok := env.inCache[key]; ok {
		return cached
	}
	result := entityInOneWork(env, entity, parent)
	env.inCache[key] = result
	return result
}
func entityInOneWork(env *Env, entity types.EntityUID, parent types.EntityUID) bool {
	if entity == parent {
		return true
	}
	var known map[types.EntityUID]struct{}
	var todo []types.EntityUID
	var candidate = entity
	for {
		if fe, ok := env.Entities[candidate]; ok {
			for _, k := range fe.Parents {
				if k == parent {
					return true
				}
			}
			for _, k := range fe.Parents {
				p, ok := env.Entities[k]
				if !ok || len(p.Parents) == 0 || k == entity || hasKnown(known, k) {
					continue
				}
				todo = append(todo, k)
				if known == nil {
					known = map[types.EntityUID]struct{}{}
				}
				known[k] = struct{}{}
			}
		}
		if len(todo) == 0 {
			return false
		}
		candidate, todo = todo[len(todo)-1], todo[:len(todo)-1]
	}
}

func entityInSet(env *Env, entity types.EntityUID, parents map[types.EntityUID]struct{}) bool {
	if _, ok := parents[entity]; ok {
		return true
	}
	var known map[types.EntityUID]struct{}
	var todo []types.EntityUID
	var candidate = entity
	for {
		if fe, ok := env.Entities[candidate]; ok {
			for _, k := range fe.Parents {
				if _, ok := parents[k]; ok {
					return true
				}
			}
			for _, k := range fe.Parents {
				p, ok := env.Entities[k]
				if !ok || len(p.Parents) == 0 || k == entity || hasKnown(known, k) {
					continue
				}
				todo = append(todo, k)
				if known == nil {
					known = map[types.EntityUID]struct{}{}
				}
				known[k] = struct{}{}
			}
		}
		if len(todo) == 0 {
			return false
		}
		candidate, todo = todo[len(todo)-1], todo[:len(todo)-1]
	}
}

func (n *inEval) Eval(env *Env) (types.Value, error) {
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

func doInEval(env *Env, lhs types.EntityUID, rhs types.Value) (types.Value, error) {
	switch rhsv := rhs.(type) {
	case types.EntityUID:
		return types.Boolean(entityInOne(env, lhs, rhsv)), nil
	case types.Set:
		query := make(map[types.EntityUID]struct{}, len(rhsv))
		for _, rhv := range rhsv {
			e, err := ValueToEntity(rhv)
			if err != nil {
				return zeroValue(), err
			}
			query[e] = struct{}{}
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

func (n *isEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalEntity(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Type == n.rhs), nil
}

type isInEval struct {
	lhs Evaler
	is  types.EntityType
	rhs Evaler
}

func newIsInEval(lhs Evaler, is types.EntityType, rhs Evaler) Evaler {
	return &isInEval{lhs: lhs, is: is, rhs: rhs}
}

func (n *isInEval) Eval(env *Env) (types.Value, error) {
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

func (n *decimalLiteralEval) Eval(env *Env) (types.Value, error) {
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

func (n *datetimeLiteralEval) Eval(env *Env) (types.Value, error) {
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

type ipLiteralEval struct {
	literal Evaler
}

func newIPLiteralEval(literal Evaler) *ipLiteralEval {
	return &ipLiteralEval{literal: literal}
}

func (n *ipLiteralEval) Eval(env *Env) (types.Value, error) {
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

func (n *ipTestEval) Eval(env *Env) (types.Value, error) {
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

func (n *ipIsInRangeEval) Eval(env *Env) (types.Value, error) {
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
			return newErrorEval(fmt.Errorf("%w: %s takes %d parameter(s)", errArity, name, i.Args))
		}
		switch {
		case name == "ip":
			return newIPLiteralEval(args[0])
		case name == "decimal":
			return newDecimalLiteralEval(args[0])

		case name == "lessThan":
			return newDecimalLessThanEval(args[0], args[1])
		case name == "lessThanOrEqual":
			return newDecimalLessThanOrEqualEval(args[0], args[1])
		case name == "greaterThan":
			return newDecimalGreaterThanEval(args[0], args[1])
		case name == "greaterThanOrEqual":
			return newDecimalGreaterThanOrEqualEval(args[0], args[1])

		case name == "isIpv4":
			return newIPTestEval(args[0], ipTestIPv4)
		case name == "isIpv6":
			return newIPTestEval(args[0], ipTestIPv6)
		case name == "isLoopback":
			return newIPTestEval(args[0], ipTestLoopback)
		case name == "isMulticast":
			return newIPTestEval(args[0], ipTestMulticast)
		case name == "isInRange":
			return newIPIsInRangeEval(args[0], args[1])

		case name == "toDate":
			return newToDateEval(args[0])
		}
	}
	return newErrorEval(fmt.Errorf("%w: %s", errUnknownExtensionFunction, name))
}

// virtualLessThanEval struct
type virtualLessThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newVirtualLessThanEval(lhs Evaler, rhs Evaler) *virtualLessThanEval {
	return &virtualLessThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *virtualLessThanEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalLesser(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLesser(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.Less(rhs)), nil
}

var _ Evaler = &virtualLessThanEval{}

// virtualGreaterThanEval struct
type virtualGreaterThanEval struct {
	lhs Evaler
	rhs Evaler
}

func newVirtualGreaterThanEval(lhs Evaler, rhs Evaler) *virtualGreaterThanEval {
	return &virtualGreaterThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *virtualGreaterThanEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalLesser(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLesser(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(!lhs.LessEqual(rhs)), nil
}

// virtualLessEqualThanEval struct
type virtualLessThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newVirtualLessThanOrEqualEval(lhs Evaler, rhs Evaler) *virtualLessThanOrEqualEval {
	return &virtualLessThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *virtualLessThanOrEqualEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalLesser(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLesser(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(lhs.LessEqual(rhs)), nil
}

// virtualGreaterEqualThanEval struct
type virtualGreaterThanOrEqualEval struct {
	lhs Evaler
	rhs Evaler
}

func newVirtualGreaterThanOrEqualEval(lhs Evaler, rhs Evaler) *virtualGreaterThanOrEqualEval {
	return &virtualGreaterThanOrEqualEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *virtualGreaterThanOrEqualEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalLesser(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLesser(n.rhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Boolean(!lhs.Less(rhs)), nil
}

const millisPerDay = int64(1000 * 60 * 60 * 24)

type toDateEval struct {
	lhs Evaler
}

func newToDateEval(lhs Evaler) *toDateEval {
	return &toDateEval{lhs: lhs}
}

func (n *toDateEval) Eval(env *Env) (types.Value, error) {
	lhs, err := evalDatetime(n.lhs, env)
	if err != nil {
		return zeroValue(), err
	}
	return types.Datetime{Value: lhs.Value - (lhs.Value % millisPerDay)}, nil
}
