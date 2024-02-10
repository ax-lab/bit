package bit

import (
	"fmt"
	"sync"
)

// Node values that can delimit a scope must implement this.
type HasScope interface {
	IsScope(node *Node) (is bool, sta, end int)
}

type Scope struct {
	Node *Node
	Sta  int
	End  int

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
}

func (v Variable) String() string {
	return fmt.Sprintf("Var(%s@%s)", v.Name, v.Decl.Span().Location().String())
}

func (scope *Scope) Declare(decl *Node, name string, offset int) *Variable {
	if offset < scope.Sta || scope.End <= offset {
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

func (node *Node) GetParentScope() *Scope {
	if node == nil || node.Parent() == nil {
		return nil
	}
	return node.Parent().GetScope()
}

func (node *Node) GetScope() *Scope {
	program := node.program
	program.scopeMutex.Lock()
	defer program.scopeMutex.Unlock()

	var (
		isScope  bool
		scope    *Scope
		sta, end int
	)

	if scope, ok := program.scopes[node]; ok {
		return scope
	}

	cur := node
	for cur != nil {
		if v, ok := cur.Value().(HasScope); ok {
			if isScope, sta, end = v.IsScope(cur); isScope {
				break
			}
		}
		cur = cur.Parent()
	}

	if isScope {
		scope = &Scope{
			Node: cur,
			Sta:  sta,
			End:  end,
		}
	}

	program.scopes[node] = scope

	if scope == nil {
		panic(fmt.Sprintf("scope resolution returned nil for node `%s`", node.Describe()))
	}

	return scope
}
