package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func TypeName(v types.Value) string {
	switch t := v.(type) {
	case types.Boolean:
		return "bool"
	case types.Decimal:
		return "decimal"
	case types.Datetime:
		return "datetime"
	case types.EntityUID:
		return fmt.Sprintf("(entity of type `%s`)", t.Type)
	case types.IPAddr:
		return "IP"
	case types.Long:
		return "long"
	case types.Record:
		return "record"
	case types.Set:
		return "set"
	case types.String:
		return "string"
	default:
		return "unknown type"
	}
}

var ErrType = fmt.Errorf("type error")

func ValueToBool(v types.Value) (types.Boolean, error) {
	bv, ok := v.(types.Boolean)
	if !ok {
		return false, fmt.Errorf("%w: expected bool, got %v", ErrType, TypeName(v))
	}
	return bv, nil
}

func ValueToLong(v types.Value) (types.Long, error) {
	lv, ok := v.(types.Long)
	if !ok {
		return 0, fmt.Errorf("%w: expected long, got %v", ErrType, TypeName(v))
	}
	return lv, nil
}

func ValueToString(v types.Value) (types.String, error) {
	sv, ok := v.(types.String)
	if !ok {
		return "", fmt.Errorf("%w: expected string, got %v", ErrType, TypeName(v))
	}
	return sv, nil
}

func ValueToSet(v types.Value) (types.Set, error) {
	sv, ok := v.(types.Set)
	if !ok {
		return types.Set{}, fmt.Errorf("%w: expected set, got %v", ErrType, TypeName(v))
	}
	return sv, nil
}

func ValueToRecord(v types.Value) (types.Record, error) {
	rv, ok := v.(types.Record)
	if !ok {
		return types.Record{}, fmt.Errorf("%w: expected record got %v", ErrType, TypeName(v))
	}
	return rv, nil
}

func ValueToEntity(v types.Value) (types.EntityUID, error) {
	ev, ok := v.(types.EntityUID)
	if !ok {
		return types.EntityUID{}, fmt.Errorf("%w: expected (entity of type `any_entity_type`), got %v", ErrType, TypeName(v))
	}
	return ev, nil
}

func ValueToDatetime(v types.Value) (types.Datetime, error) {
	d, ok := v.(types.Datetime)
	if !ok {
		return types.Datetime{}, fmt.Errorf("%w: expected datetime, got %v", ErrType, TypeName(v))
	}
	return d, nil
}

func ValueToDecimal(v types.Value) (types.Decimal, error) {
	d, ok := v.(types.Decimal)
	if !ok {
		return types.Decimal{}, fmt.Errorf("%w: expected decimal, got %v", ErrType, TypeName(v))
	}
	return d, nil
}

func ValueToDuration(v types.Value) (types.Duration, error) {
	d, ok := v.(types.Duration)
	if !ok {
		return types.Duration{}, fmt.Errorf("%w: expected duration, got %v", ErrType, TypeName(v))
	}
	return d, nil
}

func ValueToIP(v types.Value) (types.IPAddr, error) {
	i, ok := v.(types.IPAddr)
	if !ok {
		return types.IPAddr{}, fmt.Errorf("%w: expected ipaddr, got %v", ErrType, TypeName(v))
	}
	return i, nil
}
