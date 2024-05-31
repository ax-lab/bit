package core

import (
	"fmt"
	"io"
	"os"
	"slices"
	"sync"
	"sync/atomic"
)

const MaxErrors = 32

type OpFunc func(mod *Module, list NodeList)

type Compiler struct {
	Sources SourceLoader
	Lexer   Lexer

	sync sync.Mutex

	sources []string
	ops     []OpFunc
	out     OpFunc

	redirectedStdOut io.Writer
	redirectedStdErr io.Writer

	errorCount atomic.Uint32

	fatalSync   sync.Mutex
	fatalErrors []error

	modules map[Source]*Module

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

func (compiler *Compiler) AddSource(name string) {
	compiler.sync.Lock()
	defer compiler.sync.Unlock()
	compiler.sources = append(compiler.sources, name)
}

func (compiler *Compiler) CreateRuntime() *Runtime {
	compiler.sync.Lock()
	var (
		ops = slices.Clone(compiler.ops)
		out = compiler.out
	)
	compiler.sync.Unlock()

	rt := runtimeNew(compiler)
	rt.ops = ops
	rt.out = out

	return rt
}

func (compiler *Compiler) StdOut() io.Writer {
	if compiler.redirectedStdOut != nil {
		return compiler.redirectedStdOut
	}
	return os.Stdout
}

func (compiler *Compiler) StdErr() io.Writer {
	if compiler.redirectedStdErr != nil {
		return compiler.redirectedStdErr
	}
	return os.Stderr
}

func (compiler *Compiler) RedirectStdOut(out io.Writer) {
	compiler.redirectedStdOut = out
}

func (compiler *Compiler) RedirectStdErr(out io.Writer) {
	compiler.redirectedStdErr = out
}

func (compiler *Compiler) DeclareOp(op OpFunc) {
	compiler.ops = append(compiler.ops, op)
}

func (compiler *Compiler) SetOutput(op OpFunc) {
	compiler.out = op
}

func (compiler *Compiler) HasErrors() bool {
	return compiler.errorCount.Load() > 0
}

func (compiler *Compiler) ShouldStop() bool {
	count := compiler.errorCount.Load()

	compiler.fatalSync.Lock()
	fatal := len(compiler.fatalErrors)
	compiler.fatalSync.Unlock()

	return count >= MaxErrors || fatal > 0
}

func (compiler *Compiler) Fatal(err error) {
	if err != nil {
		compiler.fatalSync.Lock()
		compiler.fatalErrors = append(compiler.fatalErrors, err)
		compiler.fatalSync.Unlock()
		compiler.incrementErrorCount()
	}
}

func (compiler *Compiler) Execute() bool {
	compiler.sync.Lock()
	defer compiler.sync.Unlock()

	loader := &compiler.Sources
	_, err := loader.getBaseDir()
	if err != nil {
		compiler.Fatal(err)
		goto end
	}

	for _, name := range compiler.sources {
		source, err := loader.Load(name)
		if err != nil {
			compiler.Fatal(err)
			continue
		}

		if mod := compiler.modules[source]; mod != nil {
			continue
		}

		if compiler.modules == nil {
			compiler.modules = make(map[Source]*Module)
		}

		mod := moduleNew(compiler, source)
		compiler.modules[source] = mod

		node := NodeNew(source.Span(), source)
		nodeList := mod.Nodes()
		nodeList.Push(node)

		compiler.Eval(mod, nodeList)
	}

	if compiler.HasErrors() {
		goto end
	}

	for _, op := range compiler.ops {
		compiler.nodeSync.Lock()
		if compiler.nodeChanged {
			slices.SortFunc(compiler.nodeLists, func(a, b nodeListEntry) int {
				return a.list.Span().Compare(b.list.Span())
			})
			compiler.nodeChanged = false
		}

		list := compiler.nodeLists[:]
		compiler.nodeSync.Unlock()

		for _, entry := range list {
			op(entry.module, entry.list)
		}

		if compiler.HasErrors() {
			break
		}
	}

	if compiler.HasErrors() {
		if stopped := compiler.ShouldStop(); stopped {
			compiler.Fatal(fmt.Errorf("too many errors, aborting compilation"))
		}
		goto end
	}

	if compiler.out != nil {
		var modules []*Module
		for _, it := range compiler.modules {
			modules = append(modules, it)
		}

		slices.SortFunc(modules, func(a, b *Module) int {
			return SourceCompare(a.source, b.source)
		})

		for _, module := range modules {
			compiler.out(module, module.nodes)
		}
	}

end:

	ok := compiler.outputErrors()
	return ok
}

func (compiler *Compiler) incrementErrorCount() (stop bool) {
	compiler.errorCount.Add(1)
	return compiler.ShouldStop()
}

func (compiler *Compiler) outputErrors() bool {
	var errors []error
	for _, ls := range compiler.modules {
		errors = append(errors, ls.Errors()...)
	}

	if len(errors) == 0 {
		return true
	}

	compiler.Dump(true)

	stdOut := compiler.StdOut()
	stdErr := compiler.StdErr()

	fmt.Fprintln(stdOut)
	SortErrors(errors)
	if cnt := len(errors); cnt == 1 {
		fmt.Fprintf(stdErr, "[FAIL] with error: %s\n", errors[0])
	} else {
		fmt.Fprintf(stdErr, "[FAIL] compilation generated %d errors:\n", cnt)
		for idx, err := range errors {
			fmt.Fprintf(stdErr, "\n[%d] %s\n", idx+1, err)
		}
	}

	if fatal := compiler.fatalErrors; len(fatal) > 0 {
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

func (compiler *Compiler) Dump(full bool) {
	out := compiler.StdOut()

	fmt.Fprintf(out, "\n-- COMPILER DUMP --\n\n")
	fmt.Fprintf(out, "-> Modules:     %d\n", len(compiler.modules))
	fmt.Fprintf(out, "-> Node Lists:  %d\n", len(compiler.nodeLists))

	var lists []nodeListEntry
	if !full {
		for _, it := range compiler.modules {
			lists = append(lists, nodeListEntry{it, it.nodes})
		}
	} else {
		lists = append(lists, compiler.nodeLists...)
	}

	slices.SortFunc(lists, func(a, b nodeListEntry) int {
		return a.list.Span().Compare(b.list.Span())
	})

	for idx, entry := range lists {
		repr := Indent(entry.list.Dump())
		fmt.Fprintf(out, "\n\t[%d of %d] = %s\n", idx+1, len(lists), repr)
	}
}

func (compiler *Compiler) OutputCode(expr Expr) {
	compiler.codeSync.Lock()
	defer compiler.codeSync.Unlock()
	compiler.codeList = append(compiler.codeList, expr)
}

func (compiler *Compiler) Eval(mod *Module, list NodeList) {
	compiler.nodeSync.Lock()
	compiler.nodeLists = append(compiler.nodeLists, nodeListEntry{mod, list})
	compiler.nodeChanged = true
	compiler.nodeSync.Unlock()
}
