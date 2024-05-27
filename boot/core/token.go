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
