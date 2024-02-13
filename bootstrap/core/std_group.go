package core

import "axlab.dev/bit/bit"

type Group struct{}

func (val Group) Bind(node *Node) {
	node.Bind(Group{})
}

func (val Group) Repr(oneline bool) string {
	return "Group"
}

func (val Group) IsEqual(other Key) bool {
	if v, ok := other.(Group); ok {
		return val == v
	}
	return false
}

func (val Group) Output(ctx *bit.CodeContext) Code {
	return ctx.OutputChild(ctx.Node)
}
