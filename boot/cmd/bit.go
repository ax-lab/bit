package main

import (
	"fmt"
	"os"

	"axlab.dev/bit/core"
	"axlab.dev/bit/lang"
)

func main() {
	if len(os.Args) > 1 {
		args := os.Args[1:]
		if err := runMain(args...); err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("\nBit language %s\n\n", core.Version())
	}
}

func runMain(args ...string) error {
	loader, err := core.SourceLoaderNew(".")
	if err != nil {
		return err
	}

	compiler := core.Compiler{}
	if err := lang.Declare(&compiler); err != nil {
		return err
	}

	for _, it := range args {
		src, err := loader.Load(it)
		if err != nil {
			return err
		}

		node := core.NodeNew(src.Span(), src)
		list := core.NodeListNew(src.Span(), node)
		compiler.Add(list)
	}

	compiler.Run()
	return nil
}
