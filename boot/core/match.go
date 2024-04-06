package core

import "regexp"

type Matcher interface {
	FindMatchIndex(input string) int
	FindMatch(input string) (sta, end int)
	MatchNext(input string) int
}

func MatchIf(pred func(chr rune) bool) Matcher {
	return matcherIf{pred}
}

func MatchRegex(src string) Matcher {
	return matcherRegex{
		reFind:  regexp.MustCompile(src),
		reMatch: regexp.MustCompile("^(" + src + ")"),
	}
}

type matcherRegex struct {
	reFind  *regexp.Regexp
	reMatch *regexp.Regexp
}

func (m matcherRegex) FindMatchIndex(input string) int {
	sta, _ := m.FindMatch(input)
	return sta
}

func (m matcherRegex) FindMatch(input string) (sta, end int) {
	pos := m.reFind.FindStringIndex(input)
	if len(pos) == 0 {
		last := len(input)
		return last, last
	}
	return pos[0], pos[1]
}

func (m matcherRegex) MatchNext(input string) int {
	pos := m.reMatch.FindStringIndex(input)
	if len(pos) == 0 {
		return 0
	}
	return pos[1]
}

type matcherIf struct {
	pred func(chr rune) bool
}

func (m matcherIf) FindMatchIndex(input string) int {
	for n, chr := range input {
		if m.pred(chr) {
			return n
		}
	}
	return len(input)
}

func (m matcherIf) FindMatch(input string) (sta, end int) {
	sta = m.FindMatchIndex(input)
	end = sta + m.MatchNext(input[sta:])
	return
}

func (m matcherIf) MatchNext(input string) int {
	for n, chr := range input {
		if !m.pred(chr) {
			return n
		}
	}
	return len(input)
}
