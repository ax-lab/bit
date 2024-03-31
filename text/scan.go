package text

import (
	"strings"
	"unicode/utf8"
)

type Scanner struct {
	text string
}

func ScannerNew(input string) Scanner {
	return Scanner{input}
}

func (scan *Scanner) Len() int {
	return len(scan.text)
}

func (scan *Scanner) Text() string {
	return scan.text
}

func (scan *Scanner) Peek() rune {
	for _, chr := range scan.text {
		return chr
	}
	return 0
}

func (scan *Scanner) Read() (out rune, ok bool) {
	chr, n := utf8.DecodeRuneInString(scan.text)
	scan.text = scan.text[n:]
	return chr, n > 0
}

func (scan *Scanner) SkipIf(prefix string) bool {
	if strings.HasPrefix(scan.text, prefix) {
		scan.text = scan.text[len(prefix):]
		return true
	}
	return false
}

func (scan *Scanner) PeekChars(count int) string {
	bytes := 0
	for count > 0 {
		_, n := utf8.DecodeRuneInString(scan.text[bytes:])
		bytes += n
		count -= 1
		if n == 0 {
			break
		}
	}
	return scan.text[:bytes]
}

func (scan *Scanner) ReadChars(count int) (out string) {
	out = scan.PeekChars(count)
	scan.text = scan.text[len(out):]
	return out
}

func (scan *Scanner) ReadWhile(match Matcher) (out string) {
	var bytes int
	for {
		next := match.MatchNext(scan.text[bytes:])
		if next > 0 {
			bytes += next
		} else {
			break
		}
	}
	out = scan.text[:bytes]
	scan.text = scan.text[bytes:]
	return out
}

func (scan *Scanner) ReadUntil(match Matcher) (out string) {
	pos := match.FindMatchIndex(scan.text)
	out = scan.text[:pos]
	scan.text = scan.text[pos:]
	return out
}

func (scan *Scanner) ReadMatch(match Matcher) (prefix, matched string) {
	sta, end := match.FindMatch(scan.text)
	if sta == end && sta < len(scan.text) {
		panic("Scanner.ReadMatch: invalid match for the empty string")
	}
	prefix = scan.text[:sta]
	matched = scan.text[sta:end]
	scan.text = scan.text[end:]
	return
}
