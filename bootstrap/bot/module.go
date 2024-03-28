package bot

import (
	"axlab.dev/bit/input"
)

type Module struct {
	data *moduleData
}

type moduleData struct {
	program *Program
	nodes   NodeList
}

func moduleNew(program *Program, nodes NodeList) Module {
	data := &moduleData{
		program: program,
		nodes:   nodes,
	}
	return Module{data}
}

func (mod Module) Valid() bool {
	return mod.data != nil
}

func (mod Module) Src() input.Source {
	return mod.data.nodes.Src()
}

func (mod Module) Name() string {
	return mod.Src().Name()
}

func (mod Module) Program() *Program {
	return mod.data.program
}

func (mod Module) Nodes() NodeList {
	return mod.data.nodes
}

func (mod Module) Cmp(other Module) int {
	return mod.Src().Cmp(other.Src())
}

func (mod Module) GoOutput(program *GoProgram, block *GoBlock) (out GoVar) {
	nodes := mod.Nodes().Slice()
	for _, node := range nodes {
		if code, ok := node.(GoCode); ok {
			out = code.GoOutput(block)
		} else {
			program.AddError(node.Span().NewError("node `%s` cannot be output as Go code", node.Repr()))
			return GoVarNone
		}
	}

	return
}
