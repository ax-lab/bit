package bit

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

func (val If) Output(ctx *CodeContext) Code {
	return Code{Invalid{}, nil}
}

type ParseIf struct{}

func (op ParseIf) IsSame(other Binding) bool {
	if v, ok := other.(ParseIf); ok {
		return v == op
	}
	return false
}

func (op ParseIf) Precedence() Precedence {
	return PrecIf
}

func (op ParseIf) Process(args *BindArgs) {
	for _, it := range args.ReverseNodes() {
		par, idx := it.Parent(), it.Index()
		if par == nil || it.Next() == nil {
			it.Undo()
			continue
		}

		src := par.RemoveNodes(idx, par.Len())
		expr := src[1:]
		node := args.Program.NewNode(If{}, SpanFromSlice(src))
		if _, ok := expr[len(expr)-1].Value().(Block); ok {
			expr, body := expr[:len(expr)-1], expr[len(expr)-1]
			node.AddChildren(
				args.Program.NewNode(Group{}, SpanFromSlice(expr)).WithChildren(expr...),
				body,
			)
		} else {
			split := LastSymbolIndex(expr, ":")
			if split < 0 {
				node.AddError("invalid `if` expression is missing `:`")
				continue
			}

			expr[split].FlagDone()
			expr, body := expr[:split], expr[split+1:]
			if len(body) == 0 {
				expr[split].AddError("missing `if` body after `:`")
				continue
			}
			node.AddChildren(
				args.Program.NewNode(Group{}, SpanFromSlice(expr)).WithChildren(expr...),
				args.Program.NewNode(Group{}, SpanFromSlice(body)).WithChildren(body...),
			)
		}
		par.InsertNodes(idx, node)
	}
}

func (op ParseIf) String() string {
	return "ParseIf"
}
