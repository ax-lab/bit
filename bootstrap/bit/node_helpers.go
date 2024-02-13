package bit

func (node *Node) ParseName() (name string, next *Node) {
	if node.IsName() {
		return node.Text(), node.Next()
	}
	return
}

func (node *Node) IsName() bool {
	if node != nil {
		if v, ok := node.Value().(TokenType); ok {
			return v == TokenWord
		}
	}
	return false
}

func (node *Node) IsSymbol(symbol string) bool {
	if node != nil {
		if v, ok := node.Value().(TokenType); ok {
			return v == TokenSymbol && node.Text() == symbol
		}
	}
	return false
}

func (node *Node) IsWord(word string) bool {
	if node != nil {
		if v, ok := node.Value().(TokenType); ok {
			return v == TokenWord && node.Text() == word
		}
	}
	return false
}

func SymbolIndex(nodes []*Node, symbol string) int {
	for n, it := range nodes {
		if it.IsSymbol(symbol) {
			return n
		}
	}
	return -1
}

func LastSymbolIndex(nodes []*Node, symbol string) int {
	for n := len(nodes) - 1; n >= 0; n-- {
		if it := nodes[n]; it.IsSymbol(symbol) {
			return n
		}
	}
	return -1
}

func WordIndex(nodes []*Node, word string) int {
	for n, it := range nodes {
		if it.IsWord(word) {
			return n
		}
	}
	return -1
}

func LastWordIndex(nodes []*Node, word string) int {
	for n := len(nodes) - 1; n >= 0; n-- {
		if it := nodes[n]; it.IsWord(word) {
			return n
		}
	}
	return -1
}
