package boot

type Node struct {
	inner *nodeInner
}

func (node Node) Span() Span {
	return node.inner.span
}

func (node Node) Value() any {
	return node.inner.value
}

type nodeMap struct{}

func (nm *nodeMap) NewNode(value any, span Span) Node {
	data := &nodeInner{}
	return Node{data}
}

type nodeInner struct {
	value any
	span  Span
}
