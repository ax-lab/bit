package bit

import (
	"fmt"
	"sort"
	"strings"
	"sync/atomic"

	"axlab.dev/bit/text"
)

type Key interface {
	IsEqual(other Key) bool
	String() string
}

type Value interface {
	String() string
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
}

var idCounter atomic.Int32

func (program *Program) NewNode(value Value, span Span) *Node {
	node := &Node{
		program: program,
		value:   value,
		span:    span,
		id:      int(idCounter.Add(1)),
	}

	program.allNodes = append(program.allNodes, node)
	node.value.Bind(node)

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

func (node *Node) String() string {
	return fmt.Sprintf("Node(%s#%d @%s)", node.value.String(), node.id, node.span.String())
}

func (node *Node) Dump() string {
	header := fmt.Sprintf("#%d = ", node.id)
	out := strings.Builder{}
	out.WriteString(header)
	out.WriteString(node.value.String())
	if diff := 30 - out.Len(); diff > 0 {
		out.WriteString(strings.Repeat(" ", diff))
	} else {
		out.WriteString("  ")
	}
	out.WriteString("@ ")
	out.WriteString(node.span.String())
	if txt := node.span.DisplayText(120); txt != "" {
		if len(txt) <= 20 {
			if diff := 60 - out.Len(); diff > 0 {
				out.WriteString(strings.Repeat(" ", diff))
			} else {
				out.WriteString("  ")
			}
			out.WriteString("# ")
			out.WriteString(txt)
			out.WriteString(" ")
		} else {
			indent := strings.Repeat(" ", len(header))
			out.WriteString(fmt.Sprintf("\n%s> ", indent))
			out.WriteString(txt)
			out.WriteString("\n")
		}
	}
	if len(node.nodes) > 0 {
		out.WriteString("{")
		for n, it := range node.nodes {
			out.WriteString(fmt.Sprintf("\n\t[%03d] ", n))
			out.WriteString(text.Indented(it.Dump()))
		}
		out.WriteString("\n}")
	}
	return out.String()
}

func (node *Node) AddError(msg string, args ...any) {
	err := node.span.CreateError(msg, args)
	node.program.HandleError(err)
}

func (node *Node) Id() int {
	return node.id
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

func (node *Node) Nodes() []*Node {
	return node.nodes
}

func (node *Node) Index() int {
	return node.index
}

func (node *Node) Parent() *Node {
	return node.parent
}

func (node *Node) Done() bool {
	return node.done.Load()
}

func (node *Node) Undo() {
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
		if index > 0 && index < len(nodes) {
			return nodes[index]
		}
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

func (node *Node) RemoveNodes(sta, end int) []*Node {
	nodes := node.nodes

	removed := nodes[sta:end]
	for _, it := range removed {
		it.index = 0
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

	nodes := node.nodes
	node.nodes = nil
	node.nodes = append(node.nodes, nodes[:at]...)
	node.nodes = append(node.nodes, newNodes...)
	node.nodes = append(node.nodes, nodes[at:]...)
	for n, it := range node.nodes[at:] {
		it.parent = node
		it.index = n + at
	}
}

func (node *Node) Remove() {
	if node.parent != nil {
		node.parent.RemoveNodes(node.index, node.index+1)
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
	for _, it := range nodes {
		out.WriteString(text.Indented("\n\n"+it.Dump()) + "\n")
	}

	if len(nodes) == 0 {
		out.WriteString("  (no nodes)\n")
	}
	fmt.Println(out.String())
}
