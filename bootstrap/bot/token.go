package bot

import (
	"fmt"
	"strconv"
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

func (tok Token) GoType() GoType {
	switch tok.kind {
	case TokenInt:
		return GoTypeInt
	case TokenStr:
		return GoTypeStr
	default:
		return GoTypeNone
	}
}

func (tok Token) GoOutput(blk *GoBlock) GoVar {
	switch tok.kind {
	case TokenInt:
		return blk.Expr("%d", TokenParseInt(tok.Span().Text()))
	case TokenStr:
		return blk.Expr("%#v", TokenParseStr(tok.Span().Text()))
	case TokenComment:
		return GoVarNone
	default:
		blk.AddError(fmt.Errorf("cannot output Go code for %s", tok.Repr()))
		return GoVarNone
	}
}

func TokenParseInt(input string) int64 {
	text := strings.ReplaceAll(input, "_", "")
	out, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid string literal: %s -- %v", input, err))
	}
	return out
}

func TokenParseStr(str string) string {
	out := strings.Builder{}
	raw := false
	if strings.HasPrefix(str, "r") {
		str = str[1:]
		raw = true
	}

	delim, str, dbl := str[:1], str[1:], ""
	if raw {
		dbl = delim + delim
	}

	if strings.HasSuffix(str, delim) {
		str = str[:len(str)-1]
	}

	for len(str) > 0 {
		if raw {
			if pos := strings.Index(str, dbl); pos >= 0 {
				out.WriteString(str[:pos])
				out.WriteString(delim)
				str = str[pos+len(dbl):]
			} else {
				out.WriteString(str)
				str = ""
			}
		} else {
			if pos := strings.Index(str, "\\"); pos >= 0 {
				out.WriteString(str[:pos])
				str = str[pos+1:]
				if strings.HasPrefix(str, "\\") {
					out.WriteString("\\")
					str = str[1:]
				}
			} else {
				out.WriteString(str)
				str = ""
			}
		}
	}

	return out.String()
}
