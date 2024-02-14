package bit

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

/*
	TODO: CppContext

	- sub-context init is error prone (return new context instead of init)
*/

type CppContext struct {
	Names   *NameScope
	Parent  *CppContext
	Program *Program
	File    *CppFile
	Func    *CppFunc
	Body    *CppBody
	Expr    *strings.Builder
}

func NewCppContext(program *Program) *CppContext {
	ctx := &CppContext{
		Names:   &program.names.root,
		Program: program,
		Expr:    &strings.Builder{},
	}
	ctx.File = &CppFile{
		Context: ctx,
	}
	ctx.Func = &CppFunc{
		Context: ctx,
		File:    ctx.File,
	}
	ctx.Body = &ctx.Func.Body
	return ctx
}

func (ctx *CppContext) initFrom(parent *CppContext) {
	ctx.Parent = parent
	ctx.Program = parent.Program
	ctx.Names = parent.Names
	ctx.File = parent.File
	ctx.Func = parent.Func
	ctx.Body = parent.Body
	ctx.Expr = &strings.Builder{}
}

func (ctx *CppContext) NewName(base string) string {
	return ctx.Names.DeclareUnique(base)
}

func (ctx *CppContext) NewExpr(parent *CppContext) {
	ctx.initFrom(parent)
}

func (ctx *CppContext) NewFunc(parent *CppContext) {
	ctx.initFrom(parent)
	ctx.Func = &CppFunc{
		Context: ctx,
		File:    ctx.File,
	}
	ctx.Body = &ctx.Func.Body
}

func (ctx *CppContext) NewBody(parent *CppContext) {
	ctx.initFrom(parent)
	ctx.Body = &CppBody{}
}

func (ctx *CppContext) IncludeSystem(file string) {
	root := ctx.File
	if !root.includeSystemMap[file] {
		if root.includeSystemMap == nil {
			root.includeSystemMap = make(map[string]bool)
		}
		root.includeSystemMap[file] = true
		root.includeSystem = append(root.includeSystem, file)
	}
}

func (ctx *CppContext) IncludeLocal(file string) {
	root := ctx.File
	if !root.includeLocalMap[file] {
		if root.includeLocalMap == nil {
			root.includeLocalMap = make(map[string]bool)
		}
		root.includeLocalMap[file] = true
		root.includeLocal = append(root.includeLocal, file)
	}
}

func (ctx *CppContext) GetOutputFiles(mainFile string) (out map[string]string) {
	out = make(map[string]string)
	out[mainFile] = ctx.File.Text()
	return
}

type CppFile struct {
	Context *CppContext
	Header  CppLines
	Body    CppLines
	Footer  CppLines

	includeSystem    []string
	includeLocal     []string
	includeSystemMap map[string]bool
	includeLocalMap  map[string]bool
}

type CppFunc struct {
	Context *CppContext
	Decl    string
	File    *CppFile
	Body    CppBody
}

type CppBody struct {
	CppLines
	Decl CppLines
}

func (body *CppBody) Len() int {
	return body.Decl.Len() + body.CppLines.Len()
}

func (body *CppBody) AppendTo(lines *CppLines) {
	if body.Len() > 0 {
		lines.Push("{")
		lines.Indent()
		body.Decl.AppendTo(lines)
		body.CppLines.AppendTo(lines)
		lines.Dedent()
		lines.Push("}")
	} else {
		lines.Push("{ }")
	}
}

type CppLines struct {
	indent string
	lines  []string
	text   string
}

func (txt *CppLines) Len() int {
	return len(txt.lines)
}

func (txt *CppLines) Indent() {
	txt.indent += "\t"
}

func (txt *CppLines) Dedent() {
	txt.indent = txt.indent[:len(txt.indent)-1]
}

func (txt *CppLines) NewLine() {
	txt.lines = append(txt.lines, "")
}

func (txt *CppLines) EnsureBlank() {
	if len(txt.lines) > 0 && txt.lines[len(txt.lines)-1] != "" {
		txt.Push("")
	}
}

func (txt *CppLines) AppendTo(other *CppLines) {
	for _, it := range txt.lines {
		other.Push(it)
	}
}

func (txt *CppLines) Push(text string, args ...any) {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	if len(text) > 0 {
		text = txt.indent + text
	}
	txt.lines = append(txt.lines, text)
	txt.text = ""
}

func (txt *CppLines) WriteString(s string) (n int, err error) {
	txt.Write(s)
	return len(s), nil
}

func (txt *CppLines) Write(s string) {
	if len(txt.lines) == 0 {
		txt.lines = append(txt.lines, s)
	} else {
		txt.lines[len(txt.lines)-1] += s
	}
}

func (txt *CppLines) Text() string {
	if len(txt.text) == 0 {
		txt.text = strings.Join(txt.lines, "\n")
	}
	return txt.text
}

func (file *CppFile) Text() string {
	header := fmt.Sprintf("// AUTO-GENERATED BY BUILD AT %s\n", time.Now().Format(time.RFC3339))
	text := strings.Builder{}
	text.WriteString(header)

	if len(file.includeSystem) > 0 {
		text.WriteString("\n")
		for _, it := range file.includeSystem {
			text.WriteString(fmt.Sprintf("#include <%s>\n", it))
		}
	}

	if len(file.includeLocal) > 0 {
		text.WriteString("\n")
		for _, it := range file.includeSystem {
			text.WriteString(fmt.Sprintf("#include \"%s\"\n", it))
		}
	}

	if txt := file.Header.Text(); len(txt) > 0 {
		text.WriteString("\n")
		text.WriteString(txt)
		text.WriteString("\n")
	}

	if txt := file.Body.Text(); len(txt) > 0 {
		text.WriteString("\n")
		text.WriteString(txt)
		text.WriteString("\n")
	}

	if txt := file.Footer.Text(); len(txt) > 0 {
		text.WriteString("\n")
		text.WriteString(txt)
		text.WriteString("\n")
	}

	return text.String()
}

func (fn *CppFunc) AppendTo(file *CppFile) {
	file.Header.EnsureBlank()
	file.Header.Push(fn.Decl + ";")

	file.Body.EnsureBlank()
	file.Body.Push(fn.Decl)
	fn.Body.AppendTo(&file.Body)
}

func WriteLiteralString(out io.StringWriter, str string) {
	out.WriteString("\"")
	for _, chr := range str {
		cppOutputChar(chr, out)
	}
	out.WriteString("\"")
}

func cppOutputChar(chr rune, out io.StringWriter) {
	seq := ""
	switch chr {
	case '?':
		seq = "\\?"
	case '"':
		seq = "\\\""
	case '\'':
		seq = "\\'"
	case '\\':
		seq = "\\\\"
	case '\x00':
		seq = "\\0"
	case '\t':
		seq = "\\t"
	case '\n':
		seq = "\\n"
	case '\r':
		seq = "\\r"
	case '\x08':
		seq = "\\b"
	default:
		if cppIsSafeStrChar(chr) {
			out.WriteString(string(chr))
		} else {
			buf := [utf8.UTFMax]byte{}
			len := utf8.EncodeRune(buf[:], chr)
			for _, b := range buf[:len] {
				out.WriteString(fmt.Sprintf("\\x%X", b))
			}
		}
	}
	if seq != "" {
		out.WriteString(seq)
	}
}

func cppIsSafeStrChar(chr rune) bool {
	switch chr {
	case
		'_', ' ', '!', '#', '$', '%', '&', '(', ')', '*', '+', ',', '-', '.', '/',
		':', ';', '<', '=', '>', '@', '[', ']', '^', '`', '{', '|', '}', '~':
		return true
	}

	if 'A' <= chr && chr <= 'Z' {
		return true
	}

	if 'a' <= chr && chr <= 'z' {
		return true
	}

	if '0' <= chr && chr <= '9' {
		return true
	}

	return false
}
