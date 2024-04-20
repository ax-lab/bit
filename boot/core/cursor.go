package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Cursor struct {
	span Span
	row  int
	col  int
	ind  int
}

func (src Source) Cursor() Cursor {
	return Cursor{span: src.Span()}
}

func (cur *Cursor) Span() Span {
	return cur.span
}

func (cur *Cursor) Empty() bool {
	return cur.Len() == 0
}

func (cur *Cursor) Text() string {
	return cur.span.Text()
}

func (cur *Cursor) Len() int {
	return cur.span.Len()
}

func (cur *Cursor) Line() int {
	return cur.row + 1
}

func (cur *Cursor) Column() int {
	return cur.col + 1
}

func (cur *Cursor) Indent() int {
	return cur.ind
}

func (cur *Cursor) Offset() int {
	return cur.span.Sta()
}

func (cur *Cursor) IsLineStart() bool {
	return cur.col == 0
}

func (cur *Cursor) IsLineEmpty() bool {
	return cur.col == cur.ind
}

func (cur *Cursor) Peek() rune {
	for _, chr := range cur.Text() {
		return chr
	}
	return 0
}

func (cur *Cursor) Read() rune {
	txt := cur.Text()
	if len(txt) == 0 {
		return 0
	}

	out, len := utf8.DecodeRuneInString(txt)
	if len > 0 {
		if out == utf8.RuneError {
			panic(fmt.Sprintf("Cursor: invalid UTF-8 in %s", cur.span.Src().Name()))
		}
		cur.Advance(len)
	} else {
		panic("Cursor: read at  the end of the input")
	}

	return out
}

func (cur *Cursor) ReadIf(str string) bool {
	if str == "" {
		return false
	}

	if txt := cur.Text(); strings.HasPrefix(txt, str) {
		cur.Advance(len(str))
		return true
	}
	return false
}

func (cur *Cursor) ReadAny(ls ...string) string {
	for _, it := range ls {
		if cur.ReadIf(it) {
			return it
		}
	}
	return ""
}

func (cur *Cursor) PeekChars(count int) string {
	bytes := 0
	text := cur.Text()
	for count > 0 {
		_, n := utf8.DecodeRuneInString(text[bytes:])
		bytes += n
		count -= 1
		if n == 0 {
			break
		}
	}
	return text[:bytes]
}

func (cur *Cursor) ReadChars(count int) (out string) {
	out = cur.PeekChars(count)
	cur.Advance(len(out))
	return out
}

func (cur *Cursor) ReadWhile(match Matcher) (out string) {
	var bytes int
	text := cur.Text()
	for {
		next := match.MatchNext(text[bytes:])
		if next > 0 {
			bytes += next
		} else {
			break
		}
	}
	out = text[:bytes]
	cur.Advance(bytes)
	return out
}

func (cur *Cursor) ReadUntil(match Matcher) (out string) {
	text := cur.Text()
	pos := match.FindMatchIndex(text)
	out = text[:pos]
	cur.Advance(pos)
	return out
}

func (cur *Cursor) ReadMatch(match Matcher) (prefix, matched string) {
	text := cur.Text()
	sta, end := match.FindMatch(text)
	if sta == end && sta < len(text) {
		panic("Scanner.ReadMatch: invalid match for the empty string")
	}
	prefix = text[:sta]
	matched = text[sta:end]
	cur.Advance(end)
	return
}

func (cur *Cursor) SkipWhile(skip func(chr rune) bool) int {
	sta := cur.Offset()
	for cur.Len() > 0 && skip(cur.Peek()) {
		cur.Read()
	}
	return cur.Offset() - sta
}

func (cur *Cursor) SkipSpaces() int {
	return cur.SkipWhile(IsSpace)
}

func (cur *Cursor) Advance(length int) {
	if length == 0 {
		return
	}

	text := cur.Text()
	if length > len(text) {
		panic("Cursor: advance length out of bounds")
	}

	cr := false
	tw := cur.span.Src().TabSize()

	offset := 0
	for pos, chr := range text {
		if pos != offset {
			panic("Cursor: invalid calculated offset")
		}

		if chr == '\r' || chr == '\n' {
			if !cr || chr == '\r' {
				cur.row += 1
				cur.col = 0
				cur.ind = 0
			}
			cr = chr == '\r'
		} else if chr == '\t' {
			indent := cur.ind == cur.col
			cur.col += tw - (cur.col % tw)
			if indent {
				cur.ind = cur.col
			}
		} else {
			if IsSpace(chr) && cur.ind == cur.col {
				cur.ind += 1
			}
			cur.col += 1
		}

		size := utf8.RuneLen(chr)
		if size <= 0 {
			panic(fmt.Sprintf("Cursor: invalid rune U+%06X in %s", chr, cur.span.Src().Name()))
		}
		offset += size

		if offset >= length {
			break
		}
	}

	if offset != length {
		panic(fmt.Sprintf("Cursor: invalid advance length (expected %d, was %d)", length, offset))
	}

	if cr && offset < len(text) && text[offset] == '\n' {
		offset += 1
	}

	cur.span = cur.span.Skip(offset)
}
