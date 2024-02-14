package core

import (
	"fmt"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/common"
)

type If struct{}

func (val If) IsEqual(other Key) bool {
	if v, ok := other.(If); ok {
		return v == val
	}
	return false
}

func (val If) Repr(oneline bool) string {
	return "If"
}

func (val If) Bind(node *Node) {
	node.Bind(If{})
}

func (val If) Output(ctx *bit.CodeContext) Code {
	common.Assert(2 <= ctx.Node.Len() && ctx.Node.Len() <= 3, "invalid if node length")
	out := IfExpr{
		Cond: ctx.Output(ctx.Node.Get(0)),
		If:   ctx.Output(ctx.Node.Get(1)),
	}
	if ctx.Node.Len() == 3 {
		out.Else = ctx.Output(ctx.Node.Get(2))
	} else {
		out.Else = Code{Expr: Void{}, Node: ctx.Node}
	}
	return Code{Expr: out}
}

// TODO: (resolution) this should have precedence even over more specific bindings, we need a general mechanism for that
type ParseIf struct{}

func (op ParseIf) IsSame(other bit.Binding) bool {
	if v, ok := other.(ParseIf); ok {
		return v == op
	}
	return false
}

func (op ParseIf) Precedence() bit.Precedence {
	return bit.PrecParseIf
}

func (op ParseIf) Process(args *bit.BindArgs) {
	// Parse in reverse order so that the innermost `if` takes precedence
	for _, it := range args.ReverseNodes() {
		par, idx := it.Parent(), it.Index()
		if par == nil || it.Next() == nil {
			it.Undo()
			continue
		}

		makeNode := func(node *Node) *Node {
			if node != nil {
				return node
			}
			src := par.RemoveNodes(idx, par.Len())
			node = args.Program.NewNode(If{}, SpanFromSlice(src))
			par.InsertNodes(idx, node)
			return node
		}

		src := par.Nodes()[idx:]
		list := src[1:]

		// parse an inline else
		var (
			node     *Node
			elseNode *Node
		)
		if split := WordIndex(list, "else"); split >= 0 {
			node = makeNode(node)
			list[split].FlagDone()
			rest := list[split+1:]
			list = list[:split]
			if len(list) == 0 || len(rest) == 0 {
				list[split].AddError("if..else branch cannot be empty")
				continue
			}
			elseNode = args.Program.NewNode(Group{}, SpanFromSlice(rest)).WithChildren(rest...)
		}

		// parse an if : block
		last := list[len(list)-1]
		if _, ok := last.Value().(Block); ok {
			node = makeNode(node)
			expr, body := list[:len(list)-1], list[len(list)-1]
			node.AddChildren(
				args.Program.NewNode(Group{}, SpanFromSlice(expr)).WithChildren(expr...),
				body,
			)
		} else if split := SymbolIndex(list, ":"); split >= 0 {
			node = makeNode(node)
			list[split].FlagDone()
			expr, body := list[:split], list[split+1:]
			if len(expr) == 0 {
				list[split].AddError("invalid `if` with empty expression")
			} else if len(body) == 0 {
				list[split].AddError("invalid `if` with empty body")
			}

			node.AddChildren(
				args.Program.NewNode(Group{}, SpanFromSlice(expr)).WithChildren(expr...),
				args.Program.NewNode(Group{}, SpanFromSlice(body)).WithChildren(body...),
			)
		} else {
			it.Undo()
			continue
		}

		if elseNode == nil {
			if next := Succ(node); next != nil && IsWord(next, "else") {
				next.FlagDone()
				parent := next.Parent()
				children := parent.RemoveNodes(next.Index(), parent.Len())[1:]
				elseNode = args.Program.NewNode(Group{}, SpanFromSlice(children)).WithChildren(children...)
			}
		}

		if elseNode != nil {
			node.AddChildren(elseNode)
		}
	}
}

func (op ParseIf) String() string {
	return "ParseIf"
}

type IfExpr struct {
	Cond Code
	If   Code
	Else Code
}

func (expr IfExpr) Type() Type {
	// TODO: this should be or of both types
	return expr.If.Type()
}

func (expr IfExpr) Eval(rt *bit.RuntimeContext) {
	cond := rt.Eval(expr.Cond)
	if bit.IsError(cond) {
		rt.Result = cond
	} else if ToBool(cond) {
		rt.Result = rt.Eval(expr.If)
	} else {
		rt.Result = rt.Eval(expr.Else)
	}
}

func (val IfExpr) OutputCpp(ctx *bit.CppContext, node *Node) {
	name := ctx.NewName("if_res")
	ctx.IncludeSystem("stdbool.h")
	ctx.Body.Decl.Push("%s %s;", val.Type().CppType(), name)

	cond := bit.CppContext{}
	cond.NewExpr(ctx)
	val.Cond.OutputCpp(&cond)

	// If

	ctx.Body.Push("if (%s) {", cond.Expr.String())
	expr_if := bit.CppContext{}
	expr_if.NewBody(ctx)
	val.If.OutputCpp(&expr_if)

	ctx.Body.Indent()
	expr_if.Body.AppendTo(&ctx.Body.CppLines)
	ctx.Body.Push("%s = %s;", name, expr_if.Expr.String())
	ctx.Body.Dedent()
	ctx.Body.Push("}")

	// Else

	ctx.Body.Push("else {")
	expr_else := bit.CppContext{}
	expr_else.NewBody(ctx)
	val.Else.OutputCpp(&expr_else)

	ctx.Body.Indent()
	expr_else.Body.AppendTo(&ctx.Body.CppLines)
	ctx.Body.Push("%s = %s;", name, expr_else.Expr.String())
	ctx.Body.Dedent()
	ctx.Body.Push("}")

	ctx.Expr.WriteString(name)
}

func (expr IfExpr) Repr(oneline bool) string {
	return fmt.Sprintf("if %s: %s else: %s", expr.Cond.Repr(true), expr.If.Repr(true), expr.Else.Repr(true))
}
