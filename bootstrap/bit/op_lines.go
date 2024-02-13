package bit

// Implemented by non-semantic group nodes that can be "flattened" without
// losing meaning.
type CanFlatten interface {
	Flatten(node *Node) []*Node
}

func (val Group) Flatten(node *Node) []*Node {
	return node.Nodes()
}

func (val Line) Flatten(node *Node) []*Node {
	return node.Nodes()
}

func FlattenNodes(nodes ...*Node) (out []*Node) {
	for _, it := range nodes {
		if v, ok := it.Value().(CanFlatten); ok {
			out = append(out, v.Flatten(it)...)
		} else {
			out = append(out, it)
		}
	}
	return out
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
				span := SpanFromSlice(line)
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
