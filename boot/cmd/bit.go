package main

import (
	"fmt"
	"os"

	"axlab.dev/bit/core"
)

func main() {
	fmt.Println()
	if len(os.Args) > 1 {
		args := os.Args[1:]
		if err := runMain(args...); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Bit language %s\n\n", core.Version())
	}
}

func runMain(args ...string) error {
	loader, err := core.SourceLoaderNew(".")
	if err != nil {
		return err
	}

	for _, it := range args {
		src, err := loader.Load(it)
		if err != nil {
			return err
		}

		fmt.Printf("- Loaded `%s` with %d bytes\n", src.Name(), len(src.Text()))
	}
	fmt.Println()

	return nil
}
