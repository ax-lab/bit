package core

import (
	"io"
	"os"
	"slices"
	"sync"
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
}

func (compiler *Compiler) AddSource(name string) {
	compiler.sync.Lock()
	defer compiler.sync.Unlock()
	compiler.sources = append(compiler.sources, name)
}

func (compiler *Compiler) CreateRuntime() *Runtime {
	compiler.sync.Lock()
	var (
		sources = slices.Clone(compiler.sources)
		ops     = slices.Clone(compiler.ops)
		out     = compiler.out
	)
	compiler.sync.Unlock()

	rt := runtimeNew(compiler)
	rt.ops = ops
	rt.out = out
	rt.sources = sources

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
