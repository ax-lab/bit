package bit

type Print struct{}

func (val Print) IsEqual(other Key) bool {
	if v, ok := other.(Print); ok {
		return v == val
	}
	return false
}

func (val Print) Repr() string {
	return "Print"
}

func (val Print) Bind(node *Node) {
	node.Bind(Print{})
}

type ParsePrint struct{}

func (op ParsePrint) IsSame(other Binding) bool {
	if v, ok := other.(ParsePrint); ok {
		return v == op
	}
	return false
}

func (op ParsePrint) Precedence() Precedence {
	return PrecPrint
}

func (op ParsePrint) Process(args *BindArgs) {
	for _, it := range args.Nodes {
		par, idx := it.Parent(), it.Index()
		if par == nil {
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
