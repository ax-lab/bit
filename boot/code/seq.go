package code

import (
	"fmt"
	"strings"

	"axlab.dev/bit/core"
)

type Seq struct {
	List []core.Expr
}

func (seq Seq) String() string {
	out := strings.Builder{}
	for n, it := range seq.List {
		if n > 0 {
			out.WriteString("\n")
		}
		out.WriteString(it.String())
	}

	txt := fmt.Sprintf("Seq(%s)", core.IndentBlock(out.String()))
	return txt
}

func (seq Seq) Eval(rt *core.Runtime) (out core.Value, err error) {
	for _, it := range seq.List {
		out, err = it.Eval(rt)
		if err != nil {
			break
		}
	}
	return
}
