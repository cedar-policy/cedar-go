package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

var ErrType = fmt.Errorf("type error")

func ValueToBool(v types.Value) (types.Boolean, error) {
	bv, ok := v.(types.Boolean)
	if !ok {
		return false, fmt.Errorf("%w: expected bool, got %v", ErrType, v.TypeName())
	}
	return bv, nil
}

func ValueToLong(v types.Value) (types.Long, error) {
	lv, ok := v.(types.Long)
	if !ok {
		return 0, fmt.Errorf("%w: expected long, got %v", ErrType, v.TypeName())
	}
	return lv, nil
}

func ValueToString(v types.Value) (types.String, error) {
	sv, ok := v.(types.String)
	if !ok {
		return "", fmt.Errorf("%w: expected string, got %v", ErrType, v.TypeName())
	}
	return sv, nil
}

func ValueToSet(v types.Value) (types.Set, error) {
	sv, ok := v.(types.Set)
	if !ok {
		return nil, fmt.Errorf("%w: expected set, got %v", ErrType, v.TypeName())
	}
	return sv, nil
}

func ValueToRecord(v types.Value) (types.Record, error) {
	rv, ok := v.(types.Record)
	if !ok {
		return nil, fmt.Errorf("%w: expected record got %v", ErrType, v.TypeName())
	}
	return rv, nil
}

func ValueToEntity(v types.Value) (types.EntityUID, error) {
	ev, ok := v.(types.EntityUID)
	if !ok {
		return types.EntityUID{}, fmt.Errorf("%w: expected (entity of type `any_entity_type`), got %v", ErrType, v.TypeName())
	}
	return ev, nil
}

func ValueToEntityType(v types.Value) (types.EntityType, error) {
	ev, ok := v.(types.EntityType)
	if !ok {
		return "", fmt.Errorf("%w: expected (EntityType of type `any_entity_type`), got %v", ErrType, v.TypeName())
	}
	return ev, nil
}

func ValueToDecimal(v types.Value) (types.Decimal, error) {
	d, ok := v.(types.Decimal)
	if !ok {
		return 0, fmt.Errorf("%w: expected decimal, got %v", ErrType, v.TypeName())
	}
	return d, nil
}

func ValueToIP(v types.Value) (types.IPAddr, error) {
	i, ok := v.(types.IPAddr)
	if !ok {
		return types.IPAddr{}, fmt.Errorf("%w: expected ipaddr, got %v", ErrType, v.TypeName())
	}
	return i, nil
}
