package glob

import (
	"fmt"
	"regexp"
	"strings"

	"axlab.dev/test/text"
)

const globRegexIgnoreCase = "(?i)"

func Match(input, pattern string) bool {
	if pattern == "" {
		return true
	}

	re := regexp.MustCompile(globRegexIgnoreCase + Regex(pattern))
	return re.MatchString(input)
}

func Regex(pattern string) string {
	if pattern == "" {
		return ""
	}

	output := []string{globRegexIgnoreCase}

	glob := text.MatchIf(globSpecial)
	scan := text.ScannerNew(pattern)
	for scan.Len() > 0 {
		if literal := scan.ReadUntil(glob); len(literal) > 0 {
			output = append(output, regexp.QuoteMeta(literal))
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
			panic(fmt.Sprintf("unsupported special character: %c", next))
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
