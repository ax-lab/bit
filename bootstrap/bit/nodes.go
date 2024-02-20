package bit

import (
	"fmt"
	"sort"
	"strings"
	"sync/atomic"

	"axlab.dev/bit/code"
	"axlab.dev/bit/common"
)

type Key interface {
	IsEqual(val any) bool
	Repr(oneline bool) string
}

type Value interface {
	Repr(oneline bool) string
	Bind(node *Node)
}

type Node struct {
	program *Program
	value   Value
	span    Span
	done    atomic.Bool
	id      int

	nodes  []*Node
	parent *Node
	index  int
	scope  *code.Scope
}

var idCounter atomic.Int32

func (program *Program) NewNode(value Value, span Span) *Node {
	node := &Node{
		program: program,
		value:   value,
		span:    span,
		id:      int(idCounter.Add(1)),
	}

	if v, ok := value.(HasScope); ok {
		if v.IsScope(node) {
			parent := node.GetParentScope()
			if parent == nil {
				parent = program.scope
			}
			span := node.Span()
			node.scope = parent.NewChild(span.Sta(), span.End(), node)
		}
	}

	program.allNodes = append(program.allNodes, node)
	node.value.Bind(node)

	return node
}

func (node *Node) WithChildren(nodes ...*Node) *Node {
	node.AddChildren(nodes...)
	return node
}

func (node *Node) Program() *Program {
	return node.program
}

func (node *Node) Bind(key Key) {
	node.program.BindNodes(key, node)
}

func (node *Node) DeclareStatic(key Key, binding Binding) {
	node.program.bindings.BindStatic(key, node.span.Source(), binding)
}

func (node *Node) Declare(key Key, binding Binding) {
	node.program.bindings.Bind(key, node.span, binding)
}

func (node *Node) DeclareAt(key Key, sta, end int, binding Binding) {
	node.program.bindings.BindAt(key, sta, end, node.Span().Source(), binding)
}

func (node *Node) Describe() string {
	return fmt.Sprintf("%s at %s", node.value.Repr(true), node.span.String())
}

func (node *Node) String() string {
	return fmt.Sprintf("Node(%s#%d @%s)", node.value.Repr(true), node.id, node.span.String())
}

func (node *Node) Dump(full bool) string {
	header := fmt.Sprintf("#%d = ", node.id)
	out := strings.Builder{}
	out.WriteString(header)
	out.WriteString(common.Indented(node.value.Repr(false)))
	if diff := 30 - out.Len(); diff > 0 {
		out.WriteString(strings.Repeat(" ", diff))
	} else {
		out.WriteString("  ")
	}
	out.WriteString("@ ")
	out.WriteString(node.span.String())
	hasTxt := false
	if txt := node.span.DisplayText(120); txt != "" {
		hasTxt = true
		if len(txt) <= 20 {
			if diff := 60 - out.Len(); diff > 0 {
				out.WriteString(strings.Repeat(" ", diff))
			} else {
				out.WriteString("  ")
			}
			out.WriteString(" # ")
			out.WriteString(txt)
		} else {
			indent := strings.Repeat(".", len(header)-3)
			out.WriteString(fmt.Sprintf("\n[%s] %s", indent, txt))
		}
	}
	if len(node.nodes) > 0 && full {
		if hasTxt {
			out.WriteString("\n")
		}
		out.WriteString("{")
		for n, it := range node.nodes {
			out.WriteString(fmt.Sprintf("\n\t[%03d] ", n))
			out.WriteString(common.Indented(it.Dump(true)))
		}
		out.WriteString("\n}")
	}
	return out.String()
}

func (node *Node) AddError(msg string, args ...any) {
	err := node.span.CreateError(msg, args...)
	node.program.HandleError(err)
}

func (node *Node) Id() int {
	return node.id
}

func (node *Node) Indent() int {
	return node.Span().Indent()
}

func (node *Node) Value() Value {
	return node.value
}

func (node *Node) Span() Span {
	return node.span
}

func (node *Node) Offset() int {
	return node.span.Sta()
}

func (node *Node) Len() int {
	return len(node.nodes)
}

func (node *Node) Text() string {
	return node.span.Text()
}

func (node *Node) Get(index int) *Node {
	if index < len(node.nodes) {
		return node.nodes[index]
	}
	return nil
}

func (node *Node) Nodes() []*Node {
	return node.nodes
}

func (node *Node) Index() int {
	return node.index
}

func (node *Node) Parent() *Node {
	return node.parent
}

func (node *Node) IsDone() bool {
	return node.done.Load()
}

func (node *Node) FlagDone() {
	node.done.Store(true)
}

func (node *Node) Undo() {
	/*
		TODO: correct the behavior for node undoing with same-key bindings

		- Equal or more specific bindings will override the pre-existing one.
		- Even if the previous binding has evaluation precedence, it must not
		  pick the nodes bound to the more specific binding.
		- However, if the more specific binding undoes any node, those should
		  become available to the previous bindings.
		- This should work regardless of the evaluation order of the bindings.
	*/
	node.done.Store(false)
}

func (node *Node) Next() *Node {
	if node.parent != nil {
		nodes := node.parent.nodes
		index := node.index + 1
		if index < len(nodes) {
			return nodes[index]
		}
	}
	return nil
}

func (node *Node) Prev() *Node {
	if node.parent != nil {
		nodes := node.parent.nodes
		index := node.index - 1
		if index >= 0 && index < len(nodes) {
			return nodes[index]
		}
	}
	return nil
}

func (node *Node) Head() *Node {
	return node.nodes[0]
}

func (node *Node) Last() *Node {
	if l := len(node.nodes); l > 0 {
		return node.nodes[l-1]
	}
	return nil
}

func (node *Node) Compare(other *Node) int {
	if node == other {
		return 0
	}

	if cmp := node.span.Compare(other.span); cmp != 0 {
		return cmp
	}

	return 0
}

func (node *Node) AddChildren(nodes ...*Node) {
	node.InsertNodes(len(node.nodes), nodes...)
}

func (node *Node) RemoveRange(sta, end *Node) []*Node {
	if sta.parent != node || end.parent != node {
		panic("RemoveRange: range nodes are not children of the current node")
	}

	s, e := sta.Index(), end.Index()
	if e < s {
		s, e = e, s
	}

	return node.RemoveNodes(s, e+1)
}

func (node *Node) RemoveNodes(sta, end int) []*Node {
	nodes := node.nodes

	removed := nodes[sta:end]
	for _, it := range removed {
		// keep the index intact in case operators are still referencing it
		it.parent = nil
	}

	if len(removed) == 0 {
		return nil
	}

	node.nodes = nil
	node.nodes = append(node.nodes, nodes[:sta]...)
	node.nodes = append(node.nodes, nodes[end:]...)
	for n, it := range node.nodes {
		it.parent = node
		it.index = n
	}

	return removed
}

func (node *Node) InsertNodes(at int, newNodes ...*Node) {
	if len(newNodes) == 0 {
		return
	}

	for _, it := range newNodes {
		if it.parent != nil {
			panic(fmt.Sprintf("Node `%s` already has a parent", it.Describe()))
		}
	}

	nodes := node.nodes
	node.nodes = nil
	node.nodes = append(node.nodes, nodes[:at]...)
	node.nodes = append(node.nodes, newNodes...)
	node.nodes = append(node.nodes, nodes[at:]...)
	for n, it := range node.nodes[at:] {
		it.parent = node
		it.index = n + at
	}

	// we can safely modify the node span end since it does not change
	// its offset or location
	if last := node.nodes[len(node.nodes)-1]; last != nil {
		span := last.span
		if span.Source() == node.span.Source() && span.End() > node.span.End() {
			node.span.SetEnd(span.End())
		}
	}
}

func (node *Node) Remove() {
	if node.parent != nil {
		node.parent.RemoveNodes(node.index, node.index+1)
	}
}

func (node *Node) ReplaceWithValue(value Value) *Node {
	newNode := node.program.NewNode(value, node.span)
	node.Replace(newNode)
	return newNode
}

func (node *Node) Replace(nodes ...*Node) {
	if par := node.parent; par != nil {
		index := node.index
		par.RemoveNodes(index, index+1)
		par.InsertNodes(index, nodes...)
	}
}

func SortNodes(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		a, b := nodes[i], nodes[j]
		return a.Compare(b) < 0
	})
}

func DebugNodes(msg string, nodes ...*Node) {
	out := strings.Builder{}
	out.WriteString(msg)
	for n, it := range nodes {
		if n == 0 {
			out.WriteString("\n\n")
		}
		out.WriteString(common.Indent(it.Dump(false)) + "\n")
	}

	if len(nodes) == 0 {
		out.WriteString("  (no nodes)\n")
	}
	fmt.Println(out.String())
}

// Node values that can delimit a scope must implement this.
type HasScope interface {
	IsScope(node *Node) bool
}

func (node *Node) GetParentScope() *code.Scope {
	if node == nil || node.Parent() == nil {
		return nil
	}
	return node.Parent().GetScope()
}

func (node *Node) GetScope() *code.Scope {
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

type HasOutput interface {
	Type(node *Node) code.Type

	// TODO: review output model
	Output(ctx *code.OutputContext, node *Node, ans *code.Variable)
}

func (node *Node) Type() code.Type {
	if node != nil {
		if v, ok := node.Value().(HasOutput); ok {
			return v.Type(node)
		}
		return code.InvalidType()
	}
	return code.VoidType()
}

func (node *Node) Output(ctx *code.OutputContext, ans *code.Variable) {
	if !ctx.Valid() {
		return
	}
	if v, ok := node.value.(HasOutput); ok {
		if node.scope != nil {
			inner := ctx.NewScope(node.scope)
			v.Output(inner, node, ans)
			ctx.Output(inner.Block())
		} else {
			v.Output(ctx, node, ans)
		}
	} else {
		err := node.span.CreateError("cannot output node `%s`", node.Describe())
		ctx.Error(err)
	}
}

func (node *Node) CreateError(msg string, args ...any) error {
	return node.span.CreateError(msg, args)
}

func (node *Node) OutputError(ctx *code.OutputContext, msg string, args ...any) {
	err := node.CreateError(msg, args...)
	ctx.Error(err)
}

func (node *Node) CheckEmpty(ctx *code.OutputContext) bool {
	return node.CheckArity(ctx, 0)
}

func (node *Node) CheckArity(ctx *code.OutputContext, n int) bool {
	if node.Len() != n {
		node.OutputError(ctx, "node `%s` should have arity %d", node.Describe(), n)
		return false
	}
	return true
}

func (node *Node) CheckRange(ctx *code.OutputContext, a, b int) bool {
	if len := node.Len(); len < a || b < len {
		node.OutputError(ctx, "node `%s` should have arity between %d and %d", node.Describe(), a, b)
		return false
	}
	return true
}

func (node *Node) OutputChildren(ctx *code.OutputContext, ans *code.Variable) {
	nodes := node.Nodes()
	for n, it := range nodes {
		if n == len(nodes)-1 {
			it.Output(ctx, ans)
		} else {
			it.Output(ctx, nil)
		}
		if !ctx.Valid() {
			break
		}
	}
}

// TODO: figure out allowEmpty and void values
func (node *Node) OutputChild(ctx *code.OutputContext, ans *code.Variable, allowEmpty bool) {
	list := node.Nodes()
	switch len(list) {
	case 0:
		if !allowEmpty {
			node.OutputError(ctx, "node `%s` cannot be empty", node.Describe())
		}
	case 1:
		list[0].Output(ctx, ans)
	default:
		node.OutputError(ctx, "node `%s` cannot have multiple children", node.Describe())
	}
}
