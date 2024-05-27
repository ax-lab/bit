package lang

import (
	"fmt"
	"regexp"

	"axlab.dev/bit/core"
)

const (
	TokenInteger core.TokenType = "Integer"
	TokenFloat   core.TokenType = "Float"
	TokenString  core.TokenType = "String"
)

func MatchWithRE(regex string, token core.TokenType) core.LexMatcher {
	regex = fmt.Sprintf(`^(%s)`, regex)
	re := regexp.MustCompile(regex)
	return func(input *core.Cursor) (core.Value, error) {
		if m := re.FindString(input.Text()); len(m) > 0 {
			input.Advance(len(m))
			return core.Token{Type: token}, nil
		}
		return nil, nil
	}
}

func MatchNumber(input *core.Cursor) (core.Value, error) {
	base := 10
	switch input.ReadAny("0x", "0c", "0o", "0b") {
	case "0x":
		base = 16
	case "0c", "0o":
		base = 8
	case "0b":
		base = 2
	}
	_ = base
	panic("TODO")
}

func MatchWord(input *core.Cursor) (core.Value, error) {

	sta := *input
	next := input.Peek()
	if next != '_' && !core.IsLetter(next) {
		return nil, nil
	}

	input.SkipWhile(core.IsWord)

	for input.Peek() == '-' {
		tmp := *input
		tmp.Read()
		if tmp.SkipWhile(core.IsWord) {
			*input = tmp
		} else {
			break
		}
	}

	word := input.GetSpan(sta).Text()
	return core.Word(word), nil
}
