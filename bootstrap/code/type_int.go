package code

func IntType() Type {
	return Type{typeInt}
}

var (
	typeInt *typeData = &typeData{TypeInt, "", intImpl{}}
)

type intImpl struct{}

func (intImpl) CppType() string {
	return "int"
}

func (intImpl) String() string {
	return "int"
}

func (intImpl) CppPrint(ctx *CppContext, expr Expr) {
	ctx.Body.Write(`printf("%d", `)
	expr.OutputCpp(ctx)
	ctx.Body.Write(`);`)
}

func (intImpl) CppPrintCondition(ctx *CppContext, expr Expr) string {
	return ""
}
