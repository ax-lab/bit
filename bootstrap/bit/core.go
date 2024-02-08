package bit

import "fmt"

func (program *Program) InitCore() {
	program.DeclareGlobal(TokenBreak, SplitLines{})
	program.DeclareGlobal(Line{}, ParseIndent{})

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
	for _, nodes := range args.NodesByParent() {
		par := nodes[0].Parent()
		cur, children := 0, par.RemoveNodes(0, par.Len())
		push := func(line []*Node) {
			if len(line) > 0 {
				span := SliceSpan(line)
				node := args.Program.NewNode(Line{span.Indent()}, span)
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

type Line struct {
	Level int
}

func (val Line) IsEqual(other Key) bool {
	if v, ok := other.(Line); ok {
		return val == v
	}
	return false
}

func (val Line) Bind(node *Node) {
	node.Bind(Line{})
}

func (val Line) String() string {
	return fmt.Sprintf("Line[%d]", val.Level)
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

type ParseIndent struct{}

func (op ParseIndent) Precedence() Precedence {
	return PrecBrackets
}

func (op ParseIndent) IsSame(other Binding) bool {
	if v, ok := other.(ParseIndent); ok {
		return op == v
	}
	return false
}

func (op ParseIndent) String() string {
	return "ParseIndent"
}

func (op ParseIndent) Process(args *BindArgs) {

	type stackItem struct {
		base  int
		level int
	}

	for _, par := range args.ParentNodes() {
		nodes := par.Nodes()
		stack := []stackItem{
			{
				base:  0,
				level: nodes[0].Indent(),
			},
		}

		pop := func(end int) (level int, newEnd int) {
			size := len(stack) - 1
			head := stack[size]
			stack = stack[:size]

			sta := head.base - 1
			pos := nodes[sta].Index()

			list := par.RemoveRange(nodes[sta], nodes[end])
			block := args.Program.NewNode(Indented{}, SliceSpan(list[1:]))
			block.AddChildren(list[1:]...)

			group := args.Program.NewNode(Group{}, RangeSpan(list[0], block))
			group.AddChildren(list[0], block)

			par.InsertNodes(pos, group)

			nodes = par.Nodes()
			return stack[size-1].level, sta + 1
		}

		for index := 1; index < len(nodes); index++ {
			head := stack[len(stack)-1]
			curLevel := head.level
			newLevel := nodes[index].Indent()
			if newLevel > curLevel {
				stack = append(stack, stackItem{
					base:  index,
					level: newLevel,
				})
			} else {
				for newLevel < curLevel {
					if len(stack) == 1 {
						break
					}
					curLevel, index = pop(index - 1)
					nodes = par.Nodes()
				}

				if newLevel != curLevel {
					nodes[index].AddError("invalid indentation for line")
				}
			}
		}

		for len(stack) > 1 {
			pop(len(nodes) - 1)
		}
	}
}

type Indented struct{}

func (val Indented) Bind(node *Node) {
	node.Bind(Indented{})
}

func (val Indented) String() string {
	return "Indented"
}

func (val Indented) IsEqual(other Key) bool {
	if v, ok := other.(Indented); ok {
		return val == v
	}
	return false
}

type Group struct{}

func (val Group) Bind(node *Node) {
	node.Bind(Group{})
}

func (val Group) String() string {
	return "Group"
}

func (val Group) IsEqual(other Key) bool {
	if v, ok := other.(Group); ok {
		return val == v
	}
	return false
}
