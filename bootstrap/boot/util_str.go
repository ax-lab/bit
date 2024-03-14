package boot

import (
	"strings"
	"unicode"
)

type IndentPrefix string

func IsSpace(chr rune) bool {
	return chr != '\n' && chr != '\r' && unicode.IsSpace(chr)
}

func StrLines(input string) []string {
	return strings.Split(input, "\n")
}

func StrTrim(input string) string {
	return strings.TrimFunc(input, IsSpace)
}

func StrTrimEnd(input string) string {
	return strings.TrimRightFunc(input, IsSpace)
}

func StrIndent(input string, prefix ...IndentPrefix) string {
	return strDoIndent(true, input, prefix...)
}

func strDoIndent(indentNext bool, input string, prefix ...IndentPrefix) string {
	tabBuffer := strings.Builder{}
	for _, it := range prefix {
		tabBuffer.WriteString(string(it))
	}

	tab := "\t"
	if tabBuffer.Len() > 0 {
		tab = tabBuffer.String()
	}

	nonSpace := StrTrim(tab) != ""
	output, hasOutput := strings.Builder{}, false

	for _, line := range StrLines(input) {
		if hasOutput {
			output.WriteString("\n")
		}

		line = StrTrimEnd(line)
		if indentNext && (len(line) > 0 || nonSpace) {
			output.WriteString(tab)
		}
		output.WriteString(line)
		indentNext = true
		hasOutput = true
	}

	return output.String()
}
