package bot

import "axlab.dev/bit/input"

type ParseContext interface {
	Error(err error)
	Nodes() NodeList
	Parse(nodes NodeList)

	Output(node ...Node)
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

func ParseLines(ctx ParseContext) {
	nodes := ctx.Nodes()
	items := nodes.Slice()

	cur := 0

	push := func(idx int) {
		if line := nodes.Range(cur, idx-1); line.Len() > 0 {
			ctx.Parse(line)
			ctx.Output(Line{line})
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
