package bit_core

import (
	"axlab.dev/bit/code"
)

type Line struct{}

func (val Line) Bind(node *Node) {
	node.Bind(Line{})
	node.Bind(Indented{})
}

func (val Line) Repr(oneline bool) string {
	return "Line"
}

func (val Line) IsEqual(other Key) bool {
	if v, ok := other.(Line); ok {
		return val == v
	}
	return false
}

func (val Line) Type(node *Node) code.Type {
	return node.Get(0).Type()
}

func (val Line) Output(ctx *code.OutputContext, node *Node) {
	node.OutputChild(ctx, true)
}