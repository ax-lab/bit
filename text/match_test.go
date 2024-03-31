package text_test

import (
	"testing"

	"axlab.dev/test/text"
	"github.com/stretchr/testify/require"
)

func TestMatchRegex(t *testing.T) {
	matcher := text.MatchRegex(`\d+`)
	testMatch(t, "", matcher, "", "")
	testMatch(t, "123456", matcher, "", "123456")
	testMatch(t, "123456!!!", matcher, "", "123456")
	testMatch(t, "prefix:123456!!!", matcher, "prefix:", "123456")
}

func TestMatchIf(t *testing.T) {
	matcher := text.MatchIf(func(chr rune) bool { return '0' <= chr && chr <= '9' })
	testMatch(t, "", matcher, "", "")
	testMatch(t, "123456", matcher, "", "123456")
	testMatch(t, "123456!!!", matcher, "", "123456")
	testMatch(t, "prefix:123456!!!", matcher, "prefix:", "123456")
}

func testMatch(t *testing.T, input string, m text.Matcher, prefix, match string) {
	test := require.New(t)

	sta, end := m.FindMatch(input)
	test.Equal(prefix, input[:sta])
	test.Equal(match, input[sta:end])

	pos := m.FindMatchIndex(input)
	test.Equal(len(prefix), pos)

	matched := m.MatchNext(input)
	if len(prefix) > 0 {
		test.Equal(0, matched)
		test.Equal(len(match), m.MatchNext(input[len(prefix):]))
	} else {
		test.Equal(len(match), matched)
	}
}
