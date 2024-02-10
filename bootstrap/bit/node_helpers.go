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
