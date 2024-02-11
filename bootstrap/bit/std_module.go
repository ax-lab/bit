package bit

import "fmt"

type Module struct {
	Source *Source
}

func (mod Module) IsScope(node *Node) bool {
	return true
}

func (mod Module) IsEqual(other Key) bool {
	if v, ok := other.(Module); ok {
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

func (mod Module) Output(ctx *CodeContext) Code {
	return ctx.OutputChildren(ctx.Node)
}
