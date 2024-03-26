package bot

import (
	"fmt"

	"axlab.dev/bit/input"
)

type Bracket struct {
	kind  string
	nodes NodeList
	sta   Node
	end   Node
}

func (node Bracket) Span() input.Span {
	return node.sta.Span().Merged(node.end.Span())
}

func (node Bracket) Repr() string {
	return fmt.Sprintf("Bracket%s%s", node.sta.Span().Text(), node.end.Span().Text())
}

func ParseBrackets(ctx ParseContext, nodes NodeList) {
	var stack []Bracket

	push := func(node Node) {
		if last := len(stack) - 1; last >= 0 {
			stack[last].nodes.Push(node)
		} else {
			ctx.Push(node)
		}
	}

	hasError := false
	items := nodes.Slice()
	for idx := 0; !hasError && idx < len(items); idx++ {
		node := items[idx]
		if isOpen, kind := bracketIsOpen(node); isOpen {
			bracketSta := Bracket{
				kind:  kind,
				sta:   node,
				nodes: nodes.Range(idx+1, idx+1),
			}
			stack = append(stack, bracketSta)
		} else if isClose, kind := bracketIsClose(node); isClose {
			last := len(stack) - 1
			if last < 0 || stack[last].kind != kind {
				ctx.ErrorAt(node.Span(), "unmatched close bracket `%s`", node.Span().Text())
				hasError = true
			}

			bracketEnd := stack[last]
			bracketEnd.end = node
			stack = stack[:last]
			ctx.Queue(bracketEnd.nodes)
			push(bracketEnd)
		} else {
			push(node)
		}
	}

	if len(stack) > 0 && !hasError {
		last := stack[len(stack)-1]
		ctx.ErrorAt(last.sta.Span(), "missing close bracket for `%s`", last.sta.Span().Text())
	}
}

func bracketIsOpen(node Node) (is bool, kind string) {
	token, ok := node.(Token)
	if !ok || token.Kind() != TokenSymbol {
		return
	}

	switch txt := token.Span().Text(); txt {
	case "(":
		kind = "()"
	case "[":
		kind = "[]"
	case "{":
		kind = "{}"
	}

	return kind != "", kind
}

func bracketIsClose(node Node) (is bool, kind string) {
	token, ok := node.(Token)
	if !ok || token.Kind() != TokenSymbol {
		return
	}

	switch txt := token.Span().Text(); txt {
	case ")":
		kind = "()"
	case "]":
		kind = "[]"
	case "}":
		kind = "{}"
	}

	return kind != "", kind
}
