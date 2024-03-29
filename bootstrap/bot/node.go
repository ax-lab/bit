package bot

import (
	"fmt"

	"axlab.dev/bit/input"
)

type Node interface {
	Span() input.Span
	Repr() string
}

type NodeList struct {
	data *nodeListData
}

type nodeListData struct {
	shared bool
	source input.Source
	items  []Node
	offset int
	sta    int
	end    int
}

func NodeListNew[T Node](src input.Source, items ...T) NodeList {
	data := &nodeListData{
		source: src,
		items:  make([]Node, 0, len(items)),
		sta:    0,
		end:    len(items),
	}
	for _, it := range items {
		data.items = append(data.items, it)
	}
	return NodeList{data}
}

func (ls NodeList) Src() input.Source {
	return ls.data.source
}

func (ls NodeList) Len() int {
	return ls.data.end - ls.data.sta
}

func (ls NodeList) Get(index int) Node {
	index += ls.data.sta
	if index < ls.data.sta || ls.data.end <= index {
		panic("NodeList: index out of bounds")
	}
	return ls.data.items[index]
}

func (ls NodeList) Range(pos ...int) NodeList {
	sta, end := ls.data.GetRange(pos...)
	ls.data.shared = true
	return NodeList{
		&nodeListData{items: ls.data.items, sta: sta, end: end, shared: true},
	}
}

func (ls NodeList) Slice(pos ...int) []Node {
	sta, end := ls.data.GetRange(pos...)
	return ls.data.items[sta:end]
}

func (ls NodeList) Span() input.Span {
	list := ls.data.items
	size := len(list)
	if size == 0 {
		offset := ls.data.offset
		return ls.data.source.Span().Range(offset, offset)
	}

	if ls.data.sta >= size {
		span := list[size-1].Span()
		return span.Range(span.Len(), span.Len())
	}

	if ls.data.sta == ls.data.end {
		span := list[ls.data.sta].Span()
		return span.WithLen(0)
	}

	sta := list[ls.data.sta].Span()
	end := list[ls.data.end-1].Span()
	return sta.Merged(end)
}

func (ls NodeList) Push(nodes ...Node) {
	if len(nodes) == 0 {
		return
	}
	if ls.data.shared {
		ls.data.items = append([]Node(nil), ls.data.items...)
		ls.data.shared = false
	}

	ls.data.items = append(ls.data.items, nodes...)
}

func (ls *nodeListData) GetRange(pos ...int) (sta, end int) {
	if len(pos) > 2 {
		panic("NodeList: invalid range")
	}

	cnt := ls.end - ls.sta
	sta, end = 0, cnt
	if len(pos) > 0 {
		sta = pos[0]
		if len(pos) > 1 {
			end = pos[1]
		}
	}

	if sta < 0 || end < sta || end > cnt {
		panic(fmt.Sprintf("NodeList: out of bounds range (%d-%d)", sta, end))
	}

	sta += ls.sta
	end += ls.sta
	return
}

func (ls *nodeListData) Override(nodes []Node) {
	ls.sta = 0
	ls.end = len(nodes)
	ls.items = nodes
	ls.shared = false
	if len(nodes) > 0 {
		ls.offset = nodes[0].Span().Sta()
	}
}
