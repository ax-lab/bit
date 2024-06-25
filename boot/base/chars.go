package base

import "unicode"

func IsSpace(chr rune) bool {
	return chr != '\r' && chr != '\n' && unicode.IsSpace(chr)
}
