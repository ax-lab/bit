package code

import "fmt"

type Str struct {
	Value string
}

func (expr Str) IsExpr() {}

func (expr Str) String() string {
	return fmt.Sprintf("Str(%#v)", expr.Value)
}

type typeStr struct{}

func (typeStr) IsType() {}

func (typeStr) String() string {
	return "Str"
}
