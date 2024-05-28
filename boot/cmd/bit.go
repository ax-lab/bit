package main

import (
	"fmt"
	"os"

	"axlab.dev/bit/core"
	"axlab.dev/bit/lang"
)

func main() {
	defer fmt.Println()
	if len(os.Args) > 1 {
		args := os.Args[1:]
		if err := runMain(args...); err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("\nBit language %s\n", core.Version())
	}
}

func runMain(args ...string) error {
	compiler := core.Compiler{}
	if err := compiler.Sources.SetBaseDir("."); err != nil {
		return err
	}

	if err := lang.Declare(&compiler); err != nil {
		return err
	}

	for _, it := range args {
		compiler.AddSource(it)
	}

	rt := compiler.CreateRuntime()
	if !rt.Run() {
		return fmt.Errorf("compilation failed")
	}

	return nil
}
