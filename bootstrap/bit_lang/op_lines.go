package bit_lang

import (
	"axlab.dev/bit/bit"
	"axlab.dev/bit/common"
)

// Implemented by non-semantic group nodes that can be "flattened" without
// losing meaning.
type CanFlatten interface {
	Flatten(node *bit.Node) []*bit.Node
}

func (val Group) Flatten(node *bit.Node) []*bit.Node {
	return node.Nodes()
}

func (val Line) Flatten(node *bit.Node) []*bit.Node {
	return node.Nodes()
}

func FlattenNodes(nodes ...*bit.Node) (out []*bit.Node) {
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

func (op SplitLines) IsSame(other bit.Binding) bool {
	if v, ok := other.(SplitLines); ok {
		return v == op
	}
	return false
}

func (op SplitLines) Precedence() bit.Precedence {
	return bit.PrecLines
}

func (op SplitLines) Process(args *bit.BindArgs) {
	for _, nodes := range args.NodesByParent() {
		par := nodes[0].Parent()
		cur, children := 0, par.RemoveNodes(0, par.Len())
		push := func(line []*bit.Node) {
			if len(line) > 0 {
				span := common.SpanFromSlice(line)
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