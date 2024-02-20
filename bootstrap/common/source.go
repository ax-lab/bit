package common

import (
	"fmt"
	"strings"
	"sync/atomic"
	"unsafe"
)

type Source struct {
	name     string
	text     string
	file     *DirFile
	tabWidth atomic.Uint32
}

func (file *DirFile) CreateSource() *Source {
	return &Source{
		file: file,
		name: file.Name(),
		text: file.Text(),
	}
}

func (src *Source) IsEqual(key any) bool {
	if v, ok := key.(*Source); ok {
		return src == v
	}
	return false
}

func (src *Source) Repr(oneline bool) string {
	return fmt.Sprintf("Source(%s)", src.name)
}

func (src *Source) Name() string {
	return src.name
}

func (src *Source) Text() string {
	return src.text
}

func (src *Source) Len() int {
	return len(src.text)
}

func (src *Source) Dir() Dir {
	return src.file.Dir()
}

func (src *Source) File() *DirFile {
	return src.file
}

func (src *Source) Compare(other *Source) int {
	if src == other {
		return 0
	}

	srcHasFile := src.file != nil
	otherHasFile := other.file != nil
	if srcHasFile != otherHasFile {
		if srcHasFile {
			return -1
		} else {
			return +1
		}
	}

	if cmp := strings.Compare(src.name, other.name); cmp != 0 {
		return cmp
	}

	srcPtr := uintptr(unsafe.Pointer(src))
	otherPtr := uintptr(unsafe.Pointer(other))
	if srcPtr < otherPtr {
		return -1
	} else {
		return +1
	}
}

func (src *Source) Span() Span {
	return Span{
		src: src,
		loc: Location{},
		sta: 0,
		end: len(src.text),
	}
}

func (src *Source) Cursor() *Cursor {
	return src.Span().Cursor()
}

func (src *Source) TabWidth() uint32 {
	tw := src.tabWidth.Load()
	if tw == 0 {
		tw = DefaultTabSize
	}
	return tw
}

func (src *Source) SetTabWidth(tabWidth uint32) {
	src.tabWidth.Store(tabWidth)
}
