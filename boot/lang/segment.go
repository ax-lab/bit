package lang

import (
	"io"

	"axlab.dev/bit/core"
)

type Line core.NodeList

type Bracket struct {
	Kind string
	Expr core.NodeList
}

type Block struct {
	Lines []core.NodeList
}

func OpSegment(list core.NodeList) {
	compiler := list.Compiler()
	for _, it := range list.Nodes() {
		src, ok := it.Value().(core.Source)
		if !ok {
			continue
		}

		lexer := compiler.Lexer.Copy()
		input := src.Span().Cursor()
		for input.Len() > 0 && !compiler.ShouldStop() {
			next := ReadNext(lexer, &input)
			if !next.Empty() {
				compiler.Eval(next)
			}
		}
	}
}

func ReadNext(lexer *core.Lexer, input *core.Cursor) (out core.NodeList) {
	out = core.NodeListNew(input.ToSpan())
	for {
		next, err := lexer.Read(input)
		if err == io.EOF {
			break
		} else if err != nil {
			if stop := out.PushError(err); stop {
				break
			}
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
