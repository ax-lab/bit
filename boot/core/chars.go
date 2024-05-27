package core

import "unicode"

func IsSpace(chr rune) bool {
	return chr != '\r' && chr != '\n' && unicode.IsSpace(chr)
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

func IsLetterOrDigit(chr rune) bool {
	return IsAlphaNum(chr) || unicode.IsLetter(chr) || unicode.IsNumber(chr)
}

func IsWord(chr rune) bool {
	return chr == '_' || IsLetterOrDigit(chr)
}
