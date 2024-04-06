package core

import (
	"fmt"
	"regexp"
	"strings"
)

const globRegexIgnoreCase = "(?i)"

type GlobOption string

const (
	GlobIncludeDirs GlobOption = "+D"
)

func GlobMatch(input, pattern string) bool {
	if pattern == "" {
		return true
	}

	re := regexp.MustCompile(GlobRegex(pattern))
	return re.MatchString(input)
}

func GlobRegex(pattern string) string {
	if pattern == "" {
		return ""
	}

	output := []string{globRegexIgnoreCase}

	glob := MatchIf(globSpecial)
	scan := ScannerNew(pattern)
	for scan.Len() > 0 {
		if literal := scan.ReadUntil(glob); len(literal) > 0 {
			output = append(output, regexp.QuoteMeta(literal))
			continue
		}

		if scan.SkipIf("**") {
			output = append(output, `.*`)
			continue
		} else if scan.SkipIf("[^") {
			output = append(output, `[^`)
			continue
		}

		next, _ := scan.Read()
		switch next {
		case '\\':
			if escaped := scan.ReadChars(1); len(escaped) > 0 {
				output = append(output, regexp.QuoteMeta(escaped))
			}
		case '(', ')', '|':
			output = append(output, string(next))
		case '?':
			output = append(output, `[^/\\]`)
		case '*':
			output = append(output, `[^/\\]*`)
		case '/':
			output = append(output, `[/\\]`)
		case '^':
			output = append(output, `\^`)
		default:
			panic(fmt.Sprintf("unsupported special character: %c (U+%04X)", next, next))
		}
	}

	return strings.Join(output, "")
}

func globSpecial(chr rune) bool {
	switch chr {
	case '\\', '/', '*', '?', '(', ')', '|', '^':
		return true
	default:
		return false
	}
}
