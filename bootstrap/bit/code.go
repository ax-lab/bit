package bit

type IsCode interface {
	Output(ctx *CodeContext) Code
}

type Expr interface {
	Eval(rt *RuntimeContext)
	Repr(oneline bool) string
	Type() Type
	OutputCpp(ctx *CppContext, node *Node)
}

type Code struct {
	Expr Expr
	Node *Node
}

func (code Code) Type() Type {
	return code.Expr.Type()
}

func (code Code) Span() Span {
	return code.Node.Span()
}

func (code Code) Repr(oneline bool) string {
	if code.Expr == nil {
		return "(no code)"
	}
	return code.Expr.Repr(oneline)
}

func (code Code) OutputCpp(ctx *CppContext) {
	ctx.Expr.Reset()
	code.Expr.OutputCpp(ctx, code.Node)
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

func (code Code) wrapScope(node *Node) (out Code) {
	if _, ok := code.Expr.(WithScope); ok {
		return code
	}
	if scope := node.getOwnScope(); scope != nil && scope.Len() > 0 {
		expr := WithScope{scope, code, nil}
		expr.initVars()
		code.Expr = expr
	}
	return code
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

	if out.Node == nil {
		out.Node = node
	}

	if out.Expr == nil {
		out.Expr = Invalid{}
	}

	return out.wrapScope(node)
}

func (ctx *CodeContext) OutputChild(node *Node) (out Code) {
	nodes := node.Nodes()
	if len(nodes) == 0 {
		out = Code{Sequence{}, node}
	} else if len(nodes) == 1 {
		out = ctx.Output(nodes[0])
	} else {
		node.AddError("node `%s` cannot have multiple children", node.Value().Repr(true))
		out = Code{}
	}
	return out.wrapScope(node)
}

func (ctx *CodeContext) OutputChildren(node *Node) (out Code) {
	list := make([]Code, 0, node.Len())
	for _, it := range node.Nodes() {
		if !ctx.Valid {
			break
		}
		list = append(list, ctx.Output(it))
	}

	out = Code{Sequence{list}, node}
	return out.wrapScope(node)
}

type Invalid struct{}

func (Invalid) Type() Type {
	return InvalidType{}
}

func (Invalid) Eval(rt *RuntimeContext) {
	rt.Panic("cannot evaluate invalid code")
}

func (Invalid) OutputCpp(ctx *CppContext, node *Node) {
	ctx.Body.Push("#error Trying to output invalid code")
}

func (Invalid) Repr(oneline bool) string {
	return "Invalid"
}
