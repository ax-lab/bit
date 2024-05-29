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
	return func(mod *core.Module, lexer *core.Lexer, input *core.Cursor) core.Value {
		if m := re.FindString(input.Text()); len(m) > 0 {
			input.Advance(len(m))
			return core.Token{Type: token}
		}
		return nil
	}
}

func MatchWord(mod *core.Module, lexer *core.Lexer, input *core.Cursor) core.Value {

	sta := *input
	next := input.Peek()
	if next != '_' && !core.IsLetter(next) {
		return nil
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
	return core.Word(word)
}

func MatchNumber(mod *core.Module, lexer *core.Lexer, input *core.Cursor) core.Value {
	if !core.IsDigit(input.Peek()) {
		return nil
	}

	var err error

	sta := *input
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

	if err != nil {
		err = core.ErrorAt(input.GetSpan(sta), err)
		mod.Error(err)
	}

	val := core.Integer{
		Digits: out.String(),
		Base:   base,
		Suffix: suffix,
	}
	return val
}

func MatchString(mod *core.Module, lexer *core.Lexer, input *core.Cursor) core.Value {

	sta := *input
	delim := input.ReadAny(`r"`, `r'`, `"`, `'`)
	if delim == "" {
		return nil
	}

	prefix := ""
	raw := delim[0] == 'r'
	if raw {
		delim = delim[1:]
		prefix = "r"
	}

	doubleDelim := delim + delim
	closed := false

	textSta := *input
	textEnd := textSta
	segments := []core.LiteralExprSegment(nil)

	pushText := func() {
		text := textSta.GetSpan(textEnd).Text()
		if len(text) > 0 {
			segments = append(segments, core.LiteralExprSegment{Text: text})
		}
		textSta = textEnd
	}

	for input.Len() > 0 {
		var evalSta, evalEnd string
		if !raw {
			evalSta = input.ReadAny("${", "$[", "$(")
			switch evalSta {
			case "${":
				evalEnd = "}"
			case "$[":
				evalEnd = "]"
			case "$(":
				evalEnd = ")"
			}
		}

		if len(evalSta) > 0 {
			pushText()
			stop := func(input *core.Cursor) bool { return input.ReadIf(evalEnd) }
			nodes := lexer.TokenizeUntil(mod, input, stop)
			expr := core.NodeListNew(core.SpanForRange(nodes), nodes...)
			segments = append(segments, core.LiteralExprSegment{Expr: expr})
			textSta = *input
			textEnd = textSta
		} else if escape := !raw && input.ReadIf(`\`); escape {
			input.Read()
		} else if double := input.ReadIf(doubleDelim); !double {
			if input.ReadIf(delim) {
				closed = true
				break
			} else {
				input.Read()
			}
		}
		textEnd = *input
	}

	pushText()

	if !closed {
		err := core.Errorf(input.GetSpan(sta), "unclosed string literal")
		mod.Error(err)
	}

	var str core.Value
	if len(segments) == 0 {
		str = core.Literal{
			RawText: "",
			Delim:   delim,
			Prefix:  prefix,
		}
	} else if seg := segments[0]; len(segments) == 1 && !seg.Expr.Valid() {
		str = core.Literal{
			RawText: seg.Text,
			Delim:   delim,
			Prefix:  prefix,
		}
	} else {
		str = core.LiteralExpr{
			Segments: segments,
			Delim:    delim,
			Prefix:   prefix,
		}
	}

	return str
}

func MatcherLineComment(prefixes ...string) core.LexMatcher {
	return func(mod *core.Module, lexer *core.Lexer, input *core.Cursor) core.Value {
		prefix := input.ReadAny(prefixes...)
		if len(prefix) == 0 {
			return nil
		}

		text := input.ReadWhile(core.Not(core.IsLineBreak))
		text = strings.TrimFunc(text, core.IsSpace)

		val := core.Comment{
			Text: text,
			Sta:  prefix,
		}
		return val
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

	return func(mod *core.Module, lexer *core.Lexer, input *core.Cursor) core.Value {
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
			return nil
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
		return val
	}
}
