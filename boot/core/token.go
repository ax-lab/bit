package core

import "fmt"

type Invalid string

func (inv Invalid) String() string {
	return fmt.Sprintf("Invalid(%#v)", inv)
}

type Symbol string

func (sym Symbol) String() string {
	return fmt.Sprintf("Symbol(%#v)", sym)
}

type LineBreak string

func (LineBreak) String() string {
	return "LineBreak"
}
