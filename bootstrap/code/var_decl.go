package code

import (
	"fmt"
	"slices"
	"sync"
)

type Decl struct {
	scope *Scope
	mutex sync.Mutex
	slots []*varSlot
}

func NewDecl(scope *Scope) *Decl {
	return &Decl{scope: scope}
}

func (decl *Decl) Len() int {
	decl.mutex.Lock()
	defer decl.mutex.Unlock()
	return len(decl.slots)
}

func (decl *Decl) Add(v *Variable) {
	if v.slot != nil {
		panic(fmt.Sprintf("variable `%s` already bound to a declaration", v.Name()))
	}

	decl.mutex.Lock()
	defer decl.mutex.Unlock()
	index := decl.Len()
	v.slot = &varSlot{decl: decl, variable: v, index: index}
	decl.slots = append(decl.slots, v.slot)
}

func (decl *Decl) Init(rt *Runtime) {
	if new := len(decl.slots); new > 0 {
		newLen := len(rt.Stack) + new
		rt.Stack = slices.Grow(rt.Stack, new)
		rt.Stack = rt.Stack[:newLen]
	}
}

func (decl *Decl) Drop(rt *Runtime) {
	newLen := len(rt.Stack) - len(decl.slots)
	rt.Stack = rt.Stack[:newLen]
}

func (decl *Decl) OutputCpp(ctx *CppContext) {
	for _, it := range decl.slots {
		ctx.Body.Push("%s %s;", it.variable.Type().CppType(), it.variable.OutputName())
	}
}

type varSlot struct {
	decl     *Decl
	variable *Variable
	index    int
}
