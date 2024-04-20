package bit

import (
	"fmt"
	"slices"
	"strings"

	"axlab.dev/bit/boot/core"
)

type TokenKind string

const (
	TokenNone    TokenKind = ""
	TokenInt     TokenKind = "Int"
	TokenStr     TokenKind = "Str"
	TokenWord    TokenKind = "Word"
	TokenSymbol  TokenKind = "Symbol"
	TokenBreak   TokenKind = "Break"
	TokenComment TokenKind = "Comment"
)

func Lex(cursor *core.Cursor, lex *Lexer) (out []Token, err error) {
	for cursor.Len() > 0 {
		token, tokenErr := lex.ReadNext(cursor)
		if token.Valid() {
			out = append(out, token)
		}

		if !token.Valid() || tokenErr != nil {
			err = tokenErr
			break
		}
	}

	return
}

type Token struct {
	kind TokenKind
	span core.Span
}

func (tok Token) Valid() bool {
	return tok.kind != TokenNone
}

func (tok Token) Kind() TokenKind {
	return tok.kind
}

func (tok Token) Span() core.Span {
	return tok.span
}

type LexMatcher func(cursor *core.Cursor) (TokenKind, error)

type Lexer struct {
	symbols  symbolTable
	matchers []LexMatcher
}

func (lex *Lexer) Copy() Lexer {
	out := Lexer{}
	out.symbols = lex.symbols.Copy()
	out.matchers = append(out.matchers, lex.matchers...)
	return out
}

func (lex *Lexer) AddSymbols(symbols ...string) {
	lex.symbols.Add(symbols...)
}

func (lex *Lexer) AddMatcher(matcher LexMatcher) {
	if matcher == nil {
		panic("invalid matcher")
	}
	lex.matchers = append(lex.matchers, matcher)
}

func (lex *Lexer) ReadNext(cursor *core.Cursor) (out Token, err error) {
	cursor.SkipSpaces()
	if cursor.Empty() {
		return
	}

	var (
		tokenKind TokenKind
		tokenErr  error
	)
	span := cursor.Span()
	text := cursor.Text()

	if text[0] == '\r' || text[0] == '\n' {
		tokenKind = TokenBreak
		if strings.HasPrefix(text, "\r\n") {
			cursor.Advance(2)
		} else {
			cursor.Advance(1)
		}
	} else {
		for _, matcher := range lex.matchers {
			cur := *cursor
			tokenKind, tokenErr = matcher(&cur)
			if tokenKind != TokenNone || tokenErr != nil {
				*cursor = cur
				break
			}
		}

		if tokenKind == TokenNone {
			if sym := lex.symbols.Read(cursor); sym != "" {
				tokenKind = TokenSymbol
			}
		}
	}

	if tokenKind == "" {
		cursor.Read()
		tokenErr = fmt.Errorf("invalid token")
	}

	size := cursor.Offset() - span.Sta()
	if size == 0 {
		panic("token with size zero -- " + tokenKind)
	}

	span = span.Range(0, size)
	if tokenErr != nil {
		err = span.ErrorAt(tokenErr)
	}

	out = Token{tokenKind, span}
	return
}

type symbolTable struct {
	symbols map[string]bool
	sorted  []string
}

func (tb *symbolTable) Copy() symbolTable {
	out := symbolTable{}
	out.sorted = append(out.sorted, tb.sorted...)
	out.symbols = make(map[string]bool)
	for _, it := range out.sorted {
		out.symbols[it] = true
	}
	return out
}

func (tb *symbolTable) Add(symbols ...string) {
	for _, symbol := range symbols {
		symbol = strings.TrimSpace(symbol)
		if symbol == "" {
			panic("SymbolTable: invalid empty symbol")
		}

		if !tb.symbols[symbol] {
			if tb.symbols == nil {
				tb.symbols = make(map[string]bool)
			}
			tb.symbols[symbol] = true
			tb.sorted = append(tb.sorted, symbol)
			slices.SortFunc(tb.sorted, func(a, b string) int {
				return len(b) - len(a)
			})
		}
	}
}

func (tb *symbolTable) Read(cursor *core.Cursor) string {
	for _, sym := range tb.sorted {
		if cursor.ReadIf(sym) {
			return sym
		}
	}
	return ""
}
