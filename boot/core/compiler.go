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

type Compiler struct {
	Lexer Lexer

	redirectedStdOut io.Writer
	redirectedStdErr io.Writer

	errorCount atomic.Uint32

	fatalSync   sync.Mutex
	fatalErrors []error

	list []NodeList
	ops  []func(list NodeList)
	out  func(list NodeList)
}

func (compiler *Compiler) Run() bool {
	if compiler.out == nil {
		panic("Compiler: no output function defined")
	}

	slices.SortFunc(compiler.list, func(a, b NodeList) int {
		return a.Span().Compare(b.Span())
	})

	for _, op := range compiler.ops {
		for _, ls := range compiler.list {
			op(ls)
		}
	}

	if compiler.HasErrors() {
		if stopped := compiler.ShouldStop(); stopped {
			compiler.Fatal(fmt.Errorf("too many errors, aborting compilation"))
		}
		compiler.outputErrors()
		return false
	}

	for _, ls := range compiler.list {
		compiler.out(ls)
	}

	ok := compiler.outputErrors()
	return ok
}

func (compiler *Compiler) AddSource(src Source) {
	if src.Loader().Compiler() != compiler {
		panic("cannot add source from a different compiler instance")
	}
	node := NodeNew(src.Span(), src)
	list := NodeListNew(src.Span(), node)
	compiler.Eval(list)
}

func (compiler *Compiler) Fatal(err error) {
	if err != nil {
		compiler.fatalSync.Lock()
		compiler.fatalErrors = append(compiler.fatalErrors, err)
		compiler.fatalSync.Unlock()
		compiler.incrementErrorCount()
	}
}

func (compiler *Compiler) Eval(list NodeList) {
	compiler.list = append(compiler.list, list)
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

func (compiler *Compiler) DeclareOp(op func(list NodeList)) {
	compiler.ops = append(compiler.ops, op)
}

func (compiler *Compiler) SetOutput(op func(list NodeList)) {
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

func (compiler *Compiler) incrementErrorCount() (stop bool) {
	compiler.errorCount.Add(1)
	return compiler.ShouldStop()
}

func (compiler *Compiler) outputErrors() bool {
	var errors []error
	for _, ls := range compiler.list {
		errors = append(errors, ls.Errors()...)
	}

	if len(errors) == 0 {
		return true
	}

	fmt.Fprintln(compiler.StdOut())
	stdErr := compiler.StdErr()
	SortErrors(errors)
	if cnt := len(errors); cnt == 1 {
		fmt.Fprintf(stdErr, "Error: %s\n", errors[0])
	} else {
		fmt.Fprintf(stdErr, "Compilation failed with %d errors:\n", cnt)
		for idx, err := range errors {
			fmt.Fprintf(stdErr, "\n[%d] %s\n", idx+1, err)
		}
	}

	if fatal := compiler.fatalErrors; len(fatal) > 0 {
		fmt.Fprintln(stdErr)
		for _, err := range fatal {
			fmt.Fprintf(stdErr, "Fatal: %s\n", err)
		}
	}

	fmt.Fprintln(compiler.StdOut())

	compiler.Dump()

	return false
}

func (compiler *Compiler) Dump() {
	out := compiler.StdOut()

	count := len(compiler.list)

	fmt.Fprintf(out, "\n-- COMPILER DUMP --\n\n")
	fmt.Fprintf(out, ">>> Lists (%d) <<<\n", count)

	for idx, list := range compiler.list {
		repr := Indent(list.Dump())
		fmt.Fprintf(out, "\n%s[%d of %d] = %s\n", DefaultIndent, idx+1, count, repr)
	}

	fmt.Fprintf(out, "\n")
}
