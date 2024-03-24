package bot

import (
	"fmt"

	"axlab.dev/bit/input"
)

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
	kind TokenKind
	span input.Span
}

func (tok Token) Kind() TokenKind {
	return tok.kind
}

func (tok Token) Span() input.Span {
	return tok.span
}

func (tok Token) Repr() string {
	return fmt.Sprintf("Token(%s)", tok.kind)
}
