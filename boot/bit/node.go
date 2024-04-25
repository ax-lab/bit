package bit

import (
	"slices"

	"axlab.dev/bit/boot/core"
)

type Node struct {
	value any
	span  core.Span
}

func (node Node) Value() any {
	return node.value
}

func (node Node) Span() core.Span {
	return node.span
}

type NodeList struct {
	data *nodeListData
}

func NodeListNew(module *Module, nodes ...Node) NodeList {
	data := nodeListDataNew(module, nodes)
	data.Init()
	return NodeList{data}
}

func (ls NodeList) Len() int {
	return len(ls.data.nodes)
}

func (ls NodeList) Span() core.Span {
	return ls.data.span
}

func (ls NodeList) Module() *Module {
	return ls.data.module
}

func (ls NodeList) Get(nth int) Node {
	return ls.data.nodes[nth]
}

func (ls NodeList) Nodes() []Node {
	return ls.data.nodes
}

func (ls NodeList) Slice(sta, end int) NodeList {
	data := ls.data.Slice(sta, end)
	return NodeList{data}
}

func (ls NodeList) Set(nth int, node Node) {
	ls.data.Set(nth, node)
}

func (ls NodeList) SetSpan(nth int, span core.Span) {
	node := ls.data.nodes[nth]
	node.span = span
	ls.data.Set(nth, node)
}

func (ls NodeList) SetValue(nth int, value any) {
	node := ls.data.nodes[nth]
	node.value = value
	ls.data.Set(nth, node)
}

func (ls NodeList) Push(value any, span core.Span) {
	ls.Append(Node{value, span})
}

func (ls NodeList) Append(nodes ...Node) {
	last := ls.data.Len()
	ls.data.Replace(last, last, nodes...)
}

func (ls NodeList) Insert(index int, nodes ...Node) {
	ls.data.Replace(index, index, nodes...)
}

func (ls NodeList) Remove(sta, end int) {
	ls.data.Replace(sta, end)
}

func (ls NodeList) Replace(sta, end int, nodes ...Node) {
	ls.data.Replace(sta, end, nodes...)
}

type nodeListData struct {
	module *Module
	nodes  []Node
	span   core.Span
}

func nodeListDataNew(module *Module, nodes []Node) *nodeListData {
	if module == nil {
		panic("NodeList: invalid module")
	}

	var span core.Span
	if len(nodes) > 0 {
		span = nodes[0].span.Merged(nodes[len(nodes)-1].span)
	} else {
		span = module.Source().Span().WithLen(0)
	}

	data := &nodeListData{module, nodes, span}
	return data
}

func (ls *nodeListData) Init() {
	for _, it := range ls.nodes {
		ls.includeNode(it)
	}
}

func (ls *nodeListData) Slice(sta, end int) *nodeListData {
	nodes := append([]Node{}, ls.nodes[sta:end]...)
	data := nodeListDataNew(ls.module, nodes)
	return data
}

func (ls *nodeListData) Len() int {
	return len(ls.nodes)
}

func (ls *nodeListData) Set(nth int, node Node) {
	ls.includeNode(node)
	ls.nodes[nth] = node
}

func (ls *nodeListData) Replace(sta, end int, nodes ...Node) {
	last := len(ls.nodes)
	if sta < 0 || end > last || end < sta {
		panic("NodeList: invalid splice range")
	}

	for _, it := range nodes {
		ls.includeNode(it)
	}
	ls.nodes = slices.Replace(ls.nodes, sta, end, nodes...)
}

func (ls *nodeListData) includeNode(node Node) {
	new := node.span
	if len(ls.nodes) == 0 && ls.span.Len() == 0 && new.Src() == ls.span.Src() {
		ls.span = new
	} else {
		ls.span.Merge(new)
	}
}
