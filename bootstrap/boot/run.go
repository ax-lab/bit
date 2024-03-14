package boot

import (
	"fmt"
	"os"
	"strings"
)

func Main() {
	st := State{}
	st.RunFile("src/main.bit")
}

func (st *State) RunFile(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		Fatal(err)
	}

	text := string(data)
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
	fmt.Println(string(data))
}

func FatalAt(file string, line int, err error) {
	Fatal(fmt.Errorf("at %s:%d: %v", file, line, err))
}

func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "\nfatal: %v\n\n", err)
	os.Exit(1)
}
