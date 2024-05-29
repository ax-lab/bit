package core

import "sync"

type Module struct {
	runtime *Runtime
	source  Source
	nodes   NodeList

	errorSync sync.Mutex
	errorSort bool
	errors    []error
}

func moduleNew(runtime *Runtime, source Source) *Module {
	if source == nil {
		panic("Module: invalid source")
	}

	module := &Module{
		runtime: runtime,
		source:  source,
	}
	module.nodes = NodeListNew(source.Span())
	module.nodes.checkValid()
	return module
}

func (mod *Module) Runtime() *Runtime {
	return mod.runtime
}

func (mod *Module) Nodes() NodeList {
	return mod.nodes
}

func (mod *Module) NewLexer() *Lexer {
	return mod.runtime.lexer.Copy()
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

	return mod.runtime.incrementErrorCount()
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

func (mod *Module) checkValid() {
	if mod.runtime == nil {
		panic("Module: runtime is invalid")
	}
}
