package core

import (
	"fmt"
	"strings"
)

type CodeText struct {
	text   strings.Builder
	level  int
	tabs   int
	column int
}

func (code *CodeText) String() string {
	code.NewLine()
	out := code.text.String()
	if strings.HasSuffix(out, "\n\n") {
		out = out[:len(out)-1]
	}
	return out
}

func (code *CodeText) Append(other *CodeText) {
	code.NewLine()
	code.writeText(other.String())
}

func (code *CodeText) TabSize() int {
	if code.tabs == 0 {
		return DefaultTabSize
	}
	return code.tabs
}

func (code *CodeText) Indent() {
	code.level += 1
}

func (code *CodeText) Dedent() {
	if code.level == 0 {
		panic("Code: invalid dedent")
	}
	code.level -= 1
}

func (code *CodeText) NewLine() {
	if code.column > 0 {
		code.writeLineBreak()
	}
}

func (code *CodeText) BlankLine() {
	code.NewLine()
	if !strings.HasSuffix(code.text.String(), "\n\n") {
		code.writeLineBreak()
	}
}

func (code *CodeText) WriteLine(txt string, args ...any) {
	code.NewLine()
	code.Write(txt, args...)
	code.NewLine()
}

func (code *CodeText) Write(txt string, args ...any) {
	if len(args) > 0 {
		txt = fmt.Sprintf(txt, args...)
	}
	code.writeText(txt)
}

func (code *CodeText) writeText(txt string) {
	tab := code.TabSize()
	for n, line := range reLines.Split(txt, -1) {
		if n > 0 {
			code.writeLineBreak()
		}
		if line = TrimEnd(line); len(line) == 0 {
			continue
		}

		if code.column == 0 && code.level > 0 {
			code.writeIndent()
		}
		code.text.WriteString(line)

		for _, chr := range line {
			if chr == '\t' {
				code.column += tab - (code.column % tab)
			} else {
				code.column++
			}
		}
	}
}

func (code *CodeText) writeIndent() {
	tab := code.TabSize()
	for i := 0; i < code.level; i++ {
		code.text.WriteRune('\t')
		code.column += tab - (code.column % tab)
	}
}

func (code *CodeText) writeLineBreak() {
	code.text.WriteRune('\n')
	code.column = 0
}
