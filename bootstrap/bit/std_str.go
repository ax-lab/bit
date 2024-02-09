package bit

import "fmt"

type Str string

func (str Str) IsEqual(other Key) bool {
	if v, ok := other.(Str); ok {
		return v == str
	}
	return false
}

func (str Str) String() string {
	return string(str)
}

func (str Str) Repr() string {
	return fmt.Sprintf("Str(%v)", string(str))
}

func (str Str) Bind(node *Node) {
	node.Bind(Str(""))
}

func (val Str) Output(ctx *CodeContext) Code {
	return Code{val, nil}
}

func (val Str) Eval(rt *RuntimeContext) {
	rt.Result = val
}

type ParseString struct{}

func (ParseString) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(TokenType); ok && tok == TokenString {
		return Str(node.Text()), nil
	}
	return nil, nil
}
