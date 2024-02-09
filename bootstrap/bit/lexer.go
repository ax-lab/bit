package bit

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	COMMENT_CHAR      = '#'
	COMMENT_DELIM     = '/'
	COMMENT_DELIM_STA = "/#"
	COMMENT_DELIM_END = "#/"
)

type Symbol string

func (sym Symbol) IsEqual(key Key) bool {
	if v, ok := key.(Symbol); ok {
		return sym == v
	}
	return false
}

func (sym Symbol) Repr(oneline bool) string {
	return fmt.Sprintf("Symbol(`%s`)", string(sym))
}

type Word string

func (w Word) IsEqual(key Key) bool {
	if v, ok := key.(Word); ok {
		return w == v
	}
	return false
}

func (w Word) Repr(oneline bool) string {
	return fmt.Sprintf("Word(%s)", string(w))
}

type TokenType string

const (
	TokenNone    TokenType = ""
	TokenBreak   TokenType = "Break"
	TokenSymbol  TokenType = "Symbol"
	TokenWord    TokenType = "Word"
	TokenInteger TokenType = "Integer"
	TokenString  TokenType = "String"
	TokenComment TokenType = "Comment"
)

func (typ TokenType) Bind(node *Node) {
	node.Bind(typ)
	switch typ {
	case TokenWord:
		node.Bind(Word(node.Span().Text()))
	case TokenSymbol:
		node.Bind(Symbol(node.Span().Text()))
	}
}

func (typ TokenType) IsEqual(key Key) bool {
	if v, ok := key.(TokenType); ok {
		return typ == v
	}
	return false
}

func (typ TokenType) Repr(oneline bool) string {
	return fmt.Sprintf("Token%s", string(typ))
}

type Token struct {
	Type TokenType
	Span Span
}

func (token Token) Indent() int {
	return token.Span.Indent()
}

type Lexer struct {
	symbolSet  map[string]bool
	symbolList []string
	symbolRe   *regexp.Regexp
	matchers   []Matcher
}

func NewLexer() *Lexer {
	lex := &Lexer{}

	lex.AddSymbols(
		// punctuation
		".", "..", ",", ";", ":",
		// brackets
		"(", ")", "{", "}", "[", "]",
		// operators
		"!", "?",
		"=", "+", "-", "*", "/", "%",
		"==", "!=", "<", "<=", ">", ">=",
	)

	lex.AddMatcher(MatchWord)
	lex.AddMatcher(MatchString)

	return lex
}

func (lexer *Lexer) CopyOrDefault() *Lexer {
	if lexer == nil {
		return NewLexer()
	}

	copy := &Lexer{
		symbolSet:  make(map[string]bool),
		symbolList: append([]string{}, lexer.symbolList...),
		symbolRe:   lexer.symbolRe,
		matchers:   append([]Matcher{}, lexer.matchers...),
	}

	for it := range lexer.symbolSet {
		copy.symbolSet[it] = true
	}

	return copy
}

func (lexer *Lexer) AddMatcher(m Matcher) {
	lexer.matchers = append(lexer.matchers, m)
}

func (lexer *Lexer) AddSymbols(symbols ...string) {
	if lexer.symbolSet == nil {
		lexer.symbolSet = make(map[string]bool)
	}

	set, list := lexer.symbolSet, lexer.symbolList

	changed := false
	for _, it := range symbols {
		if !set[it] {
			set[it] = true
			list = append(list, it)
			changed = true
		}
	}

	if changed {
		sort.Slice(list, func(i, j int) bool {
			a, b := list[i], list[j]
			return len(a) < len(b)
		})

		re := strings.Builder{}
		re.WriteString(`^(`)
		for n, it := range list {
			if n > 0 {
				re.WriteRune('|')
			}
			re.WriteString(regexp.QuoteMeta(it))
		}
		re.WriteString(`)`)

		lexer.symbolList = list
		lexer.symbolRe = regexp.MustCompile(re.String())
	}
}

func (lexer *Lexer) Tokenize(src *Source) (out []Token, err error) {
	cur := src.Cursor()
	for !cur.IsEnd() {
		if cur.SkipSpaces() {
			continue
		}

		var (
			sta  = cur.Pos()
			span = cur.Span()
			text = cur.Text()
		)

		token := TokenNone
		switch cur.Peek() {
		case '\r':
			token = TokenBreak
			if strings.HasPrefix(text, "\r\n") {
				cur.Advance(2)
			} else {
				cur.Advance(1)
			}
		case '\n':
			token = TokenBreak
			cur.Advance(1)
		case COMMENT_CHAR:
			token = TokenComment
			cur.SkipWhile(func(chr rune) bool { return chr != '\n' && chr != '\r' })
		case COMMENT_DELIM:
			if strings.HasPrefix(text, COMMENT_DELIM_STA) {
				token = TokenComment
				skipMultilineComment(cur)
			}
		}

		if token == TokenNone {
			for _, matcher := range lexer.matchers {
				mCur := *cur
				if token, err = matcher(&mCur); err != nil {
					return out, err
				} else if token != TokenNone {
					if mCur.Pos() <= cur.Pos() {
						panic("Lexer: matcher generated empty token")
					}
					*cur = mCur
					break
				}
			}
		}

		if token == TokenNone && lexer.symbolRe != nil {
			symbol := lexer.symbolRe.FindString(text)
			if len := len(symbol); len > 0 {
				token = TokenSymbol
				cur.Advance(len)
			}
		}

		if token == TokenNone {
			return out, cur.Error(0, "invalid token")
		}

		len := cur.Pos() - sta
		out = append(out, Token{
			Type: token,
			Span: span.Truncated(len),
		})
	}
	return
}

func skipMultilineComment(cur *Cursor) {
	noComment := func(chr rune) bool { return chr != COMMENT_CHAR && chr != COMMENT_DELIM }
	count := 0
	for !cur.IsEnd() {
		next := cur.Text()
		if strings.HasPrefix(next, COMMENT_DELIM_STA) {
			cur.Advance(len(COMMENT_DELIM_STA))
			count += 1
		} else if strings.HasPrefix(next, COMMENT_DELIM_END) {
			cur.Advance(len(COMMENT_DELIM_END))
			count -= 1
			if count == 0 {
				break
			}
		} else {
			cur.Advance(1)
			cur.SkipWhile(noComment)
		}
	}
}
