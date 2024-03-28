package bot

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"

	"axlab.dev/bit/input"
)

type HasTextRepr interface {
	OutputRepr(repr *ReprWriter)
}

type ReprArg interface {
	IsReprArg()
}

type ReprMaxLen int
type ReprColumn int
type ReprPrefix string
type ReprSuffix string

func (ReprMaxLen) IsReprArg() {}
func (ReprColumn) IsReprArg() {}
func (ReprPrefix) IsReprArg() {}
func (ReprSuffix) IsReprArg() {}

func ReprNodes(out io.Writer, nodes ...Node) (n int, err error) {
	repr := ReprWriter{output: out}
	for n, it := range nodes {
		if n > 0 {
			repr.Write("\n")
		}
		repr.OutputNode(it)
	}
	return repr.total, repr.err
}

type ReprWriter struct {
	total   int
	line    int
	err     error
	output  io.Writer
	indent  int
	hasText bool
	cycle   map[Node]bool
}

func ReprNew(output io.Writer) *ReprWriter {
	return &ReprWriter{output: output}
}

func (repr *ReprWriter) OutputNode(node Node) {
	isCycle := repr.cycle[node]
	if repr.cycle == nil {
		repr.cycle = make(map[Node]bool)
	}
	repr.cycle[node] = true

	if isCycle {
		repr.Write("DUP{%s @ %s}", node.Repr(), node.Span().Location())
		return
	}

	if it, hasRepr := node.(HasTextRepr); hasRepr {
		it.OutputRepr(repr)
	} else {
		repr.Header(node)
	}
	repr.Write(" at %s", node.Span().Location())
	repr.Snippet(node.Span().Text(), ReprPrefix(" = `"), ReprSuffix("`"), ReprColumn(60), ReprMaxLen(40))
}

func (repr *ReprWriter) Header(node Node, fields ...map[string]any) {
	var keyVals [][2]string
	for _, it := range fields {
		for key, val := range it {
			keyVals = append(keyVals, [2]string{key, fmt.Sprint(val)})
		}
	}

	slices.SortStableFunc(keyVals, func(a, b [2]string) int {
		return cmp.Compare(a[0], b[0])
	})

	repr.Write(node.Repr())
	if len(keyVals) > 0 {
		repr.Indent()
		repr.Write("(")

		for _, it := range keyVals {
			key, val := it[0], it[1]
			repr.Write("\n%s: ", key)
			repr.Indent()
			repr.Write(val)
			repr.Dedent()
		}

		repr.Dedent()
		repr.Write("\n)")
	}
}

func (repr *ReprWriter) Snippet(text string, args ...ReprArg) {
	var (
		maxLen int
		prefix string
		suffix string
		column int
	)

	for _, it := range args {
		switch val := it.(type) {
		case ReprMaxLen:
			maxLen = int(val)
		case ReprPrefix:
			prefix = string(val)
		case ReprSuffix:
			suffix = string(val)
		case ReprColumn:
			column = int(val)
		default:
			panic("Snippet: invalid ReprArg")
		}
	}

	lines := input.Lines(strings.TrimSpace(text))
	pre, pos := lines[0], ""
	if len(lines) > 1 {
		pos = lines[len(lines)-1]
	}

	if maxLen > 0 && pos == "" && len(pre) > maxLen {
		mid := maxLen / 2
		pos = pre[mid:]
		pre = pre[:mid]
	}

	pre = strings.TrimSpace(pre)
	pos = strings.TrimSpace(pos)

	if maxLen > 0 && len(pre)+len(pos) > maxLen {
		half := (maxLen - 1) / 2
		pos = pos[len(pos)-min(half, len(pos)):]

		rest := maxLen - len(pos) - 1
		pre = pre[:min(rest, len(pre))]
	}

	if len(pre)+len(pos) > 0 {
		if column > 0 && repr.line < column {
			repr.Write(strings.Repeat(" ", column-repr.line))
		}
		repr.Write(prefix)
		repr.Write(pre)
		if len(pos) > 0 {
			repr.Write("…%s", pos)
		}
		repr.Write(suffix)
	}
}

func (repr *ReprWriter) Items(nodes []Node, args ...ReprArg) {
	var (
		prefix string
		suffix string
	)

	for _, it := range args {
		switch val := it.(type) {
		case ReprPrefix:
			prefix = string(val)
		case ReprSuffix:
			suffix = string(val)
		default:
			panic("Items: invalid ReprArg")
		}
	}

	if len(nodes) > 0 {
		repr.Write(prefix)
		repr.Indent()
		for _, it := range nodes {
			repr.Write("\n")
			repr.OutputNode(it)
		}

		repr.Dedent()
		repr.Write("\n%s", suffix)
	}
}

func (repr *ReprWriter) Indent() {
	repr.indent++
}

func (repr *ReprWriter) Dedent() {
	repr.indent--
}

func (repr *ReprWriter) Write(txt string, args ...any) {
	if len(args) > 0 {
		txt = fmt.Sprintf(txt, args...)
	}

	indent := func() {
		for i := 0; i < repr.indent; i++ {
			repr.writeChunk("    ")
		}
	}

	lines := input.Lines(txt)
	for n, it := range lines {
		if n > 0 {
			repr.writeChunk("\n")
			repr.hasText = false
			repr.line = 0
		}

		if it != "" {
			if !repr.hasText {
				indent()
			}
			repr.hasText = true
			repr.writeChunk(it)
		}
	}
}

func (repr *ReprWriter) writeChunk(txt string) {
	if repr.err != nil {
		return
	}
	len, err := repr.output.Write([]byte(txt))
	repr.total += len
	repr.line += len
	repr.err = err
}
