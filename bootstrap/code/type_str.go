package code

import "fmt"

func StrType() Type {
	return Type{typeStr}
}

var (
	typeStr *typeData = &typeData{TypeStr, "", strImpl{}}
)

type strImpl struct{}

func (strImpl) CppType() string {
	return "const char *"
}

func (strImpl) String() string {
	return "str"
}

func (strImpl) CppPrint(ctx *CppContext, expr Expr) {
	ctx.Body.Write(`printf("%s", `)
	expr.OutputCpp(ctx)
	ctx.Body.Write(`);`)
}

func (strImpl) CppPrintCondition(ctx *CppContext, expr Expr) string {
	code := ctx.ExprString(expr)
	return fmt.Sprintf("(%s)[0] != 0", code)
}
