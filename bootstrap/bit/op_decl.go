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

type HasScope interface {
	IsScope(node *Node) (is bool, sta, end int)
}

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
	return PrecVar
}

func (op BindVar) Process(args *BindArgs) {
	for _, it := range args.Nodes {
		it.ReplaceWithValue(Var{Name: op.Name, Decl: op.Let})
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

		scope, sta, end := par, node.Span().End(), par.Span().End()
		for scope != nil {
			if v, ok := scope.Value().(HasScope); ok {
				if isScope, _, e := v.IsScope(scope); isScope {
					end = e
					break
				}
			}
			scope = scope.Parent()
		}
		it.DeclareAt(Word(name), sta, end, BindVar{name, node})
	}
}

func (op ParseLet) String() string {
	return "ParseLet"
}
