package boot

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"axlab.dev/bit/input"
)

type State struct {
	nodeMap
	bindingMap
	input.ErrorList

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

func (st *State) CheckValid(stdErr io.Writer, prefix string) bool {
	list := st.Errors()
	if len(list) == 0 {
		return true
	}

	fmt.Fprint(stdErr, prefix)
	for n, err := range list {
		if n > 0 {
			fmt.Fprintf(stdErr, "\n")
		}
		text := fmt.Sprintf("[%d] %s\n", n+1, err)
		fmt.Fprint(stdErr, input.Indent(text))
	}
	return false
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

		for n, line := range input.Lines(header) {
			line = input.Trim(line)
			if strings.HasPrefix(line, PragmaLoadPrefix+" ") {
				load := input.Trim(line[len(PragmaLoadPrefix):])
				if err := st.PragmaLoad(info.node, load); err != nil {
					st.AddError(input.ErrorAt(err, src, n+1))
				}
			}
		}
	}
}
