package bit_lang

import (
	"fmt"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
)

type Str string

func (val Str) IsEqual(other any) bool {
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

func (val Str) Bind(node *bit.Node) {
	node.Bind(Str(""))
}

func (val Str) Type(node *bit.Node) code.Type {
	return code.StrType()
}

func (val Str) Output(ctx *code.OutputContext, node *bit.Node, ans *code.Variable) {
	node.CheckEmpty(ctx)
	ctx.Output(ans.SetVar(code.NewStr(string(val))))
}

type ParseString struct{}

func (ParseString) Get(node *bit.Node) (bit.Value, error) {
	if tok, ok := node.Value().(bit.TokenType); ok && tok == bit.TokenString {
		str := bit.ParseStringLiteral(node.Text())
		return Str(str), nil
	}
	return nil, nil
}
