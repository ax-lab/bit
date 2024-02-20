package bit_core

import (
	"fmt"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
)

type Int int

func (val Int) IsEqual(other any) bool {
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

func (val Int) Type(node *Node) code.Type {
	return code.IntType()
}

func (val Int) Output(ctx *code.OutputContext, node *Node, ans *code.Variable) {
	node.CheckEmpty(ctx)
	ctx.Output(ans.SetVar(code.NewInt(int(val))))
}

type ParseInt struct{}

func (ParseInt) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(bit.TokenType); ok && tok == bit.TokenInteger {
		val := bit.ParseIntegerLiteral(node.Text())
		return Int(val), nil
	}
	return nil, nil
}
