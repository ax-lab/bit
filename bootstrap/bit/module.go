package bit

import "fmt"

type Module struct {
	Source *Source
}

func (mod Module) Key() Key {
	return mod.Source
}

func (mod Module) String() string {
	return fmt.Sprintf("Module(%s)", mod.Source.Name())
}
