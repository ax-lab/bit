package bit

type Line struct{}

func (val Line) IsEqual(other Key) bool {
	if v, ok := other.(Line); ok {
		return val == v
	}
	return false
}

func (val Line) Bind(node *Node) {
	node.Bind(Line{})
	node.Bind(Indented{})
}

func (val Line) Repr() string {
	return "Line"
}

type SplitLines struct{}

func (op SplitLines) IsSame(other Binding) bool {
	if v, ok := other.(SplitLines); ok {
		return v == op
	}
	return false
}

func (op SplitLines) Precedence() Precedence {
	return PrecLines
}

func (op SplitLines) Process(args *BindArgs) {
	for _, nodes := range args.NodesByParent() {
		par := nodes[0].Parent()
		cur, children := 0, par.RemoveNodes(0, par.Len())
		push := func(line []*Node) {
			if len(line) > 0 {
				span := SliceSpan(line)
				node := args.Program.NewNode(Line{}, span)
				node.AddChildren(line...)
				par.AddChildren(node)
			}
		}
		for _, it := range nodes {
			pos := it.Index()
			line := children[cur:pos]
			push(line)
			cur = pos + 1
		}

		push(children[cur:])
	}
}

func (op SplitLines) String() string {
	return "SplitLines"
}
