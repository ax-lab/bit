package code

import "fmt"

type Number struct {
	Value int64
}

func (expr Number) IsExpr() {}

func (expr Number) String() string {
	return fmt.Sprintf("Number(%d)", expr.Value)
}

type typeNumber struct{}

func (typeNumber) IsType() {}

func (typeNumber) String() string {
	return "Number"
}
