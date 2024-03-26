package bot

func NodeIsWord(node Node, word string) bool {
	if token, ok := node.(Token); ok && token.Kind() == TokenWord {
		return token.Span().Text() == word
	}
	return false
}

func NodeIsBracketSta(node Node) (is bool, kind string) {
	token, ok := node.(Token)
	if !ok || token.Kind() != TokenSymbol {
		return
	}

	switch txt := token.Span().Text(); txt {
	case "(":
		kind = "()"
	case "[":
		kind = "[]"
	case "{":
		kind = "{}"
	}

	return kind != "", kind
}

func NodeIsBracketEnd(node Node) (is bool, kind string) {
	token, ok := node.(Token)
	if !ok || token.Kind() != TokenSymbol {
		return
	}

	switch txt := token.Span().Text(); txt {
	case ")":
		kind = "()"
	case "]":
		kind = "[]"
	case "}":
		kind = "{}"
	}

	return kind != "", kind
}
