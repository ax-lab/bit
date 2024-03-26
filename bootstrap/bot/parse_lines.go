package bot

import "axlab.dev/bit/input"

func ParseLines(ctx ParseContext, nodes NodeList) {
	items := nodes.Slice()
	cur := 0

	push := func(idx int) {
		if idx > cur {
			line := nodes.Range(cur, idx)
			ctx.Queue(line)
			ctx.Push(Line{line})
		}
		cur = idx + 1
	}

	for idx := 0; idx < len(items); idx++ {
		tok, ok := items[idx].(Token)
		if !ok || tok.Kind() != TokenBreak {
			continue
		}
		push(idx)
	}

	if last := len(items); cur < last {
		push(last)
	}
}

type Line struct {
	nodes NodeList
}

func (line Line) Span() input.Span {
	return line.nodes.Span()
}

func (line Line) Repr() string {
	return "Line"
}

func (line Line) NodeRepr(repr *NodeRepr) {
	repr.Header(line)
	repr.Items(line.nodes.Slice(), ReprPrefix(" {"), ReprSuffix("}"))
}
