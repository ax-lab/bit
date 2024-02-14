package core

import (
	"fmt"

	"axlab.dev/bit/bit"
)

func ToBool(res bit.Result) bool {
	switch v := res.(type) {
	case Bool:
		return bool(v)
	case Str:
		return string(v) != ""
	case Int:
		return v != 0
	default:
		return true
	}
}

type Bool bool

func (val Bool) Type() Type {
	return bit.BoolType{}
}

func (val Bool) IsEqual(other Key) bool {
	if v, ok := other.(Bool); ok {
		return v == val
	}
	return false
}

func (val Bool) String() string {
	if val {
		return "true"
	} else {
		return "false"
	}
}

func (val Bool) Repr(oneline bool) string {
	return fmt.Sprintf("Bool(%s)", val.String())
}

func (val Bool) Bind(node *Node) {
	node.Bind(Bool(false))
}

func (val Bool) Output(ctx *bit.CodeContext) Code {
	return Code{Expr: val}
}

func (val Bool) Eval(rt *bit.RuntimeContext) {
	rt.Result = val
}

func (val Bool) OutputCpp(ctx *bit.CppContext, node *Node) {
	ctx.IncludeSystem("stdbool.h")
	ctx.Expr.WriteString(val.String())
}

func (val Bool) OutputCppPrint(ctx *bit.CppContext, node *Node) {
	ctx.IncludeSystem("stdio.h")
	ctx.Body.EnsureBlank()
	ctx.Body.Write(`printf("`)
	ctx.Body.Write(val.String())
	ctx.Body.Write(`");`)
}

type ParseBool struct{}

func (ParseBool) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(bit.TokenType); ok && tok == bit.TokenWord {
		switch node.Text() {
		case "true":
			return Bool(true), nil
		case "false":
			return Bool(false), nil
		}
	}
	return nil, nil
}
