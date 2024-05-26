package core

import (
	"fmt"
	"io"
	"os"
	"slices"
)

type Compiler struct {
	Lexer Lexer

	redirectedStdOut io.Writer
	redirectedStdErr io.Writer

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

	if !compiler.checkAndOutputErrors() {
		return false
	}

	for _, ls := range compiler.list {
		compiler.out(ls)
	}

	return true
}

func (compiler *Compiler) AddSource(src Source) {
	if src.Loader().Compiler() != compiler {
		panic("cannot add source from a different compiler instance")
	}
	node := NodeNew(src.Span(), src)
	list := NodeListNew(src.Span(), node)
	compiler.Eval(list)
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

func (compiler *Compiler) checkAndOutputErrors() bool {
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
	fmt.Fprintln(compiler.StdOut())
	return false
}
