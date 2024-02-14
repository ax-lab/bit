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
	Node   *Node
	Parent *Scope
	Names  *NameScope

	mutex sync.Mutex
	vars  map[varKey]*Variable
}

func newScope(node *Node) *Scope {
	out := &Scope{
		Node:   node,
		Parent: node.GetParentScope(),
	}

	if out.Parent != nil {
		out.Names = out.Parent.Names.NewChild()
	} else {
		out.Names = node.program.names.NewChild()
	}

	return out
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
	Type   Type

	value   Result
	escaped string
}

func (v *Variable) Value() Result {
	return v.value
}

func (v *Variable) SetValue(val Result) {
	v.value = val
}

func (v *Variable) String() string {
	return fmt.Sprintf("Var(%s@%s)", v.Name, v.Decl.Span().Location().String())
}

func (v *Variable) EncodedName() string {
	if v.escaped == "" {
		v.escaped = v.Scope.DeclareUnique(v.Name)
	}
	return v.escaped
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

func (scope *Scope) DeclareGlobal(name string, node *Node) {
	encoded := EncodeIdentifier(name)
	if !scope.Names.nameMap.DeclareGlobal(encoded) {
		if node == nil {
			node = scope.Node
		}
		node.AddError("the global name `%s` was already declared", name)
	}
}

func (scope *Scope) DeclareUnique(name string) string {
	name = EncodeIdentifier(name)
	return scope.Names.DeclareUnique(name)
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

func (code WithScope) Type() Type {
	return code.Inner.Type()
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
	name := ctx.NewName("scope")
	ctx.Body.Decl.Push("%s %s; // scope at %s", code.Inner.Type().CppType(), name, code.Inner.Span().String())

	block := CppContext{}
	block.NewBody(ctx)
	block.Names = code.Scope.Names

	for _, it := range code.Vars {
		block.Body.Decl.Push("%s %s; // %s @%s", it.Type.CppType(), it.EncodedName(), it.Name, it.Decl.Span().String())
	}
	code.Inner.OutputCpp(&block)
	block.Body.AppendTo(&ctx.Body.CppLines)

	ctx.Body.Push("%s = %s;", name, block.Expr.String())
	ctx.Expr.WriteString(name)
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
