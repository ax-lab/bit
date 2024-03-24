package input

import (
	"cmp"
	"fmt"
	"strings"
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

func (span Span) Merged(other Span) Span {
	out := span
	out.Merge(other)
	return out
}

func (span *Span) Merge(other Span) {
	if span.src != other.src {
		panic("Span: cannot merge from different sources")
	}
	if !span.src.Valid() {
		return
	}

	span.sta = min(span.sta, other.sta)
	span.end = max(span.end, other.end)
}

func (span Span) ExtendedTo(cursor *Cursor) Span {
	out := span
	out.ExtendTo(cursor)
	return out
}

func (span *Span) ExtendTo(cursor *Cursor) {
	if span.src != cursor.span.src {
		panic("Span: cannot extend to cursor from different source")
	}

	if !span.src.Valid() {
		panic("Span: cannot extend invalid span")
	}

	offset := cursor.Offset()
	if offset < span.sta {
		panic("Span: cannot extend to cursor out of range")
	}

	span.end = offset
}

func (span Span) NewError(msg string, args ...any) ErrorWithLocation {
	return Error(msg, args...).AtLocation(span.Location())
}

func (span Span) ErrorAt(err error) ErrorWithLocation {
	return Error(err.Error()).AtLocation(span.Location())
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

func (span Span) Cmp(other Span) int {
	if res := span.src.Cmp(other.src); res != 0 {
		return res
	}
	if res := cmp.Compare(span.sta, other.sta); res != 0 {
		return res
	}
	return cmp.Compare(span.Len(), other.Len())
}

func Location(file string, pos ...int) string {
	var rowSta, colSta, rowEnd, colEnd int

	valid := true
	for i, it := range pos {
		if it < 0 {
			valid = false
		}
		switch i {
		case 0:
			rowSta = it
		case 1:
			colSta = it
		case 2:
			rowEnd = it
		case 3:
			colEnd = it
		default:
			valid = false
		}

		if !valid {
			break
		}
	}

	valid = valid &&
		((rowEnd == 0 || rowEnd >= rowSta) &&
			(colEnd == 0 || colEnd >= colSta || rowEnd != 0))

	if !valid {
		panic("Location: invalid position")
	}

	out := strings.Builder{}
	if file != "" {
		out.WriteString(file)
	}

	if rowSta > 0 {
		if out.Len() == 0 {
			out.WriteString("@ ")
		} else {
			out.WriteString(" @ ")
		}

		out.WriteString(fmt.Sprintf("L%03d", rowSta))
		if colSta > 0 {
			out.WriteString(fmt.Sprintf(":%02d", colSta))
		}

		if rowEnd > rowSta {
			out.WriteString(fmt.Sprintf("…L%03d", rowEnd))
			if colEnd > 0 {
				out.WriteString(fmt.Sprintf(":%02d", colEnd))
			}
		} else if (rowEnd == 0 || rowEnd == rowSta) && colEnd > colSta {
			out.WriteString(fmt.Sprintf("…%02d", colEnd))
		}
	}

	return out.String()
}
