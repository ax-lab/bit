package code

import "strings"

type Print struct {
	Args []Expr
}

func (expr Print) IsExpr() {}

func (expr Print) String() string {
	out := strings.Builder{}
	out.WriteString("Print(")
	for n, arg := range expr.Args {
		if n > 0 {
			out.WriteString(", ")
		}
		out.WriteString(arg.String())
	}
	out.WriteString(")")
	return out.String()
}
