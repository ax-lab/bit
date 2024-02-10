package bit

import "fmt"

type Var struct {
	Name string
	Decl *Node
}

func (val Var) IsEqual(other Key) bool {
	if v, ok := other.(Var); ok {
		return v == val
	}
	return false
}

func (val Var) Repr(oneline bool) string {
	if val.Name == "" && val.Decl == nil {
		return "Var()"
	}
	return fmt.Sprintf("Var(%s@%s)", val.Name, val.Decl.Span().Location().String())
}

func (val Var) Bind(node *Node) {
	node.Bind(Var{})
}

type BindVar struct {
	Var *Variable
}

func (op BindVar) IsSame(other Binding) bool {
	if v, ok := other.(BindVar); ok {
		return v == op
	}
	return false
}

func (op BindVar) Precedence() Precedence {
	return PrecVar
}

func (op BindVar) Process(args *BindArgs) {
	for _, it := range args.Nodes {
		it.ReplaceWithValue(Var{Name: op.Var.Name, Decl: op.Var.Decl})
	}
}

func (op BindVar) String() string {
	return fmt.Sprintf("BindVar(%s)", op.Var.Name)
}

type Let struct {
	Var *Variable
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
		if it.Index() != 0 || it.Parent() == nil {
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

		it.DeclareAt(Word(name), variable.Offset, scope.End, BindVar{variable})
	}
}

func (op ParseLet) String() string {
	return "ParseLet"
}
