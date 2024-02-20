package bit_core

import (
	"axlab.dev/bit/bit"
	"axlab.dev/bit/common"
)

type Indented struct{}

func (val Indented) Repr(oneline bool) string {
	return "Indented"
}

func (val Indented) IsEqual(other any) bool {
	if v, ok := other.(Indented); ok {
		return val == v
	}
	return false
}

type IndentedGroup struct{}

func (val IndentedGroup) Bind(node *bit.Node) {
	node.Bind(IndentedGroup{})
}

func (val IndentedGroup) Repr(oneline bool) string {
	return "IndentedGroup"
}

func (val IndentedGroup) IsEqual(other any) bool {
	if v, ok := other.(IndentedGroup); ok {
		return val == v
	}
	return false
}

type ParseIndent struct{}

func (op ParseIndent) Precedence() bit.Precedence {
	return bit.PrecIndent
}

func (op ParseIndent) IsSame(other bit.Binding) bool {
	if v, ok := other.(ParseIndent); ok {
		return op == v
	}
	return false
}

func (op ParseIndent) String() string {
	return "ParseIndent"
}

func (op ParseIndent) Process(args *bit.BindArgs) {

	type stackItem struct {
		base  int
		level int
	}

	args.RequeueNodes()

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
			block := args.Program.NewNode(IndentedGroup{}, common.SpanFromSlice(list[1:]))
			block.AddChildren(list[1:]...)

			group := args.Program.NewNode(Group{}, common.SpanFromRange(list[0], block))
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
