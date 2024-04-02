package main

import (
	"fmt"
	"os"
	"time"

	"axlab.dev/test/files"
	"axlab.dev/test/text"
)

const outputFiles = true

func main() {
	start := time.Now()

	args := os.Args[1:]
	baseDir := "."
	if len(args) > 0 {
		baseDir = args[0]
	}

	root := Check(files.Root(baseDir))
	stack := []files.File{
		Check(root.OpenFile(".")),
	}

	count := 0
	size := int64(0)

	for len(stack) > 0 {
		last := len(stack) - 1
		item := stack[last]
		stack = stack[:last]

		stat := Check(item.Stat())
		if outputFiles {
			fmt.Println(item.Path())
		}

		count += 1
		if stat.IsDir() {
			list := Check(item.ListDir())
			stack = append(stack, list...)
		} else {
			size += stat.Size()
		}
	}

	fmt.Printf("\nProcessed %d files with %s in %s\n\n", count, text.Bytes(size), time.Since(start))
}

func Check[T any](val T, err error) T {
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nfatal error: %v\n\n", err)
		os.Exit(1)
	}
	return val
}
