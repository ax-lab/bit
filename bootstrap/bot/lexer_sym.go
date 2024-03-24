package bot

import (
	"slices"
	"strings"

	"axlab.dev/bit/input"
)

type SymbolTable struct {
	symbols map[string]bool
	sorted  []string
}

func (tb *SymbolTable) Add(symbols ...string) {
	for _, symbol := range symbols {
		symbol = strings.TrimSpace(symbol)
		if symbol == "" {
			panic("SymbolTable: invalid empty symbol")
		}

		if !tb.symbols[symbol] {
			if tb.symbols == nil {
				tb.symbols = make(map[string]bool)
			}
			tb.symbols[symbol] = true
			tb.sorted = append(tb.sorted, symbol)
			slices.SortFunc(tb.sorted, func(a, b string) int {
				return len(b) - len(a)
			})
		}
	}
}

func (tb *SymbolTable) Read(cursor *input.Cursor) string {
	for _, sym := range tb.sorted {
		if cursor.ReadString(sym) {
			return sym
		}
	}
	return ""
}
