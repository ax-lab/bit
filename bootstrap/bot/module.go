package bot

import (
	"sync"

	"axlab.dev/bit/input"
)

type Module struct {
	data *moduleData
}

func (mod Module) Valid() bool {
	return mod.data != nil
}

type moduleData struct {
	mutex   sync.Mutex
	program *Program
	src     input.Source
	tokens  []Token
}
