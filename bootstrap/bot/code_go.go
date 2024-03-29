package bot

import (
	"fmt"
	"strings"
)

type GoType string

type GoVar string

const (
	GoVarNone  GoVar  = ""
	GoTypeNone GoType = ""

	GoTypeStr GoType = "string"
	GoTypeInt GoType = "int"
)

type GoCode interface {
	GoType() GoType
	GoOutput(blk *GoBlock) GoVar
}

type GoProgram struct {
	module   string
	errors   []error
	files    map[string]*GoFile
	mainFile string
}

func GoProgramNew(module string, mainFile string) (*GoProgram, *GoFile) {
	out := &GoProgram{module: module, mainFile: mainFile}
	main := out.NewFile(mainFile, "main")
	return out, main
}

func (program *GoProgram) Module() string {
	return program.module
}

func (program *GoProgram) NewFile(name, module string) *GoFile {
	file := &GoFile{
		module:  module,
		program: program,
	}
	if program.files == nil {
		program.files = make(map[string]*GoFile)
	}
	program.files[name] = file
	return file
}

func (program *GoProgram) HasErrors() bool {
	return len(program.errors) > 0
}

func (program *GoProgram) AddError(err error) {
	if err != nil {
		program.errors = append(program.errors, err)
	}
}

func (program *GoProgram) Errors() (out []error) {
	out = append(out, program.errors...)
	return out
}

type GoFile struct {
	program *GoProgram
	module  string
	imports map[string]bool
	funcs   map[string]*GoBlock
}

func (file *GoFile) Program() *GoProgram {
	return file.program
}

func (file *GoFile) Import(name string) {
	if file.imports == nil {
		file.imports = make(map[string]bool)
	}
	file.imports[name] = true
}

func (file *GoFile) HasErrors() bool {
	return file.Program().HasErrors()
}

func (file *GoFile) AddError(err error) {
	file.Program().AddError(err)
}

func (file *GoFile) Func(name, result string, args ...string) *GoBlock {
	var header strings.Builder
	header.WriteString(fmt.Sprintf("func %s(", name))
	for n, arg := range args {
		if n > 0 {
			header.WriteString(", ")
		}
		header.WriteString(arg)
	}
	header.WriteString(")")
	if result != "" {
		header.WriteString(" ")
		header.WriteString(result)
	}
	header.WriteString(" {")

	out := &GoBlock{
		file:   file,
		header: header.String(),
		footer: "}",
		indent: 1,
	}

	if file.funcs == nil {
		file.funcs = make(map[string]*GoBlock)
	}
	file.funcs[name] = out
	return out
}

type GoBlock struct {
	file     *GoFile
	header   string
	footer   string
	vars     [][2]string
	varCount int
	lines    []string
	indent   int
}

func (blk *GoBlock) File() *GoFile {
	return blk.file
}

func (blk *GoBlock) Import(name string) {
	blk.File().Import(name)
}

func (blk *GoBlock) Program() *GoProgram {
	return blk.file.Program()
}

func (blk *GoBlock) HasErrors() bool {
	return blk.Program().HasErrors()
}

func (blk *GoBlock) AddError(err error) {
	blk.Program().AddError(err)
}

func (blk *GoBlock) Expr(code string, args ...any) GoVar {
	if len(args) > 0 {
		code = fmt.Sprintf(code, args)
	}

	name := blk.VarName()
	blk.Push("%s := %s", name, code)
	return name
}

func (blk *GoBlock) Push(code string, args ...any) {
	if len(args) > 0 {
		code = fmt.Sprintf(code, args)
	}
	if blk.indent > 0 {
		code = strings.Repeat("\t", blk.indent) + code
	}
	blk.lines = append(blk.lines, code)
}

func (blk *GoBlock) Indent() {
	blk.indent++
}

func (blk *GoBlock) Dedent() {
	blk.indent--
}

func (blk *GoBlock) VarName() GoVar {
	name := fmt.Sprintf("v%04X", blk.varCount)
	blk.varCount++
	return GoVar(name)
}

func (blk *GoBlock) Declare(name GoVar, typ GoType) {
	blk.vars = append(blk.vars, [2]string{string(name), string(typ)})
}
