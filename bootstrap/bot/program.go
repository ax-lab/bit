package bot

import (
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"sync"

	"axlab.dev/bit/input"
)

const (
	debugNodes = true
)

type Program struct {
	symbols SymbolTable
	sources input.SourceMap

	stdErr io.Writer
	stdOut io.Writer

	moduleLock sync.Mutex
	moduleMap  map[input.Source]Module
	mainModule Module

	errorLock sync.Mutex
	errorList []error
}

func ProgramNew() *Program {
	out := &Program{
		stdErr: os.Stderr,
		stdOut: os.Stdout,
	}

	out.symbols.Add(
		// punctuation
		".", "..", ",", ";", ":",
		// brackets
		"(", ")", "{", "}", "[", "]",
		// operators
		"!", "?",
		"=", "+", "-", "*", "/", "%",
		"==", "!=", "<", "<=", ">", ">=",
	)

	return out
}

func (program *Program) AddError(err error) {
	if err == nil {
		return
	}

	program.errorLock.Lock()
	defer program.errorLock.Unlock()
	program.errorList = append(program.errorList, err)
}

func (program *Program) HasErrors() bool {
	program.errorLock.Lock()
	defer program.errorLock.Unlock()
	return len(program.errorList) > 0
}

func (program *Program) LoadFile(file string) Module {
	src, err := program.sources.LoadFile(file)
	if err != nil {
		program.AddError(fmt.Errorf("loading file `%s`: %v", file, err))
		return Module{}
	}
	return program.loadSource(src)
}

func (program *Program) LoadString(name, text string) Module {
	src := program.sources.NewSource(name, text)
	return program.loadSource(src)
}

func (program *Program) Eval() {
	main := program.mainModule
	if debugNodes && main.Valid() {
		repr := ReprNew(os.Stdout)
		for _, it := range main.Nodes().Slice() {
			repr.Write("\n=> ")
			repr.OutputNode(it)
			repr.Write("\n")
		}
		fmt.Printf("\n")
	}
}

func (program *Program) Run() {
	if program.OutputErrors() {
		return
	}

	program.moduleLock.Lock()
	defer program.moduleLock.Unlock()

	main := program.mainModule
	if !main.Valid() {
		return
	}
}

func (program *Program) OutputErrors() (hasErrors bool) {
	var errs []error
	program.errorLock.Lock()
	errs = append(errs, program.errorList...)
	program.errorLock.Unlock()

	if len(errs) == 0 {
		return false
	}

	fmt.Fprintf(program.stdErr, "\nErrors:\n")
	for n, err := range errs {
		fmt.Fprintf(program.stdErr, "\n\t[%d] %s\n", n+1, input.TrimSta(input.Indent(err.Error())))
	}
	fmt.Fprintf(program.stdErr, "\n")
	return true
}

func (program *Program) GoOutput(goProgram *GoProgram, mainFile *GoFile) {
	const modMainFunc = "InitModule"

	if program.HasErrors() {
		return
	}

	moduleList := make([]Module, 0, len(program.moduleMap))
	for _, it := range program.moduleMap {
		if it != program.mainModule {
			moduleList = append(moduleList, it)
		}
	}
	slices.SortFunc(moduleList, func(a, b Module) int { return a.Cmp(b) })

	mainFunc := mainFile.Func("main", "")

	for _, mod := range moduleList {
		name := mod.Src().Name()
		base, name := path.Dir(name), path.Base(name)

		base = strings.ReplaceAll(base, "/", "_")
		name = strings.TrimSuffix(name, path.Ext(name))
		if base == "" || base == "." || base == "/" {
			base = name
		}

		modImport := fmt.Sprintf("%s/%s", goProgram.Module(), name)
		mainFile.Import(modImport)
		mainFunc.Push("%s.%s()", name, modMainFunc)

		filePath := path.Join(base, name+".go")
		file := goProgram.NewFile(filePath, name)
		block := file.Func(modMainFunc, "")
		mod.GoOutput(goProgram, block)
	}

	outVar := program.mainModule.GoOutput(goProgram, mainFunc)
	if outVar != GoVarNone {
		mainFunc.Push("_ = %s", outVar)
	}

	for _, err := range goProgram.Errors() {
		program.AddError(err)
	}
}

func (program *Program) loadSource(src input.Source) Module {
	program.moduleLock.Lock()
	defer program.moduleLock.Unlock()

	if mod, ok := program.moduleMap[src]; ok {
		return mod
	}

	cursor := src.Cursor()
	tokens, err := Lex(&cursor, &program.symbols)

	mod := moduleNew(program, NodeListNew(src, tokens...))

	if program.moduleMap == nil {
		program.moduleMap = make(map[input.Source]Module)
	}
	program.moduleMap[src] = mod

	if !program.mainModule.Valid() {
		program.mainModule = mod
	}

	if err == nil {
		for _, parseErr := range Parse(mod.data.nodes) {
			program.AddError(parseErr)
		}
	}

	if err != nil {
		program.AddError(fmt.Errorf("loading module `%s`: %v", mod.Name(), err))
	}

	return mod
}
