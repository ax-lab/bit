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
	nodeLists   []nodeListEntry
	nodeChanged bool

	codeSync sync.Mutex
	codeList []Expr
}

type nodeListEntry struct {
	module *Module
	list   NodeList
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

func (rt *Runtime) StdOut() io.Writer {
	return rt.stdOut
}

func (rt *Runtime) StdErr() io.Writer {
	return rt.stdErr
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

func (rt *Runtime) RunCompiler() bool {

	loader := &rt.compiler.Sources
	_, err := loader.getBaseDir()
	if err != nil {
		rt.Fatal(err)
		goto end
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

		rt.Eval(mod, nodeList)
	}

	if rt.HasErrors() {
		goto end
	}

	for _, op := range rt.ops {
		rt.nodeSync.Lock()
		if rt.nodeChanged {
			slices.SortFunc(rt.nodeLists, func(a, b nodeListEntry) int {
				return a.list.Span().Compare(b.list.Span())
			})
			rt.nodeChanged = false
		}

		list := rt.nodeLists[:]
		rt.nodeSync.Unlock()

		for _, entry := range list {
			op(entry.module, entry.list)
		}

		if rt.HasErrors() {
			break
		}
	}

	if rt.HasErrors() {
		if stopped := rt.ShouldStop(); stopped {
			rt.Fatal(fmt.Errorf("too many errors, aborting compilation"))
		}
		goto end
	}

	if rt.out != nil {
		var modules []*Module
		for _, it := range rt.modules {
			modules = append(modules, it)
		}

		slices.SortFunc(modules, func(a, b *Module) int {
			return SourceCompare(a.source, b.source)
		})

		for _, module := range modules {
			rt.out(module, module.nodes)
		}
	}

end:

	ok := rt.outputErrors()
	return ok
}

func (rt *Runtime) RunCode() (out Value, err error) {
	for _, it := range rt.codeList {
		out, err = it.Eval(rt)
		if err != nil {
			break
		}
	}
	return
}

func (rt *Runtime) OutputCode(expr Expr) {
	rt.codeSync.Lock()
	defer rt.codeSync.Unlock()
	rt.codeList = append(rt.codeList, expr)
}

func (rt *Runtime) Eval(mod *Module, list NodeList) {
	rt.nodeSync.Lock()
	rt.nodeLists = append(rt.nodeLists, nodeListEntry{mod, list})
	rt.nodeChanged = true
	rt.nodeSync.Unlock()
}

func (rt *Runtime) Dump(full bool) {
	out := rt.stdOut

	fmt.Fprintf(out, "\n-- COMPILER DUMP --\n\n")
	fmt.Fprintf(out, "-> Modules:     %d\n", len(rt.modules))
	fmt.Fprintf(out, "-> Node Lists:  %d\n", len(rt.nodeLists))

	var lists []nodeListEntry
	if !full {
		for _, it := range rt.modules {
			lists = append(lists, nodeListEntry{it, it.nodes})
		}
	} else {
		lists = append(lists, rt.nodeLists...)
	}

	slices.SortFunc(lists, func(a, b nodeListEntry) int {
		return a.list.Span().Compare(b.list.Span())
	})

	for idx, entry := range lists {
		repr := Indent(entry.list.Dump())
		fmt.Fprintf(out, "\n\t[%d of %d] = %s\n", idx+1, len(lists), repr)
	}
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

	rt.Dump(true)

	fmt.Fprintln(rt.stdOut)
	stdErr := rt.stdErr
	SortErrors(errors)
	if cnt := len(errors); cnt == 1 {
		fmt.Fprintf(stdErr, "[FAIL] with error: %s\n", errors[0])
	} else {
		fmt.Fprintf(stdErr, "[FAIL] compilation generated %d errors:\n", cnt)
		for idx, err := range errors {
			fmt.Fprintf(stdErr, "\n[%d] %s\n", idx+1, err)
		}
	}

	if fatal := rt.fatalErrors; len(fatal) > 0 {
		fmt.Fprintln(stdErr)
		for n, err := range fatal {
			if n == 0 {
				fmt.Fprintf(stdErr, "\n")
			}
			fmt.Fprintf(stdErr, "Fatal: %s\n", err)
		}
	}

	return false
}
