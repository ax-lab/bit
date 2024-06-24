package code

import (
	"fmt"
	"math"
	"sync"
)

type VarId struct {
	frame uint32
	index uint32
}

type Scope struct {
	root   *Scope
	parent *Scope

	varSync  sync.Mutex
	varCount uint32
	varMap   map[Id]uint32
}

func (scope *Scope) NewChild() *Scope {
	out := &Scope{
		root:   scope.getRoot(),
		parent: scope,
	}
	return out
}

func (scope *Scope) getRoot() *Scope {
	if scope.root != nil {
		return scope.root
	}
	return scope
}

func (scope *Scope) Declare(v Var) (out VarId, err error) {
	scope.varSync.Lock()
	defer scope.varSync.Unlock()

	if scope.varCount == math.MaxUint32 {
		return out, fmt.Errorf("variable count overflow in scope")
	}

	index := scope.varCount
	scope.varCount++

	if scope.varMap == nil {
		scope.varMap = make(map[Id]uint32)
	}

	scope.varMap[v.Name] = index
	out = VarId{frame: 0, index: index}
	return out, nil
}

func (scope *Scope) Resolve(v Var) (out VarId, err error) {
	current, frame := scope, uint32(0)
	for current != nil {
		if index, found := current.tryResolve(v); found {
			out = VarId{frame: frame, index: index}
			return out, nil
		}
		current = current.parent
		frame++
	}

	return out, fmt.Errorf("variable `%s` not in the scope", v.Name)
}

func (scope *Scope) tryResolve(v Var) (index uint32, found bool) {
	scope.varSync.Lock()
	defer scope.varSync.Unlock()
	index, found = scope.varMap[v.Name]
	return
}
