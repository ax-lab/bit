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

func StrTrimSta(input string) string {
	return strings.TrimLeftFunc(input, IsSpace)
}

func StrIndent(input string, prefix ...IndentPrefix) string {
	tab := StrJoin(Sep(""), prefix...)
	if tab == "" {
		tab = "\t"
	}

	nonSpace := StrTrim(tab) != ""
	output, hasOutput := strings.Builder{}, false

	for _, line := range StrLines(input) {
		if hasOutput {
			output.WriteString("\n")
		}

		line = StrTrimEnd(line)
		if len(line) > 0 || nonSpace {
			output.WriteString(tab)
		}
		output.WriteString(line)
		hasOutput = true
	}

	return output.String()
}

type Sep string

func StrJoin[T ~string](sep Sep, parts ...T) string {
	out := strings.Builder{}
	for _, it := range parts {
		if part := string(it); len(part) > 0 {
			if out.Len() > 0 {
				out.WriteString(string(sep))
			}
			out.WriteString(part)
		}
	}
	return out.String()
}
