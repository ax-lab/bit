package bot

import (
	"fmt"
	"strings"

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

type TokenIntBase string

type Token struct {
	Kind TokenKind
	Span input.Span
}

func Lex(cursor *input.Cursor, symbols *SymbolTable) (out []Token, err error) {
	for {
		cursor.SkipSpaces()
		if cursor.Empty() {
			break
		}

		var kind TokenKind
		span := cursor.Span()
		text := cursor.Text()

		if text[0] == '\r' || text[0] == '\n' {
			kind = TokenBreak
			if strings.HasPrefix(text, "\r\n") {
				cursor.Advance(2)
			} else {
				cursor.Advance(1)
			}
		} else if word := LexWord(cursor); word > 0 {
			kind = TokenWord
		} else if dec, err := LexDecimal(cursor); dec > 0 || err != nil {
			if err != nil {
				return nil, err
			}
			kind = TokenInt
		} else if str, err := LexString(cursor); str > 0 || err != nil {
			if err != nil {
				return nil, err
			}
			kind = TokenStr
		} else if len := LexComment(cursor); len > 0 {
			kind = TokenComment
		} else if sym := symbols.Read(cursor); sym != "" {
			kind = TokenSymbol
		}

		if kind == "" {
			cursor.Read()
			span = span.ExtendedTo(cursor)
			return nil, span.NewError("invalid token")
		}

		size := cursor.Offset() - span.Sta()
		if size == 0 {
			panic("Lexer: token with zero size")
		}

		span = span.Range(0, size)

		token := Token{kind, span}
		out = append(out, token)
	}

	return
}

func LexDecimal(cursor *input.Cursor) (int, error) {
	count := 0
	if !input.IsDigit(cursor.Peek()) {
		return 0, nil
	}

	for !cursor.Empty() {
		next := cursor.SkipWhile(input.IsDigit)
		next += cursor.SkipWhile(func(chr rune) bool { return chr == '_' })
		if next > 0 {
			count += next
			continue
		}

		suffix := LexWordChars(cursor)
		if suffix > 0 {
			return 0, fmt.Errorf("invalid characters in decimal int")
		}

		if next+suffix == 0 {
			break
		}
	}
	return count, nil
}

func LexWord(cursor *input.Cursor) int {
	if input.IsDigit(cursor.Peek()) {
		return 0
	}
	return LexWordChars(cursor)
}

func LexWordChars(cursor *input.Cursor) int {
	return cursor.SkipWhile(input.IsWord)
}
