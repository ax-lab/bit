package core

import (
	"io"
	"os"
)

type Runtime struct {
	compiler *Compiler

	ops []OpFunc
	out OpFunc

	stdErr io.Writer
	stdOut io.Writer
}

func runtimeNew(compiler *Compiler) *Runtime {
	rt := &Runtime{
		compiler: compiler,
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

func (rt *Runtime) Run() (out Value, err error) {
	modules, output := rt.compiler.GetOutput()
	for n := range modules {
		expr := output[n]
		for _, it := range expr {
			out, err = it.Eval(rt)
			if err != nil {
				return
			}
		}
	}
	return
}
