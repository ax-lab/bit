package bit

func (program *Program) InitCore() {
	program.DeclareGlobal(TokenBreak, SplitLines{})
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
	for _, it := range args.Nodes {
		par := it.Parent()
		if par == nil {
			continue
		}

		index, count := it.Index(), par.Len()
		if index == 0 || index == count-1 {
			it.Remove()
		} else {
			nodes := par.RemoveNodes(0, count)
			lineA := nodes[:index]
			lineB := nodes[index+1:]
			nodeA := args.Program.NewNode(Line{}, SliceSpan(lineA))
			nodeA.AddChildren(lineA...)
			nodeB := args.Program.NewNode(Line{}, SliceSpan(lineB))
			nodeB.AddChildren(lineB...)
			par.AddChildren(nodeA, nodeB)
		}
	}
}

func (op SplitLines) String() string {
	return "SplitLines"
}

type Line struct{}

// IsEqual implements Key.
func (line Line) IsEqual(other Key) bool {
	if v, ok := other.(Line); ok {
		return line == v
	}
	return false
}

func (line Line) Bind(node *Node) {
	node.Bind(line)
}

// String implements Value.
func (line Line) String() string {
	return "Line"
}
