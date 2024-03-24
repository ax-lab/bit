package bot

import (
	"axlab.dev/bit/input"
)

type Module struct {
	data *moduleData
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

type moduleData struct {
	program *Program
	nodes   NodeList
}
