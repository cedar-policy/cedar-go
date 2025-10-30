package ast

import (
	"strings"

	"github.com/cedar-policy/cedar-go/internal/schema/token"
)

// Human readable syntax tree for Cedar schema files instead of JSON.
// The human readable format is defined here: https://docs.cedarpolicy.com/schema/human-readable-schema-grammar.html

// Schema    := {Namespace}
// Namespace := ('namespace' Path '{' {Decl} '}') | Decl
// Decl      := Entity | Action | TypeDecl
// Entity    := 'entity' Idents ['in' EntOrTyps] [['='] RecType] ['tags' Type] ';'
// Action    := 'action' Names ['in' RefOrRefs] [AppliesTo]';'
// TypeDecl  := 'type' TYPENAME '=' Type ';'
// Type      := Path | SetType | RecType
// EntType   := Path
// SetType   := 'Set' '<' Type '>'
// RecType   := '{' [AttrDecls] '}'
// AttrDecls := Name ['?'] ':' Type [',' | ',' AttrDecls]
// AppliesTo := 'appliesTo' '{' AppDecls '}'
// AppDecls  := ('principal' | 'resource') ':' EntOrTyps [',' | ',' AppDecls]
//            | 'context' ':' (Path | RecType) [',' | ',' AppDecls]
// Path      := IDENT {'::' IDENT}
// Ref       := Path '::' STR | Name
// RefOrRefs := Ref | '[' [RefOrRefs] ']'
// EntTypes  := Path {',' Path}
// EntOrTyps := EntType | '[' [EntTypes] ']'
// Name      := IDENT | STR
// Names     := Name {',' Name}
// Idents    := IDENT {',' IDENT}

// IDENT     := ['_''a'-'z''A'-'Z']['_''a'-'z''A'-'Z''0'-'9']*
// TYPENAME  := IDENT - RESERVED
// STR       := Fully-escaped Unicode surrounded by '"'s
// PRIMTYPE  := 'Long' | 'String' | 'Bool'
// WHITESPC  := Unicode whitespace
// COMMENT   := '//' ~NEWLINE* NEWLINE
// RESERVED  := 'Bool' | 'Boolean' | 'Entity' | 'Extension' | 'Long' | 'Record' | 'Set' | 'String'

// The human readable format is not 1-1 convertible with JSON. The JSON format
// is lossy. It loses formatting, such as comments, ordering of fields, etc...
//
// DO NOT ENABLE sumtype:decl YET.
//
// Enabling this causes the following error to be reported:
//
// internal/schema/ast/format.go:69:2: exhaustiveness check failed for sum type "Node" (from internal/schema/ast/ast.go:47:6): missing cases for unknownNode, unknownType (gochecksumtype)
//
//	switch n := n.(type) {
//	^
//
// 1 issues:
// * gochecksumtype: 1
type Node interface {
	isNode()
	// Pos returns first token of the node
	Pos() token.Position
	End() token.Position
}

// no-op statements are included for code coverage instrumentation
func (*Schema) isNode()         { _ = 0 }
func (*Namespace) isNode()      { _ = 0 }
func (*CommonTypeDecl) isNode() { _ = 0 }
func (*RecordType) isNode()     { _ = 0 }
func (*SetType) isNode()        { _ = 0 }
func (*Path) isNode()           { _ = 0 }
func (*Ident) isNode()          { _ = 0 }
func (*Entity) isNode()         { _ = 0 }
func (*Action) isNode()         { _ = 0 }
func (*AppliesTo) isNode()      { _ = 0 }
func (*Ref) isNode()            { _ = 0 }
func (*Attribute) isNode()      { _ = 0 }
func (*String) isNode()         { _ = 0 }
func (CommentBlock) isNode()    { _ = 0 }
func (*Comment) isNode()        { _ = 0 }
func (*Annotation) isNode()     { _ = 0 }

type NodeComments struct {
	Before CommentBlock // comments that precede the node on a separate line
	Inline *Comment     // inline, e.g. namespace name { <After>
	Footer *Comment     // all trailing comments after closing brace
}

type Schema struct {
	Decls []Declaration // either namespace or declarations in global namespace

	Remaining CommentBlock // any comments after all the declarations
}

func (s *Schema) Pos() token.Position {
	if len(s.Decls) > 0 {
		return s.Decls[0].Pos()
	}
	return token.Position{}
}

func (s *Schema) End() token.Position {
	if len(s.Remaining) > 0 {
		return s.Remaining.End()
	}
	if len(s.Decls) > 0 {
		return s.Decls[len(s.Decls)-1].End()
	}
	return token.Position{}
}

//sumtype:decl
type Declaration interface {
	Node
	isDecl()
}

// no-op statements are included for code coverage instrumentation
func (*Entity) isDecl()         { _ = 0 }
func (*Action) isDecl()         { _ = 0 }
func (*Namespace) isDecl()      { _ = 0 }
func (*CommonTypeDecl) isDecl() { _ = 0 }
func (*CommentBlock) isDecl()   { _ = 0 }

type Namespace struct {
	Annotations []*Annotation
	NodeComments
	NamespaceTok token.Position
	Name         *Path
	Decls        []Declaration
	Remaining    CommentBlock
	CloseBrace   token.Position
}

func (n *Namespace) Pos() token.Position {
	if len(n.Annotations) > 0 {
		return n.Annotations[0].Pos()
	}
	if len(n.Before) > 0 {
		return n.Before.Pos()
	}
	return n.NamespaceTok
}

func (n *Namespace) End() token.Position {
	if n.Footer != nil {
		return n.Footer.End()
	}
	return n.CloseBrace
}

type CommonTypeDecl struct {
	Annotations []*Annotation
	NodeComments
	TypeTok token.Position
	Name    *Ident
	Value   Type
}

func (t *CommonTypeDecl) Pos() token.Position {
	if len(t.Annotations) > 0 {
		return t.Annotations[0].Pos()
	}
	if len(t.Before) > 0 {
		return t.Before.Pos()
	}
	return t.TypeTok
}

func (t *CommonTypeDecl) End() token.Position {
	if t.Footer != nil {
		return t.Footer.End()
	}
	return t.Value.End()
}

// TypeValue is either:
// 1. A record type
// 2. A set type (Set<String>)
// 3. A path (Namespace::EntityType or String)
//
// sumtype:decl
type Type interface {
	Node
	isType()
}

// no-op statements are included for code coverage instrumentation
func (*RecordType) isType() { _ = 0 }
func (*SetType) isType()    { _ = 0 }
func (*Path) isType()       { _ = 0 }

type RecordType struct {
	Inner      *Comment // after initial '{'
	LeftCurly  token.Position
	Attributes []*Attribute
	RightCurly token.Position
	Remaining  CommentBlock // any comments after last attribute
}

func (r *RecordType) Pos() token.Position {
	return r.LeftCurly
}

func (r *RecordType) End() token.Position {
	return r.RightCurly
}

type Attribute struct {
	Annotations []*Annotation
	NodeComments
	Key        Name
	IsRequired bool // if true, has ? after name
	Type       Type
	Comma      token.Position
}

func (a *Attribute) Pos() token.Position {
	if a.Annotations != nil {
		return a.Annotations[0].Pos()
	}
	if a.Before != nil {
		return a.NodeComments.Before[0].SlashTok
	}
	return a.Key.Pos()
}

func (a *Attribute) End() token.Position {
	if a.Comma.Line != 0 {
		return a.Comma
	}
	return a.Type.End()
}

type SetType struct {
	SetToken   token.Position
	Element    Type
	RightAngle token.Position
}

func (s *SetType) Pos() token.Position {
	return s.SetToken
}

func (s *SetType) End() token.Position {
	return s.RightAngle
}

type Path struct {
	Parts []*Ident
}

func (p *Path) String() string {
	parts := make([]string, len(p.Parts))
	for i, part := range p.Parts {
		parts[i] = part.Value
	}
	return strings.Join(parts, "::")
}

func (p *Path) Pos() token.Position {
	if len(p.Parts) == 0 {
		return token.Position{}
	}
	return p.Parts[0].IdentTok
}

func (p *Path) End() token.Position {
	if len(p.Parts) == 0 {
		return token.Position{}
	}
	return p.Parts[len(p.Parts)-1].End()
}

type Ident struct {
	IdentTok token.Position
	Value    string
}

func (i *Ident) Pos() token.Position {
	return i.IdentTok
}

func (i *Ident) End() token.Position {
	after := i.IdentTok
	after.Column += len(i.Value)
	after.Offset += len(i.Value)
	return after
}

type Entity struct {
	Annotations []*Annotation
	NodeComments
	EntityTok token.Position
	Names     []*Ident // define multiple entities with the same shape

	// Traditional entity definition
	In    []*Path        // optional, if nil none given
	EqTok token.Position // valid if = is present before shape
	Shape *RecordType    // nil if none given
	Tags  Type

	// Enumerated entity definition
	Enum []*String

	Semicolon token.Position
}

func (e *Entity) Pos() token.Position {
	if len(e.Annotations) > 0 {
		return e.Annotations[0].Pos()
	}
	if len(e.Before) > 0 {
		return e.Before.Pos()
	}
	return e.EntityTok
}

func (e *Entity) End() token.Position {
	if e.Footer != nil {
		return e.Footer.End()
	}
	return e.Semicolon
}

type Action struct {
	Annotations []*Annotation
	NodeComments
	ActionTok token.Position
	Names     []Name
	In        []*Ref     // optional, if nil none given
	AppliesTo *AppliesTo // optional, if nil none given
	Semicolon token.Position
}

func (a *Action) Pos() token.Position {
	if len(a.Annotations) > 0 {
		return a.Annotations[0].Pos()
	}
	if len(a.Before) > 0 {
		return a.Before.Pos()
	}
	return a.ActionTok
}

func (a *Action) End() token.Position {
	if a.Footer != nil {
		return a.Footer.End()
	}
	return a.Semicolon
}

type AppliesTo struct {
	AppliesToTok token.Position
	CloseBrace   token.Position

	Principal     []*Path // one of required
	Resource      []*Path
	ContextPath   *Path       // nil if none
	ContextRecord *RecordType // nil if none

	Inline            *Comment // after {
	PrincipalComments NodeComments
	ResourceComments  NodeComments
	ContextComments   NodeComments
	Remaining         CommentBlock // leftovers after all three fields
}

func (a *AppliesTo) Pos() token.Position {
	return a.AppliesToTok
}

func (a *AppliesTo) End() token.Position {
	return a.CloseBrace
}

// Ref is like a path, but the last element can be a string instead of an ident
type Ref struct {
	Namespace []*Ident // nil if no namespace
	Name      Name
}

func (r *Ref) Pos() token.Position {
	if len(r.Namespace) == 0 {
		return r.Name.Pos()
	}
	return r.Namespace[0].IdentTok
}

func (r *Ref) End() token.Position {
	return r.Name.End()
}

// Name is an IDENT or STR
//
//sumtype:decl
type Name interface {
	Node
	isName()
	String() string
}

func (i *Ident) String() string {
	return i.Value
}

func (s *String) String() string {
	return s.Value()
}

type String struct {
	Tok       token.Position
	QuotedVal string
}

func (s *String) Value() string {
	return s.QuotedVal[1 : len(s.QuotedVal)-1]
}

// no-op statements are included for code coverage instrumentation
func (*String) isName() { _ = 0 }
func (*Ident) isName()  { _ = 0 }

func (s *String) Pos() token.Position {
	return s.Tok
}

func (s *String) End() token.Position {
	after := s.Tok
	after.Offset += len(s.QuotedVal)
	after.Column += len(s.QuotedVal)
	return after
}

type CommentBlock []*Comment

func (c CommentBlock) Pos() token.Position {
	if len(c) == 0 {
		return token.Position{}
	}
	return c[0].SlashTok
}

func (c CommentBlock) End() token.Position {
	if len(c) == 0 {
		return token.Position{}
	}
	return c[len(c)-1].End()
}

type Comment struct {
	SlashTok token.Position // position of '//'
	Value    string         // raw string value
}

func (c *Comment) Pos() token.Position {
	return c.SlashTok
}

func (c *Comment) End() token.Position {
	after := c.SlashTok
	after.Offset += len(c.Value)
	after.Column += len(c.Value)
	return after
}

func (c *Comment) Trim() string {
	return strings.TrimLeft(c.Value, " \t\n/")
}

type Annotation struct {
	NodeComments
	At         token.Position
	Key        *Ident
	LeftParen  token.Position
	Value      *String
	RightParen token.Position
}

func (a *Annotation) Pos() token.Position {
	if len(a.Before) > 0 {
		return a.Before.Pos()
	}
	return a.At
}

func (a *Annotation) End() token.Position {
	if a.Value == nil {
		return a.Key.End()
	}
	return a.RightParen
}
