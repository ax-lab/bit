package bit

import "fmt"

type Str string

func (val Str) Type() Type {
	return StrType{}
}

func (str Str) IsEqual(other Key) bool {
	if v, ok := other.(Str); ok {
		return v == str
	}
	return false
}

func (str Str) String() string {
	return string(str)
}

func (str Str) Repr(oneline bool) string {
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

func (val Str) OutputCpp(ctx *CppContext, node *Node) {
	WriteLiteralString(ctx.Expr, string(val))
}

func (val Str) OutputCppPrint(out *CppContext, node *Node) {
	out.IncludeSystem("stdio.h")
	out.Body.EnsureBlank()
	out.Body.Write(`printf("%s", `)
	WriteLiteralString(out.Body, string(val))
	out.Body.Write(`);`)
}

type ParseString struct{}

func (ParseString) Get(node *Node) (Value, error) {
	if tok, ok := node.Value().(TokenType); ok && tok == TokenString {
		str := ParseStringLiteral(node.Text())
		return Str(str), nil
	}
	return nil, nil
}
