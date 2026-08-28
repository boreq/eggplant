// genhandlers generates logging decorators for application handlers.
//
// For every type ending in "Handler" with a method Execute, it emits a
// LoggingXxxHandler that wraps the original, logs parameters, return values,
// and execution time, then delegates.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

type param struct {
	name string
	typ  string
}

type handler struct {
	name        string // e.g. GetAlbumHandler
	params      []param
	results     []param           // named if Go declared them, else generated
	imports     map[string]string // alias -> import path used by this handler's file
	usedAliases map[string]bool   // package aliases referenced in params/results
}

func main() {
	var (
		dir          string
		out          string
		pkgName      string
		loggerImport string
		loggerType   string
		loggerKey    string
	)
	flag.StringVar(&dir, "dir", ".", "package directory to scan")
	flag.StringVar(&out, "out", "handlers_logging_gen.go", "output filename (relative to -dir)")
	flag.StringVar(&pkgName, "pkg", "", "package name of generated file (defaults to detected)")
	flag.StringVar(&loggerImport, "logger-import", "github.com/boreq/eggplant/internal/logging", "import path of logger package")
	flag.StringVar(&loggerType, "logger-type", "logging.Logger", "logger type expression")
	flag.StringVar(&loggerKey, "logger-key", "handler", "log key for handler name")
	flag.Parse()

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedImports,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		fatal("load: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		fatal("load errors in %s", dir)
	}
	if len(pkgs) != 1 {
		fatal("expected exactly one package in %s, got %d", dir, len(pkgs))
	}
	pkg := pkgs[0]
	if pkgName == "" {
		pkgName = pkg.Name
	}

	handlers := collectHandlers(pkg)
	if len(handlers) == 0 {
		fatal("no *Handler types with Execute method found in %s", dir)
	}

	src, err := render(pkgName, loggerImport, loggerType, loggerKey, handlers)
	if err != nil {
		fatal("render: %v", err)
	}

	target := filepath.Join(dir, out)
	if err := os.WriteFile(target, src, 0o644); err != nil {
		fatal("write %s: %v", target, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d handlers)\n", target, len(handlers))
}

func collectHandlers(pkg *packages.Package) []handler {
	files := sourceFiles(pkg)

	// First pass: find struct types named *Handler.
	structs := map[string]bool{}
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if !strings.HasSuffix(ts.Name.Name, "Handler") {
				return true
			}
			if _, ok := ts.Type.(*ast.StructType); !ok {
				return true
			}
			structs[ts.Name.Name] = true
			return true
		})
	}

	// Second pass: find Execute methods on *XxxHandler.
	var out []handler
	for _, file := range files {
		fileImports := collectImports(file)
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) != 1 {
				continue
			}
			if fd.Name.Name != "Execute" {
				continue
			}
			recvName := receiverTypeName(fd.Recv.List[0].Type)
			if recvName == "" || !structs[recvName] {
				continue
			}
			h := handler{name: recvName, imports: fileImports, usedAliases: map[string]bool{}}
			h.params = extractFields(fd.Type.Params, "p")
			if fd.Type.Results != nil {
				h.results = extractFields(fd.Type.Results, "r")
			}
			for _, f := range fd.Type.Params.List {
				for _, q := range qualifiersUsed(f.Type) {
					h.usedAliases[q] = true
				}
			}
			if fd.Type.Results != nil {
				for _, f := range fd.Type.Results.List {
					for _, q := range qualifiersUsed(f.Type) {
						h.usedAliases[q] = true
					}
				}
			}
			out = append(out, h)
		}
	}

	slices.SortFunc(out, func(a, b handler) int { return strings.Compare(a.name, b.name) })
	return out
}

func sourceFiles(pkg *packages.Package) []*ast.File {
	var out []*ast.File
	for i, f := range pkg.Syntax {
		if i >= len(pkg.CompiledGoFiles) {
			out = append(out, f)
			continue
		}
		base := filepath.Base(pkg.CompiledGoFiles[i])
		if strings.HasSuffix(base, "_gen.go") || strings.HasSuffix(base, "_test.go") {
			continue
		}
		out = append(out, f)
	}
	return out
}

func collectImports(file *ast.File) map[string]string {
	out := map[string]string{}
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			alias = filepath.Base(path)
		}
		if alias == "_" || alias == "." {
			continue
		}
		out[alias] = path
	}
	return out
}

func qualifiersUsed(e ast.Expr) []string {
	var out []string
	ast.Inspect(e, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		out = append(out, id.Name)
		return true
	})
	return out
}

func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

func extractFields(fl *ast.FieldList, fallbackPrefix string) []param {
	var out []param
	idx := 0
	for _, f := range fl.List {
		typ := exprString(f.Type)
		if len(f.Names) == 0 {
			out = append(out, param{name: fmt.Sprintf("%s%d", fallbackPrefix, idx), typ: typ})
			idx++
			continue
		}
		for _, n := range f.Names {
			name := n.Name
			if name == "" || name == "_" {
				name = fmt.Sprintf("%s%d", fallbackPrefix, idx)
			}
			out = append(out, param{name: name, typ: typ})
			idx++
		}
	}
	return out
}

func exprString(e ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), e); err != nil {
		return fmt.Sprintf("/*err:%v*/", err)
	}
	return buf.String()
}

func render(pkgName, loggerImport, loggerType, loggerKey string, handlers []handler) ([]byte, error) {
	imports := map[string]bool{
		"time":       true,
		loggerImport: true,
	}
	for _, h := range handlers {
		for alias := range h.usedAliases {
			if path, ok := h.imports[alias]; ok {
				imports[path] = true
			}
		}
	}

	var body bytes.Buffer
	for _, h := range handlers {
		writeHandler(&body, h, loggerType, loggerKey)
	}

	var src bytes.Buffer
	fmt.Fprintln(&src, "// Code generated by genhandlers. DO NOT EDIT.")
	fmt.Fprintln(&src)
	fmt.Fprintf(&src, "package %s\n\n", pkgName)
	fmt.Fprintln(&src, "import (")
	var stdlib, third []string
	for k := range imports {
		if strings.Contains(strings.SplitN(k, "/", 2)[0], ".") {
			third = append(third, k)
		} else {
			stdlib = append(stdlib, k)
		}
	}
	slices.Sort(stdlib)
	slices.Sort(third)
	for _, k := range stdlib {
		fmt.Fprintf(&src, "\t%q\n", k)
	}
	if len(stdlib) > 0 && len(third) > 0 {
		fmt.Fprintln(&src)
	}
	for _, k := range third {
		fmt.Fprintf(&src, "\t%q\n", k)
	}
	fmt.Fprintln(&src, ")")
	fmt.Fprintln(&src)
	src.Write(body.Bytes())

	return format.Source(src.Bytes())
}

func writeHandler(w *bytes.Buffer, h handler, loggerType, loggerKey string) {
	short := strings.TrimSuffix(h.name, "Handler")
	loggingName := "Logging" + h.name
	errIdx := -1
	if n := len(h.results); n > 0 && h.results[n-1].typ == "error" {
		errIdx = n - 1
	}

	fmt.Fprintf(w, "type %s struct {\n", loggingName)
	fmt.Fprintf(w, "\tinner  *%s\n", h.name)
	fmt.Fprintf(w, "\tlogger %s\n", loggerType)
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func New%s(inner *%s, logger %s) *%s {\n", loggingName, h.name, loggerType, loggingName)
	fmt.Fprintf(w, "\treturn &%s{inner: inner, logger: logger}\n", loggingName)
	fmt.Fprintf(w, "}\n\n")

	// Signature.
	paramDecl := paramDecls(h.params)
	resultDecl := resultDecls(h.results)
	fmt.Fprintf(w, "func (h *%s) Execute(%s)", loggingName, paramDecl)
	if resultDecl != "" {
		fmt.Fprintf(w, " (%s)", resultDecl)
	}
	fmt.Fprintln(w, " {")

	// Body.
	fmt.Fprintln(w, "\tstart := time.Now()")

	// Call inner.
	callArgs := make([]string, len(h.params))
	for i, p := range h.params {
		callArgs[i] = p.name
	}
	if len(h.results) == 0 {
		fmt.Fprintf(w, "\th.inner.Execute(%s)\n", strings.Join(callArgs, ", "))
	} else {
		retNames := make([]string, len(h.results))
		for i := range h.results {
			retNames[i] = fmt.Sprintf("ret%d", i)
		}
		fmt.Fprintf(w, "\t%s := h.inner.Execute(%s)\n", strings.Join(retNames, ", "), strings.Join(callArgs, ", "))
	}

	// Build log kv list.
	var kvs []string
	kvs = append(kvs, fmt.Sprintf("%q, %q", loggerKey, short))
	for _, p := range h.params {
		if p.typ == "context.Context" {
			continue
		}
		kvs = append(kvs, fmt.Sprintf("%q, %s", p.name, p.name))
	}
	for i, r := range h.results {
		label := fmt.Sprintf("ret%d", i)
		if r.typ == "error" {
			label = "err"
		}
		kvs = append(kvs, fmt.Sprintf("%q, ret%d", label, i))
	}
	kvs = append(kvs, fmt.Sprintf("%q, time.Since(start)", "duration"))

	if errIdx >= 0 {
		fmt.Fprintf(w, "\tlogFn := h.logger.Debug\n")
		fmt.Fprintf(w, "\tmsg := %q\n", "handler executed")
		fmt.Fprintf(w, "\tif ret%d != nil && !isNonLoggableError(ret%d) {\n", errIdx, errIdx)
		fmt.Fprintf(w, "\t\tlogFn = h.logger.Error\n")
		fmt.Fprintf(w, "\t\tmsg = %q\n", "handler failed")
		fmt.Fprintf(w, "\t}\n")
		fmt.Fprintln(w, "\tlogFn(msg,")
	} else {
		fmt.Fprintf(w, "\th.logger.Debug(%q,\n", "handler executed")
	}
	for _, kv := range kvs {
		fmt.Fprintf(w, "\t\t%s,\n", kv)
	}
	if errIdx >= 0 {
		fmt.Fprintln(w, "\t)")
	} else {
		fmt.Fprintln(w, "\t)")
	}

	// Return.
	if len(h.results) > 0 {
		retNames := make([]string, len(h.results))
		for i := range h.results {
			retNames[i] = fmt.Sprintf("ret%d", i)
		}
		fmt.Fprintf(w, "\treturn %s\n", strings.Join(retNames, ", "))
	}
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)
}

func paramDecls(ps []param) string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = fmt.Sprintf("%s %s", p.name, p.typ)
	}
	return strings.Join(out, ", ")
}

func resultDecls(rs []param) string {
	if len(rs) == 0 {
		return ""
	}
	// Always use typed-only signature on the wrapper (no names needed).
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.typ
	}
	return strings.Join(out, ", ")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "genhandlers: "+format+"\n", args...)
	os.Exit(1)
}
