package bit_core

import (
	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
)

type ParseBlocks struct{}

func (op ParseBlocks) IsSame(other bit.Binding) bool {
	if v, ok := other.(ParseBlocks); ok {
		return v == op
	}
	return false
}

func (op ParseBlocks) Precedence() bit.Precedence {
	return bit.PrecBlocks
}

func (op ParseBlocks) Process(args *bit.BindArgs) {
	for _, it := range args.Nodes {
		par, prev := it.Parent(), it.Prev()

		valid := par != nil && prev != nil
		if !valid {
			it.AddError("invalid indented block")
			continue
		}

		nodes := it.RemoveNodes(0, it.Len())
		if last := prev.Last(); IsSymbol(last, ":") {
			last.FlagDone()
			last.Remove()
			block := it.ReplaceWithValue(Block{})
			block.AddChildren(nodes...)
			block.Remove()
			prev.AddChildren(block)
		} else {
			nodes = FlattenNodes(nodes...)
			prev.AddChildren(nodes...)
		}
	}
}

func (op ParseBlocks) String() string {
	return "ParseBlocks"
}

type Block struct{}

func (val Block) IsScope(node *Node) bool {
	return true
}

func (val Block) IsEqual(other Key) bool {
	if v, ok := other.(Block); ok {
		return val == v
	}
	return false
}

func (val Block) Repr(oneline bool) string {
	return "Block"
}

func (val Block) Bind(node *Node) {
	node.Bind(Block{})
}

func (val Block) Type(node *Node) code.Type {
	return node.Last().Type()
}

func (val Block) Output(ctx *code.OutputContext, node *Node, ans *code.Variable) {
	node.OutputChildren(ctx, ans)
}
