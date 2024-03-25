package bot

import (
	"errors"
	"fmt"

	"axlab.dev/bit/input"
)

type ParseContext interface {
	Nodes() NodeList
	Parse(nodes NodeList)

	SetNext(eval func(ctx ParseContext))

	Output(nodes ...Node)
	Error(err error)
	ErrorAt(span input.Span, msg string, args ...any)
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

func Parse(nodes NodeList) (errs []error) {
	return parseNodes(nodes, ParseBrackets)
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

func parseNodes(nodes NodeList, eval func(ctx ParseContext)) (errs []error) {
	ctx := parseContext{nodes: nodes}
	eval(&ctx)
	ctx.nodes.data.Override(ctx.output)
	ctx.nextList = append(ctx.nextList, ctx.nodes)

	if ctx.nextEval != nil {
		for i := 0; i < len(ctx.nextList) && len(errs) == 0; i++ {
			parseNodes(ctx.nextList[i], ctx.nextEval)
		}
	}

	return ctx.errs
}

type parseContext struct {
	nodes    NodeList
	errs     []error
	output   []Node
	nextList []NodeList
	nextEval func(ctx ParseContext)
}

func (ctx *parseContext) Nodes() NodeList {
	return ctx.nodes
}

func (ctx *parseContext) SetNext(eval func(ctx ParseContext)) {
	ctx.nextEval = eval
}

func (ctx *parseContext) Parse(nodes NodeList) {
	ctx.nextList = append(ctx.nextList, nodes)
}

func (ctx *parseContext) Output(nodes ...Node) {
	ctx.output = append(ctx.output, nodes...)
}

func (ctx *parseContext) Error(err error) {
	ctx.errs = append(ctx.errs, err)
}

func (ctx *parseContext) ErrorAt(span input.Span, msg string, args ...any) {
	var err error
	if len(args) > 0 {
		err = fmt.Errorf(msg, args...)
	} else {
		err = errors.New(msg)
	}
	ctx.Error(span.ErrorAt(err))
}
