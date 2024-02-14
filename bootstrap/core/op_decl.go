package core

import (
	"fmt"

	"axlab.dev/bit/bit"
)

type Var struct {
	Var *bit.Variable
}

func (val Var) Type() Type {
	return val.Var.Type
}

func (val Var) IsEqual(other Key) bool {
	if v, ok := other.(Var); ok {
		return v == val
	}
	return false
}

func (val Var) Repr(oneline bool) string {
	if val.Var == nil {
		return "Var()"
	}
	return val.Var.String()
}

func (val Var) Bind(node *Node) {
	node.Bind(Var{})
}

func (val Var) Output(ctx *bit.CodeContext) Code {
	return Code{Expr: val}
}

func (val Var) Eval(rt *bit.RuntimeContext) {
	res := val.Var.Value()
	if res == nil {
		rt.Panic("variable `%s` has not been initialized", val.Var.Name)
	} else {
		rt.Result = res
	}
}

func (val Var) OutputCpp(ctx *bit.CppContext, node *Node) {
	ctx.Expr.WriteString(val.Var.EncodedName())
}

// TODO: CppPrint needs better support for sub expressions

func (val Var) OutputCppPrint(ctx *bit.CppContext, node *Node) {
	typ := val.Type()
	if prn, ok := typ.(PrintCpp); ok {
		prn.OutputCppPrint(ctx, node)
	} else {
		ctx.Body.Push("#error type `%s` for variable `%s` does not support print", typ.String(), val.Repr(true))
	}
}

type BindVar struct {
	Var *bit.Variable
}

func (op BindVar) IsSame(other bit.Binding) bool {
	if v, ok := other.(BindVar); ok {
		return v == op
	}
	return false
}

func (op BindVar) Precedence() bit.Precedence {
	return bit.PrecVar
}

func (op BindVar) Process(args *bit.BindArgs) {
	for _, it := range args.Nodes {
		it.ReplaceWithValue(Var(op))
	}
}

func (op BindVar) String() string {
	return fmt.Sprintf("BindVar(%s)", op.Var.Name)
}

type Let struct {
	Var *bit.Variable
}

func (val Let) IsEqual(other Key) bool {
	if v, ok := other.(Let); ok {
		return v == val
	}
	return false
}

func (val Let) Repr(oneline bool) string {
	if val.Var == nil {
		return "Let()"
	}
	return fmt.Sprintf("Let(%s)", val.Var.Name)
}

func (val Let) Bind(node *Node) {
	node.Bind(Let{})
}

func (val Let) Output(ctx *bit.CodeContext) Code {
	expr := ctx.OutputChild(ctx.Node)
	val.Var.Type = expr.Type()
	val.Var.EncodedName() // generate name
	return Code{Expr: LetExpr{val.Var, expr}}
}

type LetExpr struct {
	Var  *bit.Variable
	Expr Code
}

func (code LetExpr) Type() Type {
	return code.Var.Type
}

func (code LetExpr) Eval(rt *bit.RuntimeContext) {
	rt.Result = rt.Eval(code.Expr)
	code.Var.SetValue(rt.Result)
}

func (code LetExpr) OutputCpp(ctx *bit.CppContext, node *Node) {
	// TODO: handle keywords here
	name := code.Var.EncodedName()

	expr := bit.CppContext{}
	expr.NewExpr(ctx)
	code.Expr.OutputCpp(&expr)

	ctx.Body.Push("%s = %s;", name, expr.Expr.String())
	ctx.Expr.WriteString(name)
}

func (code LetExpr) Repr(oneline bool) string {
	return fmt.Sprintf("Let(%s) = %s", code.Var.Name, code.Expr.Repr(oneline))
}

type ParseLet struct{}

func (op ParseLet) IsSame(other bit.Binding) bool {
	if v, ok := other.(ParseLet); ok {
		return v == op
	}
	return false
}

func (op ParseLet) Precedence() bit.Precedence {
	return bit.PrecLet
}

func (op ParseLet) Process(args *bit.BindArgs) {
	for _, it := range args.Nodes {
		if it.Index() != 0 || it.Parent() == nil {
			it.Undo()
			continue
		}

		name, next := ParseName(it.Next())
		if name == "" {
			it.Undo()
			continue
		}

		if !IsSymbol(next, "=") {
			it.Undo()
			continue
		}

		if next.Next() == nil {
			next.AddError("expected expression after `=`")
		}

		par := it.Parent()
		split := next.Index() + 1
		nodes := par.RemoveNodes(it.Index(), par.Len())
		nodesSpan := SpanFromSlice(nodes)

		scope := par.GetScope()
		offset := nodesSpan.End()
		variable := scope.Declare(it, name, offset)

		node := args.Program.NewNode(Let{variable}, nodesSpan)
		variable.Decl = node

		par.InsertNodes(it.Index(), node)

		expr := nodes[split:]
		node.AddChildren(expr...)
		node.FlagDone()

		for _, it := range nodes[:split] {
			it.FlagDone()
		}

		it.DeclareAt(Word(name), variable.Offset, scope.End(), BindVar{variable})
	}
}

func (op ParseLet) String() string {
	return "ParseLet"
}
