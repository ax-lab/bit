package main

import (
	"fmt"
	"os"
	"path/filepath"

	"axlab.dev/bit/boot/bit"
	"axlab.dev/bit/boot/core"
)

func main() {
	rootDir := core.ProjectRoot()

	compiler := bit.Compiler{}
	compiler.SetRoot(filepath.Join(rootDir, "src"))
	compiler.SetMain("boot/main.bit")

	if err := compiler.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n[Fatal] compilation failed: %v\n\n", err)
		os.Exit(1)
	}
}
