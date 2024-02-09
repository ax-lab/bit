package bit

type IsCode interface {
	Output(ctx *CodeContext) Code
}

type Expr interface {
	Eval(rt *RuntimeContext)
	Repr(oneline bool) string
	OutputCpp(ctx *CppContext, node *Node)
}

type Code struct {
	Expr Expr
	Node *Node
}

func (code Code) Span() Span {
	return code.Node.Span()
}

func (code Code) Repr(oneline bool) string {
	return code.Expr.Repr(oneline)
}

func (code Code) OutputCpp(ctx *CppContext) {
	if ctx.OutputExpr == nil {
		expr := CppContext{}
		expr.InitExpr(ctx)
		code.Expr.OutputCpp(&expr, code.Node)
		if txt := expr.OutputExpr.Text(); txt != "" {
			ctx.OutputFunc.EndStatement()
			ctx.OutputFunc.Write(txt)
			ctx.OutputFunc.EndStatement()
		}
	} else {
		code.Expr.OutputCpp(ctx, code.Node)
	}
}

func (program *Program) CompileOutput() Code {
	node, valid := program.mainNode, program.Valid()
	if valid && node == nil {
		panic("valid program must have a main node")
	}

	if node == nil {
		node = &Node{
			program: program,
			value:   Module{program.source},
			span:    program.source.Span(),
			id:      -1,
		}
	}

	if !valid {
		return Code{Invalid{}, node}
	}

	ctx := CodeContext{
		Parent: nil,
		Node:   node,
		Valid:  true,
	}
	return ctx.Output(node)
}

type CodeContext struct {
	Parent *CodeContext
	Node   *Node
	Valid  bool
}

func (ctx *CodeContext) Output(node *Node) (out Code) {
	program := node.Program()
	ctx.Valid = ctx.Valid && program.Valid()
	if ctx.Valid {
		if code, ok := node.Value().(IsCode); ok {
			sub := CodeContext{
				Parent: ctx,
				Node:   node,
				Valid:  true,
			}
			out = code.Output(&sub)
			ctx.Valid = sub.Valid
		} else {
			node.AddError("cannot output code for node `%s`", node.Value().Repr(true))
		}
	}

	ctx.Valid = ctx.Valid && program.Valid()
	if ctx.Valid && out.Expr == nil {
		node.AddError("node generated empty output")
		ctx.Valid = false
	}

	if out.Expr == nil {
		out.Expr = Invalid{}
	}
	if out.Node == nil {
		out.Node = node
	}

	return out
}

func (ctx *CodeContext) OutputChild(node *Node) Code {
	nodes := node.Nodes()
	if len(nodes) == 0 {
		return Code{Sequence{}, node}
	} else if len(nodes) == 1 {
		return ctx.Output(nodes[0])
	} else {
		node.AddError("node `%s` cannot have multiple children", node.Value().Repr(true))
		return Code{}
	}
}

func (ctx *CodeContext) OutputChildren(node *Node) Code {
	list := make([]Code, 0, node.Len())
	for _, it := range node.Nodes() {
		if !ctx.Valid {
			break
		}
		list = append(list, ctx.Output(it))
	}

	return Code{Sequence{list}, node}
}

type Invalid struct{}

func (Invalid) Eval(rt *RuntimeContext) {
	rt.Panic("cannot evaluate invalid code")
}

func (Invalid) OutputCpp(ctx *CppContext, node *Node) {
	out := ctx.OutputFilePrefix
	out.NewLine()
	out.Write("#error Trying to output invalid code\n")
}

func (Invalid) Repr(oneline bool) string {
	return "Invalid"
}
