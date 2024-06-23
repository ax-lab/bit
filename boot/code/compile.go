package code

import (
	"fmt"
	"log"
)

type EvalFunc func(rt *Runtime) (any, error)

func MustCompile(expr Expr) EvalFunc {
	out, err := Compile(expr)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func Compile(expr Expr) (EvalFunc, error) {
	switch expr.Value().(type) {
	default:
		return nil, fmt.Errorf("cannot compile expression: %s", expr)
	}
}
