package bit_core

import (
	"fmt"

	"axlab.dev/bit/bit"
)

type NodeReplacer interface {
	Get(node *Node) (Value, error)
}

type Replace struct {
	Pred NodeReplacer
}

func (op Replace) IsSame(other bit.Binding) bool {
	if v, ok := other.(Replace); ok {
		return v == op
	}
	return false
}

func (op Replace) Precedence() bit.Precedence {
	return bit.PrecReplace
}

func (op Replace) Process(args *bit.BindArgs) {
	for _, it := range args.Nodes {
		if v, err := op.Pred.Get(it); err != nil {
			args.Program.HandleError(err)
		} else if v != nil {
			it.ReplaceWithValue(v)
		} else {
			it.Undo()
		}
	}
}

func (op Replace) String() string {
	return fmt.Sprintf("Replace(%v)", op.Pred)
}
