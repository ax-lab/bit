package boot

import (
	"fmt"
	"strings"
)

func Main() {
	st := State{}
	st.RunFile("src/main.bit")
}

func (st *State) RunFile(file string) {
	src, err := st.LoadSourceFile(file)
	if err != nil {
		Fatal(err)
	}

	text := src.Text
	header := text
	if len(header) > PragmaLoadHeaderSize {
		header = header[:PragmaLoadHeaderSize]
	}

	for n, line := range StrLines(header) {
		line = StrTrim(line)
		if strings.HasPrefix(line, PragmaLoadPrefix+" ") {
			load := StrTrim(line[len(PragmaLoadPrefix):])
			if err := st.PragmaLoad(load); err != nil {
				FatalAt(file, n+1, err)
			}
		}
	}
	fmt.Println(src.Text)
}
