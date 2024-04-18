package core

import (
	"io"
	"strings"
)

type FormatWriter struct {
	lvl int
	col int
	tab int
	out io.StringWriter

	useSpaces bool
}

func FormatWriterNew(out io.StringWriter) FormatWriter {
	return FormatWriter{out: out}
}

func (wr *FormatWriter) TabSize() int {
	if wr.tab > 0 {
		return wr.tab
	}
	return int(DefaultTabSize)
}

func (wr *FormatWriter) SetTabSize(size int) {
	wr.tab = size
}

func (wr *FormatWriter) UseSpaces(useSpaces bool) {
	wr.useSpaces = useSpaces
}

func (wr *FormatWriter) Indent() {
	wr.lvl++
}

func (wr *FormatWriter) Dedent() {
	if wr.lvl == 0 {
		panic("dedent without indentation")
	}
	wr.lvl--
}

func (wr *FormatWriter) WriteString(output string) (length int, err error) {
	tabs := wr.TabSize()
	length = len(output)
	for err == nil && len(output) > 0 {
		eol := 0
		if strings.HasPrefix(output, "\r\n") {
			eol = 2
		} else if output[0] == '\r' || output[0] == '\n' {
			eol = 1
		}

		if eol > 0 {
			output = output[eol:]
			wr.col = 0
			_, err = wr.out.WriteString("\n")
			continue
		}

		if wr.col == 0 && wr.lvl > 0 {
			wr.col = wr.lvl * tabs
			if wr.useSpaces {
				lvl := wr.col
				for err == nil && lvl > 0 {
					lvl--
					_, err = wr.out.WriteString(" ")
				}
			} else {
				lvl := wr.lvl
				for err == nil && lvl > 0 {
					lvl--
					_, err = wr.out.WriteString("\t")
				}
			}
		} else if output[0] == '\t' {
			output = output[1:]
			width := tabs - (wr.col % tabs)
			wr.col += width
			if wr.useSpaces {
				for err == nil && width > 0 {
					width--
					_, err = wr.out.WriteString(" ")
				}
			} else {
				_, err = wr.out.WriteString("\t")
			}
		} else {
			chunk := output
			if next := strings.IndexAny(output, "\r\n\t"); next >= 0 {
				chunk = output[:next]
			}
			output = output[len(chunk):]

			_, err = wr.out.WriteString(chunk)
			for range chunk {
				wr.col++
			}
		}
	}

	length -= len(output)
	return length, err
}
