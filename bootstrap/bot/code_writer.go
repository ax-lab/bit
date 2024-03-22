package bot

import (
	"fmt"
	"slices"
	"strings"
)

type CodeValue string

type CodeWriter struct {
	varCount int
	varLast  CodeValue
	lines    []string
	imports  map[string]bool
}

func (cw *CodeWriter) Import(name string) {
	if cw.imports == nil {
		cw.imports = make(map[string]bool)
	}
	cw.imports[name] = true
}

func (cw *CodeWriter) Last() CodeValue {
	if cw.varLast == "" {
		panic("CodeWriter: no variable set")
	}
	return cw.varLast
}

func (cw *CodeWriter) PushExpr(code string, args ...any) CodeValue {
	if len(args) > 0 {
		code = fmt.Sprintf(code, args...)
	}

	cw.varLast = CodeValue(fmt.Sprintf("v%04d", cw.varCount))
	cw.varCount++

	cw.lines = append(cw.lines, fmt.Sprintf("%s := %s", cw.varLast, code))
	return cw.varLast
}

func (cw *CodeWriter) Push(code string, args ...any) {
	if len(args) > 0 {
		code = fmt.Sprintf(code, args...)
	}
	cw.lines = append(cw.lines, code)
}

func (cw *CodeWriter) Output(out *CodeOutput, mainFile string) {
	text := strings.Builder{}
	text.WriteString("package main\n\n")

	if cnt := len(cw.imports); cnt > 0 {
		imports := make([]string, 0, cnt)
		for name := range cw.imports {
			imports = append(imports, name)
		}
		slices.Sort(imports)

		text.WriteString("import (\n")
		for _, name := range imports {
			text.WriteString(fmt.Sprintf("\t%#v\n", name))
		}
		text.WriteString(")\n\n")
	}

	text.WriteString("func main() {\n")
	for _, it := range cw.lines {
		text.WriteString("\t")
		text.WriteString(it)
		text.WriteString("\n")
	}
	text.WriteString("}\n")

	out.AddFile(mainFile, text.String())
}
