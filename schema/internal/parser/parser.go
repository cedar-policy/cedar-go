// Package parser provides Cedar schema text parsing and formatting.
package parser

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/schema/ast"
	"github.com/cedar-policy/cedar-go/types"
)

// reservedTypeNames are identifiers that cannot be used as common type names.
var reservedTypeNames = []string{
	"Bool", "Boolean", "Entity", "Extension", "Long", "Record", "Set", "String",
}

// ParseSchema parses Cedar schema text into an AST.
func ParseSchema(filename string, src []byte) (*ast.Schema, error) {
	p := &parser{lex: newLexer(filename, src)}
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return p.parseSchema()
}

type parser struct {
	lex *lexer
	tok token
}

func (p *parser) readToken() error {
	tok, err := p.lex.next()
	if err != nil {
		return err
	}
	p.tok = tok
	return nil
}

func (p *parser) errorf(format string, args ...any) error {
	return fmt.Errorf("%s: %s", p.tok.Pos, fmt.Sprintf(format, args...))
}

func (p *parser) expect(tt tokenType) error {
	if p.tok.Type != tt {
		return p.errorf("expected %s, got %s", tokenName(tt), tokenDesc(p.tok))
	}
	return p.readToken()
}

func tokenName(tt tokenType) string {
	switch tt {
	case tokenEOF:
		return "EOF"
	case tokenIdent:
		return "identifier"
	case tokenString:
		return "string"
	case tokenAt:
		return "'@'"
	case tokenLBrace:
		return "'{'"
	case tokenRBrace:
		return "'}'"
	case tokenLBracket:
		return "'['"
	case tokenRBracket:
		return "']'"
	case tokenLAngle:
		return "'<'"
	case tokenRAngle:
		return "'>'"
	case tokenLParen:
		return "'('"
	case tokenRParen:
		return "')'"
	case tokenComma:
		return "','"
	case tokenSemicolon:
		return "';'"
	case tokenColon:
		return "':'"
	case tokenDoubleColon:
		return "'::'"
	case tokenQuestion:
		return "'?'"
	case tokenEquals:
		return "'='"
	default:
		return "unknown"
	}
}

func tokenDesc(tok token) string {
	switch tok.Type {
	case tokenEOF:
		return "EOF"
	case tokenIdent:
		return fmt.Sprintf("identifier %q", tok.Text)
	case tokenString:
		return fmt.Sprintf("string %q", tok.Text)
	default:
		return fmt.Sprintf("%q", tok.Text)
	}
}

func (p *parser) parseSchema() (*ast.Schema, error) {
	schema := &ast.Schema{}
	for p.tok.Type != tokenEOF {
		annotations, err := p.parseAnnotations()
		if err != nil {
			return nil, err
		}
		if p.tok.Type == tokenIdent && p.tok.Text == "namespace" {
			if err := p.readToken(); err != nil {
				return nil, err
			}
			ns, err := p.parseNamespace(annotations)
			if err != nil {
				return nil, err
			}
			if schema.Namespaces == nil {
				schema.Namespaces = ast.Namespaces{}
			}
			schema.Namespaces[ns.name] = ns.ns
		} else {
			if err := p.parseDecl(annotations, nil, schema); err != nil {
				return nil, err
			}
		}
	}
	return schema, nil
}

type parsedNamespace struct {
	name types.Path
	ns   ast.Namespace
}

func (p *parser) parseNamespace(annotations ast.Annotations) (parsedNamespace, error) {
	path, err := p.parsePath()
	if err != nil {
		return parsedNamespace{}, err
	}
	if err := p.expect(tokenLBrace); err != nil {
		return parsedNamespace{}, err
	}
	ns := ast.Namespace{Annotations: annotations}
	var innerSchema ast.Schema
	for p.tok.Type != tokenRBrace {
		if p.tok.Type == tokenEOF {
			return parsedNamespace{}, p.errorf("expected '}' to close namespace, got EOF")
		}
		innerAnnotations, err := p.parseAnnotations()
		if err != nil {
			return parsedNamespace{}, err
		}
		pathPtr := path
		if err := p.parseDecl(innerAnnotations, &pathPtr, &innerSchema); err != nil {
			return parsedNamespace{}, err
		}
	}
	if err := p.readToken(); err != nil { // consume '}'
		return parsedNamespace{}, err
	}
	ns.Entities = innerSchema.Entities
	ns.Enums = innerSchema.Enums
	ns.Actions = innerSchema.Actions
	ns.CommonTypes = innerSchema.CommonTypes
	return parsedNamespace{name: path, ns: ns}, nil
}

func (p *parser) parseDecl(annotations ast.Annotations, namespace *types.Path, schema *ast.Schema) error {
	if p.tok.Type != tokenIdent {
		return p.errorf("expected declaration (entity, action, or type), got %s", tokenDesc(p.tok))
	}
	switch p.tok.Text {
	case "entity":
		if err := p.readToken(); err != nil {
			return err
		}
		return p.parseEntity(annotations, namespace, schema)
	case "action":
		if err := p.readToken(); err != nil {
			return err
		}
		return p.parseAction(annotations, namespace, schema)
	case "type":
		if err := p.readToken(); err != nil {
			return err
		}
		return p.parseTypeDecl(annotations, schema)
	default:
		return p.errorf("expected declaration (entity, action, or type), got identifier %q", p.tok.Text)
	}
}

func (p *parser) parseEntity(annotations ast.Annotations, namespace *types.Path, schema *ast.Schema) error {
	names, err := p.parseIdents()
	if err != nil {
		return err
	}

	// Check for enum entity
	if p.tok.Type == tokenIdent && p.tok.Text == "enum" {
		if err := p.readToken(); err != nil {
			return err
		}
		return p.parseEnumEntity(annotations, namespace, names, schema)
	}

	// Parse optional 'in' clause
	var memberOf []ast.EntityTypeRef
	if p.tok.Type == tokenIdent && p.tok.Text == "in" {
		if err := p.readToken(); err != nil {
			return err
		}
		memberOf, err = p.parseEntityTypes()
		if err != nil {
			return err
		}
	}

	// Parse optional shape (record type), with optional '='
	var shape *ast.RecordType
	switch p.tok.Type {
	case tokenEquals:
		if err := p.readToken(); err != nil {
			return err
		}
		rec, err := p.parseRecordType()
		if err != nil {
			return err
		}
		shape = &rec
	case tokenLBrace:
		rec, err := p.parseRecordType()
		if err != nil {
			return err
		}
		shape = &rec
	}

	// Parse optional tags
	var tags ast.IsType
	if p.tok.Type == tokenIdent && p.tok.Text == "tags" {
		if err := p.readToken(); err != nil {
			return err
		}
		tags, err = p.parseType()
		if err != nil {
			return err
		}
	}

	if err := p.expect(tokenSemicolon); err != nil {
		return err
	}

	if schema.Entities == nil {
		schema.Entities = ast.Entities{}
	}
	for _, name := range names {
		entityType := qualifyEntityType(namespace, name)
		schema.Entities[entityType] = ast.Entity{
			Annotations: annotations,
			MemberOf:    memberOf,
			Shape:       shape,
			Tags:        tags,
		}
	}
	return nil
}

func (p *parser) parseEnumEntity(annotations ast.Annotations, namespace *types.Path, names []types.Ident, schema *ast.Schema) error {
	if err := p.expect(tokenLBracket); err != nil {
		return err
	}
	var values []types.String
	for p.tok.Type != tokenRBracket {
		if p.tok.Type != tokenString {
			return p.errorf("expected string literal in enum, got %s", tokenDesc(p.tok))
		}
		values = append(values, types.String(p.tok.Text))
		if err := p.readToken(); err != nil {
			return err
		}
		if p.tok.Type == tokenComma {
			if err := p.readToken(); err != nil {
				return err
			}
		} else if p.tok.Type != tokenRBracket {
			return p.errorf("expected ',' or ']' in enum, got %s", tokenDesc(p.tok))
		}
	}
	if err := p.readToken(); err != nil { // consume ']'
		return err
	}
	if err := p.expect(tokenSemicolon); err != nil {
		return err
	}

	if schema.Enums == nil {
		schema.Enums = ast.Enums{}
	}
	for _, name := range names {
		entityType := qualifyEntityType(namespace, name)
		schema.Enums[entityType] = ast.Enum{
			Annotations: annotations,
			Values:      values,
		}
	}
	return nil
}

func (p *parser) parseAction(annotations ast.Annotations, namespace *types.Path, schema *ast.Schema) error {
	names, err := p.parseNames()
	if err != nil {
		return err
	}

	// Parse optional 'in' clause
	var memberOf []ast.ParentRef
	if p.tok.Type == tokenIdent && p.tok.Text == "in" {
		if err := p.readToken(); err != nil {
			return err
		}
		memberOf, err = p.parseActionParents(namespace)
		if err != nil {
			return err
		}
	}

	// Parse optional appliesTo clause
	var appliesTo *ast.AppliesTo
	if p.tok.Type == tokenIdent && p.tok.Text == "appliesTo" {
		if err := p.readToken(); err != nil {
			return err
		}
		at, err := p.parseAppliesTo()
		if err != nil {
			return err
		}
		appliesTo = at
	}

	// Allow optional 'attributes {}' (Rust compat, deprecated)
	if p.tok.Type == tokenIdent && p.tok.Text == "attributes" {
		if err := p.readToken(); err != nil {
			return err
		}
		if err := p.expect(tokenLBrace); err != nil {
			return err
		}
		if err := p.expect(tokenRBrace); err != nil {
			return err
		}
	}

	if err := p.expect(tokenSemicolon); err != nil {
		return err
	}

	if schema.Actions == nil {
		schema.Actions = ast.Actions{}
	}
	for _, name := range names {
		schema.Actions[name] = ast.Action{
			Annotations: annotations,
			MemberOf:    memberOf,
			AppliesTo:   appliesTo,
		}
	}
	return nil
}

func (p *parser) parseTypeDecl(annotations ast.Annotations, schema *ast.Schema) error {
	if p.tok.Type != tokenIdent {
		return p.errorf("expected type name, got %s", tokenDesc(p.tok))
	}
	name := p.tok.Text
	if slices.Contains(reservedTypeNames, name) {
		return p.errorf("%q is a reserved type name", name)
	}
	if err := p.readToken(); err != nil {
		return err
	}
	if err := p.expect(tokenEquals); err != nil {
		return err
	}
	typ, err := p.parseType()
	if err != nil {
		return err
	}
	if err := p.expect(tokenSemicolon); err != nil {
		return err
	}

	if schema.CommonTypes == nil {
		schema.CommonTypes = ast.CommonTypes{}
	}
	schema.CommonTypes[types.Ident(name)] = ast.CommonType{
		Annotations: annotations,
		Type:        typ,
	}
	return nil
}

func (p *parser) parseAnnotations() (ast.Annotations, error) {
	var annotations ast.Annotations
	for p.tok.Type == tokenAt {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		if p.tok.Type != tokenIdent {
			return nil, p.errorf("expected annotation name, got %s", tokenDesc(p.tok))
		}
		key := types.Ident(p.tok.Text)
		if err := p.readToken(); err != nil {
			return nil, err
		}
		var value types.String
		hasValue := false
		if p.tok.Type == tokenLParen {
			if err := p.readToken(); err != nil {
				return nil, err
			}
			if p.tok.Type != tokenString {
				return nil, p.errorf("expected annotation value string, got %s", tokenDesc(p.tok))
			}
			value = types.String(p.tok.Text)
			hasValue = true
			if err := p.readToken(); err != nil {
				return nil, err
			}
			if err := p.expect(tokenRParen); err != nil {
				return nil, err
			}
		}
		if annotations == nil {
			annotations = ast.Annotations{}
		}
		if hasValue {
			annotations[key] = value
		} else {
			annotations[key] = ""
		}
	}
	return annotations, nil
}

// parsePath parses IDENT { '::' IDENT }
func (p *parser) parsePath() (types.Path, error) {
	if p.tok.Type != tokenIdent {
		return "", p.errorf("expected identifier, got %s", tokenDesc(p.tok))
	}
	path := p.tok.Text
	if err := p.readToken(); err != nil {
		return "", err
	}
	for p.tok.Type == tokenDoubleColon {
		if err := p.readToken(); err != nil {
			return "", err
		}
		if p.tok.Type != tokenIdent {
			return "", p.errorf("expected identifier after '::', got %s", tokenDesc(p.tok))
		}
		path += "::" + p.tok.Text
		if err := p.readToken(); err != nil {
			return "", err
		}
	}
	return types.Path(path), nil
}

// parsePathForRef parses a path that may include a trailing '::' followed by a string literal
// for action parent references. Returns the path and whether a string was found.
func (p *parser) parsePathForRef() (path types.Path, str types.String, qualified bool, err error) {
	if p.tok.Type != tokenIdent {
		return "", "", false, p.errorf("expected identifier, got %s", tokenDesc(p.tok))
	}
	pathStr := p.tok.Text
	if err := p.readToken(); err != nil {
		return "", "", false, err
	}
	for p.tok.Type == tokenDoubleColon {
		if err := p.readToken(); err != nil {
			return "", "", false, err
		}
		if p.tok.Type == tokenString {
			str := types.String(p.tok.Text)
			if err := p.readToken(); err != nil {
				return "", "", false, err
			}
			return types.Path(pathStr), str, true, nil
		}
		if p.tok.Type != tokenIdent {
			return "", "", false, p.errorf("expected identifier or string after '::', got %s", tokenDesc(p.tok))
		}
		pathStr += "::" + p.tok.Text
		if err := p.readToken(); err != nil {
			return "", "", false, err
		}
	}
	return types.Path(pathStr), "", false, nil
}

// parseIdents parses IDENT { ',' IDENT }
func (p *parser) parseIdents() ([]types.Ident, error) {
	if p.tok.Type != tokenIdent {
		return nil, p.errorf("expected identifier, got %s", tokenDesc(p.tok))
	}
	var result []types.Ident
	result = append(result, types.Ident(p.tok.Text))
	if err := p.readToken(); err != nil {
		return nil, err
	}
	for p.tok.Type == tokenComma {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		if p.tok.Type != tokenIdent {
			return nil, p.errorf("expected identifier after ',', got %s", tokenDesc(p.tok))
		}
		result = append(result, types.Ident(p.tok.Text))
		if err := p.readToken(); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// parseNames parses Name { ',' Name } where Name = IDENT | STR
func (p *parser) parseNames() ([]types.String, error) {
	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	result := []types.String{name}
	for p.tok.Type == tokenComma {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		name, err = p.parseName()
		if err != nil {
			return nil, err
		}
		result = append(result, name)
	}
	return result, nil
}

func (p *parser) parseName() (types.String, error) {
	switch p.tok.Type {
	case tokenIdent:
		name := types.String(p.tok.Text)
		if err := p.readToken(); err != nil {
			return "", err
		}
		return name, nil
	case tokenString:
		name := types.String(p.tok.Text)
		if err := p.readToken(); err != nil {
			return "", err
		}
		return name, nil
	default:
		return "", p.errorf("expected name (identifier or string), got %s", tokenDesc(p.tok))
	}
}

// parseEntityTypes parses Path | '[' [ Path { ',' Path } ] ']'
func (p *parser) parseEntityTypes() ([]ast.EntityTypeRef, error) {
	if p.tok.Type == tokenLBracket {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		var result []ast.EntityTypeRef
		for p.tok.Type != tokenRBracket {
			path, err := p.parsePath()
			if err != nil {
				return nil, err
			}
			result = append(result, ast.EntityTypeRef(types.EntityType(path)))
			if p.tok.Type == tokenComma {
				if err := p.readToken(); err != nil {
					return nil, err
				}
			} else if p.tok.Type != tokenRBracket {
				return nil, p.errorf("expected ',' or ']', got %s", tokenDesc(p.tok))
			}
		}
		return result, p.readToken() // consume ']'
	}
	path, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	return []ast.EntityTypeRef{ast.EntityTypeRef(types.EntityType(path))}, nil
}

// parseActionParents parses QualName | '[' QualName { ',' QualName } ']'
func (p *parser) parseActionParents(namespace *types.Path) ([]ast.ParentRef, error) {
	if p.tok.Type == tokenLBracket {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		var result []ast.ParentRef
		for p.tok.Type != tokenRBracket {
			ref, err := p.parseQualName(namespace)
			if err != nil {
				return nil, err
			}
			result = append(result, ref)
			if p.tok.Type == tokenComma {
				if err := p.readToken(); err != nil {
					return nil, err
				}
			} else if p.tok.Type != tokenRBracket {
				return nil, p.errorf("expected ',' or ']', got %s", tokenDesc(p.tok))
			}
		}
		return result, p.readToken() // consume ']'
	}
	ref, err := p.parseQualName(namespace)
	if err != nil {
		return nil, err
	}
	return []ast.ParentRef{ref}, nil
}

// parseQualName parses QualName = Name | Path '::' STR
func (p *parser) parseQualName(namespace *types.Path) (ast.ParentRef, error) {
	if p.tok.Type == tokenString {
		name := types.String(p.tok.Text)
		if err := p.readToken(); err != nil {
			return ast.ParentRef{}, err
		}
		return ast.ParentRefFromID(name), nil
	}
	path, str, qualified, err := p.parsePathForRef()
	if err != nil {
		return ast.ParentRef{}, err
	}
	if qualified {
		return ast.NewParentRef(types.EntityType(path), str), nil
	}
	// Bare identifier: treat as an action ID
	return ast.ParentRefFromID(types.String(path)), nil
}

// parseAppliesTo parses '{' AppDecls '}'
func (p *parser) parseAppliesTo() (*ast.AppliesTo, error) {
	if err := p.expect(tokenLBrace); err != nil {
		return nil, err
	}
	at := &ast.AppliesTo{}
	for p.tok.Type != tokenRBrace {
		if p.tok.Type == tokenEOF {
			return nil, p.errorf("expected '}' to close appliesTo, got EOF")
		}
		if p.tok.Type != tokenIdent {
			return nil, p.errorf("expected 'principal', 'resource', or 'context', got %s", tokenDesc(p.tok))
		}
		switch p.tok.Text {
		case "principal":
			if err := p.readToken(); err != nil {
				return nil, err
			}
			if err := p.expect(tokenColon); err != nil {
				return nil, err
			}
			refs, err := p.parseEntityTypes()
			if err != nil {
				return nil, err
			}
			at.Principals = refs
		case "resource":
			if err := p.readToken(); err != nil {
				return nil, err
			}
			if err := p.expect(tokenColon); err != nil {
				return nil, err
			}
			refs, err := p.parseEntityTypes()
			if err != nil {
				return nil, err
			}
			at.Resources = refs
		case "context":
			if err := p.readToken(); err != nil {
				return nil, err
			}
			if err := p.expect(tokenColon); err != nil {
				return nil, err
			}
			ctx, err := p.parseType()
			if err != nil {
				return nil, err
			}
			at.Context = ctx
		default:
			return nil, p.errorf("expected 'principal', 'resource', or 'context', got %q", p.tok.Text)
		}
		if p.tok.Type == tokenComma {
			if err := p.readToken(); err != nil {
				return nil, err
			}
		}
	}
	return at, p.readToken() // consume '}'
}

// parseType parses Path | 'Set' '<' Type '>' | '{' AttrDecls '}'
func (p *parser) parseType() (ast.IsType, error) {
	if p.tok.Type == tokenLBrace {
		rec, err := p.parseRecordType()
		if err != nil {
			return nil, err
		}
		return rec, nil
	}

	if p.tok.Type == tokenIdent && p.tok.Text == "Set" {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		if err := p.expect(tokenLAngle); err != nil {
			return nil, err
		}
		elem, err := p.parseType()
		if err != nil {
			return nil, err
		}
		if err := p.expect(tokenRAngle); err != nil {
			return nil, err
		}
		return ast.Set(elem), nil
	}

	path, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	return ast.TypeRef(path), nil
}

// parseRecordType parses '{' [ AttrDecls ] '}'
func (p *parser) parseRecordType() (ast.RecordType, error) {
	if err := p.expect(tokenLBrace); err != nil {
		return nil, err
	}
	rec := ast.RecordType{}
	for p.tok.Type != tokenRBrace {
		if p.tok.Type == tokenEOF {
			return nil, p.errorf("expected '}' to close record type, got EOF")
		}
		attrAnnotations, err := p.parseAnnotations()
		if err != nil {
			return nil, err
		}
		name, err := p.parseName()
		if err != nil {
			return nil, err
		}
		optional := false
		if p.tok.Type == tokenQuestion {
			optional = true
			if err := p.readToken(); err != nil {
				return nil, err
			}
		}
		if err := p.expect(tokenColon); err != nil {
			return nil, err
		}
		typ, err := p.parseType()
		if err != nil {
			return nil, err
		}
		rec[name] = ast.Attribute{
			Type:        typ,
			Optional:    optional,
			Annotations: attrAnnotations,
		}
		if p.tok.Type == tokenComma {
			if err := p.readToken(); err != nil {
				return nil, err
			}
		}
	}
	return rec, p.readToken() // consume '}'
}

func qualifyEntityType(namespace *types.Path, name types.Ident) types.EntityType {
	if namespace != nil {
		return types.EntityType(string(*namespace) + "::" + string(name))
	}
	return types.EntityType(name)
}
