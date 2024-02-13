package bit

type Type interface {
	CppType() string
	String() string
}

type BoolType struct{}

func (BoolType) CppType() string {
	return "bool"
}

func (BoolType) String() string {
	return "bool"
}

func (BoolType) OutputCppPrint(ctx *CppContext, node *Node) {
	expr := ctx.Expr.String()
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()
	ctx.Body.Write(`printf(`)
	ctx.Body.Write(expr)
	ctx.Body.Write(` ? "true" : "false");`)
}

type IntType struct{}

func (IntType) CppType() string {
	return "int"
}

func (IntType) String() string {
	return "int"
}

func (IntType) OutputCppPrint(ctx *CppContext, node *Node) {
	expr := ctx.Expr.String()
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()
	ctx.Body.Write(`printf("%d", `)
	ctx.Body.Write(expr)
	ctx.Body.Write(`);`)
}

type StrType struct{}

func (StrType) CppType() string {
	return "const char *"
}

func (StrType) String() string {
	return "str"
}

func (StrType) OutputCppPrint(ctx *CppContext, node *Node) {
	expr := ctx.Expr.String()
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()
	ctx.Body.Write(`printf("%s", `)
	ctx.Body.Write(expr)
	ctx.Body.Write(`);`)
}

type NoneType struct{}

func (NoneType) CppType() string {
	return "void"
}

func (NoneType) String() string {
	return "(none)"
}

type InvalidType struct{}

func (InvalidType) CppType() string {
	return "!INVALID!"
}

func (InvalidType) String() string {
	return "(invalid)"
}
