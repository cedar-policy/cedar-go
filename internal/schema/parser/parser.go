// Package schema/parser defines the parser for Cedar human-readable schema files.
//
// The grammar is defined here: https://docs.cedarpolicy.com/schema/human-readable-schema-grammar.html
package parser

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/token"
)

const maxErrors = 10

var (
	ErrBailout = errors.New("too many errors")
)

// ParseFile parses src bytes as human-readable Cedar schema and returns the AST. Parsing the human readable format
// preserves comments, ordering, etc...
//
// Parsing does not validate or ensure the schema is perfectly valid. For example, if the schema refers
// to entities that don't exist, the parsing will still be successful. Filename given will accompany all
// errors returned.
//
// To see the list of errors, you can use errors.Unwrap, for example:
//
//	_, err := ParseFile(...)
//	var errs []error
//	errors.As(err, &errs)
func ParseFile(filename string, src []byte) (schema *ast.Schema, err error) {
	lex := NewLexer(filename, src)
	p := &Parser{lex: lex}
	defer func() {
		errs := lex.Errors
		errs = append(errs, p.Errors...)
		if r := recover(); r != nil {
			if r != ErrBailout {
				panic(r)
			}
		}
		if len(errs) > 0 {
			errs.Sort()
			err = errs
		}
	}()
	schema = p.parseSchema()
	return
}

type Parser struct {
	lex     *Lexer
	nextTok *Token // next token to be consumed, or nil if none consumed yet

	Errors token.Errors
}

func (p *Parser) error(pos token.Position, err error) {
	n := len(p.Errors)
	if n > 0 && p.Errors[n-1].(token.Error).Pos.Line == pos.Line {
		return // discard - likely a spurious error
	}
	if len(p.Errors) > maxErrors {
		panic(ErrBailout)
	}
	p.Errors = append(p.Errors, token.Error{Pos: pos, Err: err})
}

func (p *Parser) errorf(pos token.Position, format string, args ...any) {
	p.error(pos, fmt.Errorf(format, args...))
}

func (p *Parser) advance(to map[token.Type]bool) (tok Token) {
	for p.peek().Type != token.EOF && !to[p.peek().Type] {
		tok = p.eat()
	}
	p.eat() // eat past the token we stopped at
	return
}

func (p *Parser) peek() (tok Token) {
	if p.nextTok != nil {
		return *p.nextTok
	}
	tok = p.eat()
	p.nextTok = &tok
	return tok
}

func (p *Parser) eat() (tok Token) {
	if p.nextTok != nil {
		tok = *p.nextTok
		p.nextTok = nil
	} else {
		tok = p.lex.NextToken()
	}
	return
}

func (p *Parser) eatOnly(tokenType token.Type, errfmt string, args ...any) (Token, bool) {
	tok := p.eat()
	if tok.Type != tokenType {
		errmsg := fmt.Sprintf(errfmt, args...)
		p.error(tok.Pos, fmt.Errorf("%s, got %s", errmsg, tok.String()))
		return tok, false
	}
	return tok, true
}

func (p *Parser) matches(want ...token.Type) bool {
	return slices.Contains(want, p.peek().Type)
}

func (p *Parser) parseSchema() *ast.Schema {
	schema := new(ast.Schema)
	var comments []*ast.Comment
	for p.peek().Type != token.EOF {
		t := p.peek()
		switch t.Type {
		case token.NAMESPACE:
			namespace := p.parseNamespace()
			if len(comments) > 0 {
				namespace.NodeComments.Before = comments
				comments = nil
			}
			schema.Decls = append(schema.Decls, namespace)
		case token.TYPE:
			typ := p.parseTypeDecl()
			if len(comments) > 0 {
				typ.NodeComments.Before = comments
				comments = nil
			}
			schema.Decls = append(schema.Decls, typ)
		case token.ACTION:
			action := p.parseAction()
			if len(comments) > 0 {
				action.NodeComments.Before = comments
				comments = nil
			}
			schema.Decls = append(schema.Decls, action)
		case token.ENTITY:
			entity := p.parseEntityDecl()
			if len(comments) > 0 {
				entity.NodeComments.Before = comments
				comments = nil
			}
			schema.Decls = append(schema.Decls, entity)
		case token.COMMENT:
			comments = append(comments, p.parseComment())
		default:
			p.error(t.Pos, fmt.Errorf("unexpected token %s", t.String()))
			p.advance(map[token.Type]bool{token.SEMICOLON: true})
		}
	}
	if len(comments) > 0 {
		schema.Remaining = comments
	}
	return schema
}

func (p *Parser) parseNamespace() (namespace *ast.Namespace) {
	namespace = new(ast.Namespace)
	nptok, _ := p.eatOnly(token.NAMESPACE, "expected namespace keyword")
	namespace.NamespaceTok = nptok.Pos
	namespace.Name = p.parsePath()
	p.eatOnly(token.LEFTBRACE, "expected { after namespace path")
	if p.peek().Type == token.COMMENT && p.peek().Pos.Line == nptok.Pos.Line {
		namespace.NodeComments.Inline = p.parseComment()
	}
	var comments []*ast.Comment
	for !p.matches(token.RIGHTBRACE, token.EOF) {
		if p.matches(token.ENTITY) {
			entity := p.parseEntityDecl()
			if len(comments) > 0 {
				entity.NodeComments.Before = comments
				comments = nil
			}
			namespace.Decls = append(namespace.Decls, entity)
		} else if p.matches(token.ACTION) {
			action := p.parseAction()
			if len(comments) > 0 {
				action.NodeComments.Before = comments
				comments = nil
			}
			namespace.Decls = append(namespace.Decls, action)
		} else if p.matches(token.TYPE) {
			typ := p.parseTypeDecl()
			if len(comments) > 0 {
				typ.NodeComments.Before = comments
				comments = nil
			}
			namespace.Decls = append(namespace.Decls, typ)
		} else if p.matches(token.COMMENT) {
			comments = append(comments, p.parseComment())
		} else {
			p.errorf(p.peek().Pos, "unexpected token %s, expected action, entity, or type", p.peek().Type)
			p.advance(map[token.Type]bool{token.SEMICOLON: true})
		}
	}
	if len(comments) > 0 {
		namespace.Remaining = comments
	}

	closebrace, _ := p.eatOnly(token.RIGHTBRACE, "expected }")
	namespace.CloseBrace = closebrace.Pos
	if p.matches(token.COMMENT) && p.peek().Pos.Line == closebrace.Pos.Line {
		namespace.Footer = p.parseComment()
	}
	return namespace
}

var reserved = map[string]struct{}{
	"Bool":      {},
	"Boolean":   {},
	"Long":      {},
	"String":    {},
	"Set":       {},
	"Entity":    {},
	"Extension": {},
	"Record":    {},
}

func (p *Parser) parseAction() (action *ast.Action) {
	action = new(ast.Action)
	actionTok, _ := p.eatOnly(token.ACTION, "expected action keyword")
	action.ActionTok = actionTok.Pos
	action.Names = append(action.Names, p.parseName())
	for p.matches(token.COMMA) {
		p.eat()
		action.Names = append(action.Names, p.parseName())
	}
	if p.matches(token.IN) {
		p.eat()
		action.In = p.parseRefOrTypes()
	}
	if p.matches(token.APPLIES_TO) {
		appliesToTok := p.eat()
		action.AppliesTo = p.parseAppliesTo()
		action.AppliesTo.AppliesToTok = appliesToTok.Pos
	}
	semi, _ := p.eatOnly(token.SEMICOLON, "expected ;")
	if p.matches(token.COMMENT) && p.peek().Pos.Line == semi.Pos.Line {
		action.Footer = p.parseComment()
	}
	action.Semicolon = semi.Pos
	return action
}

func (p *Parser) parseAppliesTo() (appliesTo *ast.AppliesTo) {
	appliesTo = new(ast.AppliesTo)
	lbrace, _ := p.eatOnly(token.LEFTBRACE, "expected {")
	if p.matches(token.COMMENT) && p.peek().Pos.Line == lbrace.Pos.Line {
		appliesTo.Inline = p.parseComment()
	}
	var comments []*ast.Comment
loop:
	for !p.matches(token.RIGHTBRACE, token.EOF) {
		n := p.peek()
		var nodeComments *ast.NodeComments
		var node ast.Node
		switch n.Type {
		case token.PRINCIPAL:
			p.eat()
			p.eatOnly(token.COLON, "expected :")
			appliesTo.Principal = p.parseEntOrTypes()
			nodeComments = &appliesTo.PrincipalComments
			node = appliesTo.Principal[len(appliesTo.Principal)-1]
		case token.RESOURCE:
			p.eat()
			p.eatOnly(token.COLON, "expected :")
			appliesTo.Resource = p.parseEntOrTypes()
			nodeComments = &appliesTo.ResourceComments
			node = appliesTo.Resource[len(appliesTo.Resource)-1]
		case token.CONTEXT:
			p.eat()
			p.eatOnly(token.COLON, "expected :")
			appliesTo.Context = p.parseRecType()
			nodeComments = &appliesTo.ContextComments
			node = appliesTo.Context
		case token.COMMENT:
			comments = append(comments, p.parseComment())
			continue
		default:
			p.errorf(p.peek().Pos, "expected principal, resource, or context")
			p.advance(map[token.Type]bool{token.RIGHTBRACE: true})
			break loop
		}

		if nodeComments != nil && len(comments) > 0 {
			nodeComments.Before = comments
			comments = nil
		}

		var comma token.Type
		if p.matches(token.COMMA) {
			comma = p.eat().Type
		}

		if p.matches(token.COMMENT) {
			if node != nil {
				if p.peek().Pos.Line == node.End().Line {
					nodeComments.Inline = p.parseComment()
				}
			}
		}

		for p.matches(token.COMMENT) { // parse the rest if they exist
			comments = append(comments, p.parseComment())
		}

		if !p.matches(token.RIGHTBRACE) && comma != token.COMMA {
			// We're missing a comma and the appliesTo block isn't closed
			p.errorf(p.peek().Pos, "expected , or }")
			p.advance(map[token.Type]bool{token.ACTION: true, token.ENTITY: true, token.TYPE: true})
			break
		}
	}
	if len(comments) > 0 {
		appliesTo.Remaining = comments
	}
	closer, _ := p.eatOnly(token.RIGHTBRACE, "expected }")
	appliesTo.CloseBrace = closer.Pos
	return appliesTo
}

func (p *Parser) parseEntityDecl() (entity *ast.Entity) {
	entity = new(ast.Entity)
	tok, _ := p.eatOnly(token.ENTITY, "expected entity keyword")
	entity.EntityTok = tok.Pos
	entity.Names = append(entity.Names, p.parseIdent())
	for p.matches(token.COMMA) {
		p.eat()
		entity.Names = append(entity.Names, p.parseIdent())
	}

	if p.matches(token.IN) {
		p.eat()
		entity.In = p.parseEntOrTypes()
	}
	if p.matches(token.EQUALS) {
		entity.EqTok = p.eat().Pos
	}
	if p.matches(token.LEFTBRACE) {
		entity.Shape = p.parseRecType()
	}
	if p.matches(token.TAGS) {
		p.eat()
		entity.Tags = p.parseType()
	}
	semi, _ := p.eatOnly(token.SEMICOLON, "expected ;")
	entity.Semicolon = semi.Pos
	if p.matches(token.COMMENT) && p.peek().Pos.Line == semi.Pos.Line {
		entity.NodeComments.Footer = p.parseComment()
	}
	return entity
}

func (p *Parser) parseEntOrTypes() (types []*ast.Path) {
	if p.matches(token.LEFTBRACKET) {
		p.eat()
		for !p.matches(token.RIGHTBRACKET, token.EOF) {
			types = append(types, p.parsePath())
			if p.matches(token.COMMA) {
				p.eat()
			} else if !p.matches(token.RIGHTBRACKET) {
				p.errorf(p.peek().Pos, "expected , or ]")
				p.advance(map[token.Type]bool{token.RIGHTBRACKET: true})
				break
			}
		}
		p.eatOnly(token.RIGHTBRACKET, "expected ]")
	} else {
		types = append(types, p.parsePath())
	}
	return types
}

func (p *Parser) parseRefOrTypes() (types []*ast.Ref) {
	if p.matches(token.LEFTBRACKET) {
		p.eat()
		for !p.matches(token.RIGHTBRACKET, token.EOF) {
			types = append(types, p.parseRef())
			if p.matches(token.COMMA) {
				p.eat()
			} else if !p.matches(token.RIGHTBRACKET) {
				p.errorf(p.peek().Pos, "expected , or ]")
				p.advance(map[token.Type]bool{token.RIGHTBRACKET: true})
				break
			}
		}
		p.eatOnly(token.RIGHTBRACKET, "expected ]")
	} else {
		types = append(types, p.parseRef())
	}
	return types
}

func (p *Parser) parseTypeDecl() (typ *ast.CommonTypeDecl) {
	typ = new(ast.CommonTypeDecl)
	typeTok, _ := p.eatOnly(token.TYPE, "expected type keyword")
	typ.TypeTok = typeTok.Pos
	typ.Name = p.parseIdent()
	if _, ok := reserved[typ.Name.Value]; ok {
		p.errorf(p.peek().Pos, "reserved typename %s", typ.Name.Value)
		p.advance(map[token.Type]bool{token.SEMICOLON: true})
		return typ
	}
	p.eatOnly(token.EQUALS, "expected = after typename")
	typ.Value = p.parseType()
	semi, _ := p.eatOnly(token.SEMICOLON, "expected ;")
	if p.matches(token.COMMENT) && p.peek().Pos.Line == semi.Pos.Line {
		typ.NodeComments.Footer = p.parseComment()
	}
	return typ
}

func (p *Parser) parseType() (typ ast.Type) {
	if p.matches(token.LEFTBRACE) {
		typ = p.parseRecType()
	} else if p.matches(token.IDENT) {
		if p.peek().Lit != "Set" {
			typ = p.parsePath()
		} else {
			setTok := p.eat()
			p.eatOnly(token.LEFTANGLE, "expected < after Set")
			element := p.parseType()
			rangle, _ := p.eatOnly(token.RIGHTANGLE, "expected >")
			typ = &ast.SetType{SetToken: setTok.Pos, Element: element, RightAngle: rangle.Pos}
		}
	} else {
		p.errorf(p.peek().Pos, "expected type, got %s", p.peek().String())
		p.advance(map[token.Type]bool{token.SEMICOLON: true})
	}
	return typ
}

func (p *Parser) parseRecType() (typ *ast.RecordType) {
	typ = new(ast.RecordType)
	lbrace, _ := p.eatOnly(token.LEFTBRACE, "expected {")
	typ.LeftCurly = lbrace.Pos
	if p.matches(token.COMMENT) && p.peek().Pos.Line == lbrace.Pos.Line {
		typ.Inner = p.parseComment()
	}
	var comments []*ast.Comment
	for !p.matches(token.RIGHTBRACE, token.EOF) {
		if p.matches(token.COMMENT) {
			comments = append(comments, p.parseComment())
			continue
		}
		attr := p.parseAttrDecl()
		if len(comments) > 0 {
			attr.NodeComments.Before = comments
			comments = nil
		}
		typ.Attributes = append(typ.Attributes, attr)
		if p.matches(token.COMMA) {
			attr.Comma = p.eat().Pos
		} else if !p.matches(token.RIGHTBRACE) {
			p.errorf(p.peek().Pos, "expected , or }")
			p.advance(map[token.Type]bool{token.RIGHTBRACE: true})
			break
		}
		if p.matches(token.COMMENT) && p.peek().Pos.Line == attr.End().Line {
			typ.Attributes[len(typ.Attributes)-1].Inline = p.parseComment()
		}
	}
	if len(comments) > 0 {
		typ.Remaining = comments
	}
	rbrace, _ := p.eatOnly(token.RIGHTBRACE, "expected }")
	typ.RightCurly = rbrace.Pos
	return typ
}

func (p *Parser) parseAttrDecl() (attr *ast.Attribute) {
	attr = new(ast.Attribute)
	attr.Key = p.parseName()
	if p.matches(token.QUESTION) {
		p.eat()
		attr.IsRequired = false
	} else {
		attr.IsRequired = true
	}
	p.eatOnly(token.COLON, "expected :")
	attr.Type = p.parseType()
	return attr
}

func (p *Parser) parsePath() *ast.Path {
	result := new(ast.Path)
	ident, ok := p.eatOnly(token.IDENT, "expected identifier for start of path")
	if !ok {
		return result
	}
	result.Parts = append(result.Parts, &ast.Ident{Value: ident.Lit, IdentTok: ident.Pos})
	for p.matches(token.DOUBLECOLON) {
		p.eat()
		ident, ok := p.eatOnly(token.IDENT, "expected identifier after ::")
		if !ok {
			continue
		}
		result.Parts = append(result.Parts, &ast.Ident{Value: ident.Lit, IdentTok: ident.Pos})
	}
	return result
}

func (p *Parser) parseRef() *ast.Ref {
	result := new(ast.Ref)
	first := p.parseName()
	if s, ok := first.(*ast.String); ok {
		result.Name = s
		return result
	} else {
		result.Namespace = append(result.Namespace, first.(*ast.Ident))
	}
	for p.matches(token.DOUBLECOLON) {
		p.eat()
		next := p.parseName()
		if s, ok := next.(*ast.String); ok {
			result.Name = s
			return result
		}
		result.Namespace = append(result.Namespace, next.(*ast.Ident))
	}
	if len(result.Namespace) > 0 {
		result.Name = result.Namespace[len(result.Namespace)-1]
		result.Namespace = result.Namespace[:len(result.Namespace)-1]
	}
	return result
}

func (p *Parser) parseName() ast.Name {
	if p.matches(token.STRING) {
		str := p.eat()
		return &ast.String{QuotedVal: str.Lit, Tok: str.Pos}
	} else if p.matches(token.IDENT) {
		ident := p.eat()
		return &ast.Ident{Value: ident.Lit, IdentTok: ident.Pos}
	} else {
		got := p.eat()
		p.errorf(got.Pos, "expected name (identifier or string)")
		return &ast.Ident{Value: got.Lit, IdentTok: got.Pos}
	}
}

func (p *Parser) parseIdent() *ast.Ident {
	ident, _ := p.eatOnly(token.IDENT, "expected identifier")
	return &ast.Ident{Value: ident.Lit, IdentTok: ident.Pos}
}

func (p *Parser) parseComment() *ast.Comment {
	tok, _ := p.eatOnly(token.COMMENT, "expected comment")
	return &ast.Comment{SlashTok: tok.Pos, Value: tok.Lit}
}
