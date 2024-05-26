package lang

import (
	"axlab.dev/bit/core"
)

func Declare(comp *core.Compiler) error {
	comp.DeclareOp(OpSegment)
	comp.SetOutput(NoOp)
	return nil
}

func NoOp(list core.NodeList) {
}
