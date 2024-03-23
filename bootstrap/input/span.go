package input

import (
	"cmp"
	"fmt"
)

type Span struct {
	src Source
	sta int
	end int
}

func (src Source) Span() Span {
	return Span{
		src: src,
		sta: 0,
		end: len(src.Text()),
	}
}

func (src Source) Range(sta, end int) Span {
	if sta < 0 || sta > end || end > len(src.Text()) {
		panic("Source range out of bounds")
	}

	return Span{
		src: src,
		sta: sta,
		end: end,
	}
}

func (span Span) Src() Source {
	return span.src
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

func (span Span) Text() string {
	txt := span.src.Text()
	return txt[span.sta:span.end]
}

func (span Span) Range(sta, end int) Span {
	sta += span.sta
	end += span.sta
	if sta > end || sta < span.sta || span.end < end {
		panic("Span: invalid slice bounds")
	}
	return Span{span.src, sta, end}
}

func (span Span) Skip(offset int) Span {
	len := span.Len()
	if offset < 0 || len < offset {
		panic("Span: invalid skip offset")
	}
	return span.Range(offset, len)
}

func (span Span) Location() string {
	cur := span.Src().Cursor()
	cur.Advance(span.Sta())

	loc := fmt.Sprintf("%s @ L%03d:%02d", span.Src().Name(), cur.Line(), cur.Column())
	if len := span.Len(); len > 0 {
		cur.Advance(len)
		loc += fmt.Sprintf(" … L%03d:%02d (+%d)", cur.Line(), cur.Column(), len)
	}
	return loc
}

func (span Span) Cmp(other Span) int {
	if res := span.src.Cmp(other.src); res != 0 {
		return res
	}
	if res := cmp.Compare(span.sta, other.sta); res != 0 {
		return res
	}
	return cmp.Compare(span.Len(), other.Len())
}
