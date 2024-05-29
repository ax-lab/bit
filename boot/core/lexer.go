package core

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

type LexMatcher func(mod *Module, lexer *Lexer, input *Cursor) Value

type LexSegmenter func(mod *Module, lexer *Lexer, input *Cursor) Node

type Lexer struct {
	sync     sync.Mutex
	matchers []LexMatcher

	segmenter LexSegmenter

	brackets   map[string]bool
	bracketSta map[string]string
	bracketEnd map[string]string

	symbolMap map[string]bool
	symbols   []string
}

func (lex *Lexer) Copy() *Lexer {
	lex.sync.Lock()
	defer lex.sync.Unlock()
	out := &Lexer{
		segmenter: lex.segmenter,
	}

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

func (lex *Lexer) Tokenize(mod *Module, input *Cursor) (out []Node) {
	rt := mod.runtime
	for input.Len() > 0 && !rt.ShouldStop() {
		var next Node
		if lex.segmenter != nil {
			next = lex.segmenter(mod, lex, input)
		} else {
			next = lex.Read(mod, input)
		}

		if !next.Valid() {
			if input.Len() > 0 {
				panic("lexer read returned an invalid node")
			}
			break
		}
		out = append(out, next)
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

func (lex *Lexer) SetSegmenter(segmenter LexSegmenter) {
	lex.sync.Lock()
	defer lex.sync.Unlock()
	lex.segmenter = segmenter
}

func (lex *Lexer) Read(mod *Module, input *Cursor) (out Node) {
	out, valid := lex.readNext(mod, input)
	if valid {
		return out
	}

	sta := *input
	input.Read()
	lex.skipInvalid(input)
	span := input.GetSpan(sta)
	out = NodeNew(span, Invalid(span.Text()))
	return out
}

func (lex *Lexer) readNext(mod *Module, input *Cursor) (out Node, valid bool) {
	input.SkipWhile(IsSpace)
	sta := *input

	text := input.Text()
	if len(text) == 0 {
		return out, true
	}

	if input.SkipAny("\n", "\r\n", "\r") {
		span := input.GetSpan(sta)
		out = NodeNew(span, LineBreak(span.Text()))
		return out, true
	}

	for _, matchFunc := range lex.matchers {
		cur := *input
		val := matchFunc(mod, lex, &cur)
		if val != nil {
			if cur == sta {
				panic("Lexer: matcher generated empty span")
			}
			*input = cur
			span := input.GetSpan(sta)
			out = NodeNew(span, val)
			return out, true
		}
	}

	for _, sym := range lex.symbols {
		if strings.HasPrefix(text, sym) {
			input.Advance(len(sym))
			span := input.GetSpan(sta)
			out = NodeNew(span, Symbol(sym))
			return out, true
		}
	}

	return
}

func (lex *Lexer) skipInvalid(input *Cursor) {
	cur := *input
	end := *input
	tmp := lex.Copy()
	for cur.Len() > 0 {
		if chr := cur.Peek(); IsSpace(chr) || chr == '\r' || chr == '\n' {
			break
		}
		_, valid := tmp.readNext(nil, &cur)
		if valid {
			break
		} else {
			cur.Read()
			end = cur
		}
	}
	*input = end
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
