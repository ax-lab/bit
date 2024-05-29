package core

import (
	"cmp"
	"fmt"
)

type withSpan interface {
	Span() Span
}

func GetSpan[T any](value ...T) (Span, bool) {
	out := Span{}
	for _, it := range value {
		val := any(it)
		if span, ok := val.(Span); ok {
			out = out.Merged(span)
		} else if withSpan, ok := val.(withSpan); ok {
			span := withSpan.Span()
			out = out.Merged(span)
		}
	}

	return out, out.Valid()
}

func SpanForRange[E ~[]T, T any](ls E) Span {
	out := Span{}
	if len(ls) == 0 {
		return out
	}

	sta, _ := GetSpan(ls[0])
	end, _ := GetSpan(ls[len(ls)-1])
	return sta.Merged(end)
}

type Span struct {
	src Source
	sta int
	end int

	line   int
	column int
	indent int
}

func spanForSource(src Source) Span {
	return Span{
		src: src,
		end: len(src.Text()),
	}
}

func (span Span) Valid() bool {
	return span.src != nil
}

func (span Span) Sta() int {
	return span.sta
}

func (span Span) End() int {
	return span.end
}

func (span Span) Len() int {
	return span.end - span.sta
}

func (span Span) Src() Source {
	return span.src
}

func (span Span) Text() string {
	if span.src == nil {
		return ""
	}
	txt := span.src.Text()
	return txt[span.sta:span.end]
}

func (span Span) Line() int {
	return span.line
}

func (span Span) Column() int {
	return span.column
}

func (span Span) Indent() int {
	return span.indent
}

func (span Span) Merged(other Span) (out Span) {
	if !span.Valid() {
		return other
	} else if !other.Valid() {
		return span
	}

	if span.src != other.src {
		panic("Span from different sources cannot be merged")
	}

	if span.sta <= other.sta {
		out = span
	} else {
		out = other
	}
	out.end = max(span.end, other.end)
	return
}

func (span Span) WithSize(size int) Span {
	if size < 0 || size > span.Len() {
		panic("Span: size out of bounds")
	}
	out := span
	out.end = out.sta + size
	return out
}

func (span Span) Location() string {
	if span.src == nil {
		if span.sta != 0 || span.end != 0 || span.line != 0 || span.column != 0 {
			panic("invalid span location")
		}
		return ""
	}
	line := span.line + 1
	column := span.column + 1

	size := ""
	if bytes := span.Len(); bytes > 0 {
		size = fmt.Sprintf("+%d", bytes)
	}

	return fmt.Sprintf("%s:%d:%d%s", span.src.Name(), line, column, size)
}

func (span Span) String() string {
	return span.Location()
}

func (span Span) Compare(other Span) int {
	if res := SourceCompare(span.src, other.src); res != 0 {
		return res
	}

	if res := cmp.Compare(span.sta, other.sta); res != 0 {
		return res
	}

	if res := cmp.Compare(span.end, other.end); res != 0 {
		return res
	}

	return 0
}
