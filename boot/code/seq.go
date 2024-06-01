package code

import (
	"fmt"
	"strings"

	"axlab.dev/bit/core"
)

type Seq struct {
	span core.Span
	list []core.Expr
}

func SeqNew(span core.Span, list ...core.Expr) Seq {
	return Seq{span, list}
}

func (seq *Seq) Push(expr ...core.Expr) {
	seq.list = append(seq.list, expr...)
}

func (seq Seq) Span() core.Span {
	return seq.span
}

func (seq Seq) List() []core.Expr {
	return seq.list
}

func (seq Seq) String() string {
	out := strings.Builder{}
	for n, it := range seq.list {
		if n > 0 {
			out.WriteString("\n")
		}
		out.WriteString(it.String())
	}

	txt := fmt.Sprintf("Seq(%s)", core.IndentBlock(out.String()))
	return txt
}

func (seq Seq) Eval(rt *core.Runtime) (out core.Value, err error) {
	for _, it := range seq.list {
		out, err = it.Eval(rt)
		if err != nil {
			break
		}
	}
	return
}
