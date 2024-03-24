package bot

import "axlab.dev/bit/input"

type Node interface {
	Span() input.Span
	Repr() string
}

type NodeList struct {
	data *nodeListData
}

type nodeListData struct {
	source input.Source
	items  []Node
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
	sta, end := ls.getRange(pos...)
	if sta < 0 || end < sta || end > ls.Len() {
		panic("NodeList: out of bounds range")
	}
	return NodeList{
		&nodeListData{items: ls.data.items, sta: ls.data.sta + sta, end: ls.data.sta + end},
	}
}

func (ls NodeList) Slice(pos ...int) []Node {
	sta, end := ls.getRange(pos...)
	return ls.data.items[sta:end]
}

func (ls NodeList) getRange(pos ...int) (sta, end int) {
	sta, end = 0, ls.Len()
	if len(pos) > 0 {
		sta = pos[0]
		if len(pos) > 1 {
			end = pos[1]
		}
	}

	if sta < 0 || end < sta || end > ls.Len() {
		panic("NodeList: out of bounds range")
	}
	return
}

func (ls NodeList) Span() input.Span {
	list := ls.data.items
	size := len(list)
	if size == 0 {
		return ls.data.source.Span().WithLen(0)
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
