package boot

type Span struct {
	src *Source
	sta int
	end int
}

func (src *Source) Span() Span {
	return Span{
		src: src,
		sta: 0,
		end: len(src.Text()),
	}
}

func (span Span) Src() *Source {
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
