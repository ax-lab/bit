package bit

import (
	"fmt"
	"strings"

	"axlab.dev/bit/common"
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
		var expr CppContext
		expr.InitExpr(ctx)
		it.OutputCpp(&expr)
		if txt := expr.OutputExpr.Text(); txt != "" {
			ctx.OutputFunc.NewLine()
			ctx.OutputFunc.Write(txt)
			ctx.OutputFunc.EndStatement()
		}
	}
}
