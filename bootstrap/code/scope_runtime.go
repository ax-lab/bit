package code

import (
	"fmt"
	"slices"
	"strings"
)

type varSlot struct {
	scope    *Scope
	variable *Variable
	index    int
}

func (scope *Scope) Init(rt *Runtime) {
	if new := len(scope.slots); new > 0 {
		newLen := len(rt.Stack) + new
		rt.Stack = slices.Grow(rt.Stack, new)
		rt.Stack = rt.Stack[:newLen]
	}
}

func (scope *Scope) Drop(rt *Runtime) {
	newLen := len(rt.Stack) - len(scope.slots)
	rt.Stack = rt.Stack[:newLen]
}

func (scope *Scope) OutputCpp(ctx *CppContext) {
	for _, it := range scope.slots {
		ctx.Body.Push("%s %s;", it.variable.Type().CppType(), it.variable.OutputName())
	}
}

func (scope *Scope) String() string {
	out := strings.Builder{}
	for n, it := range scope.slots {
		if n > 0 {
			out.WriteString("\n")
		}
		out.WriteString(fmt.Sprintf("var %s: %s", it.variable.Name(), it.variable.Type().String()))
	}
	return out.String()
}
