package bot

import "axlab.dev/bit/input"

type ParseContext interface {
	Nodes() NodeList
	Parse(nodes NodeList)

	Output(nodes ...Node)
	Error(err error)
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
		if idx > cur {
			line := nodes.Range(cur, idx)
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

func Parse(nodes NodeList) (errs []error) {
	ctx := parseContext{nodes: nodes}
	ParseLines(&ctx)
	ctx.nodes.data.Override(ctx.output)
	return ctx.errs
}

type parseContext struct {
	nodes  NodeList
	errs   []error
	output []Node
}

func (ctx *parseContext) Nodes() NodeList {
	return ctx.nodes
}

func (ctx *parseContext) Parse(nodes NodeList) {}

func (ctx *parseContext) Output(nodes ...Node) {
	ctx.output = append(ctx.output, nodes...)
}

func (ctx *parseContext) Error(err error) {
	ctx.errs = append(ctx.errs, err)
}
