package core

import (
	"strings"
	"unicode/utf8"
)

type Scanner struct {
	text string
	pos  int
}

func ScannerNew(input string) Scanner {
	return Scanner{input, 0}
}

func (scan *Scanner) Pos() int {
	return scan.pos
}

func (scan *Scanner) Len() int {
	return len(scan.text) - scan.pos
}

func (scan *Scanner) Text() string {
	return scan.text[scan.pos:]
}

func (scan *Scanner) Peek() rune {
	text := scan.Text()
	for _, chr := range text {
		return chr
	}
	return 0
}

func (scan *Scanner) Read() (out rune, ok bool) {
	text := scan.Text()
	chr, n := utf8.DecodeRuneInString(text)
	scan.pos += n
	return chr, n > 0
}

func (scan *Scanner) SkipIf(prefix string) bool {
	text := scan.Text()
	if strings.HasPrefix(text, prefix) {
		scan.pos += len(prefix)
		return true
	}
	return false
}

func (scan *Scanner) SkipAny(prefixes ...string) string {
	text := scan.Text()
	for _, prefix := range prefixes {
		if strings.HasPrefix(text, prefix) {
			scan.pos += len(prefix)
			return prefix
		}
	}
	return ""
}

func (scan *Scanner) PeekChars(count int) string {
	bytes := 0
	text := scan.Text()
	for count > 0 {
		_, n := utf8.DecodeRuneInString(text[bytes:])
		bytes += n
		count -= 1
		if n == 0 {
			break
		}
	}
	return text[:bytes]
}

func (scan *Scanner) ReadChars(count int) (out string) {
	out = scan.PeekChars(count)
	scan.pos += len(out)
	return out
}

func (scan *Scanner) ReadWhile(match Matcher) (out string) {
	var bytes int
	text := scan.Text()
	for {
		next := match.MatchNext(text[bytes:])
		if next > 0 {
			bytes += next
		} else {
			break
		}
	}
	out = text[:bytes]
	scan.pos += bytes
	return out
}

func (scan *Scanner) ReadUntil(match Matcher) (out string) {
	text := scan.Text()
	pos := match.FindMatchIndex(text)
	out = text[:pos]
	scan.pos += pos
	return out
}

func (scan *Scanner) ReadMatch(match Matcher) (prefix, matched string) {
	text := scan.Text()
	sta, end := match.FindMatch(text)
	if sta == end && sta < len(text) {
		panic("Scanner.ReadMatch: invalid match for the empty string")
	}
	prefix = text[:sta]
	matched = text[sta:end]
	scan.pos += end
	return
}
