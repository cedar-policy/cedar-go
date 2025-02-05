package format

import (
	"bytes"
	"fmt"
	"io"

	"github.com/cedar-policy/cedar-go/schema/ast"
	"github.com/cedar-policy/cedar-go/schema/internal/parser"
)

type bailout error

type Options struct {
	// Indent is the string to use for each level of indentation.
	// If empty, 1 tab is used.
	Indent string
}

// Source will pretty-print src in the returned byte slice. If src is malformed Cedar schema, an error will be returned.
func Source(src []byte, opts *Options) ([]byte, error) {
	var buf bytes.Buffer
	tree, err := parser.ParseFile("<input>", src)
	if err != nil {
		return nil, err
	}

	err = Node(tree, &buf, opts)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Node will pretty-print the AST node to out.
//
// The rules for formatting are fixed and not configurable.
func Node(n ast.Node, out io.Writer, opts *Options) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if bail, ok := r.(bailout); ok {
				err = bail
			}
		}
	}()
	if opts == nil {
		opts = &Options{}
	}
	if opts.Indent == "" {
		opts.Indent = "\t" // 4 spaces
	}
	p := &formatter{
		w:        out,
		lastchar: '\n',
		tab:      opts.Indent,
	}
	p.print(n)
	return nil
}

type formatter struct {
	indent int // 1 = 1 tab over
	w      io.Writer

	lastchar byte
	tab      string
}

func (p *formatter) printInd(s string) {
	if p.lastchar != '\n' {
		panic("lastchar must be newline when calling printInd")
	}
	for i := 0; i < p.indent; i++ {
		p.write(p.tab)
	}
	if len(s) > 0 {
		p.write(s)
		p.lastchar = s[len(s)-1]
	}
}

func (p *formatter) printIndf(format string, args ...any) {
	p.printInd(fmt.Sprintf(format, args...))
}

func (p *formatter) write(s string) {
	_, err := io.WriteString(p.w, s)
	if err != nil {
		panic(bailout(err))
	}
	p.lastchar = s[len(s)-1]
}

func (p *formatter) writef(format string, args ...any) {
	buf := fmt.Sprintf(format, args...)
	p.write(buf)
}

func (p *formatter) print(n ast.Node) {
	switch n := n.(type) {
	case *ast.Schema:
		for _, d := range n.Decls {
			p.print(d)
		}
		p.print(n.Remaining)
	case *ast.Namespace:
		p.print(n.Before)
		p.printInd("namespace ")
		p.print(n.Name)
		p.write(" {")
		if n.Inline != nil {
			p.print(n.Inline)
		}
		p.write("\n")
		for _, d := range n.Decls {
			p.indent += 1
			p.print(d)
			p.indent -= 1
		}
		if len(n.Remaining) > 0 {
			p.indent += 1
			p.print(n.Remaining)
			p.indent -= 1
		}
		p.write("}")
		if n.Footer != nil {
			p.print(n.Footer)
		}
		p.write("\n")
	case *ast.CommonTypeDecl:
		p.print(n.Before)
		p.printIndf("type %s = ", n.Name.Value)
		p.print(n.Value)
		p.write(";")
		if n.Footer != nil {
			p.print(n.Footer)
		}
		p.write("\n")
	case *ast.RecordType:
		p.write("{")
		if n.Inner != nil {
			p.print(n.Inner)
		}
		p.write("\n")
		for _, a := range n.Attributes {
			p.indent += 1
			p.print(a)
			p.indent -= 1
		}
		if len(n.Remaining) > 0 {
			p.indent += 1
			p.print(n.Remaining)
			p.indent -= 1
		}
		p.printInd("}")
	case *ast.SetType:
		p.write("Set<")
		p.print(n.Element)
		p.write(">")
	case *ast.Path:
		for i, part := range n.Parts {
			if i > 0 {
				p.write("::")
			}
			p.print(part)
		}
	case *ast.Attribute:
		p.print(n.Before)
		p.printInd("") // print indent
		p.print(n.Key)
		if !n.IsRequired {
			p.write("?")
		}
		p.write(": ")
		p.print(n.Type)
		p.write(",")
		if n.Inline != nil {
			p.print(n.Inline)
		}
		p.write("\n")
	case *ast.Ident:
		p.write(n.Value)
	case *ast.String:
		p.write(n.QuotedVal)
	case *ast.Entity:
		p.print(n.Before)
		p.printInd("entity ")
		for i, name := range n.Names {
			if i > 0 {
				p.write(", ")
			}
			p.print(name)
		}
		if n.In != nil {
			p.write(" in ")
			printBracketList(p, n.In)
		}
		if n.Shape != nil {
			if n.EqTok.Line > 0 {
				p.write(" = ")
			} else {
				p.write(" ")
			}
			p.print(n.Shape)
		}
		if n.Tags != nil {
			p.write(" tags ")
			p.print(n.Tags)
		}
		p.write(";")
		if n.Footer != nil {
			p.print(n.Footer)
		}
		p.write("\n")
	case *ast.Action:
		p.print(n.Before)
		p.printInd("action ")
		for i, name := range n.Names {
			if i > 0 {
				p.write(", ")
			}
			p.print(name)
		}
		if len(n.In) > 0 {
			p.write(" in ")
			printBracketList(p, n.In)
		}
		if n.AppliesTo != nil {
			p.write(" appliesTo {")
			if n.AppliesTo.Inline != nil {
				p.print(n.AppliesTo.Inline)
			}
			p.write("\n")
			p.indent += 1
			p.print(n.AppliesTo)
			p.indent -= 1
			p.printInd("}")
		}
		p.write(";")
		if n.Footer != nil {
			p.print(n.Footer)
		}
		p.write("\n")
	case *ast.AppliesTo:
		if len(n.Principal) > 0 {
			p.print(n.PrincipalComments.Before)
			p.printInd("principal: ")
			printBracketList(p, n.Principal)
			p.write(",")
			if n.PrincipalComments.Inline != nil {
				p.print(n.PrincipalComments.Inline)
			}
			p.write("\n")
		}
		if len(n.Resource) > 0 {
			p.print(n.ResourceComments.Before)
			p.printInd("resource: ")
			printBracketList(p, n.Resource)
			p.write(",")
			if n.ResourceComments.Inline != nil {
				p.print(n.ResourceComments.Inline)
			}
			p.write("\n")
		}
		if n.Context != nil {
			p.print(n.ContextComments.Before)
			p.printInd("context: ")
			p.print(n.Context)
			p.write(",")
			if n.ContextComments.Inline != nil {
				p.print(n.ContextComments.Inline)
			}
			p.write("\n")
		}
		p.print(n.Remaining)
	case *ast.Ref:
		for i, part := range n.Namespace {
			if i > 0 {
				p.write("::")
			}
			p.print(part)
		}
		if len(n.Namespace) > 0 {
			p.write("::")
		}
		p.print(n.Name)
	case ast.CommentBlock:
		if len(n) == 0 {
			return
		}
		for _, c := range n {
			// Print each comment line on a separate line indented
			p.printInd("")
			p.print(c)
			p.write("\n")
		}
	case *ast.Comment:
		if p.lastchar != ' ' && p.lastchar != '\t' && p.lastchar != '\x00' && p.lastchar != '\n' {
			p.write(" ")
		}
		p.writef("// %s", n.Trim())
	default:
		panic(fmt.Sprintf("unhandled node type %T", n))
	}
}

func printBracketList[T ast.Node](p *formatter, list []T) {
	if len(list) == 0 {
		panic("list must not be empty")
	}
	if len(list) > 1 {
		p.write("[")
	}
	for i, item := range list {
		if i > 0 {
			p.write(", ")
		}
		p.print(item)
	}
	if len(list) > 1 {
		p.write("]")
	}
}
