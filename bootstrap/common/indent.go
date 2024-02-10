package common

import "strings"

func Indent(input string, prefix ...string) string {
	return doIndent(true, input, prefix...)
}

func Indented(input string, prefix ...string) string {
	return doIndent(false, input, prefix...)
}

func doIndent(indentNext bool, input string, prefix ...string) string {
	tab := strings.Join(prefix, "")
	if len(tab) == 0 {
		tab = "    "
	}

	nonSpace := Trim(tab) != ""
	output, hasOutput := strings.Builder{}, false
	for _, line := range Lines(input) {
		if hasOutput {
			output.WriteString("\n")
		}

		line = TrimEnd(line)
		if indentNext && (len(line) > 0 || nonSpace) {
			output.WriteString(tab)
		}
		output.WriteString(line)
		indentNext = true
		hasOutput = true
	}

	return output.String()
}
