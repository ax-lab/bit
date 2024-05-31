package core

import "sync"

type Module struct {
	compiler *Compiler
	source   Source
	nodes    NodeList

	errorSync sync.Mutex
	errorSort bool
	errors    []error
}

func moduleNew(compiler *Compiler, source Source) *Module {
	if source == nil {
		panic("Module: invalid source")
	}

	module := &Module{
		compiler: compiler,
		source:   source,
	}
	module.nodes = NodeListNew(source.Span())
	module.nodes.checkValid()
	return module
}

func (mod *Module) Compiler() *Compiler {
	return mod.compiler
}

func (mod *Module) Nodes() NodeList {
	return mod.nodes
}

func (mod *Module) NewLexer() *Lexer {
	return mod.compiler.Lexer.Copy()
}

func (mod *Module) Error(err error) (stop bool) {
	mod.checkValid()
	if err == nil {
		return false
	}

	mod.errorSync.Lock()
	mod.errors = append(mod.errors, err)
	mod.errorSort = true
	mod.errorSync.Unlock()

	return mod.compiler.incrementErrorCount()
}

func (mod *Module) Errors() (out []error) {
	mod.checkValid()
	mod.errorSync.Lock()
	defer mod.errorSync.Unlock()
	if mod.errorSort {
		SortErrors(mod.errors)
		mod.errorSort = false
	}
	out = mod.errors

	return out
}

func (mod *Module) Compare(other *Module) int {
	return SourceCompare(mod.source, other.source)
}

func (mod *Module) checkValid() {
	if mod.compiler == nil {
		panic("Module: compiler is invalid")
	}
}
