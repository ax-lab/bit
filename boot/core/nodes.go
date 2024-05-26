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

func (node Node) Valid() bool {
	return node.data != nil
}

func (node Node) Compiler() *Compiler {
	node.checkValid()
	return node.data.span.Compiler()
}

func (node Node) Span() Span {
	node.checkValid()
	return node.data.span
}

func (node Node) Value() Value {
	node.checkValid()
	return node.data.value
}

func (node Node) checkValid() {
	if node.data == nil {
		panic("Node is invalid")
	}
}

type NodeList struct {
	data *nodeListData
}

type nodeListData struct {
	span   Span
	nodes  []Node
	errors []error
}

func NodeListNew(span Span, nodes ...Node) NodeList {
	if span.Src() == nil {
		panic("NodeList with invalid span")
	}
	data := &nodeListData{span: span}
	list := NodeList{data}
	list.Push(nodes...)
	return list
}

func (list NodeList) Valid() bool {
	return list.data != nil
}

func (list NodeList) Compiler() *Compiler {
	list.checkValid()
	return list.data.span.Compiler()
}

func (list NodeList) Nodes() []Node {
	list.checkValid()
	return list.data.nodes
}

func (list NodeList) Get(idx int) Node {
	list.checkValid()
	return list.data.nodes[idx]
}

func (list NodeList) Empty() bool {
	list.checkValid()
	return list.Len() == 0 && len(list.data.errors) == 0
}

func (list NodeList) Len() int {
	list.checkValid()
	return len(list.data.nodes)
}

func (list NodeList) Span() Span {
	list.checkValid()
	return list.data.span
}

func (list NodeList) Errors() []error {
	list.checkValid()
	return list.data.errors
}

func (list NodeList) checkValid() {
	if list.data == nil {
		panic("NodeList is invalid")
	}
}

//----------------------------------------------------------------------------//
// Writer methods
//----------------------------------------------------------------------------//

func (list NodeList) Set(idx int, node Node) {
	list.checkValid()
	list.data.nodes[idx] = node
}

func (list NodeList) Push(nodes ...Node) {
	list.checkValid()
	span := list.data.span
	for _, it := range nodes {
		it.checkValid()
		if it.Compiler() != span.Compiler() {
			panic("NodeList node has a different compiler")
		}
	}
	list.data.nodes = append(list.data.nodes, nodes...)
}

func (list NodeList) PushError(err error) (stop bool) {
	list.checkValid()
	if err == nil {
		return false
	}

	list.data.errors = append(list.data.errors, err)
	return list.Compiler().incrementErrorCount()
}
