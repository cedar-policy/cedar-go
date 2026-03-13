package exptypes

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// unmarshalValueJSON unmarshals JSON data into a types.Value using the
// resolved schema type to disambiguate (e.g. EntityUID vs Record).
func unmarshalValueJSON(data []byte, typ resolved.IsType) (types.Value, error) {
	switch typ := typ.(type) {
	case resolved.StringType:
		return unmarshalString(data)
	case resolved.LongType:
		return unmarshalLong(data)
	case resolved.BoolType:
		return unmarshalBool(data)
	case resolved.EntityType:
		return unmarshalEntityUID(data)
	case resolved.SetType:
		return unmarshalSet(data, typ)
	case resolved.RecordType:
		return unmarshalRecord(data, typ)
	case resolved.ExtensionType:
		return unmarshalExtension(data, typ)
	default:
		return nil, fmt.Errorf("unsupported schema type %T", typ)
	}
}

func unmarshalString(data []byte) (types.Value, error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("expected String: %w", err)
	}
	return types.String(s), nil
}

func unmarshalLong(data []byte) (types.Value, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var n json.Number
	if err := dec.Decode(&n); err != nil {
		return nil, fmt.Errorf("expected Long: %w", err)
	}
	i, err := n.Int64()
	if err != nil {
		return nil, fmt.Errorf("expected Long: %w", err)
	}
	return types.Long(i), nil
}

func unmarshalBool(data []byte) (types.Value, error) {
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("expected Boolean: %w", err)
	}
	return types.Boolean(b), nil
}

// unmarshalEntityUID parses an EntityUID from JSON. Both implicit
// ({"type":"T","id":"I"}) and explicit ({"__entity":{...}}) forms are
// accepted. The entity type is NOT validated against the schema here;
// that check is deferred to the validate package.
func unmarshalEntityUID(data []byte) (types.Value, error) {
	var uid types.EntityUID
	if err := json.Unmarshal(data, &uid); err != nil {
		return nil, fmt.Errorf("expected EntityUID: %w", err)
	}
	return uid, nil
}

func unmarshalExtension(data []byte, typ resolved.ExtensionType) (types.Value, error) {
	// Each extension type's UnmarshalJSON already handles both the
	// {"__extn":{"fn":"...","arg":"..."}} form and bare string form.
	switch types.Ident(typ) {
	case "ipaddr":
		var v types.IPAddr
		if err := v.UnmarshalJSON(data); err != nil {
			return nil, fmt.Errorf("expected IPAddr: %w", err)
		}
		return v, nil
	case "decimal":
		var v types.Decimal
		if err := v.UnmarshalJSON(data); err != nil {
			return nil, fmt.Errorf("expected Decimal: %w", err)
		}
		return v, nil
	case "datetime":
		var v types.Datetime
		if err := v.UnmarshalJSON(data); err != nil {
			return nil, fmt.Errorf("expected Datetime: %w", err)
		}
		return v, nil
	case "duration":
		var v types.Duration
		if err := v.UnmarshalJSON(data); err != nil {
			return nil, fmt.Errorf("expected Duration: %w", err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unknown extension type %q", typ)
	}
}

func unmarshalSet(data []byte, typ resolved.SetType) (types.Value, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("expected Set: %w", err)
	}
	if len(raw) == 0 {
		return types.NewSet(), nil
	}
	elems := make([]types.Value, 0, len(raw))
	for i, r := range raw {
		v, err := unmarshalValueJSON(r, typ.Element)
		if err != nil {
			return nil, fmt.Errorf("set element [%d]: %w", i, err)
		}
		elems = append(elems, v)
	}
	return types.NewSet(elems...), nil
}

func unmarshalRecord(data []byte, typ resolved.RecordType) (types.Value, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("expected Record: %w", err)
	}

	// Reject unknown attributes (closed record)
	for k := range raw {
		if _, ok := typ[types.String(k)]; !ok {
			return nil, fmt.Errorf("unexpected attribute %q", k)
		}
	}

	m := make(types.RecordMap, len(raw))
	for name, attr := range typ {
		rawVal, ok := raw[string(name)]
		if !ok {
			if !attr.Optional {
				return nil, fmt.Errorf("missing required attribute %q", name)
			}
			continue
		}
		v, err := unmarshalValueJSON(rawVal, attr.Type)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		m[name] = v
	}
	return types.NewRecord(m), nil
}
