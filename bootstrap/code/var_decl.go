package code

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

type Decl struct {
	scope *Scope
	mutex sync.Mutex
	slots []*varSlot

	/*
		TODO: this is a hack to allow indexing variables from upper scopes

		Since nested scopes are static, we can probably just compute an absolute
		offset for each variable from the stack. This would be non-reentrant,
		but functions should have their own stack anyway.
	*/
	rtOffset int
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
	index := len(decl.slots)
	v.slot = &varSlot{decl: decl, variable: v, index: index}
	decl.slots = append(decl.slots, v.slot)
}

func (decl *Decl) Init(rt *Runtime) {
	if new := len(decl.slots); new > 0 {
		decl.rtOffset = len(rt.Stack)
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

func (decl *Decl) String() string {
	out := strings.Builder{}
	for n, it := range decl.slots {
		if n > 0 {
			out.WriteString("\n")
		}
		out.WriteString(fmt.Sprintf("var %s: %s", it.variable.Name(), it.variable.Type().String()))
	}
	return out.String()
}

type varSlot struct {
	decl     *Decl
	variable *Variable
	index    int
}
