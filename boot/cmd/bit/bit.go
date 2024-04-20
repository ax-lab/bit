package main

import (
	"fmt"
	"os"

	"axlab.dev/bit/boot/bit"
	"axlab.dev/bit/boot/core"
)

func main() {
	rootDir := core.ProjectRoot()

	compiler := bit.Compiler{}
	compiler.SetRoot(rootDir)
	compiler.SetMain("src/boot/main.bit")

	if err := compiler.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nCompilation failed: %v\n\n", err)
		os.Exit(1)
	}
}
