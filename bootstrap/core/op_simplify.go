package core

import "axlab.dev/bit/bit"

type Simplify struct{}

func (op Simplify) IsSame(other bit.Binding) bool {
	if v, ok := other.(Simplify); ok {
		return v == op
	}
	return false
}

func (op Simplify) Precedence() bit.Precedence {
	return bit.PrecSimplify
}

func (op Simplify) Process(args *bit.BindArgs) {
	for _, it := range args.Nodes {
		par := it.Parent()
		if par == nil {
			continue
		}

		switch it.Len() {
		case 0:
			it.Remove()
		case 1:
			it.Remove()
			nodes := it.RemoveNodes(0, it.Len())
			par.InsertNodes(it.Index(), nodes[0])
		default:
			it.AddError("node `%s` cannot have multiple children", it.Value().Repr(true))
		}
	}
}

func (op Simplify) String() string {
	return "Simplify"
}
