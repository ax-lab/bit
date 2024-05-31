package lang

import (
	"fmt"

	"axlab.dev/bit/core"
)

type Line core.NodeList

func (line Line) String() string {
	list := core.NodeList(line)
	return fmt.Sprintf("Line(%s)", list.String())
}

func (line Line) Dump() string {
	list := core.NodeList(line)
	return fmt.Sprintf("Line(%s)", core.IndentBlock(list.Dump()))
}

type Bracket struct {
	Kind string
	Expr core.NodeList
}

type Block struct {
	Lines []core.NodeList
}

func OpSegment(mod *core.Module, list core.NodeList) {
	offset := 0
	for n, it := range list.Nodes() {
		idx := offset + n
		src, ok := it.Value().(core.Source)
		if !ok {
			continue
		}

		span := src.Span()
		input := span.Cursor()
		lexer := mod.NewLexer()
		lines := lexer.Tokenize(mod, &input)
		list.Replace(idx, idx+1, lines...)
		offset += len(lines) - 1
	}
}

func ParseLine(mod *core.Module, lexer *core.Lexer, input *core.Cursor) (out core.Node) {
	return readLine(mod, lexer, input)
}

func readLine(mod *core.Module, lexer *core.Lexer, input *core.Cursor) (out core.Node) {
	rt := mod.Runtime()

	var nodes []core.Node
	for !rt.ShouldStop() {
		next := lexer.Read(mod, input)
		if !next.Valid() {
			break
		}

		if _, eol := next.Value().(core.LineBreak); eol {
			if len(nodes) > 0 {
				break
			} else {
				continue
			}
		}

		if next.Valid() {
			nodes = append(nodes, next)
		}
	}

	if len(nodes) == 0 {
		return
	}

	line := core.NodeListNew(core.SpanForRange(nodes), nodes...)
	rt.Eval(mod, line)

	out = core.NodeNew(line.Span(), Line(line))
	return out
}
