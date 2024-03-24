package bot

import "axlab.dev/bit/input"

type Node interface {
	Span() input.Span
	Repr() string
}

type NodeList struct {
	source input.Source
	items  []Node
	sta    int
	end    int
}

func NodeListNew[T Node](src input.Source, items ...T) NodeList {
	out := NodeList{
		source: src,
		items:  make([]Node, 0, len(items)),
		sta:    0,
		end:    len(items),
	}
	for _, it := range items {
		out.items = append(out.items, it)
	}
	return out
}

func (ls NodeList) Src() input.Source {
	return ls.source
}

func (ls NodeList) Len() int {
	return ls.end - ls.sta
}

func (ls NodeList) Get(index int) Node {
	index += ls.sta
	if index < ls.sta || ls.end <= index {
		panic("NodeList: index out of bounds")
	}
	return ls.items[index]
}

func (ls NodeList) Range(pos ...int) NodeList {
	sta, end := ls.getRange(pos...)
	if sta < 0 || end < sta || end > ls.Len() {
		panic("NodeList: out of bounds range")
	}
	return NodeList{items: ls.items, sta: ls.sta + sta, end: ls.sta + end}
}

func (ls NodeList) Slice(pos ...int) []Node {
	sta, end := ls.getRange(pos...)
	return ls.items[sta:end]
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
	list := ls.items
	size := len(list)
	if size == 0 {
		return ls.source.Span().WithLen(0)
	}

	if ls.sta >= size {
		span := list[size-1].Span()
		return span.Range(span.Len(), span.Len())
	}

	if ls.sta == ls.end {
		span := list[ls.sta].Span()
		return span.WithLen(0)
	}

	sta := list[ls.sta].Span()
	end := list[ls.end-1].Span()
	return sta.Merged(end)
}
