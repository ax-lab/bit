package core

import (
	"fmt"
	"strings"
)

type TokenType string

type Token struct {
	Type  TokenType
	Value any
}

func (token Token) String() string {
	out := strings.Builder{}
	out.WriteString("Token(")
	out.WriteString(string(token.Type))
	if token.Value != nil {
		out.WriteString("=")
		out.WriteString(fmt.Sprint(token.Value))
	}
	out.WriteString(")")
	return out.String()
}

type Invalid string

func (inv Invalid) String() string {
	return fmt.Sprintf("Invalid(%#v)", inv)
}

type Symbol string

func (sym Symbol) String() string {
	return fmt.Sprintf("Symbol(%#v)", sym)
}

type Word string

func (sym Word) String() string {
	return fmt.Sprintf("Word(%s)", string(sym))
}

type LineBreak string

func (LineBreak) String() string {
	return "LineBreak"
}

type Integer struct {
	Base   int
	Digits string
	Suffix string
}

func (num Integer) String() string {
	out := strings.Builder{}
	out.WriteString("Int(")

	based := false
	switch num.Base {
	case 0, 10:
		break
	case 2:
		out.WriteString("0b")
	case 8:
		out.WriteString("0c")
	case 16:
		out.WriteString("0x")
	default:
		based = true
	}

	out.WriteString(num.Digits)
	if len(num.Suffix) > 0 {
		out.WriteString(fmt.Sprintf("; suffix=%s", num.Suffix))
	}

	if based {
		out.WriteString(fmt.Sprintf("; base=%d", num.Base))
	}

	out.WriteString(")")
	return out.String()
}

type Comment struct {
	Text string
	Sta  string
	End  string
}

func (comment Comment) String() string {
	const maxW = HintTextColumns / 2
	text := Clip(comment.Text, maxW, "…")

	repr := strings.Builder{}
	repr.WriteString("Comment(")
	repr.WriteString(comment.Sta)
	repr.WriteString("[")
	repr.WriteString(text)
	repr.WriteString("]")
	if comment.End != "" {
		repr.WriteString(comment.End)
	}
	repr.WriteString(")")

	out := repr.String()
	return out
}
