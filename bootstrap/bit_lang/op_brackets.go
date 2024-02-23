package bit_lang

import (
	"fmt"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/common"
)

type Bracket struct {
	Sta string
	End string
}

func (val Bracket) Bind(node *bit.Node) {}

func (val Bracket) Repr(oneline bool) string {
	return fmt.Sprintf("Bracket(`%s%s`)", val.Sta, val.End)
}

type ParseBrackets struct {
	Sta string
	End string
}

func (op ParseBrackets) Precedence() bit.Precedence {
	return bit.PrecBrackets
}

func (op ParseBrackets) IsSame(other bit.Binding) bool {
	if v, ok := other.(ParseBrackets); ok {
		return op == v
	}
	return false
}

func (op ParseBrackets) String() string {
	return fmt.Sprintf("ParseBrackets(`%s%s`)", op.Sta, op.End)
}

func (op ParseBrackets) Process(args *bit.BindArgs) {
	var stack []*bit.Node
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
				group := args.Program.NewNode(Bracket(op), common.SpanFromRange(sta, it))
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
