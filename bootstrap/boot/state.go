package boot

import (
	"strings"
	"sync"

	"axlab.dev/bit/input"
)

type State struct {
	nodeMap
	errorList
	bindingMap

	sources input.SourceMap
}

func (st *State) CheckDone() {
	if err := st.nodeMap.CheckDone(); err != nil {
		st.AddError(err)
	}
}

func (st *State) AddSource(name, text string) input.Source {
	src := st.sources.NewSource(name, text)
	st.initSource(src)
	return src
}

func (st *State) LoadSourceFile(name string) (input.Source, error) {
	src, err := st.sources.LoadFile(name)
	if err != nil {
		return src, err
	}

	st.initSource(src)
	return src, nil
}

type sourceInfo struct {
	sync sync.Mutex
	init bool
	node Node
}

func (st *State) initSource(src input.Source) {
	info := input.SourceGet[sourceInfo](src)
	info.sync.Lock()
	defer info.sync.Unlock()

	if !info.init {
		info.init = true
		info.node = st.NewNode(src, src.Span())
		st.BindSource(src)

		// parse pragmas
		text := src.Text()
		header := text
		if len(header) > PragmaLoadHeaderSize {
			header = header[:PragmaLoadHeaderSize]
		}

		for n, line := range StrLines(header) {
			line = StrTrim(line)
			if strings.HasPrefix(line, PragmaLoadPrefix+" ") {
				load := StrTrim(line[len(PragmaLoadPrefix):])
				if err := st.PragmaLoad(info.node, load); err != nil {
					st.AddError(ErrorAt(err, src, n+1))
				}
			}
		}
	}
}
