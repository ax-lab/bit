package core

import (
	"io"
	"os"
	"slices"
)

type Compiler struct {
	stdOut io.Writer
	stdErr io.Writer

	list []NodeList
	ops  []func(list NodeListWriter)
	out  func(comp *Compiler, list NodeList)
}

func (comp *Compiler) Run() {
	if comp.out == nil {
		panic("Compiler: no output function defined")
	}

	slices.SortFunc(comp.list, func(a, b NodeList) int {
		return a.Span().Compare(b.Span())
	})

	for _, op := range comp.ops {
		for _, ls := range comp.list {
			writer := NodeListWriter{ls}
			op(writer)
		}
	}

	for _, ls := range comp.list {
		comp.out(comp, ls)
	}
}

func (comp *Compiler) Add(list NodeList) {
	list.checkValid()
	comp.list = append(comp.list, list)
}

func (comp *Compiler) StdOut() io.Writer {
	if comp.stdOut != nil {
		return comp.stdOut
	}
	return os.Stdout
}

func (comp *Compiler) StdErr() io.Writer {
	if comp.stdErr != nil {
		return comp.stdErr
	}
	return os.Stderr
}

func (comp *Compiler) RedirectStdOut(out io.Writer) {
	comp.stdOut = out
}

func (comp *Compiler) RedirectStdErr(out io.Writer) {
	comp.stdErr = out
}

func (comp *Compiler) DeclareOp(op func(list NodeListWriter)) {
	comp.ops = append(comp.ops, op)
}

func (comp *Compiler) SetOutput(op func(comp *Compiler, list NodeList)) {
	comp.out = op
}
