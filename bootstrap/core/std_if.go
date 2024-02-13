package core

import "axlab.dev/bit/bit"

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
	return Code{Expr: bit.Invalid{}}
}

// TODO: (resolution) this should have precedence even over more specific bindings, we need a general mechanism for that
type ParseIf struct{}

func (op ParseIf) IsSame(other Binding) bool {
	if v, ok := other.(ParseIf); ok {
		return v == op
	}
	return false
}

func (op ParseIf) Precedence() Precedence {
	return bit.PrecParseIf
}

func (op ParseIf) Process(args *BindArgs) {
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