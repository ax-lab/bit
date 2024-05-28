package core

import "strings"

type Cursor struct {
	Span
}

func (span Span) Cursor() Cursor {
	return Cursor{span}
}

func (cursor *Cursor) GetSpan(other Cursor) (span Span) {
	if cursor.sta <= other.sta {
		span = cursor.Span
		span.end = other.sta
	} else {
		span = other.Span
		span.end = cursor.sta
	}
	return span
}

func (cursor *Cursor) Location() Span {
	return cursor.Span.WithSize(0)
}

func (cursor *Cursor) ToSpan() Span {
	return cursor.Span
}

func (cursor *Cursor) Peek() rune {
	for _, chr := range cursor.Text() {
		return chr
	}
	return 0
}

func (cursor *Cursor) Read() (out rune) {
	size := 0
	for idx, chr := range cursor.Text() {
		if idx == 0 {
			out = chr
		} else {
			size = idx
			break
		}
	}
	cursor.Advance(size)
	return out
}

func (cursor *Cursor) ReadChar(chr ...rune) bool {
	next := cursor.Peek()
	for _, it := range chr {
		if next == it {
			cursor.Read()
			return true
		}
	}
	return false
}

func (cursor *Cursor) ReadIf(prefix string) bool {
	return cursor.SkipAny(prefix)
}

func (cursor *Cursor) SkipSpaces() {
	cursor.SkipWhile(IsSpace)
}

func (cursor *Cursor) SkipWhile(pred func(chr rune) bool) (skipped bool) {
	txt := cursor.ReadWhile(pred)
	return len(txt) > 0
}

func (cursor *Cursor) ReadWhile(pred func(chr rune) bool) (read string) {
	txt := cursor.src.Text()
	sta := cursor.sta
	for cursor.Len() > 0 && pred(cursor.Peek()) {
		cursor.Read()
	}
	return txt[sta:cursor.sta]
}

func (cursor *Cursor) SkipAny(prefixes ...string) (skipped bool) {
	read := cursor.ReadAny(prefixes...)
	return len(read) > 0
}

func (cursor *Cursor) ReadAny(prefixes ...string) (read string) {
	if txt := cursor.Text(); len(txt) > 0 {
		for _, it := range prefixes {
			if len(it) > 0 && strings.HasPrefix(txt, it) {
				cursor.Advance(len(it))
				return it
			}
		}
	}
	return ""
}

func (cursor *Cursor) ReadFrom(list []string) (index int) {
	if txt := cursor.Text(); len(txt) > 0 {
		for idx, it := range list {
			if len(it) > 0 && strings.HasPrefix(txt, it) {
				cursor.Advance(len(it))
				return idx
			}
		}
	}
	return -1
}

func (cursor *Cursor) Advance(bytes int) {
	if bytes == 0 {
		return
	} else if bytes > cursor.Len() {
		panic("Cursor: advance outside of cursor bounds")
	}

	cr := false
	for idx, chr := range cursor.Text() {
		if idx == bytes {
			break
		} else if idx > bytes {
			panic("Cursor: advance outside of char boundaries")
		}

		eol := false
		ind := cursor.column == cursor.indent && IsSpace(chr)
		if chr == '\n' {
			if !cr {
				eol = true
			} else {
				cr = false
			}
		} else if cr = chr == '\r'; cr {
			eol = true
		}

		if eol {
			cursor.line++
			cursor.column = 0
			cursor.indent = 0
		} else if chr == '\t' {
			tab := cursor.src.TabSize()
			cursor.column += tab - (cursor.column % tab)
		} else {
			cursor.column++
		}

		if ind {
			cursor.indent = cursor.column
		}
	}

	cursor.sta += bytes
}
