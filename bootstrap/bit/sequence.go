package bit

import (
	"fmt"
	"strings"

	"axlab.dev/bit/common"
)

type Sequence struct {
	List []Code
}

func (seq Sequence) Type() Type {
	if len(seq.List) == 0 {
		return NoneType{}
	} else {
		return seq.List[len(seq.List)-1].Type()
	}
}

func (seq Sequence) Eval(rt *RuntimeContext) {
	for _, it := range seq.List {
		rt.Result = rt.Eval(it)
		if rt.Done() {
			break
		}
	}
}

func (seq Sequence) Repr(oneline bool) string {
	if oneline {
		return fmt.Sprintf("Sequence(%d)", len(seq.List))
	}
	out := strings.Builder{}
	out.WriteString("Sequence {")
	for _, it := range seq.List {
		out.WriteString("\n")
		out.WriteString(common.Indent(it.Expr.Repr(false)))
	}
	if len(seq.List) > 0 {
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

func (seq Sequence) OutputCpp(ctx *CppContext, node *Node) {
	for _, it := range seq.List {
		ctx.Expr.Reset()
		it.OutputCpp(ctx)
	}
}
