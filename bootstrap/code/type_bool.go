package code

func BoolType() Type {
	return Type{typeBool}
}

var (
	typeBool *typeData = &typeData{TypeBool, "", boolImpl{}}
)

type boolImpl struct{}

func (boolImpl) CppType() string {
	return "bool"
}

func (boolImpl) String() string {
	return "bool"
}

func (boolImpl) CppPrint(ctx *CppContext, expr Expr) {
	ctx.Body.Write(`printf(`)
	expr.OutputCpp(ctx)
	ctx.Body.Write(` ? "true" : "false");`)
}

func (boolImpl) CppPrintCondition(ctx *CppContext, expr Expr) string {
	return ""
}
