package lang

import (
	"fmt"
	"regexp"
	"strings"

	"axlab.dev/bit/core"
)

const (
	TokenInteger core.TokenType = "Integer"
	TokenFloat   core.TokenType = "Float"
	TokenString  core.TokenType = "String"
)

func MatcherWithRE(regex string, token core.TokenType) core.LexMatcher {
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

func MatchNumber(input *core.Cursor) (val core.Value, err error) {
	if !core.IsDigit(input.Peek()) {
		return nil, nil
	}

	out := strings.Builder{}

	base := 10
	switch input.ReadAny("0x", "0c", "0o", "0b") {
	case "0x":
		base = 16
	case "0c", "0o":
		base = 8
	case "0b":
		base = 2
	}

	skipSeparator := func(input *core.Cursor) bool {
		return input.SkipWhile(func(chr rune) bool { return chr == '_' })
	}

	if base != 10 {
		skipSeparator(input)
	}

	digit := func(chr rune) bool { return core.IsBaseDigit(chr, base) }

	chunk := input.ReadWhile(digit)
	out.WriteString(chunk)

	if valid := len(chunk) > 0; !valid {
		err = fmt.Errorf("invalid numeric literal")
	}

	for skipSeparator(input) {
		chunk = input.ReadWhile(digit)
		out.WriteString(chunk)
		if len(chunk) == 0 {
			break
		}
	}

	suffix := input.ReadWhile(core.IsLetter)
	if suffix != "" {
		suffix += input.ReadWhile(core.IsWord)
	} else {
		sta := *input
		invalid := input.ReadWhile(core.IsWord)
		if err == nil && len(invalid) > 0 {
			err = core.Errorf(input.GetSpan(sta), "invalid digits in numeric literal of base %d", base)
		}
	}

	val = core.Integer{
		Digits: out.String(),
		Base:   base,
		Suffix: suffix,
	}

	return val, err
}

func MatcherLineComment(prefixes ...string) core.LexMatcher {
	return func(input *core.Cursor) (core.Value, error) {
		prefix := input.ReadAny(prefixes...)
		if len(prefix) == 0 {
			return nil, nil
		}

		text := input.ReadWhile(core.Not(core.IsLineBreak))
		text = strings.TrimFunc(text, core.IsSpace)

		val := core.Comment{
			Text: text,
			Sta:  prefix,
		}
		return val, nil
	}
}

func MatcherBlockComment(pairs ...string) core.LexMatcher {
	var (
		staDelims []string
		endDelims []string
	)

	trim := func(val string) bool {
		return val == "" || len(strings.TrimSpace(val)) != len(val)
	}

	for _, it := range pairs {
		split := strings.Split(it, " ")
		if len(split) != 2 || trim(split[0]) || trim(split[1]) {
			panic(fmt.Sprintf("invalid block comment delimiter: %#v", it))
		}
		staDelims = append(staDelims, split[0])
		endDelims = append(endDelims, split[1])
	}

	return func(input *core.Cursor) (core.Value, error) {
		var (
			stack []string
			sta   string
			end   string
		)

		readStart := func(input *core.Cursor) (sta, end string) {
			index := input.ReadFrom(staDelims)
			if index >= 0 {
				return staDelims[index], endDelims[index]
			}
			return "", ""
		}

		if sta, end = readStart(input); len(sta) == 0 {
			return nil, nil
		} else {
			stack = []string{end}
		}

		input.SkipSpaces()
		textSta := *input
		textEnd := textSta

		for input.Len() > 0 && len(stack) > 0 {
			last := len(stack) - 1
			if endDelim := stack[last]; input.ReadIf(endDelim) {
				stack = stack[:last]
			} else if sta, end = readStart(input); len(sta) > 0 {
				stack = append(stack, end)
			} else {
				input.Read()
				textEnd = *input
				input.SkipSpaces()
			}
		}

		text := textSta.GetSpan(textEnd).Text()
		val := core.Comment{
			Text: text,
			Sta:  sta,
			End:  end,
		}
		return val, nil
	}
}
