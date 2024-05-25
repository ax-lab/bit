package core

type Node struct {
	data *nodeData
}

type nodeData struct {
	span  Span
	value Value
}

func NodeNew(span Span, data Value) Node {
	if data == nil {
		panic("Node: data value cannot be nil")
	}
	if span.Src() == nil {
		panic("Node: invalid span")
	}
	node := &nodeData{span, data}
	return Node{node}
}

func (node Node) Span() Span {
	return node.data.span
}

func (node Node) Value() Value {
	return node.data.value
}

func (node Node) checkValid() {
	if node.data == nil {
		panic("invalid Node")
	}
}

type NodeList struct {
	data *nodeListData
}

type nodeListData struct {
	span  Span
	nodes []Node
}

func NodeListNew(span Span, nodes ...Node) NodeList {
	if span.Src() == nil {
		panic("NodeList: invalid span")
	}
	for _, it := range nodes {
		it.checkValid()
	}
	data := &nodeListData{span, nodes}
	return NodeList{data}
}

func (list NodeList) Get(idx int) Node {
	return list.data.nodes[idx]
}

func (list NodeList) Len() int {
	return len(list.data.nodes)
}

func (list NodeList) Span() Span {
	return list.data.span
}

func (list NodeList) checkValid() {
	if list.data == nil {
		panic("invalid NodeList")
	}
}

type NodeListWriter struct {
	NodeList
}

func (list NodeListWriter) Set(idx int, node Node) {
	list.data.nodes[idx] = node
}
