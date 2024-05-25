package core

import "cmp"

type Span struct {
	src Source
	sta int
	end int
}

func spanForSource(src Source) Span {
	return Span{
		src: src,
		sta: 0,
		end: len(src.Text()),
	}
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
