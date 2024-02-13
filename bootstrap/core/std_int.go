package core

import (
	"fmt"

	"axlab.dev/bit/bit"
)

type Int int

func (val Int) Type() Type {
	return bit.IntType{}
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

func (val Int) Output(ctx *bit.CodeContext) Code {
	return Code{Expr: val}
}

func (val Int) Eval(rt *bit.RuntimeContext) {
	rt.Result = val
}

func (val Int) OutputCpp(ctx *bit.CppContext, node *Node) {
	ctx.Expr.WriteString(val.String())
}

func (val Int) OutputCppPrint(ctx *bit.CppContext, node *Node) {
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()
	ctx.Body.Write(`printf("`)
	ctx.Body.Write(val.String())
	ctx.Body.Write(`");`)
}

type ParseInt struct{}

func (ParseInt) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(bit.TokenType); ok && tok == bit.TokenInteger {
		val := bit.ParseIntegerLiteral(node.Text())
		return Int(val), nil
	}
	return nil, nil
}
