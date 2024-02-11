package bit

import "fmt"

type Int int

func (val Int) Type() Type {
	return IntType{}
}

func (val Int) IsEqual(other Key) bool {
	if v, ok := other.(Int); ok {
		return v == val
	}
	return false
}

func (val Int) String() string {
	return fmt.Sprintf("%d", val)
}

func (val Int) Repr(oneline bool) string {
	return fmt.Sprintf("Int(%d)", val)
}

func (val Int) Bind(node *Node) {
	node.Bind(Int(0))
}

func (val Int) Output(ctx *CodeContext) Code {
	return Code{val, nil}
}

func (val Int) Eval(rt *RuntimeContext) {
	rt.Result = val
}

func (val Int) OutputCpp(ctx *CppContext, node *Node) {
	ctx.Expr.WriteString(val.String())
}

func (val Int) OutputCppPrint(ctx *CppContext, node *Node) {
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()
	ctx.Body.Write(`printf("`)
	ctx.Body.Write(val.String())
	ctx.Body.Write(`");`)
}

type ParseInt struct{}

func (ParseInt) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(TokenType); ok && tok == TokenInteger {
		val := ParseIntegerLiteral(node.Text())
		return Int(val), nil
	}
	return nil, nil
}
