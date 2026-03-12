package validate

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// entityDeserError marks an error as a deserialization error (structural type mismatch)
// vs a conformance error (semantic type mismatch within the same structural category).
type entityDeserError struct{ msg string }

func (e *entityDeserError) Error() string { return e.msg }

func newDeserError(msg string) error { return &entityDeserError{msg: msg} }

// checkValue validates that a runtime value matches the expected schema type.
func checkValue(v types.Value, expected resolved.IsType) error {
	switch expected := expected.(type) {
	case resolved.StringType:
		if _, ok := v.(types.String); !ok {
			return fmt.Errorf("expected String, got %T", v)
		}
	case resolved.LongType:
		if _, ok := v.(types.Long); !ok {
			return fmt.Errorf("expected Long, got %T", v)
		}
	case resolved.BoolType:
		if _, ok := v.(types.Boolean); !ok {
			return fmt.Errorf("expected Boolean, got %T", v)
		}
	case resolved.EntityType:
		uid, ok := v.(types.EntityUID)
		if !ok {
			return newDeserError(fmt.Sprintf("expected EntityUID, got %T", v))
		}
		if uid.Type != types.EntityType(expected) {
			return fmt.Errorf("expected entity type %q, got %q", expected, uid.Type)
		}
	case resolved.SetType:
		set, ok := v.(types.Set)
		if !ok {
			return newDeserError(fmt.Sprintf("expected Set, got %T", v))
		}
		for elem := range set.All() {
			if err := checkValue(elem, expected.Element); err != nil {
				return fmt.Errorf("set element: %w", err)
			}
		}
	case resolved.RecordType:
		rec, ok := v.(types.Record)
		if !ok {
			return newDeserError(fmt.Sprintf("expected Record, got %T", v))
		}
		return checkRecord(rec, expected)
	case resolved.ExtensionType:
		return checkExtensionValue(v, expected)
	}
	return nil
}

// checkRecord validates a record against a record schema type.
func checkRecord(rec types.Record, expected resolved.RecordType) error {
	// Check all required attributes are present
	for name, attr := range expected {
		v, ok := rec.Get(name)
		if !ok {
			if !attr.Optional {
				return fmt.Errorf("missing required attribute %q", name)
			}
			continue
		}
		if err := checkValue(v, attr.Type); err != nil {
			return fmt.Errorf("attribute %q: %w", name, err)
		}
	}

	// Check for unexpected attributes (closed record)
	for k := range rec.All() {
		if _, ok := expected[k]; !ok {
			return newDeserError(fmt.Sprintf("unexpected attribute %q", k))
		}
	}
	return nil
}

// checkExtensionValue checks that a value matches an extension type.
func checkExtensionValue(v types.Value, expected resolved.ExtensionType) error {
	switch types.Ident(expected) {
	case "ipaddr":
		if _, ok := v.(types.IPAddr); !ok {
			return extensionMismatchError("IPAddr", v)
		}
	case "decimal":
		if _, ok := v.(types.Decimal); !ok {
			return extensionMismatchError("Decimal", v)
		}
	case "datetime":
		if _, ok := v.(types.Datetime); !ok {
			return extensionMismatchError("Datetime", v)
		}
	case "duration":
		if _, ok := v.(types.Duration); !ok {
			return extensionMismatchError("Duration", v)
		}
	}
	return nil
}

// extensionMismatchError returns the appropriate error type based on whether the
// actual value is an extension type (conformance error) or not (deserialization error).
func extensionMismatchError(expected string, v types.Value) error {
	msg := fmt.Sprintf("expected %s, got %T", expected, v)
	// If the actual value is also an extension type, it's a conformance error
	// (the JSON had the right __extn shape, just wrong fn).
	switch v.(type) {
	case types.IPAddr, types.Decimal, types.Datetime, types.Duration:
		return fmt.Errorf("%s", msg)
	}
	// Otherwise it's a deserialization error (JSON was wrong structural type).
	return newDeserError(msg)
}
