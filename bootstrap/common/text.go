package common

import (
	"regexp"
	"strings"
	"unicode"
)

const DefaultTabSize = 4

var (
	TRAILING_EOL = regexp.MustCompile(`(\r\n?|\n)$`)
	TABS         = regexp.MustCompile(`^[\t]+`)
)

func Trim(str string) string {
	return strings.TrimFunc(str, IsSpace)
}

func TrimEnd(str string) string {
	return strings.TrimRightFunc(str, IsSpace)
}

func CleanupText(input string) string {
	out := make([]string, 0)
	pre := ""
	for _, it := range TrimLines(Lines(input)) {
		if len(out) == 0 {
			if strings.TrimSpace(it) == "" {
				continue
			}

			indent := len(it) - len(strings.TrimLeftFunc(it, unicode.IsSpace))
			pre = it[:indent]
		}

		out = append(out, strings.TrimPrefix(it, pre))
	}
	text := strings.Join(out, "\n")
	return AddTrailingNewLine(text)
}

func AddTrailingNewLine(input string) string {
	if len(input) > 0 && !TRAILING_EOL.MatchString(input) {
		input = input + "\n"
	}
	return input
}

func ExpandTabs(input string, tabSize int) string {
	if tabSize < 0 {
		tabSize = DefaultTabSize
	}
	tab := strings.Repeat(" ", tabSize)
	out := make([]string, 0)
	for _, it := range Lines(input) {
		it = TABS.ReplaceAllStringFunc(it, func(input string) string {
			return strings.Replace(input, "\t", tab, -1)
		})
		out = append(out, it)
	}
	return strings.Join(out, "\n")
}
