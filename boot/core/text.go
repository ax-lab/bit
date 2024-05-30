package core

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	DefaultTabSize  = 4
	HintTextColumns = 60
)

type WithDump interface {
	Dump() string
}

var (
	reLines = regexp.MustCompile(`\r?\n|\r`)
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

type Prefix string

func IndentBlock(text string) (out string) {
	if strings.ContainsAny(text, "\r\n") || len(text) > HintTextColumns {
		out = fmt.Sprintf("\n\t%s\n", Indent(text))
	} else {
		out = text
	}
	return out
}

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
		if hasOutput && len(line) > 0 || nonSpace {
			output.WriteString(tab)
		}
		output.WriteString(line)
		hasOutput = true
	}

	return output.String()
}

func Clip(input string, maxLen int, trimSuffix ...string) string {
	text := strings.TrimSpace(input)
	clip := false

	if eol := strings.IndexAny(text, "\r\n"); eol >= 0 {
		text = text[:eol]
		clip = true
	}

	if len(text) > maxLen {
		offset := 0
		for pos := range text {
			if pos <= maxLen {
				offset = pos
			} else {
				break
			}
		}
		text = text[:offset]
		clip = true
	}

	if clip && len(trimSuffix) > 0 {
		text += strings.Join(trimSuffix, "")
	}

	return text
}
