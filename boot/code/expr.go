package code

type ExprValue interface {
	IsExpr()
	String() string
}

type Expr struct{ *exprData }

type exprData struct {
	value ExprValue
}

func ExprNew(value ExprValue) Expr {
	data := &exprData{value: value}
	return Expr{data}
}

func (expr Expr) Value() ExprValue {
	return expr.value
}

func (expr Expr) String() string {
	if expr.exprData == nil {
		return "Expr(nil)"
	} else if expr.value == nil {
		return "(nil)"
	}
	return expr.value.String()
}
