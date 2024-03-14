package boot

import (
	"fmt"
	"sync"
)

type Node struct {
	inner *nodeInner
}

func (node Node) Span() Span {
	return node.inner.span
}

func (node Node) Value() any {
	return node.inner.value
}

type nodeMap struct {
	mutex    sync.Mutex
	allNodes []Node
}

func (nm *nodeMap) NewNode(value any, span Span) Node {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	data := &nodeInner{}
	node := Node{data}

	nm.allNodes = append(nm.allNodes, node)
	return node
}

func (nm *nodeMap) CheckDone() error {
	var pending []Node
	for _, it := range nm.allNodes {
		if !it.inner.done {
			it.inner.done = true
			pending = append(pending, it)
		}
	}

	if count := len(pending); count > 0 {
		return fmt.Errorf("there are %d pending nodes", count)
	}

	return nil
}

type nodeInner struct {
	value any
	span  Span
	done  bool
}
