package bit

import (
	"fmt"
	"regexp"
)

type Matcher func(cur *Cursor) (TokenType, error)

func MatchWithRE(token TokenType, reExpr string) Matcher {
	reExpr = fmt.Sprintf(`^(%s)`, reExpr)
	re := regexp.MustCompile(reExpr)
	return func(cur *Cursor) (TokenType, error) {
		if m := re.FindString(cur.Text()); len(m) > 0 {
			cur.Advance(len(m))
			return token, nil
		} else {
			return TokenNone, nil
		}
	}
}

func MatchWord(cur *Cursor) (TokenType, error) {
	isLetter := func(chr rune) bool {
		return 'a' <= chr && chr <= 'z' || 'A' <= chr && chr <= 'Z'
	}

	isDigit := func(chr rune) bool {
		return '0' <= chr && chr <= '9'
	}

	next := cur.Read()
	charIsLetter := isLetter(next)
	if isWord := charIsLetter || next == '_'; !isWord {
		return TokenNone, nil
	}

	for !cur.IsEnd() {
		aux := *cur
		next = aux.Read()
		if isSeparator := charIsLetter && next == '-'; isSeparator {
			if charIsLetter = isLetter(aux.Read()); !charIsLetter {
				break
			}
		} else {
			charIsLetter = isLetter(next)
			if !(charIsLetter || next == '_' || isDigit(next)) {
				break
			}
		}

		*cur = aux
	}

	if next = cur.Peek(); next == '?' || next == '!' {
		cur.Read()
	}

	return TokenWord, nil
}

func MatchString(cur *Cursor) (TokenType, error) {
	startPos := *cur
	delim := cur.ReadAny(`r"`, `r'`, `"`, `'`)
	if delim == "" {
		return TokenNone, nil
	}

	raw := delim[0] == 'r'
	if raw {
		delim = delim[1:]
	}

	doubleDelim := delim + delim
	for !cur.IsEnd() {
		if !raw && cur.ReadIf(`\`) {
			cur.Read()
		} else if !raw || !cur.ReadIf(doubleDelim) {
			if cur.ReadIf(delim) {
				return TokenString, nil
			} else {
				cur.Read()
			}
		}
	}

	return TokenString, startPos.Error(0, "string literal missing closing `%s`", delim)
}
