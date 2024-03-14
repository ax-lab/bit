package boot

import (
	"fmt"
	"os"
)

func Main() {
	RunFile("src/main.bit")
}

func RunFile(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		Fatal(err)
	}

	fmt.Println(string(data))
}

func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "\nfatal: %v\n\n", err)
	os.Exit(1)
}
