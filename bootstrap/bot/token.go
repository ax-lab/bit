package bot

import "axlab.dev/bit/input"

type TokenKind string

const (
	TokenInt     TokenKind = "Int"
	TokenStr     TokenKind = "Str"
	TokenWord    TokenKind = "Word"
	TokenSymbol  TokenKind = "Symbol"
	TokenBreak   TokenKind = "Break"
	TokenComment TokenKind = "Comment"
)

type Token struct {
	Kind TokenKind
	Span input.Span
}

func (tok Token) Text() string {
	return tok.Span.Text()
}

type TokenList struct {
	src   input.Source
	items []Token
	sta   int
	end   int
}

func TokenListNew(src input.Source, items []Token) TokenList {
	return TokenList{src, items, 0, len(items)}
}

func (ls TokenList) Src() input.Source {
	return ls.src
}

func (ls TokenList) Len() int {
	return ls.end - ls.sta
}

func (ls TokenList) Get(index int) Token {
	index += ls.sta
	if index < ls.sta || ls.end <= index {
		panic("TokenList: index out of bounds")
	}
	return ls.items[index]
}

func (ls TokenList) Range(pos ...int) TokenList {
	sta, end := ls.getRange(pos...)
	if sta < 0 || end < sta || end > ls.Len() {
		panic("TokenList: out of bounds range")
	}
	return TokenList{items: ls.items, sta: ls.sta + sta, end: ls.sta + end}
}

func (ls TokenList) Slice(pos ...int) []Token {
	sta, end := ls.getRange(pos...)
	return ls.items[sta:end]
}

func (ls TokenList) getRange(pos ...int) (sta, end int) {
	sta, end = 0, ls.Len()
	if len(pos) > 0 {
		sta = pos[0]
		if len(pos) > 1 {
			end = pos[1]
		}
	}

	if sta < 0 || end < sta || end > ls.Len() {
		panic("TokenList: out of bounds range")
	}
	return
}

func (ls TokenList) Span() input.Span {
	list := ls.items
	size := len(list)
	if size == 0 {
		return ls.src.Span().WithLen(0)
	}

	if ls.sta >= size {
		span := list[size-1].Span
		return span.Range(span.Len(), span.Len())
	}

	if ls.sta == ls.end {
		span := list[ls.sta].Span
		return span.WithLen(0)
	}

	sta := list[ls.sta].Span
	end := list[ls.end-1].Span
	return sta.Merged(end)
}
