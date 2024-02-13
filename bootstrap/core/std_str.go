package core

import (
	"fmt"

	"axlab.dev/bit/bit"
)

type Str string

func (val Str) Type() Type {
	return bit.StrType{}
}

func (val Str) IsEqual(other Key) bool {
	if v, ok := other.(Str); ok {
		return v == val
	}
	return false
}

func (val Str) String() string {
	return string(val)
}

func (val Str) Repr(oneline bool) string {
	return fmt.Sprintf("Str(%v)", string(val))
}

func (val Str) Bind(node *Node) {
	node.Bind(Str(""))
}

func (val Str) Output(ctx *bit.CodeContext) Code {
	return Code{Expr: val}
}

func (val Str) Eval(rt *bit.RuntimeContext) {
	rt.Result = val
}

func (val Str) OutputCpp(ctx *bit.CppContext, node *Node) {
	bit.WriteLiteralString(ctx.Expr, string(val))
}

func (val Str) OutputCppPrint(ctx *bit.CppContext, node *Node) {
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()
	ctx.Body.Write(`printf("%s", `)
	bit.WriteLiteralString(ctx.Body, string(val))
	ctx.Body.Write(`);`)
}

type ParseString struct{}

func (ParseString) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(bit.TokenType); ok && tok == bit.TokenString {
		str := bit.ParseStringLiteral(node.Text())
		return Str(str), nil
	}
	return nil, nil
}
