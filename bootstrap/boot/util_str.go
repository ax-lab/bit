package boot

import (
	"strings"
	"unicode"
)

func IsSpace(chr rune) bool {
	return chr != '\n' && chr != '\r' && unicode.IsSpace(chr)
}

func StrLines(input string) []string {
	return strings.Split(input, "\n")
}

func StrTrim(input string) string {
	return strings.TrimFunc(input, IsSpace)
}
