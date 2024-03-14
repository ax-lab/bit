package boot

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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

type sourceMap struct {
	mutex sync.Mutex
	files map[string]*Source
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

	data, err := os.ReadFile(file)
	if err != nil {
		src.err = err
		return nil, err
	}

	src.text = string(data)
	src.node = st.NewNode(src, src.Span())
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
