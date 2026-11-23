// Command reexport generates a file that re-exports the API of another
// package: an alias for each exported type, a constant for each exported
// constant, and a wrapper for each exported function, each with the original's
// doc comment.
//
// It's meant to be run from a go:generate directive in the package that does
// the re-exporting:
//
//	//go:generate go run ./internal/cmd/reexport -from <import path> -out <file> [-exclude <name>,...]
//
// Everything that the package at -from exports is re-exported, except the
// names listed in -exclude. Exported variables aren't supported, since a copy
// wouldn't follow changes to the original, so they must be excluded.
//
// Generation fails if a re-exported function or generic type refers to a type
// that isn't re-exported, since callers couldn't name it.
package main

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("reexport: ")

	from := flag.String("from", "", "import path of the package to re-export")
	out := flag.String("out", "", "file to write")
	exclude := flag.String("exclude", "", "comma-separated names not to re-export")
	flag.Parse()

	pkgName := os.Getenv("GOPACKAGE")
	if *from == "" || *out == "" || pkgName == "" {
		log.Fatal("-from, -out, and $GOPACKAGE (set by go generate) are required")
	}
	var excluded []string
	if *exclude != "" {
		excluded = strings.Split(*exclude, ",")
	}

	src, err := generate(*from, pkgName, excluded)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, src, 0o644); err != nil {
		log.Fatal(err)
	}
}

type generator struct {
	fset *token.FileSet
	pkg  *build.Package
	doc  *doc.Package

	// topLevel holds every top-level name declared in pkg, exported or not.
	topLevel map[string]bool
	// reexported holds the names that are re-exported.
	reexported map[string]bool
	// fileImports maps each file in pkg to the imports it refers to by name.
	fileImports map[string]map[string]string
	// imports holds the imports, by name, that the generated code needs.
	imports map[string]string

	buf bytes.Buffer
}

func generate(from, pkgName string, excluded []string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	pkg, err := build.Import(from, wd, 0)
	if err != nil {
		return nil, err
	}

	g := &generator{
		fset:        token.NewFileSet(),
		pkg:         pkg,
		topLevel:    map[string]bool{},
		reexported:  map[string]bool{},
		fileImports: map[string]map[string]string{},
		imports:     map[string]string{pkg.Name: pkg.ImportPath},
	}

	var files []*ast.File
	for _, name := range pkg.GoFiles {
		f, err := parser.ParseFile(g.fset, filepath.Join(pkg.Dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
		g.recordFile(f)
	}
	g.doc, err = doc.NewFromFiles(g.fset, files, pkg.ImportPath, doc.PreserveAST)
	if err != nil {
		return nil, err
	}

	types := slices.Clone(g.doc.Types)
	consts := slices.Clone(g.doc.Consts)
	vars := slices.Clone(g.doc.Vars)
	funcs := slices.Clone(g.doc.Funcs)
	for _, t := range g.doc.Types {
		consts = append(consts, t.Consts...)
		vars = append(vars, t.Vars...)
		funcs = append(funcs, t.Funcs...)
	}

	exported := map[string]bool{}
	for _, t := range types {
		exported[t.Name] = true
	}
	for _, v := range slices.Concat(consts, vars) {
		for _, n := range v.Names {
			exported[n] = true
		}
	}
	for _, f := range funcs {
		exported[f.Name] = true
	}

	isExcluded := map[string]bool{}
	for _, n := range excluded {
		if !exported[n] {
			return nil, fmt.Errorf("-exclude %s: %s doesn't export %s", n, pkg.ImportPath, n)
		}
		isExcluded[n] = true
	}
	for n := range exported {
		g.reexported[n] = !isExcluded[n]
	}
	for _, v := range vars {
		for _, n := range v.Names {
			if g.reexported[n] {
				return nil, fmt.Errorf("can't re-export variable %s; exclude it", n)
			}
		}
	}

	types = slices.DeleteFunc(types, func(t *doc.Type) bool { return !g.reexported[t.Name] })
	funcs = slices.DeleteFunc(funcs, func(f *doc.Func) bool { return !g.reexported[f.Name] })
	slices.SortFunc(types, func(a, b *doc.Type) int { return cmp.Compare(a.Name, b.Name) })
	slices.SortFunc(consts, func(a, b *doc.Value) int { return cmp.Compare(a.Names[0], b.Names[0]) })
	slices.SortFunc(funcs, func(a, b *doc.Func) int { return cmp.Compare(a.Name, b.Name) })

	for _, t := range types {
		if err := g.writeType(t); err != nil {
			return nil, err
		}
	}
	for _, c := range consts {
		g.writeConsts(c)
	}
	for _, f := range funcs {
		if err := g.writeFunc(f); err != nil {
			return nil, err
		}
	}

	return g.finish(pkgName, from)
}

// recordFile records the top-level names declared in f and the names by which
// f refers to its imports.
func (g *generator) recordFile(f *ast.File) {
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {
				g.topLevel[decl.Name.Name] = true
			}
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					g.topLevel[spec.Name.Name] = true
				case *ast.ValueSpec:
					for _, n := range spec.Names {
						g.topLevel[n.Name] = true
					}
				}
			}
		}
	}

	imports := map[string]string{}
	for _, spec := range f.Imports {
		p := strings.Trim(spec.Path.Value, `"`)
		name := defaultImportName(p)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		imports[name] = p
	}
	g.fileImports[g.fset.File(f.Pos()).Name()] = imports
}

var majorVersion = regexp.MustCompile(`^v[0-9]+$`)

// defaultImportName returns the name that an import of p is conventionally
// referred to by when the import doesn't name it.
func defaultImportName(p string) string {
	name := path.Base(p)
	if majorVersion.MatchString(name) && path.Dir(p) != "." {
		name = path.Base(path.Dir(p))
	}
	return name
}

func (g *generator) writeDoc(text string) {
	if text == "" {
		return
	}
	// Comment returns the text without comment markers.
	text = string(g.doc.Printer().Comment(g.doc.Parser().Parse(text)))
	for line := range strings.Lines(text) {
		if line == "\n" {
			g.buf.WriteString("//\n")
		} else {
			g.buf.WriteString("// " + line)
		}
	}
}

func (g *generator) writeType(t *doc.Type) error {
	var spec *ast.TypeSpec
	for _, s := range t.Decl.Specs {
		if s := s.(*ast.TypeSpec); s.Name.Name == t.Name {
			spec = s
		}
	}

	g.buf.WriteString("\n")
	g.writeDoc(t.Doc)
	if spec.TypeParams == nil {
		fmt.Fprintf(&g.buf, "type %s = %s.%s\n", t.Name, g.pkg.Name, t.Name)
		return nil
	}

	params := typeParamNames(spec.TypeParams)
	if err := g.checkFields(t.Name, spec.TypeParams, params); err != nil {
		return err
	}
	fmt.Fprintf(&g.buf, "type %s[%s] = %s.%s[%s]\n",
		t.Name, g.fieldList(spec.TypeParams), g.pkg.Name, t.Name, strings.Join(params, ", "))
	return nil
}

func (g *generator) writeConsts(v *doc.Value) {
	names := slices.DeleteFunc(slices.Clone(v.Names), func(n string) bool { return !g.reexported[n] })
	switch len(names) {
	case 0:
		return
	case 1:
		g.buf.WriteString("\n")
		g.writeDoc(v.Doc)
		fmt.Fprintf(&g.buf, "const %s = %s.%s\n", names[0], g.pkg.Name, names[0])
	default:
		g.buf.WriteString("\n")
		g.writeDoc(v.Doc)
		g.buf.WriteString("const (\n")
		for _, n := range names {
			fmt.Fprintf(&g.buf, "\t%s = %s.%s\n", n, g.pkg.Name, n)
		}
		g.buf.WriteString(")\n")
	}
}

func (g *generator) writeFunc(f *doc.Func) error {
	typ := f.Decl.Type
	typeParams := typeParamNames(typ.TypeParams)
	for _, fields := range []*ast.FieldList{typ.TypeParams, typ.Params, typ.Results} {
		if err := g.checkFields(f.Name, fields, typeParams); err != nil {
			return err
		}
	}

	// Name every parameter, so the wrapper can pass it along, avoiding the
	// name of the package being called.
	params := &ast.FieldList{}
	var args []string
	if typ.Params != nil {
		for _, field := range typ.Params.List {
			names := field.Names
			if len(names) == 0 {
				names = []*ast.Ident{{Name: "_"}}
			}
			var renamed []*ast.Ident
			for _, n := range names {
				name := n.Name
				if name == "_" || name == g.pkg.Name {
					name = fmt.Sprintf("p%d", len(args))
				}
				renamed = append(renamed, ast.NewIdent(name))
				if _, ok := field.Type.(*ast.Ellipsis); ok {
					name += "..."
				}
				args = append(args, name)
			}
			params.List = append(params.List, &ast.Field{Names: renamed, Type: field.Type})
		}
	}

	sig := &ast.FuncDecl{
		Name: ast.NewIdent(f.Name),
		Type: &ast.FuncType{TypeParams: typ.TypeParams, Params: params, Results: typ.Results},
	}
	callee := g.pkg.Name + "." + f.Name
	if len(typeParams) > 0 {
		callee += "[" + strings.Join(typeParams, ", ") + "]"
	}
	ret := ""
	if typ.Results != nil {
		ret = "return "
	}

	g.buf.WriteString("\n")
	g.writeDoc(f.Doc)
	if err := printer.Fprint(&g.buf, g.fset, sig); err != nil {
		return err
	}
	fmt.Fprintf(&g.buf, " {\n\t%s%s(%s)\n}\n", ret, callee, strings.Join(args, ", "))
	return nil
}

func typeParamNames(fields *ast.FieldList) []string {
	var names []string
	if fields != nil {
		for _, field := range fields.List {
			for _, n := range field.Names {
				names = append(names, n.Name)
			}
		}
	}
	return names
}

// checkFields checks that the types in fields, which belong to the declaration
// of name, refer only to re-exported names, type parameters, predeclared
// identifiers, and imported packages. It records the imports they need.
func (g *generator) checkFields(name string, fields *ast.FieldList, typeParams []string) error {
	if fields == nil {
		return nil
	}
	var err error
	for _, field := range fields.List {
		ast.Inspect(field.Type, func(n ast.Node) bool {
			if err != nil {
				return false
			}
			switch n := n.(type) {
			case *ast.SelectorExpr:
				if x, ok := n.X.(*ast.Ident); ok {
					file := g.fset.File(n.Pos()).Name()
					p, ok := g.fileImports[file][x.Name]
					if !ok {
						err = fmt.Errorf("%s refers to %s.%s, which isn't an import", name, x.Name, n.Sel.Name)
						return false
					}
					g.imports[x.Name] = p
					return false
				}
			case *ast.Ident:
				if g.topLevel[n.Name] && !g.reexported[n.Name] && !slices.Contains(typeParams, n.Name) {
					err = fmt.Errorf("%s refers to %s, which isn't re-exported", name, n.Name)
					return false
				}
			}
			return true
		})
	}
	return err
}

func (g *generator) fieldList(fields *ast.FieldList) string {
	var parts []string
	for _, field := range fields.List {
		var names []string
		for _, n := range field.Names {
			names = append(names, n.Name)
		}
		var typ bytes.Buffer
		_ = printer.Fprint(&typ, g.fset, field.Type)
		parts = append(parts, strings.Join(names, ", ")+" "+typ.String())
	}
	return strings.Join(parts, ", ")
}

// finish returns the generated file, with its header and imports, formatted.
func (g *generator) finish(pkgName, from string) ([]byte, error) {
	module := strings.SplitN(g.pkg.ImportPath, "/", 4)
	local := strings.Join(module[:min(len(module), 3)], "/")

	var std, thirdParty, own []string
	for name, p := range g.imports {
		spec := fmt.Sprintf("%q", p)
		if name != defaultImportName(p) {
			spec = name + " " + spec
		}
		switch {
		case !strings.Contains(strings.Split(p, "/")[0], "."):
			std = append(std, spec)
		case p == local || strings.HasPrefix(p, local+"/"):
			own = append(own, spec)
		default:
			thirdParty = append(thirdParty, spec)
		}
	}

	var out bytes.Buffer
	fmt.Fprintf(&out, "// Code generated by internal/cmd/reexport from %s; DO NOT EDIT.\n\n", from)
	fmt.Fprintf(&out, "package %s\n\nimport (\n", pkgName)
	for _, group := range [][]string{std, thirdParty, own} {
		if len(group) == 0 {
			continue
		}
		slices.Sort(group)
		for _, spec := range group {
			fmt.Fprintf(&out, "\t%s\n", spec)
		}
		out.WriteString("\n")
	}
	out.WriteString(")\n")
	out.Write(g.buf.Bytes())

	src, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("formatting generated code: %w\n%s", err, out.Bytes())
	}
	return src, nil
}
