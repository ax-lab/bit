package bit

import "fmt"

type BindVar struct {
	Name string
	Let  *Node
}

func (op BindVar) IsSame(other Binding) bool {
	if v, ok := other.(BindVar); ok {
		return v == op
	}
	return false
}

func (op BindVar) Precedence() Precedence {
	return PrecLet
}

func (op BindVar) Process(args *BindArgs) {
	for _, it := range args.Nodes {
		it.AddError("variable binding is not implemented yet")
	}
}

func (op BindVar) String() string {
	return fmt.Sprintf("BindVar(%s)", op.Name)
}

type Let struct {
	Name string
}

func (val Let) IsEqual(other Key) bool {
	if v, ok := other.(Let); ok {
		return v == val
	}
	return false
}

func (val Let) Repr(oneline bool) string {
	return fmt.Sprintf("Let(%s)", val.Name)
}

func (val Let) Bind(node *Node) {
	node.Bind(Let{})
}

type ParseLet struct{}

func (op ParseLet) IsSame(other Binding) bool {
	if v, ok := other.(ParseLet); ok {
		return v == op
	}
	return false
}

func (op ParseLet) Precedence() Precedence {
	return PrecLet
}

func (op ParseLet) Process(args *BindArgs) {
	for _, it := range args.Nodes {
		if it.Index() != 0 {
			it.Undo()
			continue
		}

		name, next := it.Next().ParseName()
		if name == "" {
			it.Undo()
			continue
		}

		if !next.IsSymbol("=") {
			it.Undo()
			continue
		}

		if next.Next() == nil {
			next.AddError("expected expression after `=`")
		}

		par := it.Parent()
		split := next.Index() + 1
		nodes := par.RemoveNodes(it.Index(), par.Len())

		node := args.Program.NewNode(Let{name}, SpanFromSlice(nodes))
		expr := nodes[split:]
		for _, it := range nodes[:split] {
			it.FlagDone()
		}

		node.AddChildren(expr...)
		node.FlagDone()
		par.InsertNodes(it.Index(), node)

		// TODO: this is not strictly correct, need to bind to the scope
		span := it.Span()
		it.DeclareAt(Word(name), span.End(), span.Source().Len(), BindVar{name, it})
	}
}

func (op ParseLet) String() string {
	return "ParseLet"
}
