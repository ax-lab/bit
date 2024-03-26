package bot

import "axlab.dev/bit/input"

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
