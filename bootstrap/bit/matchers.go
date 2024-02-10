package bit

import (
	"fmt"
	"regexp"
	"strings"
)

type Matcher func(cur *Cursor) (TokenType, error)

func MatchInteger() Matcher {
	return MatchWithRE(TokenInteger, `\d+`)
}

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

func ParseStringLiteral(str string) string {
	out := strings.Builder{}
	raw := false
	if strings.HasPrefix(str, "r") {
		str = str[1:]
		raw = true
	}

	delim, str, dbl := str[:1], str[1:], ""
	if raw {
		dbl = delim + delim
	}

	if strings.HasSuffix(str, delim) {
		str = str[:len(str)-1]
	}

	for len(str) > 0 {
		if raw {
			if pos := strings.Index(str, dbl); pos >= 0 {
				out.WriteString(str[:pos])
				out.WriteString(delim)
				str = str[pos+len(dbl):]
			} else {
				out.WriteString(str)
				str = ""
			}
		} else {
			if pos := strings.Index(str, "\\"); pos >= 0 {
				out.WriteString(str[:pos])
				str = str[pos+1:]
				if strings.HasPrefix(str, "\\") {
					out.WriteString("\\")
					str = str[1:]
				}
			} else {
				out.WriteString(str)
				str = ""
			}
		}
	}

	return out.String()
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
