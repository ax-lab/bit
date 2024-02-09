package bit

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type CppContext struct {
	program          *Program
	root             *CppContext
	includeSystem    []string
	includeLocal     []string
	includeSystemMap map[string]bool
	includeLocalMap  map[string]bool

	OutputFilePrefix *CppWriter
	OutputFileSuffix *CppWriter
	OutputFunc       *CppWriter
	OutputExpr       *CppWriter
}

type CppWriter struct {
	Context *CppContext
	Output  strings.Builder

	indent string
}

func NewCppContext(program *Program) *CppContext {
	cpp := &CppContext{
		program: program,
	}
	cpp.root = cpp
	cpp.OutputFilePrefix = &CppWriter{Context: cpp}
	cpp.OutputFileSuffix = &CppWriter{Context: cpp}
	return cpp
}

func (ctx *CppContext) InitExpr(parent *CppContext) {
	*ctx = *parent
	ctx.OutputExpr = &CppWriter{Context: parent}
}

func (ctx *CppContext) InitFunc(parent *CppContext) {
	*ctx = *parent
	ctx.OutputFunc = &CppWriter{Context: parent}
	ctx.OutputExpr = &CppWriter{Context: parent}
}

func (ctx *CppContext) IncludeSystem(file string) {
	root := ctx.root
	if !root.includeSystemMap[file] {
		if root.includeSystemMap == nil {
			root.includeSystemMap = make(map[string]bool)
		}
		root.includeSystemMap[file] = true
		root.includeSystem = append(root.includeSystem, file)
	}
}

func (ctx *CppContext) IncludeLocal(file string) {
	root := ctx.root
	if !root.includeLocalMap[file] {
		if root.includeLocalMap == nil {
			root.includeLocalMap = make(map[string]bool)
		}
		root.includeLocalMap[file] = true
		root.includeLocal = append(root.includeLocal, file)
	}
}

func (cpp *CppWriter) Indent() {
	cpp.indent = cpp.indent + "\t"
}

func (cpp *CppWriter) Dedent() {
	cpp.indent = cpp.indent[:len(cpp.indent)-1]
}

func (cpp *CppWriter) IsNewLine() bool {
	str := cpp.Output.String()
	return str == "" || strings.HasSuffix(str, "\n")
}

func (cpp *CppWriter) NewLine() {
	if !cpp.IsNewLine() {
		cpp.Output.WriteRune('\n')
	}
}

func (cpp *CppWriter) Write(str string, args ...any) {
	if len(args) > 0 {
		str = fmt.Sprintf(str, args...)
	}
	if len(str) == 0 {
		return
	}

	if len(cpp.indent) > 0 {
		eol := cpp.IsNewLine()
		for n, line := range strings.Split(str, "\n") {
			if n > 0 {
				cpp.Output.WriteRune('\n')
				eol = true
			}
			if len(line) > 0 {
				if eol {
					cpp.Output.WriteString(cpp.indent)
				}
				cpp.Output.WriteString(line)
			}
		}

	} else {
		cpp.Output.WriteString(str)
	}
}

func (cpp *CppWriter) WriteLiteralString(str string) {
	out := &cpp.Output
	out.WriteRune('"')
	for _, chr := range str {
		cppOutputChar(chr, out)
	}
	out.WriteRune('"')
}

func cppOutputChar(chr rune, out *strings.Builder) {
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
			out.WriteRune(chr)
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

	if 'a' <= chr && chr <= 'a' {
		return true
	}

	if '0' <= chr && chr <= '9' {
		return true
	}

	return false
}
