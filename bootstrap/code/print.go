package code

import (
	"fmt"
	"strings"
)

type Print struct {
	Id
	Args []Expr
}

func NewPrint(args ...Expr) *Print {
	return &Print{Args: args}
}

func (prn *Print) Exec(rt *Runtime) error {
	hasOut := false
	for _, it := range prn.Args {
		if val, err := it.Eval(rt); err == nil {
			txt := val.String()
			if len(txt) > 0 {
				if hasOut {
					rt.Out(" ")
				}
				rt.Out(txt)
				hasOut = true
			}
		} else {
			rt.Out("\n")
			return err
		}
	}

	rt.Out("\n")
	return nil
}

func (prn *Print) OutputCpp(ctx *CppContext) {
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()

	if len(prn.Args) == 1 {
		arg := prn.Args[0]
		arg.Type().CppPrint(ctx, arg)
	} else if len(prn.Args) > 1 {
		hasOut := ctx.NewName("has_out")
		sep := fmt.Sprintf(`if (%s) printf(" ");`, hasOut)
		set := fmt.Sprintf(`%s = 1;`, hasOut)

		ctx.Body.Push("{")
		ctx.Body.Indent()

		for n, arg := range prn.Args {
			typ := arg.Type()
			if n > 0 {
				ctx.Body.Push(sep)
			}
			if cond := typ.CppPrintCondition(ctx, arg); cond != "" {
				ctx.Body.Push("if (%s) {", cond)
				ctx.Body.Indent()
				ctx.Body.Push(set)
				typ.CppPrint(ctx, arg)
				ctx.Body.Dedent()
				ctx.Body.Push("}")
			} else {
				ctx.Body.Push(set)
				typ.CppPrint(ctx, arg)
			}
		}

		ctx.Body.Dedent()
		ctx.Body.Push("}")
	}

	ctx.Body.Push(`printf("\n");`)
}

func (prn *Print) Repr(mode Repr) string {
	if mode == ReprLabel {
		return fmt.Sprintf("print(%d)", len(prn.Args))
	}

	out := strings.Builder{}
	out.WriteString("print(")
	for n, it := range prn.Args {
		if n > 0 {
			out.WriteString(", ")
		}
		out.WriteString(it.Repr(mode))
	}
	out.WriteString(")")
	return out.String()
}
