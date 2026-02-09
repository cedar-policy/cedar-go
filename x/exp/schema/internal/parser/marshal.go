package parser

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	ast2 "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// MarshalSchema formats an AST schema as Cedar text.
func MarshalSchema(schema *ast2.Schema) []byte {
	var buf bytes.Buffer
	m := marshaler{w: &buf}
	m.marshalSchema(schema)
	return buf.Bytes()
}

type marshaler struct {
	w      *bytes.Buffer
	indent int
}

func (m *marshaler) writeIndent() {
	for range m.indent {
		m.w.WriteByte('\t')
	}
}

func (m *marshaler) marshalSchema(schema *ast2.Schema) {
	first := true

	// Marshal bare declarations
	m.marshalDecls(&first, schema.Entities, schema.Enums, schema.Actions, schema.CommonTypes)

	// Marshal namespaces in sorted order
	nsNames := sortedKeys(schema.Namespaces)
	for _, name := range nsNames {
		ns := schema.Namespaces[name]
		if !first {
			m.w.WriteByte('\n')
		}
		first = false
		m.marshalAnnotations(ns.Annotations)
		m.writeIndent()
		fmt.Fprintf(m.w, "namespace %s {\n", name)
		m.indent++
		innerFirst := true
		m.marshalDecls(&innerFirst, ns.Entities, ns.Enums, ns.Actions, ns.CommonTypes)
		m.indent--
		m.writeIndent()
		m.w.WriteString("}\n")
	}
}

func (m *marshaler) marshalDecls(first *bool, entities ast2.Entities, enums ast2.Enums, actions ast2.Actions, commonTypes ast2.CommonTypes) {
	// Type declarations
	typeNames := sortedIdents(commonTypes)
	for _, name := range typeNames {
		ct := commonTypes[name]
		if !*first {
			m.w.WriteByte('\n')
		}
		*first = false
		m.marshalAnnotations(ct.Annotations)
		m.writeIndent()
		fmt.Fprintf(m.w, "type %s = ", name)
		m.marshalType(ct.Type)
		m.w.WriteString(";\n")
	}

	// Entity declarations
	entityNames := sortedEntityTypes(entities)
	for _, name := range entityNames {
		entity := entities[name]
		if !*first {
			m.w.WriteByte('\n')
		}
		*first = false
		m.marshalAnnotations(entity.Annotations)
		m.writeIndent()
		fmt.Fprintf(m.w, "entity %s", unqualifyEntityType(name))
		if len(entity.MemberOf) > 0 {
			m.w.WriteString(" in ")
			m.marshalEntityTypeRefs(entity.MemberOf)
		}
		if entity.Shape != nil {
			m.w.WriteByte(' ')
			m.marshalRecordType(*entity.Shape)
		}
		if entity.Tags != nil {
			m.w.WriteString(" tags ")
			m.marshalType(entity.Tags)
		}
		m.w.WriteString(";\n")
	}

	// Enum entity declarations
	enumNames := sortedEntityTypes(enums)
	for _, name := range enumNames {
		enum := enums[name]
		if !*first {
			m.w.WriteByte('\n')
		}
		*first = false
		m.marshalAnnotations(enum.Annotations)
		m.writeIndent()
		fmt.Fprintf(m.w, "entity %s enum [", unqualifyEntityType(name))
		for i, v := range enum.Values {
			if i > 0 {
				m.w.WriteString(", ")
			}
			m.w.WriteString(quoteCedar(string(v)))
		}
		m.w.WriteString("];\n")
	}

	// Action declarations
	actionNames := sortedStrings(actions)
	for _, name := range actionNames {
		action := actions[name]
		if !*first {
			m.w.WriteByte('\n')
		}
		*first = false
		m.marshalAnnotations(action.Annotations)
		m.writeIndent()
		m.w.WriteString("action ")
		m.marshalActionName(name)
		if len(action.MemberOf) > 0 {
			m.w.WriteString(" in ")
			m.marshalParentRefs(action.MemberOf)
		}
		if action.AppliesTo != nil {
			m.marshalAppliesTo(action.AppliesTo)
		}
		m.w.WriteString(";\n")
	}
}

func (m *marshaler) marshalAnnotations(annotations ast2.Annotations) {
	keys := sortedAnnotationKeys(annotations)
	for _, key := range keys {
		val := annotations[key]
		m.writeIndent()
		if val == "" {
			fmt.Fprintf(m.w, "@%s\n", key)
		} else {
			fmt.Fprintf(m.w, "@%s(%s)\n", key, quoteCedar(string(val)))
		}
	}
}

func (m *marshaler) marshalType(t ast2.IsType) {
	switch t := t.(type) {
	case ast2.StringType:
		m.w.WriteString("String")
	case ast2.LongType:
		m.w.WriteString("Long")
	case ast2.BoolType:
		m.w.WriteString("Bool")
	case ast2.ExtensionType:
		m.w.WriteString(string(t))
	case ast2.SetType:
		m.w.WriteString("Set<")
		m.marshalType(t.Element)
		m.w.WriteByte('>')
	case ast2.RecordType:
		m.marshalRecordType(t)
	case ast2.EntityTypeRef:
		m.w.WriteString(string(t))
	case ast2.TypeRef:
		m.w.WriteString(string(t))
	}
}

func (m *marshaler) marshalRecordType(rec ast2.RecordType) {
	m.w.WriteByte('{')
	keys := sortedRecordKeys(rec)
	if len(keys) > 0 {
		m.w.WriteByte('\n')
		m.indent++
		for i, key := range keys {
			attr := rec[key]
			m.marshalAnnotations(attr.Annotations)
			m.writeIndent()
			m.marshalAttrName(key)
			if attr.Optional {
				m.w.WriteByte('?')
			}
			m.w.WriteString(": ")
			m.marshalType(attr.Type)
			if i < len(keys)-1 {
				m.w.WriteByte(',')
			}
			m.w.WriteByte('\n')
		}
		m.indent--
		m.writeIndent()
	}
	m.w.WriteByte('}')
}

func (m *marshaler) marshalEntityTypeRefs(refs []ast2.EntityTypeRef) {
	if len(refs) == 1 {
		m.w.WriteString(string(refs[0]))
		return
	}
	m.w.WriteByte('[')
	for i, ref := range refs {
		if i > 0 {
			m.w.WriteString(", ")
		}
		m.w.WriteString(string(ref))
	}
	m.w.WriteByte(']')
}

func (m *marshaler) marshalParentRefs(refs []ast2.ParentRef) {
	if len(refs) == 1 {
		m.marshalParentRef(refs[0])
		return
	}
	m.w.WriteByte('[')
	for i, ref := range refs {
		if i > 0 {
			m.w.WriteString(", ")
		}
		m.marshalParentRef(ref)
	}
	m.w.WriteByte(']')
}

func (m *marshaler) marshalParentRef(ref ast2.ParentRef) {
	if types.EntityType(ref.Type) == "" {
		m.marshalActionName(ref.ID)
	} else {
		fmt.Fprintf(m.w, "%s::%s", ref.Type, quoteCedar(string(ref.ID)))
	}
}

func (m *marshaler) marshalAppliesTo(at *ast2.AppliesTo) {
	m.w.WriteString(" appliesTo {\n")
	m.indent++
	hasPrev := false
	if at.Principals != nil {
		m.writeIndent()
		m.w.WriteString("principal: ")
		m.marshalEntityTypeRefs(at.Principals)
		hasPrev = true
	}
	if at.Resources != nil {
		if hasPrev {
			m.w.WriteString(",\n")
		}
		m.writeIndent()
		m.w.WriteString("resource: ")
		m.marshalEntityTypeRefs(at.Resources)
		hasPrev = true
	}
	if at.Context != nil {
		if hasPrev {
			m.w.WriteString(",\n")
		}
		m.writeIndent()
		m.w.WriteString("context: ")
		m.marshalType(at.Context)
		hasPrev = true
	}
	if hasPrev {
		m.w.WriteByte('\n')
	}
	m.indent--
	m.writeIndent()
	m.w.WriteByte('}')
}

func (m *marshaler) marshalActionName(name types.String) {
	s := string(name)
	if isValidIdent(s) {
		m.w.WriteString(s)
	} else {
		m.w.WriteString(quoteCedar(s))
	}
}

func (m *marshaler) marshalAttrName(name types.String) {
	s := string(name)
	if isValidIdent(s) {
		m.w.WriteString(s)
	} else {
		m.w.WriteString(quoteCedar(s))
	}
}

func isValidIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !isIdentStart(r) {
				return false
			}
		} else {
			if !isIdentContinue(r) {
				return false
			}
		}
	}
	return true
}

// unqualifyEntityType extracts the basename from a fully qualified entity type.
// For types within a namespace (e.g., the namespace "NS" contains entity type
// "NS::Foo"), the marshal output only writes "Foo" because the namespace is implied.
func unqualifyEntityType(et types.EntityType) string {
	s := string(et)
	if idx := strings.LastIndex(s, "::"); idx >= 0 {
		return s[idx+2:]
	}
	return s
}

func sortedKeys[V any](m map[types.Path]V) []types.Path {
	keys := make([]types.Path, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedIdents[V any](m map[types.Ident]V) []types.Ident {
	keys := make([]types.Ident, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedEntityTypes[V any](m map[types.EntityType]V) []types.EntityType {
	keys := make([]types.EntityType, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedStrings[V any](m map[types.String]V) []types.String {
	keys := make([]types.String, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedRecordKeys(m ast2.RecordType) []types.String {
	return sortedStrings(map[types.String]ast2.Attribute(m))
}

func sortedAnnotationKeys(m ast2.Annotations) []types.Ident {
	return sortedIdents(map[types.Ident]types.String(m))
}

// quoteCedar produces a double-quoted string literal using only Cedar-valid
// escape sequences: \n, \r, \t, \\, \", \0, and \u{hex} for all other
// non-printable or non-ASCII characters.
func quoteCedar(s string) string {
	var buf strings.Builder
	buf.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '\x00':
			buf.WriteString(`\0`)
		default:
			if r >= 0x20 && r < 0x7f {
				buf.WriteRune(r)
			} else {
				fmt.Fprintf(&buf, `\u{%x}`, r)
			}
		}
	}
	buf.WriteByte('"')
	return buf.String()
}

func init() {
	// Ensure reservedTypeNames is sorted for consistency.
	slices.Sort(reservedTypeNames)
}
