package code

import (
	"sync"

	"axlab.dev/bit/base"
)

type Program struct {
	Errors base.ErrorSet

	types TypeSet

	codeSync sync.Mutex
	codeList []Expr

	scope Scope
}

func (program *Program) Types() *TypeSet {
	program.types.source.CompareAndSwap(nil, program)
	return &program.types
}

func (program *Program) Append(code ...Expr) {
	program.codeSync.Lock()
	defer program.codeSync.Unlock()
	program.codeList = append(program.codeList, code...)
}

func (program *Program) HasErrors() bool {
	return program.Errors.Len() > 0
}
