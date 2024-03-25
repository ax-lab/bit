package bot

import (
	"fmt"
	"io"
	"os"
	"sync"

	"axlab.dev/bit/input"
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

func (program *Program) Run() {
	var errs []error
	program.errorLock.Lock()
	errs = append(errs, program.errorList...)
	program.errorLock.Unlock()

	if len(errs) > 0 {
		fmt.Fprintf(program.stdErr, "\nErrors:\n")
		for n, err := range errs {
			fmt.Fprintf(program.stdErr, "\n\t[%d] %s\n", n+1, input.TrimSta(input.Indent(err.Error())))
		}
		fmt.Fprintf(program.stdErr, "\n")
		return
	}

	program.moduleLock.Lock()
	defer program.moduleLock.Unlock()

	main := program.mainModule
	if !main.Valid() {
		return
	}

	for _, it := range main.Nodes().Slice() {
		repr := it.Repr()
		span := it.Span()
		fmt.Printf("\n=> %s\n   at %s: %#v\n", repr, span.Location(), span.Text())
	}
	fmt.Printf("\n")
}

func (program *Program) loadSource(src input.Source) Module {
	program.moduleLock.Lock()
	defer program.moduleLock.Unlock()

	if mod, ok := program.moduleMap[src]; ok {
		return mod
	}

	cursor := src.Cursor()
	tokens, err := Lex(&cursor, &program.symbols)

	mod := Module{
		&moduleData{
			nodes: NodeListNew(src, tokens...),
		},
	}

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
