package core

import "cmp"

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

func (span Span) Merged(other Span) Span {
	out := span
	out.Merge(other)
	return out
}

func (span *Span) Merge(other Span) {
	if !other.src.Valid() {
		return
	}

	if !span.src.Valid() && span.sta == 0 && span.end == 0 {
		*span = other
		return
	}

	if span.src != other.src {
		panic("Span: cannot merge from different sources")
	}
	if !span.src.Valid() || !other.src.Valid() {
		return
	}

	span.sta = min(span.sta, other.sta)
	span.end = max(span.end, other.end)
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

func (span Span) WithLen(len int) Span {
	return span.Range(0, len)
}

func (span Span) From(offset int) Span {
	len := span.Len()
	if offset < 0 || len < offset {
		panic("Span: invalid skip offset")
	}
	return span.Range(offset, len)
}

func (span Span) ErrorAt(err error) error {
	return ErrorAt(err, span.Location())
}

func (span Span) Compare(other Span) int {
	if res := span.src.Compare(other.src); res != 0 {
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

func (span Span) Contains(other Span) bool {
	if !other.src.Valid() || other.src != span.src {
		return false
	}
	return span.sta <= other.sta && other.sta < span.end &&
		span.sta <= other.end && other.end <= span.end
}

func (span Span) Location() LocationPos {
	cur := span.Src().Cursor()
	cur.Advance(span.Sta())

	var (
		rowSta = cur.Line()
		colSta = cur.Column()
		rowEnd = 0
		colEnd = 0
	)
	if len := span.Len(); len > 0 {
		cur.Advance(len)
		rowEnd = cur.Line()
		colEnd = cur.Column()
	}
	loc := Location(span.Src().Name(), rowSta, colSta, rowEnd, colEnd)
	return loc
}
