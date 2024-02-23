package bit_lang

import (
	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
	"axlab.dev/bit/common"
)

type Print struct{}

func (val Print) IsEqual(other any) bool {
	if v, ok := other.(Print); ok {
		return v == val
	}
	return false
}

func (val Print) Repr(oneline bool) string {
	return "Print"
}

func (val Print) Bind(node *bit.Node) {
	node.Bind(Print{})
}

func (val Print) Type(node *bit.Node) code.Type {
	return node.Last().Type()
}

func (val Print) Output(ctx *code.OutputContext, node *bit.Node, ans *code.Variable) {
	var vars []code.Expr
	block := ctx.NewBlock()
	for _, it := range node.Nodes() {
		v := block.TempVar("p_arg", it.Type(), it)
		vars = append(vars, v)
		it.Output(block, v)
	}

	ctx.Output(block.Block())

	code := code.NewPrint(vars...)
	ctx.Output(code)
	if len(vars) > 0 {
		ctx.Output(ans.SetVar(vars[len(vars)-1]))
	}
}

type ParsePrint struct{}

func (op ParsePrint) IsSame(other bit.Binding) bool {
	if v, ok := other.(ParsePrint); ok {
		return v == op
	}
	return false
}

func (op ParsePrint) Precedence() bit.Precedence {
	return bit.PrecPrint
}

func (op ParsePrint) Process(args *bit.BindArgs) {
	for _, it := range args.Nodes {
		par, idx := it.Parent(), it.Index()
		if par == nil {
			it.Undo()
			continue
		}
		src := par.RemoveNodes(idx, par.Len())
		node := args.Program.NewNode(Print{}, common.SpanFromSlice(src))
		node.AddChildren(src[1:]...)
		par.InsertNodes(idx, node)
	}
}

func (op ParsePrint) String() string {
	return "ParsePrint"
}
