package boot

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Node struct {
	inner *nodeInner
}

func (node Node) Span() Span {
	return node.inner.span
}

func (node Node) Value() Value {
	return node.inner.value
}

type nodeMap struct {
	mutex    sync.Mutex
	allNodes []Node
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
	return node
}

func (nm *nodeMap) CheckDone() error {
	pending := make(map[Type][]Node)
	count := 0
	for _, it := range nm.allNodes {
		if !it.inner.done {
			it.inner.done = true
			key := it.Value().Type()
			pending[key] = append(pending[key], it)
			count++
		}
	}

	if count > 0 {
		out := strings.Builder{}
		if count == 1 {
			out.WriteString("there is one pending node:\n")
		} else {
			out.WriteString(fmt.Sprintf("there are %d pending nodes:\n", count))
		}

		maxPer := max(count/len(pending), 1)
		for key, list := range pending {
			for n, it := range list {
				if n == maxPer {
					break
				}
				out.WriteString("\n\t--> ")
				out.WriteString(it.Value().Repr())
				out.WriteString("\n\t    ↳ at ")
				out.WriteString(it.Span().Location())
			}

			if diff := len(list) - maxPer; diff > 0 {
				out.WriteString("\n\n\t... And ")
				if diff == 1 {
					out.WriteString("1 other")
				} else {
					out.WriteString(fmt.Sprintf("%d others", diff))
				}
				out.WriteString(" of type ")
				out.WriteString(key.Name())
			}
		}

		return errors.New(out.String())
	}

	return nil
}

type nodeInner struct {
	value Value
	span  Span
	done  bool
}
