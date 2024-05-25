package lang

import (
	"fmt"

	"axlab.dev/bit/core"
)

func Declare(comp *core.Compiler) error {
	comp.DeclareOp(OpSourceInfo)
	comp.SetOutput(OutputSourceInfo)
	return nil
}

type sourceInfo struct {
	src core.Source
}

func (info sourceInfo) String() string {
	return fmt.Sprintf("SourceInfo(src=%s)", info.src.Name())
}

func OpSourceInfo(list core.NodeListWriter) {
	for n := 0; n < list.Len(); n++ {
		node := list.Get(n)
		if src, isSrc := node.Value().(core.Source); isSrc {
			node = core.NodeNew(node.Span(), sourceInfo{src})
			list.Set(n, node)
		}
	}
}

func OutputSourceInfo(comp *core.Compiler, list core.NodeList) {
	out := comp.StdOut()
	for n := 0; n < list.Len(); n++ {
		info := list.Get(n).Value().(sourceInfo)
		fmt.Fprintf(out, "- Source %s with %d bytes\n", info.src.Name(), len(info.src.Text()))
	}
}
