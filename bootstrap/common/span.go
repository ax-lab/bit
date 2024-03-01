package common

import (
	"fmt"
	"strings"
)

type Span struct {
	src *Source
	loc Location
	sta int
	end int
}

func (span Span) Source() *Source {
	return span.src
}

func (span Span) Cursor() *Cursor {
	return &Cursor{span}
}

func (span Span) Location() Location {
	return span.loc
}

func (span Span) Indent() int {
	return span.loc.Indent()
}

func (span Span) Sta() int {
	return span.sta
}

func (span Span) End() int {
	return span.end
}

func (span *Span) SetEnd(end int) {
	if end < span.sta {
		panic("invalid span end")
	}
	span.end = end
}

func (span Span) Len() int {
	return span.end - span.sta
}

func (span Span) Text() string {
	text := span.src.Text()
	return text[span.sta:span.end]
}

func (span Span) Compare(other Span) int {
	if cmp := span.src.Compare(other.src); cmp != 0 {
		return cmp
	} else if span.sta < other.sta {
		return -1
	} else if other.sta < span.sta {
		return +1
	} else if span.end < other.end {
		return -1
	} else if other.end < span.end {
		return +1
	} else {
		return 0
	}
}

func (span Span) CreateError(msg string, args ...any) error {
	err := ErrorWithLocation{
		Span:    span,
		Message: msg,
		Args:    args,
	}
	return err
}

func (span Span) DisplayText(maxChars int) string {
	if maxChars == 0 {
		maxChars = 16
	}

	var text string
	if span.Len() == 0 {
		text = span.Source().Text()[span.sta:]
	} else {
		text = span.Text()
	}

	trimL := false
	trimR := false

	if idx := strings.IndexAny(text, "\r\n"); idx >= 0 {
		text = text[:idx]
		trimR = true
	}

	if trimmed := strings.TrimRightFunc(text, IsSpace); len(trimmed) != len(text) {
		text = trimmed
		trimR = true
	}

	if trimmed := strings.TrimLeftFunc(text, IsSpace); len(trimmed) != len(text) {
		text = trimmed
		trimL = true
	}

	cnt := 0
	for pos := range text {
		cnt += 1
		if cnt > maxChars {
			text = text[:pos]
			trimR = true
			break
		}
	}

	if text == "" {
		return ""
	}

	if trimL || trimR {
		pre, pos := "", ""
		if trimL {
			pre = "…"
		}
		if trimR {
			pos = "…"
		}
		text = fmt.Sprintf("%s%s%s", pre, text, pos)
	}

	return text
}

func (span Span) Truncated(len int) Span {
	span.end = min(span.end, span.sta+len)
	return span
}

func (span Span) Merged(other Span) Span {
	if span.src != other.src {
		panic("merging Span from different source")
	}
	span.sta = min(span.sta, other.sta)
	span.end = max(span.end, other.end)
	return span
}

func (span Span) String() string {
	if len := span.Len(); len > 0 {
		return fmt.Sprintf("%s:%s+%d", span.src.Name(), span.loc.String(), len)
	} else {
		return fmt.Sprintf("%s:%s", span.src.Name(), span.loc.String())
	}
}

type Location struct {
	row int
	col int
	ind int
}

func (loc Location) Line() int {
	return loc.row + 1
}

func (loc Location) Column() int {
	return loc.col + 1
}

func (loc Location) Indent() int {
	return loc.ind
}

func (loc Location) String() string {
	return fmt.Sprintf("%d:%d", loc.Line(), loc.Column())
}

func (loc *Location) Advance(tabWidth uint32, text string) {
	wasCR := false
	tab := int(tabWidth)
	for _, chr := range text {
		if chr == '\r' || chr == '\n' {
			if chr == '\n' && wasCR {
				continue
			}
			wasCR = chr == '\r'
			loc.row += 1
			loc.col = 0
			loc.ind = 0
		} else {
			wasCR = false
			indenting := loc.col == loc.ind && IsSpace(chr)
			if chr == '\t' {
				loc.col += tab - (loc.col % tab)
			} else {
				loc.col += 1
			}
			if indenting {
				loc.ind = loc.col
			}
		}
	}
}

type Cursor struct {
	span Span
}

func (cur *Cursor) Span() Span {
	return cur.span
}

func (cur *Cursor) Error(size int, msg string, args ...any) error {
	err := cur.span.Truncated(size).CreateError(msg, args...)
	cur.Advance(size)
	return err
}

func (cur *Cursor) Advance(len int) {
	tab := cur.span.src.TabWidth()
	txt := cur.span.Text()[:len]
	cur.span.loc.Advance(tab, txt)
	cur.span.sta += len
}

func (cur *Cursor) Text() string {
	return cur.span.Text()
}

func (cur *Cursor) Pos() int {
	return cur.span.Sta()
}

func (cur *Cursor) End() int {
	return cur.span.End()
}

func (cur *Cursor) Len() int {
	return cur.span.Len()
}

func (cur *Cursor) IsEnd() bool {
	return cur.Len() == 0
}

func (cur *Cursor) Peek() rune {
	for _, chr := range cur.Text() {
		return chr
	}
	return 0
}

func (cur *Cursor) Read() rune {
	var (
		out rune
		len int
	)
	for pos, chr := range cur.Text() {
		if pos == 0 {
			out = chr
		} else {
			len = pos
			break
		}
	}
	cur.Advance(len)
	return out
}

func (cur *Cursor) ReadIf(str string) bool {
	if len(str) > 0 && strings.HasPrefix(cur.Text(), str) {
		cur.Advance(len(str))
		return true
	}
	return false
}

func (cur *Cursor) ReadAny(str ...string) string {
	for _, it := range str {
		if cur.ReadIf(it) {
			return it
		}
	}
	return ""
}

func (cur *Cursor) SkipSpaces() bool {
	return cur.SkipWhile(IsSpace) > 0
}

func (cur *Cursor) SkipWhile(cond func(rune) bool) int {
	text := cur.Text()
	skip := strings.TrimLeftFunc(text, cond)
	size := len(text) - len(skip)
	if size > 0 {
		cur.Advance(size)
	}
	return size
}