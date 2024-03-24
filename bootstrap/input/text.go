package input

import (
	"regexp"
	"strings"
	"unicode"
)

const DefaultTabSize = 4

func IsSpace(chr rune) bool {
	return chr != '\n' && chr != '\r' && unicode.IsSpace(chr)
}

func IsAlpha(chr rune) bool {
	return 'A' <= chr && chr <= 'Z' || 'a' <= chr && chr <= 'z'
}

func IsDigit(chr rune) bool {
	return '0' <= chr && chr <= '9'
}

func IsWord(chr rune) bool {
	return chr == '_' || IsAlpha(chr) || IsDigit(chr)
}

var (
	reLines = regexp.MustCompile(`\r?\n`)
)

func Lines(text string) []string {
	return reLines.Split(text, -1)
}

func Trim(text string) string {
	return strings.TrimFunc(text, IsSpace)
}

func TrimEnd(text string) string {
	return strings.TrimRightFunc(text, IsSpace)
}

func TrimSta(text string) string {
	return strings.TrimLeftFunc(text, IsSpace)
}

type Prefix string

func Indent(text string, prefix ...Prefix) string {
	tab := Join(Sep(""), prefix...)
	if tab == "" {
		tab = "\t"
	}

	nonSpace := Trim(tab) != ""
	output, hasOutput := strings.Builder{}, false

	for _, line := range Lines(text) {
		if hasOutput {
			output.WriteString("\n")
		}

		line = TrimEnd(line)
		if len(line) > 0 || nonSpace {
			output.WriteString(tab)
		}
		output.WriteString(line)
		hasOutput = true
	}

	return output.String()
}

type Sep string

func Join[T ~string](sep Sep, parts ...T) string {
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

func Text(text string) string {
	lines := Lines(TrimEnd(text))
	if len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}

	if len(lines) > 0 {
		line := lines[0]
		trim := TrimSta(line)
		diff := line[:len(line)-len(trim)]
		for i, it := range lines {
			if len(diff) > 0 && strings.HasPrefix(it, diff) {
				lines[i] = it[len(diff):]
			}
			lines[i] = TrimEnd(lines[i])
		}
	}

	if len(lines) == 0 || lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}