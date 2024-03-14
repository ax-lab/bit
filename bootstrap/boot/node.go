package boot

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type Node struct {
	inner *nodeInner
}

func (node Node) Span() Span {
	return node.inner.span
}

func (node Node) Offset() int {
	return node.inner.span.Sta()
}

func (node Node) Value() Value {
	return node.inner.value
}

func (node Node) Cmp(other Node) int {
	if res := node.Span().Cmp(other.Span()); res != 0 {
		return res
	}
	if res := node.Value().Type().Cmp(other.Value().Type()); res != 0 {
		return res
	}
	return 0
}

func (nm *nodeMap) CheckDone() error {
	pending := make(map[Type][]Node)
	count := 0
	for _, it := range nm.allNodes {
		if !it.inner.done {
			it.inner.done = true
			key := it.Value().Type()
			pending[key] = append(pending[key], it)
			count++
		}
	}

	if count > 0 {
		out := strings.Builder{}
		if count == 1 {
			out.WriteString("there is one pending node:\n")
		} else {
			out.WriteString(fmt.Sprintf("there are %d pending nodes:\n", count))
		}

		maxPer := max(count/len(pending), 1)
		for key, list := range pending {
			slices.SortStableFunc(list, func(a, b Node) int {
				return a.Cmp(b)
			})
			for n, it := range list {
				if n == maxPer {
					break
				}
				out.WriteString("\n\t--> ")
				out.WriteString(it.Value().Repr())
				out.WriteString("\n\t    ↳ at ")
				out.WriteString(it.Span().Location())
			}

			if diff := len(list) - maxPer; diff > 0 {
				out.WriteString("\n\n\t... And ")
				if diff == 1 {
					out.WriteString("1 other")
				} else {
					out.WriteString(fmt.Sprintf("%d others", diff))
				}
				out.WriteString(" of type ")
				out.WriteString(key.Name())
			}
		}

		return errors.New(out.String())
	}

	return nil
}

type nodeInner struct {
	value Value
	span  Span
	done  bool
}
