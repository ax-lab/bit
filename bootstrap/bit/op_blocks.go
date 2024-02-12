package bit

type ParseBlocks struct{}

func (op ParseBlocks) IsSame(other Binding) bool {
	if v, ok := other.(ParseBlocks); ok {
		return v == op
	}
	return false
}

func (op ParseBlocks) Precedence() Precedence {
	return PrecBlocks
}

func (op ParseBlocks) Process(args *BindArgs) {
	for _, it := range args.Nodes {
		par, prev := it.Parent(), it.Prev()

		valid := par != nil && prev != nil
		if !valid {
			it.AddError("invalid indented block")
			continue
		}

		nodes := it.RemoveNodes(0, it.Len())
		if last := prev.Last(); last.IsSymbol(":") {
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

func (val Block) Output(ctx *CodeContext) Code {
	return ctx.OutputChildren(ctx.Node)
}
