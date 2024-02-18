package code

import (
	"strings"

	"axlab.dev/bit/common"
)

type If struct {
	Id
	Cond Expr
	True Stmt
	Else Stmt
}

func NewIf(cond Expr, ifTrue Stmt, ifFalse Stmt) *If {
	return &If{
		Cond: cond,
		True: ifTrue,
		Else: ifFalse,
	}
}

func (ifExpr *If) Exec(rt *Runtime) error {
	cond, err := ifExpr.Cond.Eval(rt)
	if err != nil {
		return err
	}

	if cond.Bool() {
		return ifExpr.True.Exec(rt)
	} else if ifExpr.Else != nil {
		return ifExpr.Else.Exec(rt)
	}

	return nil
}

func (ifExpr *If) OutputCpp(ctx *CppContext) {
	ctx.Body.Write("if (")
	ifExpr.Cond.OutputCpp(ctx)
	ctx.Body.Write(") ")
	ifExpr.True.OutputCpp(ctx)
	if ifExpr.Else != nil {
		ctx.Body.EnsureBlank()
		ctx.Body.Write("else ")
		ifExpr.Else.OutputCpp(ctx)
	}
	ctx.Body.EnsureBlank()
}

func (ifExpr *If) Repr(mode Repr) string {
	switch mode {
	case ReprLabel:
		if ifExpr.Else != nil {
			return "if else"
		} else {
			return "if"
		}

	case ReprLine:
		repr_cond := ifExpr.Cond.Repr(mode)
		repr_true := ifExpr.True.Repr(mode)

		var repr_else string
		if ifExpr.Else != nil {
			repr_else = " else " + ifExpr.Else.Repr(mode)
		}

		out := strings.Builder{}
		out.WriteString("if (")
		out.WriteString(repr_cond)
		out.WriteString(") ")

		len_full := out.Len() + len(repr_true) + len(repr_else)
		if len_full <= MaxLine {
			out.WriteString(repr_true)
			out.WriteString(repr_else)
		} else {
			if repr_else != "" {
				len_full -= len(repr_else)
				repr_else = " else …"
				len_full += len(repr_else)
			}

			if len_full > MaxLine {
				repr_true = "…"
			}

			out.WriteString(repr_true)
			out.WriteString(repr_else)
		}

		return out.String()

	default:
		out := strings.Builder{}
		out.WriteString("if (")
		out.WriteString(common.Indented(ifExpr.True.Repr(mode)))
		out.WriteString(") ")
		out.WriteString(ifExpr.True.Repr(mode))
		if ifExpr.Else != nil {
			out.WriteString(" else ")
			out.WriteString(ifExpr.Else.Repr(mode))
		}
		return out.String()
	}
}
