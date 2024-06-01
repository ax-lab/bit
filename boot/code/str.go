package code

import (
	"fmt"

	"axlab.dev/bit/core"
)

type Str struct {
	span core.Span
	text string
}

func StrNew(span core.Span, text string) Str {
	return Str{span, text}
}

func (str Str) Span() core.Span {
	return str.span
}

func (str Str) Text() string {
	return str.text
}

func (str Str) String() string {
	return str.text
}

func (str Str) Debug() string {
	return fmt.Sprintf("Str(%#v)", str.text)
}

func (str Str) Eval(rt *core.Runtime) (core.Value, error) {
	return str, nil
}
