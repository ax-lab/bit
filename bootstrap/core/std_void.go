package core

import "axlab.dev/bit/bit"

type Void struct{}

func (Void) Type() Type {
	return bit.NoneType{}
}

func (Void) Eval(rt *bit.RuntimeContext) {
	rt.Panic("cannot evaluate void code")
}

func (Void) OutputCpp(ctx *bit.CppContext, node *Node) {
	ctx.Expr.WriteString("((void)0)")
}

func (Void) Repr(oneline bool) string {
	return "Void"
}
