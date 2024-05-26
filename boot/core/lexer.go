package core

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
)

type LexMatcher func(input *Cursor) (Value, error)

type Lexer struct {
	sync     sync.Mutex
	matchers []LexMatcher

	brackets   map[string]bool
	bracketSta map[string]string
	bracketEnd map[string]string

	symbolMap map[string]bool
	symbols   []string
}

func (lex *Lexer) Copy() *Lexer {
	lex.sync.Lock()
	defer lex.sync.Unlock()
	out := &Lexer{}

	out.matchers = append(out.matchers, lex.matchers...)
	if len(lex.bracketSta) > 0 {
		out.initBrackets()
		for k, v := range out.bracketSta {
			out.brackets[k] = true
			out.brackets[v] = true
			out.bracketSta[k] = v
			out.bracketEnd[v] = k
		}
	}

	if len(lex.symbolMap) > 0 {
		out.symbols = append(out.symbols, lex.symbols...)
		out.symbolMap = make(map[string]bool)
		for k := range lex.symbolMap {
			out.symbolMap[k] = true
		}
	}

	return out
}

func (lex *Lexer) AddMatcher(matcher LexMatcher) {
	lex.sync.Lock()
	defer lex.sync.Unlock()
	lex.matchers = append(lex.matchers, matcher)
}

func (lex *Lexer) AddSymbols(symbols ...string) {
	lex.sync.Lock()
	defer lex.sync.Unlock()
	lex.registerSymbols(symbols)
}

func (lex *Lexer) AddBrackets(sta, end string) {
	lex.sync.Lock()
	defer lex.sync.Unlock()
	lex.registerSymbols([]string{sta, end})

	lex.initBrackets()
	lex.brackets[sta] = true
	lex.brackets[end] = true
	lex.bracketSta[sta] = end
	lex.bracketSta[end] = sta
}

func (lexer *Lexer) Read(input *Cursor) (out Node, err error) {
	input.SkipWhile(IsSpace)
	sta := *input

	text := input.Text()
	if len(text) == 0 {
		return out, io.EOF
	}

	if input.SkipAny("\n", "\r\n", "\r") {
		span := input.GetSpan(sta)
		out = NodeNew(span, LineBreak(span.Text()))
		return
	}

	for _, matchFunc := range lexer.matchers {
		cur := *input
		val, err := matchFunc(&cur)
		if val != nil || err != nil {
			if cur == sta {
				panic("Lexer: matcher generated empty span")
			}
			*input = cur
			span := input.GetSpan(sta)
			if val != nil {
				out = NodeNew(span, val)
			}
			if err != nil {
				err = ErrorAt(span, err)
			}
			return out, err
		}
	}

	for _, sym := range lexer.symbols {
		if strings.HasPrefix(text, sym) {
			input.Advance(len(sym))
			span := input.GetSpan(sta)
			out = NodeNew(span, Symbol(sym))
			return out, nil
		}
	}

	input.Read()
	span := input.GetSpan(sta)
	out = NodeNew(span, Invalid(span.Text()))
	err = Errorf(span, "invalid token")
	return out, err
}

func (lex *Lexer) initBrackets() {
	if lex.brackets == nil {
		lex.brackets = make(map[string]bool)
		lex.bracketSta = make(map[string]string)
		lex.bracketEnd = make(map[string]string)
	}
}

func (lex *Lexer) registerSymbols(symbols []string) {
	if lex.symbolMap == nil {
		lex.symbolMap = make(map[string]bool)
	}

	changed := false
	for _, it := range symbols {
		sym := it
		if sym == "" || len(sym) != len(strings.TrimSpace(sym)) {
			panic(fmt.Sprintf("invalid symbol `%s`", sym))
		}

		if !lex.symbolMap[sym] {
			lex.symbolMap[sym] = true
			lex.symbols = append(lex.symbols, sym)
			changed = false
		}
	}

	if changed {
		slices.SortFunc(lex.symbols, func(a, b string) int {
			return len(b) - len(a)
		})
	}
}
