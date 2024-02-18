package common

import "unicode"

func IsSpace(chr rune) bool {
	return chr != '\n' && chr != '\r' && unicode.IsSpace(chr)
}

func IsDigit(chr rune) bool {
	return '0' <= chr && chr <= '9'
}

func IsAlpha(chr rune) bool {
	return 'A' <= chr && chr <= 'Z' || 'a' <= chr && chr <= 'z'
}
