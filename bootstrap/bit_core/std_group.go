package bit_core

import (
	"axlab.dev/bit/code"
)

type Group struct{}

func (val Group) Bind(node *Node) {
	node.Bind(Group{})
}

func (val Group) Repr(oneline bool) string {
	return "Group"
}

func (val Group) IsEqual(other any) bool {
	if v, ok := other.(Group); ok {
		return val == v
	}
	return false
}

func (val Group) Type(node *Node) code.Type {
	return node.Get(0).Type()
}

func (val Group) Output(ctx *code.OutputContext, node *Node, ans *code.Variable) {
	node.OutputChild(ctx, ans, true)
}
