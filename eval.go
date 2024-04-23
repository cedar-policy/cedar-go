package cedar

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/x/exp/parser"
)

var errOverflow = fmt.Errorf("integer overflow")
var errType = fmt.Errorf("type error")
var errUnknownMethod = fmt.Errorf("unknown method")
var errUnknownExtensionFunction = fmt.Errorf("function does not exist")
var errArity = fmt.Errorf("wrong number of arguments provided to extension function")
var errAttributeAccess = fmt.Errorf("does not have the attribute")
var errDecimal = fmt.Errorf("error parsing decimal value")
var errIP = fmt.Errorf("error parsing ip value")
var errEntityNotExist = fmt.Errorf("does not exist")
var errUnspecifiedEntity = fmt.Errorf("unspecified entity")

type evalContext struct {
	Entities                    Entities
	Principal, Action, Resource Value
	Context                     Value
}

type evaler interface {
	Eval(*evalContext) (Value, error)
}

func valueToBool(v Value) (Boolean, error) {
	bv, ok := v.(Boolean)
	if !ok {
		return false, fmt.Errorf("%w: expected bool, got %v", errType, v.typeName())
	}
	return bv, nil
}

func evalBool(n evaler, ctx *evalContext) (Boolean, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return false, err
	}
	b, err := valueToBool(v)
	if err != nil {
		return false, err
	}
	return b, nil
}

func valueToLong(v Value) (Long, error) {
	lv, ok := v.(Long)
	if !ok {
		return 0, fmt.Errorf("%w: expected long, got %v", errType, v.typeName())
	}
	return lv, nil
}

func evalLong(n evaler, ctx *evalContext) (Long, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return 0, err
	}
	l, err := valueToLong(v)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func valueToString(v Value) (String, error) {
	sv, ok := v.(String)
	if !ok {
		return "", fmt.Errorf("%w: expected string, got %v", errType, v.typeName())
	}
	return sv, nil
}

func evalString(n evaler, ctx *evalContext) (String, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return "", err
	}
	s, err := valueToString(v)
	if err != nil {
		return "", err
	}
	return s, nil
}

func valueToSet(v Value) (Set, error) {
	sv, ok := v.(Set)
	if !ok {
		return nil, fmt.Errorf("%w: expected set, got %v", errType, v.typeName())
	}
	return sv, nil
}

func evalSet(n evaler, ctx *evalContext) (Set, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return nil, err
	}
	s, err := valueToSet(v)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func valueToRecord(v Value) (Record, error) {
	rv, ok := v.(Record)
	if !ok {
		return nil, fmt.Errorf("%w: expected record got %v", errType, v.typeName())
	}
	return rv, nil
}

func valueToEntity(v Value) (EntityUID, error) {
	ev, ok := v.(EntityUID)
	if !ok {
		return EntityUID{}, fmt.Errorf("%w: expected (entity of type `any_entity_type`), got %v", errType, v.typeName())
	}
	return ev, nil
}

func valueToPath(v Value) (path, error) {
	ev, ok := v.(path)
	if !ok {
		return "", fmt.Errorf("%w: expected (path of type `any_entity_type`), got %v", errType, v.typeName())
	}
	return ev, nil
}

func evalEntity(n evaler, ctx *evalContext) (EntityUID, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return EntityUID{}, err
	}
	e, err := valueToEntity(v)
	if err != nil {
		return EntityUID{}, err
	}
	return e, nil
}

func evalPath(n evaler, ctx *evalContext) (path, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return "", err
	}
	e, err := valueToPath(v)
	if err != nil {
		return "", err
	}
	return e, nil
}

func valueToDecimal(v Value) (Decimal, error) {
	d, ok := v.(Decimal)
	if !ok {
		return 0, fmt.Errorf("%w: expected decimal, got %v", errType, v.typeName())
	}
	return d, nil
}

func evalDecimal(n evaler, ctx *evalContext) (Decimal, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return Decimal(0), err
	}
	d, err := valueToDecimal(v)
	if err != nil {
		return Decimal(0), err
	}
	return d, nil
}

func valueToIP(v Value) (IPAddr, error) {
	i, ok := v.(IPAddr)
	if !ok {
		return IPAddr{}, fmt.Errorf("%w: expected ipaddr, got %v", errType, v.typeName())
	}
	return i, nil
}

func evalIP(n evaler, ctx *evalContext) (IPAddr, error) {
	v, err := n.Eval(ctx)
	if err != nil {
		return IPAddr{}, err
	}
	i, err := valueToIP(v)
	if err != nil {
		return IPAddr{}, err
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

func (n *errorEval) Eval(_ *evalContext) (Value, error) {
	return zeroValue(), n.err
}

// literalEval
type literalEval struct {
	value Value
}

func newLiteralEval(value Value) *literalEval {
	return &literalEval{value: value}
}

func (n *literalEval) Eval(_ *evalContext) (Value, error) {
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

func (n *orEval) Eval(ctx *evalContext) (Value, error) {
	v, err := n.lhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	b, err := valueToBool(v)
	if err != nil {
		return zeroValue(), err
	}
	if b {
		return v, nil
	}
	v, err = n.rhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	_, err = valueToBool(v)
	if err != nil {
		return zeroValue(), err
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

func (n *andEval) Eval(ctx *evalContext) (Value, error) {
	v, err := n.lhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	b, err := valueToBool(v)
	if err != nil {
		return zeroValue(), err
	}
	if !b {
		return v, nil
	}
	v, err = n.rhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	_, err = valueToBool(v)
	if err != nil {
		return zeroValue(), err
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

func (n *notEval) Eval(ctx *evalContext) (Value, error) {
	v, err := n.inner.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	b, err := valueToBool(v)
	if err != nil {
		return zeroValue(), err
	}
	return !b, nil
}

// Overflow
// The Go spec specifies that overflow results in defined and deterministic
// behavior (https://go.dev/ref/spec#Integer_overflow), so we can go ahead and
// do the operations and then check for overflow ex post facto.

func checkedAddI64(lhs, rhs Long) (Long, bool) {
	result := lhs + rhs
	if (result > lhs) != (rhs > 0) {
		return result, false
	}
	return result, true
}

func checkedSubI64(lhs, rhs Long) (Long, bool) {
	result := lhs - rhs
	if (result > lhs) != (rhs < 0) {
		return result, false
	}
	return result, true
}

func checkedMulI64(lhs, rhs Long) (Long, bool) {
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

func checkedNegI64(a Long) (Long, bool) {
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

func (n *addEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
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
	lhs evaler
	rhs evaler
}

func newSubtractEval(lhs evaler, rhs evaler) *subtractEval {
	return &subtractEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *subtractEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
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
	lhs evaler
	rhs evaler
}

func newMultiplyEval(lhs evaler, rhs evaler) *multiplyEval {
	return &multiplyEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *multiplyEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
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
	inner evaler
}

func newNegateEval(inner evaler) *negateEval {
	return &negateEval{
		inner: inner,
	}
}

func (n *negateEval) Eval(ctx *evalContext) (Value, error) {
	inner, err := evalLong(n.inner, ctx)
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
	lhs evaler
	rhs evaler
}

func newLongLessThanEval(lhs evaler, rhs evaler) *longLessThanEval {
	return &longLessThanEval{
		lhs: lhs,
		rhs: rhs,
	}
}

func (n *longLessThanEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs < rhs), nil
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

func (n *longLessThanOrEqualEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs <= rhs), nil
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

func (n *longGreaterThanEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs > rhs), nil
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

func (n *longGreaterThanOrEqualEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalLong(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalLong(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs >= rhs), nil
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

func (n *decimalLessThanEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalDecimal(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs < rhs), nil
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

func (n *decimalLessThanOrEqualEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalDecimal(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs <= rhs), nil
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

func (n *decimalGreaterThanEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalDecimal(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs > rhs), nil
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

func (n *decimalGreaterThanOrEqualEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalDecimal(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalDecimal(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs >= rhs), nil
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

func (n *ifThenElseEval) Eval(ctx *evalContext) (Value, error) {
	cond, err := evalBool(n.if_, ctx)
	if err != nil {
		return zeroValue(), err
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

func (n *equalEval) Eval(ctx *evalContext) (Value, error) {
	lv, err := n.lhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	rv, err := n.rhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lv.equal(rv)), nil
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

func (n *notEqualEval) Eval(ctx *evalContext) (Value, error) {
	lv, err := n.lhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	rv, err := n.rhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(!lv.equal(rv)), nil
}

// setLiteralEval
type setLiteralEval struct {
	elements []evaler
}

func newSetLiteralEval(elements []evaler) *setLiteralEval {
	return &setLiteralEval{elements: elements}
}

func (n *setLiteralEval) Eval(ctx *evalContext) (Value, error) {
	var vals Set
	for _, e := range n.elements {
		v, err := e.Eval(ctx)
		if err != nil {
			return zeroValue(), err
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

func (n *containsEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalSet(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := n.rhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(lhs.contains(rhs)), nil
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

func (n *containsAllEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalSet(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalSet(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	result := true
	for _, e := range rhs {
		if !lhs.contains(e) {
			result = false
			break
		}
	}
	return Boolean(result), nil
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

func (n *containsAnyEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalSet(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalSet(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	result := false
	for _, e := range rhs {
		if lhs.contains(e) {
			result = true
			break
		}
	}
	return Boolean(result), nil
}

// recordLiteralEval
type recordLiteralEval struct {
	elements map[string]evaler
}

func newRecordLiteralEval(elements map[string]evaler) *recordLiteralEval {
	return &recordLiteralEval{elements: elements}
}

func (n *recordLiteralEval) Eval(ctx *evalContext) (Value, error) {
	vals := Record{}
	for k, en := range n.elements {
		v, err := en.Eval(ctx)
		if err != nil {
			return zeroValue(), err
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

func (n *attributeAccessEval) Eval(ctx *evalContext) (Value, error) {
	v, err := n.object.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	var record Record
	key := "record"
	switch vv := v.(type) {
	case EntityUID:
		key = "`" + vv.String() + "`"
		var unspecified EntityUID
		if vv == unspecified {
			return zeroValue(), fmt.Errorf("cannot access attribute `%s` of %w", n.attribute, errUnspecifiedEntity)
		}
		rec, ok := ctx.Entities[vv]
		if !ok {
			return zeroValue(), fmt.Errorf("entity `%v` %w", vv.String(), errEntityNotExist)
		} else {
			record = rec.Attributes
		}
	case Record:
		record = vv
	default:
		return zeroValue(), fmt.Errorf("%w: expected one of [record, (entity of type `any_entity_type`)], got %v", errType, v.typeName())
	}
	val, ok := record[n.attribute]
	if !ok {
		return zeroValue(), fmt.Errorf("%s %w `%s`", key, errAttributeAccess, n.attribute)
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

func (n *hasEval) Eval(ctx *evalContext) (Value, error) {
	v, err := n.object.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}
	var record Record
	switch vv := v.(type) {
	case EntityUID:
		rec, ok := ctx.Entities[vv]
		if !ok {
			record = Record{}
		} else {
			record = rec.Attributes
		}
	case Record:
		record = vv
	default:
		return zeroValue(), fmt.Errorf("%w: expected one of [record, (entity of type `any_entity_type`)], got %v", errType, v.typeName())
	}
	_, ok := record[n.attribute]
	return Boolean(ok), nil
}

// likeEval
type likeEval struct {
	lhs     evaler
	pattern parser.Pattern
}

func newLikeEval(lhs evaler, pattern parser.Pattern) *likeEval {
	return &likeEval{lhs: lhs, pattern: pattern}
}

func (l *likeEval) Eval(ctx *evalContext) (Value, error) {
	v, err := evalString(l.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(match(l.pattern, string(v))), nil
}

type variableName func(ctx *evalContext) Value

func variableNamePrincipal(ctx *evalContext) Value { return ctx.Principal }
func variableNameAction(ctx *evalContext) Value    { return ctx.Action }
func variableNameResource(ctx *evalContext) Value  { return ctx.Resource }
func variableNameContext(ctx *evalContext) Value   { return ctx.Context }

// variableEval
type variableEval struct {
	variableName variableName
}

func newVariableEval(variableName variableName) *variableEval {
	return &variableEval{variableName: variableName}
}

func (n *variableEval) Eval(ctx *evalContext) (Value, error) {
	return n.variableName(ctx), nil
}

// inEval
type inEval struct {
	lhs, rhs evaler
}

func newInEval(lhs, rhs evaler) *inEval {
	return &inEval{lhs: lhs, rhs: rhs}
}

func entityIn(entity EntityUID, query map[EntityUID]struct{}, entities Entities) bool {
	checked := map[EntityUID]struct{}{}
	toCheck := []EntityUID{entity}
	for len(toCheck) > 0 {
		var candidate EntityUID
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

func (n *inEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalEntity(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}

	rhs, err := n.rhs.Eval(ctx)
	if err != nil {
		return zeroValue(), err
	}

	query := map[EntityUID]struct{}{}
	switch rhsv := rhs.(type) {
	case EntityUID:
		query[rhsv] = struct{}{}
	case Set:
		for _, rhv := range rhsv {
			e, err := valueToEntity(rhv)
			if err != nil {
				return zeroValue(), err
			}
			query[e] = struct{}{}
		}
	default:
		return zeroValue(), fmt.Errorf(
			"%w: expected one of [set, (entity of type `any_entity_type`)], got %v", errType, rhs.typeName())
	}
	return Boolean(entityIn(lhs, query, ctx.Entities)), nil
}

// isEval
type isEval struct {
	lhs, rhs evaler
}

func newIsEval(lhs, rhs evaler) *isEval {
	return &isEval{lhs: lhs, rhs: rhs}
}

func (n *isEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalEntity(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}

	rhs, err := evalPath(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}

	return Boolean(path(lhs.Type) == rhs), nil
}

// decimalLiteralEval
type decimalLiteralEval struct {
	literal evaler
}

func newDecimalLiteralEval(literal evaler) *decimalLiteralEval {
	return &decimalLiteralEval{literal: literal}
}

func (n *decimalLiteralEval) Eval(ctx *evalContext) (Value, error) {
	literal, err := evalString(n.literal, ctx)
	if err != nil {
		return zeroValue(), err
	}

	d, err := ParseDecimal(string(literal))
	if err != nil {
		return zeroValue(), err
	}

	return d, nil
}

type ipLiteralEval struct {
	literal evaler
}

func newIPLiteralEval(literal evaler) *ipLiteralEval {
	return &ipLiteralEval{literal: literal}
}

func (n *ipLiteralEval) Eval(ctx *evalContext) (Value, error) {
	literal, err := evalString(n.literal, ctx)
	if err != nil {
		return zeroValue(), err
	}

	i, err := ParseIPAddr(string(literal))
	if err != nil {
		return zeroValue(), err
	}

	return i, nil
}

type ipTestType func(v IPAddr) bool

func ipTestIPv4(v IPAddr) bool      { return v.isIPv4() }
func ipTestIPv6(v IPAddr) bool      { return v.isIPv6() }
func ipTestLoopback(v IPAddr) bool  { return v.isLoopback() }
func ipTestMulticast(v IPAddr) bool { return v.isMulticast() }

// ipTestEval
type ipTestEval struct {
	object evaler
	test   ipTestType
}

func newIPTestEval(object evaler, test ipTestType) *ipTestEval {
	return &ipTestEval{object: object, test: test}
}

func (n *ipTestEval) Eval(ctx *evalContext) (Value, error) {
	i, err := evalIP(n.object, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(n.test(i)), nil
}

// ipIsInRangeEval

type ipIsInRangeEval struct {
	lhs, rhs evaler
}

func newIPIsInRangeEval(lhs, rhs evaler) *ipIsInRangeEval {
	return &ipIsInRangeEval{lhs: lhs, rhs: rhs}
}

func (n *ipIsInRangeEval) Eval(ctx *evalContext) (Value, error) {
	lhs, err := evalIP(n.lhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	rhs, err := evalIP(n.rhs, ctx)
	if err != nil {
		return zeroValue(), err
	}
	return Boolean(rhs.contains(lhs)), nil
}
