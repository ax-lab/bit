package input

import (
	"cmp"
	"fmt"
	"reflect"
	"sync"
	"unicode"
)

const DefaultTabSize = 4

func IsSpace(chr rune) bool {
	return chr != '\n' && chr != '\r' && unicode.IsSpace(chr)
}

type Source struct {
	data *sourceData
}

type sourceData struct {
	index   uint64
	parent  *SourceMap
	err     error
	name    string
	text    string
	tabSize int

	infoMap sync.Map
}

func (src Source) Valid() bool {
	return src.data != nil
}

func (src Source) Name() string {
	if src.data == nil {
		return ""
	}
	return src.data.name
}

func (src Source) Len() int {
	if src.data == nil {
		return 0
	}
	return len(src.data.text)
}

func (src Source) Text() string {
	if src.data == nil {
		return ""
	}
	return src.data.text
}

func (src Source) TabSize() int {
	if data := src.data; data != nil && data.tabSize > 0 {
		return data.tabSize
	}
	return DefaultTabSize
}

func (src Source) SetTabSize(size int) {
	if size <= 0 {
		panic("Source: invalid tab size")
	}
	if src.data != nil {
		src.data.tabSize = size
	}
}

func (src Source) Repr() string {
	if src.data == nil {
		return "Source()"
	}
	return fmt.Sprintf("Source(%s / %d bytes)", src.Name(), len(src.Text()))
}

func (src Source) Cmp(other Source) int {
	a := src.data
	b := other.data
	if a == b {
		return 0
	} else if a == nil {
		return -1
	} else if b == nil {
		return 1
	}

	if res := cmp.Compare(a.name, b.name); res != 0 {
		return res
	}

	if res := cmp.Compare(len(a.text), len(b.text)); res != 0 {
		return res
	}

	return cmp.Compare(a.index, b.index)
}

type sourceTypeData struct {
	init chan struct{}
	data any
}

func SourceGet[T any](src Source) *T {
	if src.data == nil {
		panic("SourceInfo: invalid source")
	}

	var (
		data T
		info *sourceTypeData
	)

	key := reflect.TypeOf(data)
	if val, ok := src.data.infoMap.Load(key); ok {
		info = val.(*sourceTypeData)
	} else {
		val, loaded := src.data.infoMap.LoadOrStore(key, &sourceTypeData{
			init: make(chan struct{}),
			data: &data,
		})

		info = val.(*sourceTypeData)
		if !loaded {
			close(info.init)
		}
	}

	<-info.init
	return info.data.(*T)
}
