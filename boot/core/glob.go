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

func GlobParse(input string) (prefix string, pattern string) {
	if input == "" {
		return
	}

	glob := MatchIf(globSpecial)
	scan := ScannerNew(input)

	preList := []string{}
	slashPos, slashIndex, done := 0, 0, false
	for !done && scan.Len() > 0 {
		if literal := scan.ReadUntil(glob); len(literal) > 0 {
			preList = append(preList, literal)
			continue
		}

		next := scan.Peek()
		switch next {
		case '\\':
			scan.Read()
			escaped := scan.ReadChars(1)
			preList = append(preList, escaped)
		case '/':
			preList = append(preList, "/")
			scan.Read()
			slashPos, slashIndex = scan.Pos(), len(preList)
		default:
			done = true
		}
	}

	if scan.Text() != "" && slashIndex > 0 {
		prefix = strings.Join(preList[:slashIndex], "")
		pattern = input[slashPos:]
	} else {
		prefix = strings.Join(preList, "")
		pattern = input[scan.Pos():]
	}
	return
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
		} else if pos := scan.SkipAny("]+", "]*", "]?", "]"); pos != "" {
			output = append(output, pos)
			continue
		}

		next, _ := scan.Read()
		switch next {
		case '\\':
			if escaped := scan.ReadChars(1); len(escaped) > 0 {
				output = append(output, regexp.QuoteMeta(escaped))
			}
		case '(', ')', '|', '[', ']':
			output = append(output, string(next))
		case '?':
			output = append(output, `[^/\\]`)
		case '*':
			output = append(output, `[^/\\]*`)
		case '/':
			output = append(output, `[/\\]`)
		default:
			panic(fmt.Sprintf("unsupported special character: %c (U+%04X)", next, next))
		}
	}

	return strings.Join(output, "")
}

func globSpecial(chr rune) bool {
	switch chr {
	case '\\', '/', '*', '?', '(', ')', '[', ']', '|':
		return true
	default:
		return false
	}
}
