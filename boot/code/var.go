package code

import "fmt"

type Var struct {
	Name Id
	Type Type
}

func (expr Var) IsExpr() {}

func (expr Var) String() string {
	return fmt.Sprintf("Var(%s: %s)", expr.Name, expr.Type)
}
