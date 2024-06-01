package core

import (
	"unicode"
)

func IsSpace(chr rune) bool {
	return chr != '\r' && chr != '\n' && unicode.IsSpace(chr)
}

func IsIdentifierChar(chr rune, start bool) bool {
	id := chr == '_' || IsAlphaNum(chr) || IsLetterOrNumber(chr)
	if start && id {
		id = !(IsDigit(chr) || unicode.IsNumber(chr))
	}
	return id
}

func IsAlpha(chr rune) bool {
	return 'a' <= chr && chr <= 'z' || 'A' <= chr && chr <= 'Z'
}

func IsDigit(chr rune) bool {
	return '0' <= chr && chr <= '9'
}

func IsAlphaNum(chr rune) bool {
	return IsDigit(chr) || IsAlpha(chr)
}

func IsLetter(chr rune) bool {
	return IsAlpha(chr) || unicode.IsLetter(chr)
}

func IsLetterOrNumber(chr rune) bool {
	return IsAlphaNum(chr) || unicode.IsLetter(chr) || unicode.IsNumber(chr)
}

func IsWord(chr rune) bool {
	return chr == '_' || IsLetterOrNumber(chr)
}

func IsBaseDigit(chr rune, base int) bool {
	if digit := int(chr - '0'); digit >= 0 && digit < base {
		return true
	}

	if base <= 10 {
		return false
	}

	digit := base
	if 'A' <= chr && chr <= 'Z' {
		digit = int(chr - 'A')
	} else if 'a' <= chr && chr <= 'z' {
		digit = int(chr - 'a')
	}
	return (digit + 10) < base
}

func IsLineBreak(chr rune) bool {
	return chr == '\r' || chr == '\n'
}
