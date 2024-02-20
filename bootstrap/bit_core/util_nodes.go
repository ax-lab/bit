package bit_core

import "axlab.dev/bit/bit"

func Succ(node *bit.Node) *bit.Node {
	if next := node.Next(); next != nil {
		return EnterGroup(next)
	}
	if par := node.Parent(); par != nil {
		return Succ(par)
	}
	return nil
}

func EnterGroup(node *bit.Node) *bit.Node {
	if node.Len() > 0 {
		if _, ok := node.Value().(CanFlatten); ok {
			return EnterGroup(node.Head())
		}
	}
	return node
}

func ParseName(node *bit.Node) (name string, next *bit.Node) {
	if IsName(node) {
		return node.Text(), node.Next()
	}
	return
}

func IsName(node *bit.Node) bool {
	if node != nil {
		if v, ok := node.Value().(bit.TokenType); ok {
			return v == bit.TokenWord
		}
	}
	return false
}

func IsSymbol(node *bit.Node, symbol string) bool {
	if node != nil {
		if v, ok := node.Value().(bit.TokenType); ok {
			return v == bit.TokenSymbol && node.Text() == symbol
		}
	}
	return false
}

func IsWord(node *bit.Node, word string) bool {
	if node != nil {
		if v, ok := node.Value().(bit.TokenType); ok {
			return v == bit.TokenWord && node.Text() == word
		}
	}
	return false
}

func SymbolIndex(nodes []*bit.Node, symbol string) int {
	for n, it := range nodes {
		if IsSymbol(it, symbol) {
			return n
		}
	}
	return -1
}

func LastSymbolIndex(nodes []*bit.Node, symbol string) int {
	for n := len(nodes) - 1; n >= 0; n-- {
		if it := nodes[n]; IsSymbol(it, symbol) {
			return n
		}
	}
	return -1
}

func WordIndex(nodes []*bit.Node, word string) int {
	for n, it := range nodes {
		if IsWord(it, word) {
			return n
		}
	}
	return -1
}

func LastWordIndex(nodes []*bit.Node, word string) int {
	for n := len(nodes) - 1; n >= 0; n-- {
		if it := nodes[n]; IsWord(it, word) {
			return n
		}
	}
	return -1
}
