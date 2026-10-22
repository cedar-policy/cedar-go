package types

import (
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

var (
	errJSONInvalidExtn     = fmt.Errorf("invalid extension")
	errJSONDecode          = fmt.Errorf("error decoding json")
	errJSONLongOutOfRange  = fmt.Errorf("long out of range")
	errJSONUnsupportedType = fmt.Errorf("unsupported type")
	errJSONExtFnMatch      = fmt.Errorf("json extn mismatch")
	errJSONExtNotFound     = fmt.Errorf("json extn not found")
	errJSONEntityNotFound  = fmt.Errorf("json entity not found")
	errJSONReservedKey     = fmt.Errorf("record uses reserved key")
)

type extn struct {
	Fn  string `json:"fn"`
	Arg string `json:"arg"`
}

type extValueJSON struct {
	Extn *extn `json:"__extn,omitempty"`
}

type extEntity struct {
	Type *string `json:"type"`
	ID   *string `json:"id"`
}

type entityValueJSON struct {
	Type   *string    `json:"type,omitempty"`
	ID     *string    `json:"id,omitempty"`
	Entity *extEntity `json:"__entity,omitempty"`
}

type explicitValue struct {
	Value Value
}

func (v *explicitValue) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	switch kind := dec.PeekKind(); kind {
	case jsontext.KindString, jsontext.KindFalse, jsontext.KindTrue, jsontext.KindNumber:
		tok, err := dec.ReadToken()
		if err != nil {
			return fmt.Errorf("%w: %w", errJSONDecode, err)
		}
		switch tok.Kind() {
		case jsontext.KindString:
			v.Value = String(tok.String())
		case jsontext.KindFalse, jsontext.KindTrue:
			v.Value = Boolean(tok.Bool())
		case jsontext.KindNumber:
			// Match Rust behavior for -0
			if tok.String() == "-0" {
				return fmt.Errorf("%w: -0 is not a valid Long", errJSONLongOutOfRange)
			}
			l, err := tok.Int()
			if err != nil {
				return fmt.Errorf("%w: %w", errJSONLongOutOfRange, err)
			}
			v.Value = Long(l)
		}
	case jsontext.KindBeginArray:
		var ev []explicitValue
		if err := json.UnmarshalDecode(dec, &ev); err != nil {
			return err
		}
		vals := make([]Value, len(ev))
		for i := range ev {
			vals[i] = ev[i].Value
		}
		v.Value = NewSet(vals...)
	case jsontext.KindBeginObject:
		var m map[string]explicitValue
		if err := json.UnmarshalDecode(dec, &m); err != nil {
			return err
		}
		// Check to see if the Record is _actually_ an escaped entity UID or
		// extension function. Records that don't precisely fit the shape of
		// the escaped types are kept as Records. This matches Rust's behavior.
		if len(m) == 1 {
			if entity, ok := m["__entity"]; ok {
				if uid, ok := coerceValueToEntityUID(entity.Value); ok {
					v.Value = uid
					return nil
				}
			} else if extn, ok := m["__extn"]; ok {
				val, ok, err := coerceValueToExtensionFunction(extn.Value)
				if err != nil {
					return err
				}
				if ok {
					v.Value = val
					return nil
				}
			}
		}
		recMap := make(RecordMap, len(m))
		for k, val := range m {
			recMap[String(k)] = val.Value
		}
		v.Value = NewRecord(recMap)
	default:
		return fmt.Errorf("%w: %v", errJSONUnsupportedType, kind)
	}
	return nil
}

// UnmarshalJSON parses a single JSON value into a Cedar Value without the aid
// of a schema, following the same rules as Rust Cedar's schemaless parsing:
//
//   - A JSON string becomes a String
//   - A JSON boolean becomes a Boolean
//   - A JSON number shaped like a 64-bit signed integer becomes a Long. Other
//     numbers return an error.
//   - A JSON array becomes a Set; duplicate elements are removed.
//   - A JSON object with exactly one member, "__entity", whose value is an
//     object with String "type" and "id" members, becomes an EntityUID.
//   - A JSON object with exactly one member, "__extn", whose value is an
//     object with a String "fn" member and an "arg" member, becomes the
//     extension value produced by calling that function on "arg", which must
//     be a String. Unknown functions and invalid arguments are errors.
//   - Any other JSON object becomes a Record. In particular, the implicit
//     entity form {"type": ..., "id": ...} is a Record, since only a schema
//     can tell it apart from an ordinary record.
//   - JSON null is an error.
//
// To match Rust Cedar's behavior, extra members inside "__entity" and
// "__extn" objects are ignored.
//
// Unlike Rust Cedar, UnmarshalJSON doesn't currently validate entity type
// names.
func UnmarshalJSON(b []byte, v *Value) error {
	var val explicitValue
	if err := json.Unmarshal(b, &val); err != nil {
		return fmt.Errorf("%w: %w", errJSONDecode, err)
	}
	*v = val.Value
	return nil
}

func coerceValueToEntityUID(v Value) (EntityUID, bool) {
	rec, ok := v.(Record)
	if !ok {
		return EntityUID{}, false
	}
	typ, ok := rec.Get("type")
	if !ok {
		return EntityUID{}, false
	}
	id, ok := rec.Get("id")
	if !ok {
		return EntityUID{}, false
	}
	typStr, ok := typ.(String)
	if !ok {
		return EntityUID{}, false
	}
	idStr, ok := id.(String)
	if !ok {
		return EntityUID{}, false
	}
	return NewEntityUID(EntityType(typStr), idStr), true
}

func coerceValueToExtensionFunction(v Value) (Value, bool, error) {
	// First, check to see if this Value is a Record of the right shape.
	rec, ok := v.(Record)
	if !ok {
		return nil, false, nil
	}
	fnVal, ok := rec.Get("fn")
	if !ok {
		return nil, false, nil
	}
	arg, ok := rec.Get("arg")
	if !ok {
		return nil, false, nil
	}
	fn, ok := fnVal.(String)
	if !ok {
		return nil, false, nil
	}

	// At this point, we treat the Record as an extension function and validate
	// its values
	parsers := map[string]func(string) (Value, error){
		"ip":       func(s string) (Value, error) { return ParseIPAddr(s) },
		"decimal":  func(s string) (Value, error) { return ParseDecimal(s) },
		"datetime": func(s string) (Value, error) { return ParseDatetime(s) },
		"duration": func(s string) (Value, error) { return ParseDuration(s) },
	}
	parse, ok := parsers[string(fn)]
	if !ok {
		return nil, false, errJSONInvalidExtn
	}

	argStr, ok := arg.(String)
	if !ok {
		return nil, false, fmt.Errorf("%w: argument to %s must be a string", errJSONInvalidExtn, fn)
	}
	val, err := parse(string(argStr))
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

func unmarshalExtensionValue[T any](b []byte, extName string, parse func(string) (T, error)) (T, error) {
	var zeroT T
	var arg string
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &arg); err != nil {
			return zeroT, errors.Join(errJSONDecode, err)
		}
	} else {
		var res extValueJSON
		if err := json.Unmarshal(b, &res); err != nil {
			return zeroT, errors.Join(errJSONDecode, err)
		}
		if res.Extn == nil {
			// If we didn't find an Extn, maybe it's just an extn.
			var res2 extn

			if err := json.Unmarshal(b, &res2); err != nil {
				return zeroT, errors.Join(errJSONDecode, err)
			}

			// We've tried Ext.Fn and Fn, so no good.
			if res2.Fn == "" {
				return zeroT, errJSONExtNotFound
			}
			if res2.Fn != extName {
				return zeroT, errJSONExtFnMatch
			}
			arg = res2.Arg
		} else if res.Extn.Fn != extName {
			return zeroT, errJSONExtFnMatch
		} else {
			arg = res.Extn.Arg
		}
	}

	v, err := parse(arg)
	if err != nil {
		return zeroT, err
	}

	return v, nil
}
