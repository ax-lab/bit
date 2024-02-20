package bit

import (
	"fmt"

	"axlab.dev/bit/code"
	"axlab.dev/bit/common"
)

type Module struct {
	Source *common.Source
}

func (mod Module) IsScope(node *Node) bool {
	return true
}

func (mod Module) IsEqual(val any) bool {
	if v, ok := val.(Module); ok {
		return v == mod
	}
	return false
}

func (mod Module) Bind(node *Node) {
	node.Bind(mod.Source)
	node.Bind(Module{})
}

func (mod Module) Repr(oneline bool) string {
	if mod.Source == nil {
		return "Module(nil)"
	}
	return fmt.Sprintf("Module(%s)", mod.Source.Name())
}

func (mod Module) Type(node *Node) code.Type {
	return node.Last().Type()
}

func (mod Module) Output(code *code.OutputContext, node *Node, ans *code.Variable) {
	node.OutputChildren(code, ans)
}
