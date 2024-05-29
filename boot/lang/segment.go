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
	rt := mod.Runtime()

	offset := 0
	for n, it := range list.Nodes() {
		idx := offset + n
		src, ok := it.Value().(core.Source)
		if !ok {
			continue
		}

		span := src.Span()
		lexer := mod.NewLexer()
		lines := []core.Node(nil)
		input := span.Cursor()
		for input.Len() > 0 && !rt.ShouldStop() {
			next := ParseLine(mod, lexer, &input)
			if next.Len() > 0 {
				rt.Eval(mod, next)
				line := core.NodeNew(next.Span(), Line(next))
				lines = append(lines, line)
			}
		}

		list.Replace(idx, idx+1, lines...)
		offset += len(lines) - 1
	}
}

func ParseLine(mod *core.Module, lexer *core.Lexer, input *core.Cursor) (out core.NodeList) {
	out = core.NodeListNew(input.ToSpan())
	for !mod.Runtime().ShouldStop() {
		next := lexer.Read(mod, input)
		if !next.Valid() {
			break
		}

		if _, eol := next.Value().(core.LineBreak); eol {
			if out.Len() > 0 {
				break
			} else {
				continue
			}
		}

		if next.Valid() {
			out.Push(next)
		}
	}
	return out
}
