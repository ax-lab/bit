package bit

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"axlab.dev/bit/common"
)

// Node values that can delimit a scope must implement this.
type HasScope interface {
	IsScope(node *Node) bool
}

type Scope struct {
	Node *Node

	mutex sync.Mutex
	vars  map[varKey]*Variable
}

type varKey struct {
	Name   string
	Offset int
}

type Variable struct {
	Scope  *Scope
	Decl   *Node
	Name   string
	Offset int

	value Result
}

func (v *Variable) String() string {
	return fmt.Sprintf("Var(%s@%s)", v.Name, v.Decl.Span().Location().String())
}

func (scope *Scope) Len() int {
	return len(scope.vars)
}

func (scope *Scope) Sta() int {
	return scope.Node.Span().Sta()
}

func (scope *Scope) End() int {
	return scope.Node.Span().End()
}

func (scope *Scope) Declare(decl *Node, name string, offset int) *Variable {
	if offset < scope.Sta() || scope.End() <= offset {
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
		Scope:  scope,
		Decl:   decl,
		Name:   name,
		Offset: offset,
	}
	scope.vars[key] = v
	return v
}

type WithScope struct {
	Scope *Scope
	Inner Code
	Vars  []*Variable
}

func (code *WithScope) initVars() {
	if len(code.Vars) > 0 {
		return
	}

	var keys []varKey
	scope := code.Scope
	for k := range scope.vars {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(a, b int) bool {
		ka, kb := keys[a], keys[b]
		if ka.Offset != kb.Offset {
			return ka.Offset < kb.Offset
		}

		if pa, pb := scope.vars[ka].Decl.Offset(), scope.vars[kb].Decl.Offset(); pa != pb {
			return pa < pb
		}
		return ka.Name < kb.Name
	})

	for _, k := range keys {
		code.Vars = append(code.Vars, scope.vars[k])
	}
}

func (code WithScope) Eval(rt *RuntimeContext) {
	for _, it := range code.Vars {
		it.value = nil
	}

	rt.Result = rt.Eval(code.Inner)

	for _, it := range code.Vars {
		it.value = nil
	}
}

func (code WithScope) OutputCpp(ctx *CppContext, node *Node) {
	out := ctx.OutputFilePrefix
	out.NewLine()
	out.Write("#error Scope evaluation not implemented\n")
}

func (code WithScope) Repr(oneline bool) string {
	if oneline {
		return fmt.Sprintf("WithScope { %s }", code.Inner.Expr.Repr(true))
	}

	scope := code.Scope

	out := strings.Builder{}
	out.WriteString("WithScope {\n")
	out.WriteString(fmt.Sprintf("\t# SRC: %s\n", scope.Node.Describe()))

	for _, v := range code.Vars {
		out.WriteString(fmt.Sprintf("\t# VAR: %s at %s with offset %d\n", v.Name, v.Decl.Span().String(), v.Offset))
	}

	out.WriteString(common.Indent(code.Inner.Repr(false)))
	out.WriteString("\n}")
	return out.String()
}

func (node *Node) getOwnScope() *Scope {
	return node.scope
}

func (node *Node) GetParentScope() *Scope {
	if node == nil || node.Parent() == nil {
		return nil
	}
	return node.Parent().GetScope()
}

func (node *Node) GetScope() *Scope {
	cur, scope := node, node.scope
	for scope == nil && cur.Parent() != nil {
		cur = cur.Parent()
		scope = cur.scope
	}

	if scope == nil {
		panic(fmt.Sprintf("scope resolution returned nil for node `%s`", node.Describe()))
	}

	return scope
}
