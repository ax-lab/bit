package core

import (
	"fmt"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/common"
)

type Print struct{}

func (val Print) IsEqual(other Key) bool {
	if v, ok := other.(Print); ok {
		return v == val
	}
	return false
}

func (val Print) Repr(oneline bool) string {
	return "Print"
}

func (val Print) Bind(node *Node) {
	node.Bind(Print{})
}

func (val Print) Output(ctx *CodeContext) Code {
	code := ctx.OutputChild(ctx.Node)
	return Code{Expr: PrintExpr{code}}
}

type PrintCpp interface {
	OutputCppPrint(ctx *CppContext, node *Node)
}

type ParsePrint struct{}

func (op ParsePrint) IsSame(other Binding) bool {
	if v, ok := other.(ParsePrint); ok {
		return v == op
	}
	return false
}

func (op ParsePrint) Precedence() Precedence {
	return bit.PrecPrint
}

func (op ParsePrint) Process(args *BindArgs) {
	for _, it := range args.Nodes {
		par, idx := it.Parent(), it.Index()
		if par == nil {
			it.Undo()
			continue
		}
		src := par.RemoveNodes(idx, par.Len())
		node := args.Program.NewNode(Print{}, bit.SpanFromSlice(src))
		node.AddChildren(src[1:]...)
		par.InsertNodes(idx, node)
	}
}

func (op ParsePrint) String() string {
	return "ParsePrint"
}

type PrintExpr struct {
	args Code
}

func (expr PrintExpr) Type() Type {
	return expr.args.Type()
}

func (expr PrintExpr) Eval(rt *RuntimeContext) {
	rt.Result = rt.Eval(expr.args)
	rt.OutputStd(rt.Result.String())
	rt.OutputStd("\n")
}

func (val PrintExpr) OutputCpp(ctx *CppContext, node *Node) {
	if v, ok := val.args.Expr.(PrintCpp); ok {
		ctx.IncludeSystem("stdio.h")
		v.OutputCppPrint(ctx, val.args.Node)
		ctx.Body.Push(`printf("\n");`)
	} else {
		ctx.Body.Push("#error Cannot output print for `%s`", val.args.Expr.Repr(true))
	}
}

func (expr PrintExpr) Repr(oneline bool) string {
	return fmt.Sprintf("print(%s)", common.Indented(expr.args.Repr(oneline)))
}
