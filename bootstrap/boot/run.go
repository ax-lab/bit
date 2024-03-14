package boot

import (
	"fmt"
	"os"
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

	if !st.CheckValid(os.Stderr, "\nErrors:\n\n") {
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(src.Text())
}
