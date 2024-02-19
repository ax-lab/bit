package bit_core

import (
	"fmt"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
)

type Str string

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

func (val Str) Type(node *Node) code.Type {
	return code.StrType()
}

func (val Str) Output(ctx *code.OutputContext, node *Node, ans *code.Variable) {
	node.CheckEmpty(ctx)
	ctx.Output(ans.SetVar(code.NewStr(string(val))))
}

type ParseString struct{}

func (ParseString) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(bit.TokenType); ok && tok == bit.TokenString {
		str := bit.ParseStringLiteral(node.Text())
		return Str(str), nil
	}
	return nil, nil
}
