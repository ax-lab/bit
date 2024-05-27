package core

import (
	"fmt"
	"io"
	"os"
	"slices"
	"sync"
	"sync/atomic"
)

type Runtime struct {
	errorCount atomic.Uint32

	fatalSync   sync.Mutex
	fatalErrors []error

	compiler *Compiler
	lexer    *Lexer
	sources  []string
	modules  map[Source]*Module

	ops []OpFunc
	out OpFunc

	stdErr io.Writer
	stdOut io.Writer

	nodeSync    sync.Mutex
	nodeLists   []NodeList
	nodeChanged bool
}

func runtimeNew(compiler *Compiler) *Runtime {
	rt := &Runtime{
		compiler: compiler,
		lexer:    compiler.Lexer.Copy(),
		modules:  make(map[Source]*Module),
	}

	if compiler.redirectedStdErr != nil {
		rt.stdErr = compiler.redirectedStdErr
	} else {
		rt.stdErr = os.Stderr
	}

	if compiler.redirectedStdOut != nil {
		rt.stdOut = compiler.redirectedStdOut
	} else {
		rt.stdOut = os.Stdout
	}

	return rt
}

func (rt *Runtime) Compiler() *Compiler {
	return rt.compiler
}

func (rt *Runtime) HasErrors() bool {
	return rt.errorCount.Load() > 0
}

func (rt *Runtime) ShouldStop() bool {
	count := rt.errorCount.Load()

	rt.fatalSync.Lock()
	fatal := len(rt.fatalErrors)
	rt.fatalSync.Unlock()

	return count >= MaxErrors || fatal > 0
}

func (rt *Runtime) Fatal(err error) {
	if err != nil {
		rt.fatalSync.Lock()
		rt.fatalErrors = append(rt.fatalErrors, err)
		rt.fatalSync.Unlock()
		rt.incrementErrorCount()
	}
}

func (rt *Runtime) Run() bool {

	loader := &rt.compiler.Sources
	_, err := loader.getBaseDir()
	if err != nil {
		rt.Fatal(err)
		return false
	}

	for _, name := range rt.sources {
		source, err := loader.Load(name)
		if err != nil {
			rt.Fatal(err)
			continue
		}

		if mod := rt.modules[source]; mod != nil {
			continue
		}

		mod := moduleNew(rt, source)
		rt.modules[source] = mod

		node := NodeNew(source.Span(), source)
		nodeList := mod.Nodes()
		nodeList.Push(node)

		rt.Eval(nodeList)
	}

	if rt.HasErrors() {
		return false
	}

	for _, op := range rt.ops {
		rt.nodeSync.Lock()
		if rt.nodeChanged {
			slices.SortFunc(rt.nodeLists, func(a, b NodeList) int {
				return a.Span().Compare(b.Span())
			})
			rt.nodeChanged = false
		}

		list := rt.nodeLists[:]
		rt.nodeSync.Unlock()

		for _, ls := range list {
			op(ls)
		}

		if rt.HasErrors() {
			break
		}
	}

	if rt.HasErrors() {
		if stopped := rt.ShouldStop(); stopped {
			rt.Fatal(fmt.Errorf("too many errors, aborting compilation"))
		}
		rt.outputErrors()
		return false
	}

	if rt.out != nil {
		for _, ls := range rt.nodeLists {
			rt.out(ls)
		}
	}

	ok := rt.outputErrors()

	rt.Dump()

	return ok
}

func (rt *Runtime) Eval(list NodeList) {
	rt.nodeSync.Lock()
	rt.nodeLists = append(rt.nodeLists, list)
	rt.nodeChanged = true
	rt.nodeSync.Unlock()
}

func (rt *Runtime) Dump() {
	out := rt.stdOut

	count := len(rt.nodeLists)

	fmt.Fprintf(out, "\n-- COMPILER DUMP --\n\n")
	fmt.Fprintf(out, ">>> Lists (%d) <<<\n", count)

	for idx, list := range rt.nodeLists {
		repr := Indent(list.Dump())
		fmt.Fprintf(out, "\n%s[%d of %d] = %s\n", DefaultIndent, idx+1, count, repr)
	}

	fmt.Fprintf(out, "\n")
}

func (rt *Runtime) incrementErrorCount() (stop bool) {
	rt.errorCount.Add(1)
	return rt.ShouldStop()
}

func (rt *Runtime) outputErrors() bool {
	var errors []error
	for _, ls := range rt.modules {
		errors = append(errors, ls.Errors()...)
	}

	if len(errors) == 0 {
		return true
	}

	fmt.Fprintln(rt.stdOut)
	stdErr := rt.stdErr
	SortErrors(errors)
	if cnt := len(errors); cnt == 1 {
		fmt.Fprintf(stdErr, "Error: %s\n", errors[0])
	} else {
		fmt.Fprintf(stdErr, "Compilation failed with %d errors:\n", cnt)
		for idx, err := range errors {
			fmt.Fprintf(stdErr, "\n[%d] %s\n", idx+1, err)
		}
	}

	if fatal := rt.fatalErrors; len(fatal) > 0 {
		fmt.Fprintln(stdErr)
		for _, err := range fatal {
			fmt.Fprintf(stdErr, "Fatal: %s\n", err)
		}
	}

	fmt.Fprintln(rt.stdOut)

	return false
}
