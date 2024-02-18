package bit_core

import (
	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
)

type Print struct{}

func (val Print) IsEqual(other Key) bool {
	if v, ok := other.(Print); ok {
		return v == val
	}
	return false
}

func (val Print) Repr(oneline bool) string {
	return "Print"
}

func (val Print) Bind(node *Node) {
	node.Bind(Print{})
}

func (val Print) Type(node *Node) code.Type {
	return node.Last().Type()
}

func (val Print) Output(ctx *code.OutputContext, node *Node) {
	args := node.OutputChildren(ctx)
	code := code.NewPrint(args...)
	ctx.Output(code)
}

type ParsePrint struct{}

func (op ParsePrint) IsSame(other bit.Binding) bool {
	if v, ok := other.(ParsePrint); ok {
		return v == op
	}
	return false
}

func (op ParsePrint) Precedence() bit.Precedence {
	return bit.PrecPrint
}

func (op ParsePrint) Process(args *bit.BindArgs) {
	for _, it := range args.Nodes {
		par, idx := it.Parent(), it.Index()
		if par == nil {
			it.Undo()
			continue
		}
		src := par.RemoveNodes(idx, par.Len())
		node := args.Program.NewNode(Print{}, SpanFromSlice(src))
		node.AddChildren(src[1:]...)
		par.InsertNodes(idx, node)
	}
}

func (op ParsePrint) String() string {
	return "ParsePrint"
}
