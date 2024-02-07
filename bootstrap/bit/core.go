package bit

import "fmt"

func (program *Program) InitCore() {
	program.DeclareGlobal(TokenBreak, SplitLines{})

	program.DeclareGlobal(Symbol("("), ParseBrackets{"(", ")"})
	program.DeclareGlobal(Symbol(")"), ParseBrackets{"(", ")"})

	program.DeclareGlobal(Symbol("["), ParseBrackets{"[", "]"})
	program.DeclareGlobal(Symbol("]"), ParseBrackets{"[", "]"})

	program.DeclareGlobal(Symbol("{"), ParseBrackets{"{", "}"})
	program.DeclareGlobal(Symbol("}"), ParseBrackets{"{", "}"})
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

type ParseBrackets struct {
	Sta string
	End string
}

func (op ParseBrackets) Precedence() Precedence {
	return PrecBrackets
}

func (op ParseBrackets) IsSame(other Binding) bool {
	if v, ok := other.(ParseBrackets); ok {
		return op == v
	}
	return false
}

func (op ParseBrackets) String() string {
	return fmt.Sprintf("ParseBrackets(`%s%s`)", op.Sta, op.End)
}

func (op ParseBrackets) Process(args *BindArgs) {
	var stack []*Node
	for _, it := range args.Nodes {
		if it.Text() == op.Sta {
			stack = append(stack, it)
		} else if it.Text() == op.End {
			if l := len(stack); l > 0 {
				sta := stack[l-1]
				pos := sta.Index()
				par := sta.Parent()
				stack = stack[:l-1]
				nodes := par.RemoveRange(sta, it)
				group := args.Program.NewNode(Bracket(op), RangeSpan(sta, it))
				group.AddChildren(nodes[1 : len(nodes)-1]...)
				par.InsertNodes(pos, group)
			} else {
				it.AddError("closing bracket has no matching `%s`", op.Sta)
			}
		} else {
			it.AddError("invalid bracket for operator %s", op.String())
		}
	}

	for _, it := range stack {
		it.AddError("bracket `%s` is missing close `%s`", op.Sta, op.End)
	}
}

type Bracket struct {
	Sta string
	End string
}

func (val Bracket) Bind(node *Node) {}

func (val Bracket) String() string {
	return fmt.Sprintf("Bracket(`%s%s`)", val.Sta, val.End)
}
