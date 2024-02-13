package core

import (
	"fmt"

	"axlab.dev/bit/bit"
)

type Str string

func (val Str) Type() Type {
	return bit.StrType{}
}

func (str Str) IsEqual(other Key) bool {
	if v, ok := other.(Str); ok {
		return v == str
	}
	return false
}

func (str Str) String() string {
	return string(str)
}

func (str Str) Repr(oneline bool) string {
	return fmt.Sprintf("Str(%v)", string(str))
}

func (str Str) Bind(node *Node) {
	node.Bind(Str(""))
}

func (val Str) Output(ctx *CodeContext) Code {
	return Code{Expr: val}
}

func (val Str) Eval(rt *RuntimeContext) {
	rt.Result = val
}

func (val Str) OutputCpp(ctx *CppContext, node *Node) {
	bit.WriteLiteralString(ctx.Expr, string(val))
}

func (val Str) OutputCppPrint(ctx *CppContext, node *Node) {
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()
	ctx.Body.Write(`printf("%s", `)
	bit.WriteLiteralString(ctx.Body, string(val))
	ctx.Body.Write(`);`)
}

type ParseString struct{}

func (ParseString) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(TokenType); ok && tok == bit.TokenString {
		str := bit.ParseStringLiteral(node.Text())
		return Str(str), nil
	}
	return nil, nil
}
