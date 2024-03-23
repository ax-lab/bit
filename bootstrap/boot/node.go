package boot

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"axlab.dev/bit/input"
)

type Node struct {
	inner *nodeInner
}

func (node Node) Span() input.Span {
	return node.inner.span
}

func (node Node) Offset() int {
	return node.inner.span.Sta()
}

func (node Node) Value() Value {
	return node.inner.value
}

func (node Node) Type() Type {
	return node.inner.typ
}

func (node Node) Keys() (out []Key) {
	if val, ok := node.Value().(WithKey); ok {
		out = []Key{val.Key()}
	} else if val, ok := node.Value().(WithKeys); ok {
		out = val.Keys()
	}
	return
}

func (node Node) Cmp(other Node) int {
	if res := node.Span().Cmp(other.Span()); res != 0 {
		return res
	}
	if res := node.Type().Cmp(other.Type()); res != 0 {
		return res
	}

	ka := node.Keys()
	kb := other.Keys()
	for i := 0; i < max(len(ka), len(kb)); i++ {
		if i >= len(ka) {
			return -1
		}
		if i >= len(kb) {
			return +1
		}
		if res := ka[i].Cmp(kb[i]); res != 0 {
			return res
		}
	}

	return 0
}

func (node Node) SetDone(done bool) {
	node.inner.done = done
}

func (nm *nodeMap) CheckDone() error {
	pending := make(map[Type][]Node)
	count := 0
	for _, it := range nm.allNodes {
		if !it.inner.done {
			it.inner.done = true
			key := it.Type()
			pending[key] = append(pending[key], it)
			count++
		}
	}

	if count > 0 {
		out := strings.Builder{}
		out.WriteString("compilation finished with ")
		if count == 1 {
			out.WriteString("1 unparsed node")
		} else {
			out.WriteString(fmt.Sprintf("%d unparsed nodes", count))
		}
		out.WriteString(":\n")

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
	typ   Type
	span  input.Span
	done  bool
}
