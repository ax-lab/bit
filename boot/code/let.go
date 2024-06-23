package code

import "fmt"

type Let struct {
	Decl Var
	Init Expr
}

func (expr Let) IsExpr() {}

func (expr Let) String() string {
	return fmt.Sprintf("Let(%s: %s = %s)", expr.Decl.Name, expr.Decl.Type, expr.Init)
}
