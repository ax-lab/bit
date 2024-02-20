package common

type HasSpan interface {
	Span() Span
}

func SpanFromSlice[T HasSpan](elems []T) Span {
	switch len(elems) {
	case 0:
		panic("cannot get span for empty slice")
	case 1:
		return elems[0].Span()
	default:
		return elems[0].Span().Merged(elems[len(elems)-1].Span())
	}
}

func SpanFromRange[T HasSpan](elems ...T) Span {
	return SpanFromSlice(elems)
}
