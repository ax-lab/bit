package bot

import (
	"axlab.dev/bit/input"
)

func ParsePrint(ctx ParseContext, nodes NodeList) {
	if head := nodes.Get(0); NodeIsWord(head, "print") {
		print := Print{
			head: head,
			args: nodes.Range(1),
		}
		ctx.Queue(print.args)
		ctx.Push(print)
	}
}

type Print struct {
	head Node
	args NodeList
}

func (node Print) Span() input.Span {
	return node.head.Span().Merged(node.args.Span())
}

func (node Print) Repr() string {
	return "Print"
}

func (node Print) OutputRepr(repr *ReprWriter) {
	repr.Header(node)
	repr.Items(node.args.Slice(), ReprPrefix(" ("), ReprSuffix(")"))
}

func (node Print) GoType() GoType {
	return node.args.GoType()
}

func (node Print) GoOutput(blk *GoBlock) (out GoVar) {
	nodes := node.args.Slice()

	var args []GoVar
	for _, it := range nodes {
		code, ok := it.(GoCode)
		if !ok {
			blk.AddError(it.Span().NewError("cannot output `%s` as Go code", it.Repr()))
		} else if out = code.GoOutput(blk); out != GoVarNone {
			args = append(args, out)
		}

		if blk.HasErrors() {
			break
		}
	}

	if !blk.HasErrors() && len(args) > 0 {
		blk.Import("fmt")
		blk.Push("fmt.Print(%s)", input.Join(", ", args...))
	}
	blk.Push("fmt.Println()")

	return out
}
