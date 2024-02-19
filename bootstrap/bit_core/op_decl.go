package bit_core

import (
	"fmt"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
)

type Var struct {
	Var *code.Variable
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
	return fmt.Sprintf("Var(%s)", val.Var.Name())
}

func (val Var) Bind(node *Node) {
	node.Bind(Var{})
}

func (val Var) Type(node *Node) Type {
	return val.Var.Type()
}

func (val Var) Output(ctx *code.OutputContext, node *Node, ans *code.Variable) {
	node.CheckEmpty(ctx)
	val.Var.CheckBound()
	ctx.Output(ans.SetVar(val.Var))
}

type BindVar struct {
	Var *code.Variable
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
	return fmt.Sprintf("BindVar(%s)", op.Var.Name())
}

type Let struct {
	Var *code.Variable
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
	return fmt.Sprintf("Let(%s)", val.Var.Name())
}

func (val Let) Bind(node *Node) {
	node.Bind(Let{})
}

func (val Let) Type(node *Node) Type {
	return val.Var.Type()
}

func (val Let) Output(ctx *code.OutputContext, node *Node, ans *code.Variable) {
	decl := ctx.GetDecl()
	decl.Add(val.Var)
	val.Var.SetType(node.Get(0).Type())
	node.OutputChild(ctx, val.Var, false)
	ctx.Output(ans.SetVar(val.Var))
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
		variable := scope.Declare(name, offset)

		node := args.Program.NewNode(Let{variable}, nodesSpan)
		variable.Source = node

		par.InsertNodes(it.Index(), node)

		expr := nodes[split:]
		node.AddChildren(expr...)
		node.FlagDone()

		for _, it := range nodes[:split] {
			it.FlagDone()
		}

		it.DeclareAt(Word(name), variable.Offset(), scope.End(), BindVar{variable})
	}
}

func (op ParseLet) String() string {
	return "ParseLet"
}
