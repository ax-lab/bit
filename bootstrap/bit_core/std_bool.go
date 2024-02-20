package bit_core

import (
	"fmt"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
)

type Bool bool

func (val Bool) IsEqual(other any) bool {
	if v, ok := other.(Bool); ok {
		return v == val
	}
	return false
}

func (val Bool) String() string {
	if val {
		return "true"
	} else {
		return "false"
	}
}

func (val Bool) Repr(oneline bool) string {
	return fmt.Sprintf("Bool(%s)", val.String())
}

func (val Bool) Bind(node *Node) {
	node.Bind(Bool(false))
}

func (val Bool) Type(node *Node) code.Type {
	return code.BoolType()
}

func (val Bool) Output(ctx *code.OutputContext, node *Node, ans *code.Variable) {
	node.CheckEmpty(ctx)
	ctx.Output(ans.SetVar(code.NewBool(bool(val))))
}

type ParseBool struct{}

func (ParseBool) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(bit.TokenType); ok && tok == bit.TokenWord {
		switch node.Text() {
		case "true":
			return Bool(true), nil
		case "false":
			return Bool(false), nil
		}
	}
	return nil, nil
}
