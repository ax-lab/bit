package boot

import (
	"slices"
	"sync"
)

type nodeMap struct {
	mutex    sync.Mutex
	allNodes []Node
	pending  nodeMapByType
}

func (nm *nodeMap) NewNode(value Value, span Span) Node {
	if value == nil {
		panic("Node: value cannot be nil")
	}

	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	data := &nodeInner{
		value: value,
		span:  span,
	}

	node := Node{data}
	nm.allNodes = append(nm.allNodes, node)

	if nm.pending == nil {
		nm.pending = make(nodeMapByType)
	}
	nm.pending.Add(node)

	return node
}

type nodeMapByType map[Type]nodeMapBySource

func (mTyp nodeMapByType) Add(node Node) {
	key := node.Value().Type()
	mSrc, ok := mTyp[key]
	if !ok {
		mSrc = make(nodeMapBySource)
		mTyp[key] = mSrc
	}
	mSrc.Add(node)
}

type nodeMapBySource map[*Source]*nodeList

func (mSrc nodeMapBySource) Add(node Node) {
	key := node.Span().Src()
	list, ok := mSrc[key]
	if !ok {
		list = &nodeList{}
		mSrc[key] = list
	}
	list.Add(node)
}

type nodeList struct {
	sorted bool
	nodes  []Node
}

func (ls *nodeList) Add(node Node) {
	cnt := len(ls.nodes)
	ls.sorted = cnt == 0 || (ls.sorted && node.Offset() >= ls.nodes[cnt-1].Offset())
	ls.nodes = append(ls.nodes, node)
}

func (ls *nodeList) Sort() {
	if !ls.sorted {
		ls.sorted = true
		slices.SortStableFunc(ls.nodes, func(a, b Node) int {
			return a.Cmp(b)
		})
	}
}
