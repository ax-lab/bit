package core

type Line struct{}

func (val Line) Bind(node *Node) {
	node.Bind(Line{})
	node.Bind(Indented{})
}

func (val Line) Repr(oneline bool) string {
	return "Line"
}

func (val Line) IsEqual(other Key) bool {
	if v, ok := other.(Line); ok {
		return val == v
	}
	return false
}

func (val Line) Output(ctx *CodeContext) Code {
	return ctx.OutputChild(ctx.Node)
}
