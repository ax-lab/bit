package code

import "strings"

type Block struct {
	List []Expr
}

func (expr Block) IsExpr() {}

func (expr Block) String() string {
	out := strings.Builder{}
	out.WriteString("Block{")
	for n, expr := range expr.List {
		if n > 0 {
			out.WriteString("; ")
		}
		out.WriteString(expr.String())
	}
	out.WriteString("}")
	return out.String()
}
