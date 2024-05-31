package code

import (
	"fmt"

	"axlab.dev/bit/core"
)

type Str string

func (str Str) String() string {
	return string(str)
}

func (str Str) Debug() string {
	return fmt.Sprintf("Str(%#v)", string(str))
}

func (str Str) Eval(rt *core.Runtime) (core.Value, error) {
	return str, nil
}
