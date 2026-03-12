package validate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
)

// Request validates a request against the schema.
func (v *Validator) Request(req types.Request) error {
	// Look up action
	action, ok := v.schema.Actions[req.Action]
	if !ok {
		return fmt.Errorf("action `%s` does not exist in the supplied schema", req.Action)
	}

	// Validate principal type
	if err := v.validateRequestEntityType(req.Principal, "principal"); err != nil {
		return err
	}
	if action.AppliesTo == nil || !slices.Contains(action.AppliesTo.Principals, req.Principal.Type) {
		return fmt.Errorf("principal type `%s` is not valid for `%s`", req.Principal.Type, req.Action)
	}

	// Validate resource type
	if err := v.validateRequestEntityType(req.Resource, "resource"); err != nil {
		return err
	}
	if !slices.Contains(action.AppliesTo.Resources, req.Resource.Type) {
		return fmt.Errorf("resource type `%s` is not valid for `%s`", req.Resource.Type, req.Action)
	}

	// Validate context
	if err := checkRecord(req.Context, action.AppliesTo.Context); err != nil {
		return fmt.Errorf("context `%s` is not valid for `%s`", formatContextRecord(req.Context), req.Action)
	}

	return nil
}

func (v *Validator) validateRequestEntityType(uid types.EntityUID, role string) error {
	if v.isKnownEntityType(uid.Type) {
		return nil
	}
	return fmt.Errorf("%s type `%s` is not declared in the schema", role, uid.Type)
}

// formatContextRecord formats a record in Rust Cedar's display format for error messages.
// Format: {key: value, key: value} with unquoted keys and Cedar value display.
func formatContextRecord(rec types.Record) string {
	var sb strings.Builder
	sb.WriteRune('{')
	first := true
	keys := make([]string, 0)
	for k := range rec.All() {
		keys = append(keys, string(k))
	}
	slices.Sort(keys)
	for _, k := range keys {
		val, _ := rec.Get(types.String(k))
		if !first {
			sb.WriteString(", ")
		}
		first = false
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(formatCedarValue(val))
	}
	sb.WriteRune('}')
	return sb.String()
}

// formatCedarValue formats a Cedar value in Rust Cedar's display format.
func formatCedarValue(v types.Value) string {
	switch val := v.(type) {
	case types.Long:
		return fmt.Sprintf("%d", val)
	case types.String:
		return fmt.Sprintf("%q", string(val))
	case types.Boolean:
		if val {
			return "true"
		}
		return "false"
	case types.EntityUID:
		return val.String()
	}
	return fmt.Sprintf("%v", v)
}
