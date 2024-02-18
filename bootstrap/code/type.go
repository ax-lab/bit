package code

import "fmt"

type Type interface {
	String() string
	CppType() string
	CppPrint(ctx *CppContext, expr Expr)
	CppPrintCondition(ctx *CppContext, expr Expr) string
}

func AddTypes(types ...Type) Type {
	panic("TODO: AddTypes")
}

func InvalidType() Type {
	panic("TODO: InvalidType")
}

func VoidType() Type {
	panic("TODO: VoidType")
}

type BoolType struct{}

func (BoolType) CppType() string {
	return "bool"
}

func (BoolType) String() string {
	return "bool"
}

func (BoolType) CppPrint(ctx *CppContext, expr Expr) {
	ctx.Body.Write(`printf(`)
	expr.OutputCpp(ctx)
	ctx.Body.Write(` ? "true" : "false");`)
}

func (BoolType) CppPrintCondition(ctx *CppContext, expr Expr) string {
	return ""
}

type IntType struct{}

func (IntType) CppType() string {
	return "int"
}

func (IntType) String() string {
	return "int"
}

func (IntType) CppPrint(ctx *CppContext, expr Expr) {
	ctx.Body.Write(`printf("%d", `)
	expr.OutputCpp(ctx)
	ctx.Body.Write(`);`)
}

func (IntType) CppPrintCondition(ctx *CppContext, expr Expr) string {
	return ""
}

type StrType struct{}

func (StrType) CppType() string {
	return "const char *"
}

func (StrType) String() string {
	return "str"
}

func (StrType) CppPrint(ctx *CppContext, expr Expr) {
	ctx.Body.Write(`printf("%s", `)
	expr.OutputCpp(ctx)
	ctx.Body.Write(`);`)
}

func (StrType) CppPrintCondition(ctx *CppContext, expr Expr) string {
	code := ctx.ExprString(expr)
	return fmt.Sprintf("(%s)[0] != 0", code)
}
