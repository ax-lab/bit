package boot

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

type Source struct {
	st   *State
	err  error
	node Node

	tabWidth int

	name string
	text string
}

func (src *Source) Name() string { return src.name }

func (src *Source) Text() string { return src.text }

func (src *Source) TabWidth() int {
	if src.tabWidth <= 0 {
		return 4
	}
	return src.tabWidth
}

func (src *Source) SetTabWidth(width int) {
	if width <= 0 {
		panic("Source: invalid tab width")
	}
	src.tabWidth = width
}

func (src *Source) Repr() string {
	return fmt.Sprintf("%s (%d bytes)", src.Name(), len(src.Text()))
}

func (src *Source) Cmp(other *Source) int {
	if res := cmp.Compare(src.Name(), other.Name()); res != 0 {
		return res
	}
	if res := cmp.Compare(len(src.Text()), len(other.Text())); res != 0 {
		return res
	}

	a := uintptr(unsafe.Pointer(src))
	b := uintptr(unsafe.Pointer(other))
	return cmp.Compare(a, b)
}

type sourceMap struct {
	mutex sync.Mutex
	files map[string]*Source
}

func (st *State) AddSource(name, text string) *Source {
	src := &Source{
		st:   st,
		name: name,
		text: text,
	}
	src.node = st.NewNode(src, src.Span())

	st.BindSource(src)
	src.parsePragmas()

	return src
}

func (st *State) LoadSourceFile(file string) (*Source, error) {
	st.sourceMap.mutex.Lock()
	defer st.sourceMap.mutex.Unlock()

	fullPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	if src, ok := st.files[fullPath]; ok {
		if src.err != nil {
			return nil, src.err
		}
		return src, nil
	}

	src := &Source{
		st:   st,
		name: file,
	}
	if st.files == nil {
		st.files = make(map[string]*Source)
	}
	st.files[fullPath] = src

	data, err := os.ReadFile(file)
	if err != nil {
		src.err = err
		return nil, err
	}

	src.text = string(data)
	src.node = st.NewNode(src, src.Span())

	st.BindSource(src)
	src.parsePragmas()

	return src, nil
}

func (src *Source) parsePragmas() {
	text := src.text
	header := text
	if len(header) > PragmaLoadHeaderSize {
		header = header[:PragmaLoadHeaderSize]
	}

	for n, line := range StrLines(header) {
		line = StrTrim(line)
		if strings.HasPrefix(line, PragmaLoadPrefix+" ") {
			load := StrTrim(line[len(PragmaLoadPrefix):])
			if err := src.st.PragmaLoad(src.node, load); err != nil {
				src.st.AddError(ErrorAt(err, src, n+1))
			}
		}
	}
}
