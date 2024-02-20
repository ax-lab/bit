package code

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

type Scope struct {
	Source any

	parent *Scope
	root   *rootScopeData
	sta    int
	end    int

	mutex sync.Mutex
	vars  map[varKey]*Variable
	uniq  []*Variable

	children []*Scope

	slots      []*varSlot
	slotOffset int
}

func NewScope(source any) *Scope {
	out := &Scope{
		Source: source,
		parent: nil,
		root:   &rootScopeData{},
		sta:    0,
		end:    math.MaxInt,
	}
	out.root.Scope = out
	return out
}

func (scope *Scope) NewChild(sta, end int, source any) *Scope {
	if sta >= end || sta < scope.sta || scope.end < end {
		panic("invalid child scope range")
	}

	out := &Scope{
		Source: source,
		parent: scope,
		root:   scope.root,
		sta:    sta,
		end:    end,
	}

	scope.mutex.Lock()
	defer scope.mutex.Unlock()
	scope.children = append(scope.children, out)

	return out
}

func (scope *Scope) ProcessNames() {
	if scope.root.compiled.CompareAndSwap(false, true) {
		scope.root.ProcessNames()
	}
}

func (scope *Scope) BindVars() {
	scope.root.ProcessSlots()
}

func (scope *Scope) Parent() *Scope {
	return scope.parent
}

func (scope *Scope) Root() *Scope {
	return scope.root.Scope
}

func (scope *Scope) Sta() int {
	return scope.sta
}

func (scope *Scope) End() int {
	return scope.end
}

func (scope *Scope) Len() int {
	return len(scope.vars) + len(scope.uniq)
}

func (scope *Scope) DeclareUnique(baseName string, typ Type, source any) *Variable {
	scope.mutex.Lock()
	defer scope.mutex.Unlock()

	v := &Variable{
		scope:  scope,
		Source: source,
		name:   baseName,
		typ:    typ,
		offset: 0,
	}
	scope.uniq = append(scope.uniq, v)

	return v
}

func (scope *Scope) Declare(name string, offset int) *Variable {
	if offset < scope.sta || scope.end <= offset {
		panic("invalid offset for scope variable")
	}

	scope.mutex.Lock()
	defer scope.mutex.Unlock()

	key := varKey{name, offset}

	if _, ok := scope.vars[key]; ok {
		panic("declaring duplicated variable in scope")
	}

	if scope.vars == nil {
		scope.vars = make(map[varKey]*Variable)
	}

	v := &Variable{
		scope:  scope,
		name:   name,
		offset: offset,
	}
	scope.vars[key] = v
	return v
}

type varKey struct {
	Name   string
	Offset int
}

type rootScopeData struct {
	Scope    *Scope
	nameMap  map[string]int
	compiled atomic.Bool
}

func (root *rootScopeData) ProcessSlots() {
	root.Scope.processSlots()
}

func (scope *Scope) processSlots() {
	if scope.parent != nil {
		scope.slotOffset = scope.parent.slotOffset + len(scope.parent.slots)
	}

	for _, it := range scope.vars {
		index := len(scope.slots)
		it.slot = &varSlot{scope, it, index}
		scope.slots = append(scope.slots, it.slot)

	}

	for _, it := range scope.uniq {
		index := len(scope.slots)
		it.slot = &varSlot{scope, it, index}
		scope.slots = append(scope.slots, it.slot)
	}

	for _, it := range scope.children {
		it.processSlots()
	}
}

func (root *rootScopeData) ProcessNames() {
	root.nameMap = make(map[string]int)
	root.Scope.processVariables()
	root.Scope.processUniques()
}

func (root *rootScopeData) processVar(v *Variable) {
	v.counter = root.nameMap[v.name]
	root.nameMap[v.name] = v.counter + 1
	if v.counter > 0 {
		v.processed = EncodeIdentifier(fmt.Sprintf("%s%d", v.name, v.counter))
	} else {
		v.processed = EncodeIdentifier(v.name)
	}
}

func (scope *Scope) processVariables() {
	for _, v := range scope.vars {
		scope.root.processVar(v)
	}
	for _, it := range scope.children {
		it.processVariables()
	}
}

func (scope *Scope) processUniques() {
	for _, v := range scope.uniq {
		scope.root.processVar(v)
	}
	for _, it := range scope.children {
		it.processUniques()
	}
}
