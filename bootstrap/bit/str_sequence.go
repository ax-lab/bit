package bit

import (
	"strings"

	"axlab.dev/bit/text"
)

type Sequence struct {
	List []Code
}

func (seq Sequence) Eval(rt *RuntimeContext) {
	for _, it := range seq.List {
		rt.Result = rt.Eval(it)
		if rt.Done() {
			break
		}
	}
}

func (seq Sequence) Repr() string {
	out := strings.Builder{}
	out.WriteString("Sequence {")
	for _, it := range seq.List {
		out.WriteString("\n")
		out.WriteString(text.Indent(it.Expr.Repr()))
	}
	if len(seq.List) > 0 {
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}
